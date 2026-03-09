package domain

import "time"

type MutationDirection string

const (
	MutationDirectionCredit MutationDirection = "CREDIT"
	MutationDirectionDebit  MutationDirection = "DEBIT" 
)

type WalletMutation struct {
	ID            string            `db:"id"             json:"id"`
	WalletID      string            `db:"wallet_id"      json:"wallet_id"`
	TransactionID string            `db:"transaction_id" json:"transaction_id"`
	Direction     MutationDirection `db:"direction"      json:"direction"`
	Amount        int64             `db:"amount"         json:"amount"`      
	BalanceAfter  int64             `db:"balance_after"  json:"balance_after"`
	Description   string            `db:"description"    json:"description,omitempty"`
	CreatedAt     time.Time         `db:"created_at"     json:"created_at"`
}

type MutationListResponse struct {
	Mutations []*WalletMutation `json:"mutations"`
	Total     int               `json:"total"`
	Page      int               `json:"page"`
	Limit     int               `json:"limit"`
}

type MutationListRequest struct {
	WalletID  string `form:"wallet_id"  validate:"required,uuid4"`
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=20"`
}
