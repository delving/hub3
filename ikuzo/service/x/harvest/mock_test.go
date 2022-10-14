package harvest

import (
	"io"
	"log"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

// nolint:gocritic
func TestNewMockItems(t *testing.T) {
	is := is.New(t)

	items := newMockItems(1, 10, time.Now())

	is.Equal(len(items), 10)

	first := items[0]
	is.Equal(first.GetIdentifier(), "id-1")

	firstData, err := io.ReadAll(first.GetData())
	is.NoErr(err)
	is.Equal(string(firstData), "doc id-1")

	last := items[len(items)-1]
	is.Equal(last.GetIdentifier(), "id-10")

	lastData, err := io.ReadAll(last.GetData())
	is.NoErr(err)
	is.Equal(string(lastData), "doc id-10")
}

// nolint:gocritic
func TestMockPage(t *testing.T) {
	is := is.New(t)

	page := mockPage{
		cursor:           10,
		completeListSize: 50,
	}

	is.Equal(page.cursor, page.GetCursor())
	is.Equal(page.completeListSize, page.GetCompleteListSize())
	is.True(len(page.GetItems()) == 0)
}

func Test_mockService_HasNext(t *testing.T) {
	type fields struct {
		completeListSize int
		currentPage      int
		cursor           int
		items            []Item
		maxItems         int
		pageSize         int
		query            Query
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			"empty service",
			fields{pageSize: 50},
			false,
		},
		{
			"non-empty service without next page",
			fields{pageSize: 50, completeListSize: 10, cursor: 1},
			false,
		},
		{
			"non-empty service with next page",
			fields{pageSize: 50, completeListSize: 100, cursor: 1},
			true,
		},
		{
			"on last page",
			fields{pageSize: 50, completeListSize: 100, cursor: 51},
			false,
		},
		{
			"on last page small completeListSize",
			fields{pageSize: 50, completeListSize: 60, cursor: 51},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ms := &mockService{
				completeListSize: tt.fields.completeListSize,
				currentPage:      tt.fields.currentPage,
				cursor:           tt.fields.cursor,
				items:            tt.fields.items,
				maxItems:         tt.fields.maxItems,
				pageSize:         tt.fields.pageSize,
				query:            tt.fields.query,
			}
			if got := ms.HasNext(); got != tt.want {
				t.Errorf("mockService.HasNext() = %v, want %v", got, tt.want)
			}
		})
	}
}

