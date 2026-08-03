package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/sah4ez/ducalis-tg/pkg/types"
)

// Common errors
var (
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrInvalidInput     = errors.New("invalid input")
	ErrNotFound         = errors.New("not found")
	ErrAlreadyExists    = errors.New("already exists")
	ErrWorkspaceLimit   = errors.New("workspace limit reached")
	ErrMemberLimit      = errors.New("member limit reached")
	ErrInvalidScoring   = errors.New("invalid scoring config")
)

// ContextKey type for context values
type ContextKey string

const (
	UserIDKey    ContextKey = "userID"
	AdminIDKey   ContextKey = "adminID"
	WorkspaceKey ContextKey = "workspace"
)

// ===== WORKSPACE SERVICE =====

type WorkspaceService struct {
	workspaceRepo WorkspaceRepository
	memberRepo    MemberRepository
	userRepo      UserRepository
	logger        zerolog.Logger
}

type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *types.Workspace) error
	Get(ctx context.Context, id string) (*types.Workspace, error)
	Update(ctx context.Context, workspace *types.Workspace) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, userID string, limit, offset int) ([]types.Workspace, int, error)
}

type MemberRepository interface {
	Add(ctx context.Context, member *types.Member) error
	List(ctx context.Context, workspaceID string) ([]types.Member, error)
	Remove(ctx context.Context, workspaceID, memberID string) error
}

type UserRepository interface {
	Create(ctx context.Context, user *types.User, passwordHash string) error
	GetByID(ctx context.Context, id string) (*types.User, error)
	GetByEmail(ctx context.Context, email string) (*types.User, string, error)
}

func NewWorkspaceService(
	workspaceRepo WorkspaceRepository,
	memberRepo MemberRepository,
	userRepo UserRepository,
	logger zerolog.Logger,
) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo: workspaceRepo,
		memberRepo:    memberRepo,
		userRepo:      userRepo,
		logger:        logger,
	}
}

func (s *WorkspaceService) Create(ctx context.Context, req types.CreateWorkspaceRequest) (*types.Workspace, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return nil, ErrUnauthorized
	}

	if req.Name == "" {
		return nil, ErrInvalidInput
	}

	// Set default scoring config if not provided
	scoring := req.Scoring
	if scoring.Type == "" {
		scoring = s.getDefaultScoringConfig()
	}

	workspace := &types.Workspace{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     userID,
		Scoring:     scoring,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.workspaceRepo.Create(ctx, workspace); err != nil {
		s.logger.Error().Err(err).Msg("failed to create workspace")
		return nil, err
	}

	// Add owner as first member
	member := &types.Member{
		ID:          uuid.New().String(),
		WorkspaceID: workspace.ID,
		UserID:      userID,
		Role:        "owner",
		VoteWeight:  1.0,
		JoinedAt:    time.Now(),
	}

	if err := s.memberRepo.Add(ctx, member); err != nil {
		s.logger.Error().Err(err).Msg("failed to add owner as member")
	}

	s.logger.Info().Str("workspaceID", workspace.ID).Str("ownerID", userID).Msg("workspace created")
	return workspace, nil
}

func (s *WorkspaceService) Get(ctx context.Context, id string) (*types.Workspace, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return nil, ErrUnauthorized
	}

	workspace, err := s.workspaceRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check access
	if !s.hasAccess(ctx, workspace.ID, userID) {
		return nil, ErrForbidden
	}

	return workspace, nil
}

func (s *WorkspaceService) Update(ctx context.Context, id string, req types.UpdateWorkspaceRequest) (*types.Workspace, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return nil, ErrUnauthorized
	}

	workspace, err := s.workspaceRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user is owner or admin
	if workspace.OwnerID != userID && !s.isAdmin(ctx, workspace.ID, userID) {
		return nil, ErrForbidden
	}

	// Update fields
	if req.Name != "" {
		workspace.Name = req.Name
	}
	if req.Description != "" {
		workspace.Description = req.Description
	}
	if req.Scoring.Type != "" {
		workspace.Scoring = req.Scoring
	}

	workspace.UpdatedAt = time.Now()

	if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

