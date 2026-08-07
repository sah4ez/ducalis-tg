package adapter

import (
	"context"
	"errors"
	"testing"

	"github.com/sah4ez/ducalis-tg/internal/service"
	"github.com/sah4ez/ducalis-tg/pkg/types"
	"github.com/rs/zerolog"
)

func TestAuthAdapter_Register(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Logger()
	svc := service.NewAuthService(newStubUserRepo(), logger, "test-secret")
	adapter := NewAuthAdapter(svc)

	user, token, err := adapter.Register(context.Background(), types.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestAuthAdapter_Login(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Logger()
	repo := newStubUserRepo()
	svc := service.NewAuthService(repo, logger, "test-secret")
	adapter := NewAuthAdapter(svc)

	// Register first
	_, _, err := adapter.Register(context.Background(), types.RegisterRequest{
		Name:     "Login Test",
		Email:    "login@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Login
	user, token, err := adapter.Login(context.Background(), "login@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestAuthAdapter_Login_WrongPassword(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Logger()
	svc := service.NewAuthService(newStubUserRepo(), logger, "test-secret")
	adapter := NewAuthAdapter(svc)

	_, _, err := adapter.Login(context.Background(), "nonexistent@example.com", "wrong")
	if !errors.Is(err, service.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAuthAdapter_GetMe(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Logger()
	repo := newStubUserRepo()
	svc := service.NewAuthService(repo, logger, "test-secret")
	adapter := NewAuthAdapter(svc)

	user, _, _ := adapter.Register(context.Background(), types.RegisterRequest{
		Name:     "Me Test",
		Email:    "me@example.com",
		Password: "password123",
	})

	retrieved, err := adapter.GetMe(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.Email != "me@example.com" {
		t.Errorf("expected email me@example.com, got %s", retrieved.Email)
	}
}

func TestAuthAdapter_OAuthNotImplemented(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Logger()
	svc := service.NewAuthService(newStubUserRepo(), logger, "test-secret")
	adapter := NewAuthAdapter(svc)

	_, err := adapter.OAuthAuthorize(context.Background(), "github", "http://callback")
	if err == nil {
		t.Error("expected error for unimplemented OAuthAuthorize")
	}
}

func TestAuthAdapter_RefreshTokenNotImplemented(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Logger()
	svc := service.NewAuthService(newStubUserRepo(), logger, "test-secret")
	adapter := NewAuthAdapter(svc)

	_, err := adapter.RefreshToken(context.Background(), "some-token")
	if err == nil {
		t.Error("expected error for unimplemented RefreshToken")
	}
}

// stubUserRepo is a simple in-memory user repository for testing
type stubUserRepo struct {
	users map[string]*types.User
	hashes map[string]string
}

func newStubUserRepo() *stubUserRepo {
	return &stubUserRepo{
		users:  make(map[string]*types.User),
		hashes: make(map[string]string),
	}
}

func (r *stubUserRepo) Create(ctx context.Context, user *types.User, passwordHash string) error {
	r.users[user.ID] = user
	r.hashes[user.Email] = passwordHash
	return nil
}

func (r *stubUserRepo) GetByID(ctx context.Context, id string) (*types.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, errors.New("not found")
}

func (r *stubUserRepo) GetByEmail(ctx context.Context, email string) (*types.User, string, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, r.hashes[email], nil
		}
	}
	return nil, "", errors.New("not found")
}
