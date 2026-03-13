package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/service"
)

type TaskHandlers struct {
	svc    service.TaskService
	logger *slog.Logger
}

func NewTaskHandlers(svc service.TaskService, logger *slog.Logger) *TaskHandlers {
	return &TaskHandlers{svc: svc, logger: logger}
}

type createTaskRequest struct {
	AgentID string          `json:"agent_id"`
	Prompt  string          `json:"prompt"`
	Context json.RawMessage `json:"context,omitempty"`
}

func (h *TaskHandlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	task, err := h.svc.Create(r.Context(), req.AgentID, req.Prompt, req.Context, nil)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (h *TaskHandlers) ListTasks(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")

	var (
		tasks []*domain.Task
		err   error
	)
	if agentID != "" {
		tasks, err = h.svc.ListByAgent(r.Context(), agentID)
	} else {
		tasks, err = h.svc.List(r.Context())
	}
	if err != nil {
		h.handleError(w, err)
		return
	}
	if tasks == nil {
		tasks = []*domain.Task{}
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandlers) GetTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "taskID")
	task, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

// ListAgentTasks lists tasks for a specific agent (nested under /agents/{id}/tasks).
func (h *TaskHandlers) ListAgentTasks(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")
	tasks, err := h.svc.ListByAgent(r.Context(), agentID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	if tasks == nil {
		tasks = []*domain.Task{}
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandlers) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrTaskNotFound):
		writeError(w, http.StatusNotFound, "task not found")
	case errors.Is(err, domain.ErrInvalidTask):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.logger.Error("internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
