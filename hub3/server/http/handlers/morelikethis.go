package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/index"
)

// MoreLikeThisNestedSearch performs a moreLikeThis query and returns just the inner hit sources
func MoreLikeThisNestedSearch(
	orgID string,
	hubID string,
	searchLabels []string,
	likeText string,
	size int,
) ([]json.RawMessage, error) {
	// Create a client
	client := index.ESClient()

	rawQuery := fmt.Sprintf(`{
    "query": {
        "bool": {
            "must": [
                {
                    "nested": {
                        "path": "resources.entries",
                        "query": {
                            "bool": {
                                "filter": {
                                    "terms": {
                                        "resources.entries.searchLabel": %s
                                    }
                                },
                                "must": {
                                    "more_like_this": {
                                        "fields": ["resources.entries.@value"],
                                        "like": %s,
                                        "min_term_freq": 1,
                                        "max_query_terms": 25,
                                        "min_doc_freq": 1
                                    }
                                }
                            }
                        },
                        "inner_hits": {
                            "_source": true,
                            "size": 5
                        },
                        "score_mode": "avg"
                    }
                }
            ],
            "must_not": [
                {
                    "term": {
                        "meta.hubID": "%s"
                    }
                }
            ]
        }
    },
    "size": %d
}`,
		marshalStringSlice(searchLabels),
		fmt.Sprintf(`"%s"`, escapeJSON(likeText)),
		escapeJSON(hubID),
		size)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Perform the search using raw query
	res, err := client.Search().
		Index(c.Config.ElasticSearch.GetIndexName(orgID)).
		Source(strings.NewReader(rawQuery)).
		Pretty(true).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("search error: %w", err)
	}

	// Process results - collect all inner hit sources
	var innerHitSources []json.RawMessage

	for _, hit := range res.Hits.Hits {
		innerHitSources = append(innerHitSources, hit.Source)
	}

	return innerHitSources, nil
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
