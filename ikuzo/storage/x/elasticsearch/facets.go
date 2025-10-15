package elasticsearch

import (
	"fmt"

	"github.com/delving/hub3/ikuzo/domain/semantic"
	"github.com/olivere/elastic/v7"
)

// buildFacetAggregation builds an Elasticsearch aggregation for a facet request.
func (s *SemanticStore) buildFacetAggregation(
	facet semantic.FacetRequest,
	config *semantic.ResourceConfig,
) elastic.Aggregation {
	// Get facet config to determine aggregation type
	facetConfig, ok := config.GetFacetConfig(facet.Field)
	if !ok {
		s.log.Warn().Str("field", facet.Field).Msg("facet not configured")
		return nil
	}

	switch facetConfig.FacetType {
	case "enum":
		return s.buildEnumFacetAggregation(facet, facetConfig)

	case "range":
		return s.buildRangeFacetAggregation(facet, facetConfig)

	case "date":
		return s.buildDateFacetAggregation(facet, facetConfig)

	case "boolean":
		return s.buildBooleanFacetAggregation(facet, facetConfig)

	default:
		s.log.Warn().
			Str("field", facet.Field).
			Str("type", facetConfig.FacetType).
			Msg("unknown facet type")
		return nil
	}
}

// buildEnumFacetAggregation builds a terms aggregation for enum facets.
func (s *SemanticStore) buildEnumFacetAggregation(
	facet semantic.FacetRequest,
	config *semantic.FacetConfig,
) elastic.Aggregation {
	size := config.Size
	if facet.Limit > 0 && facet.Limit < size {
		size = facet.Limit
	}

	// Translate field name: dc.creator -> dc_creator
	esField := translateFieldName(facet.Field)

	agg := elastic.NewTermsAggregation().
		Field(esField + ".keyword").
		Size(size).
		MinDocCount(int(config.MinDocCount))

	// Apply sort
	switch config.SortBy {
	case "count":
		if config.SortOrder == "asc" {
			agg = agg.OrderByCountAsc()
		} else {
			agg = agg.OrderByCountDesc()
		}
	case "term":
		if config.SortOrder == "asc" {
			agg = agg.OrderByKeyAsc()
		} else {
			agg = agg.OrderByKeyDesc()
		}
	}

	// Include missing values if configured
	if config.Missing {
		agg = agg.Missing("_missing_")
	}

	return agg
}

// buildRangeFacetAggregation builds a range aggregation for range facets.
func (s *SemanticStore) buildRangeFacetAggregation(
	facet semantic.FacetRequest,
	config *semantic.FacetConfig,
) elastic.Aggregation {
	// Translate field name: dc.creator -> dc_creator
	esField := translateFieldName(facet.Field)
	agg := elastic.NewRangeAggregation().Field(esField)

	for _, r := range config.Ranges {
		if r.From > 0 && r.To > 0 {
			agg = agg.AddRange(r.From, r.To)
		} else if r.From > 0 {
			agg = agg.AddUnboundedFrom(r.From)
		} else if r.To > 0 {
			agg = agg.AddUnboundedTo(r.To)
		}
	}

	return agg
}

// buildDateFacetAggregation builds a date histogram aggregation for date facets.
func (s *SemanticStore) buildDateFacetAggregation(
	facet semantic.FacetRequest,
	config *semantic.FacetConfig,
) elastic.Aggregation {
	interval := "year"
	if len(config.DateIntervals) > 0 {
		interval = config.DateIntervals[0].Interval
	}

	// Translate field name: dc.creator -> dc_creator
	esField := translateFieldName(facet.Field)

	agg := elastic.NewDateHistogramAggregation().
		Field(esField).
		CalendarInterval(interval).
		MinDocCount(config.MinDocCount)

	if config.SortOrder == "desc" {
		agg = agg.OrderByKeyDesc()
	} else {
		agg = agg.OrderByKeyAsc()
	}

	return agg
}

// buildBooleanFacetAggregation builds a terms aggregation for boolean facets.
func (s *SemanticStore) buildBooleanFacetAggregation(
	facet semantic.FacetRequest,
	config *semantic.FacetConfig,
) elastic.Aggregation {
	// Translate field name: dc.creator -> dc_creator
	esField := translateFieldName(facet.Field)
	return elastic.NewTermsAggregation().
		Field(esField).
		Size(2). // Only true/false
		MinDocCount(int(config.MinDocCount))
}

