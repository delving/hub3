// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package domain_test

import (
	"fmt"
	"testing"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/matryer/is"
)

const (
	dcTitle = "http://purl.org/dc/elements/1.1/title"
	dcNS    = "http://purl.org/dc/elements/1.1/"
	dcAltNS = "http://purl.org/dc/elements/1.2/"
)

func TestSplitURI(t *testing.T) {
	type args struct {
		uri string
	}

	tests := []struct {
		name     string
		args     args
		wantBase string
		wantName string
	}{
		{
			"split by /",
			args{dcTitle},
			dcNS,
			"title",
		},
		{
			"split by #",
			args{"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
			"http://www.w3.org/1999/02/22-rdf-syntax-ns#",
			"type",
		},
		{
			"unable to split URI",
			args{"urn:123"},
			"",
			"urn:123",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotBase, gotName := domain.SplitURI(tt.args.uri)
			if gotBase != tt.wantBase {
				t.Errorf("SplitURI() gotBase = %v, want %v", gotBase, tt.wantBase)
			}
			if gotName != tt.wantName {
				t.Errorf("SplitURI() gotName = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}

func ExampleSplitURI() {
	fmt.Println(domain.SplitURI("http://purl.org/dc/elements/1.1/subject"))
	// output: http://purl.org/dc/elements/1.1/ subject
}

// nolint:gocritic
func TestURI_String(t *testing.T) {
	is := is.New(t)

	is.Equal(
		domain.URI(dcTitle).String(),
		dcTitle,
	)
}
