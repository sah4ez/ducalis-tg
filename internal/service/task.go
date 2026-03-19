package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/sah4ez/priora/pkg/types"
)

// TaskService implements contract.TaskService
type TaskService struct {
	store  TaskStore
	logger zerolog.Logger
}

// TaskStore defines storage interface for tasks
type TaskStore interface {
	Create(ctx context.Context, task *types.Task) error
	Get(ctx context.Context, id string) (*types.Task, error)
	Update(ctx context.Context, task *types.Task) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, req types.ListTasksRequest) ([]types.Task, int, error)
	AddVote(ctx context.Context, taskID, userID string, weight float64) error
	RemoveVote(ctx context.Context, taskID, userID string) error
	AddEstimation(ctx context.Context, taskID, userID string, value float64) error
	SetDependencies(ctx context.Context, taskID string, dependencyIDs []string) error
	GetRanked(ctx context.Context, workspaceID string, limit, offset int) ([]types.TaskWithRank, error)
}

// NewTaskService creates new task service
func NewTaskService(store TaskStore, logger zerolog.Logger) *TaskService {
	return &TaskService{
		store:  store,
		logger: logger,
	}
}

// Create creates a new task
func (s *TaskService) Create(ctx context.Context, req types.CreateTaskRequest) (*types.Task, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	task := &types.Task{
		ID:           uuid.New().String(),
		WorkspaceID:  req.WorkspaceID,
		ExternalID:   req.ExternalID,
		ExternalType: req.ExternalType,
		ExternalURL:  req.ExternalURL,
		Title:        req.Title,
		Description:  req.Description,
		Scores:       req.Scores,
		Status:       req.Status,
		Priority:     req.Priority,
		Labels:       req.Labels,
		AssigneeID:   req.AssigneeID,
		CreatedBy:    userID,
	}

	if task.Status == "" {
		task.Status = "backlog"
	}

	if err := s.store.Create(ctx, task); err != nil {
		s.logger.Error().Err(err).Msg("failed to create task")
		return nil, err
	}

	return task, nil
}

// Get returns task by ID
func (s *TaskService) Get(ctx context.Context, id string) (*types.Task, error) {
	return s.store.Get(ctx, id)
}

// Update updates task properties
func (s *TaskService) Update(ctx context.Context, id string, req types.UpdateTaskRequest) (*types.Task, error) {
	task, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Status != "" {
		task.Status = req.Status
	}
	if req.Priority != "" {
		task.Priority = req.Priority
	}
	if req.Labels != nil {
		task.Labels = req.Labels
	}
	if req.AssigneeID != "" {
		task.AssigneeID = req.AssigneeID
	}

	if err := s.store.Update(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

// Delete deletes task
func (s *TaskService) Delete(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}

// List returns tasks with filters
func (s *TaskService) List(ctx context.Context, req types.ListTasksRequest) ([]types.Task, int, error) {
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	return s.store.List(ctx, req)
}

// SetScores sets scoring values for a task
func (s *TaskService) SetScores(ctx context.Context, taskID string, scores map[string]float64) (*types.Task, error) {
	task, err := s.store.Get(ctx, taskID)
	if err != nil {
		return nil, err
	}

	task.Scores = scores

	// TODO: Calculate final score based on workspace scoring config

	if err := s.store.Update(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

// Vote adds user's vote to task
func (s *TaskService) Vote(ctx context.Context, taskID string, userID string, weight float64) (*types.Task, error) {
	if err := s.store.AddVote(ctx, taskID, userID, weight); err != nil {
		return nil, err
	}

	return s.store.Get(ctx, taskID)
}

// RemoveVote removes user's vote from task
func (s *TaskService) RemoveVote(ctx context.Context, taskID string, userID string) (*types.Task, error) {
	if err := s.store.RemoveVote(ctx, taskID, userID); err != nil {
		return nil, err
	}

	return s.store.Get(ctx, taskID)
}

// Estimate adds user's estimation to task
func (s *TaskService) Estimate(ctx context.Context, taskID string, userID string, value float64) (*types.Task, error) {
	if err := s.store.AddEstimation(ctx, taskID, userID, value); err != nil {
		return nil, err
	}

	return s.store.Get(ctx, taskID)
}

// SetDependencies sets task dependencies
func (s *TaskService) SetDependencies(ctx context.Context, taskID string, dependencyIDs []string) (*types.Task, error) {
	if err := s.store.SetDependencies(ctx, taskID, dependencyIDs); err != nil {
		return nil, err
	}

	return s.store.Get(ctx, taskID)
}

// GetRanked returns tasks ranked by final score
func (s *TaskService) GetRanked(ctx context.Context, workspaceID string, limit int, offset int) ([]types.TaskWithRank, error) {
	if limit <= 0 {
		limit = 50
	}

	return s.store.GetRanked(ctx, workspaceID, limit, offset)
}
