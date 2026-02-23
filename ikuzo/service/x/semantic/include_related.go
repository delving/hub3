package semantic

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	domainSemantic "github.com/delving/hub3/ikuzo/domain/semantic"
)

// relatedItemsProvider finds similar documents using More-Like-This.
type relatedItemsProvider struct {
	store domainSemantic.SimilarStore
}

// NewRelatedItemsProvider creates a new related items provider.
// It requires a SimilarStore implementation.
func NewRelatedItemsProvider(store domainSemantic.SimilarStore) domainSemantic.IncludeProvider {
	return &relatedItemsProvider{store: store}
}

func (p *relatedItemsProvider) Name() string {
	return "relatedItems"
}

func (p *relatedItemsProvider) Provide(ctx context.Context, req *domainSemantic.IncludeRequest) (any, error) {
	if p.store == nil {
		return nil, fmt.Errorf("no SimilarStore configured")
	}

	opts := domainSemantic.DefaultSimilarOptions()

	// Apply params from query string
	if countStr, ok := req.Params["count"]; ok {
		if count, err := strconv.Atoi(countStr); err == nil && count > 0 {
			if count > 20 {
				count = 20
			}
			opts.Count = count
		}
	}

	if fieldsStr, ok := req.Params["fields"]; ok && fieldsStr != "" {
		opts.Fields = strings.Split(fieldsStr, ",")
	}

	result, err := p.store.FindSimilar(ctx, req.DocumentID, opts, req.Config)
	if err != nil {
		return nil, fmt.Errorf("FindSimilar failed: %w", err)
	}

	if result == nil || len(result.Results) == 0 {
		return nil, nil
	}

	// Build Hydra Collection response
	members := make([]any, len(result.Results))
	for i, item := range result.Results {
		members[i] = item
	}

	collection := map[string]any{
		"@type":      "hydra:Collection",
		"totalItems": result.Total,
		"member":     members,
	}

	// Add similarity fields info if known
	if len(opts.Fields) > 0 {
		collection["hub3:similarityFields"] = opts.Fields
	}

	return collection, nil
}
