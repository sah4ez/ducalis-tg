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

// List maps the contract's inline query params onto ListTasksRequest.
func (a *TaskAdapter) List(ctx context.Context, workspaceID string, status string, search string, sortBy string, sortDesc bool, limit int, offset int) ([]types.Task, int, error) {
	return a.svc.List(ctx, types.ListTasksRequest{
		WorkspaceID: workspaceID,
		Status:      status,
		Search:      search,
		SortBy:      sortBy,
		SortDesc:    sortDesc,
		Limit:       limit,
		Offset:      offset,
	})
}

func (a *TaskAdapter) SetScores(ctx context.Context, id string, scores map[string]float64) (*types.Task, error) {
	return a.svc.SetScores(ctx, id, scores)
}

// Vote prefers the explicit userID contract param; falls back to the JWT
// middleware's context value when the param is empty.
func (a *TaskAdapter) Vote(ctx context.Context, id string, userID string, weight float64) (*types.Task, error) {
	if userID != "" {
		ctx = context.WithValue(ctx, service.UserIDKey, userID)
	}
	return a.svc.Vote(ctx, id, weight)
}

func (a *TaskAdapter) RemoveVote(ctx context.Context, id string, userID string) (*types.Task, error) {
	if userID != "" {
		ctx = context.WithValue(ctx, service.UserIDKey, userID)
	}
	return a.svc.RemoveVote(ctx, id)
}

func (a *TaskAdapter) Estimate(ctx context.Context, id string, userID string, value float64, unit string) (*types.Task, error) {
	if userID != "" {
		ctx = context.WithValue(ctx, service.UserIDKey, userID)
	}
	if unit == "" {
		unit = "points"
	}
	return a.svc.Estimate(ctx, id, value, unit)
}

// SetDependencies is not yet implemented in the service layer.
func (a *TaskAdapter) SetDependencies(ctx context.Context, id string, dependencyIDs []string) (*types.Task, error) {
	return nil, errors.New("not implemented")
}

func (a *TaskAdapter) GetRanked(ctx context.Context, workspaceID string, limit int, offset int) (types.RankedTasks, error) {
	return a.svc.GetRanked(ctx, workspaceID, limit, offset)
}
