package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type txContextKey struct{}

type dbCtx interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

type TxProvider struct {
	db *sqlx.DB
}

func NewTxProvider(db *sqlx.DB) *TxProvider {
	return &TxProvider{db: db}
}

func (p *TxProvider) ExecTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	txCtx := context.WithValue(ctx, txContextKey{}, tx)

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

func extractDB(ctx context.Context, fallback *sqlx.DB) dbCtx {
	if tx, ok := ctx.Value(txContextKey{}).(*sqlx.Tx); ok && tx != nil {
		return tx
	}
	return fallback
}
