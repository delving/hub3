// nolint:funlen
package eadpb

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/delving/hub3/hub3/fragments"
	proto "github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestNewPager(t *testing.T) {
	type args struct {
		total   int64
		request *ViewRequest
	}

	tests := []struct {
		name    string
		args    args
		want    Pager
		wantErr bool
	}{
		{
			"no results",
			args{total: 0, request: &ViewRequest{PageSize: int32(10), Page: int32(1)}},
			Pager{
				HasNext:        false,
				HasPrevious:    false,
				TotalCount:     0,
				NrPages:        0,
				PageCurrent:    1,
				PageNext:       0,
				PagePrevious:   0,
				PageSize:       10,
				ActiveFilename: "",
				ActiveSortKey:  0,
				Paging:         false,
			},
			false,
		},
		{
			"1 page results",
			args{total: 3, request: &ViewRequest{PageSize: int32(10), Page: int32(1)}},
			Pager{
				HasNext:        false,
				HasPrevious:    false,
				TotalCount:     3,
				NrPages:        1,
				PageCurrent:    1,
				PageNext:       0,
				PagePrevious:   0,
				PageSize:       10,
				ActiveFilename: "",
				ActiveSortKey:  0,
				Paging:         false,
			},
			false,
		},
		{
			"first page; 2 page results; 11 hits",
			args{total: 11, request: &ViewRequest{PageSize: int32(10), Page: int32(1)}},
			Pager{
				HasNext:        true,
				HasPrevious:    false,
				TotalCount:     11,
				NrPages:        2,
				PageCurrent:    1,
				PageNext:       2,
				PagePrevious:   0,
				PageSize:       10,
				ActiveFilename: "",
				ActiveSortKey:  0,
				Paging:         false,
			},
			false,
		},
		{
			"second page; 2 page results; 11 hits",
			args{total: 11, request: &ViewRequest{PageSize: int32(10), Page: int32(2)}},
			Pager{
				HasNext:        false,
				HasPrevious:    true,
				TotalCount:     11,
				NrPages:        2,
				PageCurrent:    2,
				PageNext:       0,
				PagePrevious:   1,
				PageSize:       10,
				ActiveFilename: "",
				ActiveSortKey:  0,
				Paging:         false,
			},
			false,
		},
		{
			"1 page results; 10 hits; active file index",
			args{total: 10, request: &ViewRequest{
				PageSize: int32(10),
				Page:     int32(1),
				SortKey:  4,
			}},
			Pager{
				HasNext:        false,
				HasPrevious:    false,
				TotalCount:     10,
				NrPages:        1,
				PageCurrent:    1,
				PageNext:       0,
				PagePrevious:   0,
				PageSize:       10,
				ActiveFilename: "",
				ActiveSortKey:  4,
				Paging:         false,
			},
			false,
		},
		{
			"1 page results; 10 hits; active fileName",
			args{total: 10, request: &ViewRequest{
				PageSize: int32(10),
				Page:     int32(1),
				Filename: "123.jpg",
			}},
			Pager{
				HasNext:        false,
				HasPrevious:    false,
				TotalCount:     10,
				NrPages:        1,
				PageCurrent:    1,
				PageNext:       0,
				PagePrevious:   0,
				PageSize:       10,
				ActiveFilename: "123.jpg",
				ActiveSortKey:  0,
				Paging:         false,
			},
			false,
		},
		{
			"10 page results; 105 hits",
			args{total: 105, request: &ViewRequest{
				PageSize: int32(10),
				Page:     int32(5),
				Paging:   true,
			}},
			Pager{
				HasNext:        true,
				HasPrevious:    true,
				TotalCount:     105,
				NrPages:        11,
				PageCurrent:    5,
				PageNext:       6,
				PagePrevious:   4,
				PageSize:       10,
				ActiveFilename: "",
				ActiveSortKey:  0,
				Paging:         true,
			},
			false,
		},
		{
			"1 page results, 15 hits",
			args{total: 15, request: &ViewRequest{
				PageSize: int32(1),
				Page:     int32(5),
			}},
			Pager{
				HasNext:        true,
				HasPrevious:    true,
				TotalCount:     15,
				NrPages:        15,
				PageCurrent:    5,
				PageNext:       6,
				PagePrevious:   4,
				PageSize:       1,
				ActiveFilename: "",
				ActiveSortKey:  0,
				Paging:         false,
			},
			false,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPager(tt.args.total, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPager(); %s error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(Pager{})) {
				t.Errorf("NewPager(); %s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestSetPage(t *testing.T) {
	type args struct {
		total   int64
		sortKey int32
		request *ViewRequest
	}

	tests := []struct {
		name    string
		args    args
		want    Pager
		wantErr bool
	}{
		{
			"page 1 on 2 page results",
			args{
				total:   17,
				request: &ViewRequest{PageSize: int32(10), Page: int32(1)},
				sortKey: int32(5),
			},
			Pager{
				HasNext:        true,
				HasPrevious:    false,
				TotalCount:     17,
				NrPages:        2,
				PageCurrent:    1,
				PageNext:       2,
				PagePrevious:   0,
				PageSize:       10,
				ActiveFilename: "",
				ActiveSortKey:  5,
				Paging:         false,
			},
			false,
		},
		{
			"page 2 on 2 page results",
			args{
				total:   17,
				request: &ViewRequest{PageSize: int32(10), Page: int32(1)},
				sortKey: int32(15),
			},
			Pager{
				HasNext:        false,
				HasPrevious:    true,
				TotalCount:     17,
				NrPages:        2,
				PageCurrent:    2,
				PageNext:       0,
				PagePrevious:   1,
				PageSize:       10,
				ActiveFilename: "",
				ActiveSortKey:  15,
				Paging:         false,
			},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPager(tt.args.total, tt.args.request)
			got.SetPage(tt.args.sortKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPager(); %s error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(Pager{})) {
				t.Errorf("NewPager(); %s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_DecodePBFile(t *testing.T) {
	setupFile := func(text string) string {
		// setup the file
		file := &File{Filename: text}

		b, err := proto.Marshal(file)
		if err != nil {
			t.Fatalf("unable to marshal file; %#v", err)
			return ""
		}

		return fmt.Sprintf("%x", b)
	}
	createRawMessage := func(text, messageType string) json.RawMessage {
		// setup pbWrapper
		fw := pbWrapper{
			Protobuf: &fragments.ProtoBuf{
				MessageType: messageType,
				Data:        text,
			},
		}

		// marshal json
		jsonRaw, err := json.MarshalIndent(fw, "", " ")
		if err != nil {
			t.Fatalf("unable to marshal json; %#v", err)
			return nil
		}

		// store in RawMessage
		return json.RawMessage(jsonRaw)
	}

	type args struct {
		hit json.RawMessage
	}

	tests := []struct {
		name    string
		args    args
		want    *File
		wantErr bool
	}{
		{
			"simple file",
			args{createRawMessage(setupFile("123.jpg"), "pb.File")},
			&File{Filename: "123.jpg"},
			false,
		},
		{
			"wrong protobuf messageType",
			args{createRawMessage(setupFile("123.jpg"), "File2")},
			nil,
			true,
		},
		{
			"wrong json.RawMessage input",
			args{func() json.RawMessage { data := []byte(""); return json.RawMessage(data) }()},
			nil,
			true,
		},
		{
			"wrong protobuf input",
			args{createRawMessage("super bad pb", "pb.File")},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodePBFile(tt.args.hit)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodePBFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(File{})) {
				t.Errorf("DecodePBFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
