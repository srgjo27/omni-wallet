package domain

import "time"

// MutationDirection indicates whether the mutation increased or decreased the balance.
type MutationDirection string

const (
	MutationDirectionCredit MutationDirection = "CREDIT" // balance increased (+)
	MutationDirectionDebit  MutationDirection = "DEBIT"  // balance decreased (-)
)

// WalletMutation is a single immutable entry in the double-entry ledger.
// Every balance change (top-up, transfer out, transfer in) MUST produce a mutation record.
// The balance_after field provides a point-in-time snapshot for reconciliation.
type WalletMutation struct {
	ID            string            `db:"id"             json:"id"`
	WalletID      string            `db:"wallet_id"      json:"wallet_id"`
	TransactionID string            `db:"transaction_id" json:"transaction_id"`
	Direction     MutationDirection `db:"direction"      json:"direction"`
	Amount        int64             `db:"amount"         json:"amount"`        // always positive; direction signals sign
	BalanceAfter  int64             `db:"balance_after"  json:"balance_after"` // snapshot of the wallet balance after this entry
	Description   string            `db:"description"    json:"description,omitempty"`
	CreatedAt     time.Time         `db:"created_at"     json:"created_at"`
}

// MutationListResponse wraps a paginated list of wallet mutations.
type MutationListResponse struct {
	Mutations []*WalletMutation `json:"mutations"`
	Total     int               `json:"total"`
	Page      int               `json:"page"`
	Limit     int               `json:"limit"`
}

// MutationListRequest holds filters for querying wallet mutations.
type MutationListRequest struct {
	WalletID  string `form:"wallet_id"  validate:"required,uuid4"`
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=20"`
}
