package agent

import (
	"encoding/json"
	"io"
	"time"
)

// EventType identifies the kind of LLM event.
type EventType string

const (
	EventRequestStart  EventType = "llm.request.start"
	EventRequestEnd    EventType = "llm.request.end"
	EventTextDelta     EventType = "llm.text"
	EventToolUse       EventType = "llm.tool_use"
	EventToolResult    EventType = "llm.tool_result"
	EventError         EventType = "llm.error"
	EventLoopStart     EventType = "llm.loop.start"
	EventLoopEnd       EventType = "llm.loop.end"
)

// Event is a structured LLM event emitted as JSON for log ingestion.
type Event struct {
	Timestamp string    `json:"timestamp"`
	Type      EventType `json:"type"`
	AgentID   string    `json:"agent_id,omitempty"`
	Iteration int       `json:"iteration,omitempty"`

	// Request fields
	Model        string `json:"model,omitempty"`
	SystemPrompt string `json:"system_prompt,omitempty"`
	MessageCount int    `json:"message_count,omitempty"`
	ToolCount    int    `json:"tool_count,omitempty"`

	// Response fields
	StopReason    string `json:"stop_reason,omitempty"`
	InputTokens   int64  `json:"input_tokens,omitempty"`
	OutputTokens  int64  `json:"output_tokens,omitempty"`

	// Content fields
	Text      string          `json:"text,omitempty"`
	ToolName  string          `json:"tool_name,omitempty"`
	ToolID    string          `json:"tool_id,omitempty"`
	ToolInput json.RawMessage `json:"tool_input,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
	Result    string          `json:"result,omitempty"`

	// Error fields
	Error string `json:"error,omitempty"`

	// Loop summary
	Iterations int    `json:"iterations,omitempty"`
	FinalText  string `json:"final_text,omitempty"`
}

// EventWriter emits structured LLM events as newline-delimited JSON.
type EventWriter struct {
	w       io.Writer
	enc     *json.Encoder
	agentID string
}

// NewEventWriter creates an event writer that writes NDJSON to the given writer.
// Pass os.Stdout for console output, or any io.Writer for custom sinks.
func NewEventWriter(w io.Writer, agentID string) *EventWriter {
	return &EventWriter{
		w:       w,
		enc:     json.NewEncoder(w),
		agentID: agentID,
	}
}

// Emit writes a single event as a JSON line.
func (ew *EventWriter) Emit(event Event) {
	event.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	if event.AgentID == "" {
		event.AgentID = ew.agentID
	}
	ew.enc.Encode(event)
}

// NopEventWriter discards all events.
type NopEventWriter struct{}

func (NopEventWriter) Emit(Event) {}
