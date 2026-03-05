package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/omni-wallet/wallet-service/internal/core/domain"
	"github.com/omni-wallet/wallet-service/internal/core/ports"
)

var (
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrWalletAlreadyExists = errors.New("wallet already exists for this user")
	ErrWalletFrozen        = errors.New("wallet is frozen and cannot be used")
)

// WalletService handles operations that do NOT move money between wallets:
// creating wallets, reading balances, and listing mutations.
type WalletService struct {
	walletRepo   ports.WalletRepository
	mutationRepo ports.MutationRepository
	txRepo       ports.TransactionRepository
}

func NewWalletService(
	walletRepo ports.WalletRepository,
	mutationRepo ports.MutationRepository,
	txRepo ports.TransactionRepository,
) *WalletService {
	return &WalletService{
		walletRepo:   walletRepo,
		mutationRepo: mutationRepo,
		txRepo:       txRepo,
	}
}

// CreateWallet opens a new wallet with zero balance for the given user.
// This is typically called right after a user registers.
func (s *WalletService) CreateWallet(ctx context.Context, req domain.CreateWalletRequest) (*domain.Wallet, error) {
	existing, err := s.walletRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("checking existing wallet: %w", err)
	}
	if existing != nil {
		return nil, ErrWalletAlreadyExists
	}

	now := time.Now()
	wallet := &domain.Wallet{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Balance:   0,
		Status:    domain.WalletStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	created, err := s.walletRepo.Create(ctx, wallet)
	if err != nil {
		return nil, fmt.Errorf("persisting wallet: %w", err)
	}

	return created, nil
}

func (s *WalletService) GetBalance(ctx context.Context, userID string) (*domain.BalanceResponse, error) {
	wallet, err := s.walletRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("finding wallet: %w", err)
	}
	if wallet == nil {
		return nil, ErrWalletNotFound
	}

	return &domain.BalanceResponse{
		WalletID: wallet.ID,
		UserID:   wallet.UserID,
		Balance:  wallet.Balance,
		Currency: "IDR",
		Status:   string(wallet.Status),
	}, nil
}

func (s *WalletService) GetMutations(ctx context.Context, userID string, page, limit int) (*domain.MutationListResponse, error) {
	wallet, err := s.walletRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("finding wallet: %w", err)
	}
	if wallet == nil {
		return nil, ErrWalletNotFound
	}

	mutations, total, err := s.mutationRepo.ListByWalletID(ctx, wallet.ID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("listing mutations: %w", err)
	}

	return &domain.MutationListResponse{
		Mutations: mutations,
		Total:     total,
		Page:      page,
		Limit:     limit,
	}, nil
}

func (s *WalletService) GetTransactionHistory(ctx context.Context, userID string, page, limit int) (*domain.TransactionHistoryResponse, error) {
	wallet, err := s.walletRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("finding wallet: %w", err)
	}
	if wallet == nil {
		return nil, ErrWalletNotFound
	}

	transactions, total, err := s.txRepo.ListByWalletID(ctx, wallet.ID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("listing transactions: %w", err)
	}

	return &domain.TransactionHistoryResponse{
		Transactions: transactions,
		Total:        total,
		Page:         page,
		Limit:        limit,
	}, nil
}
