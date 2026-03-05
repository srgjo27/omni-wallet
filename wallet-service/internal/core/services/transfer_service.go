package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/omni-wallet/wallet-service/internal/core/domain"
	"github.com/omni-wallet/wallet-service/internal/core/ports"
)

const (
	// distributedLockTTL defines how long a wallet lock is held before auto-expiry.
	// This acts as a safety net in case the service crashes mid-transaction.
	distributedLockTTL = 10 * time.Second

	// idempotencyKeyTTL defines how long the cached result of a transaction is stored.
	idempotencyKeyTTL = 24 * time.Hour

	// idempotencyKeyPrefix is the Redis key namespace for idempotency values.
	idempotencyKeyPrefix = "idempotency:"

	// lockKeyPrefix is the Redis key namespace for distributed wallet locks.
	lockKeyPrefix = "wallet_lock:"
)

var (
	ErrInsufficientBalance   = errors.New("insufficient wallet balance")
	ErrSameWallet            = errors.New("source and target wallet must be different")
	ErrTransactionDuplicate  = errors.New("transaction with this reference_no already exists")
	ErrInvalidPIN            = errors.New("invalid transaction PIN")
	ErrLockAcquireFailed     = errors.New("could not acquire wallet lock, please try again")
)

// TransferService handles all money-movement operations: top-up and P2P transfer.
// It enforces:
//  1. Idempotency  — duplicate reference_no returns the cached result without reprocessing.
//  2. Distributed Lock — wallet balance is protected by a Redis mutex before modification.
//  3. ACID transaction — all DB mutations happen in a single MySQL transaction via TxProvider.
type TransferService struct {
	walletRepo    ports.WalletRepository
	txRepo        ports.TransactionRepository
	mutationRepo  ports.MutationRepository
	txProvider    ports.TxProvider
	lockRepo      ports.DistributedLockRepository
	idempotency   ports.IdempotencyRepository
	userClient    ports.UserServiceClient
	// eventPublisher is optional — when nil, event publishing is skipped.
	eventPublisher ports.EventPublisher
}

func NewTransferService(
	walletRepo ports.WalletRepository,
	txRepo ports.TransactionRepository,
	mutationRepo ports.MutationRepository,
	txProvider ports.TxProvider,
	lockRepo ports.DistributedLockRepository,
	idempotency ports.IdempotencyRepository,
	userClient ports.UserServiceClient,
	eventPublisher ports.EventPublisher,
) *TransferService {
	return &TransferService{
		walletRepo:     walletRepo,
		txRepo:         txRepo,
		mutationRepo:   mutationRepo,
		txProvider:     txProvider,
		lockRepo:       lockRepo,
		idempotency:    idempotency,
		userClient:     userClient,
		eventPublisher: eventPublisher,
	}
}