// nolint:funlen // table tests can have longer function length
func Test_mockService_First(t *testing.T) {
	type fields struct {
		maxItems    int
		seedTime    time.Time
		startCursor int
		itemWindow  int
	}

	type args struct {
		q Query
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Page
		wantErr bool
		hasNext bool
	}{
		{
			"all items",
			fields{
				maxItems:    10,
				seedTime:    time.Now(),
				startCursor: 1,
				itemWindow:  10,
			},
			args{},
			mockPage{
				completeListSize: 10,
				cursor:           1,
			},
			false,
			false,
		},
		{
			"has next page",
			fields{
				maxItems:    100,
				seedTime:    time.Now(),
				startCursor: 1,
				itemWindow:  50,
			},
			args{},
			mockPage{
				completeListSize: 100,
				cursor:           1,
			},
			false,
			true,
		},
		{
			"empty result",
			fields{
				maxItems:    100,
				seedTime:    time.Now(),
				startCursor: 1,
				itemWindow:  0,
			},
			args{
				Query{
					From:  time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
					Until: time.Date(2010, 1, 1, 12, 0, 0, 0, time.UTC),
				},
			},
			mockPage{
				completeListSize: 0,
				cursor:           1,
				items:            []Item{},
			},
			false,
			false,
		},
		{
			"date filtered from",
			fields{
				maxItems:    100,
				seedTime:    time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
				startCursor: 11,
				itemWindow:  30,
			},
			args{
				Query{
					From:  time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC).Add(10 * time.Second),
					Until: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC).Add(41 * time.Second),
				},
			},
			mockPage{
				completeListSize: 30,
				cursor:           1,
			},
			false,
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ms := newMockService(tt.fields.maxItems)
			ms.seedTime = tt.fields.seedTime

			page := tt.want.(mockPage)

			for _, item := range newMockItems(tt.fields.startCursor, tt.fields.itemWindow, tt.fields.seedTime) {
				page.items = append(page.items, item)
			}

			got, err := ms.First(tt.args.q)
			if (err != nil) != tt.wantErr {
				t.Errorf("mockService.First() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.hasNext != ms.HasNext() {
				t.Errorf("mockService.HasNext() error = got %T, want %T", ms.HasNext(), tt.hasNext)
				return
			}

			if diff := cmp.Diff(page, got, cmp.AllowUnexported(mockPage{}, mockItem{})); diff != "" {
				t.Errorf("mockService.First() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func Test_mockService_Next(t *testing.T) {
	type fields struct {
		maxItems    int
		seedTime    time.Time
		startCursor int
		itemWindow  int
	}

	type args struct {
		q Query
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Page
		wantErr bool
		hasNext bool
	}{
		{
			"all items",
			fields{
				maxItems:    80,
				seedTime:    time.Now(),
				startCursor: 51,
				itemWindow:  30,
			},
			args{},
			mockPage{
				completeListSize: 80,
				cursor:           51,
			},
			false,
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ms := newMockService(tt.fields.maxItems)
			ms.seedTime = tt.fields.seedTime

			page := tt.want.(mockPage)

			for _, item := range newMockItems(tt.fields.startCursor, tt.fields.itemWindow, tt.fields.seedTime) {
				log.Printf("last modified: %s", item.lastModified)
				page.items = append(page.items, item)
			}

			_, err := ms.First(tt.args.q)
			if (err != nil) != tt.wantErr {
				t.Errorf("mockService.Next() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if ms.HasNext() {
				got, err := ms.Next()
				if (err != nil) != tt.wantErr {
					t.Errorf("mockService.Next() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if diff := cmp.Diff(page, got, cmp.AllowUnexported(mockPage{}, mockItem{})); diff != "" {
					t.Errorf("mockService.Next() %s = mismatch (-want +got):\n%s", tt.name, diff)
				}

				if tt.hasNext != ms.HasNext() {
					t.Errorf("mockService.HasNext() error = got %t, want %t", ms.HasNext(), tt.hasNext)
					return
				}
			}
		})
	}
}

// nolint:gocritic
func TestMockServiceAddNew(t *testing.T) {
	is := is.New(t)

	ms := newMockService(10)
	ms.seedTime = time.Now()

	_, err := ms.First(Query{})
	is.NoErr(err)

	is.Equal(ms.completeListSize, 10)

	ms.addNew(2)

	is.Equal(ms.completeListSize, 12)

	last := ms.items[len(ms.items)-1].(*mockItem)

	is.Equal(last.GetIdentifier(), "id-12")
}

// nolint:gocritic
func TestMockService_Modify(t *testing.T) {
	is := is.New(t)

	ms := newMockService(10)
	ms.seedTime = time.Now()

	_, err := ms.First(Query{})
	is.NoErr(err)

	is.Equal(ms.completeListSize, 10)

	ms.modify([]string{"id-5", "id-8"}, false)

	is.Equal(ms.completeListSize, 10)

	last := ms.items[len(ms.items)-1].(*mockItem)

	is.Equal(last.GetIdentifier(), "id-8")
	is.True(!last.deleted)

	ms.modify([]string{"id-1"}, true)

	is.Equal(ms.completeListSize, 10)

	lastDeleted := ms.items[len(ms.items)-1].(*mockItem)

	is.Equal(lastDeleted.GetIdentifier(), "id-1")
	is.True(lastDeleted.deleted)

	first := ms.items[0].(*mockItem)

	is.Equal(first.GetIdentifier(), "id-2")
	is.True(!first.deleted)
}
