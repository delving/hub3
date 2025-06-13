package fragments

import (
	"regexp"
	"strings"
)

// FieldInfo contains information about detected fields in a query
type FieldInfo struct {
	HasFields bool     // Whether any fields were found
	Fields    []string // List of unique field names found
	Count     int      // Total number of field occurrences
}

// detectSearchFields analyzes a query string and returns information about fielded searches
func detectSearchFields(query string) FieldInfo {
	// Custom parsing implementation for field detection
	// This is more robust than using a single regex pattern
	
	// Handle special cases first
	if query == "" {
		return FieldInfo{
			HasFields: false,
			Fields:    []string{},
			Count:     0,
		}
	}
	
	// Check for valid field pattern: field:value
	// A valid field name must:
	// 1. Start with a letter or underscore
	// 2. Contain only letters, numbers, dots, or underscores
	// 3. Be followed by a colon and not at the end of a string
	// 4. Not be preceded by another alphanumeric, dot or underscore
	
	// Basic check for valid field patterns
	validFieldRegex := regexp.MustCompile(`(^|[^a-zA-Z0-9_])([a-zA-Z_][a-zA-Z0-9_.]*):\S`)
	
	matches := validFieldRegex.FindAllStringSubmatch(query, -1)
	
	if len(matches) == 0 {
		return FieldInfo{
			HasFields: false,
			Fields:    []string{},
			Count:     0,
		}
	}
	
	// Extract unique field names
	fieldSet := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 2 {
			fieldName := match[2] // Field name is in the second capture group
			fieldSet[fieldName] = true
		}
	}
	
	// Handle specific edge cases that should not match
	// Check for "text with colon: but no field"
	invalidColonRegex := regexp.MustCompile(`\s\w+:(\s|$)`)
	invalidMatches := invalidColonRegex.FindAllString(query, -1)
	
	// If we only have invalid matches (like "text: " or "colon: "), return no fields
	if len(invalidMatches) > 0 && len(fieldSet) == 1 {
		for field := range fieldSet {
			for _, invalid := range invalidMatches {
				invalid = strings.TrimSpace(invalid)
				invalid = strings.TrimSuffix(invalid, ":")
				if field == invalid {
					return FieldInfo{
						HasFields: false,
						Fields:    []string{},
						Count:     0,
					}
				}
			}
		}
	}
	
	// Convert to slice
	fields := make([]string, 0, len(fieldSet))
	for field := range fieldSet {
		fields = append(fields, field)
	}
	
	return FieldInfo{
		HasFields: true,
		Fields:    fields,
		Count:     len(matches),
	}
}

// hasSearchFields is a simple boolean check
func hasSearchFields(query string) bool {
	info := detectSearchFields(query)
	return info.HasFields
}

// getSearchFields returns just the list of field names
func getSearchFields(query string) []string {
	info := detectSearchFields(query)
	return info.Fields
}

// shouldRemoveFieldsParam returns true if the query contains fielded searches
// and the Elasticsearch fields parameter should be removed to allow those
// fielded searches to work properly
func shouldRemoveFieldsParam(query string) bool {
	return hasSearchFields(query)
}
