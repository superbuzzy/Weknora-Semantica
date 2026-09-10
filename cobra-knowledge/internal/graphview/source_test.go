package graphview

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileOntologyBindings(t *testing.T) {
	dir := t.TempDir()
	ontology := `{"id":"o1","domain":"demo","version":"1.0","status":"approved","created_at":"2026-09-10T00:00:00Z","classes":[{"id":"line","label":"线路","support":1,"confidence":1,"status":"approved"}],"properties":[],"relations":[]}`
	if err := os.WriteFile(filepath.Join(dir, "ontology.json"), []byte(ontology), 0o644); err != nil {
		t.Fatal(err)
	}
	bindings := `{"knowledge_bases":{"kb-1":"ontology.json"}}`
	bindingsPath := filepath.Join(dir, "bindings.json")
	if err := os.WriteFile(bindingsPath, []byte(bindings), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := LoadFileOntologyBindings(bindingsPath)
	if err != nil {
		t.Fatal(err)
	}
	view, err := b.OntologyView(context.Background(), "kb-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Nodes) != 1 || view.Nodes[0].Label != "线路" {
		t.Fatalf("unexpected view: %+v", view)
	}
}
