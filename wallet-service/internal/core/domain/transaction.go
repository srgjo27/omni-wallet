package domain

import "time"

type TransactionType string

const (
	TransactionTypeTopup   TransactionType = "TOPUP"
	TransactionTypeP2P     TransactionType = "P2P"
	TransactionTypePayment TransactionType = "PAYMENT"
)

type TransactionStatus string

const (
	TransactionStatusPending TransactionStatus = "PENDING"
	TransactionStatusSuccess TransactionStatus = "SUCCESS"
	TransactionStatusFailed  TransactionStatus = "FAILED"
)

type Transaction struct {
	ID              string            `db:"id"               json:"id"`
	ReferenceNo     string            `db:"reference_no"     json:"reference_no"`
	Type            TransactionType   `db:"type"             json:"type"`
	Amount          int64             `db:"amount"           json:"amount"`
	Status          TransactionStatus `db:"status"           json:"status"`
	SourceWalletID  string            `db:"source_wallet_id"  json:"source_wallet_id,omitempty"`
	TargetWalletID  string            `db:"target_wallet_id"  json:"target_wallet_id,omitempty"`
	Description     string            `db:"description"      json:"description,omitempty"`
	CreatedAt       time.Time         `db:"created_at"       json:"created_at"`
	UpdatedAt       time.Time         `db:"updated_at"       json:"updated_at"`
}

type TopupRequest struct {
	ReferenceNo string `json:"reference_no" validate:"required,max=64"`
	UserID      string `json:"user_id"      validate:"required,uuid4"`
	Amount      int64  `json:"amount"       validate:"required,gt=0"`
}

type TransferRequest struct {
	ReferenceNo    string `json:"reference_no"    validate:"required,max=64"`
	SourceUserID   string `json:"source_user_id"  validate:"required,uuid4"`
	TargetEmail    string `json:"target_email"    validate:"required,email"`
	TargetUserID   string `json:"-"`
	Amount         int64  `json:"amount"          validate:"required,gt=0"`
	Description    string `json:"description"     validate:"max=255"`
	TransactionPIN string `json:"transaction_pin" validate:"required,len=6,numeric"`
}

type TransactionHistoryRequest struct {
	WalletID  string     `form:"wallet_id"`
	StartDate *time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate   *time.Time `form:"end_date"   time_format:"2006-01-02"`
	Page      int        `form:"page,default=1"`
	Limit     int        `form:"limit,default=20"`
}

type TransactionHistoryResponse struct {
	Transactions []*Transaction `json:"transactions"`
	Total        int            `json:"total"`
	Page         int            `json:"page"`
	Limit        int            `json:"limit"`
}
