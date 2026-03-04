package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/mock"
	"github.com/robwittman/pillar/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTemplateService(
	templateRepo *mock.IntegrationTemplateRepository,
	integrationRepo *mock.IntegrationRepository,
	agentRepo *mock.AgentRepository,
) service.IntegrationTemplateService {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return service.NewIntegrationTemplateService(templateRepo, integrationRepo, agentRepo, logger)
}

// --- Create ---

func TestCreateTemplate_Success(t *testing.T) {
	var captured *domain.IntegrationTemplate
	templateRepo := &mock.IntegrationTemplateRepository{
		CreateFn: func(ctx context.Context, template *domain.IntegrationTemplate) error {
			captured = template
			return nil
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	template := &domain.IntegrationTemplate{
		Type:     "vault",
		Name:     "prod-vault",
		Config:   map[string]any{"address": "https://vault:8200"},
		Selector: map[string]string{"env": "prod"},
	}
	err := svc.Create(context.Background(), template)
	require.NoError(t, err)
	assert.NotEmpty(t, captured.ID)
	assert.Equal(t, "vault", captured.Type)
	assert.Equal(t, "prod-vault", captured.Name)
	assert.NotNil(t, captured.Config)
	assert.NotNil(t, captured.Selector)
}

func TestCreateTemplate_NilConfigAndSelector(t *testing.T) {
	var captured *domain.IntegrationTemplate
	templateRepo := &mock.IntegrationTemplateRepository{
		CreateFn: func(ctx context.Context, template *domain.IntegrationTemplate) error {
			captured = template
			return nil
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	template := &domain.IntegrationTemplate{
		Type: "vault",
		Name: "prod-vault",
	}
	err := svc.Create(context.Background(), template)
	require.NoError(t, err)
	assert.NotNil(t, captured.Config)
	assert.NotNil(t, captured.Selector)
}

func TestCreateTemplate_MissingType(t *testing.T) {
	svc := newTestTemplateService(&mock.IntegrationTemplateRepository{}, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	template := &domain.IntegrationTemplate{Name: "prod-vault"}
	err := svc.Create(context.Background(), template)
	assert.ErrorIs(t, err, domain.ErrInvalidIntegrationTemplate)
}

func TestCreateTemplate_MissingName(t *testing.T) {
	svc := newTestTemplateService(&mock.IntegrationTemplateRepository{}, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	template := &domain.IntegrationTemplate{Type: "vault"}
	err := svc.Create(context.Background(), template)
	assert.ErrorIs(t, err, domain.ErrInvalidIntegrationTemplate)
}

func TestCreateTemplate_RepoError(t *testing.T) {
	templateRepo := &mock.IntegrationTemplateRepository{
		CreateFn: func(ctx context.Context, template *domain.IntegrationTemplate) error {
			return domain.ErrInvalidIntegrationTemplate
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	template := &domain.IntegrationTemplate{Type: "vault", Name: "prod-vault"}
	err := svc.Create(context.Background(), template)
	assert.Error(t, err)
}

// --- Get ---

func TestGetTemplate_Success(t *testing.T) {
	expected := &domain.IntegrationTemplate{ID: "tmpl-1", Type: "vault", Name: "prod-vault"}
	templateRepo := &mock.IntegrationTemplateRepository{
		GetFn: func(ctx context.Context, id string) (*domain.IntegrationTemplate, error) {
			return expected, nil
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	template, err := svc.Get(context.Background(), "tmpl-1")
	require.NoError(t, err)
	assert.Equal(t, expected, template)
}

func TestGetTemplate_NotFound(t *testing.T) {
	templateRepo := &mock.IntegrationTemplateRepository{
		GetFn: func(ctx context.Context, id string) (*domain.IntegrationTemplate, error) {
			return nil, domain.ErrIntegrationTemplateNotFound
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	_, err := svc.Get(context.Background(), "missing")
	assert.ErrorIs(t, err, domain.ErrIntegrationTemplateNotFound)
}

// --- List ---

func TestListTemplates_Success(t *testing.T) {
	expected := []*domain.IntegrationTemplate{
		{ID: "tmpl-1", Type: "vault", Name: "prod-vault"},
		{ID: "tmpl-2", Type: "keycloak", Name: "identity"},
	}
	templateRepo := &mock.IntegrationTemplateRepository{
		ListFn: func(ctx context.Context) ([]*domain.IntegrationTemplate, error) {
			return expected, nil
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	templates, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, templates, 2)
}

func TestListTemplates_Empty(t *testing.T) {
	templateRepo := &mock.IntegrationTemplateRepository{
		ListFn: func(ctx context.Context) ([]*domain.IntegrationTemplate, error) {
			return []*domain.IntegrationTemplate{}, nil
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	templates, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Empty(t, templates)
}

// --- Update ---

func TestUpdateTemplate_Success(t *testing.T) {
	templateRepo := &mock.IntegrationTemplateRepository{
		UpdateFn: func(ctx context.Context, template *domain.IntegrationTemplate) error {
			return nil
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	template := &domain.IntegrationTemplate{
		ID:       "tmpl-1",
		Name:     "updated-name",
		Config:   map[string]any{"address": "https://vault:8200"},
		Selector: map[string]string{"env": "staging"},
	}
	err := svc.Update(context.Background(), template)
	require.NoError(t, err)
}

func TestUpdateTemplate_MissingName(t *testing.T) {
	svc := newTestTemplateService(&mock.IntegrationTemplateRepository{}, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	template := &domain.IntegrationTemplate{ID: "tmpl-1"}
	err := svc.Update(context.Background(), template)
	assert.ErrorIs(t, err, domain.ErrInvalidIntegrationTemplate)
}

func TestUpdateTemplate_NilConfigAndSelector(t *testing.T) {
	var captured *domain.IntegrationTemplate
	templateRepo := &mock.IntegrationTemplateRepository{
		UpdateFn: func(ctx context.Context, template *domain.IntegrationTemplate) error {
			captured = template
			return nil
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	template := &domain.IntegrationTemplate{ID: "tmpl-1", Name: "updated-name"}
	err := svc.Update(context.Background(), template)
	require.NoError(t, err)
	assert.NotNil(t, captured.Config)
	assert.NotNil(t, captured.Selector)
}

func TestUpdateTemplate_NotFound(t *testing.T) {
	templateRepo := &mock.IntegrationTemplateRepository{
		UpdateFn: func(ctx context.Context, template *domain.IntegrationTemplate) error {
			return domain.ErrIntegrationTemplateNotFound
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	template := &domain.IntegrationTemplate{ID: "missing", Name: "updated"}
	err := svc.Update(context.Background(), template)
	assert.ErrorIs(t, err, domain.ErrIntegrationTemplateNotFound)
}

// --- Delete ---

func TestDeleteTemplate_Success(t *testing.T) {
	templateRepo := &mock.IntegrationTemplateRepository{
		DeleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	err := svc.Delete(context.Background(), "tmpl-1")
	require.NoError(t, err)
}

func TestDeleteTemplate_NotFound(t *testing.T) {
	templateRepo := &mock.IntegrationTemplateRepository{
		DeleteFn: func(ctx context.Context, id string) error {
			return domain.ErrIntegrationTemplateNotFound
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	err := svc.Delete(context.Background(), "missing")
	assert.ErrorIs(t, err, domain.ErrIntegrationTemplateNotFound)
}

// --- Preview ---

func TestPreviewTemplate_MatchesAll(t *testing.T) {
	templateRepo := &mock.IntegrationTemplateRepository{
		GetFn: func(ctx context.Context, id string) (*domain.IntegrationTemplate, error) {
			return &domain.IntegrationTemplate{
				ID:       id,
				Selector: map[string]string{},
			}, nil
		},
	}
	agentRepo := &mock.AgentRepository{
		ListFn: func(ctx context.Context) ([]*domain.Agent, error) {
			return []*domain.Agent{
				{ID: "a1", Labels: map[string]string{"env": "prod"}},
				{ID: "a2", Labels: map[string]string{"env": "staging"}},
			}, nil
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, agentRepo)

	agents, err := svc.Preview(context.Background(), "tmpl-1")
	require.NoError(t, err)
	assert.Len(t, agents, 2)
}

func TestPreviewTemplate_MatchesByLabel(t *testing.T) {
	templateRepo := &mock.IntegrationTemplateRepository{
		GetFn: func(ctx context.Context, id string) (*domain.IntegrationTemplate, error) {
			return &domain.IntegrationTemplate{
				ID:       id,
				Selector: map[string]string{"env": "prod"},
			}, nil
		},
	}
	agentRepo := &mock.AgentRepository{
		ListFn: func(ctx context.Context) ([]*domain.Agent, error) {
			return []*domain.Agent{
				{ID: "a1", Labels: map[string]string{"env": "prod", "tier": "backend"}},
				{ID: "a2", Labels: map[string]string{"env": "staging"}},
				{ID: "a3", Labels: map[string]string{"env": "prod"}},
			}, nil
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, agentRepo)

	agents, err := svc.Preview(context.Background(), "tmpl-1")
	require.NoError(t, err)
	assert.Len(t, agents, 2)
	assert.Equal(t, "a1", agents[0].ID)
	assert.Equal(t, "a3", agents[1].ID)
}

func TestPreviewTemplate_NoMatches(t *testing.T) {
	templateRepo := &mock.IntegrationTemplateRepository{
		GetFn: func(ctx context.Context, id string) (*domain.IntegrationTemplate, error) {
			return &domain.IntegrationTemplate{
				ID:       id,
				Selector: map[string]string{"env": "prod"},
			}, nil
		},
	}
	agentRepo := &mock.AgentRepository{
		ListFn: func(ctx context.Context) ([]*domain.Agent, error) {
			return []*domain.Agent{
				{ID: "a1", Labels: map[string]string{"env": "staging"}},
			}, nil
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, agentRepo)

	agents, err := svc.Preview(context.Background(), "tmpl-1")
	require.NoError(t, err)
	assert.Empty(t, agents)
}

func TestPreviewTemplate_TemplateNotFound(t *testing.T) {
	templateRepo := &mock.IntegrationTemplateRepository{
		GetFn: func(ctx context.Context, id string) (*domain.IntegrationTemplate, error) {
			return nil, domain.ErrIntegrationTemplateNotFound
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	_, err := svc.Preview(context.Background(), "missing")
	assert.ErrorIs(t, err, domain.ErrIntegrationTemplateNotFound)
}

// --- ProvisionForAgent ---

func TestProvisionForAgent_Success(t *testing.T) {
	var createdIntegrations []*domain.Integration
	templateRepo := &mock.IntegrationTemplateRepository{
		FindMatchingTemplatesFn: func(ctx context.Context, labels map[string]string) ([]*domain.IntegrationTemplate, error) {
			return []*domain.IntegrationTemplate{
				{ID: "tmpl-1", Type: "vault", Name: "prod-vault", Config: map[string]any{"address": "https://vault:8200"}},
				{ID: "tmpl-2", Type: "keycloak", Name: "identity", Config: map[string]any{"url": "https://keycloak:8080"}},
			}, nil
		},
	}
	integrationRepo := &mock.IntegrationRepository{
		CreateFn: func(ctx context.Context, integration *domain.Integration) error {
			createdIntegrations = append(createdIntegrations, integration)
			return nil
		},
	}
	svc := newTestTemplateService(templateRepo, integrationRepo, &mock.AgentRepository{})

	err := svc.ProvisionForAgent(context.Background(), "agent-1", map[string]string{"env": "prod"})
	require.NoError(t, err)
	assert.Len(t, createdIntegrations, 2)
	assert.Equal(t, "agent-1", createdIntegrations[0].AgentID)
	assert.Equal(t, "vault", createdIntegrations[0].Type)
	assert.Equal(t, "tmpl-1", createdIntegrations[0].TemplateID)
	assert.Equal(t, "keycloak", createdIntegrations[1].Type)
	assert.Equal(t, "tmpl-2", createdIntegrations[1].TemplateID)
}

func TestProvisionForAgent_NoMatchingTemplates(t *testing.T) {
	templateRepo := &mock.IntegrationTemplateRepository{
		FindMatchingTemplatesFn: func(ctx context.Context, labels map[string]string) ([]*domain.IntegrationTemplate, error) {
			return nil, nil
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	err := svc.ProvisionForAgent(context.Background(), "agent-1", map[string]string{"env": "prod"})
	require.NoError(t, err)
}

func TestProvisionForAgent_SkipsDuplicates(t *testing.T) {
	templateRepo := &mock.IntegrationTemplateRepository{
		FindMatchingTemplatesFn: func(ctx context.Context, labels map[string]string) ([]*domain.IntegrationTemplate, error) {
			return []*domain.IntegrationTemplate{
				{ID: "tmpl-1", Type: "vault", Name: "prod-vault", Config: map[string]any{}},
			}, nil
		},
	}
	integrationRepo := &mock.IntegrationRepository{
		CreateFn: func(ctx context.Context, integration *domain.Integration) error {
			return domain.ErrInvalidIntegration
		},
	}
	svc := newTestTemplateService(templateRepo, integrationRepo, &mock.AgentRepository{})

	err := svc.ProvisionForAgent(context.Background(), "agent-1", map[string]string{"env": "prod"})
	require.NoError(t, err)
}

func TestProvisionForAgent_ContinuesOnError(t *testing.T) {
	var createdCount int
	templateRepo := &mock.IntegrationTemplateRepository{
		FindMatchingTemplatesFn: func(ctx context.Context, labels map[string]string) ([]*domain.IntegrationTemplate, error) {
			return []*domain.IntegrationTemplate{
				{ID: "tmpl-1", Type: "vault", Name: "prod-vault", Config: map[string]any{}},
				{ID: "tmpl-2", Type: "keycloak", Name: "identity", Config: map[string]any{}},
			}, nil
		},
	}
	integrationRepo := &mock.IntegrationRepository{
		CreateFn: func(ctx context.Context, integration *domain.Integration) error {
			if integration.Type == "vault" {
				return errors.New("db error")
			}
			createdCount++
			return nil
		},
	}
	svc := newTestTemplateService(templateRepo, integrationRepo, &mock.AgentRepository{})

	err := svc.ProvisionForAgent(context.Background(), "agent-1", map[string]string{"env": "prod"})
	require.NoError(t, err)
	assert.Equal(t, 1, createdCount)
}

func TestProvisionForAgent_FindTemplatesError(t *testing.T) {
	templateRepo := &mock.IntegrationTemplateRepository{
		FindMatchingTemplatesFn: func(ctx context.Context, labels map[string]string) ([]*domain.IntegrationTemplate, error) {
			return nil, errors.New("db error")
		},
	}
	svc := newTestTemplateService(templateRepo, &mock.IntegrationRepository{}, &mock.AgentRepository{})

	err := svc.ProvisionForAgent(context.Background(), "agent-1", map[string]string{"env": "prod"})
	assert.Error(t, err)
}
