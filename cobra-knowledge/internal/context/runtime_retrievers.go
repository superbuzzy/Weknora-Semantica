package context

import (
	"context"
	"fmt"
	"strings"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

// EntityQuerySource is the narrow runtime contract implemented by the WeKnora
// Neo4j adapter. It is intentionally independent from visualization storage details.
type EntityQuerySource interface {
	EntityQuery(ctx context.Context, knowledgeBaseID string, terms []string, limit int) (model.GraphView, error)
}

type EntityGraphRuntimeRetriever struct {
	SourceAdapter    EntityQuerySource
	KnowledgeBaseIDs []string
	Limit            int
}

func NewEntityGraphRuntimeRetriever(source EntityQuerySource, knowledgeBaseIDs []string) *EntityGraphRuntimeRetriever {
	return &EntityGraphRuntimeRetriever{SourceAdapter: source, KnowledgeBaseIDs: append([]string(nil), knowledgeBaseIDs...), Limit: 200}
}

func (r *EntityGraphRuntimeRetriever) Source() model.RetrievalSource { return model.SourceEntityGraph }

func (r *EntityGraphRuntimeRetriever) Retrieve(ctx context.Context, req model.QueryRequest, step model.RetrievalStep) (model.RetrievalResult, error) {
	if r.SourceAdapter == nil {
		return model.RetrievalResult{Source: model.SourceEntityGraph}, fmt.Errorf("entity graph source is not configured")
	}
	if len(r.KnowledgeBaseIDs) == 0 {
		return model.RetrievalResult{Source: model.SourceEntityGraph, Gaps: []string{"entity graph requires explicit workspace knowledge-base scope"}}, nil
	}
	terms := runtimeTerms(req.Query, step.TargetTerms)
	limit := r.Limit
	if limit <= 0 {
		limit = 200
	}
	result := model.RetrievalResult{Source: model.SourceEntityGraph}
	for _, kbID := range r.KnowledgeBaseIDs {
		view, err := r.SourceAdapter.EntityQuery(ctx, kbID, terms, limit)
		if err != nil {
			result.Gaps = append(result.Gaps, fmt.Sprintf("entity graph %s: %v", kbID, err))
			continue
		}
		partial := graphViewToResult(view, req, step)
		result.Assertions = append(result.Assertions, partial.Assertions...)
		result.Paths = append(result.Paths, partial.Paths...)
		result.Gaps = append(result.Gaps, partial.Gaps...)
	}
	if len(result.Assertions) == 0 && len(result.Paths) == 0 && len(result.Gaps) == 0 {
		result.Gaps = append(result.Gaps, "entity graph returned no matching facts or relations")
	}
	return result, nil
}

func graphViewToResult(view model.GraphView, req model.QueryRequest, step model.RetrievalStep) model.RetrievalResult {
	result := model.RetrievalResult{Source: model.SourceEntityGraph}
	propertyTargets := stringSet(step.TargetPropertyIDs)
	targetTerms := lowerSet(step.TargetTerms)
	nodeEvidence := map[string][]model.EvidenceRef{}
	selected := map[string]bool{}

	for _, node := range view.Nodes {
		attrs := stringSlice(node.Metadata["attributes"])
		props, reserved := parseGraphAttributes(attrs)
		evidence := evidenceFromGraphNode(view.Meta.KnowledgeBaseID, node)
		nodeEvidence[node.ID] = evidence
		if nodeMatches(node, attrs, req.Query, targetTerms) {
			selected[node.ID] = true
		}
		typeLabel := strings.TrimSpace(reserved["__type__"])
		if typeLabel != "" && targetTerms[strings.ToLower(typeLabel)] {
			selected[node.ID] = true
		}
		if !selected[node.ID] {
			continue
		}
		for key, value := range props {
			predicateID := model.StableID("prop", req.Domain, key)
			if len(propertyTargets) > 0 && !propertyTargets[predicateID] && !targetTerms[strings.ToLower(key)] {
				continue
			}
			result.Assertions = append(result.Assertions, model.Assertion{
				ID:          model.StableID("assert", "entity_graph", node.ID+"\x00"+key+"\x00"+fmt.Sprint(value)),
				SlotID:      node.ID + "\x00" + predicateID,
				SubjectID:   node.ID,
				PredicateID: predicateID,
				Value:       value,
				Scope:       cloneMap(req.Scope),
				Source:      "entity_graph",
				SourceType:  string(model.SourceEntityGraph),
				Authority:   0.8,
				Confidence:  0.8,
				Evidence:    evidence,
			})
		}
	}

	relationTargets := stringSet(step.TargetRelationIDs)
	for _, edge := range view.Edges {
		if !selected[edge.Source] && !selected[edge.Target] {
			continue
		}
		relationID := model.StableID("obj", req.Domain, edge.Label)
		if len(relationTargets) > 0 && !relationTargets[relationID] && !targetTerms[strings.ToLower(edge.Label)] {
			continue
		}
		evidence := mergeEvidence(nodeEvidence[edge.Source], nodeEvidence[edge.Target])
		result.Paths = append(result.Paths, model.RelationPath{Nodes: []string{edge.Source, edge.Target}, Relations: []string{relationID}, Evidence: evidence})
	}
	if len(result.Assertions) == 0 && len(result.Paths) == 0 && len(view.Nodes) > 0 {
		result.Gaps = append(result.Gaps, "entity graph matched nodes but no requested property or relation was available")
	}
	return result
}

func runtimeTerms(query string, targets []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range targets {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	if len(out) == 0 {
		value := strings.ToLower(strings.TrimSpace(query))
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func nodeMatches(node model.GraphViewNode, attrs []string, query string, targets map[string]bool) bool {
	q := strings.ToLower(query)
	if strings.TrimSpace(node.Label) != "" && strings.Contains(q, strings.ToLower(node.Label)) {
		return true
	}
	text := strings.ToLower(node.Label + " " + node.Subtitle + " " + strings.Join(attrs, " "))
	for term := range targets {
		if term != "" && strings.Contains(text, term) {
			return true
		}
	}
	return false
}

func evidenceFromGraphNode(kbID string, node model.GraphViewNode) []model.EvidenceRef {
	chunks := stringSlice(node.Metadata["chunks"])
	knowledgeID, _ := node.Metadata["knowledge_id"].(string)
	out := make([]model.EvidenceRef, 0, len(chunks))
	for _, chunkID := range chunks {
		out = append(out, model.EvidenceRef{
			ID:              model.StableID("ev", "entity_graph", kbID+"\x00"+knowledgeID+"\x00"+chunkID),
			Source:          "weknora",
			SourceType:      string(model.SourceEntityGraph),
			SourceID:        node.ID,
			KnowledgeBaseID: kbID,
			KnowledgeID:     knowledgeID,
			EntityID:        node.ID,
			ChunkID:         chunkID,
			Confidence:      0.8,
			Quality:         model.EvidenceExact,
			Provenance:      map[string]interface{}{"adapter": "weknora-neo4j", "entity": node.Label},
		})
	}
	return out
}

func parseGraphAttributes(attrs []string) (map[string]interface{}, map[string]string) {
	props := map[string]interface{}{}
	reserved := map[string]string{}
	for _, raw := range attrs {
		parts := strings.SplitN(strings.TrimSpace(raw), "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if strings.HasPrefix(key, "__") {
			reserved[key] = value
			continue
		}
		if key != "" {
			props[key] = value
		}
	}
	return props, reserved
}

func stringSlice(value interface{}) []string {
	switch items := value.(type) {
	case []string:
		return append([]string(nil), items...)
	case []interface{}:
		out := make([]string, 0, len(items))
		for _, item := range items {
			if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
				out = append(out, strings.TrimSpace(text))
			}
		}
		return out
	default:
		return nil
	}
}

func stringSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
}

func lowerSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "" {
			out[value] = true
		}
	}
	return out
}

func cloneMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := map[string]string{}
	for key, value := range in {
		out[key] = value
	}
	return out
}

func mergeEvidence(groups ...[]model.EvidenceRef) []model.EvidenceRef {
	seen := map[string]bool{}
	out := []model.EvidenceRef{}
	for _, group := range groups {
		for _, evidence := range group {
			if evidence.ID != "" && !seen[evidence.ID] {
				seen[evidence.ID] = true
				out = append(out, evidence)
			}
		}
	}
	return out
}

type NamespacedOntology struct {
	Namespace string
	Ontology  model.Ontology
}

type FederatedOntologyRetriever struct{ Ontologies []NamespacedOntology }

func (r *FederatedOntologyRetriever) Source() model.RetrievalSource { return model.SourceOntologyGraph }

func (r *FederatedOntologyRetriever) Retrieve(ctx context.Context, req model.QueryRequest, step model.RetrievalStep) (model.RetrievalResult, error) {
	result := model.RetrievalResult{Source: model.SourceOntologyGraph}
	for _, item := range r.Ontologies {
		partial, err := (&StaticOntologyRetriever{Ontology: item.Ontology}).Retrieve(ctx, req, step)
		if err != nil {
			return result, err
		}
		for _, knowledge := range partial.Knowledge {
			knowledge.ID = item.Namespace + "::" + knowledge.ID
			knowledge.Title = "[" + item.Namespace + "] " + knowledge.Title
			result.Knowledge = append(result.Knowledge, knowledge)
		}
	}
	if len(result.Knowledge) == 0 {
		result.Gaps = append(result.Gaps, "federated ontology did not match the requested semantics")
	}
	return result, nil
}
