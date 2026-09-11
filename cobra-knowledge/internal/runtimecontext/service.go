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
	EntitySource          ctxsvc.EntityQuerySource
	BusinessGateway       ctxsvc.BusinessQuerySource
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

func (s *Service) ontologiesFor(ctx context.Context, ids []string) ([]ctxsvc.NamespacedOntology, retrieval.SemanticCatalog, error) {
	items := []ctxsvc.NamespacedOntology{}
	catalogs := []retrieval.NamespacedCatalog{}
	if s.Registry == nil {
		return items, retrieval.SemanticCatalog{}, nil
	}
	for _, rawID := range ids {
		kbID := strings.TrimSpace(rawID)
		if kbID == "" {
			continue
		}
		resolved, err := s.Registry.ResolveForKnowledgeBase(ctx, kbID)
		if err != nil {
			if errors.Is(err, ontology.ErrBindingNotFound) || errors.Is(err, ontology.ErrOntologyNotFound) || errors.Is(err, ontology.ErrVersionNotFound) {
				continue
			}
			return nil, retrieval.SemanticCatalog{}, err
		}
		items = append(items, ctxsvc.NamespacedOntology{Namespace: kbID, Ontology: resolved.Ontology})
		catalogs = append(catalogs, retrieval.NamespacedCatalog{Namespace: kbID, Catalog: retrieval.CompileCatalogWithOverlay(resolved.Ontology, s.Overlay)})
	}
	return items, retrieval.FederateCatalogs(catalogs), nil
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
	ontologies, catalog, err := s.ontologiesFor(ctx, in.KnowledgeBaseIDs)
	if err != nil {
		return model.ContextPack{}, fmt.Errorf("resolve knowledge-base ontologies: %w", err)
	}

	planner := retrieval.NewPlanner(catalog)
	arbiter := retrieval.NewArbiter(retrieval.ArbitrationPolicy{DefaultSourcePriority: map[string]float64{"business_data": 1.0, "entity_graph": 0.8, "weknora": 0.6}, DefaultDelta: 0.05})
	service := ctxsvc.NewService(planner, arbiter, ctxsvc.NewAssembler())
	service.Register(&ctxsvc.WeKnoraRAGRetriever{Client: client, KnowledgeBaseIDs: in.KnowledgeBaseIDs})
	if len(ontologies) > 0 {
		service.Register(&ctxsvc.FederatedOntologyRetriever{Ontologies: ontologies})
	}
	if s.EntitySource != nil {
		service.Register(ctxsvc.NewEntityGraphRuntimeRetriever(s.EntitySource, in.KnowledgeBaseIDs))
	}
	if s.BusinessGateway != nil {
		service.Register(&ctxsvc.BusinessDataRetriever{Gateway: s.BusinessGateway})
	}

	domain := strings.TrimSpace(in.Domain)
	if domain == "" && len(ontologies) == 1 {
		domain = ontologies[0].Ontology.Domain
	}
	req := model.QueryRequest{Query: query, Domain: domain, Task: strings.TrimSpace(in.Task), Scope: map[string]string{"workspace_id": p.WorkspaceID, "user_id": p.UserID}}
	if len(in.KnowledgeBaseIDs) == 1 {
		req.Scope["knowledge_base_id"] = strings.TrimSpace(in.KnowledgeBaseIDs[0])
	}
	pack, err := service.Retrieve(ctx, req)
	if err != nil {
		return model.ContextPack{}, err
	}

	if pack.Plan.AllowFallback && len(pack.Facts) == 0 && len(pack.Knowledge) == 0 && len(pack.Paths) == 0 {
		items, ragErr := client.Search(ctx, query, in.KnowledgeBaseIDs, nil)
		if ragErr == nil && len(items) > 0 {
			pack.Knowledge = items
			seen := map[string]bool{}
			for _, item := range items {
				for _, evidence := range item.Evidence {
					if evidence.ID != "" && !seen[evidence.ID] {
						seen[evidence.ID] = true
						pack.Evidence = append(pack.Evidence, evidence)
					}
				}
			}
			pack.Plan.Steps = append(pack.Plan.Steps, model.RetrievalStep{ID: "fallback-rag", Source: model.SourceWikiRAG, Operation: "search_knowledge", Query: query, Purpose: "结构化必要来源无结果时补充当前 Workspace 的权威文档证据", Required: false})
			pack.SourceStatus = append(pack.SourceStatus, model.SourceStatus{StepID: "fallback-rag", Source: model.SourceWikiRAG, Required: false, Satisfied: true})
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
