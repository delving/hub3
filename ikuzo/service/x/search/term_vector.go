package search

type TermVector struct {
	Positions map[int]bool // wordpositions not tokenpositions
	Split     bool         // replace with elasticsearch term for synonym in same position
}

func NewTermVector() *TermVector {
	return &TermVector{
		Positions: make(map[int]bool),
	}
}

func (tv *TermVector) Size() int {
	return len(tv.Positions)
}
