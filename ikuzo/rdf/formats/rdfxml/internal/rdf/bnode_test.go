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
			Subj: Blank{id: "_:b66765fb9ea2d8811"},
			Pred: IRI{str: "http://example.org/prop1"},
			Obj:  Blank{id: "_:b140008127eafebb0"},
		},
		{
			Subj: Blank{id: "_:b66765fb9ea2d8811"},
			Pred: IRI{str: "http://example.org/prop2"},
			Obj:  Blank{id: "_:abc"},
		},
		{
			Subj: Blank{id: "_:b481d10b23ee2da68"},
			Pred: IRI{str: "http://example.org/prop3"},
			Obj:  Blank{id: "_:b32cb4b4cdf946253"},
		},
		{
			Subj: Blank{id: "_:b481d10b23ee2da68"},
			Pred: IRI{str: "http://example.org/prop3"},
			Obj:  Blank{id: "_:b9e6956d1598d1fec"},
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

// Two sibling node elements without rdf:about or rdf:nodeID are two blank nodes.
// Their path used to hold only element names and depths, so whether two such
// siblings got distinct labels depended on what else was on the path stack: in
// fries-museum P2023-025 (testdata/sibling-blank-concepts.rdf) both blank
// skos:Concept nodes got one label and their triples merged.
func TestSiblingBlankNodesGetDistinctIDs(t *testing.T) {
	for name, open := range map[string]func() (io.Reader, error){
		"minimal": func() (io.Reader, error) {
			return strings.NewReader(`
        <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
                 xmlns:skos="http://www.w3.org/2004/02/skos/core#">
            <skos:Concept><skos:prefLabel>digital born</skos:prefLabel></skos:Concept>
            <skos:Concept rdf:about="http://example.org/c"><skos:prefLabel>video-installatie</skos:prefLabel></skos:Concept>
            <skos:Concept><skos:prefLabel>digitale film</skos:prefLabel></skos:Concept>
        </rdf:RDF>`), nil
		},
		"fries-museum P2023-025": func() (io.Reader, error) { return os.Open("testdata/sibling-blank-concepts.rdf") },
	} {
		t.Run(name, func(t *testing.T) {
			r, err := open()
			if err != nil {
				t.Fatal(err)
			}
			dec := NewRDFXMLDecoder(r, WithSeed("dcn_fries-museum_P2023-025"), WithIDStrategy(DeterministicIDs))
			labels := map[string]string{} // blank concept -> prefLabel
			for {
				triple, err := dec.Decode()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("Decode error: %v", err)
				}
				b, ok := triple.Subj.(Blank)
				if !ok || !strings.HasSuffix(triple.Pred.String(), "skos/core#prefLabel") {
					continue
				}
				if prev, dup := labels[b.id]; dup {
					t.Fatalf("blank node %s carries both %s and %s", b.id, prev, triple.Obj.String())
				}
				labels[b.id] = triple.Obj.String()
			}
			if len(labels) != 2 {
				t.Fatalf("want 2 blank concepts, got %d: %v", len(labels), labels)
			}
		})
	}
}
