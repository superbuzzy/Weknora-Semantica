package access

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// WeKnoraAccessChecker delegates authorization to WeKnora instead of duplicating its RBAC.
// The caller's supported WeKnora identity headers are forwarded to the standard KB read API.
type WeKnoraAccessChecker struct {
	BaseURL string
	Client  *http.Client
}

func NewWeKnoraAccessChecker(baseURL string) *WeKnoraAccessChecker {
	return &WeKnoraAccessChecker{BaseURL: strings.TrimRight(baseURL, "/"), Client: &http.Client{Timeout: 5 * time.Second}}
}

func (a *WeKnoraAccessChecker) CheckKnowledgeBaseAccess(ctx context.Context, kbID string, headers http.Header) error {
	if a.BaseURL == "" {
		return fmt.Errorf("weknora base url is not configured")
	}
	endpoint := a.BaseURL + "/api/v1/knowledge-bases/" + url.PathEscape(kbID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	for _, name := range []string{"Authorization", "X-API-Key", "X-Tenant-ID", "X-External-User-ID", "X-External-User-Token", "Accept-Language"} {
		if value := headers.Get(name); value != "" {
			req.Header.Set(name, value)
		}
	}
	resp, err := a.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("weknora access check returned %d", resp.StatusCode)
	}
	return nil
}
