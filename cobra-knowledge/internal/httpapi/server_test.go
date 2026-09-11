package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cobraknowledge.local/cobra-knowledge/internal/model"
	"cobraknowledge.local/cobra-knowledge/internal/ontology"
	"cobraknowledge.local/cobra-knowledge/internal/runtimecontext"
)

type fakeEntity struct{}

func (fakeEntity) EntityView(_ context.Context, kb string, _ int) (model.GraphView, error) {
	return model.GraphView{Nodes: []model.GraphViewNode{{ID: "e1", Label: "实体", Kind: "entity"}}, Meta: model.GraphViewMeta{View: "entity", KnowledgeBaseID: kb, TotalNodes: 1, ReturnedNodes: 1}}, nil
}

type fakeOntology struct{}

func (fakeOntology) OntologyView(_ context.Context, kb string) (model.GraphView, error) {
	return model.GraphView{Nodes: []model.GraphViewNode{{ID: "c1", Label: "线路", Kind: "class"}}, Meta: model.GraphViewMeta{View: "ontology", KnowledgeBaseID: kb, TotalNodes: 1, ReturnedNodes: 1}}, nil
}

func TestGraphSwitchAPI(t *testing.T) {
	s := &Server{EntitySource: fakeEntity{}, OntologySource: fakeOntology{}}
	for _, tc := range []struct{ view, want string }{{"entity", "实体"}, {"ontology", "线路"}} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/knowledge-bases/kb-1/graph?view="+tc.view, nil)
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s returned %d: %s", tc.view, rec.Code, rec.Body.String())
		}
		if body := rec.Body.String(); !contains(body, tc.want) {
			t.Fatalf("%s missing %q in %s", tc.view, tc.want, body)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestRegistryGovernanceAPI(t *testing.T) {
	root := t.TempDir()
	registry := ontology.NewFSRegistry(root)
	s := &Server{Registry: registry, RegistryAdminToken: "secret"}

	o := model.Ontology{
		ID: "ont-grid", Domain: "distribution_network", Version: "1.0.0", Status: model.StatusApproved,
		CreatedAt:  time.Now().UTC(),
		Classes:    []model.OntologyClass{{ID: "cls-line", Label: "线路", Support: 1, Confidence: 1, Status: model.StatusApproved}},
		Properties: []model.DataProperty{}, Relations: []model.ObjectRelation{},
	}
	body, _ := json.Marshal(map[string]interface{}{"ontology": o, "actor": "tester", "notes": "v1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/registry/ontologies/versions", bytes.NewReader(body))
	req.Header.Set("X-Cobra-Admin-Token", "secret")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register returned %d: %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/registry/ontologies/ont-grid/versions/1.0.0/publish", nil)
	req.Header.Set("X-Cobra-Admin-Token", "secret")
	req.Header.Set("X-Cobra-Actor", "reviewer")
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("publish returned %d: %s", rec.Code, rec.Body.String())
	}

	bindingBody := []byte(`{"ontology_id":"ont-grid","mode":"active"}`)
	req = httptest.NewRequest(http.MethodPut, "/api/v1/registry/knowledge-bases/kb-1/binding", bytes.NewReader(bindingBody))
	req.Header.Set("X-Cobra-Admin-Token", "secret")
	req.Header.Set("X-Cobra-Actor", "operator")
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("bind returned %d: %s", rec.Code, rec.Body.String())
	}

	resolved, err := registry.ResolveForKnowledgeBase(context.Background(), "kb-1")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Ontology.Version != "1.0.0" {
		t.Fatalf("unexpected resolved version: %s", resolved.Ontology.Version)
	}
}

func TestRegistryAPIRequiresSeparateAdminToken(t *testing.T) {
	s := &Server{Registry: ontology.NewFSRegistry(t.TempDir()), RegistryAdminToken: "secret"}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/registry/ontologies", nil)
	req.Header.Set("Authorization", "Bearer normal-weknora-user-token")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRuntimeContextAPIRequiresInternalBearerAndDoesNotCORSExposeTrustedHeaders(t *testing.T) {
	s := &Server{RuntimeContext: &runtimecontext.Service{}, RuntimeToken: "runtime-secret", AllowedOrigin: "https://openclaw.example"}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runtime/context/retrieve", bytes.NewReader([]byte(`{"query":"q"}`)))
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodOptions, "/api/v1/runtime/context/retrieve", nil)
	req.Header.Set("Origin", "https://openclaw.example")
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type" {
		t.Fatalf("trusted headers exposed to browser CORS: %q", got)
	}
}
