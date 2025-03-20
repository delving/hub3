package index

import (
	"fmt"
	"log/slog"
	"strings"
)

// QueryResult represents the result of a PathQuery execution
type QueryResult struct {
	Resources []*Resource // Matching resources
	Entries   []*Entry    // Extracted entries (if a value path was provided)
}

// Query traverses the graph following the given path and returns matching results
func (g *Graph) Query(pathStr string) (*QueryResult, error) {
	if len(g.roots) == 0 {
		err := g.Inline()
		if err != nil {
			return nil, fmt.Errorf("failed to inline graph: %w", err)
		}
	}

	query, err := ParsePathString(pathStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path: %w", err)
	}

	if query == nil {
		return &QueryResult{}, nil
	}

	slog.Info("pathQuery", "query", query, "resource", query.Resource, "filter", query.Filter, "value", query.Value)

	// Find resources matching the resource path
	matchedResources, err := g.findMatchingResources(query.Resource)
	if err != nil {
		return nil, fmt.Errorf("error finding resources: %w", err)
	}

	slog.Info("Found matching resources", "resources", len(matchedResources))

	// Apply filter if present
	if !query.Filter.Empty() {
		matchedResources, err = g.applyFilter(matchedResources, query.Filter)
		if err != nil {
			return nil, fmt.Errorf("error applying filter: %w", err)
		}
	}

	result := &QueryResult{
		Resources: matchedResources,
	}

	// Extract values if value path is provided
	if query.Value != "" {
		entries, err := g.extractValues(matchedResources, query.Value)
		if err != nil {
			return nil, fmt.Errorf("error extracting values: %w", err)
		}
		result.Entries = entries
	}

	return result, nil
}

// findMatchingResources follows the resource path and returns matching resources
func (g *Graph) findMatchingResources(resourcePath string) ([]*Resource, error) {
	if resourcePath == "" {
		return nil, fmt.Errorf("resource path cannot be empty")
	}

	// Start with root resources
	currentResources := g.roots
	if len(currentResources) == 0 {
		// If roots not set, use all resources
		currentResources = g.Resources
	}

	return findMatchingResources(currentResources, resourcePath, g)
}

func findMatchingResources(resources []*Resource, resourcePath string, g *Graph) ([]*Resource, error) {
	currentResources := resources

	// Parse the path into path entries
	pathEntries, err := ParsePath(resourcePath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource path: %w", err)
	}

	// No path entries, return all root resources
	if len(pathEntries) == 0 {
		return currentResources, nil
	}

	slog.Info("Finding matching resources", "path", resourcePath, "resources", len(currentResources), "roots", len(g.roots), "pathEntries", pathEntries)

	// Process each path segment
	for i, pathEntry := range pathEntries {
		if err := pathEntry.ExpandNamespace(); err != nil {
			return nil, fmt.Errorf("failed to expand namespace: %w", err)
		}

		var nextResources []*Resource

		for _, resource := range currentResources {
			slog.Info("Checking resource", "resource", resource.ID, "resourceClass", resource.Types, "pathEntryClass", pathEntry.Class, "matchedClass", matchesClass(resource, pathEntry.Class))
			// For the first entry, check if resource type matches
			if !matchesClass(resource, pathEntry.Class) {
				continue
			}

			slog.Info("Checking resource", "resource", resource.ID, "class", pathEntry, "matchedClass", matchesClass(resource, pathEntry.Class))

			// If this is the last entry and it's marked as a leaf, we're done
			if i == len(pathEntries)-1 || pathEntry.IsLeaf {
				if matchesClass(resource, pathEntry.Class) {
					nextResources = append(nextResources, resource)
				}
				continue
			}

			// Otherwise, follow the predicate links to the next resources
			for _, entry := range resource.Entries {
				if entry.Predicate == pathEntry.Predicate {
					if entry.Inline != nil {
						// For the next path entry, verify the class if specified
						if i+1 < len(pathEntries) {
							nextEntry := pathEntries[i+1]
							if nextEntry.Class == "" || matchesClass(entry.Inline, nextEntry.Class) {
								nextResources = append(nextResources, entry.Inline)
							}
						} else {
							nextResources = append(nextResources, entry.Inline)
						}
					} else if entry.EntryType == ResourceType || entry.EntryType == Bnode {
						// Try to find the linked resource in the graph
						for _, r := range g.Resources {
							if r.ID == entry.ID {
								if i+1 < len(pathEntries) {
									nextEntry := pathEntries[i+1]
									if nextEntry.Class == "" || matchesClass(r, nextEntry.Class) {
										nextResources = append(nextResources, r)
									}
								} else {
									nextResources = append(nextResources, r)
								}
								break
							}
						}
					}
				}
			}
		}

		// If no resources found for this path segment, stop early
		if len(nextResources) == 0 {
			return nil, nil
		}

		// Continue with the next resources for the next path segment
		currentResources = nextResources

		// Skip to the next class+predicate pair
		if i+1 < len(pathEntries) && pathEntries[i+1].Class != "" {
			i++
		}
	}
	return currentResources, nil
}

