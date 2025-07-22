package search

import (
	"errors"
	"fmt"
)

var ErrInvalidPage = errors.New("invalid page range requested")

type Paginator struct {
	Start              int        `json:"start"` //  start is 1 based
	Rows               int        `json:"rows"`
	NumFound           int        `json:"numFound"`
	FirstPage          int        `json:"firstPage"`
	LastPage           int        `json:"lastPage"`
	CurrentPage        int        `json:"currentPage"`
	HasNext            bool       `json:"hasNext"`
	HasPrevious        bool       `json:"hasPrevious"`
	NextPageNumber     int        `json:"nextPageNumber"`
	PreviousPageNumber int        `json:"previousPageNumber"`
	NextPage           int        `json:"nextPage"`
	PreviousPage       int        `json:"previousPage"`
	Links              []PageLink `json:"links"`
	// When backend has a hard limit like ElasticSearch for paging you can
	// set a max here and it will return an error when it is exceeded.
	MaxPagingWindow int `json:"-"`
}

type PageLink struct {
	Start      int  `json:"start"`
	IsLinked   bool `json:"isLinked"`
	PageNumber int  `json:"pageNumber"`
}

// NewPaginator creates a Paginator without PageLinks.
// You need to call AddPageLinks() to add them to the Paginator.
// Cursor is zero-based. If page is not zero, the cursor value is ignored
func NewPaginator(total, pageSize, currentPage, cursor int) (*Paginator, error) {
	if currentPage < 1 {
		currentPage = 1
	}

	p := &Paginator{
		NumFound:    total,
		Rows:        pageSize,
		CurrentPage: currentPage,
	}

	// cursor is zero based but this is 1 based
	if cursor != 0 {
		p.Start = cursor + 1
	}

	if p.CurrentPage == 1 && cursor != 0 {
		page, err := p.getPageNumber()
		if err != nil {
			return nil, err
		}

		p.CurrentPage = page
	}

	if err := p.setPaging(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Paginator) getPageNumber() (int, error) {
	if p.NumFound < 1 || p.Rows < 1 {
		return 1, nil
	}

	if p.CurrentPage > 1 {
		return p.CurrentPage, nil
	}

	page := p.Start / p.Rows
	if p.Start%p.Rows != 0 {
		page++
	}

	return page, nil
}

func (p *Paginator) AddPageLinks() error {
	links, err := p.getPageLinks()
	if err != nil {
		return fmt.Errorf("unable to add PageLinks to Paginator; %w", err)
	}

	p.Links = links

	return nil
}

func (p *Paginator) getPageLinks() ([]PageLink, error) {
	totalLinksWanted := 10
	
	// Calculate the ideal window
	firstPage := 1
	lastPage := p.LastPage
	
	if p.LastPage <= totalLinksWanted {
		// If total pages <= 10, show all pages
		firstPage = 1
		lastPage = p.LastPage
	} else {
		// Try to center the current page with 5 on each side
		firstPage = p.CurrentPage - 5
		lastPage = p.CurrentPage + 4
		
		// Adjust if we're too close to the beginning
		if firstPage < 1 {
			firstPage = 1
			lastPage = totalLinksWanted
		}
		
		// Adjust if we're too close to the end
		if lastPage > p.LastPage {
			lastPage = p.LastPage
			firstPage = p.LastPage - totalLinksWanted + 1
			if firstPage < 1 {
				firstPage = 1
			}
		}
	}

	links := []PageLink{}

	for i := firstPage; i <= lastPage; i++ {
		start := ((i - 1) * p.Rows) + 1
		if start < 1 {
			start = 1
		}

		links = append(
			links,
			PageLink{
				PageNumber: i,
				IsLinked:   i != p.CurrentPage,
				Start:      start,
			},
		)
	}

	if len(links) == 0 {
		links = append(links, PageLink{PageNumber: 1, IsLinked: false, Start: 1})
	}

	return links, nil
}

func (p *Paginator) setPaging() error {
	if p.NumFound > 0 && p.Rows > 0 {
		p.FirstPage = 1
		p.LastPage = (p.NumFound / p.Rows)

		if p.NumFound%p.Rows != 0 {
			p.LastPage++
		}

		if p.Start == 0 {
			p.Start = (p.Rows * (p.CurrentPage - 1)) + 1
		}

		if p.CurrentPage > p.LastPage {
			return ErrInvalidPage
		}
	}

	if p.CurrentPage+1 <= p.LastPage {
		p.HasNext = true
		p.NextPage = p.Rows + p.Start
		p.NextPageNumber = p.CurrentPage + 1
	}

	if p.CurrentPage > 1 {
		p.HasPrevious = true
		p.PreviousPageNumber = p.CurrentPage - 1
		p.PreviousPage = p.Start - p.Rows
	}

	return nil
}
