package adapter

import (
	"context"
	"errors"

	"github.com/sah4ez/ducalis-tg/internal/service"
	"github.com/sah4ez/ducalis-tg/pkg/types"
)

// TaskAdapter implements contract.TaskService by wrapping service.TaskService
type TaskAdapter struct {
	svc *service.TaskService
}

func NewTaskAdapter(svc *service.TaskService) *TaskAdapter {
	return &TaskAdapter{svc: svc}
}

func (a *TaskAdapter) Create(ctx context.Context, req types.CreateTaskRequest) (*types.Task, error) {
	return a.svc.Create(ctx, req)
}

func (a *TaskAdapter) Get(ctx context.Context, id string) (*types.Task, error) {
	return a.svc.Get(ctx, id)
}

func (a *TaskAdapter) Update(ctx context.Context, id string, req types.UpdateTaskRequest) (*types.Task, error) {
	return a.svc.Update(ctx, id, req)
}

func (a *TaskAdapter) Delete(ctx context.Context, id string) error {
	return a.svc.Delete(ctx, id)
}

func (a *TaskAdapter) List(ctx context.Context, req types.ListTasksRequest) ([]types.Task, int, error) {
	return a.svc.List(ctx, req)
}

func (a *TaskAdapter) SetScores(ctx context.Context, taskID string, scores map[string]float64) (*types.Task, error) {
	return a.svc.SetScores(ctx, taskID, scores)
}

func (a *TaskAdapter) Vote(ctx context.Context, taskID string, userID string, weight float64) (*types.Task, error) {
	// Contract passes userID explicitly; service reads from context.
	ctx = context.WithValue(ctx, service.UserIDKey, userID)
	return a.svc.Vote(ctx, taskID, weight)
}

func (a *TaskAdapter) RemoveVote(ctx context.Context, taskID string, userID string) (*types.Task, error) {
	ctx = context.WithValue(ctx, service.UserIDKey, userID)
	return a.svc.RemoveVote(ctx, taskID)
}

func (a *TaskAdapter) Estimate(ctx context.Context, taskID string, userID string, value float64) (*types.Task, error) {
	ctx = context.WithValue(ctx, service.UserIDKey, userID)
	return a.svc.Estimate(ctx, taskID, value, "points")
}

// SetDependencies is not yet implemented in the service layer.
func (a *TaskAdapter) SetDependencies(ctx context.Context, taskID string, dependencyIDs []string) (*types.Task, error) {
	return nil, errors.New("not implemented")
}

// GetRanked is not yet implemented in the service layer.
func (a *TaskAdapter) GetRanked(ctx context.Context, workspaceID string, limit int, offset int) ([]types.TaskWithRank, error) {
	return nil, errors.New("not implemented")
}
