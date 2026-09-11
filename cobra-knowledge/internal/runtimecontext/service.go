package runtimecontext

import (
	"context"
	"errors"
	"fmt"
	"strings"

	wk "cobraknowledge.local/cobra-knowledge/internal/adapters/weknora"
	ctxsvc "cobraknowledge.local/cobra-knowledge/internal/context"
	"cobraknowledge.local/cobra-knowledge/internal/model"
	"cobraknowledge.local/cobra-knowledge/internal/ontology"
	"cobraknowledge.local/cobra-knowledge/internal/retrieval"
)

type Principal struct {
	WorkspaceID    string
	UserID         string
	TenantID       string
	WeKnoraAPIKey  string
	WeKnoraBaseURL string
}

type RetrieveRequest struct {
	Query            string   `json:"query"`
	Domain           string   `json:"domain,omitempty"`
	Task             string   `json:"task,omitempty"`
	KnowledgeBaseIDs []string `json:"knowledge_base_ids,omitempty"`
}

type EvidenceRequest struct {
	ChunkID          string   `json:"chunk_id"`
	KnowledgeBaseIDs []string `json:"knowledge_base_ids,omitempty"`
}

var ErrEvidenceOutsideScope = errors.New("evidence chunk is outside the active workspace knowledge scope")

type Service struct {
	DefaultWeKnoraBaseURL string
	Registry              ontology.Registry
	Overlay               retrieval.CatalogOverlay
}

func (s *Service) weknoraClient(p Principal) (*wk.SearchClient, error) {
	base := strings.TrimSpace(p.WeKnoraBaseURL)
	if base == "" {
		base = strings.TrimSpace(s.DefaultWeKnoraBaseURL)
	}
	if base == "" || strings.TrimSpace(p.WeKnoraAPIKey) == "" || strings.TrimSpace(p.TenantID) == "" || strings.TrimSpace(p.UserID) == "" {
		return nil, fmt.Errorf("runtime context principal is incomplete")
	}
	return wk.NewSearchClientWithPrincipal(base, p.WeKnoraAPIKey, p.TenantID, p.UserID), nil
}

func (s *Service) ontologyFor(ctx context.Context, ids []string) (model.Ontology, bool, error) {
	if s.Registry == nil || len(ids) != 1 || strings.TrimSpace(ids[0]) == "" {
		return model.Ontology{}, false, nil
	}
	resolved, err := s.Registry.ResolveForKnowledgeBase(ctx, strings.TrimSpace(ids[0]))
	if err != nil {
		if errors.Is(err, ontology.ErrBindingNotFound) || errors.Is(err, ontology.ErrOntologyNotFound) || errors.Is(err, ontology.ErrVersionNotFound) {
			return model.Ontology{}, false, nil
		}
		return model.Ontology{}, false, err
	}
	return resolved.Ontology, true, nil
}

func (s *Service) Retrieve(ctx context.Context, p Principal, in RetrieveRequest) (model.ContextPack, error) {
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return model.ContextPack{}, fmt.Errorf("query is required")
	}
	client, err := s.weknoraClient(p)
	if err != nil {
		return model.ContextPack{}, err
	}
	onto, hasOntology, err := s.ontologyFor(ctx, in.KnowledgeBaseIDs)
	if err != nil {
		return model.ContextPack{}, fmt.Errorf("resolve knowledge-base ontology: %w", err)
	}
	catalog := retrieval.CompileCatalogWithOverlay(onto, s.Overlay)
	planner := retrieval.NewPlanner(catalog)
	arbiter := retrieval.NewArbiter(retrieval.ArbitrationPolicyFromOntology(onto, map[string]float64{"business_data": 1.0, "entity_graph": 0.8, "weknora": 0.6}))
	service := ctxsvc.NewService(planner, arbiter, ctxsvc.NewAssembler())
	service.Register(&ctxsvc.WeKnoraRAGRetriever{Client: client, KnowledgeBaseIDs: in.KnowledgeBaseIDs})
	if hasOntology {
		service.Register(&ctxsvc.StaticOntologyRetriever{Ontology: onto})
	}
	req := model.QueryRequest{Query: query, Domain: strings.TrimSpace(in.Domain), Task: strings.TrimSpace(in.Task), Scope: map[string]string{"workspace_id": p.WorkspaceID, "user_id": p.UserID}}
	if len(in.KnowledgeBaseIDs) == 1 {
		req.Scope["knowledge_base_id"] = in.KnowledgeBaseIDs[0]
	}
	pack, err := service.Retrieve(ctx, req)
	if err != nil {
		return model.ContextPack{}, err
	}
	// The v0.8 runtime always has an authoritative RAG fallback. Some ontology-driven
	// plans can target an entity/business retriever that is not yet configured; when
	// that produces no usable answer, retrieve WeKnora once rather than fabricating facts.
	if len(pack.Facts) == 0 && len(pack.Knowledge) == 0 && len(pack.Paths) == 0 {
		items, ragErr := client.Search(ctx, query, in.KnowledgeBaseIDs, nil)
		if ragErr == nil && len(items) > 0 {
			pack.Knowledge = items
			seen := map[string]bool{}
			for _, item := range items {
				for _, ev := range item.Evidence {
					if !seen[ev.ID] {
						seen[ev.ID] = true
						pack.Evidence = append(pack.Evidence, ev)
					}
				}
			}
			// RAG fallback is supporting evidence, not a substitute for a required
			// structured/live source. Preserve the original gaps and incomplete flag
			// so the agent cannot mistake documentary evidence for a live business fact.
			pack.Plan.Steps = append(pack.Plan.Steps, model.RetrievalStep{ID: "fallback-rag", Source: model.SourceWikiRAG, Operation: "search_knowledge", Query: query, Purpose: "结构化检索无可用事实，补充当前 Workspace 的 WeKnora 权威文档证据；不消除原结构化数据缺口"})
		}
	}
	return pack, nil
}

func (s *Service) GetEvidence(ctx context.Context, p Principal, in EvidenceRequest) (wk.ChunkDetail, error) {
	if strings.TrimSpace(in.ChunkID) == "" {
		return wk.ChunkDetail{}, fmt.Errorf("chunk_id is required")
	}
	client, err := s.weknoraClient(p)
	if err != nil {
		return wk.ChunkDetail{}, err
	}
	detail, err := client.GetChunk(ctx, strings.TrimSpace(in.ChunkID))
	if err != nil {
		return wk.ChunkDetail{}, err
	}
	if len(in.KnowledgeBaseIDs) > 0 {
		kbID := strings.TrimSpace(detail.KnowledgeBaseID)
		allowed := false
		for _, id := range in.KnowledgeBaseIDs {
			if strings.TrimSpace(id) != "" && strings.TrimSpace(id) == kbID {
				allowed = true
				break
			}
		}
		if !allowed {
			return wk.ChunkDetail{}, ErrEvidenceOutsideScope
		}
	}
	return detail, nil
}
