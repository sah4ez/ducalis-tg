// Package contract declares the service APIs as @tg-annotated Go interfaces.
// The HTTP (fiber) transport is GENERATED from these contracts by
// `tg server -o internal/transport` (tg v3, tgp-go plugins).
package contract

import (
	"context"

	"github.com/sah4ez/ducalis-tg/pkg/types"
)

// WorkspaceService manages workspaces and their configuration.
//
// @tg title=`Workspace Service`
// @tg version=`1.0.0`
// @tg desc=`Workspace CRUD, scoring configuration and members.`
// @tg http-server
// @tg metrics
type WorkspaceService interface {

	// Create creates a new workspace with custom scoring configuration.
	//
	// @tg http-method=POST
	// @tg http-path=`/api/v1/workspaces`
	// @tg http-success=201
	// @tg summary=`Create workspace`
	// @tg desc=`Creates a new workspace with custom scoring configuration`
	Create(ctx context.Context, req types.CreateWorkspaceRequest) (workspace *types.Workspace, err error)

	// Get returns workspace by ID.
	//
	// @tg http-method=GET
	// @tg http-path=`/api/v1/workspaces/:id`
	// @tg http-args=id|id|explicit
	// @tg summary=`Get workspace`
	Get(ctx context.Context, id string) (workspace *types.Workspace, err error)

	// Update updates workspace settings.
	//
	// @tg http-method=PATCH
	// @tg http-path=`/api/v1/workspaces/:id`
	// @tg http-args=id|id|explicit
	// @tg summary=`Update workspace`
	Update(ctx context.Context, id string, req types.UpdateWorkspaceRequest) (workspace *types.Workspace, err error)

	// Delete deletes workspace.
	//
	// @tg http-method=DELETE
	// @tg http-path=`/api/v1/workspaces/:id`
	// @tg http-args=id|id|explicit
	// @tg http-success=204
	// @tg summary=`Delete workspace`
	Delete(ctx context.Context, id string) (err error)

	// List returns user's workspaces.
	//
	// @tg http-method=GET
	// @tg http-path=`/api/v1/workspaces`
	// @tg http-args=userID|userID|explicit
	// @tg http-args=limit|limit|explicit
	// @tg http-args=offset|offset|explicit
	// @tg summary=`List workspaces`
	List(ctx context.Context, userID string, limit int, offset int) (workspaces []types.Workspace, total int, err error)

	// SetScoringConfig configures scoring criteria and weights.
	//
	// @tg http-method=PUT
	// @tg http-path=`/api/v1/workspaces/:id/scoring`
	// @tg http-args=id|id|explicit
	// @tg enableInlineSingle
	// @tg summary=`Set scoring configuration`
	SetScoringConfig(ctx context.Context, id string, config types.ScoringConfig) (workspace *types.Workspace, err error)

	// InviteMember invites a new member to workspace.
	//
	// @tg http-method=POST
	// @tg http-path=`/api/v1/workspaces/:id/members`
	// @tg http-args=id|id|explicit
	// @tg http-success=201
	// @tg summary=`Invite member`
	InviteMember(ctx context.Context, id string, email string, role string) (member *types.Member, err error)

	// ListMembers returns workspace members.
	//
	// @tg http-method=GET
	// @tg http-path=`/api/v1/workspaces/:id/members`
	// @tg http-args=id|id|explicit
	// @tg summary=`List members`
	ListMembers(ctx context.Context, id string) (members []types.Member, err error)

	// RemoveMember removes member from workspace.
	//
	// @tg http-method=DELETE
	// @tg http-path=`/api/v1/workspaces/:id/members/:memberID`
	// @tg http-args=id|id|explicit
	// @tg enableInlineSingle
	// @tg summary=`Remove member`
	RemoveMember(ctx context.Context, id string, memberID string) (err error)
}
