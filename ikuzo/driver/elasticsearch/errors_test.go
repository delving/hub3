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

package elasticsearch

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_GetErrorType(t *testing.T) {
	type args struct {
		r io.Reader
	}

	tests := []struct {
		name string
		args args
		want ErrorType
	}{
		{
			"missing fields",
			args{
				strings.NewReader(
					`
				{
					"error": {
						"type": "index_not_found_exception",
						"resource.type": "index_or_alias",
						"resource.id": "hub3test",
						"index_uuid": "_na_",
					},
					"status": 404
					}
				`,
				),
			},
			ErrorType{
				Type: "index_not_found_exception",
			},
		},
		{
			"index does not exist error",
			args{
				strings.NewReader(
					`
				{
					"error": {
						"root_cause": [
						{
							"type": "index_not_found_exception",
							"reason": "no such index [hub3test]",
							"resource.type": "index_or_alias",
							"resource.id": "hub3test",
							"index_uuid": "_na_",
							"index": "hub3test"
						}
						],
						"type": "index_not_found_exception",
						"reason": "no such index [hub3test]",
						"resource.type": "index_or_alias",
						"resource.id": "hub3test",
						"index_uuid": "_na_",
						"index": "hub3test"
					},
					"status": 404
					}
				`,
				),
			},
			ErrorType{
				Index:  "hub3test",
				Type:   "index_not_found_exception",
				Reason: "no such index [hub3test]",
			},
		},
		{
			"empty json",
			args{strings.NewReader("")},
			ErrorType{},
		},
		{
			"alias missing",
			args{strings.NewReader(`
				{"error":"alias [unknownalias] missing","status":404}
				`)},
			ErrorType{
				Reason: "alias [unknownalias] missing",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got := GetErrorType(tt.args.r)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getError() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestGetErrorType(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name string
		args args
		want ErrorType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetErrorType(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetErrorType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorType_Error(t *testing.T) {
	type fields struct {
		Type   string
		Reason string
	}

	tests := []struct {
		name    string
		fields  fields
		errType error
	}{
		{
			"missing index",
			fields{Reason: "alias [unknownalias] missing"},
			ErrAliasNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			et := ErrorType{
				Type:   tt.fields.Type,
				Reason: tt.fields.Reason,
			}
			if !errors.Is(et.Error(), tt.errType) {
				t.Errorf("ErrorType.Error() error = %v, wantErr %v", et.Error(), tt.errType)
			}
		})
	}
}
