package domain

import "time"

// KYCStatus represents the KYC verification status of a user.
type KYCStatus string

const (
	KYCStatusUnverified KYCStatus = "UNVERIFIED"
	KYCStatusPending    KYCStatus = "PENDING"
	KYCStatusVerified   KYCStatus = "VERIFIED"
)

// User is the core entity representing a registered user of OmniWallet.
// This layer has NO dependency on any framework, database library, or external package.
type User struct {
	ID           string    `db:"id"            json:"id"`
	Name         string    `db:"name"          json:"name"`
	Email        string    `db:"email"         json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"` // never exposed via API
	PinHash      string    `db:"pin_hash"      json:"-"` // transaction PIN, never exposed
	KYCStatus    KYCStatus `db:"kyc_status"    json:"kyc_status"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"    json:"updated_at"`
}

// RegisterRequest holds the data required to register a new user.
type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest holds the credentials for user authentication.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse is returned upon successful authentication.
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"` // in seconds
	User        *User  `json:"user"`
}

// SetPinRequest holds the payload to set or update the transaction PIN.
type SetPinRequest struct {
	Pin        string `json:"pin"         validate:"required,len=6,numeric"`
	ConfirmPin string `json:"confirm_pin" validate:"required,len=6,numeric"`
}

// UpdateKYCRequest holds the payload for submitting KYC information.
type UpdateKYCRequest struct {
	// KYCStatus is set to PENDING when a user submits their KYC documents.
	// Full KYC document fields (NIK, photo, etc.) would be added here in a real system.
	FullName   string `json:"full_name"   validate:"required,min=2,max=150"`
	NationalID string `json:"national_id" validate:"required,len=16,numeric"`
}
