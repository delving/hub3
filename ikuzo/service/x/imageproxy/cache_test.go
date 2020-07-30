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

package imageproxy

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	imgURL       = "http://example.com/123.jpg"
	testCacheKey = "aHR0cDovL2V4YW1wbGUuY29tLzEyMy5qcGc="
)

func TestNewRequest(t *testing.T) {
	type args struct {
		key     string
		options []RequestOption
	}

	tests := []struct {
		name    string
		args    args
		want    *Request
		wantErr bool
	}{
		{
			"raw http request",
			args{key: imgURL},
			&Request{
				cacheKey:  testCacheKey,
				sourceURL: imgURL,
			},
			false,
		},
		{
			"encoded request",
			args{key: testCacheKey},
			&Request{
				cacheKey:  testCacheKey,
				sourceURL: imgURL,
			},
			false,
		},
		{
			"raw http request with params",
			args{
				key: imgURL,
				options: []RequestOption{
					SetRawQueryString("size=200"),
				},
			},
			&Request{
				cacheKey:       "aHR0cDovL2V4YW1wbGUuY29tLzEyMy5qcGc_c2l6ZT0yMDA=",
				sourceURL:      "http://example.com/123.jpg?size=200",
				rawQueryString: "size=200",
			},
			false,
		},
		{
			"raw http request with params (ADLIB)",
			args{
				key: "http://rabk.adlibhosting.com/wwwopacx/wwwopac.ashx",
				options: []RequestOption{
					SetRawQueryString(`command=getcontent&amp;server=images&amp;value=\kerncollectie\3781.jpg`),
				},
			},
			&Request{
				cacheKey: `aHR0cDovL3JhYmsuYWRsaWJob3N0aW5nLmNvbS93d3dvcGFjeC93d3dvcGFjLmFzaHg_Y29tbWFuZD1nZXRj` +
					`b250ZW50JnNlcnZlcj1pbWFnZXMmdmFsdWU9XGtlcm5jb2xsZWN0aWVcMzc4MS5qcGc=`,
				sourceURL:      `http://rabk.adlibhosting.com/wwwopacx/wwwopac.ashx?command=getcontent&server=images&value=\kerncollectie\3781.jpg`,
				rawQueryString: `command=getcontent&amp;server=images&amp;value=\kerncollectie\3781.jpg`,
			},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRequest(tt.args.key, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(Request{})); diff != "" {
				t.Errorf("NewRequest() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestRequest_storePaths(t *testing.T) {
	type fields struct {
		cacheKey         string
		transformOptions string
	}

	tests := []struct {
		name           string
		fields         fields
		wantSource     string
		wantDerivative string
	}{
		{
			name: "cacheKey only",
			fields: fields{
				cacheKey:         testCacheKey,
				transformOptions: "",
			},
			wantSource:     "ef4/6db/375/aHR0cDovL2V4YW1wbGUuY29tLzEyMy5qcGc=",
			wantDerivative: "",
		},
		{
			name: "cacheKey with transform options",
			fields: fields{
				cacheKey:         testCacheKey,
				transformOptions: "200x,fit",
			},
			wantSource:     "ef4/6db/375/aHR0cDovL2V4YW1wbGUuY29tLzEyMy5qcGc=",
			wantDerivative: "ef4/6db/375/aHR0cDovL2V4YW1wbGUuY29tLzEyMy5qcGc=#/200x,fit",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			req := &Request{
				cacheKey:         tt.fields.cacheKey,
				transformOptions: tt.fields.transformOptions,
			}

			if diff := cmp.Diff(tt.wantSource, req.sourcePath()); diff != "" {
				t.Errorf("Request.sourcePath() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}

			if diff := cmp.Diff(tt.wantDerivative, req.derivativePath()); diff != "" {
				t.Errorf("Request.derivativePath() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
