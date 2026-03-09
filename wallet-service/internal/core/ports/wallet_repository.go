package ports

import (
	"context"

	"github.com/omni-wallet/wallet-service/internal/core/domain"
)

type WalletRepository interface {
	Create(ctx context.Context, wallet *domain.Wallet) (*domain.Wallet, error)
	FindByID(ctx context.Context, id string) (*domain.Wallet, error)
	FindByUserID(ctx context.Context, userID string) (*domain.Wallet, error)
	UpdateBalance(ctx context.Context, walletID string, newBalance int64) error
}

type TransactionRepository interface {
	Create(ctx context.Context, tx *domain.Transaction) (*domain.Transaction, error)
	FindByID(ctx context.Context, id string) (*domain.Transaction, error)
	FindByReferenceNo(ctx context.Context, referenceNo string) (*domain.Transaction, error)
	UpdateStatus(ctx context.Context, id string, status domain.TransactionStatus) error
	ListByWalletID(ctx context.Context, walletID string, page, limit int) ([]*domain.Transaction, int, error)
	GetAdminStats(ctx context.Context) (totalTx int, totalVolume int64, err error)
}

type MutationRepository interface {
	Create(ctx context.Context, mutation *domain.WalletMutation) (*domain.WalletMutation, error)
	ListByWalletID(ctx context.Context, walletID string, page, limit int) ([]*domain.WalletMutation, int, error)
}

type TxProvider interface {
	ExecTx(ctx context.Context, fn func(ctx context.Context) error) error
}
