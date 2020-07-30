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

package search

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBreadCrumbBuilder_GetLast(t *testing.T) {
	type fields struct {
		hrefPath []string
		crumbs   []*BreadCrumb
	}

	tests := []struct {
		name   string
		fields fields
		want   *BreadCrumb
	}{
		{
			"empty list of breadcrumbs",
			fields{},
			nil,
		},
		{
			"single breadcrumb",
			fields{crumbs: []*BreadCrumb{{Field: "last"}}},
			&BreadCrumb{Field: "last"},
		},
		{
			"list of breadcrumbs",
			fields{crumbs: []*BreadCrumb{
				{Field: "first"},
				{Field: "last"},
			}},
			&BreadCrumb{Field: "last"},
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			bcb := &BreadCrumbBuilder{
				hrefPath: tt.fields.hrefPath,
				crumbs:   tt.fields.crumbs,
			}
			got := bcb.GetLast()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("BreadCrumbBuilder.GetLast() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBreadCrumbBuilder_GetPath(t *testing.T) {
	type fields struct {
		hrefPath []string
		crumbs   []*BreadCrumb
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"empty href path",
			fields{},
			"",
		},
		{
			"single path",
			fields{hrefPath: []string{"q=test"}},
			"q=test",
		},
		{
			"double path",
			fields{
				hrefPath: []string{
					"q=Супрематизм» Suprematism",
					"qf=dc_creator:malevich",
				},
			},
			"q=Супрематизм» Suprematism&qf=dc_creator:malevich",
		},
		{
			"2 or more",
			fields{
				hrefPath: []string{
					"q=Супрематизм» Suprematism",
					"qf=dc_creator:malevich",
					"qf=dc_date:1915",
				},
			},
			"q=Супрематизм» Suprematism&qf=dc_creator:malevich&qf=dc_date:1915",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			bcb := &BreadCrumbBuilder{
				hrefPath: tt.fields.hrefPath,
				crumbs:   tt.fields.crumbs,
			}
			got := bcb.GetPath()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("BreadCrumbBuilder.GetPath() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
