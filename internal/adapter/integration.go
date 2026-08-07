package adapter

import (
	"context"
	"errors"

	"github.com/sah4ez/ducalis-tg/internal/service"
	"github.com/sah4ez/ducalis-tg/pkg/types"
)

// IntegrationAdapter implements contract.IntegrationService by wrapping service.IntegrationService
type IntegrationAdapter struct {
	svc *service.IntegrationService
}

func NewIntegrationAdapter(svc *service.IntegrationService) *IntegrationAdapter {
	return &IntegrationAdapter{svc: svc}
}

func (a *IntegrationAdapter) CreateGitHub(ctx context.Context, req types.CreateGitHubIntegrationRequest) (*types.Integration, error) {
	return a.svc.CreateGitHub(ctx, req)
}

func (a *IntegrationAdapter) CreateJira(ctx context.Context, req types.CreateJiraIntegrationRequest) (*types.Integration, error) {
	return a.svc.CreateJira(ctx, req)
}

func (a *IntegrationAdapter) CreateLinear(ctx context.Context, req types.CreateLinearIntegrationRequest) (*types.Integration, error) {
	return a.svc.CreateLinear(ctx, req)
}

func (a *IntegrationAdapter) List(ctx context.Context, workspaceID string) ([]types.Integration, error) {
	return a.svc.List(ctx, workspaceID)
}

func (a *IntegrationAdapter) Delete(ctx context.Context, id string) error {
	return a.svc.Delete(ctx, id)
}

// Sync is not yet implemented in the service layer.
func (a *IntegrationAdapter) Sync(ctx context.Context, id string) (types.SyncResult, error) {
	return types.SyncResult{}, errors.New("not implemented")
}

func (a *IntegrationAdapter) GetSyncStatus(ctx context.Context, id string) (types.SyncStatus, error) {
	return a.svc.GetSyncStatus(ctx, id)
}
