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
		result model.RetrievalResult
		err    error
	}
	ch := make(chan item, len(plan.Steps))
	var wg sync.WaitGroup
	for _, step := range plan.Steps {
		r, ok := s.Retrievers[step.Source]
		if !ok {
			ch <- item{result: model.RetrievalResult{Source: step.Source, Gaps: []string{"retriever not configured: " + string(step.Source)}}}
			continue
		}
		wg.Add(1)
		go func(r Retriever, step model.RetrievalStep) {
			defer wg.Done()
			res, err := r.Retrieve(ctx, req, step)
			ch <- item{result: res, err: err}
		}(r, step)
	}
	wg.Wait()
	close(ch)
	var results []model.RetrievalResult
	var assertions []model.Assertion
	for x := range ch {
		if x.err != nil {
			results = append(results, model.RetrievalResult{Source: x.result.Source, Gaps: []string{x.err.Error()}})
			continue
		}
		results = append(results, x.result)
		assertions = append(assertions, x.result.Assertions...)
	}
	qt := time.Now().UTC()
	if req.At != nil {
		qt = req.At.UTC()
	}
	arbitration := s.Arbiter.Arbitrate(assertions, qt)
	return s.Assembler.Assemble(plan, results, arbitration), nil
}