// applyFilter filters resources based on the filter criteria
func (g *Graph) applyFilter(resources []*Resource, filter FilterPath) ([]*Resource, error) {
	if filter.Empty() {
		return resources, nil
	}

	// Parse the filter path
	filterEntries, err := ParsePath(filter.Path, true)
	if err != nil {
		return nil, fmt.Errorf("failed to parse filter path: %w", err)
	}

	if len(filterEntries) == 0 {
		return nil, fmt.Errorf("invalid filter path: no entries found")
	}

	var filteredResources []*Resource

	// Handle multi-segment filter paths
	if len(filterEntries) > 1 {
		// For multi-segment paths, we need to traverse the path first
		for _, resource := range resources {
			// Start with the current resource
			currentResources := []*Resource{resource}

			// Traverse through all but the last filter entry
			for i := 0; i < len(filterEntries)-1; i++ {
				var nextResources []*Resource
				for _, r := range currentResources {
					// Find matching entries based on predicate
					for _, entry := range r.Entries {
						if entry.Predicate == filterEntries[i].Predicate ||
							entry.SearchLabel == filterEntries[i].Predicate {
							// Follow link to next resource
							if entry.Inline != nil {
								if matchesClass(entry.Inline, filterEntries[i+1].Class) {
									nextResources = append(nextResources, entry.Inline)
								}
							} else if entry.EntryType == ResourceType || entry.EntryType == Bnode {
								// Find the resource by ID
								for _, linkedR := range g.Resources {
									if linkedR.ID == entry.ID &&
										matchesClass(linkedR, filterEntries[i+1].Class) {
										nextResources = append(nextResources, linkedR)
									}
								}
							}
						}
					}
				}
				// Update current resources for next iteration
				currentResources = nextResources
			}

			// Now check the last filter entry against the values
			lastEntry := filterEntries[len(filterEntries)-1]
			matched := false

			for _, r := range currentResources {
				// Check each entry against the filter
				for _, entry := range r.Entries {
					if (entry.Predicate == lastEntry.Predicate ||
						entry.SearchLabel == lastEntry.Predicate) &&
						containsValue(entry, filter.Values) {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}

			// If a match was found, add the original resource to filtered results
			if matched {
				filteredResources = append(filteredResources, resource)
			}
		}
	} else {
		// Simple single segment filter path
		lastEntry := filterEntries[len(filterEntries)-1]

		for _, resource := range resources {
			// For each resource, check entries against filter
			for _, entry := range resource.Entries {
				if (entry.Predicate == lastEntry.Predicate ||
					entry.SearchLabel == lastEntry.Predicate) &&
					containsValue(entry, filter.Values) {
					filteredResources = append(filteredResources, resource)
					break
				}

				// Also check entries in inlined resources
				if entry.Inline != nil {
					for _, inlineEntry := range entry.Inline.Entries {
						if (inlineEntry.Predicate == lastEntry.Predicate ||
							inlineEntry.SearchLabel == lastEntry.Predicate) &&
							containsValue(inlineEntry, filter.Values) {
							filteredResources = append(filteredResources, resource)
							break
						}
					}
				}
			}
		}
	}

	return filteredResources, nil
}

// extractValues extracts values from resources based on the value path
func (g *Graph) extractValues(resources []*Resource, valuePath string) ([]*Entry, error) {
	// Parse the value path
	valueEntries, err := ParsePath(valuePath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to parse value path: %w", err)
	}

	if len(valueEntries) == 0 {
		return nil, fmt.Errorf("invalid value path: no entries found")
	}

	// Get the last entry which contains the predicate to extract
	valueEntry := valueEntries[len(valueEntries)-1]

	var matchingEntries []*Entry

	// First, check if it's a direct property of the resources
	for _, resource := range resources {
		for _, entry := range resource.Entries {
			if entry.Predicate == valueEntry.Predicate || entry.SearchLabel == valueEntry.Predicate {
				entryCopy := *entry
				matchingEntries = append(matchingEntries, &entryCopy)
			}
		}
	}

	// If no entries found directly, try to resolve the path and then extract values
	if len(matchingEntries) == 0 && len(valueEntries) > 1 {
		innerResources, err := findMatchingResources(resources, valuePath, g)
		if err != nil {
			return nil, err
		}

		matchingEntries, err = g.extractValues(innerResources, valueEntry.String())
		if err != nil {
			return nil, err
		}
	}

	return matchingEntries, nil
}

// Helper functions

// matchesClass checks if a resource matches a class specification (empty class matches any resource)
func matchesClass(resource *Resource, class string) bool {
	if class == "" {
		return true // Empty class matches any resource
	}

	for _, t := range resource.Types {
		if t == class {
			return true
		}
	}

	return false
}

// containsValue checks if an entry's value is in the list of allowed values
func containsValue(entry *Entry, values []string) bool {
	if len(values) == 0 {
		return true // Empty values matches any entry
	}

	for _, v := range values {
		if entry.Value == v {
			return true
		}

		// For resource type entries, check the ID
		if entry.EntryType == ResourceType && entry.ID == v {
			return true
		}
	}

	return false
}

// traversePath traverses a path starting from a specific resource
func traversePath(resource *Resource, path string) ([]*Entry, error) {
	pathEntries, err := ParsePath(path, true)
	if err != nil {
		return nil, err
	}

	if len(pathEntries) == 0 {
		return nil, nil
	}

	// Get the first path entry
	firstEntry := pathEntries[0]

	// Check if resource matches the class
	if !matchesClass(resource, firstEntry.Class) {
		return nil, nil
	}

	// If there's only one entry, find matching entries in this resource
	if len(pathEntries) == 1 {
		var results []*Entry
		for _, entry := range resource.Entries {
			if entry.Predicate == firstEntry.Predicate || entry.SearchLabel == firstEntry.Predicate {
				results = append(results, entry)
			}
		}
		return results, nil
	}

	// If there are more entries, traverse to the next resources
	var results []*Entry
	for _, entry := range resource.Entries {
		if entry.Predicate == firstEntry.Predicate || entry.SearchLabel == firstEntry.Predicate {
			if entry.Inline != nil {
				// Recursively traverse the rest of the path
				nextPath := strings.Join([]string{pathEntries[1].Class, pathEntries[1].Predicate}, "")
				for i := 2; i < len(pathEntries); i++ {
					nextPath += "->" + strings.Join([]string{pathEntries[i].Class, pathEntries[i].Predicate}, "")
				}

				nextEntries, err := traversePath(entry.Inline, nextPath)
				if err != nil {
					return nil, err
				}
				results = append(results, nextEntries...)
			}
		}
	}

	return results, nil
}
