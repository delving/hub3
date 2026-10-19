package oaipmh

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestService_Do_missingMetadataPrefix(t *testing.T) {
	tests := []struct {
		name string
		req  Request
	}{
		{"ListRecords without metadataPrefix", Request{Verb: string(VerbListRecords), Set: "nk"}},
		{"ListIdentifiers without metadataPrefix", Request{Verb: string(VerbListIdentifiers), Set: "nk"}},
		{"GetRecord without metadataPrefix", Request{Verb: string(VerbGetRecord), Identifier: "nk_1"}},
		{"GetRecord without identifier", Request{Verb: string(VerbGetRecord), MetadataPrefix: "rdfxml"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.req.BaseURL = "http://localhost/api/oai-pmh"

			resp, err := (&Service{}).Do(context.Background(), &tt.req)
			if err != nil {
				t.Fatalf("Do() error = %v", err)
			}

			if diff := cmp.Diff([]Error{ErrBadArgument}, resp.Error, cmp.AllowUnexported(Error{})); diff != "" {
				t.Errorf("Do() errors mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(&Request{BaseURL: tt.req.BaseURL}, resp.Request, cmp.AllowUnexported(Request{})); diff != "" {
				t.Errorf("Do() request must be reduced to base URL (-want +got):\n%s", diff)
			}
		})
	}
}
