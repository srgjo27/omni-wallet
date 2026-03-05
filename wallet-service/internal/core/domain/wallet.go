package domain

import "time"

// WalletStatus represents the operational state of a wallet.
type WalletStatus string

const (
	WalletStatusActive   WalletStatus = "ACTIVE"
	WalletStatusInactive WalletStatus = "INACTIVE"
	WalletStatusFrozen   WalletStatus = "FROZEN"
)

// Wallet stores the current balance snapshot for a user.
// It is the mutable "account" record; all balance history lives in WalletMutation.
type Wallet struct {
	ID        string       `db:"id"         json:"id"`
	UserID    string       `db:"user_id"    json:"user_id"`
	Balance   int64        `db:"balance"    json:"balance"` // stored in smallest unit (e.g. cents / rupiah)
	Status    WalletStatus `db:"status"     json:"status"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
}

// CreateWalletRequest holds the data required to open a new wallet.
type CreateWalletRequest struct {
	UserID string `json:"user_id" validate:"required,uuid4"`
}

// BalanceResponse is the payload returned for a balance inquiry.
type BalanceResponse struct {
	WalletID string `json:"wallet_id"`
	UserID   string `json:"user_id"`
	Balance  int64  `json:"balance"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
}
