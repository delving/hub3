package index

import (
	"errors"
	"fmt"
	"strings"

	"github.com/delving/hub3/ikuzo/rdf"
)

// PathQuery defines a complex query with resource path, optional filters, and value path
type PathQuery struct {
	Resource string     // The main resource path using the enhanced syntax
	Filter   FilterPath // Optional filter conditions
	Value    string     // Path to the value to extract from matching resources
}

// FilterPath represents a filter condition with a path and allowed values
type FilterPath struct {
	Path   string   // Path to filter by, using the enhanced syntax
	Values []string // Values that the path must match
}

// Empty returns true if the filter path is empty
func (fp *FilterPath) Empty() bool {
	return fp.Path == "" && len(fp.Values) == 0
}

// PathEntry represents a single step in a path query
type PathEntry struct {
	Class     string
	Predicate string
	IsLeaf    bool // IsLeaf indicates that this is the last entry in the path
}

func (pe *PathEntry) String() string {
	return fmt.Sprintf("[%s]%s", pe.Class, pe.Predicate)
}

func (pe *PathEntry) ExpandNamespace() error {
	var err error
	mgr := rdf.DefaultNamespaceManager
	if pe.Predicate != "" {
		pe.Predicate, err = mgr.Expand(pe.Predicate)
		if err != nil {
			return fmt.Errorf("unable to expand predicate %s: %w", pe.Predicate, err)
		}
	}
	if pe.Class != "" {
		pe.Class, err = mgr.Expand(pe.Class)
		if err != nil {
			return fmt.Errorf("unable to expand class %s: %w", pe.Class, err)
		}
	}

	return nil
}

// ParsePathString parses a string into a PathQuery structure.
// The path string follows the syntax:
// - Resource path: [class]predicate->[class]predicate...
// - Optional filter: | [class]predicate=value1 | [class]predicate=value2...
// - Optional value path: | [class]predicate...
//
// Empty class brackets ([]) represent a wildcard for node traversal.
//
// Examples:
// - Simple path: "[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title]crm:P190_has_symbolic_content"
// - With filter: "[class]pred | [class]pred=value"
// - With filter and value: "[class]pred | [class]pred=value | [class]valuePathPred"
// - With wildcard: "[]crm:P102_has_title"
//
// See the accompanying documentation for detailed syntax rules.
func ParsePathString(path string) (*PathQuery, error) {
	if path == "" {
		return nil, nil
	}

	// Split by pipe delimiter
	parts := strings.Split(path, " | ")

	// Validate the resource path has balanced brackets
	if !hasBalancedBrackets(parts[0]) {
		return nil, errors.New("unclosed bracket in resource path")
	}

	// Initialize the PathQuery
	pathQuery := &PathQuery{
		Resource: parts[0],
		Filter:   FilterPath{},
	}

	// Process filters and value path if they exist
	for i := 1; i < len(parts); i++ {
		part := parts[i]

		// Check if it's a filter (contains '=')
		if idx := strings.Index(part, "="); idx != -1 {
			filterPath := part[:idx]
			filterValue := part[idx+1:]

			// Validate the filter path has balanced brackets
			if !hasBalancedBrackets(filterPath) {
				return nil, errors.New("unclosed bracket in filter path")
			}

			// If this is the first filter value, set the filter path
			if pathQuery.Filter.Path == "" {
				pathQuery.Filter.Path = filterPath
			} else if pathQuery.Filter.Path != filterPath {
				// If we already have a different filter path, this is an error
				return nil, errors.New("multiple different filter paths not supported")
			}

			// Add the filter value
			pathQuery.Filter.Values = append(pathQuery.Filter.Values, filterValue)
		} else {
			// If there's no '=', it's a value path
			// Validate the value path has balanced brackets
			if !hasBalancedBrackets(part) {
				return nil, errors.New("unclosed bracket in value path")
			}

			// Only set it if we haven't already (to handle edge cases)
			if pathQuery.Value == "" {
				pathQuery.Value = part
			} else {
				return nil, errors.New("multiple value paths not supported")
			}
		}
	}

	if !strings.HasSuffix(pathQuery.Resource, "]") && pathQuery.Value == "" {
		parts := strings.Split(pathQuery.Resource, "->")
		if len(parts) > 1 {
			pathQuery.Value = parts[len(parts)-1]
		}
	}

	return pathQuery, nil
}

