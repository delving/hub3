package search

type Vector struct {
	DocID    int
	Location int
}

type Vectors struct {
	Locations     map[Vector]bool
	Docs          map[int]bool
	PhraseVectors int
}

func NewVectors() *Vectors {
	return &Vectors{
		Locations: make(map[Vector]bool),
		Docs:      make(map[int]bool),
	}
}

// pos must not be 0
func (tv *Vectors) Add(doc, pos int) {
	v := Vector{Location: pos, DocID: doc}
	tv.AddVector(v)
}

func (tv *Vectors) AddVector(vector Vector) {
	if _, ok := tv.Docs[vector.DocID]; !ok {
		tv.Docs[vector.DocID] = true
	}

	tv.Locations[vector] = true
}

func (tv *Vectors) AddPhraseVector(vector Vector) {
	if tv.HasVector(vector) {
		// increment phraseVectors must be idempotent
		return
	}

	tv.AddVector(vector)
	tv.PhraseVectors++
}

func (tv *Vectors) DocCount() int {
	return len(tv.Docs)
}

func (tv *Vectors) HasDoc(doc int) bool {
	_, ok := tv.Docs[doc]
	return ok
}

func (tv *Vectors) HasVector(vector Vector) bool {
	_, ok := tv.Locations[vector]
	return ok
}

func (tv *Vectors) Merge(vectors *Vectors) {
	for vector := range vectors.Locations {
		if !tv.HasVector(vector) {
			tv.AddVector(vector)
		}
	}
}

func (tv *Vectors) Size() int {
	return len(tv.Locations) - tv.PhraseVectors
}
