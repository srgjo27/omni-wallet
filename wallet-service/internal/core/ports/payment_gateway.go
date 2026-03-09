package ports

import "github.com/omni-wallet/wallet-service/internal/core/domain"

type PaymentGateway interface {
	CreateFixedVA(externalID, name, bankCode string) (*domain.VirtualAccount, error)
	VerifyWebhookToken(token string) bool
}
