package graphview

import (
	"context"
	"testing"
	"time"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

type fakeRegistryResolver struct{}

func (fakeRegistryResolver) ResolveForKnowledgeBase(_ context.Context, kbID string) (model.OntologyResolution, error) {
	return model.OntologyResolution{
		Ontology: model.Ontology{
			ID: "o1", Domain: "demo", Version: "1.0.0", Status: model.StatusApproved, CreatedAt: time.Now(),
			Classes: []model.OntologyClass{{ID: "line", Label: "线路", Support: 1, Confidence: 1, Status: model.StatusApproved}},
		},
		Binding: model.OntologyBinding{KnowledgeBaseID: kbID, OntologyID: "o1", Mode: "active"},
		Version: model.OntologyVersionMeta{OntologyID: "o1", Version: "1.0.0", State: model.RegistryPublished},
	}, nil
}

func TestRegistryOntologySource(t *testing.T) {
	s := NewRegistryOntologySource(fakeRegistryResolver{})
	view, err := s.OntologyView(context.Background(), "kb-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Nodes) != 1 || view.Nodes[0].Label != "线路" {
		t.Fatalf("unexpected view: %+v", view)
	}
	if view.Meta.OntologyState != "published" || view.Meta.BindingMode != "active" {
		t.Fatalf("missing registry metadata: %+v", view.Meta)
	}
}
