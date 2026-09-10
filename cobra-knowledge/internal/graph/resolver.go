package graph

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

type ResolveGroup struct {
	Canonical model.Entity   `json:"canonical"`
	Members   []model.Entity `json:"members"`
	Reason    string         `json:"reason"`
}

type ResolveResult struct {
	Entities  []model.Entity          `json:"entities"`
	IDMap     map[string]string       `json:"id_map"`
	Groups    []ResolveGroup          `json:"groups,omitempty"`
	Conflicts []PropertyMergeConflict `json:"property_conflicts,omitempty"`
}

type PropertyMergeConflict struct {
	CanonicalID string      `json:"canonical_id"`
	Property    string      `json:"property"`
	Existing    interface{} `json:"existing"`
	Incoming    interface{} `json:"incoming"`
	SourceID    string      `json:"source_id"`
}

type EntityResolver struct{}

func NewEntityResolver() *EntityResolver { return &EntityResolver{} }

// Resolve performs conservative deterministic identity resolution.
// It only merges when a business key is identical, or when type + normalized
// canonical/alias name is identical. Fuzzy/semantic merging is deliberately
// excluded from the hard path and should go through human/LLM review.
func (r *EntityResolver) Resolve(in []model.Entity) ResolveResult {
	result := ResolveResult{IDMap: map[string]string{}}
	canonicalByKey := map[string]int{}

	for _, entity := range in {
		key, reason := identityKey(entity)
		if idx, ok := canonicalByKey[key]; ok {
			canonical := &result.Entities[idx]
			result.IDMap[entity.ID] = canonical.ID
			result.Conflicts = append(result.Conflicts, mergeEntity(canonical, entity)...)

			found := false
			for i := range result.Groups {
				if result.Groups[i].Canonical.ID == canonical.ID {
					result.Groups[i].Canonical = *canonical
					result.Groups[i].Members = append(result.Groups[i].Members, entity)
					found = true
					break
				}
			}
			if !found {
				result.Groups = append(result.Groups, ResolveGroup{
					Canonical: *canonical,
					Members:   []model.Entity{entity},
					Reason:    reason,
				})
			}
			continue
		}
		canonicalByKey[key] = len(result.Entities)
		result.Entities = append(result.Entities, cloneEntity(entity))
		result.IDMap[entity.ID] = entity.ID
	}

	sort.Slice(result.Entities, func(i, j int) bool { return result.Entities[i].ID < result.Entities[j].ID })
	return result
}

func RewriteRelations(rels []model.Relation, idMap map[string]string) []model.Relation {
	seen := map[string]struct{}{}
	out := make([]model.Relation, 0, len(rels))
	for _, rel := range rels {
		if id, ok := idMap[rel.SourceID]; ok {
			rel.SourceID = id
		}
		if id, ok := idMap[rel.TargetID]; ok {
			rel.TargetID = id
		}
		key := rel.SourceID + "\x00" + rel.TypeID + "\x00" + rel.TargetID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		rel.ID = model.StableID("rel", "resolved", key)
		out = append(out, rel)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func identityKey(e model.Entity) (string, string) {
	typ := e.TypeID
	if strings.TrimSpace(e.BusinessKey) != "" {
		return "business:" + typ + ":" + normalize(e.BusinessKey), "same business key"
	}
	names := append([]string{e.Name}, e.Aliases...)
	normalized := ""
	for _, n := range names {
		if x := normalize(n); x != "" && (normalized == "" || x < normalized) {
			normalized = x
		}
	}
	if normalized == "" {
		return "id:" + e.ID, "missing usable identity key; keep separate"
	}
	return "name:" + typ + ":" + normalized, "same type and normalized name/alias"
}

func normalize(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func mergeEntity(dst *model.Entity, src model.Entity) []PropertyMergeConflict {
	dst.Aliases = uniqueStrings(append(append(dst.Aliases, src.Aliases...), src.Name))
	dst.Evidence = mergeEvidence(dst.Evidence, src.Evidence)
	if dst.Properties == nil {
		dst.Properties = map[string]interface{}{}
	}
	var conflicts []PropertyMergeConflict
	for k, v := range src.Properties {
		if old, ok := dst.Properties[k]; ok && normalizedValue(old) != normalizedValue(v) {
			conflicts = append(conflicts, PropertyMergeConflict{
				CanonicalID: dst.ID,
				Property:    k,
				Existing:    old,
				Incoming:    v,
				SourceID:    src.ID,
			})
			continue
		}
		dst.Properties[k] = v
	}
	return conflicts
}

func cloneEntity(e model.Entity) model.Entity {
	out := e
	out.Aliases = append([]string(nil), e.Aliases...)
	out.Evidence = append([]model.EvidenceRef(nil), e.Evidence...)
	out.Properties = map[string]interface{}{}
	for k, v := range e.Properties {
		out.Properties[k] = v
	}
	out.Metadata = map[string]interface{}{}
	for k, v := range e.Metadata {
		out.Metadata[k] = v
	}
	return out
}

func mergeEvidence(a, b []model.EvidenceRef) []model.EvidenceRef {
	seen := map[string]struct{}{}
	out := make([]model.EvidenceRef, 0, len(a)+len(b))
	all := append(append([]model.EvidenceRef(nil), a...), b...)
	for _, e := range all {
		if _, ok := seen[e.ID]; ok {
			continue
		}
		seen[e.ID] = struct{}{}
		out = append(out, e)
	}
	return out
}

func uniqueStrings(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func normalizedValue(v interface{}) string {
	return normalize(fmt.Sprint(v))
}
