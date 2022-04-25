package handlers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_ensureSparqlLimit(t *testing.T) {
	type args struct {
		query string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"no limit in query",
			args{query: "select * where {?s ?p ?o}"},
			"select * where {?s ?p ?o} LIMIT 25",
			false,
		},
		{
			"with limit in query",
			args{query: "select * where {?s ?p ?o} LIMIT 100"},
			"select * where {?s ?p ?o} LIMIT 100",
			false,
		},
		{
			"with limit in query",
			args{query: "select * where {?s ?p ?o} LIMIT 1500"},
			"",
			true,
		},
		{
			"with limit badly formatted limit",
			args{query: "select * where {?s ?p ?o} LIMIT A15AA"},
			"",
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got, err := ensureSparqlLimit(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureSparqlLimit() %s; error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ensureSparqlLimit() %s; mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
