package model

import "time"

type EvidenceQuality string

const (
	EvidenceExact   EvidenceQuality = "exact"
	EvidenceDerived EvidenceQuality = "derived"
)

// EvidenceRef v2 is the normalized provenance contract shared by every retriever.
// Legacy fields remain so existing WeKnora graph/chunk integrations stay compatible.
type EvidenceRef struct {
	ID              string                 `json:"id"`
	Source          string                 `json:"source"`
	SourceType      string                 `json:"source_type,omitempty"`
	SourceID        string                 `json:"source_id,omitempty"`
	KnowledgeBaseID string                 `json:"knowledge_base_id,omitempty"`
	KnowledgeID     string                 `json:"knowledge_id,omitempty"`
	EntityID        string                 `json:"entity_id,omitempty"`
	ChunkID         string                 `json:"chunk_id,omitempty"`
	DocumentTitle   string                 `json:"document_title,omitempty"`
	Value           interface{}            `json:"value,omitempty"`
	Unit            string                 `json:"unit,omitempty"`
	Version         string                 `json:"version,omitempty"`
	PublishedAt     *time.Time             `json:"published_at,omitempty"`
	EffectiveAt     *time.Time             `json:"effective_time,omitempty"`
	ObservedAt      *time.Time             `json:"observed_at,omitempty"`
	ValidFrom       *time.Time             `json:"valid_from,omitempty"`
	ValidTo         *time.Time             `json:"valid_to,omitempty"`
	Confidence      float64                `json:"confidence,omitempty"`
	Quality         EvidenceQuality        `json:"quality"`
	Provenance      map[string]interface{} `json:"provenance,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type Entity struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	BusinessType string                 `json:"business_type"`
	TypeID       string                 `json:"type_id"`
	Aliases      []string               `json:"aliases,omitempty"`
	BusinessKey  string                 `json:"business_key,omitempty"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
	Evidence     []EvidenceRef          `json:"evidence,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type Relation struct {
	ID           string                 `json:"id"`
	SourceID     string                 `json:"source_id"`
	TargetID     string                 `json:"target_id"`
	BusinessType string                 `json:"business_type"`
	TypeID       string                 `json:"type_id"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
	Evidence     []EvidenceRef          `json:"evidence,omitempty"`
	Quality      EvidenceQuality        `json:"evidence_quality"`
}

type GraphSnapshot struct {
	Domain          string            `json:"domain"`
	KnowledgeBaseID string            `json:"knowledge_base_id,omitempty"`
	KnowledgeID     string            `json:"knowledge_id,omitempty"`
	Entities        []Entity          `json:"entities"`
	Relations       []Relation        `json:"relations"`
	Warnings        []string          `json:"warnings,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}