// Topup credits the wallet of the given user. It is called from the mock
// Virtual Account webhook and is fully idempotent.
func (s *TransferService) Topup(ctx context.Context, req domain.TopupRequest) (*domain.Transaction, error) {
	idempotencyKey := idempotencyKeyPrefix + req.ReferenceNo
	cached, err := s.idempotency.Get(ctx, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("checking idempotency: %w", err)
	}
	if cached != "" {
		var cachedTx domain.Transaction
		if err := json.Unmarshal([]byte(cached), &cachedTx); err == nil {
			return &cachedTx, nil
		}
	}

	wallet, err := s.walletRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("finding wallet: %w", err)
	}
	if wallet == nil {
		return nil, ErrWalletNotFound
	}
	if wallet.Status != domain.WalletStatusActive {
		return nil, ErrWalletFrozen
	}

	lockKey := lockKeyPrefix + wallet.ID
	acquired, err := s.lockRepo.Acquire(ctx, lockKey, distributedLockTTL)
	if err != nil {
		return nil, fmt.Errorf("acquiring wallet lock: %w", err)
	}
	if !acquired {
		return nil, ErrLockAcquireFailed
	}
	defer s.lockRepo.Release(ctx, lockKey) //nolint:errcheck

	now := time.Now()
	txRecord := &domain.Transaction{
		ID:             uuid.New().String(),
		ReferenceNo:    req.ReferenceNo,
		Type:           domain.TransactionTypeTopup,
		Amount:         req.Amount,
		Status:         domain.TransactionStatusPending,
		TargetWalletID: wallet.ID,
		Description:    "Top-up via Virtual Account",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	var createdTx *domain.Transaction

	if err := s.txProvider.ExecTx(ctx, func(txCtx context.Context) error {
		created, err := s.txRepo.Create(txCtx, txRecord)
		if err != nil {
			return fmt.Errorf("creating transaction: %w", err)
		}
		createdTx = created

		newBalance := wallet.Balance + req.Amount
		if err := s.walletRepo.UpdateBalance(txCtx, wallet.ID, newBalance); err != nil {
			return fmt.Errorf("updating wallet balance: %w", err)
		}

		mutation := &domain.WalletMutation{
			ID:            uuid.New().String(),
			WalletID:      wallet.ID,
			TransactionID: createdTx.ID,
			Direction:     domain.MutationDirectionCredit,
			Amount:        req.Amount,
			BalanceAfter:  newBalance,
			Description:   createdTx.Description,
			CreatedAt:     now,
		}
		if _, err := s.mutationRepo.Create(txCtx, mutation); err != nil {
			return fmt.Errorf("creating mutation: %w", err)
		}

		if err := s.txRepo.UpdateStatus(txCtx, createdTx.ID, domain.TransactionStatusSuccess); err != nil {
			return fmt.Errorf("updating transaction status: %w", err)
		}
		createdTx.Status = domain.TransactionStatusSuccess

		return nil
	}); err != nil {
		_ = s.txRepo.UpdateStatus(ctx, txRecord.ID, domain.TransactionStatusFailed)
		return nil, fmt.Errorf("topup transaction failed: %w", err)
	}

	s.cacheTransaction(ctx, idempotencyKey, createdTx)

	s.publishEvent(ctx, domain.OutboundEvent{
		EventType:   domain.OutboundEventTopupSuccess,
		ReferenceNo: createdTx.ReferenceNo,
		UserID:      req.UserID,
		Amount:      createdTx.Amount,
		OccurredAt:  createdTx.CreatedAt,
	})

	return createdTx, nil
}

