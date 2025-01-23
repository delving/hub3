package index

type PathFilter struct {
	ID        string
	Type      string
	Predicate string
	// Path      string
	// NSPath    string
	// Filter    string
}

// Query traverses resources following the given path and returns matching entries
func Query(resources []*Resource, path []PathFilter) []*Entry {
	if len(path) == 0 || len(resources) == 0 {
		return nil
	}

	var matches []*Entry
	currentFilter := path[0]

	// For each resource at current level
	for _, resource := range resources {
		// Check if resource matches current filter's type and ID if specified
		if !matchesResource(resource, currentFilter) {
			continue
		}

		// If this is the last filter in path, collect matching entries
		if len(path) == 1 {
			matches = append(matches, findMatchingEntries(resource, currentFilter.Predicate)...)
			continue
		}

		// Find all entries matching the current predicate that point to other resources
		matchingEntries := findMatchingEntries(resource, currentFilter.Predicate)

		// Collect target resource IDs from matching entries
		var targetIDs []string
		for _, entry := range matchingEntries {
			if entry.EntryType == ResourceType || entry.EntryType == Bnode {
				targetIDs = append(targetIDs, entry.ID)
			}
		}

		// Filter resources for the next level based on collected IDs
		var nextResources []*Resource
		for _, res := range resources {
			for _, targetID := range targetIDs {
				if res.ID == targetID {
					nextResources = append(nextResources, res)
					break
				}
			}
		}

		// Recursively query the next level with remaining path
		matches = append(matches, Query(nextResources, path[1:])...)
	}

	return matches
}

// matchesResource checks if a resource matches the type and ID filters
func matchesResource(resource *Resource, filter PathFilter) bool {
	// If filter specifies a Type, check if resource has it
	if filter.Type != "" {
		hasType := false
		for _, t := range resource.Types {
			if t == filter.Type {
				hasType = true
				break
			}
		}
		if !hasType {
			return false
		}
	}

	// If filter specifies an ID, check if resource matches
	if filter.ID != "" && resource.ID != filter.ID {
		return false
	}

	return true
}

// findMatchingEntries returns entries that match the given predicate
func findMatchingEntries(resource *Resource, predicate string) []*Entry {
	var matches []*Entry
	for _, entry := range resource.Entries {
		if entry.Predicate == predicate {
			matches = append(matches, entry)
		}
	}
	return matches
}

// type PathQuery struct {
// 	Paths []PathFilter
// }
//
// type PathResult struct {
// 	triples []*rdf.Triple
// }
//
// func (g *Graph) Query(q PathQuery) (*PathResult, error) {
// 	res := &PathRekult{}
// 	for _, p := range q.Paths {
//
//
// 	}
// 	return res, nil
// }

// func Query(path string) (*Result, error) {
// 	return nil, nil
// }
//
// type Result struct {
// 	path string
// 	res  gjson.Result
// }
//
// func (res *Result) Exists() bool {
// 	return res.res.Exists()
// }
//
// func (res *Result) First() string {
// 	return ""
// }
//
// func (res *Result) Array() []string {
// 	return []string{}
// }
//
// func (res *Result) Raw() gjson.Result {
// 	return res.res
// }
