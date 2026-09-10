package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cobraknowledge.local/cobra-knowledge/internal/graphview"
	"cobraknowledge.local/cobra-knowledge/internal/model"
)

type AccessChecker interface {
	CheckKnowledgeBaseAccess(ctx context.Context, knowledgeBaseID string, headers http.Header) error
}

type Server struct {
	EntitySource   graphview.EntitySource
	OntologySource graphview.OntologySource
	AccessChecker  AccessChecker
	AllowedOrigin  string
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /api/v1/knowledge-bases/{kbID}/graph", s.graph)
	return s.withCORS(mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "ok", "version": "0.3.0"})
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

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(s.AllowedOrigin)
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Tenant-ID")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
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
