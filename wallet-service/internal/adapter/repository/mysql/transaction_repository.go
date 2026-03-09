package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/omni-wallet/wallet-service/internal/core/domain"
)

type TransactionRepository struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *domain.Transaction) (*domain.Transaction, error) {
	query := `
		INSERT INTO transactions
			(id, reference_no, type, amount, status, source_wallet_id, target_wallet_id, description, created_at, updated_at)
		VALUES
			(:id, :reference_no, :type, :amount, :status, :source_wallet_id, :target_wallet_id, :description, :created_at, :updated_at)
	`
	db := extractDB(ctx, r.db)
	if _, err := db.NamedExecContext(ctx, query, tx); err != nil {
		return nil, fmt.Errorf("inserting transaction: %w", err)
	}
	return tx, nil
}

func (r *TransactionRepository) FindByID(ctx context.Context, id string) (*domain.Transaction, error) {
	db := extractDB(ctx, r.db)
	var tx domain.Transaction
	if err := db.GetContext(ctx, &tx,
		`SELECT id, reference_no, type, amount, status, source_wallet_id, target_wallet_id, description, created_at, updated_at
		 FROM transactions WHERE id = ? LIMIT 1`, id,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying transaction by id: %w", err)
	}
	return &tx, nil
}

func (r *TransactionRepository) FindByReferenceNo(ctx context.Context, referenceNo string) (*domain.Transaction, error) {
	db := extractDB(ctx, r.db)
	var tx domain.Transaction
	if err := db.GetContext(ctx, &tx,
		`SELECT id, reference_no, type, amount, status, source_wallet_id, target_wallet_id, description, created_at, updated_at
		 FROM transactions WHERE reference_no = ? LIMIT 1`, referenceNo,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying transaction by reference_no: %w", err)
	}
	return &tx, nil
}

func (r *TransactionRepository) UpdateStatus(ctx context.Context, id string, status domain.TransactionStatus) error {
	db := extractDB(ctx, r.db)
	result, err := db.ExecContext(ctx,
		`UPDATE transactions SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("updating transaction status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("no transaction found with id %s", id)
	}
	return nil
}

func (r *TransactionRepository) ListByWalletID(ctx context.Context, walletID string, page, limit int) ([]*domain.Transaction, int, error) {
	db := extractDB(ctx, r.db)

	offset := (page - 1) * limit

	var total int
	if err := db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM transactions WHERE source_wallet_id = ? OR target_wallet_id = ?`,
		walletID, walletID,
	); err != nil {
		return nil, 0, fmt.Errorf("counting transactions: %w", err)
	}

	var transactions []*domain.Transaction
	if err := db.SelectContext(ctx, &transactions,
		`SELECT id, reference_no, type, amount, status, source_wallet_id, target_wallet_id, description, created_at, updated_at
		 FROM transactions
		 WHERE source_wallet_id = ? OR target_wallet_id = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		walletID, walletID, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("listing transactions: %w", err)
	}

	return transactions, total, nil
}

func (r *TransactionRepository) GetAdminStats(ctx context.Context) (totalTx int, totalVolume int64, err error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE status = 'SUCCESS'
	`)
	if err := row.Scan(&totalTx, &totalVolume); err != nil {
		return 0, 0, fmt.Errorf("querying transaction stats: %w", err)
	}
	return totalTx, totalVolume, nil
}
