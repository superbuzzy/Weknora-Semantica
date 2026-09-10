package graphview

import (
	"context"

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

// OntologyResolver is the storage-neutral registry contract required by visualization.
// It intentionally exposes no filesystem path or database implementation detail.
type OntologyResolver interface {
	ResolveForKnowledgeBase(ctx context.Context, knowledgeBaseID string) (model.OntologyResolution, error)
}

// RegistryOntologySource makes the WeKnora UI consume the active/pinned ontology
// through the registry lifecycle instead of reading a binding-to-file configuration.
type RegistryOntologySource struct {
	Registry OntologyResolver
}

func NewRegistryOntologySource(registry OntologyResolver) *RegistryOntologySource {
	return &RegistryOntologySource{Registry: registry}
}

func (s *RegistryOntologySource) OntologyView(ctx context.Context, kbID string) (model.GraphView, error) {
	resolution, err := s.Registry.ResolveForKnowledgeBase(ctx, kbID)
	if err != nil {
		return model.GraphView{}, err
	}
	view := BuildOntologyView(kbID, resolution.Ontology)
	view.Meta.OntologyState = string(resolution.Version.State)
	view.Meta.BindingMode = resolution.Binding.Mode
	return view, nil
}
