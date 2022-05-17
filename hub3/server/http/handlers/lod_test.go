// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"net/http"
	"testing"

	"github.com/delving/hub3/config"
)

func Test_getResolveURL(t *testing.T) {
	config.Config.RDF.BaseURL = "http://data.hub3.org"

	type args struct {
		url string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"simple id url",
			args{"http://localhost:3000/id/123"},
			"http://data.hub3.org/doc/123",
		},
		{
			"simple doc url",
			args{"http://localhost:3000/doc/123"},
			"http://data.hub3.org/doc/123",
		},
		{
			"query params ignored",
			args{"http://localhost:3000/id/123?id=bla"},
			"http://data.hub3.org/doc/123",
		},
		{
			"simple resource url",
			args{"http://localhost:3000/resource/document/dataset/123"},
			"http://data.hub3.org/data/document/dataset/123",
		},
	}

	for _, tt := range tests {
		tt := tt

		r, _ := http.NewRequest(http.MethodGet, tt.args.url, http.NoBody)
		t.Run(tt.name, func(t *testing.T) {
			if got := getResolveURL(r); got != tt.want {
				t.Errorf("getResolveURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSparqlSubject(t *testing.T) {
	type args struct {
		iri      string
		fragment string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"with resource",
			args{iri: "https://test.nl/data/123", fragment: ""},
			"https://test.nl/resource/123",
			false,
		},
		{
			"without fragments",
			args{iri: "https://test.nl/doc/123"},
			"https://test.nl/id/123",
			false,
		},
		{
			"with fragments",
			args{iri: "https://test.nl/doc/123", fragment: "hello"},
			"https://test.nl/id/123#hello",
			false,
		},
		{
			"with fragments and def",
			args{iri: "https://test.nl/def/ontology", fragment: "label"},
			"https://test.nl/def/ontology#label",
			false,
		},
		{
			"with fragments in iri",
			args{iri: "https://test.nl/def/ontology#label", fragment: ""},
			"https://test.nl/def/ontology#label",
			false,
		},
		{
			"invalid uri",
			args{iri: "", fragment: ""},
			"",
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got, err := getSparqlSubject(tt.args.iri, tt.args.fragment)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSparqlSubject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getSparqlSubject() = %v, want %v", got, tt.want)
			}
		})
	}
}
