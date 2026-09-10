package weknora

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

type SearchClient struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

type searchRequest struct {
	Query            string   `json:"query"`
	KnowledgeBaseID  string   `json:"knowledge_base_id,omitempty"`
	KnowledgeBaseIDs []string `json:"knowledge_base_ids,omitempty"`
	KnowledgeIDs     []string `json:"knowledge_ids,omitempty"`
}

type searchChunk struct {
	ID                string                 `json:"id"`
	Content           string                 `json:"content"`
	KnowledgeID       string                 `json:"knowledge_id"`
	KnowledgeTitle    string                 `json:"knowledge_title"`
	Score             float64                `json:"score"`
	Metadata          map[string]interface{} `json:"metadata"`
	KnowledgeFilename string                 `json:"knowledge_filename"`
	KnowledgeSource   string                 `json:"knowledge_source"`
}

type searchResponse struct {
	Data    []searchChunk `json:"data"`
	Success bool          `json:"success"`
}

func NewSearchClient(baseURL, apiKey string) *SearchClient {
	return &SearchClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Client:  &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *SearchClient) Search(ctx context.Context, query string, knowledgeBaseIDs []string, knowledgeIDs []string) ([]model.KnowledgeItem, error) {
	if c.BaseURL == "" {
		return nil, fmt.Errorf("weknora base URL is empty")
	}
	reqBody := searchRequest{Query: query, KnowledgeBaseIDs: knowledgeBaseIDs, KnowledgeIDs: knowledgeIDs}
	if len(knowledgeBaseIDs) == 1 {
		reqBody.KnowledgeBaseID = knowledgeBaseIDs[0]
		reqBody.KnowledgeBaseIDs = nil
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/v1/knowledge-search", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("weknora search returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var parsed searchResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode weknora search: %w", err)
	}
	out := make([]model.KnowledgeItem, 0, len(parsed.Data))
	for _, chunk := range parsed.Data {
		ev := model.EvidenceRef{
			ID:            model.StableID("ev", "weknora", chunk.KnowledgeID+"\x00"+chunk.ID),
			Source:        "weknora",
			KnowledgeID:   chunk.KnowledgeID,
			ChunkID:       chunk.ID,
			DocumentTitle: chunk.KnowledgeTitle,
			Quality:       model.EvidenceExact,
			Metadata:      chunk.Metadata,
		}
		out = append(out, model.KnowledgeItem{
			ID: chunk.ID, Title: chunk.KnowledgeTitle, Content: chunk.Content,
			Source: model.SourceWikiRAG, Score: chunk.Score, Evidence: []model.EvidenceRef{ev},
		})
	}
	return out, nil
}

type ChunkDetail struct {
	ID              string                 `json:"id"`
	KnowledgeID     string                 `json:"knowledge_id"`
	KnowledgeBaseID string                 `json:"knowledge_base_id"`
	Content         string                 `json:"content"`
	ChunkIndex      int                    `json:"chunk_index"`
	StartAt         int                    `json:"start_at"`
	EndAt           int                    `json:"end_at"`
	Metadata        map[string]interface{} `json:"metadata"`
	UpdatedAt       *time.Time             `json:"updated_at,omitempty"`
}

func (c *SearchClient) GetChunk(ctx context.Context, id string) (ChunkDetail, error) {
	var empty ChunkDetail
	if c.BaseURL == "" {
		return empty, fmt.Errorf("weknora base URL is empty")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/v1/chunks/by-id/"+id, nil)
	if err != nil {
		return empty, err
	}
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return empty, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return empty, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return empty, fmt.Errorf("weknora get chunk returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var raw struct {
		Data    ChunkDetail `json:"data"`
		Success bool        `json:"success"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return empty, fmt.Errorf("decode weknora chunk: %w", err)
	}
	return raw.Data, nil
}
