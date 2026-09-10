package model

import "time"

type EvidenceQuality string

const (
	EvidenceExact   EvidenceQuality = "exact"
	EvidenceDerived EvidenceQuality = "derived"
)

type EvidenceRef struct {
	ID              string                 `json:"id"`
	Source          string                 `json:"source"`
	KnowledgeBaseID string                 `json:"knowledge_base_id,omitempty"`
	KnowledgeID     string                 `json:"knowledge_id,omitempty"`
	ChunkID         string                 `json:"chunk_id,omitempty"`
	DocumentTitle   string                 `json:"document_title,omitempty"`
	Version         string                 `json:"version,omitempty"`
	PublishedAt     *time.Time             `json:"published_at,omitempty"`
	ObservedAt      *time.Time             `json:"observed_at,omitempty"`
	ValidFrom       *time.Time             `json:"valid_from,omitempty"`
	ValidTo         *time.Time             `json:"valid_to,omitempty"`
	Quality         EvidenceQuality        `json:"quality"`
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
