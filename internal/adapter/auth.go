package adapter

import (
	"context"
	"errors"

	"github.com/sah4ez/ducalis-tg/internal/service"
	"github.com/sah4ez/ducalis-tg/pkg/types"
)

// AuthAdapter implements contract.AuthService by wrapping service.AuthService
type AuthAdapter struct {
	svc *service.AuthService
}

func NewAuthAdapter(svc *service.AuthService) *AuthAdapter {
	return &AuthAdapter{svc: svc}
}

func (a *AuthAdapter) Register(ctx context.Context, req types.RegisterRequest) (*types.User, string, error) {
	return a.svc.Register(ctx, req)
}

func (a *AuthAdapter) Login(ctx context.Context, email string, password string) (*types.User, string, error) {
	return a.svc.Login(ctx, email, password)
}

func (a *AuthAdapter) GetMe(ctx context.Context, userID string) (*types.User, error) {
	// Contract passes userID explicitly; service reads from context.
	// Inject userID into context so service.GetMe works.
	ctx = context.WithValue(ctx, service.UserIDKey, userID)
	return a.svc.GetMe(ctx)
}

// OAuthAuthorize is not yet implemented in the service layer.
func (a *AuthAdapter) OAuthAuthorize(ctx context.Context, provider string, redirectURL string) (string, error) {
	return "", errors.New("not implemented")
}

// OAuthCallback is not yet implemented in the service layer.
func (a *AuthAdapter) OAuthCallback(ctx context.Context, provider string, code string, state string) (*types.User, string, error) {
	return nil, "", errors.New("not implemented")
}

// RefreshToken is not yet implemented in the service layer.
func (a *AuthAdapter) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	return "", errors.New("not implemented")
}
