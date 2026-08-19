// Package contract declares the service APIs as @tg-annotated Go interfaces.
// The HTTP (fiber) transport is GENERATED from these contracts by
// `tg server -o internal/transport` (tg v3, tgp-go plugins).
package contract

import (
	"context"

	"github.com/sah4ez/ducalis-tg/pkg/types"
)

// TaskService manages tasks and their scoring.
//
// @tg title=`Task Service`
// @tg version=`1.0.0`
// @tg desc=`Task CRUD, scoring (RICE/ICE/WSJF), voting, estimations and ranked priorities.`
// @tg http-server
// @tg metrics
type TaskService interface {

	// Create creates a new task in a workspace.
	//
	// @tg http-method=POST
	// @tg http-path=`/api/v1/tasks`
	// @tg http-success=201
	// @tg summary=`Create task`
	Create(ctx context.Context, req types.CreateTaskRequest) (task *types.Task, err error)

	// Get returns task by ID.
	//
	// @tg http-method=GET
	// @tg http-path=`/api/v1/tasks/:id`
	// @tg http-args=id|id|explicit
	// @tg enableInlineSingle
	// @tg summary=`Get task`
	Get(ctx context.Context, id string) (task *types.Task, err error)

	// Update updates task properties.
	//
	// @tg http-method=PATCH
	// @tg http-path=`/api/v1/tasks/:id`
	// @tg http-args=id|id|explicit
	// @tg enableInlineSingle
	// @tg summary=`Update task`
	Update(ctx context.Context, id string, req types.UpdateTaskRequest) (task *types.Task, err error)

	// Delete deletes task.
	//
	// @tg http-method=DELETE
	// @tg http-path=`/api/v1/tasks/:id`
	// @tg http-args=id|id|explicit
	// @tg enableInlineSingle
	// @tg http-success=204
	// @tg summary=`Delete task`
	Delete(ctx context.Context, id string) (err error)

	// List returns tasks with filters and sorting (all via query params).
	//
	// @tg http-method=GET
	// @tg http-path=`/api/v1/tasks`
	// @tg http-args=workspaceID|workspaceID|explicit
	// @tg http-args=status|status|explicit
	// @tg http-args=search|search|explicit
	// @tg http-args=sortBy|sortBy|explicit
	// @tg http-args=sortDesc|sortDesc|explicit
	// @tg http-args=limit|limit|explicit
	// @tg http-args=offset|offset|explicit
	// @tg summary=`List tasks`
	List(ctx context.Context, workspaceID string, status string, search string, sortBy string, sortDesc bool, limit int, offset int) (tasks []types.Task, total int, err error)

	// SetScores sets scoring values for a task.
	//
	// @tg http-method=PUT
	// @tg http-path=`/api/v1/tasks/:id/scores`
	// @tg http-args=id|id|explicit
	// @tg enableInlineSingle
	// @tg summary=`Set task scores`
	SetScores(ctx context.Context, id string, scores map[string]float64) (task *types.Task, err error)

	// Vote adds user's vote to task.
	//
	// @tg http-method=POST
	// @tg http-path=`/api/v1/tasks/:id/vote`
	// @tg http-args=id|id|explicit
	// @tg enableInlineSingle
	// @tg summary=`Vote for task`
	Vote(ctx context.Context, id string, userID string, weight float64) (task *types.Task, err error)

	// RemoveVote removes user's vote from task.
	//
	// @tg http-method=DELETE
	// @tg http-path=`/api/v1/tasks/:id/vote`
	// @tg http-args=id|id|explicit
	// @tg enableInlineSingle
	// @tg summary=`Remove vote`
	RemoveVote(ctx context.Context, id string, userID string) (task *types.Task, err error)

	// Estimate adds user's estimation to task.
	//
	// @tg http-method=POST
	// @tg http-path=`/api/v1/tasks/:id/estimate`
	// @tg http-args=id|id|explicit
	// @tg enableInlineSingle
	// @tg summary=`Estimate task`
	Estimate(ctx context.Context, id string, userID string, value float64, unit string) (task *types.Task, err error)

	// SetDependencies sets task dependencies.
	//
	// @tg http-method=PUT
	// @tg http-path=`/api/v1/tasks/:id/dependencies`
	// @tg http-args=id|id|explicit
	// @tg enableInlineSingle
	// @tg summary=`Set dependencies`
	SetDependencies(ctx context.Context, id string, dependencyIDs []string) (task *types.Task, err error)

	// GetRanked returns tasks ranked by final score.
	//
	// @tg http-method=GET
	// @tg http-path=`/api/v1/workspaces/:workspaceID/tasks/ranked`
	// @tg http-args=workspaceID|workspaceID|explicit
	// @tg http-args=limit|limit|explicit
	// @tg http-args=offset|offset|explicit
	// @tg summary=`Get ranked tasks`
	GetRanked(ctx context.Context, workspaceID string, limit int, offset int) (result types.RankedTasks, err error)
}
