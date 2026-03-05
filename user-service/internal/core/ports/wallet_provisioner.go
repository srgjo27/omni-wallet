package ports

import "context"

// WalletProvisioner defines the contract for creating a wallet for a newly
// registered user. This keeps the core service layer decoupled from the HTTP
// adapter that actually calls the wallet-service.
type WalletProvisioner interface {
	// ProvisionWallet creates an empty wallet for the given user ID.
	// It is called synchronously after a successful user registration.
	ProvisionWallet(ctx context.Context, userID string) error
}
