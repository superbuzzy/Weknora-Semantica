package httpapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cobraknowledge.local/cobra-knowledge/internal/graphview"
	"cobraknowledge.local/cobra-knowledge/internal/model"
	"cobraknowledge.local/cobra-knowledge/internal/ontology"
)

type AccessChecker interface {
	CheckKnowledgeBaseAccess(ctx context.Context, knowledgeBaseID string, headers http.Header) error
}

type Server struct {
	EntitySource       graphview.EntitySource
	OntologySource     graphview.OntologySource
	AccessChecker      AccessChecker
	Registry           ontology.Registry
	RegistryAdminToken string
	AllowedOrigin      string
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	// Stable WeKnora-facing API. v0.6 keeps this route unchanged.
	mux.HandleFunc("GET /api/v1/knowledge-bases/{kbID}/graph", s.graph)

	// CobraKnowledge governance API. These routes are not called by the WeKnora overlay.
	mux.HandleFunc("GET /api/v1/registry/ontologies", s.registryListOntologies)
	mux.HandleFunc("GET /api/v1/registry/ontologies/{ontologyID}", s.registryGetManifest)
	mux.HandleFunc("GET /api/v1/registry/ontologies/{ontologyID}/versions/{version}", s.registryGetVersion)
	mux.HandleFunc("POST /api/v1/registry/ontologies/versions", s.registryRegisterVersion)
	mux.HandleFunc("POST /api/v1/registry/ontologies/{ontologyID}/versions/{version}/publish", s.registryPublish)
	mux.HandleFunc("POST /api/v1/registry/ontologies/{ontologyID}/activate", s.registryActivate)
	mux.HandleFunc("GET /api/v1/registry/knowledge-bases/{kbID}/binding", s.registryGetBinding)
	mux.HandleFunc("PUT /api/v1/registry/knowledge-bases/{kbID}/binding", s.registryPutBinding)
	mux.HandleFunc("GET /api/v1/registry/audit", s.registryAudit)
	return s.withCORS(mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "ok", "version": "0.6.0"})
}

func (s *Server) graph(w http.ResponseWriter, r *http.Request) {
	kbID := strings.TrimSpace(r.PathValue("kbID"))
	if kbID == "" {
		writeError(w, http.StatusBadRequest, "knowledge base id is required")
		return
	}
	if s.AccessChecker != nil {
		if err := s.AccessChecker.CheckKnowledgeBaseAccess(r.Context(), kbID, r.Header); err != nil {
			writeError(w, http.StatusForbidden, "knowledge base access denied")
			return
		}
	}

	view := strings.TrimSpace(r.URL.Query().Get("view"))
	if view == "" {
		view = "entity"
	}
	limit := 160
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()
	var graph model.GraphView
	var err error
	switch view {
	case "entity":
		if s.EntitySource == nil {
			err = errors.New("entity graph source is not configured")
		} else {
			graph, err = s.EntitySource.EntityView(ctx, kbID, limit)
		}
	case "ontology":
		if s.OntologySource == nil {
			err = errors.New("ontology graph source is not configured")
		} else {
			graph, err = s.OntologySource.OntologyView(ctx, kbID)
		}
	default:
		writeError(w, http.StatusBadRequest, "view must be entity or ontology")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, graph)
}

