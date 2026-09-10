package model

import "time"

// RegistryVersionState describes the release state of an immutable ontology snapshot.
// The Ontology payload itself is immutable after registration; state transitions live
// in registry metadata so publishing/rollback never rewrites historical content.
type RegistryVersionState string

const (
	RegistryCandidate RegistryVersionState = "candidate"
	RegistryPublished RegistryVersionState = "published"
)

// OntologyVersionMeta is mutable release metadata for one immutable ontology payload.
type OntologyVersionMeta struct {
	OntologyID    string               `json:"ontology_id"`
	Domain        string               `json:"domain"`
	Version       string               `json:"version"`
	State         RegistryVersionState `json:"state"`
	ContentSHA256 string               `json:"content_sha256"`
	CreatedAt     time.Time            `json:"created_at"`
	CreatedBy     string               `json:"created_by,omitempty"`
	PublishedAt   *time.Time           `json:"published_at,omitempty"`
	PublishedBy   string               `json:"published_by,omitempty"`
	ParentVersion string               `json:"parent_version,omitempty"`
	Notes         string               `json:"notes,omitempty"`
}

// OntologyManifest is the registry index for one logical ontology.
type OntologyManifest struct {
	ID            string                `json:"id"`
	Domain        string                `json:"domain"`
	ActiveVersion string                `json:"active_version,omitempty"`
	Versions      []OntologyVersionMeta `json:"versions"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

// OntologyBinding binds a WeKnora knowledge base to one logical ontology.
// Mode=active follows the registry's active published version. Mode=pinned locks
// the knowledge base to one explicit published version.
type OntologyBinding struct {
	KnowledgeBaseID string    `json:"knowledge_base_id"`
	OntologyID      string    `json:"ontology_id"`
	Mode            string    `json:"mode"` // active | pinned
	Version         string    `json:"version,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
	UpdatedBy       string    `json:"updated_by,omitempty"`
}

// OntologyResolution is the authoritative result used by retrieval and UI.
type OntologyResolution struct {
	Ontology Ontology            `json:"-"`
	Binding  OntologyBinding     `json:"binding"`
	Version  OntologyVersionMeta `json:"version"`
}

// RegistryAuditEvent records registry mutations. It is append-only.
type RegistryAuditEvent struct {
	ID              string                 `json:"id"`
	At              time.Time              `json:"at"`
	Actor           string                 `json:"actor,omitempty"`
	Action          string                 `json:"action"`
	OntologyID      string                 `json:"ontology_id,omitempty"`
	Version         string                 `json:"version,omitempty"`
	KnowledgeBaseID string                 `json:"knowledge_base_id,omitempty"`
	Details         map[string]interface{} `json:"details,omitempty"`
}
