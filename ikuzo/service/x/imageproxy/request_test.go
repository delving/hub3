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
				CacheKey:  testCacheKey,
				SourceURL: imgURL,
			},
			false,
		},
		{
			"filename too long",
			args{key: "https://service.archief.nl/iipsrv?IIIF=%2F4d%2F68%2F78%2F49%2Ffe%2F2c%2F4f%2F85%2F8b%2F7a%2F97%2F88%2F78%2Fe0%2F2f%2F1d%2F7bb1c07c-14e9-4064-92a0-ffb9fd0132f8.jp2%2Ffull%2F213%2C%2F0%2Fdefault.jpg"},
			&Request{
				CacheKey:  "xxhash64-3a748e637be3be2a",
				SourceURL: "https://service.archief.nl/iipsrv?IIIF=%2F4d%2F68%2F78%2F49%2Ffe%2F2c%2F4f%2F85%2F8b%2F7a%2F97%2F88%2F78%2Fe0%2F2f%2F1d%2F7bb1c07c-14e9-4064-92a0-ffb9fd0132f8.jp2%2Ffull%2F213%2C%2F0%2Fdefault.jpg",
			},
			false,
		},
		{
			"encoded request",
			args{key: testCacheKey},
			&Request{
				CacheKey:  testCacheKey,
				SourceURL: imgURL,
			},
			false,
		},
		{
			"raw http request with params",
			args{
				key: imgURL,
				options: []RequestOption{
					SetRawQueryString("size=200"),
					SetEnableTransform(true),
				},
			},
			&Request{
				CacheKey:        "aHR0cDovL2V4YW1wbGUuY29tLzEyMy5qcGc_c2l6ZT0yMDA=",
				SourceURL:       "http://example.com/123.jpg?size=200",
				RawQueryString:  "size=200",
				EnableTransform: true,
			},
			false,
		},
		{
			"raw http request with params (ADLIB)",
			args{
				key: "http://rabk.adlibhosting.com/wwwopacx/wwwopac.ashx",
				options: []RequestOption{
					SetRawQueryString(`command=getcontent&amp;server=images&amp;value=\kerncollectie\3781.jpg`),
					SetEnableTransform(true),
				},
			},
			&Request{
				CacheKey: `aHR0cDovL3JhYmsuYWRsaWJob3N0aW5nLmNvbS93d3dvcGFjeC93d3dvcGFjLmFzaHg_Y29tbWFuZD1nZXRj` +
					`b250ZW50JnNlcnZlcj1pbWFnZXMmdmFsdWU9XGtlcm5jb2xsZWN0aWVcMzc4MS5qcGc=`,
				SourceURL:       `http://rabk.adlibhosting.com/wwwopacx/wwwopac.ashx?command=getcontent&server=images&value=\kerncollectie\3781.jpg`,
				RawQueryString:  `command=getcontent&amp;server=images&amp;value=\kerncollectie\3781.jpg`,
				EnableTransform: true,
			},
			false,
		},
		{
			"raw",
			args{
				key: imgURL,
				options: []RequestOption{
					SetTransform("raw"),
					SetEnableTransform(true),
				},
			},
			&Request{
				CacheKey:         testCacheKey,
				SourceURL:        imgURL,
				TransformOptions: "raw",
				EnableTransform:  true,
			},
			false,
		},
		{
			"thumbnail",
			args{
				key: imgURL,
				options: []RequestOption{
					SetTransform("500,smartcrop"),
					SetEnableTransform(true),
				},
			},
			&Request{
				CacheKey:         testCacheKey + "_500,smartcrop_tn.jpg",
				SourceURL:        imgURL,
				TransformOptions: "500,smartcrop",
				thumbnailOpts:    "500",
				SubPath:          "_500,smartcrop_tn.jpg",
				EnableTransform:  true,
			},
			false,
		},
		{
			"deepzoom dzi",
			args{
				key: imgURL + ".dzi",
				options: []RequestOption{
					SetTransform("deepzoom"),
					SetEnableTransform(true),
				},
			},
			&Request{
				CacheKey:         testCacheKey + ".dzi",
				SourceURL:        imgURL,
				TransformOptions: "deepzoom",
				SubPath:          ".dzi",
				EnableTransform:  true,
			},
			false,
		},
		{
			"deepzoom tiles",
			args{
				key: imgURL + "_files/9/0_0.jpeg",
				options: []RequestOption{
					SetTransform("deepzoom"),
					SetEnableTransform(true),
				},
			},
			&Request{
				CacheKey:         testCacheKey + "_files/9/0_0.jpeg",
				SourceURL:        imgURL,
				TransformOptions: "deepzoom",
				SubPath:          "_files/9/0_0.jpeg",
				EnableTransform:  true,
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
