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

package fragments

import (
	"bytes"
	fmt "fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/delving/hub3/config"
	fuzz "github.com/google/gofuzz"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Naa", func() {
	Context("when parsing a record definition", func() {
		It("should read the whole model", func() {
			path, err := filepath.Abs("./testdata/naa_0.0.5_record-definition.xml")
			Expect(err).ToNot(HaveOccurred())
			f, err := os.Open(path)
			Expect(err).ToNot(HaveOccurred())
			recDef, err := NewRecDef(f)
			Expect(err).ToNot(HaveOccurred())
			fuzz, err := NewFuzzer(recDef)
			Expect(err).ToNot(HaveOccurred())
			nsSize, _ := fuzz.nm.Len()
			Expect(nsSize).ToNot(Equal(0))
			Expect(fuzz.resource).ToNot(BeEmpty())
		})

		It("should fuzz the object", func() {
			f := fuzz.New()
			var ns Cattr
			f.Fuzz(&ns)
			// fmt.Printf("ns => %#v\n", ns)
		})
	})
})

func TestNewRecDef(t *testing.T) {
	// path, err := filepath.Abs("./testdata/naa_0.0.5_record-definition.xml")
	// assert.NoError(t, err, "Unable to create absolute path")
	// raw, err := ioutil.ReadFile(path)
	// assert.NoError(t, err, "Unable to open test data")

	type args struct {
		input io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *Crecord_dash_definition
		wantErr bool
	}{
		{"bad []byte", args{input: bytes.NewReader([]byte{})}, nil, true},
		//{"load full recdef", args{input: raw}, &Crecord_dash_definition{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer GinkgoRecover()
			got, err := NewRecDef(tt.args.input)
			if (err != nil) != tt.wantErr {
				Fail(fmt.Sprintf("NewRecDef() error = %v, wantErr %v", err, tt.wantErr))
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				Fail(fmt.Sprintf("NewRecDef() = %v, want %v", got, tt.want))
			}
		})
	}
}

func TestFuzzer_ExpandNameSpace(t *testing.T) {
	nm := config.NewNamespaceMap()
	defURL := "http://example.com/def/"
	nm.Add("test", defURL)
	type fields struct {
		nm       *config.NamespaceMap
		resource []*FuzzResource
		BaseURL  string
		f        *fuzz.Fuzzer
	}

	type args struct {
		xmlLabel string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			"empty string",
			fields{nm: nm, BaseURL: "http://example.com/def/", f: fuzz.New()},
			args{xmlLabel: ""},
			"", true,
		},
		{
			"bad xml label",
			fields{nm: nm, BaseURL: "http://example.com/def/", f: fuzz.New()},
			args{xmlLabel: "testtitle"},
			"", true,
		},
		{
			"simple namespace",
			fields{nm: nm, BaseURL: "http://example.com/def/", f: fuzz.New()},
			args{xmlLabel: "test:title"},
			fmt.Sprintf("%stitle", defURL), false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer GinkgoRecover()
			fz := &Fuzzer{
				nm:       tt.fields.nm,
				resource: tt.fields.resource,
				BaseURL:  tt.fields.BaseURL,
				f:        tt.fields.f,
			}
			got, err := fz.ExpandNamespace(tt.args.xmlLabel)
			if (err != nil) != tt.wantErr {
				Fail(fmt.Sprintf("Fuzzer.ExpandNameSpace() error = %v, wantErr %v", err, tt.wantErr))
				return
			}
			if got != tt.want {
				Fail(fmt.Sprintf("Fuzzer.ExpandNameSpace() = %v, want %v", got, tt.want))
			}
		})
	}
}
