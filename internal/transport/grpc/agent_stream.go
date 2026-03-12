package grpc

import (
	"io"
	"log/slog"
	"sync"

	"google.golang.org/protobuf/types/known/timestamppb"

	pillarv1 "github.com/robwittman/pillar/gen/proto/pillar/v1"
	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/service"
)

type AgentStreamService struct {
	pillarv1.UnimplementedAgentStreamServiceServer
	svc       service.AgentService
	configSvc service.ConfigService
	attrSvc   service.AttributeService
	logSvc    *service.LogService
	logger    *slog.Logger
	streams   *StreamManager
}

func NewAgentStreamService(svc service.AgentService, configSvc service.ConfigService, attrSvc service.AttributeService, logSvc *service.LogService, logger *slog.Logger) *AgentStreamService {
	return &AgentStreamService{
		svc:       svc,
		configSvc: configSvc,
		attrSvc:   attrSvc,
		logSvc:    logSvc,
		logger:    logger,
		streams:   NewStreamManager(),
	}
}

// NewAgentStreamServiceWithStreams creates an AgentStreamService using an externally
// provided StreamManager. This allows main.go to share the StreamManager with a
// StreamNotifier so that Start/Stop directives can reach connected agents.
func NewAgentStreamServiceWithStreams(svc service.AgentService, configSvc service.ConfigService, attrSvc service.AttributeService, logSvc *service.LogService, streams *StreamManager, logger *slog.Logger) *AgentStreamService {
	return &AgentStreamService{
		svc:       svc,
		configSvc: configSvc,
		attrSvc:   attrSvc,
		logSvc:    logSvc,
		logger:    logger,
		streams:   streams,
	}
}

// StreamNotifier implements service.AgentNotifier by sending directives over
// gRPC streams managed by a StreamManager.
type StreamNotifier struct {
	streams *StreamManager
	logger  *slog.Logger
}

func NewStreamNotifier(streams *StreamManager, logger *slog.Logger) *StreamNotifier {
	return &StreamNotifier{streams: streams, logger: logger}
}

func (n *StreamNotifier) NotifyDirective(agentID, directiveType, payload string) error {
	return n.streams.SendToAgent(agentID, &pillarv1.ServerMessage{
		Payload: &pillarv1.ServerMessage_Directive{
			Directive: &pillarv1.Directive{Type: directiveType, Payload: payload},
		},
	})
}

func (s *AgentStreamService) AgentStream(stream pillarv1.AgentStreamService_AgentStreamServer) error {
	var agentID string

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			if agentID != "" {
				s.streams.Remove(agentID)
				s.logger.Info("agent stream closed", "agent_id", agentID)
			}
			return nil
		}
		if err != nil {
			if agentID != "" {
				s.streams.Remove(agentID)
				s.logger.Warn("agent stream error", "agent_id", agentID, "error", err)
			}
			return err
		}

		switch p := msg.Payload.(type) {
		case *pillarv1.AgentMessage_Connect:
			agentID = p.Connect.AgentId
			s.streams.Add(agentID, stream)
			s.logger.Info("agent connected", "agent_id", agentID)

			ack := &pillarv1.ConnectAck{
				Accepted:                 true,
				HeartbeatIntervalSeconds: 15,
				Message:                  "connected to pillar",
				Status:                   pillarv1.AgentStatus_AGENT_STATUS_PENDING,
			}

			if agent, err := s.svc.Get(stream.Context(), agentID); err != nil {
				s.logger.Debug("could not look up agent status on connect", "agent_id", agentID, "error", err)
			} else {
				ack.Status = domainStatusToProto(agent.Status)
			}

			if s.configSvc != nil {
				cfg, credential, err := s.configSvc.GetConfigWithSecrets(stream.Context(), agentID)
				if err != nil {
					s.logger.Debug("no config for agent", "agent_id", agentID, "error", err)
				} else {
					ack.Config = toProtoConfig(cfg, credential)
				}
			}

			if s.attrSvc != nil {
				attrs, err := s.attrSvc.List(stream.Context(), agentID)
				if err != nil {
					s.logger.Debug("failed to fetch attributes", "agent_id", agentID, "error", err)
				} else if len(attrs) > 0 {
					ack.Attributes = make(map[string][]byte, len(attrs))
					for _, attr := range attrs {
						ack.Attributes[attr.Namespace] = attr.Value
					}
				}
			}

			if err := stream.Send(&pillarv1.ServerMessage{
				Payload: &pillarv1.ServerMessage_ConnectAck{
					ConnectAck: ack,
				},
			}); err != nil {
				return err
			}

		case *pillarv1.AgentMessage_Heartbeat:
			ctx := stream.Context()
			if err := s.svc.Heartbeat(ctx, p.Heartbeat.AgentId); err != nil {
				s.logger.Warn("heartbeat failed", "agent_id", p.Heartbeat.AgentId, "error", err)
			}

			if err := stream.Send(&pillarv1.ServerMessage{
				Payload: &pillarv1.ServerMessage_HeartbeatAck{
					HeartbeatAck: &pillarv1.HeartbeatAck{
						ServerTime: timestamppb.Now(),
					},
				},
			}); err != nil {
				return err
			}

		case *pillarv1.AgentMessage_Event:
			s.logger.Debug("agent event",
				"agent_id", p.Event.AgentId,
				"type", p.Event.EventType,
			)
			if s.logSvc != nil {
				s.logSvc.Ingest(stream.Context(), p.Event.AgentId, p.Event.Payload)
			}

		case *pillarv1.AgentMessage_TaskResult:
			s.logger.Info("task result",
				"agent_id", p.TaskResult.AgentId,
				"task_id", p.TaskResult.TaskId,
				"success", p.TaskResult.Success,
			)
		}
	}
}

