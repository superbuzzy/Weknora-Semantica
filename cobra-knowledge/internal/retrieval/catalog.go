package retrieval

import (
	"fmt"
	"sort"
	"strings"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

type CatalogTerm struct {
	ID               string                  `json:"id"`
	Label            string                  `json:"label"`
	Aliases          []string                `json:"aliases,omitempty"`
	Kind             string                  `json:"kind"`
	PreferredSources []model.RetrievalSource `json:"preferred_sources,omitempty"`
}

type CatalogConflict struct {
	Kind       string `json:"kind"`
	Term       string `json:"term"`
	NamespaceA string `json:"namespace_a"`
	NamespaceB string `json:"namespace_b"`
	Detail     string `json:"detail"`
}

type SemanticCatalog struct {
	Domain     string            `json:"domain"`
	Namespaces []string          `json:"namespaces,omitempty"`
	Classes    []CatalogTerm     `json:"classes"`
	Properties []CatalogTerm     `json:"properties"`
	Relations  []CatalogTerm     `json:"relations"`
	Conflicts  []CatalogConflict `json:"conflicts,omitempty"`
}

type NamespacedCatalog struct {
	Namespace string
	Catalog   SemanticCatalog
}

type TermOverlay struct {
	Aliases          []string `json:"aliases,omitempty"`
	PreferredSources []string `json:"preferred_sources,omitempty"`
}

type CatalogOverlay struct {
	Classes    map[string]TermOverlay `json:"classes,omitempty"`
	Properties map[string]TermOverlay `json:"properties,omitempty"`
	Relations  map[string]TermOverlay `json:"relations,omitempty"`
}

func CompileCatalog(o model.Ontology) SemanticCatalog {
	return CompileCatalogWithOverlay(o, CatalogOverlay{})
}

func CompileCatalogWithOverlay(o model.Ontology, overlay CatalogOverlay) SemanticCatalog {
	c := SemanticCatalog{Domain: o.Domain}
	for _, cls := range o.Classes {
		ov := overlay.Classes[cls.Label]
		c.Classes = append(c.Classes, CatalogTerm{ID: cls.ID, Label: cls.Label, Aliases: mergeStrings(cls.Aliases, ov.Aliases), Kind: "class", PreferredSources: overlaySources(ov.PreferredSources, []model.RetrievalSource{model.SourceEntityGraph})})
	}
	for _, prop := range o.Properties {
		sources := []model.RetrievalSource{model.SourceEntityGraph}
		if prop.RetrievalPolicy != nil && len(prop.RetrievalPolicy.PreferredSources) > 0 {
			sources = nil
			for _, source := range prop.RetrievalPolicy.PreferredSources {
				sources = append(sources, model.RetrievalSource(strings.TrimSpace(source)))
			}
		}
		ov := overlay.Properties[prop.Label]
		c.Properties = append(c.Properties, CatalogTerm{ID: prop.ID, Label: prop.Label, Aliases: mergeStrings(prop.Aliases, ov.Aliases), Kind: "property", PreferredSources: overlaySources(ov.PreferredSources, sources)})
	}
	for _, rel := range o.Relations {
		ov := overlay.Relations[rel.Label]
		c.Relations = append(c.Relations, CatalogTerm{ID: rel.ID, Label: rel.Label, Aliases: mergeStrings(rel.Aliases, ov.Aliases), Kind: "relation", PreferredSources: overlaySources(ov.PreferredSources, []model.RetrievalSource{model.SourceEntityGraph})})
	}
	sortTerms(c.Classes)
	sortTerms(c.Properties)
	sortTerms(c.Relations)
	return c
}

// FederateCatalogs performs request-scoped ontology federation. Equal concept IDs
// are merged; equal labels with different IDs are kept visible as conflicts instead
// of being silently flattened into one concept sense.
func FederateCatalogs(inputs []NamespacedCatalog) SemanticCatalog {
	out := SemanticCatalog{Domain: "federated"}
	if len(inputs) == 1 {
		out = inputs[0].Catalog
		out.Namespaces = uniqueStrings([]string{inputs[0].Namespace})
		return out
	}
	for _, input := range inputs {
		if strings.TrimSpace(input.Namespace) != "" {
			out.Namespaces = append(out.Namespaces, strings.TrimSpace(input.Namespace))
		}
	}
	out.Namespaces = uniqueStrings(out.Namespaces)
	out.Classes, out.Conflicts = federateTerms("class", inputs, func(c SemanticCatalog) []CatalogTerm { return c.Classes }, out.Conflicts)
	out.Properties, out.Conflicts = federateTerms("property", inputs, func(c SemanticCatalog) []CatalogTerm { return c.Properties }, out.Conflicts)
	out.Relations, out.Conflicts = federateTerms("relation", inputs, func(c SemanticCatalog) []CatalogTerm { return c.Relations }, out.Conflicts)
	return out
}

func federateTerms(kind string, inputs []NamespacedCatalog, pick func(SemanticCatalog) []CatalogTerm, conflicts []CatalogConflict) ([]CatalogTerm, []CatalogConflict) {
	byID := map[string]CatalogTerm{}
	idNamespace := map[string]string{}
	labelToID := map[string]string{}
	labelNamespace := map[string]string{}
	for _, input := range inputs {
		ns := strings.TrimSpace(input.Namespace)
		for _, term := range pick(input.Catalog) {
			id := strings.TrimSpace(term.ID)
			labelKey := strings.ToLower(strings.TrimSpace(term.Label))
			if id == "" || labelKey == "" {
				continue
			}
			if priorID, ok := labelToID[labelKey]; ok && priorID != id {
				conflicts = append(conflicts, CatalogConflict{Kind: kind, Term: term.Label, NamespaceA: labelNamespace[labelKey], NamespaceB: ns, Detail: fmt.Sprintf("same label maps to different concept ids: %s vs %s", priorID, id)})
			} else {
				labelToID[labelKey] = id
				labelNamespace[labelKey] = ns
			}
			if prior, ok := byID[id]; ok {
				if !strings.EqualFold(strings.TrimSpace(prior.Label), strings.TrimSpace(term.Label)) {
					conflicts = append(conflicts, CatalogConflict{Kind: kind, Term: term.Label, NamespaceA: idNamespace[id], NamespaceB: ns, Detail: fmt.Sprintf("same concept id has different labels: %s vs %s", prior.Label, term.Label)})
					continue
				}
				prior.Aliases = mergeStrings(prior.Aliases, term.Aliases)
				prior.PreferredSources = mergeSources(prior.PreferredSources, term.PreferredSources)
				byID[id] = prior
				continue
			}
			byID[id] = term
			idNamespace[id] = ns
		}
	}
	out := make([]CatalogTerm, 0, len(byID))
	for _, term := range byID {
		out = append(out, term)
	}
	sortTerms(out)
	return out, conflicts
}

func sortTerms(items []CatalogTerm) {
	sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })
}

