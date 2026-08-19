package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/sah4ez/ducalis-tg/pkg/types"
)

// Common errors
var (
	ErrNotFound      = errors.New("entity not found")
	ErrAlreadyExists = errors.New("entity already exists")
)

// NotFoundError represents not found error
type NotFoundError struct {
	Entity string
	ID     string
}

func (e *NotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("%s with id '%s' not found", e.Entity, e.ID)
	}
	return fmt.Sprintf("%s not found", e.Entity)
}

// ===== USER REPOSITORY =====

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *types.User, passwordHash string) error {
	query := `
		INSERT INTO users (id, email, name, password_hash, avatar_url, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Pool().Exec(ctx, query,
		user.ID, user.Email, user.Name, passwordHash, user.AvatarURL, user.CreatedAt,
	)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*types.User, error) {
	query := `SELECT id, email, name, password_hash, avatar_url, created_at FROM users WHERE id = $1`
	user := &types.User{}
	var passwordHash string
	var avatarURL sql.NullString

	err := r.db.Pool().QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Name, &passwordHash, &avatarURL, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &NotFoundError{Entity: "user", ID: id}
		}
		return nil, err
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*types.User, string, error) {
	query := `SELECT id, email, name, password_hash, avatar_url, created_at FROM users WHERE email = $1`
	user := &types.User{}
	var passwordHash string
	var avatarURL sql.NullString

	err := r.db.Pool().QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Name, &passwordHash, &avatarURL, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", &NotFoundError{Entity: "user", ID: email}
		}
		return nil, "", err
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}
	return user, passwordHash, nil
}

// ===== WORKSPACE REPOSITORY =====

type WorkspaceRepository struct {
	db *DB
}

func NewWorkspaceRepository(db *DB) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

func (r *WorkspaceRepository) Create(ctx context.Context, workspace *types.Workspace) error {
	scoringJSON, _ := json.Marshal(workspace.Scoring)
	query := `
		INSERT INTO workspaces (id, name, description, owner_id, scoring_config, created_at, updated_at)
		VALUES ($1, $2, NULLIF($3,''), $4, $5::jsonb, $6, $7)
	`
	_, err := r.db.Pool().Exec(ctx, query,
		workspace.ID, workspace.Name, workspace.Description, workspace.OwnerID,
		scoringJSON, workspace.CreatedAt, workspace.UpdatedAt,
	)
	return err
}

func (r *WorkspaceRepository) Get(ctx context.Context, id string) (*types.Workspace, error) {
	query := `SELECT id, name, COALESCE(description,''), owner_id, scoring_config, created_at, updated_at FROM workspaces WHERE id = $1`
	workspace := &types.Workspace{}
	var scoringJSON []byte

	err := r.db.Pool().QueryRow(ctx, query, id).Scan(
		&workspace.ID, &workspace.Name, &workspace.Description, &workspace.OwnerID,
		&scoringJSON, &workspace.CreatedAt, &workspace.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &NotFoundError{Entity: "workspace", ID: id}
		}
		return nil, err
	}
	if len(scoringJSON) > 0 {
		json.Unmarshal(scoringJSON, &workspace.Scoring)
	}
	return workspace, nil
}

func (r *WorkspaceRepository) Update(ctx context.Context, workspace *types.Workspace) error {
	scoringJSON, _ := json.Marshal(workspace.Scoring)
	query := `UPDATE workspaces SET name = $2, description = NULLIF($3,''), scoring_config = $4::jsonb, updated_at = $5 WHERE id = $1`
	result, err := r.db.Pool().Exec(ctx, query,
		workspace.ID, workspace.Name, workspace.Description, scoringJSON, workspace.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &NotFoundError{Entity: "workspace", ID: workspace.ID}
	}
	return nil
}

func (r *WorkspaceRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Pool().Exec(ctx, `DELETE FROM workspaces WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &NotFoundError{Entity: "workspace", ID: id}
	}
	return nil
}

