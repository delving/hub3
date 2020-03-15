// nolint:gocritic
package elasticsearch

import (
	"encoding/json"
	"testing"

	"github.com/delving/hub3/ikuzo/service/x/search"
	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestQuery(t *testing.T) {
	is := is.New(t)

	qp, err := search.NewQueryParser()
	is.NoErr(err)

	qt, err := qp.Parse("two words")
	is.NoErr(err)

	bq := NewBoolQuery(qt)
	is.True(bq != nil)

	bqSource, err := bq.Source()
	is.NoErr(err)

	got, err := json.Marshal(bqSource)
	is.NoErr(err)

	if diff := cmp.Diff("{\"bool\":{}}", string(got)); diff != "" {
		t.Errorf("NewBoolQuery(); %s = mismatch (-want +got):\n%s", "first", diff)
	}
}

func TestNewBoolQuery(t *testing.T) {
	is := is.New(t)

	type args struct {
		q string
	}

	tests := []struct {
		name       string
		args       args
		defaultAnd bool
		want       string
	}{
		{
			"empty query is match all",
			args{q: ""},
			false,
			"{\"bool\":{}}",
		},
		{
			"single word OR query",
			args{q: "one"},
			false,
			"{\"bool\":{}}",
			// "{\"match\":{}}",
		},
	}

	for _, tt := range tests {
		tt := tt

		qp, err := search.NewQueryParser()
		is.NoErr(err)

		qt, err := qp.Parse("two words")
		is.NoErr(err)

		bq := NewBoolQuery(qt)
		is.True(bq != nil)

		bqSource, err := bq.Source()
		is.NoErr(err)

		got, err := json.Marshal(bqSource)
		is.NoErr(err)
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, string(got)); diff != "" {
				t.Errorf("NewBoolQuery(); %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
