package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/robwittman/pillar/internal/auth"
	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/service"
)

type OrgHandlers struct {
	orgSvc service.OrgService
	logger *slog.Logger
}

func NewOrgHandlers(orgSvc service.OrgService, logger *slog.Logger) *OrgHandlers {
	return &OrgHandlers{orgSvc: orgSvc, logger: logger}
}

// --- Organization CRUD ---

func (h *OrgHandlers) CreateOrg(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Slug == "" {
		writeError(w, http.StatusBadRequest, "name and slug are required")
		return
	}

	org, err := h.orgSvc.Create(r.Context(), req.Name, req.Slug)
	if err != nil {
		h.logger.Error("failed to create org", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create organization")
		return
	}
	writeJSON(w, http.StatusCreated, org)
}

func (h *OrgHandlers) ListOrgs(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	orgs, err := h.orgSvc.ListByUser(r.Context(), principal.ID)
	if err != nil {
		h.logger.Error("failed to list orgs", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list organizations")
		return
	}
	if orgs == nil {
		orgs = []*domain.Organization{}
	}
	writeJSON(w, http.StatusOK, orgs)
}

func (h *OrgHandlers) GetOrg(w http.ResponseWriter, r *http.Request) {
	org, err := h.orgSvc.Get(r.Context(), chi.URLParam(r, "orgID"))
	if err != nil {
		writeError(w, http.StatusNotFound, "organization not found")
		return
	}
	writeJSON(w, http.StatusOK, org)
}

func (h *OrgHandlers) UpdateOrg(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	org, err := h.orgSvc.Update(r.Context(), chi.URLParam(r, "orgID"), req.Name, req.Slug)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, org)
}

func (h *OrgHandlers) DeleteOrg(w http.ResponseWriter, r *http.Request) {
	if err := h.orgSvc.Delete(r.Context(), chi.URLParam(r, "orgID")); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- Members ---

func (h *OrgHandlers) ListMembers(w http.ResponseWriter, r *http.Request) {
	members, err := h.orgSvc.ListMembers(r.Context(), chi.URLParam(r, "orgID"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list members")
		return
	}
	if members == nil {
		members = []*domain.Membership{}
	}
	writeJSON(w, http.StatusOK, members)
}

func (h *OrgHandlers) AddMember(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string         `json:"user_id"`
		Role   domain.OrgRole `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	if req.Role == "" {
		req.Role = domain.OrgRoleMember
	}

	m, err := h.orgSvc.AddMember(r.Context(), chi.URLParam(r, "orgID"), req.UserID, req.Role)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *OrgHandlers) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Role domain.OrgRole `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Role == "" {
		writeError(w, http.StatusBadRequest, "role is required")
		return
	}

	err := h.orgSvc.UpdateMemberRole(r.Context(), chi.URLParam(r, "orgID"), chi.URLParam(r, "userID"), req.Role)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *OrgHandlers) RemoveMember(w http.ResponseWriter, r *http.Request) {
	err := h.orgSvc.RemoveMember(r.Context(), chi.URLParam(r, "orgID"), chi.URLParam(r, "userID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- Teams ---

func (h *OrgHandlers) ListTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := h.orgSvc.ListTeams(r.Context(), chi.URLParam(r, "orgID"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list teams")
		return
	}
	if teams == nil {
		teams = []*domain.Team{}
	}
	writeJSON(w, http.StatusOK, teams)
}

func (h *OrgHandlers) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	t, err := h.orgSvc.CreateTeam(r.Context(), chi.URLParam(r, "orgID"), req.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (h *OrgHandlers) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	if err := h.orgSvc.DeleteTeam(r.Context(), chi.URLParam(r, "teamID")); err != nil {
		writeError(w, http.StatusNotFound, "team not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *OrgHandlers) AddTeamMember(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	err := h.orgSvc.AddTeamMember(r.Context(), chi.URLParam(r, "teamID"), req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *OrgHandlers) RemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	err := h.orgSvc.RemoveTeamMember(r.Context(), chi.URLParam(r, "teamID"), chi.URLParam(r, "userID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *OrgHandlers) ListTeamMembers(w http.ResponseWriter, r *http.Request) {
	members, err := h.orgSvc.ListTeamMembers(r.Context(), chi.URLParam(r, "teamID"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list team members")
		return
	}
	if members == nil {
		members = []*domain.TeamMembership{}
	}
	writeJSON(w, http.StatusOK, members)
}
