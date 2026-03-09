package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/omni-wallet/wallet-service/internal/core/domain"
	"github.com/omni-wallet/wallet-service/internal/core/ports"
)

var (
	ErrPaymentGatewayNotConfigured = errors.New("payment gateway not configured")
	ErrPaymentAlreadyProcessed     = errors.New("payment already processed")
	ErrInvalidCallbackToken        = errors.New("invalid callback token")
	ErrInvalidExternalID           = errors.New("invalid external_id format")
)

const vaCacheTTL = 24 * 365 * time.Hour

const paymentDedupTTL = 48 * time.Hour

func vaExternalID(userID, bankCode string) string {
	return fmt.Sprintf("omni-va-%s::%s", userID, bankCode)
}

func parseExternalID(externalID string) (userID, bankCode string, err error) {
	trimmed := strings.TrimPrefix(externalID, "omni-va-")
	if trimmed == externalID {
		return "", "", ErrInvalidExternalID
	}
	parts := strings.SplitN(trimmed, "::", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", ErrInvalidExternalID
	}
	return parts[0], parts[1], nil
}

type PaymentService struct {
	gateway      ports.PaymentGateway
	cache        ports.IdempotencyRepository
	walletRepo   ports.WalletRepository
	txRepo       ports.TransactionRepository
	mutationRepo ports.MutationRepository
	txProvider   ports.TxProvider
	publisher    ports.EventPublisher
}

func NewPaymentService(
	gateway ports.PaymentGateway,
	cache ports.IdempotencyRepository,
	walletRepo ports.WalletRepository,
	txRepo ports.TransactionRepository,
	mutationRepo ports.MutationRepository,
	txProvider ports.TxProvider,
	publisher ports.EventPublisher,
) *PaymentService {
	return &PaymentService{
		gateway:      gateway,
		cache:        cache,
		walletRepo:   walletRepo,
		txRepo:       txRepo,
		mutationRepo: mutationRepo,
		txProvider:   txProvider,
		publisher:    publisher,
	}
}

func (s *PaymentService) Gateway() ports.PaymentGateway {
	return s.gateway
}

func (s *PaymentService) GetOrCreateVA(ctx context.Context, userID, name, bankCode string) (*domain.VirtualAccount, error) {
	if s.gateway == nil {
		return nil, ErrPaymentGatewayNotConfigured
	}

	cacheKey := fmt.Sprintf("xendit:va:%s::%s", userID, bankCode)

	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var va domain.VirtualAccount
		if err := json.Unmarshal([]byte(cached), &va); err == nil {
			return &va, nil
		}
	}

	extID := vaExternalID(userID, bankCode)
	va, err := s.gateway.CreateFixedVA(extID, name, bankCode)
	if err != nil {
		return nil, fmt.Errorf("creating fixed VA: %w", err)
	}

	if data, marshalErr := json.Marshal(va); marshalErr == nil {
		_ = s.cache.Set(ctx, cacheKey, string(data), vaCacheTTL)
	}

	return va, nil
}

func (s *PaymentService) ProcessXenditPayment(ctx context.Context, payment domain.XenditVAPayment) (*domain.Transaction, error) {
	dedupKey := fmt.Sprintf("xendit:pay:%s", payment.PaymentID)
	existing, _ := s.cache.Get(ctx, dedupKey)
	if existing != "" {
		var tx domain.Transaction
		if err := json.Unmarshal([]byte(existing), &tx); err == nil {
			return &tx, nil
		}
		return nil, ErrPaymentAlreadyProcessed
	}

	userID, _, err := parseExternalID(payment.ExternalID)
	if err != nil {
		return nil, fmt.Errorf("parsing external_id %q: %w", payment.ExternalID, err)
	}

	wallet, err := s.walletRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("finding wallet for user %s: %w", userID, err)
	}
	if wallet == nil {
		return nil, ErrWalletNotFound
	}
	if wallet.Status != domain.WalletStatusActive {
		return nil, ErrWalletFrozen
	}

	amount := int64(payment.Amount)

	now := time.Now()
	referenceNo := fmt.Sprintf("xendit-%s", payment.PaymentID)

	txRecord := &domain.Transaction{
		ID:             uuid.New().String(),
		ReferenceNo:    referenceNo,
		Type:           domain.TransactionTypeTopup,
		Amount:         amount,
		Status:         domain.TransactionStatusPending,
		TargetWalletID: wallet.ID,
		Description:    fmt.Sprintf("Top-up via Xendit VA (%s)", payment.BankCode),
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

		newBalance := wallet.Balance + amount
		if err := s.walletRepo.UpdateBalance(txCtx, wallet.ID, newBalance); err != nil {
			return fmt.Errorf("updating balance: %w", err)
		}

		mutation := &domain.WalletMutation{
			ID:            uuid.New().String(),
			WalletID:      wallet.ID,
			TransactionID: createdTx.ID,
			Direction:     domain.MutationDirectionCredit,
			Amount:        amount,
			BalanceAfter:  newBalance,
			Description:   createdTx.Description,
			CreatedAt:     now,
		}
		if _, err := s.mutationRepo.Create(txCtx, mutation); err != nil {
			return fmt.Errorf("creating mutation: %w", err)
		}

		if err := s.txRepo.UpdateStatus(txCtx, createdTx.ID, domain.TransactionStatusSuccess); err != nil {
			return fmt.Errorf("updating tx status: %w", err)
		}
		createdTx.Status = domain.TransactionStatusSuccess
		return nil
	}); err != nil {
		_ = s.txRepo.UpdateStatus(ctx, txRecord.ID, domain.TransactionStatusFailed)
		return nil, fmt.Errorf("xendit topup failed: %w", err)
	}

	if data, marshalErr := json.Marshal(createdTx); marshalErr == nil {
		_ = s.cache.Set(ctx, dedupKey, string(data), paymentDedupTTL)
	}

	if s.publisher != nil {
		_ = s.publisher.Publish(ctx, domain.OutboundEvent{
			EventType:   domain.OutboundEventTopupSuccess,
			ReferenceNo: createdTx.ReferenceNo,
			UserID:      userID,
			Amount:      createdTx.Amount,
			OccurredAt:  createdTx.CreatedAt,
		})
	}

	return createdTx, nil
}