func termMatches(query string, term CatalogTerm) bool {
	q := strings.ToLower(query)
	if term.Label != "" && strings.Contains(q, strings.ToLower(term.Label)) {
		return true
	}
	for _, alias := range term.Aliases {
		if alias != "" && strings.Contains(q, strings.ToLower(alias)) {
			return true
		}
	}
	return false
}

func overlaySources(in []string, fallback []model.RetrievalSource) []model.RetrievalSource {
	if len(in) == 0 {
		return fallback
	}
	out := make([]model.RetrievalSource, 0, len(in))
	for _, source := range in {
		if strings.TrimSpace(source) != "" {
			out = append(out, model.RetrievalSource(strings.TrimSpace(source)))
		}
	}
	return out
}

func mergeStrings(a, b []string) []string {
	return uniqueStrings(append(append([]string(nil), a...), b...))
}

func uniqueStrings(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, value := range in {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func mergeSources(a, b []model.RetrievalSource) []model.RetrievalSource {
	seen := map[model.RetrievalSource]bool{}
	out := make([]model.RetrievalSource, 0, len(a)+len(b))
	for _, source := range append(append([]model.RetrievalSource(nil), a...), b...) {
		if source == "" || seen[source] {
			continue
		}
		seen[source] = true
		out = append(out, source)
	}
	return out
}
