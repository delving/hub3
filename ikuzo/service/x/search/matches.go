package search

type Matches struct {
	termFrequency map[string]int
	wordPositions map[int]bool
}

func NewMatches() *Matches {
	return &Matches{
		termFrequency: make(map[string]int),
		wordPositions: map[int]bool{},
	}
}

func (m *Matches) ApppendPositions(positions map[int]bool) {
	m.mergePositions(positions)
}

func (m *Matches) AppendTerm(term string, count int, positions map[int]bool) {
	m.termFrequency[term] = count
	m.mergePositions(positions)
}

func (m *Matches) Total() int {
	var total int
	for _, hitCount := range m.termFrequency {
		total += hitCount
	}

	return total
}

func (m *Matches) TermFrequency() map[string]int {
	return m.termFrequency
}

func (m *Matches) WordPositions() map[int]bool {
	return m.wordPositions
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
}

func (m *Matches) mergePositions(positions map[int]bool) {
	for key := range positions {
		_, ok := m.wordPositions[key]
		if ok {
			continue
		}

		m.wordPositions[key] = true
	}
}
