package contract

import (
	"context"

	"github.com/sah4ez/priora/pkg/types"
)

// TaskService manages tasks and their scoring
// @tg tag=tasks Tasks management
// @tg http-prefix=api/v1/tasks
type TaskService interface {

	// Create creates a new task
	// @tg summary=Create task
	// @tg desc=Creates a task in workspace with optional external reference
	// @tg http-method=POST
	// @tg http-path=/
	Create(ctx context.Context, req types.CreateTaskRequest) (task *types.Task, err error)

	// Get returns task by ID
	// @tg summary=Get task
	// @tg http-method=GET
	// @tg http-path=/{id}
	// @tg http-path-params=id:id
	Get(ctx context.Context, id string) (task *types.Task, err error)

	// Update updates task properties
	// @tg summary=Update task
	// @tg http-method=PATCH
	// @tg http-path=/{id}
	// @tg http-path-params=id:id
	Update(ctx context.Context, id string, req types.UpdateTaskRequest) (task *types.Task, err error)

	// Delete deletes task
	// @tg summary=Delete task
	// @tg http-method=DELETE
	// @tg http-path=/{id}
	// @tg http-path-params=id:id
	Delete(ctx context.Context, id string) (err error)

	// List returns tasks with filters and sorting
	// @tg summary=List tasks
	// @tg http-method=GET
	// @tg http-path=/
	List(ctx context.Context, req types.ListTasksRequest) (tasks []types.Task, total int, err error)

	// SetScores sets scoring values for a task
	// @tg summary=Set task scores
	// @tg http-method=PUT
	// @tg http-path=/{id}/scores
	// @tg http-path-params=id:taskID
	SetScores(ctx context.Context, taskID string, scores map[string]float64) (task *types.Task, err error)

	// Vote adds user's vote to task
	// @tg summary=Vote for task
	// @tg http-method=POST
	// @tg http-path=/{id}/vote
	// @tg http-path-params=id:taskID
	Vote(ctx context.Context, taskID string, userID string, weight float64) (task *types.Task, err error)

	// RemoveVote removes user's vote from task
	// @tg summary=Remove vote
	// @tg http-method=DELETE
	// @tg http-path=/{id}/vote
	// @tg http-path-params=id:taskID
	RemoveVote(ctx context.Context, taskID string, userID string) (task *types.Task, err error)

	// Estimate adds user's estimation to task
	// @tg summary=Estimate task
	// @tg http-method=POST
	// @tg http-path=/{id}/estimate
	// @tg http-path-params=id:taskID
	Estimate(ctx context.Context, taskID string, userID string, value float64) (task *types.Task, err error)

	// SetDependencies sets task dependencies
	// @tg summary=Set dependencies
	// @tg http-method=PUT
	// @tg http-path=/{id}/dependencies
	// @tg http-path-params=id:taskID
	SetDependencies(ctx context.Context, taskID string, dependencyIDs []string) (task *types.Task, err error)

	// GetRanked returns tasks ranked by final score
	// @tg summary=Get ranked tasks
	// @tg http-method=GET
	// @tg http-path=/ranked
	GetRanked(ctx context.Context, workspaceID string, limit int, offset int) (tasks []types.TaskWithRank, err error)
}
