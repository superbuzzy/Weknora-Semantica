package graphview

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

// EntityQuery is the runtime-facing query contract. It pushes term filtering to
// Neo4j and returns the same stable GraphView used by visualization.
func (s *Neo4jHTTPSource) EntityQuery(ctx context.Context, kbID string, terms []string, limit int) (model.GraphView, error) {
	label, err := weknoraKBLabel(kbID)
	if err != nil {
		return model.GraphView{}, err
	}
	if limit <= 0 {
		limit = 200
	}
	if limit > 500 {
		limit = 500
	}
	cleanTerms := make([]string, 0, len(terms))
	for _, term := range terms {
		term = strings.ToLower(strings.TrimSpace(term))
		if term != "" {
			cleanTerms = append(cleanTerms, term)
		}
	}
	statement := fmt.Sprintf(`MATCH (n:%s)
WITH n, toLower(coalesce(n.name, '')) AS lname,
     [x IN coalesce(n.attributes, []) | toLower(toString(x))] AS attrs
WHERE size($terms) = 0 OR any(term IN $terms WHERE lname CONTAINS term OR any(a IN attrs WHERE a CONTAINS term))
RETURN coalesce(n.kg, '') AS kg, coalesce(n.name, '') AS name,
       coalesce(n.attributes, []) AS attributes, coalesce(n.chunks, []) AS chunks
ORDER BY name LIMIT $limit`, quoteCypherLabel(label))
	rows, err := s.run(ctx, statement, map[string]interface{}{"terms": cleanTerms, "limit": limit})
	if err != nil {
		return model.GraphView{}, err
	}
	nodes := make([]model.GraphViewNode, 0, len(rows))
	keys := make([]string, 0, len(rows))
	for _, row := range rows {
		kg := stringValue(row, "kg")
		name := stringValue(row, "name")
		if name == "" {
			continue
		}
		id := kg + "::" + name
		keys = append(keys, id)
		nodes = append(nodes, model.GraphViewNode{ID: id, Label: name, Kind: "entity", Group: kg, Subtitle: firstString(row["attributes"]), Metadata: map[string]interface{}{"knowledge_id": kg, "attributes": row["attributes"], "chunks": row["chunks"]}})
	}
	edges := []model.GraphViewEdge{}
	if len(keys) > 0 {
		edgeStatement := fmt.Sprintf(`MATCH (n:%s)-[r]->(m:%s)
WITH n, r, m,
     coalesce(n.kg, '') + '::' + coalesce(n.name, '') AS source_id,
     coalesce(m.kg, '') + '::' + coalesce(m.name, '') AS target_id
WHERE source_id IN $keys OR target_id IN $keys
RETURN source_id, target_id, type(r) AS rel_type LIMIT $edge_limit`, quoteCypherLabel(label), quoteCypherLabel(label))
		edgeRows, edgeErr := s.run(ctx, edgeStatement, map[string]interface{}{"keys": keys, "edge_limit": limit * 4})
		if edgeErr != nil {
			return model.GraphView{}, edgeErr
		}
		seen := map[string]bool{}
		for _, row := range edgeRows {
			sourceID := stringValue(row, "source_id")
			targetID := stringValue(row, "target_id")
			relType := stringValue(row, "rel_type")
			id := sourceID + "|" + relType + "|" + targetID
			if sourceID == "" || targetID == "" || seen[id] {
				continue
			}
			seen[id] = true
			edges = append(edges, model.GraphViewEdge{ID: id, Source: sourceID, Target: targetID, Label: relType, Kind: "entity_relation"})
		}
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	sort.Slice(edges, func(i, j int) bool { return edges[i].ID < edges[j].ID })
	return model.GraphView{Nodes: nodes, Edges: edges, Meta: model.GraphViewMeta{View: "entity", KnowledgeBaseID: kbID, TotalNodes: len(nodes), ReturnedNodes: len(nodes), ReturnedEdges: len(edges), Truncated: len(nodes) >= limit}}, nil
}