func (s *WorkspaceService) Delete(ctx context.Context, id string) error {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return ErrUnauthorized
	}

	workspace, err := s.workspaceRepo.Get(ctx, id)
	if err != nil {
		return err
	}

	// Only owner can delete workspace
	if workspace.OwnerID != userID {
		return ErrForbidden
	}

	return s.workspaceRepo.Delete(ctx, id)
}

func (s *WorkspaceService) List(ctx context.Context, userID string, limit int, offset int) ([]types.Workspace, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	return s.workspaceRepo.List(ctx, userID, limit, offset)
}

func (s *WorkspaceService) SetScoringConfig(ctx context.Context, workspaceID string, config types.ScoringConfig) (*types.Workspace, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return nil, ErrUnauthorized
	}

	workspace, err := s.workspaceRepo.Get(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	if workspace.OwnerID != userID && !s.isAdmin(ctx, workspaceID, userID) {
		return nil, ErrForbidden
	}

	// Validate scoring config
	if err := s.validateScoringConfig(config); err != nil {
		return nil, err
	}

	workspace.Scoring = config
	workspace.UpdatedAt = time.Now()

	if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

func (s *WorkspaceService) InviteMember(ctx context.Context, workspaceID string, email string, role string) (*types.Member, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return nil, ErrUnauthorized
	}

	if !s.isAdmin(ctx, workspaceID, userID) {
		return nil, ErrForbidden
	}

	// Check if member already exists
	members, err := s.memberRepo.List(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	for _, m := range members {
		if m.Email == email {
			return nil, ErrAlreadyExists
		}
	}

	// Get user by email if exists
	var memberUserID string
	if user, _, err := s.userRepo.GetByEmail(ctx, email); err == nil {
		memberUserID = user.ID
	}

	member := &types.Member{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		UserID:      memberUserID,
		Email:       email,
		Role:        role,
		VoteWeight:  1.0,
		JoinedAt:    time.Now(),
	}

	if err := s.memberRepo.Add(ctx, member); err != nil {
		return nil, err
	}

	s.logger.Info().Str("workspaceID", workspaceID).Str("email", email).Str("role", role).Msg("member invited")
	return member, nil
}

func (s *WorkspaceService) ListMembers(ctx context.Context, workspaceID string) ([]types.Member, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return nil, ErrUnauthorized
	}

	if !s.hasAccess(ctx, workspaceID, userID) {
		return nil, ErrForbidden
	}

	return s.memberRepo.List(ctx, workspaceID)
}

func (s *WorkspaceService) RemoveMember(ctx context.Context, workspaceID string, memberID string) error {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return ErrUnauthorized
	}

	if !s.isAdmin(ctx, workspaceID, userID) {
		return ErrForbidden
	}

	return s.memberRepo.Remove(ctx, workspaceID, memberID)
}

func (s *WorkspaceService) hasAccess(ctx context.Context, workspaceID string, userID string) bool {
	members, err := s.memberRepo.List(ctx, workspaceID)
	if err != nil {
		return false
	}

	for _, m := range members {
		if m.UserID == userID {
			return true
		}
	}

	return false
}

func (s *WorkspaceService) isAdmin(ctx context.Context, workspaceID string, userID string) bool {
	members, err := s.memberRepo.List(ctx, workspaceID)
	if err != nil {
		return false
	}

	for _, m := range members {
		if m.UserID == userID && (m.Role == "owner" || m.Role == "admin") {
			return true
		}
	}

	return false
}

func (s *WorkspaceService) getDefaultScoringConfig() types.ScoringConfig {
	return types.ScoringConfig{
		Type: "RICE",
		Criteria: []types.Criterion{
			{ID: "reach", Name: "Reach", Type: "number", Weight: 1.0, Required: true},
			{ID: "impact", Name: "Impact", Type: "number", Weight: 1.0, Required: true},
			{ID: "confidence", Name: "Confidence", Type: "number", Weight: 1.0, Required: true},
			{ID: "effort", Name: "Effort", Type: "number", Weight: 1.0, Required: true},
		},
	}
}

func (s *WorkspaceService) validateScoringConfig(config types.ScoringConfig) error {
	if config.Type == "" {
		return ErrInvalidScoring
	}

	if len(config.Criteria) == 0 && config.Type != "RICE" && config.Type != "ICE" && config.Type != "WSJF" {
		return ErrInvalidScoring
	}

	return nil
}
