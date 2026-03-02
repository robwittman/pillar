package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/mock"
	"github.com/robwittman/pillar/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService(repo *mock.AgentRepository, status *mock.AgentStatusStore, opts ...service.AgentServiceOption) service.AgentService {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return service.NewAgentService(repo, status, logger, opts...)
}

// --- Create ---

func TestCreate_Success(t *testing.T) {
	repo := &mock.AgentRepository{
		CreateFn: func(ctx context.Context, agent *domain.Agent) error {
			return nil
		},
	}
	status := &mock.AgentStatusStore{}
	svc := newTestService(repo, status)

	agent, err := svc.Create(context.Background(), "test-agent", nil, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, agent.ID)
	assert.Equal(t, "test-agent", agent.Name)
	assert.Equal(t, domain.AgentStatusPending, agent.Status)
	assert.NotNil(t, agent.Metadata)
	assert.NotNil(t, agent.Labels)
}

func TestCreate_NilMapsBecomEmpty(t *testing.T) {
	var captured *domain.Agent
	repo := &mock.AgentRepository{
		CreateFn: func(ctx context.Context, agent *domain.Agent) error {
			captured = agent
			return nil
		},
	}
	status := &mock.AgentStatusStore{}
	svc := newTestService(repo, status)

	_, err := svc.Create(context.Background(), "a", nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, captured.Metadata)
	assert.NotNil(t, captured.Labels)
}

func TestCreate_RepoError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &mock.AgentRepository{
		CreateFn: func(ctx context.Context, agent *domain.Agent) error {
			return repoErr
		},
	}
	status := &mock.AgentStatusStore{}
	svc := newTestService(repo, status)

	_, err := svc.Create(context.Background(), "x", nil, nil)
	assert.ErrorIs(t, err, repoErr)
}

// --- Get ---

func TestGet_Success(t *testing.T) {
	expected := &domain.Agent{ID: "abc", Name: "agent1"}
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			assert.Equal(t, "abc", id)
			return expected, nil
		},
	}
	status := &mock.AgentStatusStore{}
	svc := newTestService(repo, status)

	agent, err := svc.Get(context.Background(), "abc")
	require.NoError(t, err)
	assert.Equal(t, expected, agent)
}

func TestGet_NotFound(t *testing.T) {
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return nil, domain.ErrAgentNotFound
		},
	}
	status := &mock.AgentStatusStore{}
	svc := newTestService(repo, status)

	_, err := svc.Get(context.Background(), "missing")
	assert.ErrorIs(t, err, domain.ErrAgentNotFound)
}

// --- List ---

func TestList_Success(t *testing.T) {
	expected := []*domain.Agent{{ID: "1"}, {ID: "2"}}
	repo := &mock.AgentRepository{
		ListFn: func(ctx context.Context) ([]*domain.Agent, error) {
			return expected, nil
		},
	}
	status := &mock.AgentStatusStore{}
	svc := newTestService(repo, status)

	agents, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Equal(t, expected, agents)
}

// --- Update ---

func TestUpdate_MergesFields(t *testing.T) {
	existing := &domain.Agent{
		ID:       "id1",
		Name:     "old",
		Metadata: map[string]string{"k": "v"},
		Labels:   map[string]string{"l": "v"},
	}
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return existing, nil
		},
		UpdateFn: func(ctx context.Context, agent *domain.Agent) error {
			return nil
		},
	}
	status := &mock.AgentStatusStore{}
	svc := newTestService(repo, status)

	newMeta := map[string]string{"new": "meta"}
	agent, err := svc.Update(context.Background(), "id1", "new-name", newMeta, nil)
	require.NoError(t, err)
	assert.Equal(t, "new-name", agent.Name)
	assert.Equal(t, newMeta, agent.Metadata)
	assert.Equal(t, map[string]string{"l": "v"}, agent.Labels)
}

func TestUpdate_NotFound(t *testing.T) {
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return nil, domain.ErrAgentNotFound
		},
	}
	status := &mock.AgentStatusStore{}
	svc := newTestService(repo, status)

	_, err := svc.Update(context.Background(), "missing", "x", nil, nil)
	assert.ErrorIs(t, err, domain.ErrAgentNotFound)
}

// --- Delete ---

