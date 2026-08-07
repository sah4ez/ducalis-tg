package adapter

import (
	"context"
	"testing"

	"github.com/sah4ez/ducalis-tg/internal/service"
	"github.com/sah4ez/ducalis-tg/pkg/types"
	"github.com/rs/zerolog"
)

func TestWorkspaceAdapter_Create(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Logger()
	adapter := NewWorkspaceAdapter(service.NewWorkspaceService(
		newStubWorkspaceRepo(), newStubMemberRepo(), newStubUserRepo(), logger))

	ctx := context.WithValue(context.Background(), service.UserIDKey, "user-1")
	ws, err := adapter.Create(ctx, types.CreateWorkspaceRequest{Name: "Test WS"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.Name != "Test WS" {
		t.Errorf("expected name Test WS, got %s", ws.Name)
	}
	if ws.OwnerID != "user-1" {
		t.Errorf("expected owner user-1, got %s", ws.OwnerID)
	}
}

func TestWorkspaceAdapter_List(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Logger()
	adapter := NewWorkspaceAdapter(service.NewWorkspaceService(
		newStubWorkspaceRepo(), newStubMemberRepo(), newStubUserRepo(), logger))

	workspaces, total, err := adapter.List(context.Background(), "user-1", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("expected 0 workspaces, got %d", total)
	}
	if len(workspaces) != 0 {
		t.Errorf("expected empty list, got %d items", len(workspaces))
	}
}

// stub workspace repo
type stubWorkspaceRepo struct {
	workspaces map[string]*types.Workspace
}

func newStubWorkspaceRepo() *stubWorkspaceRepo {
	return &stubWorkspaceRepo{workspaces: make(map[string]*types.Workspace)}
}

func (r *stubWorkspaceRepo) Create(ctx context.Context, ws *types.Workspace) error {
	r.workspaces[ws.ID] = ws
	return nil
}

func (r *stubWorkspaceRepo) Get(ctx context.Context, id string) (*types.Workspace, error) {
	if ws, ok := r.workspaces[id]; ok {
		return ws, nil
	}
	return nil, service.ErrNotFound
}

func (r *stubWorkspaceRepo) Update(ctx context.Context, ws *types.Workspace) error {
	r.workspaces[ws.ID] = ws
	return nil
}

func (r *stubWorkspaceRepo) Delete(ctx context.Context, id string) error {
	delete(r.workspaces, id)
	return nil
}

func (r *stubWorkspaceRepo) List(ctx context.Context, userID string, limit, offset int) ([]types.Workspace, int, error) {
	var result []types.Workspace
	for _, ws := range r.workspaces {
		result = append(result, *ws)
	}
	return result, len(result), nil
}

// stub member repo
type stubMemberRepo struct {
	members map[string]*types.Member
}

func newStubMemberRepo() *stubMemberRepo {
	return &stubMemberRepo{members: make(map[string]*types.Member)}
}

func (r *stubMemberRepo) Add(ctx context.Context, m *types.Member) error {
	r.members[m.ID] = m
	return nil
}

func (r *stubMemberRepo) List(ctx context.Context, workspaceID string) ([]types.Member, error) {
	var result []types.Member
	for _, m := range r.members {
		if m.WorkspaceID == workspaceID {
			result = append(result, *m)
		}
	}
	return result, nil
}

func (r *stubMemberRepo) Remove(ctx context.Context, workspaceID, memberID string) error {
	for id, m := range r.members {
		if m.WorkspaceID == workspaceID && m.ID == memberID {
			delete(r.members, id)
			break
		}
	}
	return nil
}
