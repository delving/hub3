package ead

import (
	"github.com/delving/hub3/ikuzo/storage/x/memory"
)

// DescriptionQuery can be used to query and highlight matches in the ead.Description
type DescriptionQuery struct {
	tq     *memory.TextQuery
	Filter bool
}

// NewDescriptionQuery returns a DescriptionQuery that can be used to filter
// and hightlight DataItems.
func NewDescriptionQuery(query string) *DescriptionQuery {
	tq, _ := memory.NewTextQueryFromString(query)

	return &DescriptionQuery{
		tq:     tq,
		Filter: true,
	}
}

// FilterMatches filters a []DataItem for query matches and highlights them.
func (dq *DescriptionQuery) FilterMatches(items []*DataItem) []*DataItem {
	matches := []*DataItem{}

	for _, item := range items {
		text, ok := dq.highlightQuery(item.Text)
		if !ok && dq.Filter {
			continue
		}
		item.Text = text
		matches = append(matches, item)
	}

	return matches
}

func (dq *DescriptionQuery) Seen() int {
	return dq.tq.Hits.Total()
}

func (dq *DescriptionQuery) Hits() *memory.SearchHits {
	return dq.tq.Hits
}

// HightlightSummary applied query highlights to the ead.Summary.
func (dq *DescriptionQuery) HightlightSummary(s Summary) Summary {
	if s.Profile != nil {
		s.Profile.Creation, _ = dq.highlightQuery(s.Profile.Creation)
		s.Profile.Language, _ = dq.highlightQuery(s.Profile.Language)
	}

	if s.File != nil {
		s.File.Author, _ = dq.highlightQuery(s.File.Author)
		s.File.Copyright, _ = dq.highlightQuery(s.File.Copyright)
		s.File.PublicationDate, _ = dq.highlightQuery(s.File.PublicationDate)
		s.File.Publisher, _ = dq.highlightQuery(s.File.Publisher)
		s.File.Title, _ = dq.highlightQuery(s.File.Title)
		var editions []string
		for _, e := range s.File.Edition {
			edition, _ := dq.highlightQuery(e)
			editions = append(editions, edition)
		}
		if len(editions) != 0 {
			s.File.Edition = editions
		}
	}

	if s.FindingAid != nil {
		s.FindingAid.AgencyCode, _ = dq.highlightQuery(s.FindingAid.AgencyCode)
		s.FindingAid.Country, _ = dq.highlightQuery(s.FindingAid.Country)
		s.FindingAid.ID, _ = dq.highlightQuery(s.FindingAid.ID)
		s.FindingAid.ShortTitle, _ = dq.highlightQuery(s.FindingAid.ShortTitle)
		var titles []string
		for _, t := range s.FindingAid.Title {
			title, _ := dq.highlightQuery(t)
			titles = append(titles, title)
		}
		if len(titles) != 0 {
			s.FindingAid.Title = titles
		}
	}

	if s.FindingAid != nil && s.FindingAid.UnitInfo != nil {
		unit := s.FindingAid.UnitInfo

		unit.ID, _ = dq.highlightQuery(unit.ID)
		unit.Language, _ = dq.highlightQuery(unit.Language)
		unit.DateBulk, _ = dq.highlightQuery(unit.DateBulk)
		unit.Files, _ = dq.highlightQuery(unit.Files)
		unit.Length, _ = dq.highlightQuery(unit.Length)
		unit.Material, _ = dq.highlightQuery(unit.Material)
		unit.Physical, _ = dq.highlightQuery(unit.Physical)
		unit.PhysicalLocation, _ = dq.highlightQuery(unit.PhysicalLocation)
		unit.Repository, _ = dq.highlightQuery(unit.Repository)

		var origins []string
		for _, o := range unit.Origin {
			origin, _ := dq.highlightQuery(o)
			origins = append(origins, origin)
		}
		if len(origins) != 0 {
			unit.Origin = origins
		}

		var dates []string
		for _, d := range unit.Date {
			date, _ := dq.highlightQuery(d)
			dates = append(dates, date)
		}
		if len(dates) != 0 {
			unit.Date = dates
		}

		var abstracts []string
		for _, a := range unit.Abstract {
			abstract, _ := dq.highlightQuery(a)
			abstracts = append(abstracts, abstract)
		}
		if len(abstracts) != 0 {
			unit.Abstract = abstracts
		}
		s.FindingAid.UnitInfo = unit
	}

	return s
}

func (dq *DescriptionQuery) highlightQuery(text string) (string, bool) {
	return dq.tq.Highlight(text)
}
