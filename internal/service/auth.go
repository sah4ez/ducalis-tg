package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"

	"github.com/sah4ez/priora/pkg/types"
)

// AuthService implements contract.AuthService
type AuthService struct {
	store     AuthStore
	logger    zerolog.Logger
	jwtSecret string
}

// AuthStore defines storage interface for auth
type AuthStore interface {
	CreateUser(ctx context.Context, user *types.User, passwordHash string) error
	GetUserByEmail(ctx context.Context, email string) (*types.User, string, error)
	GetUserByID(ctx context.Context, id string) (*types.User, error)
}

// NewAuthService creates new auth service
func NewAuthService(store AuthStore, logger zerolog.Logger, jwtSecret string) *AuthService {
	return &AuthService{
		store:     store,
		logger:    logger,
		jwtSecret: jwtSecret,
	}
}

// Register creates new user account
func (s *AuthService) Register(ctx context.Context, req types.RegisterRequest) (*types.User, string, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &types.User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Name:      req.Name,
		CreatedAt: time.Now(),
	}

	if err := s.store.CreateUser(ctx, user, string(hashedPassword)); err != nil {
		return nil, "", err
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Login authenticates user
func (s *AuthService) Login(ctx context.Context, email string, password string) (*types.User, string, error) {
	user, passwordHash, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// OAuthAuthorize initiates OAuth flow
func (s *AuthService) OAuthAuthorize(ctx context.Context, provider string, redirectURL string) (string, error) {
	return "", errors.New("not implemented")
}

// OAuthCallback handles OAuth callback
func (s *AuthService) OAuthCallback(ctx context.Context, provider string, code string, state string) (*types.User, string, error) {
	return nil, "", errors.New("not implemented")
}

// GetMe returns current user
func (s *AuthService) GetMe(ctx context.Context, userID string) (*types.User, error) {
	return s.store.GetUserByID(ctx, userID)
}

// RefreshToken refreshes JWT token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	return "", errors.New("not implemented")
}

// generateToken generates JWT token for user
func (s *AuthService) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
