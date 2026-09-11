package context

import (
	"testing"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

func TestAssemblerV09OnlyRequiredGapsBlockCompleteness(t *testing.T) {
	plan := model.RetrievalPlan{Query: "q"}
	results := []model.RetrievalResult{
		{StepID: "s1", Source: model.SourceEntityGraph, Required: true, Satisfied: true, Assertions: []model.Assertion{{ID: "a", SlotID: "s", SubjectID: "e", PredicateID: "p", Value: "ok", Source: "entity_graph", Authority: .8, Confidence: .9}}},
		{StepID: "s2", Source: model.SourceWikiRAG, Required: false, Satisfied: false, Gaps: []string{"supporting source unavailable"}},
	}
	arbitration := []model.ArbitrationResult{{SlotID: "s", Status: "accepted", Accepted: &results[0].Assertions[0]}}
	pack := NewAssembler().Assemble(plan, results, arbitration)
	if !pack.Complete {
		t.Fatalf("supporting source gap must not block a complete required fact: %+v", pack)
	}
	if len(pack.Gaps) != 1 || len(pack.BlockingGaps) != 0 {
		t.Fatalf("expected diagnostic non-blocking gap: %+v", pack)
	}
}

func TestAssemblerV09MissingRequiredSourceBlocksCompleteness(t *testing.T) {
	plan := model.RetrievalPlan{Query: "q"}
	results := []model.RetrievalResult{{StepID: "s1", Source: model.SourceBusinessData, Required: true, Satisfied: false, Gaps: []string{"gateway unavailable"}}}
	pack := NewAssembler().Assemble(plan, results, nil)
	if pack.Complete || len(pack.BlockingGaps) == 0 {
		t.Fatalf("missing required source must fail completeness: %+v", pack)
	}
}
