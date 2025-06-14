package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/index"
)

// MoreLikeThisSearch performs a moreLikeThis query using a document ID and returns document sources
func MoreLikeThisSearch(
	orgID string,
	hubID string,
	searchLabels []string,
	filters []string,
	size int,
) ([]json.RawMessage, error) {
	// Create a client
	client := index.ESClient()

	// Build the searchLabel fields array for more_like_this
	searchLabelFields := make([]string, len(searchLabels))
	for i, label := range searchLabels {
		searchLabel := label
		if !strings.HasPrefix(searchLabel, "fields.") {
			searchLabel = fmt.Sprintf("fields.%s", label)
		}
		searchLabelFields[i] = searchLabel
	}

	// Create the more_like_this part of the query which is common to both cases
	mltQuery := fmt.Sprintf(`{
        "fields": %s,
        "like": [
            {
                "_index": "%s",
                "_id": "%s"
            }
        ],
        "min_term_freq": 1,
        "max_query_terms": 15,
        "min_doc_freq": 1
    }`,
		marshalStringSlice(searchLabelFields),
		c.Config.ElasticSearch.GetIndexName(orgID),
		hubID)
	
	var queryJSON string
	
	// Process filters if we have any
	if len(filters) > 0 {
		// Process filters and add "fields." prefix if needed
		processedFilters := make([]string, len(filters))
		for i, filter := range filters {
			// Skip adding prefix if it's already a meta or tree field
			if strings.HasPrefix(filter, "meta.") || strings.HasPrefix(filter, "tree.") {
				processedFilters[i] = filter
			} else if !strings.HasPrefix(filter, "fields.") {
				// Add fields prefix if not already there
				processedFilters[i] = fmt.Sprintf("fields.%s", filter)
			} else {
				processedFilters[i] = filter
			}
		}
		
		// Create filter terms JSON
		var filterTerms []string
		for _, filter := range processedFilters {
			// Split the filter into field:value
			parts := strings.SplitN(filter, ":", 2)
			if len(parts) == 2 {
				field := parts[0]
				value := parts[1]
				
				// Escape any quotes in the value
				value = escapeJSON(value)
				
				// Create the term filter JSON
				filterTerm := fmt.Sprintf(`{"term":{"%s":"%s"}}`, field, value)
				filterTerms = append(filterTerms, filterTerm)
			}
		}
		
		// Join filter terms with commas
		filtersJSON := strings.Join(filterTerms, ",")
		
		// Build the combined bool query with must (more_like_this) and should (OR) for filters
		queryJSON = fmt.Sprintf(`{
    "query": {
        "bool": {
            "must": {
                "more_like_this": %s
            },
            "should": [%s],
            "minimum_should_match": 1
        }
    },
    "size": %d
}`, mltQuery, filtersJSON, size)
	} else {
		// No filters, just use the more_like_this query directly
		queryJSON = fmt.Sprintf(`{
    "query": {
        "more_like_this": %s
    },
    "size": %d
}`, mltQuery, size)
	}
	
	rawQuery := queryJSON

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	slog.Info("mlt", "rawQuery", rawQuery, "docID", hubID)
	println(rawQuery)

	// Perform the search using raw query
	res, err := client.Search().
		Index(c.Config.ElasticSearch.GetIndexName(orgID)).
		Source(rawQuery).
		Pretty(true).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("search error: %w", err)
	}

	slog.Info("mlt query", "hits", len(res.Hits.Hits), "totalHits", res.Hits.TotalHits)

	// Process results - collect document sources
	var sources []json.RawMessage
	count := 0
	for _, hit := range res.Hits.Hits {
		slog.Info("mlt query sources", "doc", hit.Id)
		if count >= size {
			break
		}
		sources = append(sources, hit.Source)
		count++
	}

	return sources, nil
}

// Helper function to properly escape JSON strings
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// Helper function to marshal string slice to JSON
func marshalStringSlice(slice []string) string {
	bytes, err := json.Marshal(slice)
	if err != nil {
		return "[]"
	}
	return string(bytes)
}