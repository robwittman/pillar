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
	"github.com/robwittman/pillar/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIntegrationRouter(agentSvc service.AgentService, configSvc service.ConfigService, integSvc service.IntegrationService) *chi.Mux {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := chi.NewRouter()
	h := NewHandlers(agentSvc, logger)
	ch := NewConfigHandlers(configSvc, logger)
	ih := NewIntegrationHandlers(integSvc, logger)
	ith := NewIntegrationTemplateHandlers(noopIntegrationTemplateService(), logger)
	RegisterRoutes(r, h, ch, ih, ith)
	return r
}

func noopConfigService() *mock.ConfigService {
	return &mock.ConfigService{
		CreateConfigFn: func(ctx context.Context, config *domain.AgentConfig, apiCredential string) error {
			return nil
		},
		GetConfigFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, error) {
			return nil, nil
		},
		GetConfigWithSecretsFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, string, error) {
			return nil, "", nil
		},
		UpdateConfigFn: func(ctx context.Context, config *domain.AgentConfig, apiCredential string) error {
			return nil
		},
		DeleteConfigFn: func(ctx context.Context, agentID string) error { return nil },
	}
}

// --- CreateIntegration ---

func TestCreateIntegration_HandlerSuccess(t *testing.T) {
	integSvc := &mock.IntegrationService{
		CreateFn: func(ctx context.Context, integration *domain.Integration) error {
			assert.Equal(t, "agent-1", integration.AgentID)
			assert.Equal(t, "vault", integration.Type)
			assert.Equal(t, "prod-vault", integration.Name)
			integration.ID = "integ-1"
			return nil
		},
	}
	r := setupIntegrationRouter(noopAgentService(), noopConfigService(), integSvc)

	body := `{"type":"vault","name":"prod-vault","config":{"address":"https://vault:8200"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/agent-1/integrations", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var integration domain.Integration
	require.NoError(t, json.NewDecoder(w.Body).Decode(&integration))
	assert.Equal(t, "integ-1", integration.ID)
	assert.Equal(t, "agent-1", integration.AgentID)
}

func TestCreateIntegration_HandlerBadJSON(t *testing.T) {
	integSvc := &mock.IntegrationService{}
	r := setupIntegrationRouter(noopAgentService(), noopConfigService(), integSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/agent-1/integrations", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateIntegration_HandlerAgentNotFound(t *testing.T) {
	integSvc := &mock.IntegrationService{
		CreateFn: func(ctx context.Context, integration *domain.Integration) error {
			return domain.ErrAgentNotFound
		},
	}
	r := setupIntegrationRouter(noopAgentService(), noopConfigService(), integSvc)

	body := `{"type":"vault","name":"prod-vault"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/missing/integrations", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateIntegration_HandlerInvalid(t *testing.T) {
	integSvc := &mock.IntegrationService{
		CreateFn: func(ctx context.Context, integration *domain.Integration) error {
			return domain.ErrInvalidIntegration
		},
	}
	r := setupIntegrationRouter(noopAgentService(), noopConfigService(), integSvc)

	body := `{"type":"","name":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/agent-1/integrations", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- ListIntegrations ---

func TestListIntegrations_HandlerSuccess(t *testing.T) {
	integSvc := &mock.IntegrationService{
		ListFn: func(ctx context.Context, agentID string) ([]*domain.Integration, error) {
			return []*domain.Integration{
				{ID: "integ-1", AgentID: agentID, Type: "vault", Name: "prod-vault"},
				{ID: "integ-2", AgentID: agentID, Type: "keystone", Name: "identity"},
			}, nil
		},
	}
	r := setupIntegrationRouter(noopAgentService(), noopConfigService(), integSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/agent-1/integrations", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var integrations []domain.Integration
	require.NoError(t, json.NewDecoder(w.Body).Decode(&integrations))
	assert.Len(t, integrations, 2)
}

func TestListIntegrations_HandlerEmpty(t *testing.T) {
	integSvc := &mock.IntegrationService{
		ListFn: func(ctx context.Context, agentID string) ([]*domain.Integration, error) {
			return nil, nil
		},
	}
	r := setupIntegrationRouter(noopAgentService(), noopConfigService(), integSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/agent-1/integrations", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var integrations []domain.Integration
	require.NoError(t, json.NewDecoder(w.Body).Decode(&integrations))
	assert.Empty(t, integrations)
}

// --- GetIntegration ---

func TestGetIntegration_HandlerSuccess(t *testing.T) {
	integSvc := &mock.IntegrationService{
		GetFn: func(ctx context.Context, id string) (*domain.Integration, error) {
			return &domain.Integration{
				ID:      id,
				AgentID: "agent-1",
				Type:    "vault",
				Name:    "prod-vault",
			}, nil
		},
	}
	r := setupIntegrationRouter(noopAgentService(), noopConfigService(), integSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/agent-1/integrations/integ-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var integration domain.Integration
	require.NoError(t, json.NewDecoder(w.Body).Decode(&integration))
	assert.Equal(t, "integ-1", integration.ID)
}

func TestGetIntegration_HandlerNotFound(t *testing.T) {
	integSvc := &mock.IntegrationService{
		GetFn: func(ctx context.Context, id string) (*domain.Integration, error) {
			return nil, domain.ErrIntegrationNotFound
		},
	}
	r := setupIntegrationRouter(noopAgentService(), noopConfigService(), integSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/agent-1/integrations/missing", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- UpdateIntegration ---

func TestUpdateIntegration_HandlerSuccess(t *testing.T) {
	integSvc := &mock.IntegrationService{
		UpdateFn: func(ctx context.Context, integration *domain.Integration) error {
			assert.Equal(t, "integ-1", integration.ID)
			assert.Equal(t, "new-name", integration.Name)
			return nil
		},
	}
	r := setupIntegrationRouter(noopAgentService(), noopConfigService(), integSvc)

	body := `{"name":"new-name","config":{"address":"https://vault:8200"}}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/agents/agent-1/integrations/integ-1", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateIntegration_HandlerBadJSON(t *testing.T) {
	integSvc := &mock.IntegrationService{}
	r := setupIntegrationRouter(noopAgentService(), noopConfigService(), integSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/agents/agent-1/integrations/integ-1", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateIntegration_HandlerNotFound(t *testing.T) {
	integSvc := &mock.IntegrationService{
		UpdateFn: func(ctx context.Context, integration *domain.Integration) error {
			return domain.ErrIntegrationNotFound
		},
	}
	r := setupIntegrationRouter(noopAgentService(), noopConfigService(), integSvc)

	body := `{"name":"new-name"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/agents/agent-1/integrations/missing", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- DeleteIntegration ---

func TestDeleteIntegration_HandlerSuccess(t *testing.T) {
	integSvc := &mock.IntegrationService{
		DeleteFn: func(ctx context.Context, id string) error {
			assert.Equal(t, "integ-1", id)
			return nil
		},
	}
	r := setupIntegrationRouter(noopAgentService(), noopConfigService(), integSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/agents/agent-1/integrations/integ-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteIntegration_HandlerNotFound(t *testing.T) {
	integSvc := &mock.IntegrationService{
		DeleteFn: func(ctx context.Context, id string) error {
			return domain.ErrIntegrationNotFound
		},
	}
	r := setupIntegrationRouter(noopAgentService(), noopConfigService(), integSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/agents/agent-1/integrations/missing", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
