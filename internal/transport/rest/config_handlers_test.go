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

func setupConfigRouter(agentSvc service.AgentService, configSvc service.ConfigService) *chi.Mux {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := chi.NewRouter()
	h := NewHandlers(agentSvc, logger)
	ch := NewConfigHandlers(configSvc, logger)
	wh := NewWebhookHandlers(noopWebhookService(), logger)
	ah := NewAttributeHandlers(noopAttributeService(), logger)
	RegisterRoutes(r, h, ch, wh, ah, nil, nil, nil, nil, nil, nil, false)
	return r
}

func noopAgentService() *mock.AgentService {
	return &mock.AgentService{
		CreateFn: func(ctx context.Context, name string, metadata, labels map[string]string) (*domain.Agent, error) {
			return nil, nil
		},
		GetFn:  func(ctx context.Context, id string) (*domain.Agent, error) { return nil, nil },
		ListFn: func(ctx context.Context) ([]*domain.Agent, error) { return nil, nil },
		UpdateFn: func(ctx context.Context, id, name string, metadata, labels map[string]string) (*domain.Agent, error) {
			return nil, nil
		},
		DeleteFn:    func(ctx context.Context, id string) error { return nil },
		StartFn:     func(ctx context.Context, id string) error { return nil },
		StopFn:      func(ctx context.Context, id string) error { return nil },
		StatusFn:    func(ctx context.Context, id string) (*service.AgentStatusInfo, error) { return nil, nil },
		HeartbeatFn: func(ctx context.Context, agentID string) error { return nil },
	}
}

// --- CreateConfig ---

func TestCreateConfig_Success(t *testing.T) {
	configSvc := &mock.ConfigService{
		CreateConfigFn: func(ctx context.Context, config *domain.AgentConfig, apiCredential string) error {
			assert.Equal(t, "agent-1", config.AgentID)
			assert.Equal(t, domain.ModelProviderClaude, config.ModelProvider)
			assert.Equal(t, "sk-test", apiCredential)
			return nil
		},
	}
	r := setupConfigRouter(noopAgentService(), configSvc)

	body := `{"model_provider":"claude","model_id":"claude-sonnet-4-20250514","api_credential":"sk-test","max_iterations":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/agent-1/config", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var config domain.AgentConfig
	require.NoError(t, json.NewDecoder(w.Body).Decode(&config))
	assert.Equal(t, "agent-1", config.AgentID)
}

func TestCreateConfig_BadJSON(t *testing.T) {
	configSvc := &mock.ConfigService{}
	r := setupConfigRouter(noopAgentService(), configSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/agent-1/config", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateConfig_AgentNotFound(t *testing.T) {
	configSvc := &mock.ConfigService{
		CreateConfigFn: func(ctx context.Context, config *domain.AgentConfig, apiCredential string) error {
			return domain.ErrAgentNotFound
		},
	}
	r := setupConfigRouter(noopAgentService(), configSvc)

	body := `{"model_provider":"claude","model_id":"claude-sonnet-4-20250514"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/missing/config", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateConfig_AlreadyExists(t *testing.T) {
	configSvc := &mock.ConfigService{
		CreateConfigFn: func(ctx context.Context, config *domain.AgentConfig, apiCredential string) error {
			return domain.ErrConfigAlreadyExists
		},
	}
	r := setupConfigRouter(noopAgentService(), configSvc)

	body := `{"model_provider":"claude","model_id":"claude-sonnet-4-20250514"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/agent-1/config", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateConfig_InvalidConfig(t *testing.T) {
	configSvc := &mock.ConfigService{
		CreateConfigFn: func(ctx context.Context, config *domain.AgentConfig, apiCredential string) error {
			return domain.ErrInvalidConfig
		},
	}
	r := setupConfigRouter(noopAgentService(), configSvc)

	body := `{"model_provider":"","model_id":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/agent-1/config", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- GetConfig ---

func TestGetConfig_Success(t *testing.T) {
	configSvc := &mock.ConfigService{
		GetConfigFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, error) {
			return &domain.AgentConfig{
				AgentID:       agentID,
				ModelProvider: domain.ModelProviderClaude,
				ModelID:       "claude-sonnet-4-20250514",
			}, nil
		},
	}
	r := setupConfigRouter(noopAgentService(), configSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/agent-1/config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var config domain.AgentConfig
	require.NoError(t, json.NewDecoder(w.Body).Decode(&config))
	assert.Equal(t, "agent-1", config.AgentID)
	assert.Equal(t, domain.ModelProviderClaude, config.ModelProvider)
}

func TestGetConfig_NotFound(t *testing.T) {
	configSvc := &mock.ConfigService{
		GetConfigFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, error) {
			return nil, domain.ErrConfigNotFound
		},
	}
	r := setupConfigRouter(noopAgentService(), configSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/missing/config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- UpdateConfig ---

func TestUpdateConfig_Success(t *testing.T) {
	configSvc := &mock.ConfigService{
		UpdateConfigFn: func(ctx context.Context, config *domain.AgentConfig, apiCredential string) error {
			assert.Equal(t, "agent-1", config.AgentID)
			assert.Equal(t, "new-key", apiCredential)
			return nil
		},
	}
	r := setupConfigRouter(noopAgentService(), configSvc)

	body := `{"model_provider":"claude","model_id":"claude-sonnet-4-20250514","api_credential":"new-key","max_iterations":200}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/agents/agent-1/config", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateConfig_NotFound(t *testing.T) {
	configSvc := &mock.ConfigService{
		UpdateConfigFn: func(ctx context.Context, config *domain.AgentConfig, apiCredential string) error {
			return domain.ErrConfigNotFound
		},
	}
	r := setupConfigRouter(noopAgentService(), configSvc)

	body := `{"model_provider":"claude","model_id":"claude-sonnet-4-20250514"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/agents/missing/config", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateConfig_BadJSON(t *testing.T) {
	configSvc := &mock.ConfigService{}
	r := setupConfigRouter(noopAgentService(), configSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/agents/agent-1/config", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- DeleteConfig ---

func TestDeleteConfig_Success(t *testing.T) {
	configSvc := &mock.ConfigService{
		DeleteConfigFn: func(ctx context.Context, agentID string) error {
			assert.Equal(t, "agent-1", agentID)
			return nil
		},
	}
	r := setupConfigRouter(noopAgentService(), configSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/agents/agent-1/config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteConfig_NotFound(t *testing.T) {
	configSvc := &mock.ConfigService{
		DeleteConfigFn: func(ctx context.Context, agentID string) error {
			return domain.ErrConfigNotFound
		},
	}
	r := setupConfigRouter(noopAgentService(), configSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/agents/missing/config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