func (r *WorkspaceRepository) List(ctx context.Context, userID string, limit, offset int) ([]types.Workspace, int, error) {
	var total int
	err := r.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM workspaces w
		INNER JOIN members m ON m.workspace_id = w.id
		WHERE m.user_id = $1
	`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Pool().Query(ctx, `
		SELECT w.id, w.name, COALESCE(w.description,''), w.owner_id, w.scoring_config, w.created_at, w.updated_at
		FROM workspaces w
		INNER JOIN members m ON m.workspace_id = w.id
		WHERE m.user_id = $1
		ORDER BY w.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var workspaces []types.Workspace
	for rows.Next() {
		var w types.Workspace
		var scoringJSON []byte
		if err := rows.Scan(
			&w.ID, &w.Name, &w.Description, &w.OwnerID, &scoringJSON, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if len(scoringJSON) > 0 {
			json.Unmarshal(scoringJSON, &w.Scoring)
		}
		workspaces = append(workspaces, w)
	}

	if workspaces == nil {
		workspaces = []types.Workspace{}
	}
	return workspaces, total, nil
}

// ===== MEMBER REPOSITORY =====

type MemberRepository struct {
	db *DB
}

func NewMemberRepository(db *DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) Add(ctx context.Context, member *types.Member) error {
	query := `
		INSERT INTO members (id, workspace_id, user_id, email, name, role, vote_weight, joined_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Pool().Exec(ctx, query,
		member.ID, member.WorkspaceID, member.UserID, member.Email,
		member.Name, member.Role, member.VoteWeight, member.JoinedAt,
	)
	return err
}

func (r *MemberRepository) List(ctx context.Context, workspaceID string) ([]types.Member, error) {
	rows, err := r.db.Pool().Query(ctx, `
		SELECT id, workspace_id, user_id, email, name, role, vote_weight, joined_at
		FROM members WHERE workspace_id = $1 ORDER BY joined_at ASC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []types.Member
	for rows.Next() {
		var m types.Member
		if err := rows.Scan(
			&m.ID, &m.WorkspaceID, &m.UserID, &m.Email, &m.Name, &m.Role, &m.VoteWeight, &m.JoinedAt,
		); err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	if members == nil {
		members = []types.Member{}
	}
	return members, nil
}

func (r *MemberRepository) Remove(ctx context.Context, workspaceID, memberID string) error {
	result, err := r.db.Pool().Exec(ctx, `DELETE FROM members WHERE workspace_id = $1 AND id = $2`, workspaceID, memberID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &NotFoundError{Entity: "member", ID: memberID}
	}
	return nil
}

// ===== TASK REPOSITORY =====

type TaskRepository struct {
	db *DB
}

func NewTaskRepository(db *DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *types.Task) error {
	scoresJSON, _ := json.Marshal(task.Scores)
	metadataJSON, _ := json.Marshal(task.Metadata)

	// NULLIF/COALESCE: пустые строки в nullable uuid/text колонках невалидны
	// ("" не кастуется в uuid), а NULL не сканируется в string — нормализуем
	// на границе SQL.
	query := `
		INSERT INTO tasks (
			id, workspace_id, external_id, external_type, external_url,
			title, description, scores, final_score, status, priority,
			labels, assignee_id, created_by, created_at, updated_at, metadata
		) VALUES ($1, $2, NULLIF($3,''), NULLIF($4,''), NULLIF($5,''), $6, NULLIF($7,''),
			$8::jsonb, $9, $10, NULLIF($11,''), $12, NULLIF($13,'')::uuid, $14, $15, $16, $17::jsonb)
	`
	_, err := r.db.Pool().Exec(ctx, query,
		task.ID, task.WorkspaceID, task.ExternalID, task.ExternalType, task.ExternalURL,
		task.Title, task.Description, scoresJSON, task.FinalScore, task.Status, task.Priority,
		task.Labels, task.AssigneeID, task.CreatedBy, task.CreatedAt, task.UpdatedAt, metadataJSON,
	)
	return err
}

func (r *TaskRepository) Get(ctx context.Context, id string) (*types.Task, error) {
	query := `
		SELECT id, workspace_id,
			COALESCE(external_id::text,''), COALESCE(external_type,''), COALESCE(external_url,''),
			title, COALESCE(description,''), scores, final_score, COALESCE(status,''), COALESCE(priority,''),
			labels, COALESCE(assignee_id::text,''), created_by, created_at, updated_at, metadata
		FROM tasks WHERE id = $1
	`
	task := &types.Task{}
	var scoresJSON, metadataJSON []byte

	err := r.db.Pool().QueryRow(ctx, query, id).Scan(
		&task.ID, &task.WorkspaceID, &task.ExternalID, &task.ExternalType, &task.ExternalURL,
		&task.Title, &task.Description, &scoresJSON, &task.FinalScore, &task.Status, &task.Priority,
		&task.Labels, &task.AssigneeID, &task.CreatedBy, &task.CreatedAt, &task.UpdatedAt, &metadataJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &NotFoundError{Entity: "task", ID: id}
		}
		return nil, err
	}
	if len(scoresJSON) > 0 {
		json.Unmarshal(scoresJSON, &task.Scores)
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &task.Metadata)
	}
	return task, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *types.Task) error {
	scoresJSON, _ := json.Marshal(task.Scores)
	metadataJSON, _ := json.Marshal(task.Metadata)

	query := `
		UPDATE tasks SET title = $2, description = NULLIF($3,''), scores = $4::jsonb, final_score = $5,
			status = $6, priority = NULLIF($7,''), labels = $8, assignee_id = NULLIF($9,'')::uuid,
			updated_at = $10, metadata = $11::jsonb
		WHERE id = $1
	`
	result, err := r.db.Pool().Exec(ctx, query,
		task.ID, task.Title, task.Description, scoresJSON, task.FinalScore,
		task.Status, task.Priority, task.Labels, task.AssigneeID, task.UpdatedAt, metadataJSON,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &NotFoundError{Entity: "task", ID: task.ID}
	}
	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Pool().Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &NotFoundError{Entity: "task", ID: id}
	}
	return nil
}

