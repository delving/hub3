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

	rawQuery := fmt.Sprintf(`{
    "query": {
        "more_like_this": {
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
        }
    },
    "size": %d
}`,
		marshalStringSlice(searchLabelFields),
		c.Config.ElasticSearch.GetIndexName(orgID),
		hubID,
		size)

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
