package index

import (
	"strings"

	"github.com/OneOfOne/xxhash"
)

// ContextRef contains the path for the Referrers that point to the
// current Resource. This is used to inline or create grouped view of
// a []Resource
type ContextRef struct {
	Subject      string   `json:"Subject,omitempty"`
	SubjectClass []string `json:"SubjectClass,omitempty"`
	Predicate    string   `json:"Predicate,omitempty"`
	SearchLabel  string   `json:"SearchLabel,omitempty"`
	Level        int32    `json:"Level,omitempty"`
	ObjectID     string   `json:"ObjectID,omitempty"`
	SortKey      int32    `json:"SortKey,omitempty"`
	Label        string   `json:"Label,omitempty"`
}

// containsContext determines if a FragmentReferrerContext is already part of list.
//
// Deduplication is important to not provide false counts for context levels
func containsContext(s []*ContextRef, e *ContextRef) bool {
	for _, cr := range s {
		if cr.ObjectID == e.ObjectID && cr.Predicate == e.Predicate {
			return true
		}
	}
	return false
}

type contextHasher struct{}

// Hash computes a unique hash for ContextRef.
// It is used for sorting and determining of equivalence.
func (ch contextHasher) Hash(cr ContextRef) uint32 {
	d := xxhash.New32()
	d.WriteString(cr.Subject)
	d.WriteString(strings.Join(cr.SubjectClass, ","))
	d.WriteString(cr.Predicate)
	d.WriteString(cr.ObjectID)

	return d.Sum32()
}

// Equal determines if two ContextRefs are equal
func (ch contextHasher) Equal(a, b ContextRef) bool {
	return ch.Hash(a) == ch.Hash(b)
}
