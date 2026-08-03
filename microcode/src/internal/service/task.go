package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/sah4ez/ducalis-tg/pkg/types"
)

// ===== TASK SERVICE =====

type TaskService struct {
	taskRepo       TaskRepository
	workspaceRepo  WorkspaceRepository
	voteRepo       VoteRepository
	estimationRepo EstimationRepository
	logger         zerolog.Logger
}

type TaskRepository interface {
	Create(ctx context.Context, task *types.Task) error
	Get(ctx context.Context, id string) (*types.Task, error)
	Update(ctx context.Context, task *types.Task) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, req types.ListTasksRequest) ([]types.Task, int, error)
}

type VoteRepository interface {
	Add(ctx context.Context, taskID string, userID string, weight float64) error
	Remove(ctx context.Context, taskID string, userID string) error
	GetByTask(ctx context.Context, taskID string) ([]types.Vote, error)
}

type EstimationRepository interface {
	Add(ctx context.Context, taskID string, userID string, value float64, unit string) error
	GetByTask(ctx context.Context, taskID string) ([]types.Estimation, error)
}

func NewTaskService(
	taskRepo TaskRepository,
	workspaceRepo WorkspaceRepository,
	voteRepo VoteRepository,
	estimationRepo EstimationRepository,
	logger zerolog.Logger,
) *TaskService {
	return &TaskService{
		taskRepo:       taskRepo,
		workspaceRepo:  workspaceRepo,
		voteRepo:       voteRepo,
		estimationRepo: estimationRepo,
		logger:         logger,
	}
}

func (s *TaskService) Create(ctx context.Context, req types.CreateTaskRequest) (*types.Task, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return nil, ErrUnauthorized
	}

	if req.WorkspaceID == "" || req.Title == "" {
		return nil, ErrInvalidInput
	}

	// Verify workspace access
	if _, err := s.workspaceRepo.Get(ctx, req.WorkspaceID); err != nil {
		return nil, err
	}

	now := time.Now()
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
		CreatedAt:    now,
		UpdatedAt:    now,
		Metadata:     req.Metadata,
	}

	if task.Status == "" {
		task.Status = "backlog"
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		s.logger.Error().Err(err).Msg("failed to create task")
		return nil, err
	}

	return task, nil
}

func (s *TaskService) Get(ctx context.Context, id string) (*types.Task, error) {
	task, err := s.taskRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load votes and estimations
	votes, _ := s.voteRepo.GetByTask(ctx, id)
	estimations, _ := s.estimationRepo.GetByTask(ctx, id)

	task.Votes = votes
	task.Estimations = estimations

	return task, nil
}

func (s *TaskService) Update(ctx context.Context, id string, req types.UpdateTaskRequest) (*types.Task, error) {
	task, err := s.taskRepo.Get(ctx, id)
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

	task.UpdatedAt = time.Now()

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) Delete(ctx context.Context, id string) error {
	return s.taskRepo.Delete(ctx, id)
}

func (s *TaskService) List(ctx context.Context, req types.ListTasksRequest) ([]types.Task, int, error) {
	return s.taskRepo.List(ctx, req)
}

func (s *TaskService) SetScores(ctx context.Context, taskID string, scores map[string]float64) (*types.Task, error) {
	task, err := s.taskRepo.Get(ctx, taskID)
	if err != nil {
		return nil, err
	}

	workspace, err := s.workspaceRepo.Get(ctx, task.WorkspaceID)
	if err != nil {
		return nil, err
	}

	// Validate scores against workspace criteria
	for _, criterion := range workspace.Scoring.Criteria {
		if criterion.Required {
			if _, ok := scores[criterion.ID]; !ok {
				return nil, ErrInvalidInput
			}
		}
	}

	task.Scores = scores
	task.FinalScore = s.calculateFinalScore(scores, workspace.Scoring)
	task.UpdatedAt = time.Now()

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) calculateFinalScore(scores map[string]float64, config types.ScoringConfig) float64 {
	// Simple weighted sum for now
	// TODO: Implement formula parsing for custom formulas
	total := 0.0
	for _, criterion := range config.Criteria {
		if val, ok := scores[criterion.ID]; ok {
			total += val * criterion.Weight
		}
	}
	return total
}

func (s *TaskService) Vote(ctx context.Context, taskID string, weight float64) (*types.Task, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return nil, ErrUnauthorized
	}

	if err := s.voteRepo.Add(ctx, taskID, userID, weight); err != nil {
		return nil, err
	}

	return s.Get(ctx, taskID)
}

func (s *TaskService) RemoveVote(ctx context.Context, taskID string) (*types.Task, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return nil, ErrUnauthorized
	}

	if err := s.voteRepo.Remove(ctx, taskID, userID); err != nil {
		return nil, err
	}

	return s.Get(ctx, taskID)
}

func (s *TaskService) Estimate(ctx context.Context, taskID string, value float64, unit string) (*types.Task, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return nil, ErrUnauthorized
	}

	if unit == "" {
		unit = "points"
	}

	if err := s.estimationRepo.Add(ctx, taskID, userID, value, unit); err != nil {
		return nil, err
	}

	return s.Get(ctx, taskID)
}
