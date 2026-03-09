package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/omni-wallet/wallet-service/internal/core/domain"
)

type WalletRepository struct {
	db *sqlx.DB
}

func NewWalletRepository(db *sqlx.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) Create(ctx context.Context, wallet *domain.Wallet) (*domain.Wallet, error) {
	query := `
		INSERT INTO wallets (id, user_id, balance, status, created_at, updated_at)
		VALUES (:id, :user_id, :balance, :status, :created_at, :updated_at)
	`
	db := extractDB(ctx, r.db)
	if _, err := db.NamedExecContext(ctx, query, wallet); err != nil {
		return nil, fmt.Errorf("inserting wallet: %w", err)
	}
	return wallet, nil
}

func (r *WalletRepository) FindByID(ctx context.Context, id string) (*domain.Wallet, error) {
	db := extractDB(ctx, r.db)
	var wallet domain.Wallet
	if err := db.GetContext(ctx, &wallet,
		`SELECT id, user_id, balance, status, created_at, updated_at FROM wallets WHERE id = ? LIMIT 1`, id,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying wallet by id: %w", err)
	}
	return &wallet, nil
}

func (r *WalletRepository) FindByUserID(ctx context.Context, userID string) (*domain.Wallet, error) {
	db := extractDB(ctx, r.db)
	var wallet domain.Wallet
	if err := db.GetContext(ctx, &wallet,
		`SELECT id, user_id, balance, status, created_at, updated_at FROM wallets WHERE user_id = ? LIMIT 1`, userID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying wallet by user_id: %w", err)
	}
	return &wallet, nil
}

func (r *WalletRepository) UpdateBalance(ctx context.Context, walletID string, newBalance int64) error {
	db := extractDB(ctx, r.db)
	result, err := db.ExecContext(ctx,
		`UPDATE wallets SET balance = ?, updated_at = NOW() WHERE id = ?`, newBalance, walletID,
	)
	if err != nil {
		return fmt.Errorf("updating wallet balance: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("no wallet found with id %s", walletID)
	}
	return nil
}
