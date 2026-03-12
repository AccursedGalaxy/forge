package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/accursedgalaxy/forge/internal/db"
)

// ProjectHandler handles project CRUD endpoints.
type ProjectHandler struct {
	db   *db.Queries
	pool *pgxpool.Pool
}

// List returns all projects.
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.db.ListProjects(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list projects")
		return
	}
	if projects == nil {
		projects = []db.Project{}
	}
	writeJSON(w, http.StatusOK, projects)
}

// Create inserts a new project owned by the default user.
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		RepoURL     string `json:"repo_url"`
		LocalPath   string `json:"local_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if body.LocalPath == "" {
		writeError(w, http.StatusBadRequest, "local_path is required")
		return
	}

	project, err := h.db.CreateProject(r.Context(), db.CreateProjectParams{
		OwnerID:     db.DefaultUserID,
		Name:        body.Name,
		Description: body.Description,
		RepoUrl:     body.RepoURL,
		LocalPath:   body.LocalPath,
		Status:      "active",
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create project")
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

// Get returns a single project with task counts.
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	project, err := h.db.GetProjectWithTaskCounts(r.Context(), id)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get project")
		return
	}
	writeJSON(w, http.StatusOK, project)
}

// Update patches a project's mutable fields.
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var body struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		RepoURL     *string `json:"repo_url"`
		LocalPath   *string `json:"local_path"`
		Status      *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	params := db.UpdateProjectParams{ID: id}
	if body.Name != nil {
		params.Name = sql.NullString{String: *body.Name, Valid: true}
	}
	if body.Description != nil {
		params.Description = sql.NullString{String: *body.Description, Valid: true}
	}
	if body.RepoURL != nil {
		params.RepoUrl = sql.NullString{String: *body.RepoURL, Valid: true}
	}
	if body.LocalPath != nil {
		params.LocalPath = sql.NullString{String: *body.LocalPath, Valid: true}
	}
	if body.Status != nil {
		params.Status = sql.NullString{String: *body.Status, Valid: true}
	}

	project, err := h.db.UpdateProject(r.Context(), params)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update project")
		return
	}
	writeJSON(w, http.StatusOK, project)
}

// Delete removes a project and all its tasks (cascade).
func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	if err := h.db.DeleteProject(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete project")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
