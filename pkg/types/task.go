package types

import "time"

// Task represents a prioritizable task
type Task struct {
	ID           string                 `json:"id"`
	WorkspaceID  string                 `json:"workspaceId"`
	ExternalID   string                 `json:"externalId,omitempty"`   // ID in external system
	ExternalType string                 `json:"externalType,omitempty"` // github, jira, linear
	ExternalURL  string                 `json:"externalUrl,omitempty"`  // Link to external task
	Title        string                 `json:"title"`
	Description  string                 `json:"description,omitempty"`
	Scores       map[string]float64     `json:"scores,omitempty"` // criterion_id -> value
	FinalScore   float64                `json:"finalScore"`
	Votes        []Vote                 `json:"votes,omitempty"`
	Estimations  []Estimation           `json:"estimations,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"` // task IDs that block this task
	Status       string                 `json:"status"`                 // backlog, in_progress, done, cancelled
	Priority     string                 `json:"priority,omitempty"`     // low, medium, high, critical
	Labels       []string               `json:"labels,omitempty"`
	AssigneeID   string                 `json:"assigneeId,omitempty"`
	CreatedBy    string                 `json:"createdBy"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"` // arbitrary external data
}

// CreateTaskRequest for creating task
type CreateTaskRequest struct {
	WorkspaceID  string                 `json:"workspaceId"`
	ExternalID   string                 `json:"externalId,omitempty"`
	ExternalType string                 `json:"externalType,omitempty"`
	ExternalURL  string                 `json:"externalUrl,omitempty"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description,omitempty"`
	Scores       map[string]float64     `json:"scores,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Status       string                 `json:"status,omitempty"`
	Priority     string                 `json:"priority,omitempty"`
	Labels       []string               `json:"labels,omitempty"`
	AssigneeID   string                 `json:"assigneeId,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTaskRequest for updating task
type UpdateTaskRequest struct {
	Title        string                 `json:"title,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Status       string                 `json:"status,omitempty"`
	Priority     string                 `json:"priority,omitempty"`
	Labels       []string               `json:"labels,omitempty"`
	AssigneeID   string                 `json:"assigneeId,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ListTasksRequest for listing tasks with filters
type ListTasksRequest struct {
	WorkspaceID  string   `json:"workspaceId"`
	Status       string   `json:"status,omitempty"`
	AssigneeID   string   `json:"assigneeId,omitempty"`
	Labels       []string `json:"labels,omitempty"`
	HasScore     *bool    `json:"hasScore,omitempty"`     // filter by having scores
	HasVotes     *bool    `json:"hasVotes,omitempty"`     // filter by having votes
	MinScore     float64  `json:"minScore,omitempty"`     // minimum final score
	MaxScore     float64  `json:"maxScore,omitempty"`     // maximum final score
	Search       string   `json:"search,omitempty"`       // search in title/description
	SortBy       string   `json:"sortBy,omitempty"`       // finalScore, createdAt, votes
	SortDesc     bool     `json:"sortDesc,omitempty"`     // sort descending
	Limit        int      `json:"limit,omitempty"`
	Offset       int      `json:"offset,omitempty"`
}

// Vote represents user's vote for a task
type Vote struct {
	UserID    string    `json:"userId"`
	Weight    float64   `json:"weight"` // vote weight (can be based on role)
	CreatedAt time.Time `json:"createdAt"`
}

// Estimation represents user's time/effort estimation
type Estimation struct {
	UserID    string    `json:"userId"`
	Value     float64   `json:"value"` // story points or hours
	Unit      string    `json:"unit"`  // "points", "hours"
	CreatedAt time.Time `json:"createdAt"`
}

// TaskWithRank is task with calculated rank position
type TaskWithRank struct {
	Task
	Rank       int     `json:"rank"`
	Percentile float64 `json:"percentile"` // position in percentile (0-100)
}
