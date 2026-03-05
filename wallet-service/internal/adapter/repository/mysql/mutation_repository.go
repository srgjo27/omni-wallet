package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/omni-wallet/wallet-service/internal/core/domain"
)

type MutationRepository struct {
	db *sqlx.DB
}

func NewMutationRepository(db *sqlx.DB) *MutationRepository {
	return &MutationRepository{db: db}
}

func (r *MutationRepository) Create(ctx context.Context, mutation *domain.WalletMutation) (*domain.WalletMutation, error) {
	query := `
		INSERT INTO wallet_mutations (id, wallet_id, transaction_id, direction, amount, balance_after, description, created_at)
		VALUES (:id, :wallet_id, :transaction_id, :direction, :amount, :balance_after, :description, :created_at)
	`
	if mutation.CreatedAt.IsZero() {
		mutation.CreatedAt = time.Now()
	}
	db := extractDB(ctx, r.db)
	if _, err := db.NamedExecContext(ctx, query, mutation); err != nil {
		return nil, fmt.Errorf("inserting wallet mutation: %w", err)
	}
	return mutation, nil
}

func (r *MutationRepository) ListByWalletID(ctx context.Context, walletID string, page, limit int) ([]*domain.WalletMutation, int, error) {
	db := extractDB(ctx, r.db)
	offset := (page - 1) * limit

	var total int
	if err := db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM wallet_mutations WHERE wallet_id = ?`, walletID,
	); err != nil {
		return nil, 0, fmt.Errorf("counting mutations: %w", err)
	}

	var mutations []*domain.WalletMutation
	if err := db.SelectContext(ctx, &mutations,
		`SELECT id, wallet_id, transaction_id, direction, amount, balance_after, description, created_at
		 FROM wallet_mutations
		 WHERE wallet_id = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		walletID, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("listing mutations: %w", err)
	}

	return mutations, total, nil
}
