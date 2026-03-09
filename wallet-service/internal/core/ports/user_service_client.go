package ports

type UserServiceClient interface {
	VerifyPIN(userID string, pin string) error
	ExistsByID(userID string) (bool, error)
	FindUserIDByEmail(email string) (string, error)
}
