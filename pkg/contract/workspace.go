// Package contract defines service interfaces for tg code generation
// @tg title=Priora API
// @tg version=1.0.0
// @tg description=Task prioritization and scoring platform
// @tg jsonRPC-server log metrics trace
package contract

import (
	"context"

	"github.com/sah4ez/priora/pkg/types"
)

// WorkspaceService manages workspaces and their configuration
// @tg tag=workspaces Workspaces management
// @tg http-prefix=api/v1/workspaces
type WorkspaceService interface {

	// Create creates a new workspace
	// @tg summary=Create workspace
	// @tg desc=Creates a new workspace with custom scoring configuration
	// @tg http-method=POST
	// @tg http-path=/
	Create(ctx context.Context, req types.CreateWorkspaceRequest) (workspace *types.Workspace, err error)

	// Get returns workspace by ID
	// @tg summary=Get workspace
	// @tg http-method=GET
	// @tg http-path=/{id}
	// @tg http-path-params=id:id
	Get(ctx context.Context, id string) (workspace *types.Workspace, err error)

	// Update updates workspace settings
	// @tg summary=Update workspace
	// @tg http-method=PATCH
	// @tg http-path=/{id}
	// @tg http-path-params=id:id
	Update(ctx context.Context, id string, req types.UpdateWorkspaceRequest) (workspace *types.Workspace, err error)

	// Delete deletes workspace
	// @tg summary=Delete workspace
	// @tg http-method=DELETE
	// @tg http-path=/{id}
	// @tg http-path-params=id:id
	Delete(ctx context.Context, id string) (err error)

	// List returns user's workspaces
	// @tg summary=List workspaces
	// @tg http-method=GET
	// @tg http-path=/
	List(ctx context.Context, userID string, limit int, offset int) (workspaces []types.Workspace, total int, err error)

	// SetScoringConfig configures scoring criteria and weights
	// @tg summary=Set scoring configuration
	// @tg http-method=PUT
	// @tg http-path=/{id}/scoring
	// @tg http-path-params=id:workspaceID
	SetScoringConfig(ctx context.Context, workspaceID string, config types.ScoringConfig) (workspace *types.Workspace, err error)

	// InviteMember invites a new member to workspace
	// @tg summary=Invite member
	// @tg http-method=POST
	// @tg http-path=/{id}/members
	// @tg http-path-params=id:workspaceID
	InviteMember(ctx context.Context, workspaceID string, email string, role string) (member *types.Member, err error)

	// ListMembers returns workspace members
	// @tg summary=List members
	// @tg http-method=GET
	// @tg http-path=/{id}/members
	// @tg http-path-params=id:workspaceID
	ListMembers(ctx context.Context, workspaceID string) (members []types.Member, err error)

	// RemoveMember removes member from workspace
	// @tg summary=Remove member
	// @tg http-method=DELETE
	// @tg http-path=/{id}/members/{memberID}
	// @tg http-path-params=id:workspaceID,memberID:memberID
	RemoveMember(ctx context.Context, workspaceID string, memberID string) (err error)
}
