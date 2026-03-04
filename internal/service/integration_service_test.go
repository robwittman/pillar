package service_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/mock"
	"github.com/robwittman/pillar/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestIntegrationService(
	integrationRepo *mock.IntegrationRepository,
	agentRepo *mock.AgentRepository,
) service.IntegrationService {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return service.NewIntegrationService(integrationRepo, agentRepo, logger)
}

// --- Create ---

func TestCreateIntegration_Success(t *testing.T) {
	var captured *domain.Integration
	integrationRepo := &mock.IntegrationRepository{
		CreateFn: func(ctx context.Context, integration *domain.Integration) error {
			captured = integration
			return nil
		},
	}
	agentRepo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id}, nil
		},
	}
	svc := newTestIntegrationService(integrationRepo, agentRepo)

	integration := &domain.Integration{
		AgentID: "agent-1",
		Type:    "vault",
		Name:    "prod-vault",
		Config:  map[string]any{"address": "https://vault:8200"},
	}
	err := svc.Create(context.Background(), integration)
	require.NoError(t, err)
	assert.NotEmpty(t, captured.ID)
	assert.Equal(t, "agent-1", captured.AgentID)
	assert.Equal(t, "vault", captured.Type)
	assert.NotNil(t, captured.Config)
}

func TestCreateIntegration_NilConfig(t *testing.T) {
	var captured *domain.Integration
	integrationRepo := &mock.IntegrationRepository{
		CreateFn: func(ctx context.Context, integration *domain.Integration) error {
			captured = integration
			return nil
		},
	}
	agentRepo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id}, nil
		},
	}
	svc := newTestIntegrationService(integrationRepo, agentRepo)

	integration := &domain.Integration{
		AgentID: "agent-1",
		Type:    "vault",
		Name:    "prod-vault",
	}
	err := svc.Create(context.Background(), integration)
	require.NoError(t, err)
	assert.NotNil(t, captured.Config)
}

func TestCreateIntegration_AgentNotFound(t *testing.T) {
	integrationRepo := &mock.IntegrationRepository{}
	agentRepo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return nil, domain.ErrAgentNotFound
		},
	}
	svc := newTestIntegrationService(integrationRepo, agentRepo)

	integration := &domain.Integration{
		AgentID: "missing",
		Type:    "vault",
		Name:    "prod-vault",
	}
	err := svc.Create(context.Background(), integration)
	assert.ErrorIs(t, err, domain.ErrAgentNotFound)
}

func TestCreateIntegration_MissingType(t *testing.T) {
	integrationRepo := &mock.IntegrationRepository{}
	agentRepo := &mock.AgentRepository{}
	svc := newTestIntegrationService(integrationRepo, agentRepo)

	integration := &domain.Integration{
		AgentID: "agent-1",
		Name:    "prod-vault",
	}
	err := svc.Create(context.Background(), integration)
	assert.ErrorIs(t, err, domain.ErrInvalidIntegration)
}

func TestCreateIntegration_MissingName(t *testing.T) {
	integrationRepo := &mock.IntegrationRepository{}
	agentRepo := &mock.AgentRepository{}
	svc := newTestIntegrationService(integrationRepo, agentRepo)

	integration := &domain.Integration{
		AgentID: "agent-1",
		Type:    "vault",
	}
	err := svc.Create(context.Background(), integration)
	assert.ErrorIs(t, err, domain.ErrInvalidIntegration)
}

func TestCreateIntegration_RepoError(t *testing.T) {
	integrationRepo := &mock.IntegrationRepository{
		CreateFn: func(ctx context.Context, integration *domain.Integration) error {
			return domain.ErrInvalidIntegration
		},
	}
	agentRepo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id}, nil
		},
	}
	svc := newTestIntegrationService(integrationRepo, agentRepo)

	integration := &domain.Integration{
		AgentID: "agent-1",
		Type:    "vault",
		Name:    "prod-vault",
	}
	err := svc.Create(context.Background(), integration)
	assert.Error(t, err)
}

// --- Get ---

func TestGetIntegration_Success(t *testing.T) {
	expected := &domain.Integration{ID: "integ-1", AgentID: "agent-1", Type: "vault", Name: "prod-vault"}
	integrationRepo := &mock.IntegrationRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Integration, error) {
			return expected, nil
		},
	}
	agentRepo := &mock.AgentRepository{}
	svc := newTestIntegrationService(integrationRepo, agentRepo)

	integration, err := svc.Get(context.Background(), "integ-1")
	require.NoError(t, err)
	assert.Equal(t, expected, integration)
}

