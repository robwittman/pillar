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

func setupAttributeRouter(attrSvc service.AttributeService) *chi.Mux {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := chi.NewRouter()
	h := NewHandlers(noopAgentService(), logger)
	ch := NewConfigHandlers(&mock.ConfigService{
		CreateConfigFn:         func(ctx context.Context, config *domain.AgentConfig, apiCredential string) error { return nil },
		GetConfigFn:            func(ctx context.Context, agentID string) (*domain.AgentConfig, error) { return nil, nil },
		GetConfigWithSecretsFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, string, error) { return nil, "", nil },
		UpdateConfigFn:         func(ctx context.Context, config *domain.AgentConfig, apiCredential string) error { return nil },
		DeleteConfigFn:         func(ctx context.Context, agentID string) error { return nil },
	}, logger)
	wh := NewWebhookHandlers(noopWebhookService(), logger)
	ah := NewAttributeHandlers(attrSvc, logger)
	RegisterRoutes(r, h, ch, wh, ah, nil)
	return r
}

func TestSetAttribute_Success(t *testing.T) {
	svc := &mock.AttributeService{
		SetFn: func(ctx context.Context, agentID, namespace string, value json.RawMessage) (*domain.AgentAttribute, error) {
			return &domain.AgentAttribute{
				AgentID:   agentID,
				Namespace: namespace,
				Value:     value,
			}, nil
		},
	}
	r := setupAttributeRouter(svc)

	body := `{"key":"value"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/agents/agent-1/attributes/keycloak", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var attr domain.AgentAttribute
	require.NoError(t, json.NewDecoder(w.Body).Decode(&attr))
	assert.Equal(t, "agent-1", attr.AgentID)
	assert.Equal(t, "keycloak", attr.Namespace)
}

func TestSetAttribute_InvalidJSON(t *testing.T) {
	svc := &mock.AttributeService{}
	r := setupAttributeRouter(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/agents/agent-1/attributes/ns", bytes.NewBufferString("{invalid"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSetAttribute_AgentNotFound(t *testing.T) {
	svc := &mock.AttributeService{
		SetFn: func(ctx context.Context, agentID, namespace string, value json.RawMessage) (*domain.AgentAttribute, error) {
			return nil, domain.ErrAgentNotFound
		},
	}
	r := setupAttributeRouter(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/agents/missing/attributes/ns", bytes.NewBufferString(`{}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListAttributes_Success(t *testing.T) {
	svc := &mock.AttributeService{
		ListFn: func(ctx context.Context, agentID string) ([]*domain.AgentAttribute, error) {
			return []*domain.AgentAttribute{
				{AgentID: agentID, Namespace: "vault"},
				{AgentID: agentID, Namespace: "keycloak"},
			}, nil
		},
	}
	r := setupAttributeRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/agent-1/attributes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetAttribute_Success(t *testing.T) {
	svc := &mock.AttributeService{
		GetFn: func(ctx context.Context, agentID, namespace string) (*domain.AgentAttribute, error) {
			return &domain.AgentAttribute{AgentID: agentID, Namespace: namespace}, nil
		},
	}
	r := setupAttributeRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/agent-1/attributes/vault", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetAttribute_NotFound(t *testing.T) {
	svc := &mock.AttributeService{
		GetFn: func(ctx context.Context, agentID, namespace string) (*domain.AgentAttribute, error) {
			return nil, domain.ErrAttributeNotFound
		},
	}
	r := setupAttributeRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/agent-1/attributes/missing", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteAttribute_Success(t *testing.T) {
	svc := &mock.AttributeService{
		DeleteFn: func(ctx context.Context, agentID, namespace string) error {
			return nil
		},
	}
	r := setupAttributeRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/agents/agent-1/attributes/vault", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteAttribute_NotFound(t *testing.T) {
	svc := &mock.AttributeService{
		DeleteFn: func(ctx context.Context, agentID, namespace string) error {
			return domain.ErrAttributeNotFound
		},
	}
	r := setupAttributeRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/agents/agent-1/attributes/missing", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
