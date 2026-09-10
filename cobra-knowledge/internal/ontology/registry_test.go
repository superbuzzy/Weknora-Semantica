package ontology

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

func testOntology(version string) model.Ontology {
	return model.Ontology{
		ID:        "ont-grid",
		Domain:    "distribution_network",
		Version:   version,
		Status:    model.StatusApproved,
		CreatedAt: time.Now().UTC(),
		Classes: []model.OntologyClass{{
			ID: "cls-line", Label: "线路", Support: 10, Confidence: 1, Status: model.StatusApproved,
		}},
		Properties: []model.DataProperty{},
		Relations:  []model.ObjectRelation{},
	}
}

func TestFSRegistryLifecycleAndRollback(t *testing.T) {
	ctx := context.Background()
	r := NewFSRegistry(t.TempDir())
	if _, err := r.RegisterVersion(ctx, testOntology("1.0.0"), "alice", "initial"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Publish(ctx, "ont-grid", "1.0.0", "reviewer-a"); err != nil {
		t.Fatal(err)
	}
	if err := r.BindKnowledgeBase(ctx, model.OntologyBinding{KnowledgeBaseID: "kb-1", OntologyID: "ont-grid", Mode: "active", UpdatedBy: "alice"}); err != nil {
		t.Fatal(err)
	}
	resolved, err := r.ResolveForKnowledgeBase(ctx, "kb-1")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Ontology.Version != "1.0.0" || resolved.Version.State != model.RegistryPublished {
		t.Fatalf("unexpected resolution: %+v", resolved)
	}

	if _, err := r.RegisterVersion(ctx, testOntology("1.1.0"), "alice", "next"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Publish(ctx, "ont-grid", "1.1.0", "reviewer-b"); err != nil {
		t.Fatal(err)
	}
	resolved, err = r.ResolveForKnowledgeBase(ctx, "kb-1")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Ontology.Version != "1.1.0" {
		t.Fatalf("active binding did not follow publication: %s", resolved.Ontology.Version)
	}
	if err := r.Activate(ctx, "ont-grid", "1.0.0", "operator"); err != nil {
		t.Fatal(err)
	}
	resolved, err = r.ResolveForKnowledgeBase(ctx, "kb-1")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Ontology.Version != "1.0.0" {
		t.Fatalf("rollback failed: %s", resolved.Ontology.Version)
	}
	if events, err := r.ListAudit(ctx, 20); err != nil || len(events) < 6 {
		t.Fatalf("audit missing, len=%d err=%v", len(events), err)
	}
}

func TestFSRegistryPinnedBindingDoesNotMove(t *testing.T) {
	ctx := context.Background()
	r := NewFSRegistry(t.TempDir())
	for _, v := range []string{"1.0.0", "1.1.0"} {
		if _, err := r.RegisterVersion(ctx, testOntology(v), "alice", ""); err != nil {
			t.Fatal(err)
		}
		if _, err := r.Publish(ctx, "ont-grid", v, "reviewer"); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.BindKnowledgeBase(ctx, model.OntologyBinding{KnowledgeBaseID: "kb-pin", OntologyID: "ont-grid", Mode: "pinned", Version: "1.0.0"}); err != nil {
		t.Fatal(err)
	}
	resolved, err := r.ResolveForKnowledgeBase(ctx, "kb-pin")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Ontology.Version != "1.0.0" {
		t.Fatalf("pinned binding moved: %s", resolved.Ontology.Version)
	}
}

func TestFSRegistryVersionIsImmutable(t *testing.T) {
	ctx := context.Background()
	r := NewFSRegistry(t.TempDir())
	if _, err := r.RegisterVersion(ctx, testOntology("1.0.0"), "alice", ""); err != nil {
		t.Fatal(err)
	}
	changed := testOntology("1.0.0")
	changed.Classes[0].Label = "馈线"
	if _, err := r.RegisterVersion(ctx, changed, "bob", ""); err == nil {
		t.Fatal("expected immutable duplicate version error")
	}
}

func TestFSRegistryRejectsUnpublishedPinnedBinding(t *testing.T) {
	ctx := context.Background()
	r := NewFSRegistry(t.TempDir())
	if _, err := r.RegisterVersion(ctx, testOntology("1.0.0"), "alice", ""); err != nil {
		t.Fatal(err)
	}
	err := r.BindKnowledgeBase(ctx, model.OntologyBinding{KnowledgeBaseID: "kb-1", OntologyID: "ont-grid", Mode: "pinned", Version: "1.0.0"})
	if err == nil {
		t.Fatal("expected unpublished version rejection")
	}
	if _, err := r.GetBinding(ctx, "missing"); !errors.Is(err, ErrBindingNotFound) {
		t.Fatalf("expected ErrBindingNotFound, got %v", err)
	}
}

func TestFSRegistryDetectsPayloadTampering(t *testing.T) {
	ctx := context.Background()
	r := NewFSRegistry(t.TempDir())
	if _, err := r.RegisterVersion(ctx, testOntology("1.0.0"), "alice", ""); err != nil {
		t.Fatal(err)
	}
	path := r.versionPath("ont-grid", "1.0.0")
	if err := os.WriteFile(path, []byte(`{"id":"tampered"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := r.Get(ctx, "ont-grid", "1.0.0"); err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum mismatch, got %v", err)
	}
}

func TestFSRegistryPublishRequiresApprovedSnapshot(t *testing.T) {
	ctx := context.Background()
	r := NewFSRegistry(t.TempDir())
	o := testOntology("1.0.0")
	o.Status = model.StatusCandidate
	if _, err := r.RegisterVersion(ctx, o, "alice", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Publish(ctx, "ont-grid", "1.0.0", "reviewer"); err == nil {
		t.Fatal("expected unapproved snapshot to be rejected")
	}
}