// Transfer executes a P2P transfer between two user wallets.
// It acquires distributed locks on BOTH wallets (in a deterministic order to prevent
// deadlocks), validates the source balance, and commits all mutations atomically.
func (s *TransferService) Transfer(ctx context.Context, req domain.TransferRequest) (*domain.Transaction, error) {
	if req.SourceUserID == req.TargetUserID {
		return nil, ErrSameWallet
	}

	idempotencyKey := idempotencyKeyPrefix + req.ReferenceNo
	cached, err := s.idempotency.Get(ctx, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("checking idempotency: %w", err)
	}
	if cached != "" {
		var cachedTx domain.Transaction
		if err := json.Unmarshal([]byte(cached), &cachedTx); err == nil {
			return &cachedTx, nil
		}
	}

	if err := s.userClient.VerifyPIN(req.SourceUserID, req.TransactionPIN); err != nil {
		return nil, ErrInvalidPIN
	}

	sourceWallet, err := s.walletRepo.FindByUserID(ctx, req.SourceUserID)
	if err != nil {
		return nil, fmt.Errorf("finding source wallet: %w", err)
	}
	if sourceWallet == nil {
		return nil, fmt.Errorf("source %w", ErrWalletNotFound)
	}
	if sourceWallet.Status != domain.WalletStatusActive {
		return nil, fmt.Errorf("source wallet: %w", ErrWalletFrozen)
	}

	targetWallet, err := s.walletRepo.FindByUserID(ctx, req.TargetUserID)
	if err != nil {
		return nil, fmt.Errorf("finding target wallet: %w", err)
	}
	if targetWallet == nil {
		return nil, fmt.Errorf("target %w", ErrWalletNotFound)
	}
	if targetWallet.Status != domain.WalletStatusActive {
		return nil, fmt.Errorf("target wallet: %w", ErrWalletFrozen)
	}

	// Acquire distributed locks in deterministic order
	// Sorting by wallet ID before locking prevents deadlocks when two concurrent transfers
	// involve the same pair of wallets in opposite directions.
	lockKeys := []string{lockKeyPrefix + sourceWallet.ID, lockKeyPrefix + targetWallet.ID}
	sort.Strings(lockKeys)

	for _, key := range lockKeys {
		acquired, err := s.lockRepo.Acquire(ctx, key, distributedLockTTL)
		if err != nil {
			return nil, fmt.Errorf("acquiring lock %s: %w", key, err)
		}
		if !acquired {
			return nil, ErrLockAcquireFailed
		}
	}
	defer func() {
		for _, key := range lockKeys {
			s.lockRepo.Release(ctx, key) //nolint:errcheck
		}
	}()

	if sourceWallet.Balance < req.Amount {
		return nil, ErrInsufficientBalance
	}

	now := time.Now()
	description := req.Description
	if description == "" {
		description = "P2P Transfer"
	}

	txRecord := &domain.Transaction{
		ID:             uuid.New().String(),
		ReferenceNo:    req.ReferenceNo,
		Type:           domain.TransactionTypeP2P,
		Amount:         req.Amount,
		Status:         domain.TransactionStatusPending,
		SourceWalletID: sourceWallet.ID,
		TargetWalletID: targetWallet.ID,
		Description:    description,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	var createdTx *domain.Transaction

	if err := s.txProvider.ExecTx(ctx, func(txCtx context.Context) error {
		created, err := s.txRepo.Create(txCtx, txRecord)
		if err != nil {
			return fmt.Errorf("creating transaction: %w", err)
		}
		createdTx = created

		sourceNewBalance := sourceWallet.Balance - req.Amount
		if err := s.walletRepo.UpdateBalance(txCtx, sourceWallet.ID, sourceNewBalance); err != nil {
			return fmt.Errorf("debiting source wallet: %w", err)
		}

		targetNewBalance := targetWallet.Balance + req.Amount
		if err := s.walletRepo.UpdateBalance(txCtx, targetWallet.ID, targetNewBalance); err != nil {
			return fmt.Errorf("crediting target wallet: %w", err)
		}

		debitMutation := &domain.WalletMutation{
			ID:            uuid.New().String(),
			WalletID:      sourceWallet.ID,
			TransactionID: createdTx.ID,
			Direction:     domain.MutationDirectionDebit,
			Amount:        req.Amount,
			BalanceAfter:  sourceNewBalance,
			Description:   description,
			CreatedAt:     now,
		}
		if _, err := s.mutationRepo.Create(txCtx, debitMutation); err != nil {
			return fmt.Errorf("creating debit mutation: %w", err)
		}

		creditMutation := &domain.WalletMutation{
			ID:            uuid.New().String(),
			WalletID:      targetWallet.ID,
			TransactionID: createdTx.ID,
			Direction:     domain.MutationDirectionCredit,
			Amount:        req.Amount,
			BalanceAfter:  targetNewBalance,
			Description:   description,
			CreatedAt:     now,
		}
		if _, err := s.mutationRepo.Create(txCtx, creditMutation); err != nil {
			return fmt.Errorf("creating credit mutation: %w", err)
		}

		if err := s.txRepo.UpdateStatus(txCtx, createdTx.ID, domain.TransactionStatusSuccess); err != nil {
			return fmt.Errorf("updating transaction status: %w", err)
		}
		createdTx.Status = domain.TransactionStatusSuccess

		return nil
	}); err != nil {
		_ = s.txRepo.UpdateStatus(ctx, txRecord.ID, domain.TransactionStatusFailed)
		return nil, fmt.Errorf("transfer transaction failed: %w", err)
	}

	s.cacheTransaction(ctx, idempotencyKey, createdTx)

	s.publishEvent(ctx, domain.OutboundEvent{
		EventType:    domain.OutboundEventTransferSuccess,
		ReferenceNo:  createdTx.ReferenceNo,
		UserID:       req.SourceUserID,
		TargetUserID: req.TargetUserID,
		Amount:       createdTx.Amount,
		OccurredAt:   createdTx.CreatedAt,
	})

	return createdTx, nil
}

// cacheTransaction serialises a transaction and stores it under the idempotency key.
// Failures here are non-fatal since the DB is already consistent.
func (s *TransferService) cacheTransaction(ctx context.Context, key string, tx *domain.Transaction) {
	data, err := json.Marshal(tx)
	if err != nil {
		return
	}
	_ = s.idempotency.Set(ctx, key, string(data), idempotencyKeyTTL)
}

// publishEvent sends an outbound event to the message broker.
// Failures are logged but not propagated — the transaction is already committed.
func (s *TransferService) publishEvent(ctx context.Context, event domain.OutboundEvent) {
	if s.eventPublisher == nil {
		return
	}
	if err := s.eventPublisher.Publish(ctx, event); err != nil {
		// Use log directly to avoid an import cycle; in production use a structured logger.
		fmt.Printf("[transfer-service] WARNING: failed to publish event %s: %v\n", event.EventType, err)
	}
}
