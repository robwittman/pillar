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

func setupRouter(svc service.AgentService) *chi.Mux {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := chi.NewRouter()
	h := NewHandlers(svc, logger)
	// Use a no-op ConfigService for agent-only tests
	configSvc := &mock.ConfigService{
		CreateConfigFn:         func(ctx context.Context, config *domain.AgentConfig, apiCredential string) error { return nil },
		GetConfigFn:            func(ctx context.Context, agentID string) (*domain.AgentConfig, error) { return nil, nil },
		GetConfigWithSecretsFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, string, error) { return nil, "", nil },
		UpdateConfigFn:         func(ctx context.Context, config *domain.AgentConfig, apiCredential string) error { return nil },
		DeleteConfigFn:         func(ctx context.Context, agentID string) error { return nil },
	}
	ch := NewConfigHandlers(configSvc, logger)
	integSvc := &mock.IntegrationService{
		CreateFn: func(ctx context.Context, integration *domain.Integration) error { return nil },
		GetFn:    func(ctx context.Context, id string) (*domain.Integration, error) { return nil, nil },
		ListFn:   func(ctx context.Context, agentID string) ([]*domain.Integration, error) { return nil, nil },
		UpdateFn: func(ctx context.Context, integration *domain.Integration) error { return nil },
		DeleteFn: func(ctx context.Context, id string) error { return nil },
	}
	ih := NewIntegrationHandlers(integSvc, logger)
	ith := NewIntegrationTemplateHandlers(noopIntegrationTemplateService(), logger)
	RegisterRoutes(r, h, ch, ih, ith)
	return r
}

func noopIntegrationTemplateService() *mock.IntegrationTemplateService {
	return &mock.IntegrationTemplateService{
		CreateFn:            func(ctx context.Context, template *domain.IntegrationTemplate) error { return nil },
		GetFn:               func(ctx context.Context, id string) (*domain.IntegrationTemplate, error) { return nil, nil },
		ListFn:              func(ctx context.Context) ([]*domain.IntegrationTemplate, error) { return nil, nil },
		UpdateFn:            func(ctx context.Context, template *domain.IntegrationTemplate) error { return nil },
		DeleteFn:            func(ctx context.Context, id string) error { return nil },
		PreviewFn:           func(ctx context.Context, id string) ([]*domain.Agent, error) { return nil, nil },
		ProvisionForAgentFn: func(ctx context.Context, agentID string, labels map[string]string) error { return nil },
	}
}

