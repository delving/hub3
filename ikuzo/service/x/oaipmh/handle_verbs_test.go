package oaipmh

import (
	"context"
	"encoding/xml"
	"strings"
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

func TestResponse_marshalEnvelope(t *testing.T) {
	resp, err := (&Service{}).Do(context.Background(), &Request{
		Verb:     string(VerbListRecords),
		BaseURL:  "http://localhost/api/oai-pmh",
		DebugOut: "/tmp/debug",
	})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	out, err := xml.Marshal(resp)
	if err != nil {
		t.Fatalf("xml.Marshal() error = %v", err)
	}

	got := string(out)

	for _, want := range []string{
		`xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`,
		`xsi:schemaLocation="http://www.openarchives.org/OAI/2.0/ http://www.openarchives.org/OAI/2.0/OAI-PMH.xsd"`,
		`<request>http://localhost/api/oai-pmh</request>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("marshaled response missing %s\ngot: %s", want, got)
		}
	}

	for _, notWant := range []string{"DebugOut", "_XMLSchema-instance"} {
		if strings.Contains(got, notWant) {
			t.Errorf("marshaled response contains %s\ngot: %s", notWant, got)
		}
	}
}
