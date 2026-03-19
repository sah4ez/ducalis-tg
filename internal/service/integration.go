package service

import (
	"context"
	"errors"

	"github.com/rs/zerolog"

	"github.com/sah4ez/ducalis-tg/pkg/types"
)

// IntegrationService implements contract.IntegrationService
type IntegrationService struct {
	store  IntegrationStore
	logger zerolog.Logger
}

// IntegrationStore defines storage interface for integrations
type IntegrationStore interface {
	Create(ctx context.Context, integration *types.Integration) error
	Get(ctx context.Context, id string) (*types.Integration, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, workspaceID string) ([]types.Integration, error)
	UpdateSyncStatus(ctx context.Context, id, status, errMsg string) error
}

// NewIntegrationService creates new integration service
func NewIntegrationService(store IntegrationStore, logger zerolog.Logger) *IntegrationService {
	return &IntegrationService{
		store:  store,
		logger: logger,
	}
}

// CreateGitHub creates GitHub integration
func (s *IntegrationService) CreateGitHub(ctx context.Context, req types.CreateGitHubIntegrationRequest) (*types.Integration, error) {
	return nil, errors.New("not implemented")
}

// CreateJira creates Jira integration
func (s *IntegrationService) CreateJira(ctx context.Context, req types.CreateJiraIntegrationRequest) (*types.Integration, error) {
	return nil, errors.New("not implemented")
}

// CreateLinear creates Linear integration
func (s *IntegrationService) CreateLinear(ctx context.Context, req types.CreateLinearIntegrationRequest) (*types.Integration, error) {
	return nil, errors.New("not implemented")
}

// List returns workspace integrations
func (s *IntegrationService) List(ctx context.Context, workspaceID string) ([]types.Integration, error) {
	return s.store.List(ctx, workspaceID)
}

// Delete deletes integration
func (s *IntegrationService) Delete(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}

// Sync triggers manual sync with external service
func (s *IntegrationService) Sync(ctx context.Context, id string) (types.SyncResult, error) {
	return types.SyncResult{}, errors.New("not implemented")
}

// GetSyncStatus returns last sync status
func (s *IntegrationService) GetSyncStatus(ctx context.Context, id string) (types.SyncStatus, error) {
	integration, err := s.store.Get(ctx, id)
	if err != nil {
		return types.SyncStatus{}, err
	}

	return types.SyncStatus{
		IntegrationID:  integration.ID,
		Status:         integration.SyncStatus,
		LastSyncAt:     integration.LastSyncAt,
		LastError:      integration.SyncError,
	}, nil
}
