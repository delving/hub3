package index

// ResolveLabels populates Value, Language, ResolvedFrom, and ResolvedLevel on
// Resource-type and Bnode-type entries by looking up the linked resource's best
// label within the same graph. Entries that already have a Value are left untouched.
//
// This is a single-pass operation: for each entry pointing at another resource,
// it calls GetLabel() on the linked resource and records which predicate the
// label came from (ResolvedFrom) and the linked resource's depth (ResolvedLevel).
func (g *Graph) ResolveLabels() {
	// Build a lookup map for O(1) resource access.
	lookup := make(map[string]*Resource, len(g.Resources))
	for _, rsc := range g.Resources {
		lookup[rsc.ID] = rsc
	}

	for _, rsc := range g.Resources {
		for _, entry := range rsc.Entries {
			// Only resolve Resource and Bnode entries.
			if entry.EntryType != ResourceType && entry.EntryType != Bnode {
				continue
			}

			// Do not overwrite existing labels.
			if entry.Value != "" {
				continue
			}

			linked, ok := lookup[entry.ID]
			if !ok {
				continue
			}

			label, lang := linked.GetLabel()
			if label == "" {
				continue
			}

			entry.Value = label
			entry.Language = lang
			entry.ResolvedLevel = linked.GetLevel()

			// Determine which predicate the label was resolved from.
			for _, p := range ResourceLabelPredicates {
				entries := linked.predicate(p)
				if len(entries) > 0 && entries[0].Value == label {
					entry.ResolvedFrom = p
					break
				}
			}
		}
	}
}
