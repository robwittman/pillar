package grpc

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	pillarv1 "github.com/robwittman/pillar/gen/proto/pillar/v1"
	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/mock"
	"github.com/robwittman/pillar/internal/service"
)

var errEOF = io.EOF

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// mockHeartbeatService implements service.AgentService for gRPC tests.
type mockHeartbeatService struct {
	heartbeatCalled bool
	getAgent        *domain.Agent
	getErr          error
}

func (m *mockHeartbeatService) Create(ctx context.Context, name string, metadata, labels map[string]string) (*domain.Agent, error) {
	return nil, nil
}
func (m *mockHeartbeatService) Get(ctx context.Context, id string) (*domain.Agent, error) {
	if m.getAgent != nil || m.getErr != nil {
		return m.getAgent, m.getErr
	}
	return &domain.Agent{ID: id, Status: domain.AgentStatusPending}, nil
}
func (m *mockHeartbeatService) List(ctx context.Context) ([]*domain.Agent, error) {
	return nil, nil
}
func (m *mockHeartbeatService) Update(ctx context.Context, id string, name string, meta, labels map[string]string) (*domain.Agent, error) {
	return nil, nil
}
func (m *mockHeartbeatService) Delete(ctx context.Context, id string) error { return nil }
func (m *mockHeartbeatService) Start(ctx context.Context, id string) error  { return nil }
func (m *mockHeartbeatService) Stop(ctx context.Context, id string) error   { return nil }
func (m *mockHeartbeatService) Status(ctx context.Context, id string) (*service.AgentStatusInfo, error) {
	return nil, nil
}
func (m *mockHeartbeatService) Heartbeat(ctx context.Context, agentID string) error {
	m.heartbeatCalled = true
	return nil
}

// mockAgentStream implements pillarv1.AgentStreamService_AgentStreamServer.
type mockAgentStream struct {
	recvMsgs []*pillarv1.AgentMessage
	recvIdx  int
	sent     []*pillarv1.ServerMessage
	recvErr  error
	mu       sync.Mutex
	ctx      context.Context
}

func newMockStream(msgs []*pillarv1.AgentMessage, finalErr error) *mockAgentStream {
	return &mockAgentStream{
		recvMsgs: msgs,
		recvErr:  finalErr,
		ctx:      context.Background(),
	}
}

func (m *mockAgentStream) Recv() (*pillarv1.AgentMessage, error) {
	if m.recvIdx < len(m.recvMsgs) {
		msg := m.recvMsgs[m.recvIdx]
		m.recvIdx++
		return msg, nil
	}
	return nil, m.recvErr
}

func (m *mockAgentStream) Send(msg *pillarv1.ServerMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, msg)
	return nil
}

func (m *mockAgentStream) Context() context.Context     { return m.ctx }
func (m *mockAgentStream) SetHeader(metadata.MD) error  { return nil }
func (m *mockAgentStream) SendHeader(metadata.MD) error { return nil }
func (m *mockAgentStream) SetTrailer(metadata.MD)       {}
func (m *mockAgentStream) SendMsg(interface{}) error    { return nil }
func (m *mockAgentStream) RecvMsg(interface{}) error    { return nil }

// --- StreamManager unit tests ---

func TestStreamManager_AddGetRemove(t *testing.T) {
	sm := NewStreamManager()

	_, ok := sm.Get("agent1")
	assert.False(t, ok)

	sm.Add("agent1", nil)
	s, ok := sm.Get("agent1")
	assert.True(t, ok)
	assert.Nil(t, s)

	sm.Remove("agent1")
	_, ok = sm.Get("agent1")
	assert.False(t, ok)
}

func TestStreamManager_SendToAgent_NotFound(t *testing.T) {
	sm := NewStreamManager()
	err := sm.SendToAgent("missing", &pillarv1.ServerMessage{})
	assert.NoError(t, err)
}

func TestStreamManager_ConcurrentAccess(t *testing.T) {
	sm := NewStreamManager()
	var wg sync.WaitGroup
	const n = 100

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(id int) {
			defer wg.Done()
			sm.Add(string(rune('a'+id%26)), nil)
		}(i)
	}
	wg.Wait()

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(id int) {
			defer wg.Done()
			sm.Get(string(rune('a' + id%26)))
		}(i)
	}
	wg.Wait()

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(id int) {
			defer wg.Done()
			sm.Remove(string(rune('a' + id%26)))
		}(i)
	}
	wg.Wait()
}

// --- AgentStream handler tests ---

