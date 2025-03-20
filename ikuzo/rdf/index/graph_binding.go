package index

import (
	"fmt"
	"log/slog"
	"strings"
)

// PathMap represents a mapping of field names to query paths
type PathMap map[string]string

// Bind executes multiple path queries according to a PathMap and returns a map of matched entries
func (g *Graph) Bind(pathMap PathMap) (map[string][]*Entry, error) {
	// Make sure the graph is inlined
	if len(g.roots) == 0 {
		if err := g.Inline(); err != nil {
			return nil, fmt.Errorf("failed to inline graph for binding: %w", err)
		}
	}

	// Initialize result map
	result := make(map[string][]*Entry)

	// Process each field and path in the map
	for field, path := range pathMap {
		// Skip empty paths
		if path == "" {
			continue
		}

		// Execute the query for this path
		queryResult, err := g.Query(path)
		if err != nil {
			return nil, fmt.Errorf("error querying path for field %s: %w", field, err)
		}

		slog.Info("query result", "result", queryResult, "path", path)

		// If there are entries, add them to the result map
		if len(queryResult.Entries) > 0 {
			result[field] = queryResult.Entries
		} else if len(queryResult.Resources) > 0 {
			// If no entries but resources exist, try to extract labels or other identifiable values
			var entries []*Entry
			for _, resource := range queryResult.Resources {
				// Try to get a label for the resource
				label, lang := resource.GetLabel()
				if label != "" {
					// Create a new entry with the label value
					entry := &Entry{
						Value:     label,
						Language:  lang,
						EntryType: Literal,
					}
					entries = append(entries, entry)
				} else {
					// If no label, use resource ID as a fallback
					entry := &Entry{
						ID:        resource.ID,
						EntryType: ResourceType,
					}
					entries = append(entries, entry)
				}
			}
			result[field] = entries
		}
	}

	return result, nil
}

// normalizePath standardizes path syntax by replacing ">" with "->" when needed
func normalizePath(path string) string {
	// Split the path to handle parts separately (resource path, filter, value path)
	parts := strings.Split(path, " | ")

	for i, part := range parts {
		// Skip filter parts with "=" as they don't need the arrow normalization
		if strings.Contains(part, "=") {
			continue
		}

		// Replace single ">" with "->" but only if not already part of "->"
		var normalizedPart string
		for j := 0; j < len(part); j++ {
			if part[j] == '>' && (j == 0 || part[j-1] != '-') {
				normalizedPart += "->"
			} else if j > 0 && part[j-1] == '-' && part[j] == '>' {
				normalizedPart += ">"
			} else {
				normalizedPart += string(part[j])
			}
		}
		parts[i] = normalizedPart
	}

	return strings.Join(parts, " | ")
}
