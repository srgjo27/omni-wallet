package ports

// UserServiceClient defines the contract for calling the User Service
// to verify user existence and validate transaction PINs.
// In a real microservice topology this would be an HTTP/gRPC client;
// for now it surfaces only what the Wallet Service needs.
type UserServiceClient interface {
	// VerifyPIN returns nil if the provided pin matches the stored hash for userID.
	VerifyPIN(userID string, pin string) error

	// ExistsByID returns true if a user with the given ID exists.
	ExistsByID(userID string) (bool, error)
}