// String converts a PathQuery back to its string representation
func (query *PathQuery) String() string {
	if query == nil || query.Resource == "" {
		return ""
	}

	var builder strings.Builder

	// Add the resource path
	builder.WriteString(query.Resource)

	// Add filter if present
	if !query.Filter.Empty() {
		for _, value := range query.Filter.Values {
			builder.WriteString(" | ")
			builder.WriteString(query.Filter.Path)
			builder.WriteString("=")
			builder.WriteString(value)
		}
	}

	// Add value path if present
	if query.Value != "" {
		builder.WriteString(" | ")
		builder.WriteString(query.Value)
	}

	return builder.String()
}

// hasBalancedBrackets checks if all opening square brackets have matching closing brackets
func hasBalancedBrackets(s string) bool {
	bracketCount := 0
	for _, char := range s {
		if char == '[' {
			bracketCount++
		} else if char == ']' {
			bracketCount--
			if bracketCount < 0 {
				return false // More closing than opening brackets
			}
		}
	}
	return bracketCount == 0 // All brackets should be matched
}

// ParsePath parses a path string (like "[class]predicate") into a PathEntry structure
// This helper function can be used if you need to break down path components.
// If expanded is true, the predicate and class will be expanded using the default namespace manager.
func ParsePath(path string, expanded bool) ([]PathEntry, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path is not allowed")
	}

	// First check if the path has balanced brackets
	if !hasBalancedBrackets(path) {
		return nil, errors.New("unclosed bracket in path")
	}

	var entries []PathEntry
	remaining := path

	for len(remaining) > 0 {
		entry := PathEntry{}

		// Extract class if present (in square brackets)
		if strings.HasPrefix(remaining, "[") {
			endIdx := strings.Index(remaining, "]")
			if endIdx == -1 {
				return nil, errors.New("unclosed bracket in path")
			}

			// +1 to skip the opening bracket
			entry.Class = remaining[1:endIdx]
			remaining = remaining[endIdx+1:]
		}

		// Extract predicate if present
		if len(remaining) > 0 {
			// Check if there's an arrow or another class
			arrowIdx := strings.Index(remaining, "->")
			nextClassIdx := strings.Index(remaining, "[")

			var predicateEnd int

			if arrowIdx != -1 && (nextClassIdx == -1 || arrowIdx < nextClassIdx) {
				// Predicate ends at arrow
				predicateEnd = arrowIdx
				entry.IsLeaf = false
			} else if nextClassIdx != -1 {
				// Predicate ends at next class
				predicateEnd = nextClassIdx
				entry.IsLeaf = false
			} else {
				// Predicate goes to the end
				predicateEnd = len(remaining)
				entry.IsLeaf = true
			}

			entry.Predicate = remaining[:predicateEnd]

			// Move past the predicate
			remaining = remaining[predicateEnd:]

			// Skip arrow if present
			if strings.HasPrefix(remaining, "->") {
				remaining = remaining[2:]
			}
		}

		entries = append(entries, entry)
	}

	if expanded {
		for i := range entries {
			if err := entries[i].ExpandNamespace(); err != nil {
				return nil, err
			}
		}
	}

	return entries, nil
}

// ParseCsvPathMap parses a CSV-style mapping into a PathQuery map
// Format: TargetField,PathQuery,Source,Notes
func ParseCsvPathMap(csvLines []string) (PathMap, error) {
	pathMap := make(PathMap)

	for i, line := range csvLines {
		// Skip header or empty lines
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		// Split the CSV line
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			return nil, fmt.Errorf("line %d: insufficient columns (need at least field and path)", i+1)
		}

		field := strings.TrimSpace(parts[0])
		pathStr := strings.Trim(strings.TrimSpace(parts[1]), "\"")

		pathMap[field] = pathStr
	}

	return pathMap, nil
}
