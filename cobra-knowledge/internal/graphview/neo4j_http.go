package graphview

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

// Neo4jHTTPSource reads WeKnora's existing Neo4j GraphRAG storage through Neo4j's
// transactional HTTP endpoint. It avoids importing WeKnora or a Neo4j driver into
// CobraKnowledge and therefore keeps the integration boundary replaceable.
type Neo4jHTTPSource struct {
	BaseURL  string
	Database string
	Username string
	Password string
	Client   *http.Client
}

func NewNeo4jHTTPSource(baseURL, database, username, password string) *Neo4jHTTPSource {
	return &Neo4jHTTPSource{
		BaseURL:  strings.TrimRight(baseURL, "/"),
		Database: database,
		Username: username,
		Password: password,
		Client:   &http.Client{Timeout: 10 * time.Second},
	}
}

var kbIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func weknoraKBLabel(kbID string) (string, error) {
	if !kbIDPattern.MatchString(kbID) {
		return "", fmt.Errorf("invalid knowledge base id")
	}
	return "ENTITY" + strings.ReplaceAll(kbID, "-", "_"), nil
}

func (s *Neo4jHTTPSource) endpoint() string {
	db := s.Database
	if db == "" {
		db = "neo4j"
	}
	return fmt.Sprintf("%s/db/%s/tx/commit", s.BaseURL, db)
}

func (s *Neo4jHTTPSource) EntityView(ctx context.Context, kbID string, limit int) (model.GraphView, error) {
	label, err := weknoraKBLabel(kbID)
	if err != nil {
		return model.GraphView{}, err
	}
	if limit <= 0 {
		limit = 160
	}
	if limit > 500 {
		limit = 500
	}

	nodeStatement := fmt.Sprintf(`MATCH (n:%s)
RETURN coalesce(n.kg, '') AS kg, coalesce(n.name, '') AS name,
       coalesce(n.attributes, []) AS attributes, coalesce(n.chunks, []) AS chunks
ORDER BY name LIMIT $limit`, quoteCypherLabel(label))
	rows, err := s.run(ctx, nodeStatement, map[string]interface{}{"limit": limit})
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
		nodes = append(nodes, model.GraphViewNode{
			ID:       id,
			Label:    name,
			Kind:     "entity",
			Group:    kg,
			Subtitle: firstString(row["attributes"]),
			Metadata: map[string]interface{}{
				"knowledge_id": kg,
				"attributes":   row["attributes"],
				"chunks":       row["chunks"],
			},
		})
	}

	edges := []model.GraphViewEdge{}
	if len(keys) > 0 {
		edgeStatement := fmt.Sprintf(`MATCH (n:%s)-[r]->(m:%s)
WITH n, r, m,
     coalesce(n.kg, '') + '::' + coalesce(n.name, '') AS source_id,
     coalesce(m.kg, '') + '::' + coalesce(m.name, '') AS target_id
WHERE source_id IN $keys AND target_id IN $keys
RETURN source_id, target_id, type(r) AS rel_type`, quoteCypherLabel(label), quoteCypherLabel(label))
		edgeRows, edgeErr := s.run(ctx, edgeStatement, map[string]interface{}{"keys": keys})
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

	total := len(nodes)
	countStatement := fmt.Sprintf(`MATCH (n:%s) RETURN count(n) AS total`, quoteCypherLabel(label))
	if countRows, countErr := s.run(ctx, countStatement, nil); countErr == nil && len(countRows) > 0 {
		if value, ok := numberAsInt(countRows[0]["total"]); ok {
			total = value
		}
	}

	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	sort.Slice(edges, func(i, j int) bool { return edges[i].ID < edges[j].ID })
	return model.GraphView{
		Nodes: nodes,
		Edges: edges,
		Meta: model.GraphViewMeta{
			View:            "entity",
			KnowledgeBaseID: kbID,
			TotalNodes:      total,
			ReturnedNodes:   len(nodes),
			ReturnedEdges:   len(edges),
			Truncated:       total > len(nodes),
		},
	}, nil
}

func quoteCypherLabel(label string) string { return "`" + strings.ReplaceAll(label, "`", "``") + "`" }

func (s *Neo4jHTTPSource) run(ctx context.Context, statement string, params map[string]interface{}) ([]map[string]interface{}, error) {
	if strings.TrimSpace(s.BaseURL) == "" {
		return nil, fmt.Errorf("neo4j base url is not configured")
	}
	payload := map[string]interface{}{
		"statements": []interface{}{map[string]interface{}{
			"statement":          statement,
			"parameters":         params,
			"resultDataContents": []string{"row"},
		}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.Username != "" {
		req.SetBasicAuth(s.Username, s.Password)
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("neo4j request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("neo4j http %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var decoded struct {
		Results []struct {
			Columns []string `json:"columns"`
			Data    []struct {
				Row []interface{} `json:"row"`
			} `json:"data"`
		} `json:"results"`
		Errors []struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, fmt.Errorf("decode neo4j response: %w", err)
	}
	if len(decoded.Errors) > 0 {
		return nil, fmt.Errorf("neo4j query %s: %s", decoded.Errors[0].Code, decoded.Errors[0].Message)
	}
	if len(decoded.Results) == 0 {
		return nil, nil
	}
	result := decoded.Results[0]
	rows := make([]map[string]interface{}, 0, len(result.Data))
	for _, item := range result.Data {
		row := make(map[string]interface{}, len(result.Columns))
		for i, col := range result.Columns {
			if i < len(item.Row) {
				row[col] = item.Row[i]
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func stringValue(row map[string]interface{}, key string) string {
	v, _ := row[key].(string)
	return v
}

func firstString(v interface{}) string {
	items, ok := v.([]interface{})
	if !ok || len(items) == 0 {
		return ""
	}
	text, _ := items[0].(string)
	return text
}

func numberAsInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	default:
		return 0, false
	}
}
