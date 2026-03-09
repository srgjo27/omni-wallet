package domain

import "time"

type KYCStatus string

const (
	KYCStatusUnverified KYCStatus = "UNVERIFIED"
	KYCStatusPending    KYCStatus = "PENDING"
	KYCStatusVerified   KYCStatus = "VERIFIED"
)

type User struct {
	ID           string    `db:"id"            json:"id"`
	Name         string    `db:"name"          json:"name"`
	Email        string    `db:"email"         json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`       
	PinHash      string    `db:"pin_hash"      json:"-"`       
	HasPIN       bool      `db:"-"             json:"has_pin"`
	KYCStatus    KYCStatus `db:"kyc_status"    json:"kyc_status"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"    json:"updated_at"`
}

type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	User        *User  `json:"user"`
}

type SetPinRequest struct {
	Pin        string `json:"pin"         validate:"required,len=6,numeric"`
	ConfirmPin string `json:"confirm_pin" validate:"required,len=6,numeric"`
}

type UpdateKYCRequest struct {
	FullName   string `json:"full_name"   validate:"required,min=2,max=150"`
	NationalID string `json:"national_id" validate:"required,len=16,numeric"`
}
