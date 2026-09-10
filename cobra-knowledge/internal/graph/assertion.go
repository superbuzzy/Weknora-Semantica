package graph

import (
	"sort"
	"strings"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

type AssertionBuilder struct {
	Source     string
	SourceType string
	Authority  float64
	Confidence float64
}

func NewAssertionBuilder(source string) *AssertionBuilder {
	return &AssertionBuilder{Source: source, SourceType: "knowledge_graph", Authority: 0.5, Confidence: 0.8}
}

func (b *AssertionBuilder) FromGraph(graph model.GraphSnapshot) []model.Assertion {
	var out []model.Assertion
	for _, entity := range graph.Entities {
		keys := make([]string, 0, len(entity.Properties))
		for key := range entity.Properties {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			predicateID := model.StableID("prop", graph.Domain, key)
			slot := SlotID(entity.ID, predicateID, nil)
			out = append(out, model.Assertion{
				ID:     model.StableID("assert", graph.Domain, entity.ID+"\x00"+predicateID+"\x00"+normalizedValue(entity.Properties[key])),
				SlotID: slot, SubjectID: entity.ID, PredicateID: predicateID, Value: entity.Properties[key],
				Source: b.Source, SourceType: b.SourceType, Authority: b.Authority, Confidence: b.Confidence,
				Evidence: append([]model.EvidenceRef(nil), entity.Evidence...),
			})
		}
	}
	for _, rel := range graph.Relations {
		slot := SlotID(rel.SourceID, rel.TypeID, map[string]string{"target": rel.TargetID})
		out = append(out, model.Assertion{
			ID:     model.StableID("assertrel", graph.Domain, rel.SourceID+"\x00"+rel.TypeID+"\x00"+rel.TargetID),
			SlotID: slot, SubjectID: rel.SourceID, PredicateID: rel.TypeID, Value: rel.TargetID,
			Source: b.Source, SourceType: "relationship", Authority: b.Authority, Confidence: b.Confidence,
			Evidence: append([]model.EvidenceRef(nil), rel.Evidence...),
		})
	}
	return out
}

func SlotID(subjectID, predicateID string, scope map[string]string) string {
	keys := make([]string, 0, len(scope))
	for k := range scope {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString(subjectID)
	b.WriteByte('\x00')
	b.WriteString(predicateID)
	for _, k := range keys {
		b.WriteByte('\x00')
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(scope[k])
	}
	return model.StableID("slot", "assertion", b.String())
}
