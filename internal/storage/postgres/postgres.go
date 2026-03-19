package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sah4ez/ducalis-tg/pkg/types"
)

// Store implements storage interfaces using PostgreSQL
type Store struct {
	pool *pgxpool.Pool
}

// New creates new PostgreSQL store
func New(databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Store{pool: pool}, nil
}

// Close closes database connection
func (s *Store) Close() {
	s.pool.Close()
}

// Workspace storage methods

// Create creates new workspace
func (s *Store) Create(ctx context.Context, workspace *types.Workspace) error {
	query := `
		INSERT INTO workspaces (id, name, description, owner_id, scoring_config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.pool.Exec(ctx, query,
		workspace.ID,
		workspace.Name,
		workspace.Description,
		workspace.OwnerID,
		workspace.Scoring,
		workspace.CreatedAt,
		workspace.UpdatedAt,
	)
	return err
}

// Get returns workspace by ID
func (s *Store) Get(ctx context.Context, id string) (*types.Workspace, error) {
	query := `
		SELECT id, name, description, owner_id, scoring_config, created_at, updated_at
		FROM workspaces
		WHERE id = $1
	`
	workspace := &types.Workspace{}
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&workspace.ID,
		&workspace.Name,
		&workspace.Description,
		&workspace.OwnerID,
		&workspace.Scoring,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return workspace, nil
}

// Update updates workspace
func (s *Store) Update(ctx context.Context, workspace *types.Workspace) error {
	query := `
		UPDATE workspaces
		SET name = $2, description = $3, scoring_config = $4, updated_at = $5
		WHERE id = $1
	`
	_, err := s.pool.Exec(ctx, query,
		workspace.ID,
		workspace.Name,
		workspace.Description,
		workspace.Scoring,
		workspace.UpdatedAt,
	)
	return err
}

// Delete deletes workspace
func (s *Store) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM workspaces WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, id)
	return err
}

// List returns user's workspaces
func (s *Store) List(ctx context.Context, userID string, limit, offset int) ([]types.Workspace, int, error) {
	// Get total count
	var total int
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM workspaces w
		 INNER JOIN members m ON m.workspace_id = w.id
		 WHERE m.user_id = $1`,
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get workspaces
	query := `
		SELECT w.id, w.name, w.description, w.owner_id, w.scoring_config, w.created_at, w.updated_at
		FROM workspaces w
		INNER JOIN members m ON m.workspace_id = w.id
		WHERE m.user_id = $1
		ORDER BY w.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := s.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var workspaces []types.Workspace
	for rows.Next() {
		var w types.Workspace
		if err := rows.Scan(
			&w.ID, &w.Name, &w.Description, &w.OwnerID, &w.Scoring, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		workspaces = append(workspaces, w)
	}

	return workspaces, total, nil
}

// AddMember adds member to workspace
func (s *Store) AddMember(ctx context.Context, member *types.Member) error {
	query := `
		INSERT INTO members (id, workspace_id, user_id, email, name, role, vote_weight, joined_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := s.pool.Exec(ctx, query,
		member.ID, member.WorkspaceID, member.UserID, member.Email,
		member.Name, member.Role, member.VoteWeight, member.JoinedAt,
	)
	return err
}

// ListMembers returns workspace members
func (s *Store) ListMembers(ctx context.Context, workspaceID string) ([]types.Member, error) {
	query := `
		SELECT id, workspace_id, user_id, email, name, role, vote_weight, joined_at
		FROM members
		WHERE workspace_id = $1
		ORDER BY joined_at ASC
	`
	rows, err := s.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []types.Member
	for rows.Next() {
		var m types.Member
		if err := rows.Scan(
			&m.ID, &m.WorkspaceID, &m.UserID, &m.Email,
			&m.Name, &m.Role, &m.VoteWeight, &m.JoinedAt,
		); err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	return members, nil
}

// RemoveMember removes member from workspace
func (s *Store) RemoveMember(ctx context.Context, workspaceID, memberID string) error {
	query := `DELETE FROM members WHERE workspace_id = $1 AND id = $2`
	_, err := s.pool.Exec(ctx, query, workspaceID, memberID)
	return err
}
