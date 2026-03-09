package service_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/mock"
	"github.com/robwittman/pillar/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttributeSet_Success(t *testing.T) {
	agentRepo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id}, nil
		},
	}
	var captured *domain.AgentAttribute
	attrRepo := &mock.AgentAttributeRepository{
		SetFn: func(ctx context.Context, attr *domain.AgentAttribute) error {
			captured = attr
			return nil
		},
	}
	svc := service.NewAttributeService(attrRepo, agentRepo, testWebhookLogger())

	value := json.RawMessage(`{"key":"value"}`)
	attr, err := svc.Set(context.Background(), "agent-1", "keycloak", value)
	require.NoError(t, err)
	assert.Equal(t, "agent-1", attr.AgentID)
	assert.Equal(t, "keycloak", attr.Namespace)
	assert.Equal(t, json.RawMessage(`{"key":"value"}`), attr.Value)
	assert.Equal(t, captured, attr)
}

func TestAttributeSet_AgentNotFound(t *testing.T) {
	agentRepo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return nil, domain.ErrAgentNotFound
		},
	}
	attrRepo := &mock.AgentAttributeRepository{}
	svc := service.NewAttributeService(attrRepo, agentRepo, testWebhookLogger())

	_, err := svc.Set(context.Background(), "missing", "ns", json.RawMessage(`{}`))
	assert.ErrorIs(t, err, domain.ErrAgentNotFound)
}

func TestAttributeGet_Success(t *testing.T) {
	expected := &domain.AgentAttribute{AgentID: "agent-1", Namespace: "vault"}
	attrRepo := &mock.AgentAttributeRepository{
		GetFn: func(ctx context.Context, agentID, namespace string) (*domain.AgentAttribute, error) {
			return expected, nil
		},
	}
	svc := service.NewAttributeService(attrRepo, nil, testWebhookLogger())

	attr, err := svc.Get(context.Background(), "agent-1", "vault")
	require.NoError(t, err)
	assert.Equal(t, expected, attr)
}

func TestAttributeGet_NotFound(t *testing.T) {
	attrRepo := &mock.AgentAttributeRepository{
		GetFn: func(ctx context.Context, agentID, namespace string) (*domain.AgentAttribute, error) {
			return nil, domain.ErrAttributeNotFound
		},
	}
	svc := service.NewAttributeService(attrRepo, nil, testWebhookLogger())

	_, err := svc.Get(context.Background(), "agent-1", "missing")
	assert.ErrorIs(t, err, domain.ErrAttributeNotFound)
}

func TestAttributeList_Success(t *testing.T) {
	expected := []*domain.AgentAttribute{{Namespace: "a"}, {Namespace: "b"}}
	attrRepo := &mock.AgentAttributeRepository{
		ListFn: func(ctx context.Context, agentID string) ([]*domain.AgentAttribute, error) {
			return expected, nil
		},
	}
	svc := service.NewAttributeService(attrRepo, nil, testWebhookLogger())

	attrs, err := svc.List(context.Background(), "agent-1")
	require.NoError(t, err)
	assert.Equal(t, expected, attrs)
}

func TestAttributeDelete_Success(t *testing.T) {
	attrRepo := &mock.AgentAttributeRepository{
		DeleteFn: func(ctx context.Context, agentID, namespace string) error {
			assert.Equal(t, "agent-1", agentID)
			assert.Equal(t, "vault", namespace)
			return nil
		},
	}
	svc := service.NewAttributeService(attrRepo, nil, testWebhookLogger())

	err := svc.Delete(context.Background(), "agent-1", "vault")
	assert.NoError(t, err)
}

func TestAttributeDelete_NotFound(t *testing.T) {
	attrRepo := &mock.AgentAttributeRepository{
		DeleteFn: func(ctx context.Context, agentID, namespace string) error {
			return domain.ErrAttributeNotFound
		},
	}
	svc := service.NewAttributeService(attrRepo, nil, testWebhookLogger())

	err := svc.Delete(context.Background(), "agent-1", "missing")
	assert.ErrorIs(t, err, domain.ErrAttributeNotFound)
}
