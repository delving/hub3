package imageproxy

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
			args{key: "http://example.com/123.jpg"},
			&Request{
				cacheKey:  "aHR0cDovL2V4YW1wbGUuY29tLzEyMy5qcGc=",
				sourceURL: "http://example.com/123.jpg",
			},
			false,
		},
		{
			"encoded request",
			args{key: "aHR0cDovL2V4YW1wbGUuY29tLzEyMy5qcGc="},
			&Request{
				cacheKey:  "aHR0cDovL2V4YW1wbGUuY29tLzEyMy5qcGc=",
				sourceURL: "http://example.com/123.jpg",
			},
			false,
		},
		{
			"raw http request with params",
			args{
				key: "http://example.com/123.jpg",
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
				cacheKey:         "aHR0cDovL2V4YW1wbGUuY29tLzEyMy5qcGc=",
				transformOptions: "",
			},
			wantSource:     "ef4/6db/375/aHR0cDovL2V4YW1wbGUuY29tLzEyMy5qcGc=",
			wantDerivative: "",
		},
		{
			name: "cacheKey with transform options",
			fields: fields{
				cacheKey:         "aHR0cDovL2V4YW1wbGUuY29tLzEyMy5qcGc=",
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
