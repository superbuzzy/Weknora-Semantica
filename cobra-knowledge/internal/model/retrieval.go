package model

import "time"

type RetrievalSource string

const (
	SourceBusinessData  RetrievalSource = "business_data"
	SourceEntityGraph   RetrievalSource = "entity_graph"
	SourceWikiRAG       RetrievalSource = "wiki_rag"
	SourceOntologyGraph RetrievalSource = "ontology_graph"
)

type QueryRequest struct {
	Query  string            `json:"query"`
	Domain string            `json:"domain,omitempty"`
	At     *time.Time        `json:"at,omitempty"`
	Scope  map[string]string `json:"scope,omitempty"`
	Task   string            `json:"task,omitempty"`
}

type QuerySemantics struct {
	Intent            string            `json:"intent"`
	MatchedClasses    []string          `json:"matched_classes,omitempty"`
	MatchedProperties []string          `json:"matched_properties,omitempty"`
	MatchedRelations  []string          `json:"matched_relations,omitempty"`
	Scope             map[string]string `json:"scope,omitempty"`
}

type RetrievalStep struct {
	ID                   string          `json:"id"`
	Source               RetrievalSource `json:"source"`
	Operation            string          `json:"operation"`
	Query                string          `json:"query"`
	Purpose              string          `json:"purpose"`
	Required             bool            `json:"required"`
	FreshnessRequirement string          `json:"freshness_requirement,omitempty"`
	TargetClassIDs       []string        `json:"target_class_ids,omitempty"`
	TargetPropertyIDs    []string        `json:"target_property_ids,omitempty"`
	TargetRelationIDs    []string        `json:"target_relation_ids,omitempty"`
	TargetTerms          []string        `json:"target_terms,omitempty"`
	DependsOn            []string        `json:"depends_on,omitempty"`
	ParallelGroup        string          `json:"parallel_group,omitempty"`
}

type RetrievalPlan struct {
	Query                   string            `json:"query"`
	Mode                    string            `json:"mode"`
	Semantics               QuerySemantics    `json:"semantics"`
	Steps                   []RetrievalStep   `json:"steps"`
	RequiredSources         []RetrievalSource `json:"required_sources,omitempty"`
	SupportingSources       []RetrievalSource `json:"supporting_sources,omitempty"`
	FreshnessRequirement    string            `json:"freshness_requirement,omitempty"`
	OntologyScope           []string          `json:"ontology_scope,omitempty"`
	AllowFallback           bool              `json:"allow_fallback"`
	BlockingIssues          []string          `json:"blocking_issues,omitempty"`
	NeedArbitration         bool              `json:"need_arbitration"`
	NeedEvidence            bool              `json:"need_evidence"`
	StopCondition           string            `json:"stop_condition"`
	Reasons                 []string          `json:"reasons,omitempty"`
	RequiresLLMPlanning     bool              `json:"requires_llm_planning"`
}

type Assertion struct {
	ID          string            `json:"id"`
	SlotID      string            `json:"slot_id"`
	SubjectID   string            `json:"subject_id"`
	PredicateID string            `json:"predicate_id"`
	Value       interface{}       `json:"value"`
	Scope       map[string]string `json:"scope,omitempty"`
	Source      string            `json:"source"`
	SourceType  string            `json:"source_type,omitempty"`
	Authority   float64           `json:"authority"`
	Confidence  float64           `json:"confidence"`
	ObservedAt  *time.Time        `json:"observed_at,omitempty"`
	ValidFrom   *time.Time        `json:"valid_from,omitempty"`
	ValidTo     *time.Time        `json:"valid_to,omitempty"`
	Version     string            `json:"version,omitempty"`
	Supersedes  []string          `json:"supersedes,omitempty"`
	Evidence    []EvidenceRef     `json:"evidence,omitempty"`
}

type ArbitrationResult struct {
	SlotID    string      `json:"slot_id"`
	Status    string      `json:"status"`
	Accepted  *Assertion  `json:"accepted,omitempty"`
	Conflicts []Assertion `json:"conflicts,omitempty"`
	Stale     []Assertion `json:"stale,omitempty"`
	Reason    string      `json:"reason"`
}

type KnowledgeItem struct {
	ID       string          `json:"id"`
	Title    string          `json:"title,omitempty"`
	Content  string          `json:"content"`
	Source   RetrievalSource `json:"source"`
	Score    float64         `json:"score,omitempty"`
	Evidence []EvidenceRef   `json:"evidence,omitempty"`
}

type RelationPath struct {
	Nodes     []string      `json:"nodes"`
	Relations []string      `json:"relations"`
	Evidence  []EvidenceRef `json:"evidence,omitempty"`
}

type RetrievalResult struct {
	StepID     string          `json:"step_id,omitempty"`
	Source     RetrievalSource `json:"source"`
	Required   bool            `json:"required"`
	Satisfied  bool            `json:"satisfied"`
	Assertions []Assertion     `json:"assertions,omitempty"`
	Knowledge  []KnowledgeItem `json:"knowledge,omitempty"`
	Paths      []RelationPath  `json:"paths,omitempty"`
	Gaps       []string        `json:"gaps,omitempty"`
}

type SourceStatus struct {
	StepID    string          `json:"step_id"`
	Source    RetrievalSource `json:"source"`
	Required  bool            `json:"required"`
	Satisfied bool            `json:"satisfied"`
	Gaps      []string        `json:"gaps,omitempty"`
}

type ContextPack struct {
	Query        string              `json:"query"`
	Plan         RetrievalPlan       `json:"plan"`
	Facts        []Assertion         `json:"facts,omitempty"`
	Knowledge    []KnowledgeItem     `json:"knowledge,omitempty"`
	Paths        []RelationPath      `json:"paths,omitempty"`
	Arbitration  []ArbitrationResult `json:"arbitration,omitempty"`
	Evidence     []EvidenceRef       `json:"evidence,omitempty"`
	Conflicts    []ArbitrationResult `json:"conflicts,omitempty"`
	SourceStatus []SourceStatus      `json:"source_status,omitempty"`
	Gaps         []string            `json:"gaps,omitempty"`
	BlockingGaps []string            `json:"blocking_gaps,omitempty"`
	Complete     bool                `json:"complete"`
}
