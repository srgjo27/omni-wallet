package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/omni-wallet/user-service/internal/core/domain"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `
		INSERT INTO users (id, name, email, password_hash, pin_hash, kyc_status, created_at, updated_at)
		VALUES (:id, :name, :email, :password_hash, :pin_hash, :kyc_status, :created_at, :updated_at)
	`
	if _, err := r.db.NamedExecContext(ctx, query, user); err != nil {
		return nil, fmt.Errorf("inserting user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, name, email, password_hash, pin_hash, kyc_status, created_at, updated_at
		FROM users WHERE id = ? LIMIT 1
	`
	var user domain.User
	if err := r.db.GetContext(ctx, &user, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying user by id: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, name, email, password_hash, pin_hash, kyc_status, created_at, updated_at
		FROM users WHERE email = ? LIMIT 1
	`
	var user domain.User
	if err := r.db.GetContext(ctx, &user, query, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) UpdatePin(ctx context.Context, userID string, pinHash string) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE users SET pin_hash = ?, updated_at = NOW() WHERE id = ?`,
		pinHash, userID,
	)
	if err != nil {
		return fmt.Errorf("updating pin: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected after pin update: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("no user found with id %s", userID)
	}
	return nil
}

func (r *UserRepository) UpdateKYCStatus(ctx context.Context, userID string, status domain.KYCStatus) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE users SET kyc_status = ?, updated_at = NOW() WHERE id = ?`,
		status, userID,
	)
	if err != nil {
		return fmt.Errorf("updating kyc status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected after kyc update: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("no user found with id %s", userID)
	}
	return nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int
	if err := r.db.GetContext(ctx, &count, `SELECT COUNT(1) FROM users WHERE email = ?`, email); err != nil {
		return false, fmt.Errorf("checking email existence: %w", err)
	}
	return count > 0, nil
}

func (r *UserRepository) ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int, error) {
	offset := (page - 1) * pageSize

	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(1) FROM users`); err != nil {
		return nil, 0, fmt.Errorf("counting users: %w", err)
	}

	var users []*domain.User
	query := `
		SELECT id, name, email, password_hash, pin_hash, kyc_status, created_at, updated_at
		FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?
	`
	if err := r.db.SelectContext(ctx, &users, query, pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("listing users: %w", err)
	}

	return users, total, nil
}

func (r *UserRepository) GetStats(ctx context.Context) (totalUsers, verifiedUsers int, err error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			SUM(CASE WHEN kyc_status = 'VERIFIED' THEN 1 ELSE 0 END)
		FROM users
	`)
	var total, verified int
	if err := row.Scan(&total, &verified); err != nil {
		return 0, 0, fmt.Errorf("querying user stats: %w", err)
	}
	return total, verified, nil
}
