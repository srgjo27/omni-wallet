package ports

import (
	"context"

	"github.com/omni-wallet/wallet-service/internal/core/domain"
)

// WalletRepository defines the contract for wallet persistence.
type WalletRepository interface {
	// Create persists a new wallet.
	Create(ctx context.Context, wallet *domain.Wallet) (*domain.Wallet, error)

	// FindByID retrieves a wallet by its primary key.
	FindByID(ctx context.Context, id string) (*domain.Wallet, error)

	// FindByUserID retrieves the wallet belonging to the given user.
	FindByUserID(ctx context.Context, userID string) (*domain.Wallet, error)

	// UpdateBalanceWithLock updates the wallet balance inside an existing DB transaction.
	// The caller is responsible for providing a transactional context (via TxProvider).
	UpdateBalance(ctx context.Context, walletID string, newBalance int64) error
}

// TransactionRepository defines the contract for transaction persistence.
type TransactionRepository interface {
	// Create persists a new transaction record.
	Create(ctx context.Context, tx *domain.Transaction) (*domain.Transaction, error)

	// FindByID retrieves a transaction by primary key.
	FindByID(ctx context.Context, id string) (*domain.Transaction, error)

	// FindByReferenceNo retrieves a transaction by its unique reference number.
	// Used for idempotency checks.
	FindByReferenceNo(ctx context.Context, referenceNo string) (*domain.Transaction, error)

	// UpdateStatus transitions the status (PENDING → SUCCESS | FAILED) of a transaction.
	UpdateStatus(ctx context.Context, id string, status domain.TransactionStatus) error

	// ListByWalletID returns a paginated list of transactions for a given wallet.
	ListByWalletID(ctx context.Context, walletID string, page, limit int) ([]*domain.Transaction, int, error)
}

// MutationRepository defines the contract for wallet mutation (ledger) persistence.
type MutationRepository interface {
	// Create appends a new mutation entry to the ledger.
	Create(ctx context.Context, mutation *domain.WalletMutation) (*domain.WalletMutation, error)

	// ListByWalletID returns a paginated list of mutations for a given wallet.
	ListByWalletID(ctx context.Context, walletID string, page, limit int) ([]*domain.WalletMutation, int, error)
}

// TxProvider defines the contract for running multiple repository operations
// inside a single ACID database transaction. This keeps DB transaction logic
// out of the service layer while still allowing atomicity.
type TxProvider interface {
	// ExecTx executes the given function atomically.
	// If fn returns an error, the transaction is rolled back; otherwise committed.
	ExecTx(ctx context.Context, fn func(ctx context.Context) error) error
}
