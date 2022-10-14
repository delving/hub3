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

package ead

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func getReader(fname string) (*os.File, int64, error) {
	f, err := os.Open(
		filepath.Join("testdata/", fname),
	)
	if err != nil {
		return nil, 0, err
	}

	info, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}

	size := info.Size()

	return f, size, nil
}

func getTestService() (*Service, error) {
	parentDir := os.TempDir()

	eadDir, err := os.MkdirTemp(parentDir, "ead-*")
	if err != nil {
		return nil, err
	}

	return NewService(
		SetDataDir(eadDir),
	)
}

// nolint:gocritic
// func TestService_Process(t *testing.T) {
// is := is.New(t)

// svc, err := getTestService()
// is.NoErr(err)

// // remove test tmpDir
// defer os.RemoveAll(svc.dataDir)

// // make sure it does not run forever
// ctx, cancel := context.WithTimeout(context.Background(), 9*time.Second)
// defer cancel()

// // read test file
// r, size, err := getReader("4.ZHPB2.xml")
// is.NoErr(err)
// is.True(size > 0)

// defer r.Close()

// // remove later. needed for now because of legacy code
// config.InitConfig()

// meta, err := svc.Process(ctx, r, size)
// is.NoErr(err)
// is.Equal(meta.DatasetID, "4.ZHPB2")
// }

// nolint:gocritic
func TestService_SaveEAD(t *testing.T) {
	is := is.New(t)

	svc, err := getTestService()
	is.NoErr(err)

	// remove test tmpDir
	defer os.RemoveAll(svc.dataDir)

	f, size, err := getReader("4.ZHPB2.xml")
	is.NoErr(err)

	meta, err := svc.SaveEAD(f, size, "4.ZHPB2", "demo")
	is.NoErr(err)
	is.Equal(meta.DatasetID, "4.ZHPB2")
}

func TestService_GetName(t *testing.T) {
	type args struct {
		r io.Reader
	}

	fn := func(fname string) io.ReadCloser {
		r, _, err := getReader(fname)
		if err != nil {
			t.Errorf("unable to get reader; %s", err)
		}

		return r
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"empty input",
			args{strings.NewReader("")},
			"",
			true,
		},
		{
			"no eadid",
			args{strings.NewReader("<EAD></EAD>")},
			"",
			true,
		},
		{
			"empty eadid",
			args{strings.NewReader("<EAD><eadid></eadid></EAD>")},
			"",
			true,
		},
		{
			"ead file",
			args{fn("4.ZHPB2.xml")},
			"4.ZHPB2",
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			s := &Service{}

			var buf bytes.Buffer
			_, err := io.Copy(&buf, tt.args.r)
			if err != nil {
				t.Errorf("Service.GetName() error = %v, wantErr %v", err, tt.wantErr)
			}

			got, err := s.GetName(buf.Bytes())
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.GetName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Service.GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_addHeader(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			"no header",
			args{b: []byte(`<ead audience="external">
          <eadheader `)},
			[]byte(`<?xml version="1.0" encoding="UTF-8"?><ead audience="external">
          <eadheader `),
		},
		{
			"with leading white-space",
			args{b: []byte(`            <ead audience="external">
          <eadheader `)},
			[]byte(`<?xml version="1.0" encoding="UTF-8"?>            <ead audience="external">
          <eadheader `),
		},
		{
			"with header",
			args{b: []byte(`<?xml version="1.0" encoding="UTF-8"?><ead audience="external">
          <eadheader `)},
			[]byte(`<?xml version="1.0" encoding="UTF-8"?><ead audience="external">
          <eadheader `),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got := addHeader(tt.args.b)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("addHeader() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