func TestGetIntegration_NotFound(t *testing.T) {
	integrationRepo := &mock.IntegrationRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Integration, error) {
			return nil, domain.ErrIntegrationNotFound
		},
	}
	agentRepo := &mock.AgentRepository{}
	svc := newTestIntegrationService(integrationRepo, agentRepo)

	_, err := svc.Get(context.Background(), "missing")
	assert.ErrorIs(t, err, domain.ErrIntegrationNotFound)
}

// --- List ---

func TestListIntegrations_Success(t *testing.T) {
	expected := []*domain.Integration{
		{ID: "integ-1", AgentID: "agent-1", Type: "vault", Name: "prod-vault"},
		{ID: "integ-2", AgentID: "agent-1", Type: "keystone", Name: "identity"},
	}
	integrationRepo := &mock.IntegrationRepository{
		ListFn: func(ctx context.Context, agentID string) ([]*domain.Integration, error) {
			return expected, nil
		},
	}
	agentRepo := &mock.AgentRepository{}
	svc := newTestIntegrationService(integrationRepo, agentRepo)

	integrations, err := svc.List(context.Background(), "agent-1")
	require.NoError(t, err)
	assert.Len(t, integrations, 2)
}

func TestListIntegrations_Empty(t *testing.T) {
	integrationRepo := &mock.IntegrationRepository{
		ListFn: func(ctx context.Context, agentID string) ([]*domain.Integration, error) {
			return []*domain.Integration{}, nil
		},
	}
	agentRepo := &mock.AgentRepository{}
	svc := newTestIntegrationService(integrationRepo, agentRepo)

	integrations, err := svc.List(context.Background(), "agent-1")
	require.NoError(t, err)
	assert.Empty(t, integrations)
}

// --- Update ---

func TestUpdateIntegration_Success(t *testing.T) {
	integrationRepo := &mock.IntegrationRepository{
		UpdateFn: func(ctx context.Context, integration *domain.Integration) error {
			return nil
		},
	}
	agentRepo := &mock.AgentRepository{}
	svc := newTestIntegrationService(integrationRepo, agentRepo)

	integration := &domain.Integration{
		ID:   "integ-1",
		Name: "updated-name",
		Config: map[string]any{"address": "https://vault:8200"},
	}
	err := svc.Update(context.Background(), integration)
	require.NoError(t, err)
}

func TestUpdateIntegration_MissingName(t *testing.T) {
	integrationRepo := &mock.IntegrationRepository{}
	agentRepo := &mock.AgentRepository{}
	svc := newTestIntegrationService(integrationRepo, agentRepo)

	integration := &domain.Integration{
		ID: "integ-1",
	}
	err := svc.Update(context.Background(), integration)
	assert.ErrorIs(t, err, domain.ErrInvalidIntegration)
}

func TestUpdateIntegration_NilConfig(t *testing.T) {
	var captured *domain.Integration
	integrationRepo := &mock.IntegrationRepository{
		UpdateFn: func(ctx context.Context, integration *domain.Integration) error {
			captured = integration
			return nil
		},
	}
	agentRepo := &mock.AgentRepository{}
	svc := newTestIntegrationService(integrationRepo, agentRepo)

	integration := &domain.Integration{
		ID:   "integ-1",
		Name: "updated-name",
	}
	err := svc.Update(context.Background(), integration)
	require.NoError(t, err)
	assert.NotNil(t, captured.Config)
}

func TestUpdateIntegration_NotFound(t *testing.T) {
	integrationRepo := &mock.IntegrationRepository{
		UpdateFn: func(ctx context.Context, integration *domain.Integration) error {
			return domain.ErrIntegrationNotFound
		},
	}
	agentRepo := &mock.AgentRepository{}
	svc := newTestIntegrationService(integrationRepo, agentRepo)

	integration := &domain.Integration{
		ID:   "missing",
		Name: "updated",
	}
	err := svc.Update(context.Background(), integration)
	assert.ErrorIs(t, err, domain.ErrIntegrationNotFound)
}

// --- Delete ---

func TestDeleteIntegration_Success(t *testing.T) {
	integrationRepo := &mock.IntegrationRepository{
		DeleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	agentRepo := &mock.AgentRepository{}
	svc := newTestIntegrationService(integrationRepo, agentRepo)

	err := svc.Delete(context.Background(), "integ-1")
	require.NoError(t, err)
}

func TestDeleteIntegration_NotFound(t *testing.T) {
	integrationRepo := &mock.IntegrationRepository{
		DeleteFn: func(ctx context.Context, id string) error {
			return domain.ErrIntegrationNotFound
		},
	}
	agentRepo := &mock.AgentRepository{}
	svc := newTestIntegrationService(integrationRepo, agentRepo)

	err := svc.Delete(context.Background(), "missing")
	assert.ErrorIs(t, err, domain.ErrIntegrationNotFound)
}
