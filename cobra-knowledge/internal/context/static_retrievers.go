package context

import (
	"context"
	"fmt"
	"strings"

	wk "cobraknowledge.local/cobra-knowledge/internal/adapters/weknora"
	"cobraknowledge.local/cobra-knowledge/internal/graph"
	"cobraknowledge.local/cobra-knowledge/internal/model"
)

type StaticEntityGraphRetriever struct {
	Graph      model.GraphSnapshot
	Assertions []model.Assertion
}

func NewStaticEntityGraphRetriever(g model.GraphSnapshot) *StaticEntityGraphRetriever {
	return &StaticEntityGraphRetriever{Graph: g, Assertions: graph.NewAssertionBuilder("entity_graph").FromGraph(g)}
}
func (r *StaticEntityGraphRetriever) Source() model.RetrievalSource { return model.SourceEntityGraph }
func (r *StaticEntityGraphRetriever) Retrieve(_ context.Context, req model.QueryRequest, step model.RetrievalStep) (model.RetrievalResult, error) {
	q := strings.ToLower(req.Query)
	entities := map[string]bool{}
	classTargets := setOf(step.TargetClassIDs)
	for _, e := range r.Graph.Entities {
		nameMatch := strings.Contains(q, strings.ToLower(e.Name))
		for _, a := range e.Aliases {
			if strings.Contains(q, strings.ToLower(a)) {
				nameMatch = true
				break
			}
		}
		typeMatch := len(classTargets) > 0 && classTargets[e.TypeID]
		if nameMatch || typeMatch {
			entities[e.ID] = true
		}
	}
	propTargets := setOf(step.TargetPropertyIDs)
	relTargets := setOf(step.TargetRelationIDs)
	var assertions []model.Assertion
	for _, a := range r.Assertions {
		if !entities[a.SubjectID] {
			continue
		}
		if len(propTargets) > 0 || len(relTargets) > 0 {
			if !propTargets[a.PredicateID] && !relTargets[a.PredicateID] {
				continue
			}
		}
		assertions = append(assertions, a)
	}
	var paths []model.RelationPath
	for _, rel := range r.Graph.Relations {
		if !entities[rel.SourceID] && !entities[rel.TargetID] {
			continue
		}
		if len(relTargets) > 0 && !relTargets[rel.TypeID] {
			continue
		}
		if len(propTargets) > 0 && len(relTargets) == 0 {
			continue
		}
		paths = append(paths, model.RelationPath{Nodes: []string{rel.SourceID, rel.TargetID}, Relations: []string{rel.TypeID}, Evidence: rel.Evidence})
	}
	gaps := []string{}
	if len(assertions) == 0 && len(paths) == 0 {
		gaps = append(gaps, "实体图未命中满足目标类型/属性/关系的事实")
	}
	return model.RetrievalResult{Source: model.SourceEntityGraph, Assertions: assertions, Paths: paths, Gaps: gaps}, nil
}

func setOf(in []string) map[string]bool {
	out := map[string]bool{}
	for _, v := range in {
		out[v] = true
	}
	return out
}

type StaticOntologyRetriever struct{ Ontology model.Ontology }

func (r *StaticOntologyRetriever) Source() model.RetrievalSource { return model.SourceOntologyGraph }
func (r *StaticOntologyRetriever) Retrieve(_ context.Context, req model.QueryRequest, _ model.RetrievalStep) (model.RetrievalResult, error) {
	q := strings.ToLower(req.Query)
	var items []model.KnowledgeItem
	for _, c := range r.Ontology.Classes {
		if matchTerm(q, c.Label, c.Aliases) {
			items = append(items, model.KnowledgeItem{ID: c.ID, Title: "本体类型: " + c.Label, Content: fmt.Sprintf("类型=%s；父类=%v；描述=%s", c.Label, c.ParentIDs, c.Description), Source: model.SourceOntologyGraph, Score: c.Confidence})
		}
	}
	for _, p := range r.Ontology.Properties {
		if matchTerm(q, p.Label, p.Aliases) {
			items = append(items, model.KnowledgeItem{ID: p.ID, Title: "本体属性: " + p.Label, Content: fmt.Sprintf("属性=%s；Domain=%v；数据类型=%s", p.Label, p.DomainIDs, p.DataType), Source: model.SourceOntologyGraph, Score: p.Confidence})
		}
	}
	for _, rel := range r.Ontology.Relations {
		if matchTerm(q, rel.Label, rel.Aliases) {
			items = append(items, model.KnowledgeItem{ID: rel.ID, Title: "本体关系: " + rel.Label, Content: fmt.Sprintf("关系=%s；Domain=%v；Range=%v", rel.Label, rel.DomainIDs, rel.RangeIDs), Source: model.SourceOntologyGraph, Score: rel.Confidence})
		}
	}
	gaps := []string{}
	if len(items) == 0 {
		gaps = append(gaps, "本体图未命中相关语义")
	}
	return model.RetrievalResult{Source: model.SourceOntologyGraph, Knowledge: items, Gaps: gaps}, nil
}
func matchTerm(q, label string, aliases []string) bool {
	if strings.Contains(q, strings.ToLower(label)) {
		return true
	}
	for _, a := range aliases {
		if strings.Contains(q, strings.ToLower(a)) {
			return true
		}
	}
	return false
}

type WeKnoraRAGRetriever struct {
	Client           *wk.SearchClient
	KnowledgeBaseIDs []string
	KnowledgeIDs     []string
}

func (r *WeKnoraRAGRetriever) Source() model.RetrievalSource { return model.SourceWikiRAG }
func (r *WeKnoraRAGRetriever) Retrieve(ctx context.Context, req model.QueryRequest, _ model.RetrievalStep) (model.RetrievalResult, error) {
	items, err := r.Client.Search(ctx, req.Query, r.KnowledgeBaseIDs, r.KnowledgeIDs)
	return model.RetrievalResult{Source: model.SourceWikiRAG, Knowledge: items}, err
}
