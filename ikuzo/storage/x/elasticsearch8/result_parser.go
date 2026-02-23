package elasticsearch8

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/delving/hub3/ikuzo/domain/semantic"
)

// esSearchResponse represents the top-level Elasticsearch search response.
type esSearchResponse struct {
	Took int                        `json:"took"`
	Hits esHits                     `json:"hits"`
	Aggs map[string]json.RawMessage `json:"aggregations"`
}

// esHits represents the hits section of an Elasticsearch response.
type esHits struct {
	Total esTotal `json:"total"`
	Hits  []esHit `json:"hits"`
}

// esTotal represents the total hits count.
type esTotal struct {
	Value    int64  `json:"value"`
	Relation string `json:"relation"`
}

// esHit represents a single hit in the search results.
type esHit struct {
	ID        string                         `json:"_id"`
	Score     *float64                       `json:"_score"`
	Source    json.RawMessage                `json:"_source"`
	InnerHits map[string]esInnerHitsWrapper  `json:"inner_hits"`
}

// esInnerHitsWrapper wraps inner hits for nested queries.
type esInnerHitsWrapper struct {
	Hits esHits `json:"hits"`
}

// esGetResponse represents the Elasticsearch GET document response.
type esGetResponse struct {
	ID     string          `json:"_id"`
	Found  bool            `json:"found"`
	Source json.RawMessage `json:"_source"`
}

// esTermsBucket represents a single bucket in a terms aggregation.
type esTermsBucket struct {
	Key      string `json:"key"`
	DocCount int64  `json:"doc_count"`
}

// esTermsAgg represents a terms aggregation result.
type esTermsAgg struct {
	Buckets          []esTermsBucket `json:"buckets"`
	SumOtherDocCount int64           `json:"sum_other_doc_count"`
}

// esNestedAgg represents a nested aggregation with filtered sub-aggregation.
type esNestedAgg struct {
	DocCount int64           `json:"doc_count"`
	Filtered json.RawMessage `json:"filtered"`
}

// esFilteredAgg represents the filtered sub-aggregation inside a nested agg.
type esFilteredAgg struct {
	DocCount int64           `json:"doc_count"`
	Values   json.RawMessage `json:"values"`
}

// ResultParser parses Elasticsearch responses into semantic domain types.
type ResultParser struct{}

// ParseSearchResponse parses a full Elasticsearch search response into a SearchResult.
func (rp *ResultParser) ParseSearchResponse(data []byte) (*semantic.SearchResult, error) {
	var esResp esSearchResponse
	if err := json.Unmarshal(data, &esResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ES search response: %w", err)
	}

	result := &semantic.SearchResult{
		Total:     esResp.Hits.Total.Value,
		Results:   make([]map[string]any, 0, len(esResp.Hits.Hits)),
		ResultIDs: make([]string, 0, len(esResp.Hits.Hits)),
		Facets:    []semantic.FacetResult{},
		Took:      time.Duration(esResp.Took) * time.Millisecond,
		Metadata: map[string]any{
			"elasticsearch_took_ms": esResp.Took,
		},
	}

	for _, hit := range esResp.Hits.Hits {
		doc, err := rp.parseHit(hit)
		if err != nil {
			// Skip malformed documents rather than failing the whole response.
			continue
		}

		result.Results = append(result.Results, doc)
		result.ResultIDs = append(result.ResultIDs, hit.ID)
	}

	return result, nil
}

// parseHit converts a single ES hit into a map, ensuring @id is set.
func (rp *ResultParser) parseHit(hit esHit) (map[string]any, error) {
	if len(hit.Source) == 0 {
		return nil, fmt.Errorf("empty _source for hit %s", hit.ID)
	}

	var doc map[string]any
	if err := json.Unmarshal(hit.Source, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal _source for hit %s: %w", hit.ID, err)
	}

	// Set @id from _id if not present in source.
	if _, ok := doc["@id"]; !ok {
		doc["@id"] = hit.ID
	}

	return doc, nil
}

// ParseFacets parses aggregation results into FacetResult slices.
// It matches each FacetRequest by field name against the aggregations map.
func (rp *ResultParser) ParseFacets(
	aggs map[string]json.RawMessage,
	facets []semantic.FacetRequest,
) []semantic.FacetResult {
	if len(aggs) == 0 || len(facets) == 0 {
		return []semantic.FacetResult{}
	}

	results := make([]semantic.FacetResult, 0, len(facets))

	for _, facet := range facets {
		raw, ok := aggs[facet.Field]
		if !ok {
			continue
		}

		fr := rp.parseSingleFacet(facet.Field, raw)
		results = append(results, fr)
	}

	return results
}

// parseSingleFacet parses a single aggregation result.
// It handles both flat terms aggregations and nested aggregations.
func (rp *ResultParser) parseSingleFacet(field string, raw json.RawMessage) semantic.FacetResult {
	fr := semantic.FacetResult{
		Field:     field,
		FacetType: "enum",
		Values:    []semantic.FacetValueResult{},
	}

	// First, try to parse as a flat terms aggregation (has "buckets" at top level).
	var termsAgg esTermsAgg
	if err := json.Unmarshal(raw, &termsAgg); err == nil && len(termsAgg.Buckets) > 0 {
		fr.SumOther = termsAgg.SumOtherDocCount
		fr.Values = bucketsToValues(termsAgg.Buckets)
		fr.TotalValues = int64(len(fr.Values))

		return fr
	}

	// Try nested aggregation: look for "filtered" -> "values" -> buckets.
	var nestedAgg esNestedAgg
	if err := json.Unmarshal(raw, &nestedAgg); err == nil && len(nestedAgg.Filtered) > 0 {
		var filteredAgg esFilteredAgg
		if err := json.Unmarshal(nestedAgg.Filtered, &filteredAgg); err == nil && len(filteredAgg.Values) > 0 {
			var innerTerms esTermsAgg
			if err := json.Unmarshal(filteredAgg.Values, &innerTerms); err == nil {
				fr.SumOther = innerTerms.SumOtherDocCount
				fr.Values = bucketsToValues(innerTerms.Buckets)
				fr.TotalValues = int64(len(fr.Values))

				return fr
			}
		}
	}

	return fr
}

// bucketsToValues converts ES term buckets to FacetValueResult slices.
func bucketsToValues(buckets []esTermsBucket) []semantic.FacetValueResult {
	values := make([]semantic.FacetValueResult, 0, len(buckets))
	for _, b := range buckets {
		values = append(values, semantic.FacetValueResult{
			Value: b.Key,
			Label: b.Key,
			Count: b.DocCount,
		})
	}

	return values
}

// ParseGetResponse parses a single-document GET response from Elasticsearch.
func (rp *ResultParser) ParseGetResponse(data []byte) (map[string]any, error) {
	var getResp esGetResponse
	if err := json.Unmarshal(data, &getResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ES GET response: %w", err)
	}

	if !getResp.Found {
		return nil, fmt.Errorf("document not found")
	}

	if len(getResp.Source) == 0 {
		return nil, fmt.Errorf("empty _source in GET response for %s", getResp.ID)
	}

	var doc map[string]any
	if err := json.Unmarshal(getResp.Source, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal _source from GET response: %w", err)
	}

	// Set @id from _id if not present in source.
	if _, ok := doc["@id"]; !ok {
		doc["@id"] = getResp.ID
	}

	return doc, nil
}
