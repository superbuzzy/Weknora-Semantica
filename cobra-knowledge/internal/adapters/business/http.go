package business

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

type HTTPGateway struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

type queryRequest struct {
	Query             string            `json:"query"`
	Domain            string            `json:"domain,omitempty"`
	Task              string            `json:"task,omitempty"`
	Scope             map[string]string `json:"scope,omitempty"`
	TargetClassIDs    []string          `json:"target_class_ids,omitempty"`
	TargetPropertyIDs []string          `json:"target_property_ids,omitempty"`
	TargetRelationIDs []string          `json:"target_relation_ids,omitempty"`
	TargetTerms       []string          `json:"target_terms,omitempty"`
	At                *time.Time        `json:"at,omitempty"`
}

type queryResponse struct {
	Assertions []model.Assertion `json:"assertions"`
	Gaps       []string          `json:"gaps,omitempty"`
}

func NewHTTPGateway(baseURL, token string) *HTTPGateway {
	return &HTTPGateway{BaseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"), Token: strings.TrimSpace(token), Client: &http.Client{Timeout: 15 * time.Second}}
}

func (g *HTTPGateway) Query(ctx context.Context, req model.QueryRequest, step model.RetrievalStep) (model.RetrievalResult, error) {
	if g.BaseURL == "" {
		return model.RetrievalResult{Source: model.SourceBusinessData}, fmt.Errorf("business gateway base URL is not configured")
	}
	payload := queryRequest{Query: req.Query, Domain: req.Domain, Task: req.Task, Scope: req.Scope, TargetClassIDs: step.TargetClassIDs, TargetPropertyIDs: step.TargetPropertyIDs, TargetRelationIDs: step.TargetRelationIDs, TargetTerms: step.TargetTerms, At: req.At}
	raw, err := json.Marshal(payload)
	if err != nil {
		return model.RetrievalResult{Source: model.SourceBusinessData}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, g.BaseURL+"/v1/query", bytes.NewReader(raw))
	if err != nil {
		return model.RetrievalResult{Source: model.SourceBusinessData}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if g.Token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+g.Token)
	}
	resp, err := g.Client.Do(httpReq)
	if err != nil {
		return model.RetrievalResult{Source: model.SourceBusinessData}, fmt.Errorf("business gateway request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return model.RetrievalResult{Source: model.SourceBusinessData}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.RetrievalResult{Source: model.SourceBusinessData}, fmt.Errorf("business gateway returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var decoded queryResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return model.RetrievalResult{Source: model.SourceBusinessData}, fmt.Errorf("decode business gateway: %w", err)
	}
	for i := range decoded.Assertions {
		if decoded.Assertions[i].Source == "" {
			decoded.Assertions[i].Source = "business_data"
		}
		if decoded.Assertions[i].SourceType == "" {
			decoded.Assertions[i].SourceType = string(model.SourceBusinessData)
		}
		if decoded.Assertions[i].Authority == 0 {
			decoded.Assertions[i].Authority = 1
		}
		if decoded.Assertions[i].Confidence == 0 {
			decoded.Assertions[i].Confidence = 1
		}
	}
	return model.RetrievalResult{Source: model.SourceBusinessData, Assertions: decoded.Assertions, Gaps: decoded.Gaps}, nil
}
