package harvest

import (
	"errors"
	"io"
	"time"
)

var (
	ErrNoMatch = errors.New("no items match harvest request")
)

type Item interface {
	GetLastModified() time.Time
	GetIdentifier() string
	GetData() io.Reader
}

type Page interface {
	GetCursor() int
	GetCompleteListSize() int
	GetItems() []Item
}

type Service interface {
	Next() (Page, error)
	First(q Query) (Page, error)
}

type Query struct {
	From  time.Time
	Until time.Time
}

func (q Query) Valid(t time.Time) bool {
	if !q.From.IsZero() && !t.After(q.From) {
		return false
	}

	if !q.Until.IsZero() && !t.Before(q.Until) {
		return false
	}

	return true
}
