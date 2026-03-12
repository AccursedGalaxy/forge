package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sqlc-dev/pqtype"

	"github.com/accursedgalaxy/forge/internal/db"
)

// ProviderHandler handles provider CRUD endpoints.
type ProviderHandler struct {
	db   *db.Queries
	pool *pgxpool.Pool
}

// List returns all configured providers.
func (h *ProviderHandler) List(w http.ResponseWriter, r *http.Request) {
	providers, err := h.db.ListProviders(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list providers")
		return
	}
	if providers == nil {
		providers = []db.AgentProvider{}
	}
	writeJSON(w, http.StatusOK, providers)
}

// Create inserts a new provider configuration.
func (h *ProviderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name         string          `json:"name"`
		ProviderType string          `json:"provider_type"`
		Config       json.RawMessage `json:"config"`
		IsDefault    bool            `json:"is_default"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if body.ProviderType == "" {
		body.ProviderType = "claude"
	}
	if body.Config == nil {
		body.Config = json.RawMessage("{}")
	}

	// If this should be default, clear existing default first
	if body.IsDefault {
		if err := h.db.ClearDefaultProvider(r.Context()); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to clear default provider")
			return
		}
	}

	provider, err := h.db.CreateProvider(r.Context(), db.CreateProviderParams{
		Name:         body.Name,
		ProviderType: body.ProviderType,
		Config:       body.Config,
		IsDefault:    body.IsDefault,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create provider")
		return
	}
	writeJSON(w, http.StatusCreated, provider)
}

// Update patches a provider. If is_default is set to true, clears other defaults first.
func (h *ProviderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid provider id")
		return
	}

	var body struct {
		Name         *string          `json:"name"`
		ProviderType *string          `json:"provider_type"`
		Config       *json.RawMessage `json:"config"`
		IsDefault    *bool            `json:"is_default"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Clear default on all providers before setting a new one
	if body.IsDefault != nil && *body.IsDefault {
		if err := h.db.ClearDefaultProvider(r.Context()); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to clear default provider")
			return
		}
	}

	params := db.UpdateProviderParams{ID: id}
	if body.Name != nil {
		params.Name = sql.NullString{String: *body.Name, Valid: true}
	}
	if body.ProviderType != nil {
		params.ProviderType = sql.NullString{String: *body.ProviderType, Valid: true}
	}
	if body.Config != nil {
		params.Config = pqtype.NullRawMessage{RawMessage: *body.Config, Valid: true}
	}
	if body.IsDefault != nil {
		params.IsDefault = sql.NullBool{Bool: *body.IsDefault, Valid: true}
	}

	provider, err := h.db.UpdateProvider(r.Context(), params)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "provider not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update provider")
		return
	}
	writeJSON(w, http.StatusOK, provider)
}
