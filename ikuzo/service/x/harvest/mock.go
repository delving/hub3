package harvest

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// make sure item satisfies harvest.Item interface.
var (
	_ Item   = (*mockItem)(nil)
	_ Syncer = (*mockService)(nil)
)

type mockItem struct {
	id           string
	lastModified time.Time
	deleted      bool
}

func (m mockItem) GetLastModified() time.Time {
	return m.lastModified
}

func (m mockItem) GetIdentifier() string {
	return m.id
}

func (m mockItem) GetData() io.Reader {
	return strings.NewReader(fmt.Sprintf("doc %s", m.id))
}

func newMockItems(start, max int, seedTime time.Time) []*mockItem {
	var items []*mockItem

	for i := start; i <= max+start-1; i++ {
		modified := seedTime.Add(time.Duration(i) * time.Second)

		items = append(items, &mockItem{id: fmt.Sprintf("id-%d", i), lastModified: modified})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].GetLastModified().Before(items[j].GetLastModified())
	})

	return items
}

func sortItems(items []Item) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].GetLastModified().Before(items[j].GetLastModified())
	})
}

type mockPage struct {
	cursor           int
	completeListSize int
	items            []Item
}

func (p mockPage) GetCursor() int {
	return p.cursor
}

func (p mockPage) GetCompleteListSize() int {
	return p.completeListSize
}

func (p mockPage) GetItems() []Item {
	return p.items
}

type mockService struct {
	completeListSize int
	currentPage      int
	cursor           int
	items            []Item
	maxItems         int
	pageSize         int
	query            Query
	seedTime         time.Time
}

func newMockService(maxItems int) *mockService {
	return &mockService{
		maxItems: maxItems,
		pageSize: 50,
	}
}

func (ms *mockService) HasNext() bool {
	return (ms.cursor + ms.pageSize) < ms.completeListSize
}

func (ms *mockService) First(q Query) (Page, error) {
	ms.query = q
	ms.currentPage = 1
	ms.cursor = 1

	res := []Item{}

	for _, val := range newMockItems(1, ms.maxItems, ms.seedTime) {
		if ms.query.Valid(val.GetLastModified()) {
			res = append(res, Item(val))
		}
	}

	ms.items = res
	ms.completeListSize = len(res)

	end := ms.pageSize
	if end > ms.completeListSize {
		end = ms.completeListSize
	}

	return mockPage{
		items:            ms.items[:end],
		completeListSize: ms.completeListSize,
		cursor:           ms.cursor,
	}, nil
}

func (ms *mockService) Next() (Page, error) {
	start := ms.currentPage * ms.pageSize

	// increment counters
	ms.currentPage++
	ms.cursor += ms.pageSize

	end := ms.currentPage * ms.pageSize

	if end > ms.completeListSize {
		end = ms.completeListSize
	}

	return mockPage{
		items:            ms.items[start:end],
		completeListSize: ms.completeListSize,
		cursor:           ms.cursor,
	}, nil
}

func (ms *mockService) addNew(nr int) {
	items := newMockItems(ms.completeListSize+1, nr, ms.seedTime)
	ms.completeListSize += nr

	for _, item := range items {
		ms.items = append(ms.items, Item(item))
	}

	sortItems(ms.items)
}

func (ms *mockService) modify(ids []string, deleted bool) {
	last := ms.items[len(ms.items)-1]

	for idx, item := range ms.items {
		for _, id := range ids {
			if item.GetIdentifier() == id {
				mItem := item.(*mockItem)
				mItem.deleted = deleted
				mItem.lastModified = last.GetLastModified().Add(time.Duration(idx+3) * time.Second)
			}
		}
	}

	sortItems(ms.items)
}
