package context

import (
	"sort"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

type Assembler struct{ MaxKnowledge int }

func NewAssembler() *Assembler { return &Assembler{MaxKnowledge: 12} }

func (a *Assembler) Assemble(plan model.RetrievalPlan, results []model.RetrievalResult, arbitration []model.ArbitrationResult) model.ContextPack {
	pack := model.ContextPack{Query: plan.Query, Plan: plan, Arbitration: arbitration, Complete: true}
	evSeen := map[string]struct{}{}
	knowSeen := map[string]struct{}{}
	for _, ar := range arbitration {
		if ar.Accepted != nil {
			pack.Facts = append(pack.Facts, *ar.Accepted)
			for _, e := range ar.Accepted.Evidence {
				if _, ok := evSeen[e.ID]; !ok {
					evSeen[e.ID] = struct{}{}
					pack.Evidence = append(pack.Evidence, e)
				}
			}
		}
		if ar.Status == "unresolved_conflict" {
			pack.Conflicts = append(pack.Conflicts, ar)
			pack.Complete = false
		}
	}
	for _, res := range results {
		for _, k := range res.Knowledge {
			key := k.ID
			if key == "" {
				key = k.Content
			}
			if _, ok := knowSeen[key]; ok {
				continue
			}
			knowSeen[key] = struct{}{}
			pack.Knowledge = append(pack.Knowledge, k)
			for _, e := range k.Evidence {
				if _, ok := evSeen[e.ID]; !ok {
					evSeen[e.ID] = struct{}{}
					pack.Evidence = append(pack.Evidence, e)
				}
			}
		}
		pack.Paths = append(pack.Paths, res.Paths...)
		pack.Gaps = append(pack.Gaps, res.Gaps...)
	}
	sort.Slice(pack.Knowledge, func(i, j int) bool { return pack.Knowledge[i].Score > pack.Knowledge[j].Score })
	if a.MaxKnowledge > 0 && len(pack.Knowledge) > a.MaxKnowledge {
		pack.Knowledge = pack.Knowledge[:a.MaxKnowledge]
	}
	if len(pack.Gaps) > 0 {
		pack.Complete = false
	}
	if len(pack.Facts) == 0 && len(pack.Knowledge) == 0 && len(pack.Paths) == 0 {
		pack.Complete = false
	}
	return pack
}
