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
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("authorization header not forwarded")
		}
		if r.Header.Get("X-Tenant-ID") != "42" {
			t.Fatalf("tenant header not forwarded")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := NewWeKnoraAccessChecker(server.URL)
	headers := http.Header{}
	headers.Set("Authorization", "Bearer token")
	headers.Set("X-Tenant-ID", "42")
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
