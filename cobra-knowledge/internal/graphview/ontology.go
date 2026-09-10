package graphview

import (
	"fmt"
	"sort"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

// BuildOntologyView projects the governed ontology into the same graph contract used
// by entity visualization. Classes and data properties are nodes. Class hierarchy,
// class-property membership, and object relations are edges.
func BuildOntologyView(kbID string, o model.Ontology) model.GraphView {
	nodes := make([]model.GraphViewNode, 0, len(o.Classes)+len(o.Properties))
	edges := make([]model.GraphViewEdge, 0)

	classByID := make(map[string]model.OntologyClass, len(o.Classes))
	for _, class := range o.Classes {
		classByID[class.ID] = class
		nodes = append(nodes, model.GraphViewNode{
			ID:       "class:" + class.ID,
			Label:    class.Label,
			Kind:     "class",
			Group:    string(class.Status),
			Subtitle: class.Description,
			Metadata: map[string]interface{}{
				"ontology_element_id": class.ID,
				"aliases":             class.Aliases,
				"support":             class.Support,
				"confidence":          class.Confidence,
				"status":              class.Status,
			},
		})
	}

	for _, property := range o.Properties {
		nodes = append(nodes, model.GraphViewNode{
			ID:       "property:" + property.ID,
			Label:    property.Label,
			Kind:     "property",
			Group:    string(property.Status),
			Subtitle: property.DataType,
			Metadata: map[string]interface{}{
				"ontology_element_id": property.ID,
				"aliases":             property.Aliases,
				"data_type":           property.DataType,
				"support":             property.Support,
				"confidence":          property.Confidence,
				"status":              property.Status,
			},
		})
		for _, domainID := range property.DomainIDs {
			if _, ok := classByID[domainID]; !ok {
				continue
			}
			edges = append(edges, model.GraphViewEdge{
				ID:     fmt.Sprintf("has-property:%s:%s", domainID, property.ID),
				Source: "class:" + domainID,
				Target: "property:" + property.ID,
				Label:  "属性",
				Kind:   "has_property",
			})
		}
	}

	for _, class := range o.Classes {
		for _, parentID := range class.ParentIDs {
			if _, ok := classByID[parentID]; !ok {
				continue
			}
			edges = append(edges, model.GraphViewEdge{
				ID:     fmt.Sprintf("subclass:%s:%s", class.ID, parentID),
				Source: "class:" + class.ID,
				Target: "class:" + parentID,
				Label:  "属于",
				Kind:   "subclass_of",
			})
		}
	}

	for _, relation := range o.Relations {
		for _, domainID := range relation.DomainIDs {
			if _, ok := classByID[domainID]; !ok {
				continue
			}
			for _, rangeID := range relation.RangeIDs {
				if _, ok := classByID[rangeID]; !ok {
					continue
				}
				edges = append(edges, model.GraphViewEdge{
					ID:     fmt.Sprintf("relation:%s:%s:%s", relation.ID, domainID, rangeID),
					Source: "class:" + domainID,
					Target: "class:" + rangeID,
					Label:  relation.Label,
					Kind:   "object_relation",
					Metadata: map[string]interface{}{
						"ontology_element_id": relation.ID,
						"confidence":          relation.Confidence,
						"support":             relation.Support,
						"status":              relation.Status,
					},
				})
			}
		}
	}

	// Stable ordering keeps snapshots and UI tests deterministic.
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	sort.Slice(edges, func(i, j int) bool { return edges[i].ID < edges[j].ID })

	return model.GraphView{
		Nodes: nodes,
		Edges: edges,
		Meta: model.GraphViewMeta{
			View:            "ontology",
			KnowledgeBaseID: kbID,
			TotalNodes:      len(nodes),
			ReturnedNodes:   len(nodes),
			ReturnedEdges:   len(edges),
			OntologyID:      o.ID,
			OntologyVersion: o.Version,
		},
	}
}