func TestHealth(t *testing.T) {
	r := setupRouter(&mock.AgentService{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	assert.Equal(t, "ok", body["status"])
}

// --- CreateAgent ---

func TestCreateAgent_Success(t *testing.T) {
	svc := &mock.AgentService{
		CreateFn: func(ctx context.Context, name string, metadata, labels map[string]string) (*domain.Agent, error) {
			return &domain.Agent{ID: "new-id", Name: name, Status: domain.AgentStatusPending}, nil
		},
	}
	r := setupRouter(svc)

	body := `{"name":"test-agent","metadata":{"k":"v"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var agent domain.Agent
	json.NewDecoder(w.Body).Decode(&agent)
	assert.Equal(t, "new-id", agent.ID)
	assert.Equal(t, "test-agent", agent.Name)
}

func TestCreateAgent_MissingName(t *testing.T) {
	svc := &mock.AgentService{}
	r := setupRouter(svc)

	body := `{"metadata":{"k":"v"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateAgent_BadJSON(t *testing.T) {
	svc := &mock.AgentService{}
	r := setupRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewBufferString("{invalid"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- GetAgent ---

func TestGetAgent_Success(t *testing.T) {
	svc := &mock.AgentService{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Name: "a"}, nil
		},
	}
	r := setupRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var agent domain.Agent
	json.NewDecoder(w.Body).Decode(&agent)
	assert.Equal(t, "abc", agent.ID)
}

func TestGetAgent_NotFound(t *testing.T) {
	svc := &mock.AgentService{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return nil, domain.ErrAgentNotFound
		},
	}
	r := setupRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/missing", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- ListAgents ---

func TestListAgents_Success(t *testing.T) {
	svc := &mock.AgentService{
		ListFn: func(ctx context.Context) ([]*domain.Agent, error) {
			return []*domain.Agent{{ID: "1"}, {ID: "2"}}, nil
		},
	}
	r := setupRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var agents []domain.Agent
	json.NewDecoder(w.Body).Decode(&agents)
	assert.Len(t, agents, 2)
}

// --- UpdateAgent ---

func TestUpdateAgent_Success(t *testing.T) {
	svc := &mock.AgentService{
		UpdateFn: func(ctx context.Context, id, name string, metadata, labels map[string]string) (*domain.Agent, error) {
			return &domain.Agent{ID: id, Name: name}, nil
		},
	}
	r := setupRouter(svc)

	body := `{"name":"updated"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/agents/abc", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var agent domain.Agent
	json.NewDecoder(w.Body).Decode(&agent)
	assert.Equal(t, "updated", agent.Name)
}

func TestUpdateAgent_NotFound(t *testing.T) {
	svc := &mock.AgentService{
		UpdateFn: func(ctx context.Context, id, name string, metadata, labels map[string]string) (*domain.Agent, error) {
			return nil, domain.ErrAgentNotFound
		},
	}
	r := setupRouter(svc)

	body := `{"name":"x"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/agents/missing", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateAgent_BadJSON(t *testing.T) {
	svc := &mock.AgentService{}
	r := setupRouter(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/agents/abc", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- DeleteAgent ---

func TestDeleteAgent_Success(t *testing.T) {
	svc := &mock.AgentService{
		DeleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	r := setupRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/agents/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteAgent_NotFound(t *testing.T) {
	svc := &mock.AgentService{
		DeleteFn: func(ctx context.Context, id string) error {
			return domain.ErrAgentNotFound
		},
	}
	r := setupRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/agents/missing", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- StartAgent ---

func TestStartAgent_Success(t *testing.T) {
	svc := &mock.AgentService{
		StartFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	r := setupRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/abc/start", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	assert.Equal(t, "started", body["status"])
}

func TestStartAgent_NotFound(t *testing.T) {
	svc := &mock.AgentService{
		StartFn: func(ctx context.Context, id string) error {
			return domain.ErrAgentNotFound
		},
	}
	r := setupRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/missing/start", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestStartAgent_InvalidTransition(t *testing.T) {
	svc := &mock.AgentService{
		StartFn: func(ctx context.Context, id string) error {
			return domain.ErrInvalidTransition
		},
	}
	r := setupRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/abc/start", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

// --- StopAgent ---

func TestStopAgent_Success(t *testing.T) {
	svc := &mock.AgentService{
		StopFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	r := setupRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/abc/stop", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStopAgent_NotFound(t *testing.T) {
	svc := &mock.AgentService{
		StopFn: func(ctx context.Context, id string) error {
			return domain.ErrAgentNotFound
		},
	}
	r := setupRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/missing/stop", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestStopAgent_InvalidTransition(t *testing.T) {
	svc := &mock.AgentService{
		StopFn: func(ctx context.Context, id string) error {
			return domain.ErrInvalidTransition
		},
	}
	r := setupRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/abc/stop", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

// --- AgentStatus ---

func TestAgentStatus_Success(t *testing.T) {
	svc := &mock.AgentService{
		StatusFn: func(ctx context.Context, id string) (*service.AgentStatusInfo, error) {
			return &service.AgentStatusInfo{AgentID: id, Status: "running", Online: true}, nil
		},
	}
	r := setupRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/abc/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var info service.AgentStatusInfo
	require.NoError(t, json.NewDecoder(w.Body).Decode(&info))
	assert.Equal(t, "abc", info.AgentID)
	assert.True(t, info.Online)
}

func TestAgentStatus_NotFound(t *testing.T) {
	svc := &mock.AgentService{
		StatusFn: func(ctx context.Context, id string) (*service.AgentStatusInfo, error) {
			return nil, domain.ErrAgentNotFound
		},
	}
	r := setupRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/missing/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
