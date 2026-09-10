package graphview

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNeo4jHTTPSourceEntityView(t *testing.T) {
	call := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		defer r.Body.Close()
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		statements := req["statements"].([]interface{})
		statement := statements[0].(map[string]interface{})["statement"].(string)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(statement, "RETURN coalesce(n.kg"):
			_, _ = w.Write([]byte(`{"results":[{"columns":["kg","name","attributes","chunks"],"data":[{"row":["doc-1","金牛线",["转供状态=不可转供"],["chunk-1"]]},{"row":["doc-1","金牛台区",[],["chunk-2"]]}]}],"errors":[]}`))
		case strings.Contains(statement, "type(r) AS rel_type"):
			_, _ = w.Write([]byte(`{"results":[{"columns":["source_id","target_id","rel_type"],"data":[{"row":["doc-1::金牛线","doc-1::金牛台区","供电"]}]}],"errors":[]}`))
		case strings.Contains(statement, "count(n) AS total"):
			_, _ = w.Write([]byte(`{"results":[{"columns":["total"],"data":[{"row":[2]}]}],"errors":[]}`))
		default:
			t.Fatalf("unexpected statement: %s", statement)
		}
	}))
	defer server.Close()

	source := NewNeo4jHTTPSource(server.URL, "neo4j", "", "")
	view, err := source.EntityView(context.Background(), "kb-123", 100)
	if err != nil {
		t.Fatal(err)
	}
	if call != 3 {
		t.Fatalf("expected 3 calls, got %d", call)
	}
	if len(view.Nodes) != 2 || len(view.Edges) != 1 {
		t.Fatalf("unexpected graph view: %+v", view)
	}
	if view.Meta.TotalNodes != 2 || view.Meta.Truncated {
		t.Fatalf("unexpected meta: %+v", view.Meta)
	}
	if view.Edges[0].Label != "供电" {
		t.Fatalf("unexpected relation: %+v", view.Edges[0])
	}
}

func TestWeknoraKBLabelRejectsUnsafeID(t *testing.T) {
	if _, err := weknoraKBLabel("kb`) MATCH (n) DETACH DELETE n //"); err == nil {
		t.Fatal("expected unsafe KB id to be rejected")
	}
	label, err := weknoraKBLabel("abc-def_123")
	if err != nil {
		t.Fatal(err)
	}
	if label != "ENTITYabc_def_123" {
		t.Fatalf("unexpected label: %s", label)
	}
}
