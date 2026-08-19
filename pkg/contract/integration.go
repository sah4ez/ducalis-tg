// Package contract declares the service APIs as @tg-annotated Go interfaces.
// The HTTP (fiber) transport is GENERATED from these contracts by
// `tg server -o internal/transport` (tg v3, tgp-go plugins).
package contract

import (
	"context"

	"github.com/sah4ez/ducalis-tg/pkg/types"
)

// IntegrationService manages external service integrations.
//
// @tg title=`Integration Service`
// @tg version=`1.0.0`
// @tg desc=`GitHub/Jira/Linear integrations and sync.`
// @tg http-server
// @tg metrics
type IntegrationService interface {

	// CreateGitHub creates GitHub integration.
	//
	// @tg http-method=POST
	// @tg http-path=`/api/v1/integrations/github`
	// @tg http-success=201
	// @tg summary=`Create GitHub integration`
	CreateGitHub(ctx context.Context, req types.CreateGitHubIntegrationRequest) (integration *types.Integration, err error)

	// CreateJira creates Jira integration.
	//
	// @tg http-method=POST
	// @tg http-path=`/api/v1/integrations/jira`
	// @tg http-success=201
	// @tg summary=`Create Jira integration`
	CreateJira(ctx context.Context, req types.CreateJiraIntegrationRequest) (integration *types.Integration, err error)

	// CreateLinear creates Linear integration.
	//
	// @tg http-method=POST
	// @tg http-path=`/api/v1/integrations/linear`
	// @tg http-success=201
	// @tg summary=`Create Linear integration`
	CreateLinear(ctx context.Context, req types.CreateLinearIntegrationRequest) (integration *types.Integration, err error)

	// List returns workspace integrations.
	//
	// @tg http-method=GET
	// @tg http-path=`/api/v1/workspaces/:workspaceID/integrations`
	// @tg http-args=workspaceID|workspaceID|explicit
	// @tg summary=`List integrations`
	List(ctx context.Context, workspaceID string) (integrations []types.Integration, err error)

	// Delete deletes integration.
	//
	// @tg http-method=DELETE
	// @tg http-path=`/api/v1/integrations/:id`
	// @tg http-args=id|id|explicit
	// @tg http-success=204
	// @tg summary=`Delete integration`
	Delete(ctx context.Context, id string) (err error)

	// Sync triggers manual sync with external service.
	//
	// @tg http-method=POST
	// @tg http-path=`/api/v1/integrations/:id/sync`
	// @tg http-args=id|id|explicit
	// @tg summary=`Sync integration`
	Sync(ctx context.Context, id string) (result types.SyncResult, err error)

	// GetSyncStatus returns last sync status.
	//
	// @tg http-method=GET
	// @tg http-path=`/api/v1/integrations/:id/status`
	// @tg http-args=id|id|explicit
	// @tg summary=`Get sync status`
	GetSyncStatus(ctx context.Context, id string) (status types.SyncStatus, err error)
}

// AuthService handles authentication.
//
// @tg title=`Auth Service`
// @tg version=`1.0.0`
// @tg desc=`Registration, login and JWT tokens.`
// @tg http-server
// @tg metrics
type AuthService interface {

	// Register creates new user account.
	//
	// @tg http-method=POST
	// @tg http-path=`/api/v1/auth/register`
	// @tg http-success=201
	// @tg summary=`Register user`
	Register(ctx context.Context, req types.RegisterRequest) (user *types.User, token string, err error)

	// Login authenticates user.
	//
	// @tg http-method=POST
	// @tg http-path=`/api/v1/auth/login`
	// @tg summary=`Login`
	Login(ctx context.Context, email string, password string) (user *types.User, token string, err error)

	// OAuthAuthorize initiates OAuth flow.
	//
	// @tg http-method=GET
	// @tg http-path=`/api/v1/auth/oauth/:provider`
	// @tg http-args=provider|provider|explicit
	// @tg summary=`OAuth authorize`
	OAuthAuthorize(ctx context.Context, provider string, redirectURL string) (authURL string, err error)

	// OAuthCallback handles OAuth callback.
	//
	// @tg http-method=GET
	// @tg http-path=`/api/v1/auth/oauth/:provider/callback`
	// @tg http-args=provider|provider|explicit
	// @tg summary=`OAuth callback`
	OAuthCallback(ctx context.Context, provider string, code string, state string) (user *types.User, token string, err error)

	// GetMe returns current user.
	//
	// @tg http-method=GET
	// @tg http-path=`/api/v1/auth/me`
	// @tg summary=`Get current user`
	GetMe(ctx context.Context, userID string) (user *types.User, err error)

	// RefreshToken refreshes JWT token.
	//
	// @tg http-method=POST
	// @tg http-path=`/api/v1/auth/refresh`
	// @tg summary=`Refresh token`
	RefreshToken(ctx context.Context, refreshToken string) (token string, err error)
}
