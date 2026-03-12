package agent

import (
	"encoding/json"
	"log/slog"
	"time"
)

// GrpcEventSender is the subset of client.Client needed to send events.
type GrpcEventSender interface {
	SendEvent(eventType, payload string) error
}

// GrpcEmitter sends structured LLM events back to Pillar via the gRPC stream.
type GrpcEmitter struct {
	sender  GrpcEventSender
	agentID string
	logger  *slog.Logger
}

// NewGrpcEmitter creates an emitter that sends LLM events over gRPC.
func NewGrpcEmitter(sender GrpcEventSender, agentID string, logger *slog.Logger) *GrpcEmitter {
	return &GrpcEmitter{sender: sender, agentID: agentID, logger: logger}
}

func (e *GrpcEmitter) Emit(event Event) {
	event.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	if event.AgentID == "" {
		event.AgentID = e.agentID
	}

	payload, err := json.Marshal(event)
	if err != nil {
		e.logger.Warn("failed to marshal event", "error", err)
		return
	}

	if err := e.sender.SendEvent(string(event.Type), string(payload)); err != nil {
		e.logger.Warn("failed to send event via gRPC", "type", event.Type, "error", err)
	}
}

// MultiEmitter fans out events to multiple emitters.
type MultiEmitter struct {
	emitters []Emitter
}

// NewMultiEmitter creates an emitter that sends to all provided emitters.
func NewMultiEmitter(emitters ...Emitter) *MultiEmitter {
	return &MultiEmitter{emitters: emitters}
}

func (m *MultiEmitter) Emit(event Event) {
	for _, e := range m.emitters {
		e.Emit(event)
	}
}