func (r *TaskRepository) List(ctx context.Context, req types.ListTasksRequest) ([]types.Task, int, error) {
	baseQuery := `FROM tasks WHERE workspace_id = $1`
	args := []interface{}{req.WorkspaceID}
	argIdx := 2

	if req.Status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, req.Status)
		argIdx++
	}
	if req.AssigneeID != "" {
		baseQuery += fmt.Sprintf(" AND assignee_id = $%d", argIdx)
		args = append(args, req.AssigneeID)
		argIdx++
	}
	if len(req.Labels) > 0 {
		baseQuery += fmt.Sprintf(" AND labels && $%d", argIdx)
		args = append(args, req.Labels)
		argIdx++
	}
	if req.Search != "" {
		baseQuery += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+req.Search+"%")
		argIdx++
	}

	var total int
	err := r.db.Pool().QueryRow(ctx, `SELECT COUNT(*) `+baseQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	query := `SELECT id, workspace_id, COALESCE(external_id::text,''), COALESCE(external_type,''), COALESCE(external_url,''), title, COALESCE(description,''), scores, final_score, COALESCE(status,''), COALESCE(priority,''), labels, COALESCE(assignee_id::text,''), created_by, created_at, updated_at, metadata ` +
		baseQuery + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", argIdx) + ` OFFSET $` + fmt.Sprintf("%d", argIdx+1)
	args = append(args, limit, req.Offset)

	rows, err := r.db.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []types.Task
	for rows.Next() {
		var t types.Task
		var scoresJSON, metadataJSON []byte
		if err := rows.Scan(
			&t.ID, &t.WorkspaceID, &t.ExternalID, &t.ExternalType, &t.ExternalURL,
			&t.Title, &t.Description, &scoresJSON, &t.FinalScore, &t.Status, &t.Priority,
			&t.Labels, &t.AssigneeID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt, &metadataJSON,
		); err != nil {
			return nil, 0, err
		}
		if len(scoresJSON) > 0 {
			json.Unmarshal(scoresJSON, &t.Scores)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &t.Metadata)
		}
		tasks = append(tasks, t)
	}

	if tasks == nil {
		tasks = []types.Task{}
	}
	return tasks, total, nil
}