func (s *Server) registryListOntologies(w http.ResponseWriter, r *http.Request) {
	if !s.requireRegistryAdmin(w, r) {
		return
	}
	items, err := s.Registry.ListManifests(r.Context())
	if err != nil {
		writeRegistryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func (s *Server) registryGetManifest(w http.ResponseWriter, r *http.Request) {
	if !s.requireRegistryAdmin(w, r) {
		return
	}
	manifest, err := s.Registry.GetManifest(r.Context(), r.PathValue("ontologyID"))
	if err != nil {
		writeRegistryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, manifest)
}

func (s *Server) registryGetVersion(w http.ResponseWriter, r *http.Request) {
	if !s.requireRegistryAdmin(w, r) {
		return
	}
	o, meta, err := s.Registry.Get(r.Context(), r.PathValue("ontologyID"), r.PathValue("version"))
	if err != nil {
		writeRegistryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"metadata": meta, "ontology": o})
}

type registerVersionRequest struct {
	Ontology model.Ontology `json:"ontology"`
	Actor    string         `json:"actor,omitempty"`
	Notes    string         `json:"notes,omitempty"`
}

func (s *Server) registryRegisterVersion(w http.ResponseWriter, r *http.Request) {
	if !s.requireRegistryAdmin(w, r) {
		return
	}
	var body registerVersionRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	actor := actorFrom(r, body.Actor)
	meta, err := s.Registry.RegisterVersion(r.Context(), body.Ontology, actor, body.Notes)
	if err != nil {
		writeRegistryError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, meta)
}

func (s *Server) registryPublish(w http.ResponseWriter, r *http.Request) {
	if !s.requireRegistryAdmin(w, r) {
		return
	}
	meta, err := s.Registry.Publish(r.Context(), r.PathValue("ontologyID"), r.PathValue("version"), actorFrom(r, ""))
	if err != nil {
		writeRegistryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, meta)
}

type activateRequest struct {
	Version string `json:"version"`
	Actor   string `json:"actor,omitempty"`
}

func (s *Server) registryActivate(w http.ResponseWriter, r *http.Request) {
	if !s.requireRegistryAdmin(w, r) {
		return
	}
	var body activateRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(body.Version) == "" {
		writeError(w, http.StatusBadRequest, "version is required")
		return
	}
	if err := s.Registry.Activate(r.Context(), r.PathValue("ontologyID"), body.Version, actorFrom(r, body.Actor)); err != nil {
		writeRegistryError(w, err)
		return
	}
	manifest, _ := s.Registry.GetManifest(r.Context(), r.PathValue("ontologyID"))
	writeJSON(w, http.StatusOK, manifest)
}

func (s *Server) registryGetBinding(w http.ResponseWriter, r *http.Request) {
	if !s.requireRegistryAdmin(w, r) {
		return
	}
	binding, err := s.Registry.GetBinding(r.Context(), r.PathValue("kbID"))
	if err != nil {
		writeRegistryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, binding)
}

func (s *Server) registryPutBinding(w http.ResponseWriter, r *http.Request) {
	if !s.requireRegistryAdmin(w, r) {
		return
	}
	var binding model.OntologyBinding
	if err := decodeJSON(r, &binding); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	binding.KnowledgeBaseID = r.PathValue("kbID")
	binding.UpdatedBy = actorFrom(r, binding.UpdatedBy)
	if err := s.Registry.BindKnowledgeBase(r.Context(), binding); err != nil {
		writeRegistryError(w, err)
		return
	}
	binding, _ = s.Registry.GetBinding(r.Context(), binding.KnowledgeBaseID)
	writeJSON(w, http.StatusOK, binding)
}

func (s *Server) registryAudit(w http.ResponseWriter, r *http.Request) {
	if !s.requireRegistryAdmin(w, r) {
		return
	}
	limit := 100
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 500 {
		limit = 500
	}
	events, err := s.Registry.ListAudit(r.Context(), limit)
	if err != nil {
		writeRegistryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": events})
}

func (s *Server) requireRegistryAdmin(w http.ResponseWriter, r *http.Request) bool {
	if s.Registry == nil {
		writeError(w, http.StatusServiceUnavailable, "ontology registry is not configured")
		return false
	}
	expected := strings.TrimSpace(s.RegistryAdminToken)
	if expected == "" {
		writeError(w, http.StatusServiceUnavailable, "registry administration API is disabled")
		return false
	}
	actual := strings.TrimSpace(r.Header.Get("X-Cobra-Admin-Token"))
	if len(actual) != len(expected) || subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) != 1 {
		writeError(w, http.StatusUnauthorized, "registry administrator authentication required")
		return false
	}
	return true
}

func actorFrom(r *http.Request, fallback string) string {
	if actor := strings.TrimSpace(r.Header.Get("X-Cobra-Actor")); actor != "" {
		return actor
	}
	return strings.TrimSpace(fallback)
}

func decodeJSON(r *http.Request, out interface{}) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 4<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("invalid json body: %w", err)
	}
	return nil
}

func writeRegistryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ontology.ErrOntologyNotFound), errors.Is(err, ontology.ErrVersionNotFound), errors.Is(err, ontology.ErrBindingNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case strings.Contains(err.Error(), "already exists"):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusBadRequest, err.Error())
	}
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(s.AllowedOrigin)
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Tenant-ID, X-External-User-ID, X-Cobra-Admin-Token, X-Cobra-Actor")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		_ = fmt.Errorf("encode response: %w", err)
	}
}
