package adapter

import (
	"context"
	"testing"

	"github.com/sah4ez/ducalis-tg/internal/service"
	"github.com/sah4ez/ducalis-tg/pkg/types"
	"github.com/rs/zerolog"
)

func TestIntegrationAdapter_CreateGitHub(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Logger()
	adapter := NewIntegrationAdapter(service.NewIntegrationService(
		newStubIntegrationRepo(), logger))

	integration, err := adapter.CreateGitHub(context.Background(), types.CreateGitHubIntegrationRequest{
		WorkspaceID: "ws-1",
		Name:        "GitHub Integration",
		Token:       "ghp_test",
		Owner:       "org",
		Repo:        "repo",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if integration.Type != "github" {
		t.Errorf("expected type github, got %s", integration.Type)
	}
}

func TestIntegrationAdapter_List(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Logger()
	repo := newStubIntegrationRepo()
	adapter := NewIntegrationAdapter(service.NewIntegrationService(repo, logger))

	_, _ = adapter.CreateGitHub(context.Background(), types.CreateGitHubIntegrationRequest{
		WorkspaceID: "ws-1",
		Name:        "Test",
		Token:       "tok",
		Owner:       "o",
		Repo:        "r",
	})

	integrations, err := adapter.List(context.Background(), "ws-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(integrations) != 1 {
		t.Errorf("expected 1 integration, got %d", len(integrations))
	}
}

func TestIntegrationAdapter_Sync_NotImplemented(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Logger()
	adapter := NewIntegrationAdapter(service.NewIntegrationService(
		newStubIntegrationRepo(), logger))

	_, err := adapter.Sync(context.Background(), "int-1")
	if err == nil {
		t.Error("expected error for unimplemented Sync")
	}
}

// stub integration repo
type stubIntegrationRepo struct {
	integrations map[string]*types.Integration
}

func newStubIntegrationRepo() *stubIntegrationRepo {
	return &stubIntegrationRepo{integrations: make(map[string]*types.Integration)}
}

func (r *stubIntegrationRepo) Create(ctx context.Context, i *types.Integration) error {
	r.integrations[i.ID] = i
	return nil
}

func (r *stubIntegrationRepo) Get(ctx context.Context, id string) (*types.Integration, error) {
	if i, ok := r.integrations[id]; ok {
		return i, nil
	}
	return nil, service.ErrNotFound
}

func (r *stubIntegrationRepo) List(ctx context.Context, workspaceID string) ([]types.Integration, error) {
	var result []types.Integration
	for _, i := range r.integrations {
		if i.WorkspaceID == workspaceID {
			result = append(result, *i)
		}
	}
	return result, nil
}

func (r *stubIntegrationRepo) Delete(ctx context.Context, id string) error {
	delete(r.integrations, id)
	return nil
}
