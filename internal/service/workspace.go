package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/sah4ez/ducalis-tg/pkg/types"
)

// WorkspaceService implements contract.WorkspaceService
type WorkspaceService struct {
	store  WorkspaceStore
	logger zerolog.Logger
}

// WorkspaceStore defines storage interface
type WorkspaceStore interface {
	Create(ctx context.Context, workspace *types.Workspace) error
	Get(ctx context.Context, id string) (*types.Workspace, error)
	Update(ctx context.Context, workspace *types.Workspace) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, userID string, limit, offset int) ([]types.Workspace, int, error)
	AddMember(ctx context.Context, member *types.Member) error
	ListMembers(ctx context.Context, workspaceID string) ([]types.Member, error)
	RemoveMember(ctx context.Context, workspaceID, memberID string) error
}

// NewWorkspaceService creates new workspace service
func NewWorkspaceService(store WorkspaceStore, logger zerolog.Logger) *WorkspaceService {
	return &WorkspaceService{
		store:  store,
		logger: logger,
	}
}

// Create creates a new workspace
func (s *WorkspaceService) Create(ctx context.Context, req types.CreateWorkspaceRequest) (*types.Workspace, error) {
	// Get user ID from context (set by auth middleware)
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	// Set default scoring if not provided
	scoring := req.Scoring
	if scoring.Type == "" {
		scoring = types.ScoringConfig{
			Type: "RICE",
			Criteria: []types.Criterion{
				{ID: "reach", Name: "Reach", Type: "number", Weight: 1.0, Required: true, Scale: &types.Scale{Min: 1, Max: 10}},
				{ID: "impact", Name: "Impact", Type: "number", Weight: 1.0, Required: true, Scale: &types.Scale{Min: 1, Max: 10}},
				{ID: "confidence", Name: "Confidence", Type: "number", Weight: 1.0, Required: true, Scale: &types.Scale{Min: 0.1, Max: 1.0}},
				{ID: "effort", Name: "Effort", Type: "number", Weight: 1.0, Required: true, Scale: &types.Scale{Min: 1, Max: 100}},
			},
			Formula: "(reach * impact * confidence) / effort",
		}
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

	if err := s.store.Create(ctx, workspace); err != nil {
		s.logger.Error().Err(err).Msg("failed to create workspace")
		return nil, err
	}

	// Add owner as first member
	ownerMember := &types.Member{
		ID:          uuid.New().String(),
		WorkspaceID: workspace.ID,
		UserID:      userID,
		Role:        "owner",
		VoteWeight:  1.0,
		JoinedAt:    time.Now(),
	}

	if err := s.store.AddMember(ctx, ownerMember); err != nil {
		s.logger.Error().Err(err).Msg("failed to add owner as member")
	}

	s.logger.Info().Str("workspaceID", workspace.ID).Msg("workspace created")
	return workspace, nil
}

// Get returns workspace by ID
func (s *WorkspaceService) Get(ctx context.Context, id string) (*types.Workspace, error) {
	workspace, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// TODO: Check if user has access to this workspace

	return workspace, nil
}

// Update updates workspace settings
func (s *WorkspaceService) Update(ctx context.Context, id string, req types.UpdateWorkspaceRequest) (*types.Workspace, error) {
	workspace, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// TODO: Check if user is owner or admin

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

	if err := s.store.Update(ctx, workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

// Delete deletes workspace
func (s *WorkspaceService) Delete(ctx context.Context, id string) error {
	workspace, err := s.store.Get(ctx, id)
	if err != nil {
		return err
	}

	// TODO: Check if user is owner

	return s.store.Delete(ctx, workspace.ID)
}

// List returns user's workspaces
func (s *WorkspaceService) List(ctx context.Context, userID string, limit int, offset int) ([]types.Workspace, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	return s.store.List(ctx, userID, limit, offset)
}

// SetScoringConfig configures scoring criteria and weights
func (s *WorkspaceService) SetScoringConfig(ctx context.Context, workspaceID string, config types.ScoringConfig) (*types.Workspace, error) {
	workspace, err := s.store.Get(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// TODO: Validate scoring config

	workspace.Scoring = config
	workspace.UpdatedAt = time.Now()

	if err := s.store.Update(ctx, workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

// InviteMember invites a new member to workspace
func (s *WorkspaceService) InviteMember(ctx context.Context, workspaceID string, email string, role string) (*types.Member, error) {
	workspace, err := s.store.Get(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// TODO: Check if user can invite
	// TODO: Send invitation email

	member := &types.Member{
		ID:          uuid.New().String(),
		WorkspaceID: workspace.ID,
		Email:       email,
		Role:        role,
		VoteWeight:  1.0,
		JoinedAt:    time.Now(),
	}

	if err := s.store.AddMember(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// ListMembers returns workspace members
func (s *WorkspaceService) ListMembers(ctx context.Context, workspaceID string) ([]types.Member, error) {
	return s.store.ListMembers(ctx, workspaceID)
}

// RemoveMember removes member from workspace
func (s *WorkspaceService) RemoveMember(ctx context.Context, workspaceID string, memberID string) error {
	// TODO: Check permissions
	return s.store.RemoveMember(ctx, workspaceID, memberID)
}
