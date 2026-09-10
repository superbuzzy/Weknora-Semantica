package access

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWeKnoraAccessCheckerForwardsIdentityHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/knowledge-bases/kb-1" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		expected := map[string]string{
			"Authorization":         "Bearer token",
			"X-API-Key":             "wk-api-key",
			"X-Tenant-ID":           "42",
			"X-External-User-ID":    "user-123",
			"X-External-User-Token": "external-token",
			"Accept-Language":       "zh-CN",
		}
		for name, want := range expected {
			if got := r.Header.Get(name); got != want {
				t.Fatalf("%s header = %q, want %q", name, got, want)
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := NewWeKnoraAccessChecker(server.URL)
	headers := http.Header{}
	headers.Set("Authorization", "Bearer token")
	headers.Set("X-API-Key", "wk-api-key")
	headers.Set("X-Tenant-ID", "42")
	headers.Set("X-External-User-ID", "user-123")
	headers.Set("X-External-User-Token", "external-token")
	headers.Set("Accept-Language", "zh-CN")
	if err := checker.CheckKnowledgeBaseAccess(context.Background(), "kb-1", headers); err != nil {
		t.Fatal(err)
	}
}

func TestWeKnoraAccessCheckerRejectsDeniedKB(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	checker := NewWeKnoraAccessChecker(server.URL)
	if err := checker.CheckKnowledgeBaseAccess(context.Background(), "kb-1", http.Header{}); err == nil {
		t.Fatal("expected denied access")
	}
}