func TestAgentStream_ConnectAck(t *testing.T) {
	svc := &mockHeartbeatService{}
	ss := NewAgentStreamService(svc, nil, nil, nil, nil, testLogger())

	stream := newMockStream([]*pillarv1.AgentMessage{
		{Payload: &pillarv1.AgentMessage_Connect{
			Connect: &pillarv1.ConnectRequest{AgentId: "agent1"},
		}},
	}, errEOF)

	err := ss.AgentStream(stream)
	assert.NoError(t, err)

	require.Len(t, stream.sent, 1)
	ack := stream.sent[0].GetConnectAck()
	require.NotNil(t, ack)
	assert.True(t, ack.Accepted)
	assert.Equal(t, int32(15), ack.HeartbeatIntervalSeconds)
}

func TestAgentStream_HeartbeatAck(t *testing.T) {
	svc := &mockHeartbeatService{}
	ss := NewAgentStreamService(svc, nil, nil, nil, nil, testLogger())

	stream := newMockStream([]*pillarv1.AgentMessage{
		{Payload: &pillarv1.AgentMessage_Connect{
			Connect: &pillarv1.ConnectRequest{AgentId: "agent1"},
		}},
		{Payload: &pillarv1.AgentMessage_Heartbeat{
			Heartbeat: &pillarv1.Heartbeat{AgentId: "agent1"},
		}},
	}, errEOF)

	err := ss.AgentStream(stream)
	assert.NoError(t, err)

	require.Len(t, stream.sent, 2)
	hbAck := stream.sent[1].GetHeartbeatAck()
	require.NotNil(t, hbAck)
	assert.True(t, svc.heartbeatCalled)
}

func TestAgentStream_EventLogged(t *testing.T) {
	svc := &mockHeartbeatService{}
	ss := NewAgentStreamService(svc, nil, nil, nil, nil, testLogger())

	stream := newMockStream([]*pillarv1.AgentMessage{
		{Payload: &pillarv1.AgentMessage_Event{
			Event: &pillarv1.EventReport{AgentId: "agent1", EventType: "test", Payload: "data"},
		}},
	}, errEOF)

	err := ss.AgentStream(stream)
	assert.NoError(t, err)
	assert.Len(t, stream.sent, 0)
}

func TestAgentStream_TaskResultLogged(t *testing.T) {
	svc := &mockHeartbeatService{}
	ss := NewAgentStreamService(svc, nil, nil, nil, nil, testLogger())

	stream := newMockStream([]*pillarv1.AgentMessage{
		{Payload: &pillarv1.AgentMessage_TaskResult{
			TaskResult: &pillarv1.TaskResult{AgentId: "agent1", TaskId: "t1", Success: true},
		}},
	}, errEOF)

	err := ss.AgentStream(stream)
	assert.NoError(t, err)
	assert.Len(t, stream.sent, 0)
}

func TestAgentStream_ErrorCleanup(t *testing.T) {
	svc := &mockHeartbeatService{}
	ss := NewAgentStreamService(svc, nil, nil, nil, nil, testLogger())

	streamErr := assert.AnError
	stream := newMockStream([]*pillarv1.AgentMessage{
		{Payload: &pillarv1.AgentMessage_Connect{
			Connect: &pillarv1.ConnectRequest{AgentId: "agent1"},
		}},
	}, streamErr)

	err := ss.AgentStream(stream)
	assert.Error(t, err)

	_, ok := ss.streams.Get("agent1")
	assert.False(t, ok)
}

func TestAgentStream_EOFCleanup(t *testing.T) {
	svc := &mockHeartbeatService{}
	ss := NewAgentStreamService(svc, nil, nil, nil, nil, testLogger())

	stream := newMockStream([]*pillarv1.AgentMessage{
		{Payload: &pillarv1.AgentMessage_Connect{
			Connect: &pillarv1.ConnectRequest{AgentId: "agent1"},
		}},
	}, errEOF)

	err := ss.AgentStream(stream)
	assert.NoError(t, err)

	_, ok := ss.streams.Get("agent1")
	assert.False(t, ok)
}

// --- Config delivery tests ---