// extractFacets extracts facet results from Elasticsearch aggregations.
func (s *SemanticStore) extractFacets(
	searchResult *elastic.SearchResult,
	facets []semantic.FacetRequest,
	config *semantic.ResourceConfig,
) []semantic.FacetResult {
	results := make([]semantic.FacetResult, 0, len(facets))

	for _, facet := range facets {
		facetConfig, ok := config.GetFacetConfig(facet.Field)
		if !ok {
			continue
		}

		agg, found := searchResult.Aggregations.Terms(facet.Field)
		if !found {
			// Try other aggregation types
			if facetConfig.FacetType == "range" {
				result := s.extractRangeFacet(searchResult, facet, facetConfig)
				if result != nil {
					results = append(results, *result)
				}
				continue
			}

			if facetConfig.FacetType == "date" {
				result := s.extractDateFacet(searchResult, facet, facetConfig)
				if result != nil {
					results = append(results, *result)
				}
				continue
			}

			s.log.Warn().Str("field", facet.Field).Msg("aggregation not found")
			continue
		}

		result := semantic.FacetResult{
			Field:       facet.Field,
			FacetType:   facetConfig.FacetType,
			TotalValues: int64(len(agg.Buckets)),
			Values:      make([]semantic.FacetValueResult, 0, len(agg.Buckets)),
		}

		for _, bucket := range agg.Buckets {
			value := semantic.FacetValueResult{
				Value: fmt.Sprintf("%v", bucket.Key),
				Count: bucket.DocCount,
			}

			// Use label from config if available
			value.Label = value.Value

			result.Values = append(result.Values, value)
		}

		// Add sum of other document counts
		result.SumOther = agg.SumOfOtherDocCount

		results = append(results, result)
	}

	return results
}

// extractRangeFacet extracts range facet results.
func (s *SemanticStore) extractRangeFacet(
	searchResult *elastic.SearchResult,
	facet semantic.FacetRequest,
	config *semantic.FacetConfig,
) *semantic.FacetResult {
	agg, found := searchResult.Aggregations.Range(facet.Field)
	if !found {
		return nil
	}

	result := &semantic.FacetResult{
		Field:       facet.Field,
		FacetType:   "range",
		TotalValues: int64(len(agg.Buckets)),
		Values:      make([]semantic.FacetValueResult, 0, len(agg.Buckets)),
	}

	// Match buckets with configured ranges
	for i, bucket := range agg.Buckets {
		value := semantic.FacetValueResult{
			Count: bucket.DocCount,
		}

		// Try to find matching range config
		if i < len(config.Ranges) {
			rangeConfig := config.Ranges[i]
			value.Value = rangeConfig.Key
			if rangeConfig.Label != "" {
				value.Label = rangeConfig.Label
			} else {
				value.Label = rangeConfig.Key
			}
		} else {
			// Fallback to bucket key
			value.Value = bucket.Key
			value.Label = bucket.Key
		}

		result.Values = append(result.Values, value)
	}

	return result
}

// extractDateFacet extracts date histogram facet results.
func (s *SemanticStore) extractDateFacet(
	searchResult *elastic.SearchResult,
	facet semantic.FacetRequest,
	config *semantic.FacetConfig,
) *semantic.FacetResult {
	agg, found := searchResult.Aggregations.DateHistogram(facet.Field)
	if !found {
		return nil
	}

	result := &semantic.FacetResult{
		Field:       facet.Field,
		FacetType:   "date",
		TotalValues: int64(len(agg.Buckets)),
		Values:      make([]semantic.FacetValueResult, 0, len(agg.Buckets)),
	}

	for _, bucket := range agg.Buckets {
		keyStr := ""
		if bucket.KeyAsString != nil {
			keyStr = *bucket.KeyAsString
		}

		value := semantic.FacetValueResult{
			Value: keyStr,
			Label: keyStr,
			Count: bucket.DocCount,
		}

		result.Values = append(result.Values, value)
	}

	return result
}
