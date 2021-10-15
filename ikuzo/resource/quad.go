package resource

// Context is the IRI namespace for the Quad
type Context interface {
	Term
	validAsSubject()
}

// Quad represents a RDF Quad; a Triple plus the context in which it occurs.
type Quad struct {
	*Triple
	ctx Context
}

func NewQuad(triple *Triple, ctx Context) (*Quad, error) {
	return &Quad{Triple: triple, ctx: ctx}, nil
}

// Equal tests if other Quad is identical.
func (q Quad) Equal(other *Quad) bool {
	if !q.ctx.Equal(other.ctx) {
		return false
	}

	if !q.Triple.Equal(other.Triple) {
		return false
	}

	return true
}
