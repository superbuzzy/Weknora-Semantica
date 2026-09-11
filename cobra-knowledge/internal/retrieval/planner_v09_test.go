package retrieval

import (
	"testing"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

func TestPlannerV09MarksLiveBusinessDataRequired(t *testing.T) {
	cat := SemanticCatalog{Properties: []CatalogTerm{{ID: "load", Label: "负载率", Kind: "property", PreferredSources: []model.RetrievalSource{model.SourceBusinessData}}}}
	plan := NewPlanner(cat).Plan(model.QueryRequest{Query: "线路当前负载率是多少", Task: "问数"})
	if len(plan.Steps) != 1 || plan.Steps[0].Source != model.SourceBusinessData || !plan.Steps[0].Required {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	if plan.FreshnessRequirement != "live" || len(plan.RequiredSources) != 1 {
		t.Fatalf("live requirement missing: %+v", plan)
	}
}

func TestFederatedCatalogKeepsConceptSenseConflictVisible(t *testing.T) {
	left := SemanticCatalog{Classes: []CatalogTerm{{ID: "line-a", Label: "线路", Kind: "class"}}}
	right := SemanticCatalog{Classes: []CatalogTerm{{ID: "line-b", Label: "线路", Kind: "class"}}}
	federated := FederateCatalogs([]NamespacedCatalog{{Namespace: "kb-a", Catalog: left}, {Namespace: "kb-b", Catalog: right}})
	if len(federated.Conflicts) != 1 {
		t.Fatalf("expected visible federation conflict, got %+v", federated)
	}
	plan := NewPlanner(federated).Plan(model.QueryRequest{Query: "线路现状"})
	if len(plan.BlockingIssues) != 1 {
		t.Fatalf("expected relevant conflict to block completeness: %+v", plan)
	}
}
