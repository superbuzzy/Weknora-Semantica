package context

import (
	"fmt"
	"sort"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

type Assembler struct{ MaxKnowledge int }

func NewAssembler() *Assembler { return &Assembler{MaxKnowledge: 12} }

func (a *Assembler) Assemble(plan model.RetrievalPlan, results []model.RetrievalResult, arbitration []model.ArbitrationResult) model.ContextPack {
	pack := model.ContextPack{Query: plan.Query, Plan: plan, Arbitration: arbitration, Complete: true}
	evidenceSeen := map[string]struct{}{}
	knowledgeSeen := map[string]struct{}{}

	for _, issue := range plan.BlockingIssues {
		appendUniqueString(&pack.BlockingGaps, issue)
	}

	for _, result := range results {
		pack.SourceStatus = append(pack.SourceStatus, model.SourceStatus{StepID: result.StepID, Source: result.Source, Required: result.Required, Satisfied: result.Satisfied, Gaps: append([]string(nil), result.Gaps...)})
		for _, gap := range result.Gaps {
			appendUniqueString(&pack.Gaps, gap)
		}
		if result.Required && !result.Satisfied {
			appendUniqueString(&pack.BlockingGaps, fmt.Sprintf("required source %s was not satisfied", result.Source))
		}
	}

	for _, ar := range arbitration {
		if ar.Accepted != nil {
			pack.Facts = append(pack.Facts, *ar.Accepted)
			appendEvidence(&pack.Evidence, evidenceSeen, ar.Accepted.Evidence)
		}
		if ar.Status == "unresolved_conflict" {
			pack.Conflicts = append(pack.Conflicts, ar)
			appendUniqueString(&pack.BlockingGaps, "unresolved fact conflict: "+ar.SlotID)
		}
	}

	for _, result := range results {
		for _, knowledge := range result.Knowledge {
			key := knowledge.ID
			if key == "" {
				key = knowledge.Content
			}
			if _, ok := knowledgeSeen[key]; ok {
				continue
			}
			knowledgeSeen[key] = struct{}{}
			pack.Knowledge = append(pack.Knowledge, knowledge)
			appendEvidence(&pack.Evidence, evidenceSeen, knowledge.Evidence)
		}
		for _, path := range result.Paths {
			pack.Paths = append(pack.Paths, path)
			appendEvidence(&pack.Evidence, evidenceSeen, path.Evidence)
		}
	}

	sort.Slice(pack.Knowledge, func(i, j int) bool { return pack.Knowledge[i].Score > pack.Knowledge[j].Score })
	if a.MaxKnowledge > 0 && len(pack.Knowledge) > a.MaxKnowledge {
		pack.Knowledge = pack.Knowledge[:a.MaxKnowledge]
	}
	sort.Slice(pack.SourceStatus, func(i, j int) bool { return pack.SourceStatus[i].StepID < pack.SourceStatus[j].StepID })

	if len(pack.BlockingGaps) > 0 {
		pack.Complete = false
	}
	if len(pack.Facts) == 0 && len(pack.Knowledge) == 0 && len(pack.Paths) == 0 {
		pack.Complete = false
	}
	return pack
}

func appendEvidence(dst *[]model.EvidenceRef, seen map[string]struct{}, items []model.EvidenceRef) {
	for _, evidence := range items {
		key := evidence.ID
		if key == "" {
			key = evidence.Source + "\x00" + evidence.SourceID + "\x00" + evidence.ChunkID
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		*dst = append(*dst, evidence)
	}
}

func appendUniqueString(dst *[]string, value string) {
	if value == "" {
		return
	}
	for _, existing := range *dst {
		if existing == value {
			return
		}
	}
	*dst = append(*dst, value)
}
