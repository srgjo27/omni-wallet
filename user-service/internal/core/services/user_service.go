package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/omni-wallet/user-service/internal/core/domain"
	"github.com/omni-wallet/user-service/internal/core/ports"
)

var (
	ErrEmailAlreadyExists  = errors.New("email address is already registered")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrUserNotFound        = errors.New("user not found")
	ErrPinMismatch         = errors.New("pin and confirm_pin do not match")
	ErrPinNotSet           = errors.New("transaction pin has not been set")
)

type jwtClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type UserService struct {
	userRepo         ports.UserRepository
	cacheRepo        ports.UserCacheRepository
	walletProvisioner ports.WalletProvisioner
	jwtSecret        string
	jwtTTL           time.Duration
}

func NewUserService(
	userRepo ports.UserRepository,
	cacheRepo ports.UserCacheRepository,
	walletProvisioner ports.WalletProvisioner,
	jwtSecret string,
	jwtTTL time.Duration,
) *UserService {
	return &UserService{
		userRepo:         userRepo,
		cacheRepo:        cacheRepo,
		walletProvisioner: walletProvisioner,
		jwtSecret:        jwtSecret,
		jwtTTL:           jwtTTL,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, req domain.RegisterRequest) (*domain.User, error) {
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("checking email existence: %w", err)
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	now := time.Now()
	user := &domain.User{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		KYCStatus:    domain.KYCStatusUnverified,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	created, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("persisting new user: %w", err)
	}

	if s.walletProvisioner != nil {
		if wErr := s.walletProvisioner.ProvisionWallet(ctx, created.ID); wErr != nil {
			fmt.Printf("[WARN] RegisterUser: failed to provision wallet for user_id=%s: %v\n", created.ID, wErr)
		}
	}

	return created, nil
}

func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.userRepo.ListUsers(ctx, page, pageSize)
}

func (s *UserService) Login(ctx context.Context, req domain.LoginRequest) (*domain.LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("finding user by email: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("generating JWT: %w", err)
	}

	ttlSeconds := int64(s.jwtTTL.Seconds())
	if err := s.cacheRepo.SetUserSession(ctx, user.ID, token, ttlSeconds); err != nil {
		fmt.Printf("[WARN] failed to cache user session for user_id=%s: %v\n", user.ID, err)
	}

	user.HasPIN = user.PinHash != ""

	return &domain.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   ttlSeconds,
		User:        user,
	}, nil
}

func (s *UserService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("finding user by id: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	user.HasPIN = user.PinHash != ""
	return user, nil
}

func (s *UserService) SetPin(ctx context.Context, userID string, req domain.SetPinRequest) error {
	if req.Pin != req.ConfirmPin {
		return ErrPinMismatch
	}

	pinHash, err := bcrypt.GenerateFromPassword([]byte(req.Pin), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing pin: %w", err)
	}

	if err := s.userRepo.UpdatePin(ctx, userID, string(pinHash)); err != nil {
		return fmt.Errorf("persisting pin: %w", err)
	}

	return nil
}

func (s *UserService) UpdateKYC(ctx context.Context, userID string, req domain.UpdateKYCRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("finding user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	if err := s.userRepo.UpdateKYCStatus(ctx, userID, domain.KYCStatusPending); err != nil {
		return fmt.Errorf("updating kyc status: %w", err)
	}

	return nil
}

func (s *UserService) VerifyPIN(ctx context.Context, userID string, pin string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("finding user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}
	if user.PinHash == "" {
		return ErrPinNotSet
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PinHash), []byte(pin)); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

func (s *UserService) AdminGetStats(ctx context.Context) (totalUsers, verifiedUsers int, err error) {
	return s.userRepo.GetStats(ctx)
}

func (s *UserService) LookupByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("finding user by email: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) AdminVerifyKYC(ctx context.Context, userID string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("finding user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}
	if err := s.userRepo.UpdateKYCStatus(ctx, userID, domain.KYCStatusVerified); err != nil {
		return fmt.Errorf("updating kyc status to verified: %w", err)
	}
	return nil
}

func (s *UserService) Logout(ctx context.Context, userID string) error {
	if err := s.cacheRepo.DeleteUserSession(ctx, userID); err != nil {
		return fmt.Errorf("deleting user session: %w", err)
	}
	return nil
}

func (s *UserService) VerifyToken(tokenString string) (*jwtClaims, error) {
	claims := &jwtClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid or expired token")
	}

	return claims, nil
}

func (s *UserService) generateJWT(user *domain.User) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtTTL)),
			Issuer:    "omni-wallet/user-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