func TestAgentStream_ConnectAckWithConfig(t *testing.T) {
	svc := &mockHeartbeatService{}
	configSvc := &mock.ConfigService{
		GetConfigWithSecretsFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, string, error) {
			return &domain.AgentConfig{
				AgentID:       agentID,
				ModelProvider: domain.ModelProviderClaude,
				ModelID:       "claude-sonnet-4-20250514",
				SystemPrompt:  "You are a helpful assistant.",
				MaxIterations: 100,
				ModelParams:   domain.ModelParams{Temperature: 0.7, MaxTokens: 4096},
				MCPServers: []domain.MCPServerConfig{
					{Name: "fs", TransportType: domain.MCPTransportStdio, Command: "mcp-fs"},
				},
				ToolPermissions: domain.ToolPermissions{AllowedTools: []string{"read_file"}},
				EscalationRules: []domain.EscalationRule{
					{Name: "error-limit", Condition: "error_count > 3", Action: domain.EscalationActionPause},
				},
			}, "sk-resolved-key", nil
		},
	}
	ss := NewAgentStreamService(svc, configSvc, nil, nil, nil, testLogger())

	stream := newMockStream([]*pillarv1.AgentMessage{
		{Payload: &pillarv1.AgentMessage_Connect{
			Connect: &pillarv1.ConnectRequest{AgentId: "agent1"},
		}},
	}, errEOF)

	err := ss.AgentStream(stream)
	assert.NoError(t, err)

	require.Len(t, stream.sent, 1)
	ack := stream.sent[0].GetConnectAck()
	require.NotNil(t, ack)
	assert.True(t, ack.Accepted)

	cfg := ack.Config
	require.NotNil(t, cfg)
	assert.Equal(t, "agent1", cfg.AgentId)
	assert.Equal(t, "claude", cfg.ModelProvider)
	assert.Equal(t, "claude-sonnet-4-20250514", cfg.ModelId)
	assert.Equal(t, "You are a helpful assistant.", cfg.SystemPrompt)
	assert.Equal(t, "sk-resolved-key", cfg.ApiCredential)
	assert.Equal(t, int32(100), cfg.MaxIterations)
	assert.InDelta(t, 0.7, cfg.ModelParams.Temperature, 0.001)
	assert.Equal(t, int32(4096), cfg.ModelParams.MaxTokens)
	require.Len(t, cfg.McpServers, 1)
	assert.Equal(t, "fs", cfg.McpServers[0].Name)
	assert.Equal(t, "stdio", cfg.McpServers[0].TransportType)
	require.NotNil(t, cfg.ToolPermissions)
	assert.Equal(t, []string{"read_file"}, cfg.ToolPermissions.AllowedTools)
	require.Len(t, cfg.EscalationRules, 1)
	assert.Equal(t, "error-limit", cfg.EscalationRules[0].Name)
}

func TestAgentStream_ConnectAckWithoutConfig(t *testing.T) {
	svc := &mockHeartbeatService{}
	configSvc := &mock.ConfigService{
		GetConfigWithSecretsFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, string, error) {
			return nil, "", domain.ErrConfigNotFound
		},
	}
	ss := NewAgentStreamService(svc, configSvc, nil, nil, nil, testLogger())

	stream := newMockStream([]*pillarv1.AgentMessage{
		{Payload: &pillarv1.AgentMessage_Connect{
			Connect: &pillarv1.ConnectRequest{AgentId: "agent1"},
		}},
	}, errEOF)

	err := ss.AgentStream(stream)
	assert.NoError(t, err)

	require.Len(t, stream.sent, 1)
	ack := stream.sent[0].GetConnectAck()
	require.NotNil(t, ack)
	assert.True(t, ack.Accepted)
	assert.Nil(t, ack.Config)
}

func TestAgentStream_ConnectAckNilConfigService(t *testing.T) {
	svc := &mockHeartbeatService{}
	ss := NewAgentStreamService(svc, nil, nil, nil, nil, testLogger())

	stream := newMockStream([]*pillarv1.AgentMessage{
		{Payload: &pillarv1.AgentMessage_Connect{
			Connect: &pillarv1.ConnectRequest{AgentId: "agent1"},
		}},
	}, errEOF)

	err := ss.AgentStream(stream)
	assert.NoError(t, err)

	require.Len(t, stream.sent, 1)
	ack := stream.sent[0].GetConnectAck()
	require.NotNil(t, ack)
	assert.True(t, ack.Accepted)
	assert.Nil(t, ack.Config)
}

// --- ConnectAck status tests ---

func TestAgentStream_ConnectAckStatusRunning(t *testing.T) {
	svc := &mockHeartbeatService{
		getAgent: &domain.Agent{ID: "agent1", Status: domain.AgentStatusRunning},
	}
	ss := NewAgentStreamService(svc, nil, nil, nil, nil, testLogger())

	stream := newMockStream([]*pillarv1.AgentMessage{
		{Payload: &pillarv1.AgentMessage_Connect{
			Connect: &pillarv1.ConnectRequest{AgentId: "agent1"},
		}},
	}, errEOF)

	err := ss.AgentStream(stream)
	assert.NoError(t, err)

	require.Len(t, stream.sent, 1)
	ack := stream.sent[0].GetConnectAck()
	require.NotNil(t, ack)
	assert.Equal(t, pillarv1.AgentStatus_AGENT_STATUS_RUNNING, ack.Status)
}

