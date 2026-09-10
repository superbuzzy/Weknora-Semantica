package retrieval

import "cobraknowledge.local/cobra-knowledge/internal/model"

func ArbitrationPolicyFromOntology(o model.Ontology, sourcePriority map[string]float64) ArbitrationPolicy {
	p := ArbitrationPolicy{DefaultSourcePriority: sourcePriority, Predicates: map[string]PredicatePolicy{}, DefaultDelta: 0.05}
	for _, prop := range o.Properties {
		pp := PredicatePolicy{SourcePriority: map[string]float64{}, PreferNewer: true}
		if prop.RetrievalPolicy != nil {
			pp.MaxAgeSeconds = prop.RetrievalPolicy.MaxAgeSeconds
			n := len(prop.RetrievalPolicy.PreferredSources)
			for i, s := range prop.RetrievalPolicy.PreferredSources { pp.SourcePriority[s] = 1.0 - float64(i)*0.1/float64(maxInt(n, 1)) }
		}
		p.Predicates[prop.ID] = pp
	}
	return p
}
func maxInt(a,b int)int{if a>b{return a};return b}
