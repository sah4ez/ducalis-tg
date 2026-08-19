package adapter

import (
	"context"

	"github.com/sah4ez/ducalis-tg/internal/service"
	"github.com/sah4ez/ducalis-tg/pkg/types"
)

// WorkspaceAdapter implements contract.WorkspaceService by wrapping service.WorkspaceService
type WorkspaceAdapter struct {
	svc *service.WorkspaceService
}

func NewWorkspaceAdapter(svc *service.WorkspaceService) *WorkspaceAdapter {
	return &WorkspaceAdapter{svc: svc}
}

func (a *WorkspaceAdapter) Create(ctx context.Context, req types.CreateWorkspaceRequest) (*types.Workspace, error) {
	return a.svc.Create(ctx, req)
}

func (a *WorkspaceAdapter) Get(ctx context.Context, id string) (*types.Workspace, error) {
	return a.svc.Get(ctx, id)
}

func (a *WorkspaceAdapter) Update(ctx context.Context, id string, req types.UpdateWorkspaceRequest) (*types.Workspace, error) {
	return a.svc.Update(ctx, id, req)
}

func (a *WorkspaceAdapter) Delete(ctx context.Context, id string) error {
	return a.svc.Delete(ctx, id)
}

func (a *WorkspaceAdapter) List(ctx context.Context, userID string, limit int, offset int) ([]types.Workspace, int, error) {
	// Contract passes userID explicitly; service reads from context.
	ctx = context.WithValue(ctx, service.UserIDKey, userID)
	return a.svc.List(ctx, userID, limit, offset)
}

func (a *WorkspaceAdapter) SetScoringConfig(ctx context.Context, id string, config types.ScoringConfig) (*types.Workspace, error) {
	return a.svc.SetScoringConfig(ctx, id, config)
}

func (a *WorkspaceAdapter) InviteMember(ctx context.Context, id string, email string, role string) (*types.Member, error) {
	return a.svc.InviteMember(ctx, id, email, role)
}

func (a *WorkspaceAdapter) ListMembers(ctx context.Context, id string) ([]types.Member, error) {
	return a.svc.ListMembers(ctx, id)
}

func (a *WorkspaceAdapter) RemoveMember(ctx context.Context, id string, memberID string) error {
	return a.svc.RemoveMember(ctx, id, memberID)
}
