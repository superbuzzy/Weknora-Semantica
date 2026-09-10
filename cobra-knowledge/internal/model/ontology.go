package model

import "time"

type LifecycleStatus string

const (
	StatusCandidate  LifecycleStatus = "candidate"
	StatusApproved   LifecycleStatus = "approved"
	StatusRejected   LifecycleStatus = "rejected"
	StatusSuperseded LifecycleStatus = "superseded"
)

type SourceBinding struct {
	Source     string            `json:"source"`
	SourceType string            `json:"source_type"`
	Locator    string            `json:"locator,omitempty"`
	Field      string            `json:"field,omitempty"`
	JoinKey    string            `json:"join_key,omitempty"`
	Priority   int               `json:"priority,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type RetrievalPolicy struct {
	PreferredSources []string `json:"preferred_sources,omitempty"`
	FreshnessMode    string   `json:"freshness_mode,omitempty"`
	MaxAgeSeconds    int64    `json:"max_age_seconds,omitempty"`
	ConflictMode     string   `json:"conflict_mode,omitempty"`
	RequireEvidence  bool     `json:"require_evidence,omitempty"`
}

type OntologyClass struct {
	ID          string          `json:"id"`
	Label       string          `json:"label"`
	Aliases     []string        `json:"aliases,omitempty"`
	ParentIDs   []string        `json:"parent_ids,omitempty"`
	Description string          `json:"description,omitempty"`
	Support     int             `json:"support"`
	Confidence  float64         `json:"confidence"`
	Status      LifecycleStatus `json:"status"`
}

type DataProperty struct {
	ID              string           `json:"id"`
	Label           string           `json:"label"`
	Aliases         []string         `json:"aliases,omitempty"`
	DomainIDs       []string         `json:"domain_ids"`
	DataType        string           `json:"data_type"`
	Description     string           `json:"description,omitempty"`
	Support         int              `json:"support"`
	Confidence      float64          `json:"confidence"`
	Status          LifecycleStatus  `json:"status"`
	SourceBindings  []SourceBinding  `json:"source_bindings,omitempty"`
	RetrievalPolicy *RetrievalPolicy `json:"retrieval_policy,omitempty"`
}

type ObjectRelation struct {
	ID          string          `json:"id"`
	Label       string          `json:"label"`
	Aliases     []string        `json:"aliases,omitempty"`
	DomainIDs   []string        `json:"domain_ids"`
	RangeIDs    []string        `json:"range_ids"`
	Description string          `json:"description,omitempty"`
	Support     int             `json:"support"`
	Confidence  float64         `json:"confidence"`
	Status      LifecycleStatus `json:"status"`
}

type Constraint struct {
	ID          string                 `json:"id"`
	Kind        string                 `json:"kind"`
	TargetID    string                 `json:"target_id"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Description string                 `json:"description,omitempty"`
	Status      LifecycleStatus        `json:"status"`
}

type ReviewItem struct {
	ID          string   `json:"id"`
	Kind        string   `json:"kind"`
	SubjectIDs  []string `json:"subject_ids,omitempty"`
	Question    string   `json:"question"`
	Risk        string   `json:"risk"`
	SuggestedBy string   `json:"suggested_by"`
	Status      string   `json:"status"`
}

type Ontology struct {
	ID          string            `json:"id"`
	Domain      string            `json:"domain"`
	Version     string            `json:"version"`
	Status      LifecycleStatus   `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	Classes     []OntologyClass   `json:"classes"`
	Properties  []DataProperty    `json:"properties"`
	Relations   []ObjectRelation  `json:"relations"`
	Constraints []Constraint      `json:"constraints,omitempty"`
	ReviewQueue []ReviewItem      `json:"review_queue,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}
