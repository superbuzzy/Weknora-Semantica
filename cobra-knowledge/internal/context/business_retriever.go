package context

import (
	"context"
	"fmt"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

// BusinessQuerySource is intentionally transport-neutral. MCP, REST or SQL-gateway
// adapters can implement it without leaking business-system details into the core.
type BusinessQuerySource interface {
	Query(ctx context.Context, req model.QueryRequest, step model.RetrievalStep) (model.RetrievalResult, error)
}

type BusinessDataRetriever struct{ Gateway BusinessQuerySource }

func (r *BusinessDataRetriever) Source() model.RetrievalSource { return model.SourceBusinessData }

func (r *BusinessDataRetriever) Retrieve(ctx context.Context, req model.QueryRequest, step model.RetrievalStep) (model.RetrievalResult, error) {
	if r.Gateway == nil {
		return model.RetrievalResult{Source: model.SourceBusinessData}, fmt.Errorf("business data gateway is not configured")
	}
	result, err := r.Gateway.Query(ctx, req, step)
	result.Source = model.SourceBusinessData
	return result, err
}
