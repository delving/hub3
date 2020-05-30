package ead

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

	eadDir, err := ioutil.TempDir(parentDir, "ead-*")
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

	r, meta, err := svc.SaveEAD(f, size)
	is.NoErr(err)
	is.Equal(meta.DatasetID, "4.ZHPB2")
	is.True(strings.HasPrefix(meta.basePath, svc.dataDir))
	is.True(r != nil)

	info, err := os.Stat(filepath.Join(meta.basePath, fmt.Sprintf("%s.xml", meta.DatasetID)))
	is.NoErr(err)
	is.Equal(info.Size(), size)
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

			got, err := s.GetName(&buf)
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
