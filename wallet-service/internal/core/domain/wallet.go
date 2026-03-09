package domain

import "time"

type WalletStatus string

const (
	WalletStatusActive   WalletStatus = "ACTIVE"
	WalletStatusInactive WalletStatus = "INACTIVE"
	WalletStatusFrozen   WalletStatus = "FROZEN"
)

type Wallet struct {
	ID        string       `db:"id"         json:"id"`
	UserID    string       `db:"user_id"    json:"user_id"`
	Balance   int64        `db:"balance"    json:"balance"`
	Status    WalletStatus `db:"status"     json:"status"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
}

type CreateWalletRequest struct {
	UserID string `json:"user_id" validate:"required,uuid4"`
}

type BalanceResponse struct {
	WalletID string `json:"wallet_id"`
	UserID   string `json:"user_id"`
	Balance  int64  `json:"balance"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
}
