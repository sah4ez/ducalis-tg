package types

import "time"

// Integration represents connection to external service
type Integration struct {
	ID           string                 `json:"id"`
	WorkspaceID  string                 `json:"workspaceId"`
	Type         string                 `json:"type"` // github, jira, linear
	Name         string                 `json:"name"`
	Config       map[string]interface{} `json:"config"` // type-specific config (encrypted)
	LastSyncAt   *time.Time             `json:"lastSyncAt,omitempty"`
	SyncStatus   string                 `json:"syncStatus"` // idle, syncing, error
	SyncError    string                 `json:"syncError,omitempty"`
	AutoSync     bool                   `json:"autoSync"`
	SyncInterval int                    `json:"syncInterval,omitempty"` // minutes
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
}

// CreateGitHubIntegrationRequest for GitHub integration
type CreateGitHubIntegrationRequest struct {
	WorkspaceID string `json:"workspaceId"`
	Name        string `json:"name"`
	// Personal access token or OAuth
	Token       string `json:"token"`
	Owner       string `json:"owner"`       // repo owner
	Repo        string `json:"repo"`        // repo name
	AutoSync    bool   `json:"autoSync"`
	SyncLabels  bool   `json:"syncLabels"`  // sync issues with specific labels
	LabelFilter string `json:"labelFilter"` // comma-separated labels
}

// CreateJiraIntegrationRequest for Jira integration
type CreateJiraIntegrationRequest struct {
	WorkspaceID string `json:"workspaceId"`
	Name        string `json:"name"`
	BaseURL     string `json:"baseUrl"`     // e.g., https://company.atlassian.net
	Username    string `json:"username"`
	APIToken    string `json:"apiToken"`
	ProjectKey  string `json:"projectKey"`
	JQLFilter   string `json:"jqlFilter"`   // JQL query to filter issues
	AutoSync    bool   `json:"autoSync"`
}

// CreateLinearIntegrationRequest for Linear integration
type CreateLinearIntegrationRequest struct {
	WorkspaceID string `json:"workspaceId"`
	Name        string `json:"name"`
	APIKey      string `json:"apiKey"`
	TeamID      string `json:"teamId"`
	AutoSync    bool   `json:"autoSync"`
}

// SyncResult contains sync operation results
type SyncResult struct {
	IntegrationID   string      `json:"integrationId"`
	Status          string      `json:"status"` // success, partial, error
	TasksCreated    int         `json:"tasksCreated"`
	TasksUpdated    int         `json:"tasksUpdated"`
	TasksDeleted    int         `json:"tasksDeleted"`
	Errors          []SyncError `json:"errors,omitempty"`
	DurationMs      int64       `json:"durationMs"`
	NextScheduledAt *time.Time  `json:"nextScheduledAt,omitempty"`
}

// SyncError represents a single sync error
type SyncError struct {
	ExternalID string `json:"externalId"`
	Error      string `json:"error"`
}

// SyncStatus represents current sync state
type SyncStatus struct {
	IntegrationID   string     `json:"integrationId"`
	Status          string     `json:"status"` // idle, syncing, error
	LastSyncAt      *time.Time `json:"lastSyncAt,omitempty"`
	LastSyncResult  string     `json:"lastSyncResult,omitempty"`
	LastError       string     `json:"lastError,omitempty"`
	NextSyncAt      *time.Time `json:"nextSyncAt,omitempty"`
	TotalTasks      int        `json:"totalTasks"`
	LastDurationMs  int64      `json:"lastDurationMs,omitempty"`
}
