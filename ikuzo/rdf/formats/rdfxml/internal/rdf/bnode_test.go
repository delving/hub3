package rdf

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/delving/hub3/ikuzo/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestPersistentBlankNodeIDs(t *testing.T) {
	input := `
        <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
                 xmlns:ex="http://example.org/">
            <rdf:Description>
                <ex:prop1 rdf:nodeID=""/>
                <ex:prop2 rdf:nodeID="abc"/>
            </rdf:Description>
            <rdf:Description>
                <ex:prop3 rdf:nodeID=""/>
                <ex:prop3 rdf:nodeID=""/>
            </rdf:Description>
        </rdf:RDF>
    `
	expected := []Triple{
		{
			Subj: Blank{id: "_:b1f63bcc868f9561"},
			Pred: IRI{str: "http://example.org/prop1"},
			Obj:  Blank{id: "_:b8b434bd122cbc67e"},
		},
		{
			Subj: Blank{id: "_:b1f63bcc868f9561"},
			Pred: IRI{str: "http://example.org/prop2"},
			Obj:  Blank{id: "_:abc"},
		},
		{
			Subj: Blank{id: "_:b77e9e79d1fcd7e99"},
			Pred: IRI{str: "http://example.org/prop3"},
			Obj:  Blank{id: "_:b98039e712ac3d16"},
		},
		{
			Subj: Blank{id: "_:b77e9e79d1fcd7e99"},
			Pred: IRI{str: "http://example.org/prop3"},
			Obj:  Blank{id: "_:bb4bd705b468ba169"},
		},
	}

	seed := "test_seed"
	dec := NewRDFXMLDecoder(strings.NewReader(input), WithSeed(seed), WithIDStrategy(DeterministicIDs))
	if dec == nil {
		t.Fatal("NewRDFXMLDecoder returned nil")
	}
	dec.SetOption(Base, IRI{str: "http://example.com/"})

	var got []Triple
	for {
		triple, err := dec.Decode()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("Decode error: %v", err)
		}
		got = append(got, triple)
	}

	// Create comparers for the unexported fields
	blankComparer := cmp.Comparer(func(x, y Blank) bool {
		return x.id == y.id
	})

	iriComparer := cmp.Comparer(func(x, y IRI) bool {
		return x.str == y.str
	})

	// Combine the comparers into options
	opts := cmp.Options{
		blankComparer,
		iriComparer,
	}

	if diff := cmp.Diff(expected, got, opts); diff != "" {
		t.Errorf("Unexpected triples (-want +got):\n%s", diff)
	}
}

func Test_rdfXMLDecoder_Files(t *testing.T) {
	tests := []struct {
		name    string
		rdfFile string
		golden  string
		seed    string
		wantErr bool
	}{
		{
			name:    "nave rdf file",
			rdfFile: "./testdata/nave.rdf",
			golden:  "nave.golden.nt",
			seed:    "test_seed_1",
		},
		{
			name:    "crm rdf file",
			rdfFile: "./testdata/crm.rdf",
			golden:  "crm.golden.nt",
			seed:    "test_seed_1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			golden := testutil.NewGoldenHelper(t, "rdfparser")

			f, err := os.Open(tt.rdfFile)
			is.NoErr(err)
			defer f.Close()

			dec := NewRDFXMLDecoder(f, WithSeed(tt.seed), WithIDStrategy(DeterministicIDs))
			triples, err := dec.DecodeAll()
			is.NoErr(err)

			// Serialize the triples to a buffer
			var buf bytes.Buffer
			for _, triple := range triples {
				_, err := buf.WriteString(triple.Serialize(0))
				is.NoErr(err)
			}

			// Compare with golden file
			golden.Compare(tt.golden, buf.Bytes())
			is.True(len(triples) > 0)
		})
	}
}

func TestMain(m *testing.M) {
	testutil.TestMainHelper(m)
}
