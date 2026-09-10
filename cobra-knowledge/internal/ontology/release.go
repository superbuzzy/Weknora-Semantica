package ontology

import (
	"fmt"
	"strings"
	"time"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

// ApproveSnapshot creates a new approved ontology snapshot from reviewed candidate
// content. It never edits an already-registered registry version.
func ApproveSnapshot(o model.Ontology, version, reviewer string) (model.Ontology, error) {
	version = strings.TrimSpace(version)
	if version == "" {
		return model.Ontology{}, fmt.Errorf("approved version is required")
	}
	for _, item := range o.ReviewQueue {
		if item.Risk == "high" && item.Status != "approved" {
			return model.Ontology{}, fmt.Errorf("high-risk review item %s is not approved", item.ID)
		}
	}
	if violations := ValidateOntology(o); hasValidationErrors(violations) {
		return model.Ontology{}, fmt.Errorf("ontology validation failed: %s", firstValidationError(violations))
	}
	approved := o
	approved.Version = version
	approved.Status = model.StatusApproved
	approved.CreatedAt = time.Now().UTC()
	if approved.Metadata == nil {
		approved.Metadata = map[string]string{}
	}
	approved.Metadata["reviewer"] = strings.TrimSpace(reviewer)
	if o.Version != "" && o.Version != version {
		approved.Metadata["parent_version"] = o.Version
	}
	for i := range approved.Classes {
		if approved.Classes[i].Status == model.StatusCandidate {
			approved.Classes[i].Status = model.StatusApproved
		}
	}
	for i := range approved.Properties {
		if approved.Properties[i].Status == model.StatusCandidate {
			approved.Properties[i].Status = model.StatusApproved
		}
	}
	for i := range approved.Relations {
		if approved.Relations[i].Status == model.StatusCandidate {
			approved.Relations[i].Status = model.StatusApproved
		}
	}
	for i := range approved.Constraints {
		if approved.Constraints[i].Status == model.StatusCandidate {
			approved.Constraints[i].Status = model.StatusApproved
		}
	}
	return approved, nil
}
