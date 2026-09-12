package promotion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type HTTPError struct { StatusCode int; Method string; URL string; Body string }
func (e *HTTPError) Error() string { return fmt.Sprintf("%s %s returned %d: %s", e.Method, e.URL, e.StatusCode, strings.TrimSpace(e.Body)) }

func doJSON(ctx context.Context, client *http.Client, method, endpoint string, headers http.Header, body interface{}) (interface{}, error) {
	var reader io.Reader
	if body != nil { raw, err := json.Marshal(body); if err != nil { return nil, err }; reader = bytes.NewReader(raw) }
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader); if err != nil { return nil, err }
	for key, values := range headers { for _, value := range values { req.Header.Add(key, value) } }
	if body != nil { req.Header.Set("Content-Type", "application/json") }
	resp, err := client.Do(req); if err != nil { return nil, err }
	defer resp.Body.Close(); raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20)); if err != nil { return nil, err }
	if resp.StatusCode < 200 || resp.StatusCode >= 300 { return nil, &HTTPError{StatusCode: resp.StatusCode, Method: method, URL: endpoint, Body: string(raw)} }
	if len(bytes.TrimSpace(raw)) == 0 { return map[string]interface{}{}, nil }
	var payload interface{}; if err := json.Unmarshal(raw, &payload); err != nil { return nil, fmt.Errorf("decode downstream response: %w", err) }; return payload, nil
}

func unwrapObject(value interface{}) map[string]interface{} {
	m, _ := value.(map[string]interface{}); if m == nil { return map[string]interface{}{} }
	for _, key := range []string{"result", "data"} { if nested, ok := m[key].(map[string]interface{}); ok { return nested } }; return m
}
func unwrapArray(value interface{}) []interface{} {
	m, _ := value.(map[string]interface{}); if m == nil { return nil }
	for _, key := range []string{"result", "data"} { if items, ok := m[key].([]interface{}); ok { return items }; if nested, ok := m[key].(map[string]interface{}); ok { for _, child := range []string{"items", "knowledge", "skills"} { if items, ok := nested[child].([]interface{}); ok { return items } } } }
	for _, key := range []string{"items", "knowledge", "skills"} { if items, ok := m[key].([]interface{}); ok { return items } }; return nil
}
func stringField(m map[string]interface{}, keys ...string) string { for _, key := range keys { if value, ok := m[key].(string); ok && strings.TrimSpace(value) != "" { return strings.TrimSpace(value) } }; return "" }
func floatField(m map[string]interface{}, keys ...string) float64 { for _, key := range keys { switch value := m[key].(type) { case float64: return value; case float32: return float64(value); case int: return float64(value) } }; return 0 }
func endpoint(baseURL, path string) string { return strings.TrimRight(strings.TrimSpace(baseURL), "/") + "/" + strings.TrimLeft(path, "/") }
