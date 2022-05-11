// Copyright 2017 Delving B.V.
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

package fragments_test

import (
	"crypto/tls"
	"net/http"
	"testing"

	. "github.com/delving/hub3/hub3/fragments"
	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestNewHyperMediaDataSet(t *testing.T) {
	is := is.New(t)

	type args struct {
		req       string
		totalHits int64
	}

	baseURL := "https://localhost:3000/fragments/museum"

	tests := []struct {
		name string
		args args
		want *HyperMediaDataSet
	}{
		{
			"first page and only page",
			args{
				req:       baseURL,
				totalHits: 10,
			},
			&HyperMediaDataSet{
				DataSetURI:   baseURL,
				PagerURI:     baseURL,
				TotalItems:   10,
				ItemsPerPage: 100,
				FirstPage:    baseURL + "?page=1",
				PreviousPage: baseURL + "?page=0",
				NextPage:     baseURL + "?page=2",
				CurrentPage:  1,
			},
		},
		{
			"first page and 5 pages",
			args{
				req:       baseURL,
				totalHits: 525,
			},
			&HyperMediaDataSet{
				DataSetURI:   baseURL,
				PagerURI:     baseURL,
				TotalItems:   525,
				ItemsPerPage: 100,
				FirstPage:    baseURL + "?page=1",
				PreviousPage: baseURL + "?page=0",
				NextPage:     baseURL + "?page=2",
				CurrentPage:  1,
			},
		},
		{
			"third page and 5 pages",
			args{
				req:       baseURL + "?page=3",
				totalHits: 525,
			},
			&HyperMediaDataSet{
				DataSetURI:   baseURL,
				PagerURI:     baseURL + "?page=3",
				TotalItems:   525,
				ItemsPerPage: 100,
				FirstPage:    baseURL + "?page=1",
				PreviousPage: baseURL + "?page=2",
				NextPage:     baseURL + "?page=4",
				CurrentPage:  3,
			},
		},
		{
			"third page and 5 pages",
			args{
				req:       baseURL + "?page=3",
				totalHits: 525,
			},
			&HyperMediaDataSet{
				DataSetURI:   baseURL,
				PagerURI:     baseURL + "?page=3",
				TotalItems:   525,
				ItemsPerPage: 100,
				FirstPage:    baseURL + "?page=1",
				PreviousPage: baseURL + "?page=2",
				NextPage:     baseURL + "?page=4",
				CurrentPage:  3,
			},
		},
		{
			"first page with filter",
			args{
				req:       baseURL + "?subject=1&predicate=2&object=3&page=3",
				totalHits: 525,
			},
			&HyperMediaDataSet{
				DataSetURI:   baseURL,
				PagerURI:     baseURL + "?subject=1&predicate=2&object=3&page=3",
				TotalItems:   525,
				ItemsPerPage: 100,
				FirstPage:    baseURL + "?subject=1&predicate=2&object=3&page=1",
				PreviousPage: baseURL + "?subject=1&predicate=2&object=3&page=2",
				NextPage:     baseURL + "?subject=1&predicate=2&object=3&page=4",
				CurrentPage:  3,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r, err := http.NewRequest("GET", tt.args.req, http.NoBody)
			is.NoErr(err)
			r.TLS = &tls.ConnectionState{}

			fr := NewFragmentRequest("test-orgID")
			err = fr.ParseQueryString(r.URL.Query())
			is.NoErr(err)

			got := NewHyperMediaDataSet(r, tt.args.totalHits, fr)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewHyperMediaDataSet() %s mismatch (-want +got):\n%s", tt.name, diff)
				// t.Errorf("NewHyperMediaDataSet() %s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
