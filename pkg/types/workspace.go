package types

import "time"

// Workspace represents a team workspace with scoring configuration
type Workspace struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	OwnerID     string        `json:"ownerId"`
	Scoring     ScoringConfig `json:"scoring"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
}

// CreateWorkspaceRequest for creating workspace
type CreateWorkspaceRequest struct {
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Scoring     ScoringConfig `json:"scoring,omitempty"`
}

// UpdateWorkspaceRequest for updating workspace
type UpdateWorkspaceRequest struct {
	Name        string        `json:"name,omitempty"`
	Description string        `json:"description,omitempty"`
	Scoring     ScoringConfig `json:"scoring,omitempty"`
}

// ScoringConfig defines scoring criteria and weights
type ScoringConfig struct {
	// Preset: RICE, ICE, WSJF, CUSTOM
	Type string `json:"type"`
	// Custom criteria (if type=CUSTOM)
	Criteria []Criterion `json:"criteria,omitempty"`
	// Formula for final score calculation
	Formula string `json:"formula,omitempty"` // e.g. "(reach * impact * confidence) / effort"
}

// Criterion defines a single scoring criterion
type Criterion struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`        // e.g., "Reach", "Impact"
	Description string   `json:"description"` // e.g., "How many users affected?"
	Type        string   `json:"type"`        // "number", "enum", "boolean"
	Scale       *Scale   `json:"scale,omitempty"`
	Options     []string `json:"options,omitempty"` // for enum type
	Weight      float64  `json:"weight"`            // relative weight in calculation
	Required    bool     `json:"required"`
}

// Scale defines numeric range
type Scale struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// Member represents workspace member
type Member struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspaceId"`
	UserID      string    `json:"userId"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Role        string    `json:"role"` // owner, admin, member
	VoteWeight  float64   `json:"voteWeight"`
	JoinedAt    time.Time `json:"joinedAt"`
}

// User represents authenticated user
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatarUrl,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// RegisterRequest for user registration
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
