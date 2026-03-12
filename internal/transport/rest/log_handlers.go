package rest

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/robwittman/pillar/internal/service"
)

type LogHandlers struct {
	logSvc *service.LogService
	logger *slog.Logger
}

func NewLogHandlers(logSvc *service.LogService, logger *slog.Logger) *LogHandlers {
	return &LogHandlers{logSvc: logSvc, logger: logger}
}

// GetLogs returns historical log entries for an agent.
// Query params: since (unix nanoseconds), limit (int, default 200).
func (h *LogHandlers) GetLogs(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")

	var sinceNano int64
	if s := r.URL.Query().Get("since"); s != "" {
		sinceNano, _ = strconv.ParseInt(s, 10, 64)
	}

	limit := 200
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}

	entries, err := h.logSvc.Query(r.Context(), agentID, sinceNano, limit)
	if err != nil {
		h.logger.Error("failed to query logs", "agent_id", agentID, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to query logs")
		return
	}

	writeJSON(w, http.StatusOK, entries)
}

// StreamLogs provides a Server-Sent Events stream of real-time log entries.
func (h *LogHandlers) StreamLogs(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	sub := h.logSvc.Subscribe(agentID)
	defer h.logSvc.Unsubscribe(sub)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-sub.Ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", entry)
			flusher.Flush()
		}
	}
}