// StreamManager tracks active gRPC streams per agent.
type StreamManager struct {
	mu      sync.RWMutex
	streams map[string]pillarv1.AgentStreamService_AgentStreamServer
}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		streams: make(map[string]pillarv1.AgentStreamService_AgentStreamServer),
	}
}

func (m *StreamManager) Add(agentID string, stream pillarv1.AgentStreamService_AgentStreamServer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.streams[agentID] = stream
}

func (m *StreamManager) Remove(agentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.streams, agentID)
}

func (m *StreamManager) Get(agentID string) (pillarv1.AgentStreamService_AgentStreamServer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.streams[agentID]
	return s, ok
}

func (m *StreamManager) SendToAgent(agentID string, msg *pillarv1.ServerMessage) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.streams[agentID]
	if !ok {
		return nil
	}
	return s.Send(msg)
}

func domainStatusToProto(s domain.AgentStatus) pillarv1.AgentStatus {
	switch s {
	case domain.AgentStatusPending:
		return pillarv1.AgentStatus_AGENT_STATUS_PENDING
	case domain.AgentStatusRunning:
		return pillarv1.AgentStatus_AGENT_STATUS_RUNNING
	case domain.AgentStatusStopped:
		return pillarv1.AgentStatus_AGENT_STATUS_STOPPED
	case domain.AgentStatusError:
		return pillarv1.AgentStatus_AGENT_STATUS_ERROR
	default:
		return pillarv1.AgentStatus_AGENT_STATUS_UNSPECIFIED
	}
}

func toProtoConfig(cfg *domain.AgentConfig, credential string) *pillarv1.AgentConfig {
	if cfg == nil {
		return nil
	}

	pc := &pillarv1.AgentConfig{
		AgentId:            cfg.AgentID,
		ModelProvider:      string(cfg.ModelProvider),
		ModelId:            cfg.ModelID,
		SystemPrompt:       cfg.SystemPrompt,
		ApiCredential:      credential,
		MaxIterations:      int32(cfg.MaxIterations),
		TokenBudget:        int32(cfg.TokenBudget),
		TaskTimeoutSeconds: int32(cfg.TaskTimeoutSeconds),
		ModelParams: &pillarv1.ModelParams{
			Temperature: cfg.ModelParams.Temperature,
			TopP:        cfg.ModelParams.TopP,
			MaxTokens:   int32(cfg.ModelParams.MaxTokens),
		},
		ToolPermissions: &pillarv1.ToolPermissions{
			AllowedTools: cfg.ToolPermissions.AllowedTools,
			DeniedTools:  cfg.ToolPermissions.DeniedTools,
		},
	}

	for _, mcp := range cfg.MCPServers {
		pc.McpServers = append(pc.McpServers, &pillarv1.MCPServerConfig{
			Name:          mcp.Name,
			TransportType: string(mcp.TransportType),
			Command:       mcp.Command,
			Args:          mcp.Args,
			Url:           mcp.URL,
			Headers:       mcp.Headers,
			Env:           mcp.Env,
		})
	}

	for _, rule := range cfg.EscalationRules {
		pc.EscalationRules = append(pc.EscalationRules, &pillarv1.EscalationRule{
			Name:      rule.Name,
			Condition: rule.Condition,
			Action:    string(rule.Action),
			Message:   rule.Message,
		})
	}

	return pc
}