func (r *TaskRepository) GetRanked(ctx context.Context, workspaceID string, limit int, offset int) ([]types.TaskWithRank, error) {
	query := `
		SELECT id, workspace_id,
			COALESCE(external_id::text,''), COALESCE(external_type,''), COALESCE(external_url,''),
			title, COALESCE(description,''), scores, final_score, COALESCE(status,''), COALESCE(priority,''),
			labels, COALESCE(assignee_id::text,''), created_by, created_at, updated_at, metadata,
			COALESCE(rank,0), COALESCE(percentile,0)
		FROM ranked_tasks
		WHERE workspace_id = $1
		ORDER BY rank ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Pool().Query(ctx, query, workspaceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []types.TaskWithRank
	for rows.Next() {
		var t types.TaskWithRank
		var scoresJSON, metadataJSON []byte
		if err := rows.Scan(
			&t.ID, &t.WorkspaceID,
			&t.ExternalID, &t.ExternalType, &t.ExternalURL,
			&t.Title, &t.Description, &scoresJSON, &t.FinalScore, &t.Status, &t.Priority,
			&t.Labels, &t.AssigneeID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt, &metadataJSON,
			&t.Rank, &t.Percentile,
		); err != nil {
			return nil, err
		}
		if len(scoresJSON) > 0 {
			json.Unmarshal(scoresJSON, &t.Scores)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &t.Metadata)
		}
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []types.TaskWithRank{}
	}
	return tasks, nil
}

// ===== VOTE REPOSITORY =====

type VoteRepository struct {
	db *DB
}

func NewVoteRepository(db *DB) *VoteRepository {
	return &VoteRepository{db: db}
}

func (r *VoteRepository) Add(ctx context.Context, taskID string, userID string, weight float64) error {
	query := `
		INSERT INTO votes (id, task_id, user_id, weight, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
		ON CONFLICT (task_id, user_id) DO UPDATE SET weight = $3, created_at = NOW()
	`
	_, err := r.db.Pool().Exec(ctx, query, taskID, userID, weight)
	return err
}

func (r *VoteRepository) Remove(ctx context.Context, taskID string, userID string) error {
	result, err := r.db.Pool().Exec(ctx, `DELETE FROM votes WHERE task_id = $1 AND user_id = $2`, taskID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &NotFoundError{Entity: "vote", ID: taskID + "/" + userID}
	}
	return nil
}

func (r *VoteRepository) GetByTask(ctx context.Context, taskID string) ([]types.Vote, error) {
	rows, err := r.db.Pool().Query(ctx, `SELECT user_id, weight, created_at FROM votes WHERE task_id = $1`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votes []types.Vote
	for rows.Next() {
		var v types.Vote
		if err := rows.Scan(&v.UserID, &v.Weight, &v.CreatedAt); err != nil {
			return nil, err
		}
		votes = append(votes, v)
	}

	if votes == nil {
		votes = []types.Vote{}
	}
	return votes, nil
}

// ===== ESTIMATION REPOSITORY =====

type EstimationRepository struct {
	db *DB
}

func NewEstimationRepository(db *DB) *EstimationRepository {
	return &EstimationRepository{db: db}
}

func (r *EstimationRepository) Add(ctx context.Context, taskID string, userID string, value float64, unit string) error {
	query := `
		INSERT INTO estimations (id, task_id, user_id, value, unit, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (task_id, user_id) DO UPDATE SET value = $3, unit = $4, updated_at = NOW()
	`
	_, err := r.db.Pool().Exec(ctx, query, taskID, userID, value, unit)
	return err
}

func (r *EstimationRepository) GetByTask(ctx context.Context, taskID string) ([]types.Estimation, error) {
	rows, err := r.db.Pool().Query(ctx, `SELECT user_id, value, unit, created_at FROM estimations WHERE task_id = $1`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var estimations []types.Estimation
	for rows.Next() {
		var e types.Estimation
		if err := rows.Scan(&e.UserID, &e.Value, &e.Unit, &e.CreatedAt); err != nil {
			return nil, err
		}
		estimations = append(estimations, e)
	}

	if estimations == nil {
		estimations = []types.Estimation{}
	}
	return estimations, nil
}

// ===== INTEGRATION REPOSITORY =====

type IntegrationRepository struct {
	db *DB
}

func NewIntegrationRepository(db *DB) *IntegrationRepository {
	return &IntegrationRepository{db: db}
}

func (r *IntegrationRepository) Create(ctx context.Context, integration *types.Integration) error {
	configJSON, _ := json.Marshal(integration.Config)
	query := `
		INSERT INTO integrations (id, workspace_id, type, name, config, auto_sync, sync_interval, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	_, err := r.db.Pool().Exec(ctx, query,
		integration.ID, integration.WorkspaceID, integration.Type, integration.Name,
		configJSON, integration.AutoSync, integration.SyncInterval,
	)
	return err
}

func (r *IntegrationRepository) Get(ctx context.Context, id string) (*types.Integration, error) {
	query := `
		SELECT id, workspace_id, type, name, config, last_sync_at, sync_status, sync_error, auto_sync, sync_interval, created_at, updated_at
		FROM integrations WHERE id = $1
	`
	integration := &types.Integration{}
	var configJSON []byte

	err := r.db.Pool().QueryRow(ctx, query, id).Scan(
		&integration.ID, &integration.WorkspaceID, &integration.Type, &integration.Name, &configJSON,
		&integration.LastSyncAt, &integration.SyncStatus, &integration.SyncError,
		&integration.AutoSync, &integration.SyncInterval, &integration.CreatedAt, &integration.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &NotFoundError{Entity: "integration", ID: id}
		}
		return nil, err
	}
	if len(configJSON) > 0 {
		json.Unmarshal(configJSON, &integration.Config)
	}
	return integration, nil
}

func (r *IntegrationRepository) List(ctx context.Context, workspaceID string) ([]types.Integration, error) {
	rows, err := r.db.Pool().Query(ctx, `
		SELECT id, workspace_id, type, name, config, last_sync_at, sync_status, sync_error, auto_sync, sync_interval, created_at, updated_at
		FROM integrations WHERE workspace_id = $1 ORDER BY created_at DESC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var integrations []types.Integration
	for rows.Next() {
		var i types.Integration
		var configJSON []byte
		if err := rows.Scan(
			&i.ID, &i.WorkspaceID, &i.Type, &i.Name, &configJSON,
			&i.LastSyncAt, &i.SyncStatus, &i.SyncError,
			&i.AutoSync, &i.SyncInterval, &i.CreatedAt, &i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if len(configJSON) > 0 {
			json.Unmarshal(configJSON, &i.Config)
		}
		integrations = append(integrations, i)
	}

	if integrations == nil {
		integrations = []types.Integration{}
	}
	return integrations, nil
}

func (r *IntegrationRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Pool().Exec(ctx, `DELETE FROM integrations WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return &NotFoundError{Entity: "integration", ID: id}
	}
	return nil
}
