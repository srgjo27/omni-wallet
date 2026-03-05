package domain

import "time"

// TransactionType classifies the direction / purpose of a transaction.
type TransactionType string

const (
	TransactionTypeTopup   TransactionType = "TOPUP"
	TransactionTypeP2P     TransactionType = "P2P"
	TransactionTypePayment TransactionType = "PAYMENT"
)

// TransactionStatus tracks the lifecycle of a transaction.
type TransactionStatus string

const (
	TransactionStatusPending TransactionStatus = "PENDING"
	TransactionStatusSuccess TransactionStatus = "SUCCESS"
	TransactionStatusFailed  TransactionStatus = "FAILED"
)

// Transaction is the master record of a financial intent.
// It is created at the start of every operation and its status transitions
// as the operation progresses (PENDING → SUCCESS | FAILED).
type Transaction struct {
	ID              string            `db:"id"               json:"id"`
	ReferenceNo     string            `db:"reference_no"     json:"reference_no"`  // unique idempotency key
	Type            TransactionType   `db:"type"             json:"type"`
	Amount          int64             `db:"amount"           json:"amount"`
	Status          TransactionStatus `db:"status"           json:"status"`
	SourceWalletID  string            `db:"source_wallet_id"  json:"source_wallet_id,omitempty"`
	TargetWalletID  string            `db:"target_wallet_id"  json:"target_wallet_id,omitempty"`
	Description     string            `db:"description"      json:"description,omitempty"`
	CreatedAt       time.Time         `db:"created_at"       json:"created_at"`
	UpdatedAt       time.Time         `db:"updated_at"       json:"updated_at"`
}

// TopupRequest is the payload delivered by the Virtual Account webhook.
type TopupRequest struct {
	ReferenceNo string `json:"reference_no" validate:"required,max=64"`
	UserID      string `json:"user_id"      validate:"required,uuid4"`
	Amount      int64  `json:"amount"       validate:"required,gt=0"`
}

// TransferRequest describes a P2P transfer between two wallet owners.
type TransferRequest struct {
	ReferenceNo        string `json:"reference_no"          validate:"required,max=64"`
	SourceUserID       string `json:"source_user_id"        validate:"required,uuid4"`
	TargetUserID       string `json:"target_user_id"        validate:"required,uuid4"`
	Amount             int64  `json:"amount"                validate:"required,gt=0"`
	Description        string `json:"description"           validate:"max=255"`
	TransactionPIN     string `json:"transaction_pin"       validate:"required,len=6,numeric"`
}

// TransactionHistoryRequest defines the filters for mutation listing.
type TransactionHistoryRequest struct {
	WalletID  string     `form:"wallet_id"`
	StartDate *time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate   *time.Time `form:"end_date"   time_format:"2006-01-02"`
	Page      int        `form:"page,default=1"`
	Limit     int        `form:"limit,default=20"`
}

// TransactionHistoryResponse wraps a paginated list of transactions.
type TransactionHistoryResponse struct {
	Transactions []*Transaction `json:"transactions"`
	Total        int            `json:"total"`
	Page         int            `json:"page"`
	Limit        int            `json:"limit"`
}
