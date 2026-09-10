package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"cobraknowledge.local/cobra-knowledge/internal/model"
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
