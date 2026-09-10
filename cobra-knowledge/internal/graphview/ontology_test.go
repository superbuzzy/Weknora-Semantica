package graphview

import (
	"testing"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

func TestBuildOntologyView(t *testing.T) {
	o := model.Ontology{
		ID: "dist", Version: "1.0.0", Status: model.StatusApproved,
		Classes: []model.OntologyClass{
			{ID: "resource", Label: "电网资源", Status: model.StatusApproved},
			{ID: "line", Label: "线路", ParentIDs: []string{"resource"}, Status: model.StatusApproved},
			{ID: "area", Label: "台区", ParentIDs: []string{"resource"}, Status: model.StatusApproved},
		},
		Properties: []model.DataProperty{
			{ID: "transfer_status", Label: "转供状态", DomainIDs: []string{"line"}, DataType: "string", Status: model.StatusApproved},
		},
		Relations: []model.ObjectRelation{
			{ID: "supplies", Label: "供电", DomainIDs: []string{"line"}, RangeIDs: []string{"area"}, Status: model.StatusApproved},
		},
	}

	view := BuildOntologyView("kb-1", o)
	if view.Meta.View != "ontology" || view.Meta.OntologyVersion != "1.0.0" {
		t.Fatalf("unexpected meta: %+v", view.Meta)
	}
	if len(view.Nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(view.Nodes))
	}
	if len(view.Edges) != 4 { // line->resource, area->resource, line->property, line->area
		t.Fatalf("expected 4 edges, got %d: %+v", len(view.Edges), view.Edges)
	}
}
