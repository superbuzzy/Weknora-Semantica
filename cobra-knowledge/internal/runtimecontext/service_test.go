package runtimecontext

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cobraknowledge.local/cobra-knowledge/internal/model"
	"cobraknowledge.local/cobra-knowledge/internal/ontology"
)

func TestRuntimeContextUsesTrustedWorkspacePrincipalForWeKnora(t *testing.T) {
	var searchHeaders http.Header
	var searchBody map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/knowledge-search":
			searchHeaders = r.Header.Clone()
			_ = json.NewDecoder(r.Body).Decode(&searchBody)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": []interface{}{map[string]interface{}{"id": "chunk-1", "content": "正式知识", "knowledge_id": "doc-1", "knowledge_title": "制度", "score": 0.93}}})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/chunks/by-id/chunk-1":
			if r.Header.Get("X-Tenant-ID") != "42" || r.Header.Get("X-External-User-ID") != "p1" {
				t.Fatalf("evidence request lost principal headers: %#v", r.Header)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": map[string]interface{}{"id": "chunk-1", "knowledge_id": "doc-1", "knowledge_base_id": "kb-1", "content": "原始证据"}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	svc := &Service{DefaultWeKnoraBaseURL: ts.URL}
	principal := Principal{WorkspaceID: "cq", UserID: "p1", TenantID: "42", WeKnoraAPIKey: "wk-secret"}
	pack, err := svc.Retrieve(context.Background(), principal, RetrieveRequest{Query: "制度是什么", KnowledgeBaseIDs: []string{"kb-1"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(pack.Knowledge) != 1 || pack.Knowledge[0].Content != "正式知识" {
		t.Fatalf("unexpected pack: %#v", pack)
	}
	if searchHeaders.Get("X-API-Key") != "wk-secret" || searchHeaders.Get("X-Tenant-ID") != "42" || searchHeaders.Get("X-External-User-ID") != "p1" {
		t.Fatalf("trusted principal not forwarded: %#v", searchHeaders)
	}
	if searchBody["knowledge_base_id"] != "kb-1" {
		t.Fatalf("workspace KB scope missing: %#v", searchBody)
	}
	if pack.Plan.Semantics.Scope["workspace_id"] != "cq" || pack.Plan.Semantics.Scope["user_id"] != "p1" {
		t.Fatalf("server scope missing from plan: %#v", pack.Plan.Semantics.Scope)
	}

	evidence, err := svc.GetEvidence(context.Background(), principal, EvidenceRequest{ChunkID: "chunk-1", KnowledgeBaseIDs: []string{"kb-1"}})
	if err != nil {
		t.Fatal(err)
	}
	if evidence.Content != "原始证据" {
		t.Fatalf("unexpected evidence: %#v", evidence)
	}
}

func TestRuntimeContextRejectsIncompletePrincipal(t *testing.T) {
	svc := &Service{DefaultWeKnoraBaseURL: "http://unused"}
	_, err := svc.Retrieve(context.Background(), Principal{WorkspaceID: "cq", UserID: "p1"}, RetrieveRequest{Query: "q"})
	if err == nil {
		t.Fatal("expected incomplete principal error")
	}
}

func TestRuntimeContextRAGFallbackDoesNotEraseStructuredFactGap(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/knowledge-search" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data": []interface{}{map[string]interface{}{
					"id": "chunk-1", "content": "制度材料中提到线路现状", "knowledge_id": "doc-1", "knowledge_title": "制度", "score": 0.9,
				}},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	registry := ontology.NewFSRegistry(t.TempDir())
	o := model.Ontology{
		ID: "ont-grid", Domain: "distribution_network", Version: "1.0.0",
		Status: model.StatusApproved, CreatedAt: time.Now().UTC(),
		Classes: []model.OntologyClass{{ID: "cls-line", Label: "线路", Support: 1, Confidence: 1, Status: model.StatusApproved}},
	}
	ctx := context.Background()
	if _, err := registry.RegisterVersion(ctx, o, "test", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Publish(ctx, o.ID, o.Version, "test"); err != nil {
		t.Fatal(err)
	}
	if err := registry.BindKnowledgeBase(ctx, model.OntologyBinding{KnowledgeBaseID: "kb-1", OntologyID: o.ID, Mode: "active"}); err != nil {
		t.Fatal(err)
	}

	svc := &Service{DefaultWeKnoraBaseURL: ts.URL, Registry: registry}
	p := Principal{WorkspaceID: "cq", UserID: "p1", TenantID: "42", WeKnoraAPIKey: "wk"}
	pack, err := svc.Retrieve(ctx, p, RetrieveRequest{Query: "线路", KnowledgeBaseIDs: []string{"kb-1"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(pack.Knowledge) != 1 {
		t.Fatalf("expected documentary fallback, got %#v", pack)
	}
	if pack.Complete {
		t.Fatalf("documentary fallback must not claim a missing structured fact is complete: %#v", pack)
	}
	if len(pack.Gaps) == 0 {
		t.Fatalf("structured retrieval gap must be preserved: %#v", pack)
	}
}

func TestRuntimeContextEvidenceRejectsChunkOutsideWorkspaceKnowledgeScope(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/chunks/by-id/chunk-outside" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"id": "chunk-outside", "knowledge_id": "doc-2", "knowledge_base_id": "kb-other", "content": "不属于当前 Workspace 的证据",
				},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	svc := &Service{DefaultWeKnoraBaseURL: ts.URL}
	principal := Principal{WorkspaceID: "cq", UserID: "p1", TenantID: "42", WeKnoraAPIKey: "wk-secret"}
	_, err := svc.GetEvidence(context.Background(), principal, EvidenceRequest{ChunkID: "chunk-outside", KnowledgeBaseIDs: []string{"kb-1", "kb-2"}})
	if !errors.Is(err, ErrEvidenceOutsideScope) {
		t.Fatalf("expected ErrEvidenceOutsideScope, got %v", err)
	}
}
