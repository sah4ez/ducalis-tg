package adapter

import (
	"context"
	"testing"

	"github.com/sah4ez/ducalis-tg/internal/service"
	"github.com/sah4ez/ducalis-tg/pkg/types"
	"github.com/rs/zerolog"
)

func TestTaskAdapter_Create(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Logger()
	wsRepo := newStubWorkspaceRepo()
	// Pre-create workspace so task creation can verify it exists
	wsRepo.workspaces["ws-1"] = &types.Workspace{ID: "ws-1", Name: "WS"}
	adapter := NewTaskAdapter(service.NewTaskService(
		newStubTaskRepo(), wsRepo, newStubVoteRepo(), newStubEstimationRepo(), logger))

	ctx := context.WithValue(context.Background(), service.UserIDKey, "user-1")
	task, err := adapter.Create(ctx, types.CreateTaskRequest{
		WorkspaceID: "ws-1",
		Title:       "Test Task",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "Test Task" {
		t.Errorf("expected title Test Task, got %s", task.Title)
	}
	if task.CreatedBy != "user-1" {
		t.Errorf("expected created by user-1, got %s", task.CreatedBy)
	}
}

func TestTaskAdapter_Vote(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Logger()
	wsRepo := newStubWorkspaceRepo()
	wsRepo.workspaces["ws-1"] = &types.Workspace{ID: "ws-1", Name: "WS"}
	taskRepo := newStubTaskRepo()
	adapter := NewTaskAdapter(service.NewTaskService(
		taskRepo, wsRepo, newStubVoteRepo(), newStubEstimationRepo(), logger))

	ctx := context.WithValue(context.Background(), service.UserIDKey, "user-1")
	task, err := adapter.Create(ctx, types.CreateTaskRequest{
		WorkspaceID: "ws-1",
		Title:       "Votable Task",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	result, err := adapter.Vote(context.Background(), task.ID, "user-2", 1.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil task")
	}
}

func TestTaskAdapter_SetDependencies_NotImplemented(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Logger()
	adapter := NewTaskAdapter(service.NewTaskService(
		newStubTaskRepo(), newStubWorkspaceRepo(), newStubVoteRepo(), newStubEstimationRepo(), logger))

	_, err := adapter.SetDependencies(context.Background(), "task-1", []string{"dep-1"})
	if err == nil {
		t.Error("expected error for unimplemented SetDependencies")
	}
}

func TestTaskAdapter_GetRanked_NotImplemented(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Logger()
	adapter := NewTaskAdapter(service.NewTaskService(
		newStubTaskRepo(), newStubWorkspaceRepo(), newStubVoteRepo(), newStubEstimationRepo(), logger))

	_, err := adapter.GetRanked(context.Background(), "ws-1", 10, 0)
	if err == nil {
		t.Error("expected error for unimplemented GetRanked")
	}
}

// stub task repo
type stubTaskRepo struct {
	tasks map[string]*types.Task
}

func newStubTaskRepo() *stubTaskRepo {
	return &stubTaskRepo{tasks: make(map[string]*types.Task)}
}

func (r *stubTaskRepo) Create(ctx context.Context, task *types.Task) error {
	r.tasks[task.ID] = task
	return nil
}

func (r *stubTaskRepo) Get(ctx context.Context, id string) (*types.Task, error) {
	if t, ok := r.tasks[id]; ok {
		return t, nil
	}
	return nil, service.ErrNotFound
}

func (r *stubTaskRepo) Update(ctx context.Context, task *types.Task) error {
	r.tasks[task.ID] = task
	return nil
}

func (r *stubTaskRepo) Delete(ctx context.Context, id string) error {
	delete(r.tasks, id)
	return nil
}

func (r *stubTaskRepo) List(ctx context.Context, req types.ListTasksRequest) ([]types.Task, int, error) {
	var result []types.Task
	for _, t := range r.tasks {
		result = append(result, *t)
	}
	return result, len(result), nil
}

// stub vote repo
type stubVoteRepo struct{}

func newStubVoteRepo() *stubVoteRepo { return &stubVoteRepo{} }
func (r *stubVoteRepo) Add(ctx context.Context, taskID, userID string, weight float64) error { return nil }
func (r *stubVoteRepo) Remove(ctx context.Context, taskID, userID string) error            { return nil }
func (r *stubVoteRepo) GetByTask(ctx context.Context, taskID string) ([]types.Vote, error) { return nil, nil }

// stub estimation repo
type stubEstimationRepo struct{}

func newStubEstimationRepo() *stubEstimationRepo { return &stubEstimationRepo{} }
func (r *stubEstimationRepo) Add(ctx context.Context, taskID, userID string, value float64, unit string) error {
	return nil
}
func (r *stubEstimationRepo) GetByTask(ctx context.Context, taskID string) ([]types.Estimation, error) {
	return nil, nil
}
