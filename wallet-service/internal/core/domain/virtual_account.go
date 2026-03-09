package domain

type VirtualAccount struct {
	ID            string `json:"id"`
	ExternalID    string `json:"external_id"`
	BankCode      string `json:"bank_code"`
	Name          string `json:"name"`
	AccountNumber string `json:"account_number"`
	MerchantCode  string `json:"merchant_code"`
	Currency      string `json:"currency"`
	IsClosed      bool   `json:"is_closed"`
	Status        string `json:"status"`
}

type RequestVARequest struct {
	BankCode string `json:"bank_code" validate:"required,oneof=BNI MANDIRI BRI PERMATA BSI"`
	Name     string `json:"name"      validate:"required,min=2,max=80"`
}

type XenditVAPayment struct {
	ID                       string  `json:"id"`
	PaymentID                string  `json:"payment_id"`
	CallbackVirtualAccountID string  `json:"callback_virtual_account_id"`
	ExternalID               string  `json:"external_id"`
	BankCode                 string  `json:"bank_code"`
	AccountNumber            string  `json:"account_number"`
	Amount                   float64 `json:"amount"`
	TransactionTimestamp     string  `json:"transaction_timestamp"`
}