func TestDelete_SetOfflineFailureStillDeletes(t *testing.T) {
	repo := &mock.AgentRepository{
		DeleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	status := &mock.AgentStatusStore{
		SetOfflineFn: func(ctx context.Context, agentID string) error {
			return errors.New("redis down")
		},
	}
	svc := newTestService(repo, status)

	err := svc.Delete(context.Background(), "id1")
	assert.NoError(t, err)
}

func TestDelete_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	repo := &mock.AgentRepository{
		DeleteFn: func(ctx context.Context, id string) error {
			return repoErr
		},
	}
	status := &mock.AgentStatusStore{
		SetOfflineFn: func(ctx context.Context, agentID string) error {
			return nil
		},
	}
	svc := newTestService(repo, status)

	err := svc.Delete(context.Background(), "id1")
	assert.ErrorIs(t, err, repoErr)
}

// --- Start ---

func TestStart_FromPending(t *testing.T) {
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusPending}, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status domain.AgentStatus) error {
			assert.Equal(t, domain.AgentStatusRunning, status)
			return nil
		},
	}
	status := &mock.AgentStatusStore{}
	svc := newTestService(repo, status)

	err := svc.Start(context.Background(), "id1")
	assert.NoError(t, err)
}

func TestStart_FromStopped(t *testing.T) {
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusStopped}, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status domain.AgentStatus) error {
			return nil
		},
	}
	status := &mock.AgentStatusStore{}
	svc := newTestService(repo, status)

	err := svc.Start(context.Background(), "id1")
	assert.NoError(t, err)
}

func TestStart_FromRunning_InvalidTransition(t *testing.T) {
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusRunning}, nil
		},
	}
	status := &mock.AgentStatusStore{}
	svc := newTestService(repo, status)

	err := svc.Start(context.Background(), "id1")
	assert.ErrorIs(t, err, domain.ErrInvalidTransition)
}

func TestStart_NotFound(t *testing.T) {
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return nil, domain.ErrAgentNotFound
		},
	}
	status := &mock.AgentStatusStore{}
	svc := newTestService(repo, status)

	err := svc.Start(context.Background(), "missing")
	assert.ErrorIs(t, err, domain.ErrAgentNotFound)
}

// --- Stop ---

func TestStop_FromRunning(t *testing.T) {
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusRunning}, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status domain.AgentStatus) error {
			assert.Equal(t, domain.AgentStatusStopped, status)
			return nil
		},
	}
	status := &mock.AgentStatusStore{
		SetOfflineFn: func(ctx context.Context, agentID string) error {
			return nil
		},
	}
	svc := newTestService(repo, status)

	err := svc.Stop(context.Background(), "id1")
	assert.NoError(t, err)
}

func TestStop_FromPending_InvalidTransition(t *testing.T) {
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusPending}, nil
		},
	}
	status := &mock.AgentStatusStore{}
	svc := newTestService(repo, status)

	err := svc.Stop(context.Background(), "id1")
	assert.ErrorIs(t, err, domain.ErrInvalidTransition)
}

func TestStop_SetOfflineFailureNonFatal(t *testing.T) {
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusRunning}, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status domain.AgentStatus) error {
			return nil
		},
	}
	status := &mock.AgentStatusStore{
		SetOfflineFn: func(ctx context.Context, agentID string) error {
			return errors.New("redis down")
		},
	}
	svc := newTestService(repo, status)

	err := svc.Stop(context.Background(), "id1")
	assert.NoError(t, err)
}

// --- Status ---

func TestStatus_CombinesRepoAndRedis(t *testing.T) {
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusRunning}, nil
		},
	}
	statusStore := &mock.AgentStatusStore{
		IsOnlineFn: func(ctx context.Context, agentID string) (bool, error) {
			return true, nil
		},
	}
	svc := newTestService(repo, statusStore)

	info, err := svc.Status(context.Background(), "id1")
	require.NoError(t, err)
	assert.Equal(t, "id1", info.AgentID)
	assert.Equal(t, "running", info.Status)
	assert.True(t, info.Online)
}

func TestStatus_IsOnlineFailureNonFatal(t *testing.T) {
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusPending}, nil
		},
	}
	statusStore := &mock.AgentStatusStore{
		IsOnlineFn: func(ctx context.Context, agentID string) (bool, error) {
			return false, errors.New("redis down")
		},
	}
	svc := newTestService(repo, statusStore)

	info, err := svc.Status(context.Background(), "id1")
	require.NoError(t, err)
	assert.False(t, info.Online)
}

func TestStatus_NotFound(t *testing.T) {
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return nil, domain.ErrAgentNotFound
		},
	}
	statusStore := &mock.AgentStatusStore{}
	svc := newTestService(repo, statusStore)

	_, err := svc.Status(context.Background(), "missing")
	assert.ErrorIs(t, err, domain.ErrAgentNotFound)
}

// --- Heartbeat ---

