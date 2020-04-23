package search

type Matches struct {
	termFrequency map[string]int
	termVectors   *Vectors
}

func NewMatches() *Matches {
	return &Matches{
		termFrequency: make(map[string]int),
		termVectors:   NewVectors(),
	}
}

func (m *Matches) AppendTerm(term string, tv *Vectors) {
	if tv.Size() == 0 {
		return
	}

	m.termFrequency[term] = tv.Size()
	m.mergeVectors(tv)
}

func (m *Matches) DocCount() int {
	return m.termVectors.DocCount()
}

func (m *Matches) HasDocID(docID int) bool {
	return m.termVectors.HasDoc(docID)
}

func (m *Matches) Merge(matches *Matches) {
	for key, count := range matches.termFrequency {
		v, ok := m.termFrequency[key]
		if ok {
			m.termFrequency[key] = v + count
			continue
		}

		m.termFrequency[key] = count
	}

	m.mergeVectors(matches.termVectors)
}

func (m *Matches) mergeVectors(tv *Vectors) {
	m.termVectors.Merge(tv)
}

func (m *Matches) TermFrequency() map[string]int {
	return m.termFrequency
}

func (m *Matches) TermCount() int {
	return len(m.termFrequency)
}

func (m *Matches) Total() int {
	var total int
	for _, hitCount := range m.termFrequency {
		total += hitCount
	}

	return total
}

func (m *Matches) Vectors() *Vectors {
	return m.termVectors
}
