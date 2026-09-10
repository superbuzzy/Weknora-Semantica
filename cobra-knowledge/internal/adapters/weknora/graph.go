package weknora

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

type GraphNode struct {
	Name       string   `json:"name,omitempty"`
	Chunks     []string `json:"chunks,omitempty"`
	Attributes []string `json:"attributes,omitempty"`
}

type GraphRelation struct {
	Node1 string `json:"node1,omitempty"`
	Node2 string `json:"node2,omitempty"`
	Type  string `json:"type,omitempty"`
}

type GraphData struct {
	Text     string          `json:"text,omitempty"`
	Node     []GraphNode     `json:"node,omitempty"`
	Relation []GraphRelation `json:"relation,omitempty"`
}

type Envelope struct {
	Domain          string    `json:"domain,omitempty"`
	KnowledgeBaseID string    `json:"knowledge_base_id,omitempty"`
	KnowledgeID     string    `json:"knowledge_id,omitempty"`
	Graph           GraphData `json:"graph"`
}

type Normalizer struct{}

func NewNormalizer() *Normalizer { return &Normalizer{} }

func (n *Normalizer) NormalizeJSON(data []byte, domain string) (model.GraphSnapshot, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err == nil && (len(env.Graph.Node) > 0 || len(env.Graph.Relation) > 0) {
		if env.Domain == "" {
			env.Domain = domain
		}
		return n.Normalize(env), nil
	}

	var graph GraphData
	if err := json.Unmarshal(data, &graph); err != nil {
		return model.GraphSnapshot{}, fmt.Errorf("decode WeKnora graph: %w", err)
	}
	return n.Normalize(Envelope{Domain: domain, Graph: graph}), nil
}

func (n *Normalizer) Normalize(env Envelope) model.GraphSnapshot {
	domain := strings.TrimSpace(env.Domain)
	if domain == "" {
		domain = "default"
	}
	snapshot := model.GraphSnapshot{
		Domain:          domain,
		KnowledgeBaseID: env.KnowledgeBaseID,
		KnowledgeID:     env.KnowledgeID,
		Metadata:        map[string]string{"adapter": "weknora-graph-v0.2"},
	}

	byName := make(map[string]model.Entity, len(env.Graph.Node))
	chunksByName := make(map[string]map[string]struct{}, len(env.Graph.Node))

	for _, raw := range env.Graph.Node {
		props, reserved := parseAttributes(raw.Attributes)
		businessType := strings.TrimSpace(reserved["__type__"])
		if businessType == "" {
			businessType = "未分类实体"
			snapshot.Warnings = append(snapshot.Warnings, fmt.Sprintf("实体 %q 缺少 __type__，归入未分类实体", raw.Name))
		}
		entityID := strings.TrimSpace(reserved["__id__"])
		if entityID == "" {
			entityID = model.StableID("ent", domain, businessType+"\x00"+raw.Name)
		}
		aliases := splitList(reserved["__aliases__"])
		businessKey := strings.TrimSpace(reserved["__business_key__"])
		ev := make([]model.EvidenceRef, 0, len(raw.Chunks))
		chunkSet := map[string]struct{}{}
		for _, chunkID := range uniqueStrings(raw.Chunks) {
			chunkSet[chunkID] = struct{}{}
			ev = append(ev, model.EvidenceRef{
				ID:              model.StableID("ev", domain, env.KnowledgeID+"\x00"+chunkID),
				Source:          "weknora",
				KnowledgeBaseID: env.KnowledgeBaseID,
				KnowledgeID:     env.KnowledgeID,
				ChunkID:         chunkID,
				Quality:         model.EvidenceExact,
			})
		}
		entity := model.Entity{
			ID:           entityID,
			Name:         strings.TrimSpace(raw.Name),
			BusinessType: businessType,
			TypeID:       model.StableID("cls", domain, businessType),
			Aliases:      aliases,
			BusinessKey:  businessKey,
			Properties:   props,
			Evidence:     ev,
			Metadata: map[string]interface{}{
				"raw_attributes": append([]string(nil), raw.Attributes...),
			},
		}
		snapshot.Entities = append(snapshot.Entities, entity)
		byName[entity.Name] = entity
		chunksByName[entity.Name] = chunkSet
	}

	for _, raw := range env.Graph.Relation {
		source, sourceOK := byName[strings.TrimSpace(raw.Node1)]
		target, targetOK := byName[strings.TrimSpace(raw.Node2)]
		if !sourceOK || !targetOK {
			snapshot.Warnings = append(snapshot.Warnings, fmt.Sprintf("关系 %q -> %q (%s) 端点不存在，已跳过", raw.Node1, raw.Node2, raw.Type))
			continue
		}
		evidence := deriveRelationEvidence(domain, env, chunksByName[source.Name], chunksByName[target.Name])
		quality := model.EvidenceDerived
		if len(evidence) > 0 {
			quality = evidence[0].Quality
		}
		label := strings.TrimSpace(raw.Type)
		snapshot.Relations = append(snapshot.Relations, model.Relation{
			ID:           model.StableID("rel", domain, source.ID+"\x00"+label+"\x00"+target.ID),
			SourceID:     source.ID,
			TargetID:     target.ID,
			BusinessType: label,
			TypeID:       model.StableID("obj", domain, label),
			Evidence:     evidence,
			Quality:      quality,
		})
	}

	sort.Slice(snapshot.Entities, func(i, j int) bool { return snapshot.Entities[i].ID < snapshot.Entities[j].ID })
	sort.Slice(snapshot.Relations, func(i, j int) bool { return snapshot.Relations[i].ID < snapshot.Relations[j].ID })
	return snapshot
}

func parseAttributes(attrs []string) (map[string]interface{}, map[string]string) {
	props := map[string]interface{}{}
	reserved := map[string]string{}
	for _, raw := range attrs {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		parts := strings.SplitN(s, "=", 2)
		if len(parts) != 2 {
			props[s] = true
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

func deriveRelationEvidence(domain string, env Envelope, left, right map[string]struct{}) []model.EvidenceRef {
	// WeKnora currently persists chunk evidence on nodes, not on the relationship
	// itself. Even when both endpoints share a chunk, relationship provenance is
	// inferred rather than exact and must remain marked as derived.
	var shared []string
	for chunk := range left {
		if _, ok := right[chunk]; ok {
			shared = append(shared, chunk)
		}
	}
	sort.Strings(shared)
	chunks := shared
	derivation := "shared_endpoint_chunk"
	if len(chunks) == 0 {
		derivation = "endpoint_chunk_union"
		union := map[string]struct{}{}
		for c := range left {
			union[c] = struct{}{}
		}
		for c := range right {
			union[c] = struct{}{}
		}
		for c := range union {
			chunks = append(chunks, c)
		}
		sort.Strings(chunks)
	}
	out := make([]model.EvidenceRef, 0, len(chunks))
	for _, chunkID := range chunks {
		out = append(out, model.EvidenceRef{
			ID:              model.StableID("evrel", domain, env.KnowledgeID+"\x00"+chunkID),
			Source:          "weknora",
			KnowledgeBaseID: env.KnowledgeBaseID,
			KnowledgeID:     env.KnowledgeID,
			ChunkID:         chunkID,
			Quality:         model.EvidenceDerived,
			Metadata: map[string]interface{}{
				"derivation": derivation,
				"note":       "WeKnora relationship does not currently persist relation-level chunk evidence",
			},
		})
	}
	return out
}

func splitList(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.FieldsFunc(v, func(r rune) bool { return r == ',' || r == '，' || r == ';' || r == '；' || r == '|' })
	return uniqueStrings(parts)
}

func uniqueStrings(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