func TestHeartbeat_Success(t *testing.T) {
	var heartbeatCalled, onlineCalled bool
	statusStore := &mock.AgentStatusStore{
		SetHeartbeatFn: func(ctx context.Context, agentID string, ttl time.Duration) error {
			heartbeatCalled = true
			assert.Equal(t, "agent1", agentID)
			assert.Equal(t, 30*time.Second, ttl)
			return nil
		},
		SetOnlineFn: func(ctx context.Context, agentID string) error {
			onlineCalled = true
			return nil
		},
	}
	repo := &mock.AgentRepository{}
	svc := newTestService(repo, statusStore)

	err := svc.Heartbeat(context.Background(), "agent1")
	assert.NoError(t, err)
	assert.True(t, heartbeatCalled)
	assert.True(t, onlineCalled)
}

func TestHeartbeat_SetHeartbeatError(t *testing.T) {
	hbErr := errors.New("redis error")
	statusStore := &mock.AgentStatusStore{
		SetHeartbeatFn: func(ctx context.Context, agentID string, ttl time.Duration) error {
			return hbErr
		},
	}
	repo := &mock.AgentRepository{}
	svc := newTestService(repo, statusStore)

	err := svc.Heartbeat(context.Background(), "agent1")
	assert.ErrorIs(t, err, hbErr)
}

func TestHeartbeat_SetOnlineError(t *testing.T) {
	onlineErr := errors.New("redis error")
	statusStore := &mock.AgentStatusStore{
		SetHeartbeatFn: func(ctx context.Context, agentID string, ttl time.Duration) error {
			return nil
		},
		SetOnlineFn: func(ctx context.Context, agentID string) error {
			return onlineErr
		},
	}
	repo := &mock.AgentRepository{}
	svc := newTestService(repo, statusStore)

	err := svc.Heartbeat(context.Background(), "agent1")
	assert.ErrorIs(t, err, onlineErr)
}

// --- Start/Stop Notification ---

func TestStart_NotifiesOnSuccess(t *testing.T) {
	var notifiedID, notifiedType string
	notifier := &mock.AgentNotifier{
		NotifyDirectiveFn: func(agentID, directiveType, payload string) error {
			notifiedID = agentID
			notifiedType = directiveType
			return nil
		},
	}
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusPending}, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status domain.AgentStatus) error {
			return nil
		},
	}
	svc := newTestService(repo, &mock.AgentStatusStore{}, service.WithNotifier(notifier))

	err := svc.Start(context.Background(), "agent1")
	assert.NoError(t, err)
	assert.Equal(t, "agent1", notifiedID)
	assert.Equal(t, service.DirectiveStart, notifiedType)
}

func TestStart_NotifyFailureNonFatal(t *testing.T) {
	notifier := &mock.AgentNotifier{
		NotifyDirectiveFn: func(agentID, directiveType, payload string) error {
			return errors.New("stream broken")
		},
	}
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusPending}, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status domain.AgentStatus) error {
			return nil
		},
	}
	svc := newTestService(repo, &mock.AgentStatusStore{}, service.WithNotifier(notifier))

	err := svc.Start(context.Background(), "agent1")
	assert.NoError(t, err)
}

func TestStop_NotifiesOnSuccess(t *testing.T) {
	var notifiedID, notifiedType string
	notifier := &mock.AgentNotifier{
		NotifyDirectiveFn: func(agentID, directiveType, payload string) error {
			notifiedID = agentID
			notifiedType = directiveType
			return nil
		},
	}
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusRunning}, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status domain.AgentStatus) error {
			return nil
		},
	}
	status := &mock.AgentStatusStore{
		SetOfflineFn: func(ctx context.Context, agentID string) error {
			return nil
		},
	}
	svc := newTestService(repo, status, service.WithNotifier(notifier))

	err := svc.Stop(context.Background(), "agent1")
	assert.NoError(t, err)
	assert.Equal(t, "agent1", notifiedID)
	assert.Equal(t, service.DirectiveStop, notifiedType)
}

func TestStop_NotifyFailureNonFatal(t *testing.T) {
	notifier := &mock.AgentNotifier{
		NotifyDirectiveFn: func(agentID, directiveType, payload string) error {
			return errors.New("stream broken")
		},
	}
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusRunning}, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status domain.AgentStatus) error {
			return nil
		},
	}
	status := &mock.AgentStatusStore{
		SetOfflineFn: func(ctx context.Context, agentID string) error {
			return nil
		},
	}
	svc := newTestService(repo, status, service.WithNotifier(notifier))

	err := svc.Stop(context.Background(), "agent1")
	assert.NoError(t, err)
}

