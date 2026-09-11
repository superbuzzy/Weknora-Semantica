package context

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cobraknowledge.local/cobra-knowledge/internal/model"
	"cobraknowledge.local/cobra-knowledge/internal/retrieval"
)

type Retriever interface {
	Source() model.RetrievalSource
	Retrieve(ctx context.Context, req model.QueryRequest, step model.RetrievalStep) (model.RetrievalResult, error)
}

type ComplexPlanRefiner interface {
	Refine(ctx context.Context, req model.QueryRequest, base model.RetrievalPlan) (model.RetrievalPlan, error)
}

type Service struct {
	Planner    *retrieval.Planner
	Arbiter    *retrieval.Arbiter
	Assembler  *Assembler
	Refiner    ComplexPlanRefiner
	Retrievers map[model.RetrievalSource]Retriever
}

func NewService(planner *retrieval.Planner, arbiter *retrieval.Arbiter, assembler *Assembler) *Service {
	return &Service{Planner: planner, Arbiter: arbiter, Assembler: assembler, Retrievers: map[model.RetrievalSource]Retriever{}}
}

func (s *Service) Register(r Retriever) {
	if r != nil {
		s.Retrievers[r.Source()] = r
	}
}

func (s *Service) Retrieve(ctx context.Context, req model.QueryRequest) (model.ContextPack, error) {
	if s.Planner == nil || s.Arbiter == nil || s.Assembler == nil {
		return model.ContextPack{}, fmt.Errorf("context service is not fully configured")
	}
	plan := s.Planner.Plan(req)
	if plan.RequiresLLMPlanning && s.Refiner != nil {
		refined, err := s.Refiner.Refine(ctx, req, plan)
		if err != nil {
			return model.ContextPack{}, fmt.Errorf("refine retrieval plan: %w", err)
		}
		plan = refined
	}

	type item struct {
		step   model.RetrievalStep
		result model.RetrievalResult
		err    error
	}
	ch := make(chan item, len(plan.Steps))
	var wg sync.WaitGroup
	for _, step := range plan.Steps {
		r, ok := s.Retrievers[step.Source]
		if !ok {
			ch <- item{step: step, result: model.RetrievalResult{StepID: step.ID, Source: step.Source, Required: step.Required, Satisfied: false, Gaps: []string{"retriever not configured: " + string(step.Source)}}}
			continue
		}
		wg.Add(1)
		go func(r Retriever, step model.RetrievalStep) {
			defer wg.Done()
			res, err := r.Retrieve(ctx, req, step)
			ch <- item{step: step, result: res, err: err}
		}(r, step)
	}
	wg.Wait()
	close(ch)

	results := make([]model.RetrievalResult, 0, len(plan.Steps))
	assertions := []model.Assertion{}
	for item := range ch {
		res := item.result
		res.StepID = item.step.ID
		res.Source = item.step.Source
		res.Required = item.step.Required
		if item.err != nil {
			res.Satisfied = false
			res.Gaps = append(res.Gaps, item.err.Error())
			results = append(results, res)
			continue
		}
		res.Satisfied = hasUsableResult(res)
		if !res.Satisfied && len(res.Gaps) == 0 {
			res.Gaps = append(res.Gaps, "source returned no usable result: "+string(res.Source))
		}
		results = append(results, res)
		assertions = append(assertions, res.Assertions...)
	}

	queryTime := time.Now().UTC()
	if req.At != nil {
		queryTime = req.At.UTC()
	}
	arbitration := s.Arbiter.Arbitrate(assertions, queryTime)
	return s.Assembler.Assemble(plan, results, arbitration), nil
}

func hasUsableResult(res model.RetrievalResult) bool {
	return len(res.Assertions) > 0 || len(res.Knowledge) > 0 || len(res.Paths) > 0
}
