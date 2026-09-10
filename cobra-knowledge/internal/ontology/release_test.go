package ontology

import (
	"testing"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

func TestApproveSnapshotCreatesNewApprovedVersion(t *testing.T) {
	o := testOntology("0.1.0-candidate")
	o.Status = model.StatusCandidate
	o.Classes[0].Status = model.StatusCandidate
	approved, err := ApproveSnapshot(o, "1.0.0", "reviewer")
	if err != nil {
		t.Fatal(err)
	}
	if approved.Version != "1.0.0" || approved.Status != model.StatusApproved || approved.Classes[0].Status != model.StatusApproved {
		t.Fatalf("unexpected approved snapshot: %+v", approved)
	}
	if o.Version != "0.1.0-candidate" || o.Status != model.StatusCandidate {
		t.Fatalf("source candidate was mutated: %+v", o)
	}
}
