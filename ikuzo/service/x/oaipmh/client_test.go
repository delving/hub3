package oaipmh

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestNewRequest(t *testing.T) {
	type args struct {
		r *http.Request
	}

	req := func(url string) *http.Request {
		r, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			t.Fatalf("unable to build request: %s", err)
		}

		return r
	}

	tests := []struct {
		name string
		args args
		want Request
	}{
		{
			"empty",
			args{req("http://localhost:3000/api/oai-pmh/ead")},
			Request{
				BaseURL: "http://localhost:3000/api/oai-pmh/ead",
			},
		},
		{
			"full request",
			args{
				req(
					"https://localhost:3000/api/oai-pmh/ead?verb=ListRecords" +
						"&metadataPrefix=oai_dc&set=all&from=1970-01-01" +
						"&until=2000-01-01",
				)},
			Request{
				BaseURL:        "https://localhost:3000/api/oai-pmh/ead",
				Set:            "all",
				MetadataPrefix: "oai_dc",
				Verb:           "ListRecords",
				From:           "1970-01-01",
				Until:          "2000-01-01",
			},
		},
		{
			"resumptionToken request",
			args{
				req(
					"https://localhost:3000/api/oai-pmh/ead?verb=ListRecords" +
						"&resumptionToken=123abc",
				)},
			Request{
				BaseURL:         "https://localhost:3000/api/oai-pmh/ead",
				Verb:            "ListRecords",
				ResumptionToken: "123abc",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got := NewRequest(tt.args.r)
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(Request{})); diff != "" {
				t.Errorf("NewRequest() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
