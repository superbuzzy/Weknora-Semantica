package graphview

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

// EntitySource returns a visualization-ready slice of the WeKnora entity graph.
type EntitySource interface {
	EntityView(ctx context.Context, knowledgeBaseID string, limit int) (model.GraphView, error)
}

// OntologySource resolves the ontology bound to a WeKnora knowledge base.
type OntologySource interface {
	OntologyView(ctx context.Context, knowledgeBaseID string) (model.GraphView, error)
}

// FileOntologyBindings is deliberately external to the ontology itself. A domain ontology
// can be reused by multiple knowledge bases without copying or mutating it.
type FileOntologyBindings struct {
	BaseDir  string            `json:"-"`
	Bindings map[string]string `json:"knowledge_bases"`
	Default  string            `json:"default_ontology,omitempty"`
}

func LoadFileOntologyBindings(path string) (*FileOntologyBindings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read ontology bindings: %w", err)
	}
	var b FileOntologyBindings
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("decode ontology bindings: %w", err)
	}
	b.BaseDir = filepath.Dir(path)
	if b.Bindings == nil {
		b.Bindings = map[string]string{}
	}
	return &b, nil
}

func (b *FileOntologyBindings) pathForKB(kbID string) (string, error) {
	candidate := strings.TrimSpace(b.Bindings[kbID])
	if candidate == "" {
		candidate = strings.TrimSpace(b.Default)
	}
	if candidate == "" {
		return "", fmt.Errorf("no ontology is bound to knowledge base %q", kbID)
	}
	if filepath.IsAbs(candidate) {
		return candidate, nil
	}
	return filepath.Clean(filepath.Join(b.BaseDir, candidate)), nil
}

func (b *FileOntologyBindings) OntologyView(_ context.Context, kbID string) (model.GraphView, error) {
	path, err := b.pathForKB(kbID)
	if err != nil {
		return model.GraphView{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return model.GraphView{}, fmt.Errorf("read ontology %s: %w", path, err)
	}
	var o model.Ontology
	if err := json.Unmarshal(data, &o); err != nil {
		return model.GraphView{}, fmt.Errorf("decode ontology %s: %w", path, err)
	}
	return BuildOntologyView(kbID, o), nil
}