func TestAgentStream_ConnectAckStatusPending(t *testing.T) {
	svc := &mockHeartbeatService{
		getAgent: &domain.Agent{ID: "agent1", Status: domain.AgentStatusPending},
	}
	ss := NewAgentStreamService(svc, nil, nil, nil, nil, testLogger())

	stream := newMockStream([]*pillarv1.AgentMessage{
		{Payload: &pillarv1.AgentMessage_Connect{
			Connect: &pillarv1.ConnectRequest{AgentId: "agent1"},
		}},
	}, errEOF)

	err := ss.AgentStream(stream)
	assert.NoError(t, err)

	require.Len(t, stream.sent, 1)
	ack := stream.sent[0].GetConnectAck()
	require.NotNil(t, ack)
	assert.Equal(t, pillarv1.AgentStatus_AGENT_STATUS_PENDING, ack.Status)
}

func TestAgentStream_ConnectAckStatusDefaultsOnNotFound(t *testing.T) {
	svc := &mockHeartbeatService{
		getErr: domain.ErrAgentNotFound,
	}
	ss := NewAgentStreamService(svc, nil, nil, nil, nil, testLogger())

	stream := newMockStream([]*pillarv1.AgentMessage{
		{Payload: &pillarv1.AgentMessage_Connect{
			Connect: &pillarv1.ConnectRequest{AgentId: "missing"},
		}},
	}, errEOF)

	err := ss.AgentStream(stream)
	assert.NoError(t, err)

	require.Len(t, stream.sent, 1)
	ack := stream.sent[0].GetConnectAck()
	require.NotNil(t, ack)
	assert.Equal(t, pillarv1.AgentStatus_AGENT_STATUS_PENDING, ack.Status)
}

// --- StreamNotifier tests ---

func TestStreamNotifier_SendsDirective(t *testing.T) {
	sm := NewStreamManager()
	stream := newMockStream(nil, errEOF)
	sm.Add("agent1", stream)

	notifier := NewStreamNotifier(sm, testLogger())
	err := notifier.NotifyDirective("agent1", "start", "")
	assert.NoError(t, err)

	require.Len(t, stream.sent, 1)
	directive := stream.sent[0].GetDirective()
	require.NotNil(t, directive)
	assert.Equal(t, "start", directive.Type)
	assert.Equal(t, "", directive.Payload)
}

func TestStreamNotifier_AgentNotConnected(t *testing.T) {
	sm := NewStreamManager()
	notifier := NewStreamNotifier(sm, testLogger())

	err := notifier.NotifyDirective("missing", "start", "")
	assert.NoError(t, err)
}

// --- ConnectAck with attributes ---

func TestAgentStream_ConnectAckWithAttributes(t *testing.T) {
	svc := &mockHeartbeatService{}
	attrSvc := &mock.AttributeService{
		ListFn: func(ctx context.Context, agentID string) ([]*domain.AgentAttribute, error) {
			return []*domain.AgentAttribute{
				{AgentID: agentID, Namespace: "vault", Value: []byte(`{"token":"abc"}`)},
				{AgentID: agentID, Namespace: "keycloak", Value: []byte(`{"client_id":"xyz"}`)},
			}, nil
		},
	}
	ss := NewAgentStreamService(svc, nil, attrSvc, nil, nil, testLogger())

	stream := newMockStream([]*pillarv1.AgentMessage{
		{Payload: &pillarv1.AgentMessage_Connect{
			Connect: &pillarv1.ConnectRequest{AgentId: "agent1"},
		}},
	}, errEOF)

	err := ss.AgentStream(stream)
	assert.NoError(t, err)

	require.Len(t, stream.sent, 1)
	ack := stream.sent[0].GetConnectAck()
	require.NotNil(t, ack)
	assert.True(t, ack.Accepted)
	require.Len(t, ack.Attributes, 2)
	assert.Equal(t, []byte(`{"token":"abc"}`), ack.Attributes["vault"])
	assert.Equal(t, []byte(`{"client_id":"xyz"}`), ack.Attributes["keycloak"])
}

func TestAgentStream_ConnectAckNilAttributeService(t *testing.T) {
	svc := &mockHeartbeatService{}
	ss := NewAgentStreamService(svc, nil, nil, nil, nil, testLogger())

	stream := newMockStream([]*pillarv1.AgentMessage{
		{Payload: &pillarv1.AgentMessage_Connect{
			Connect: &pillarv1.ConnectRequest{AgentId: "agent1"},
		}},
	}, errEOF)

	err := ss.AgentStream(stream)
	assert.NoError(t, err)

	require.Len(t, stream.sent, 1)
	ack := stream.sent[0].GetConnectAck()
	require.NotNil(t, ack)
	assert.Nil(t, ack.Attributes)
}
