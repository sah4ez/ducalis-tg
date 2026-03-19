package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"

	"github.com/sah4ez/ducalis-tg/pkg/types"
)

// ===== AUTH SERVICE =====

type AuthService struct {
	userRepo   UserRepository
	jwtSecret  string
	jwtTTL     time.Duration
	logger     zerolog.Logger
}

func NewAuthService(
	userRepo UserRepository,
	logger zerolog.Logger,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtTTL:    24 * time.Hour,
		logger:    logger,
	}
}

func (s *AuthService) Register(ctx context.Context, req types.RegisterRequest) (*types.User, string, error) {
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return nil, "", ErrInvalidInput
	}

	// Check if user already exists
	if _, _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, "", ErrAlreadyExists
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &types.User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Name:      req.Name,
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.Create(ctx, user, string(passwordHash)); err != nil {
		s.logger.Error().Err(err).Msg("failed to create user")
		return nil, "", err
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (*types.User, string, error) {
	user, passwordHash, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, "", ErrUnauthorized
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) GetMe(ctx context.Context) (*types.User, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return nil, ErrUnauthorized
	}

	return s.userRepo.GetByID(ctx, userID)
}

func (s *AuthService) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.jwtTTL).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ===== ADMIN SERVICE =====

type AdminService struct {
	userRepo      UserRepository
	workspaceRepo WorkspaceRepository
	logger        zerolog.Logger
}

func NewAdminService(
	userRepo UserRepository,
	workspaceRepo WorkspaceRepository,
	logger zerolog.Logger,
) *AdminService {
	return &AdminService{
		userRepo:      userRepo,
		workspaceRepo: workspaceRepo,
		logger:        logger,
	}
}

func (s *AdminService) ListUsers(ctx context.Context, limit int, offset int, search string) ([]types.User, int, error) {
	// TODO: Implement with search
	return []types.User{}, 0, nil
}

func (s *AdminService) GetUser(ctx context.Context, id string) (*types.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *AdminService) BanUser(ctx context.Context, id string, reason string) error {
	// TODO: Implement ban functionality
	return nil
}

func (s *AdminService) UnbanUser(ctx context.Context, id string) error {
	// TODO: Implement unban functionality
	return nil
}

func (s *AdminService) DeleteUser(ctx context.Context, id string) error {
	// TODO: Implement with cascade delete
	return nil
}

func (s *AdminService) ListWorkspaces(ctx context.Context, limit int, offset int) ([]types.Workspace, int, error) {
	// TODO: Implement admin workspace listing
	return []types.Workspace{}, 0, nil
}

func (s *AdminService) GetWorkspace(ctx context.Context, id string) (*types.Workspace, error) {
	return s.workspaceRepo.Get(ctx, id)
}

func (s *AdminService) DeleteWorkspace(ctx context.Context, id string) error {
	return s.workspaceRepo.Delete(ctx, id)
}

func (s *AdminService) GetSystemStats(ctx context.Context) (types.SystemStats, error) {
	// TODO: Implement system stats
	return types.SystemStats{}, nil
}

// ===== INTEGRATION SERVICE =====

type IntegrationService struct {
	integrationRepo IntegrationRepository
	logger          zerolog.Logger
}

type IntegrationRepository interface {
	Create(ctx context.Context, integration *types.Integration) error
	Get(ctx context.Context, id string) (*types.Integration, error)
	List(ctx context.Context, workspaceID string) ([]types.Integration, error)
	Delete(ctx context.Context, id string) error
}

func NewIntegrationService(
	integrationRepo IntegrationRepository,
	logger zerolog.Logger,
) *IntegrationService {
	return &IntegrationService{
		integrationRepo: integrationRepo,
		logger:          logger,
	}
}

func (s *IntegrationService) CreateGitHub(ctx context.Context, req types.CreateGitHubIntegrationRequest) (*types.Integration, error) {
	config := map[string]interface{}{
		"token":        req.Token,
		"owner":        req.Owner,
		"repo":         req.Repo,
		"sync_labels":  req.SyncLabels,
		"label_filter": req.LabelFilter,
	}

	integration := &types.Integration{
		ID:          uuid.New().String(),
		WorkspaceID: req.WorkspaceID,
		Type:        "github",
		Name:        req.Name,
		Config:      config,
		AutoSync:    req.AutoSync,
		SyncStatus:  "idle",
	}

	if err := s.integrationRepo.Create(ctx, integration); err != nil {
		return nil, err
	}

	return integration, nil
}

func (s *IntegrationService) CreateJira(ctx context.Context, req types.CreateJiraIntegrationRequest) (*types.Integration, error) {
	config := map[string]interface{}{
		"base_url":     req.BaseURL,
		"username":     req.Username,
		"project_key":  req.ProjectKey,
		"jql_filter":   req.JQLFilter,
	}

	integration := &types.Integration{
		ID:          uuid.New().String(),
		WorkspaceID: req.WorkspaceID,
		Type:        "jira",
		Name:        req.Name,
		Config:      config,
		AutoSync:    req.AutoSync,
		SyncStatus:  "idle",
	}

	if err := s.integrationRepo.Create(ctx, integration); err != nil {
		return nil, err
	}

	return integration, nil
}

func (s *IntegrationService) CreateLinear(ctx context.Context, req types.CreateLinearIntegrationRequest) (*types.Integration, error) {
	config := map[string]interface{}{
		"team_id": req.TeamID,
	}

	integration := &types.Integration{
		ID:          uuid.New().String(),
		WorkspaceID: req.WorkspaceID,
		Type:        "linear",
		Name:        req.Name,
		Config:      config,
		AutoSync:    req.AutoSync,
		SyncStatus:  "idle",
	}

	if err := s.integrationRepo.Create(ctx, integration); err != nil {
		return nil, err
	}

	return integration, nil
}

func (s *IntegrationService) List(ctx context.Context, workspaceID string) ([]types.Integration, error) {
	return s.integrationRepo.List(ctx, workspaceID)
}

func (s *IntegrationService) Delete(ctx context.Context, id string) error {
	return s.integrationRepo.Delete(ctx, id)
}

func (s *IntegrationService) GetSyncStatus(ctx context.Context, id string) (types.SyncStatus, error) {
	integration, err := s.integrationRepo.Get(ctx, id)
	if err != nil {
		return types.SyncStatus{}, err
	}

	return types.SyncStatus{
		IntegrationID: integration.ID,
		Status:         integration.SyncStatus,
		LastSyncAt:     integration.LastSyncAt,
		LastError:      integration.SyncError,
	}, nil
}