func TestStart_NilNotifier(t *testing.T) {
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusPending}, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status domain.AgentStatus) error {
			return nil
		},
	}
	svc := newTestService(repo, &mock.AgentStatusStore{})

	err := svc.Start(context.Background(), "agent1")
	assert.NoError(t, err)
}

// --- Runtime Integration ---

func TestStart_CallsRuntimeEnsureRunning(t *testing.T) {
	var calledWith string
	rt := &mock.AgentRuntime{
		EnsureRunningFn: func(ctx context.Context, agentID string) error {
			calledWith = agentID
			return nil
		},
	}
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusPending}, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status domain.AgentStatus) error {
			return nil
		},
	}
	svc := newTestService(repo, &mock.AgentStatusStore{}, service.WithRuntime(rt))

	err := svc.Start(context.Background(), "agent1")
	assert.NoError(t, err)
	assert.Equal(t, "agent1", calledWith)
}

func TestStart_RuntimeFailureNonFatal(t *testing.T) {
	rt := &mock.AgentRuntime{
		EnsureRunningFn: func(ctx context.Context, agentID string) error {
			return errors.New("kube down")
		},
	}
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusPending}, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status domain.AgentStatus) error {
			return nil
		},
	}
	svc := newTestService(repo, &mock.AgentStatusStore{}, service.WithRuntime(rt))

	err := svc.Start(context.Background(), "agent1")
	assert.NoError(t, err)
}

func TestStart_NilRuntime(t *testing.T) {
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusPending}, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status domain.AgentStatus) error {
			return nil
		},
	}
	svc := newTestService(repo, &mock.AgentStatusStore{})

	err := svc.Start(context.Background(), "agent1")
	assert.NoError(t, err)
}

func TestStop_CallsRuntimeEnsureStopped(t *testing.T) {
	var calledWith string
	rt := &mock.AgentRuntime{
		EnsureStoppedFn: func(ctx context.Context, agentID string) error {
			calledWith = agentID
			return nil
		},
	}
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusRunning}, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status domain.AgentStatus) error {
			return nil
		},
	}
	status := &mock.AgentStatusStore{
		SetOfflineFn: func(ctx context.Context, agentID string) error {
			return nil
		},
	}
	svc := newTestService(repo, status, service.WithRuntime(rt))

	err := svc.Stop(context.Background(), "agent1")
	assert.NoError(t, err)
	assert.Equal(t, "agent1", calledWith)
}

func TestStop_RuntimeFailureNonFatal(t *testing.T) {
	rt := &mock.AgentRuntime{
		EnsureStoppedFn: func(ctx context.Context, agentID string) error {
			return errors.New("kube down")
		},
	}
	repo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Status: domain.AgentStatusRunning}, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status domain.AgentStatus) error {
			return nil
		},
	}
	status := &mock.AgentStatusStore{
		SetOfflineFn: func(ctx context.Context, agentID string) error {
			return nil
		},
	}
	svc := newTestService(repo, status, service.WithRuntime(rt))

	err := svc.Stop(context.Background(), "agent1")
	assert.NoError(t, err)
}

func TestDelete_CallsRuntimeRemove(t *testing.T) {
	var calledWith string
	rt := &mock.AgentRuntime{
		RemoveFn: func(ctx context.Context, agentID string) error {
			calledWith = agentID
			return nil
		},
	}
	repo := &mock.AgentRepository{
		DeleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	status := &mock.AgentStatusStore{
		SetOfflineFn: func(ctx context.Context, agentID string) error {
			return nil
		},
	}
	svc := newTestService(repo, status, service.WithRuntime(rt))

	err := svc.Delete(context.Background(), "agent1")
	assert.NoError(t, err)
	assert.Equal(t, "agent1", calledWith)
}

func TestDelete_RuntimeFailureNonFatal(t *testing.T) {
	rt := &mock.AgentRuntime{
		RemoveFn: func(ctx context.Context, agentID string) error {
			return errors.New("kube down")
		},
	}
	repo := &mock.AgentRepository{
		DeleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	status := &mock.AgentStatusStore{
		SetOfflineFn: func(ctx context.Context, agentID string) error {
			return nil
		},
	}
	svc := newTestService(repo, status, service.WithRuntime(rt))

	err := svc.Delete(context.Background(), "agent1")
	assert.NoError(t, err)
}

func TestDelete_NilRuntime(t *testing.T) {
	repo := &mock.AgentRepository{
		DeleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	status := &mock.AgentStatusStore{
		SetOfflineFn: func(ctx context.Context, agentID string) error {
			return nil
		},
	}
	svc := newTestService(repo, status)

	err := svc.Delete(context.Background(), "agent1")
	assert.NoError(t, err)
}
