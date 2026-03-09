package ports

import "context"

type WalletProvisioner interface {
	ProvisionWallet(ctx context.Context, userID string) error
}
