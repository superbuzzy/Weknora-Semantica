package context

import (
	"context"
	"testing"

	"cobraknowledge.local/cobra-knowledge/internal/model"
	"cobraknowledge.local/cobra-knowledge/internal/retrieval"
)

type fakeRetriever struct {
	source model.RetrievalSource
	result model.RetrievalResult
}

func (f fakeRetriever) Source() model.RetrievalSource { return f.source }
func (f fakeRetriever) Retrieve(context.Context, model.QueryRequest, model.RetrievalStep) (model.RetrievalResult, error) {
	return f.result, nil
}
func TestContextServicePlansRetrievesArbitratesAndAssembles(t *testing.T) {
	cat := retrieval.SemanticCatalog{Classes: []retrieval.CatalogTerm{{ID: "line", Label: "线路", Kind: "class", PreferredSources: []model.RetrievalSource{model.SourceEntityGraph}}}}
	svc := NewService(retrieval.NewPlanner(cat), retrieval.NewArbiter(retrieval.ArbitrationPolicy{DefaultDelta: .05}), NewAssembler())
	svc.Register(fakeRetriever{source: model.SourceEntityGraph, result: model.RetrievalResult{Source: model.SourceEntityGraph, Assertions: []model.Assertion{{ID: "a", SlotID: "s", SubjectID: "e", PredicateID: "p", Value: "x", Source: "entity_graph", Authority: .8, Confidence: .9}}}})
	pack, err := svc.Retrieve(context.Background(), model.QueryRequest{Query: "线路状态"})
	if err != nil {
		t.Fatal(err)
	}
	if len(pack.Facts) != 1 || !pack.Complete {
		t.Fatalf("pack=%+v", pack)
	}
}

func TestStaticEntityRetrieverFiltersTargetProperty(t *testing.T) {
	g := model.GraphSnapshot{Domain: "d", Entities: []model.Entity{{ID: "e", Name: "金牛线", TypeID: "line", BusinessType: "线路", Properties: map[string]interface{}{"转供状态": "不可转供", "电压等级": "10kV"}}}}
	r := NewStaticEntityGraphRetriever(g)
	transferID := model.StableID("prop", "d", "转供状态")
	res, err := r.Retrieve(context.Background(), model.QueryRequest{Query: "金牛线转供能力"}, model.RetrievalStep{TargetPropertyIDs: []string{transferID}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Assertions) != 1 || res.Assertions[0].PredicateID != transferID {
		t.Fatalf("unexpected assertions: %+v", res.Assertions)
	}
}
