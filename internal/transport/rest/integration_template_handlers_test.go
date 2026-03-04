package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTemplateRouter(templateSvc *mock.IntegrationTemplateService) *chi.Mux {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := chi.NewRouter()
	h := NewHandlers(noopAgentService(), logger)
	ch := NewConfigHandlers(noopConfigService(), logger)
	integSvc := &mock.IntegrationService{
		CreateFn: func(ctx context.Context, integration *domain.Integration) error { return nil },
		GetFn:    func(ctx context.Context, id string) (*domain.Integration, error) { return nil, nil },
		ListFn:   func(ctx context.Context, agentID string) ([]*domain.Integration, error) { return nil, nil },
		UpdateFn: func(ctx context.Context, integration *domain.Integration) error { return nil },
		DeleteFn: func(ctx context.Context, id string) error { return nil },
	}
	ih := NewIntegrationHandlers(integSvc, logger)
	ith := NewIntegrationTemplateHandlers(templateSvc, logger)
	RegisterRoutes(r, h, ch, ih, ith)
	return r
}

// --- CreateTemplate ---

func TestCreateTemplate_HandlerSuccess(t *testing.T) {
	svc := &mock.IntegrationTemplateService{
		CreateFn: func(ctx context.Context, template *domain.IntegrationTemplate) error {
			assert.Equal(t, "vault", template.Type)
			assert.Equal(t, "prod-vault", template.Name)
			assert.Equal(t, map[string]string{"env": "prod"}, template.Selector)
			template.ID = "tmpl-1"
			return nil
		},
	}
	r := setupTemplateRouter(svc)

	body := `{"type":"vault","name":"prod-vault","config":{"address":"https://vault:8200"},"selector":{"env":"prod"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integration-templates", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var template domain.IntegrationTemplate
	require.NoError(t, json.NewDecoder(w.Body).Decode(&template))
	assert.Equal(t, "tmpl-1", template.ID)
}

func TestCreateTemplate_HandlerBadJSON(t *testing.T) {
	svc := noopIntegrationTemplateService()
	r := setupTemplateRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/integration-templates", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateTemplate_HandlerInvalid(t *testing.T) {
	svc := &mock.IntegrationTemplateService{
		CreateFn: func(ctx context.Context, template *domain.IntegrationTemplate) error {
			return domain.ErrInvalidIntegrationTemplate
		},
	}
	r := setupTemplateRouter(svc)

	body := `{"type":"","name":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integration-templates", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- ListTemplates ---

func TestListTemplates_HandlerSuccess(t *testing.T) {
	svc := &mock.IntegrationTemplateService{
		ListFn: func(ctx context.Context) ([]*domain.IntegrationTemplate, error) {
			return []*domain.IntegrationTemplate{
				{ID: "tmpl-1", Type: "vault", Name: "prod-vault"},
				{ID: "tmpl-2", Type: "keycloak", Name: "identity"},
			}, nil
		},
	}
	r := setupTemplateRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/integration-templates", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var templates []domain.IntegrationTemplate
	require.NoError(t, json.NewDecoder(w.Body).Decode(&templates))
	assert.Len(t, templates, 2)
}

func TestListTemplates_HandlerEmpty(t *testing.T) {
	svc := &mock.IntegrationTemplateService{
		ListFn: func(ctx context.Context) ([]*domain.IntegrationTemplate, error) {
			return nil, nil
		},
	}
	r := setupTemplateRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/integration-templates", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var templates []domain.IntegrationTemplate
	require.NoError(t, json.NewDecoder(w.Body).Decode(&templates))
	assert.Empty(t, templates)
}

// --- GetTemplate ---

func TestGetTemplate_HandlerSuccess(t *testing.T) {
	svc := &mock.IntegrationTemplateService{
		GetFn: func(ctx context.Context, id string) (*domain.IntegrationTemplate, error) {
			return &domain.IntegrationTemplate{
				ID:   id,
				Type: "vault",
				Name: "prod-vault",
			}, nil
		},
	}
	r := setupTemplateRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/integration-templates/tmpl-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var template domain.IntegrationTemplate
	require.NoError(t, json.NewDecoder(w.Body).Decode(&template))
	assert.Equal(t, "tmpl-1", template.ID)
}

func TestGetTemplate_HandlerNotFound(t *testing.T) {
	svc := &mock.IntegrationTemplateService{
		GetFn: func(ctx context.Context, id string) (*domain.IntegrationTemplate, error) {
			return nil, domain.ErrIntegrationTemplateNotFound
		},
	}
	r := setupTemplateRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/integration-templates/missing", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- UpdateTemplate ---

func TestUpdateTemplate_HandlerSuccess(t *testing.T) {
	svc := &mock.IntegrationTemplateService{
		UpdateFn: func(ctx context.Context, template *domain.IntegrationTemplate) error {
			assert.Equal(t, "tmpl-1", template.ID)
			assert.Equal(t, "new-name", template.Name)
			return nil
		},
	}
	r := setupTemplateRouter(svc)

	body := `{"name":"new-name","config":{"address":"https://vault:8200"},"selector":{"env":"staging"}}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/integration-templates/tmpl-1", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateTemplate_HandlerBadJSON(t *testing.T) {
	svc := noopIntegrationTemplateService()
	r := setupTemplateRouter(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/integration-templates/tmpl-1", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateTemplate_HandlerNotFound(t *testing.T) {
	svc := &mock.IntegrationTemplateService{
		UpdateFn: func(ctx context.Context, template *domain.IntegrationTemplate) error {
			return domain.ErrIntegrationTemplateNotFound
		},
	}
	r := setupTemplateRouter(svc)

	body := `{"name":"new-name"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/integration-templates/missing", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- DeleteTemplate ---

func TestDeleteTemplate_HandlerSuccess(t *testing.T) {
	svc := &mock.IntegrationTemplateService{
		DeleteFn: func(ctx context.Context, id string) error {
			assert.Equal(t, "tmpl-1", id)
			return nil
		},
	}
	r := setupTemplateRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/integration-templates/tmpl-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteTemplate_HandlerNotFound(t *testing.T) {
	svc := &mock.IntegrationTemplateService{
		DeleteFn: func(ctx context.Context, id string) error {
			return domain.ErrIntegrationTemplateNotFound
		},
	}
	r := setupTemplateRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/integration-templates/missing", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- PreviewTemplate ---

func TestPreviewTemplate_HandlerSuccess(t *testing.T) {
	svc := &mock.IntegrationTemplateService{
		PreviewFn: func(ctx context.Context, id string) ([]*domain.Agent, error) {
			return []*domain.Agent{
				{ID: "agent-1", Name: "prod-agent", Labels: map[string]string{"env": "prod"}},
			}, nil
		},
	}
	r := setupTemplateRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/integration-templates/tmpl-1/preview", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var agents []domain.Agent
	require.NoError(t, json.NewDecoder(w.Body).Decode(&agents))
	assert.Len(t, agents, 1)
	assert.Equal(t, "agent-1", agents[0].ID)
}

func TestPreviewTemplate_HandlerEmpty(t *testing.T) {
	svc := &mock.IntegrationTemplateService{
		PreviewFn: func(ctx context.Context, id string) ([]*domain.Agent, error) {
			return nil, nil
		},
	}
	r := setupTemplateRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/integration-templates/tmpl-1/preview", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var agents []domain.Agent
	require.NoError(t, json.NewDecoder(w.Body).Decode(&agents))
	assert.Empty(t, agents)
}

func TestPreviewTemplate_HandlerNotFound(t *testing.T) {
	svc := &mock.IntegrationTemplateService{
		PreviewFn: func(ctx context.Context, id string) ([]*domain.Agent, error) {
			return nil, domain.ErrIntegrationTemplateNotFound
		},
	}
	r := setupTemplateRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/integration-templates/missing/preview", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
