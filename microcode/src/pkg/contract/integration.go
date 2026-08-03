package contract

import (
	"context"

	"github.com/sah4ez/ducalis-tg/pkg/types"
)

// IntegrationService manages external service integrations
// @tg tag=integrations Integrations management
// @tg http-prefix=api/v1/integrations
type IntegrationService interface {

	// CreateGitHub creates GitHub integration
	// @tg summary=Create GitHub integration
	// @tg http-method=POST
	// @tg http-path=/github
	CreateGitHub(ctx context.Context, req types.CreateGitHubIntegrationRequest) (integration *types.Integration, err error)

	// CreateJira creates Jira integration
	// @tg summary=Create Jira integration
	// @tg http-method=POST
	// @tg http-path=/jira
	CreateJira(ctx context.Context, req types.CreateJiraIntegrationRequest) (integration *types.Integration, err error)

	// CreateLinear creates Linear integration
	// @tg summary=Create Linear integration
	// @tg http-method=POST
	// @tg http-path=/linear
	CreateLinear(ctx context.Context, req types.CreateLinearIntegrationRequest) (integration *types.Integration, err error)

	// List returns workspace integrations
	// @tg summary=List integrations
	// @tg http-method=GET
	// @tg http-path=/
	List(ctx context.Context, workspaceID string) (integrations []types.Integration, err error)

	// Delete deletes integration
	// @tg summary=Delete integration
	// @tg http-method=DELETE
	// @tg http-path=/{id}
	// @tg http-path-params=id:id
	Delete(ctx context.Context, id string) (err error)

	// Sync triggers manual sync with external service
	// @tg summary=Sync integration
	// @tg http-method=POST
	// @tg http-path=/{id}/sync
	// @tg http-path-params=id:id
	Sync(ctx context.Context, id string) (result types.SyncResult, err error)

	// GetSyncStatus returns last sync status
	// @tg summary=Get sync status
	// @tg http-method=GET
	// @tg http-path=/{id}/status
	// @tg http-path-params=id:id
	GetSyncStatus(ctx context.Context, id string) (status types.SyncStatus, err error)
}

// AuthService handles authentication
// @tg tag=auth Authentication
// @tg http-prefix=api/v1/auth
type AuthService interface {

	// Register creates new user account
	// @tg summary=Register user
	// @tg http-method=POST
	// @tg http-path=/register
	Register(ctx context.Context, req types.RegisterRequest) (user *types.User, token string, err error)

	// Login authenticates user
	// @tg summary=Login
	// @tg http-method=POST
	// @tg http-path=/login
	Login(ctx context.Context, email string, password string) (user *types.User, token string, err error)

	// OAuthAuthorize initiates OAuth flow
	// @tg summary=OAuth authorize
	// @tg http-method=GET
	// @tg http-path=/oauth/{provider}
	// @tg http-path-params=provider:provider
	OAuthAuthorize(ctx context.Context, provider string, redirectURL string) (authURL string, err error)

	// OAuthCallback handles OAuth callback
	// @tg summary=OAuth callback
	// @tg http-method=GET
	// @tg http-path=/oauth/{provider}/callback
	// @tg http-path-params=provider:provider
	OAuthCallback(ctx context.Context, provider string, code string, state string) (user *types.User, token string, err error)

	// GetMe returns current user
	// @tg summary=Get current user
	// @tg http-method=GET
	// @tg http-path=/me
	GetMe(ctx context.Context, userID string) (user *types.User, err error)

	// RefreshToken refreshes JWT token
	// @tg summary=Refresh token
	// @tg http-method=POST
	// @tg http-path=/refresh
	RefreshToken(ctx context.Context, refreshToken string) (token string, err error)
}
