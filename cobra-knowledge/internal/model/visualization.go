package model

// GraphViewNode is the stable visualization contract shared by entity and ontology graphs.
// The UI must not depend on Neo4j records or Ontology storage structures directly.
type GraphViewNode struct {
	ID       string                 `json:"id"`
	Label    string                 `json:"label"`
	Kind     string                 `json:"kind"`
	Group    string                 `json:"group,omitempty"`
	Subtitle string                 `json:"subtitle,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GraphViewEdge is the stable visualization contract shared by entity and ontology graphs.
type GraphViewEdge struct {
	ID       string                 `json:"id"`
	Source   string                 `json:"source"`
	Target   string                 `json:"target"`
	Label    string                 `json:"label,omitempty"`
	Kind     string                 `json:"kind,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GraphViewMeta describes a rendered slice. TotalNodes may be greater than len(Nodes)
// when the backend truncates a large graph for interactive rendering.
type GraphViewMeta struct {
	View            string `json:"view"`
	KnowledgeBaseID string `json:"knowledge_base_id"`
	TotalNodes      int    `json:"total_nodes"`
	ReturnedNodes   int    `json:"returned_nodes"`
	ReturnedEdges   int    `json:"returned_edges"`
	Truncated       bool   `json:"truncated"`
	OntologyID      string `json:"ontology_id,omitempty"`
	OntologyVersion string `json:"ontology_version,omitempty"`
}

// GraphView is the only graph shape the WeKnora overlay consumes.
type GraphView struct {
	Nodes []GraphViewNode `json:"nodes"`
	Edges []GraphViewEdge `json:"edges"`
	Meta  GraphViewMeta   `json:"meta"`
}
