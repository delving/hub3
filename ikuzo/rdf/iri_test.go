package rdf

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/delving/hub3/ikuzo/validator"
	"github.com/matryer/is"
)

func TestNewIRI(t *testing.T) {
	type args struct {
		iri string
	}

	tests := []struct {
		name    string
		args    args
		want    IRI
		wantErr bool
		err     error
	}{
		{
			name:    "empty IRI",
			args:    args{iri: ""},
			want:    IRI{str: ""},
			wantErr: true,
			err:     ErrEmptyIRI,
		},
		{
			name:    "disallowed character: \n",
			args:    args{iri: "http://dott\ncom"},
			want:    IRI{str: ""},
			wantErr: true,
			err:     ErrDisallowedCharacterInIRI,
		},
		{
			name:    "disallowed character: <",
			args:    args{iri: "<a>"},
			want:    IRI{str: ""},
			wantErr: true,
			err:     ErrDisallowedCharacterInIRI,
		},
		{
			name:    "disallowed character: ' '",
			args:    args{iri: "here are spaces"},
			want:    IRI{str: ""},
			wantErr: true,
			err:     ErrDisallowedCharacterInIRI,
		},
		{
			name:    "valid IRI simple",
			args:    args{iri: "1"},
			want:    IRI{str: "1"},
			wantErr: false,
			err:     nil,
		},
		{
			name:    "valid IRI",
			args:    args{iri: "myscheme://abc/xyz/伝言/æøå#hei?f=88"},
			want:    IRI{str: "myscheme://abc/xyz/伝言/æøå#hei?f=88"},
			wantErr: false,
			err:     nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewIRI(tt.args.iri)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewIRI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && !errors.Is(err, tt.err) {
				t.Errorf("NewIRI() error = %v, wantErr %v", err, tt.err)
				return
			}

			if !got.Equal(tt.want) {
				t.Errorf("NewIRI() mismatch = got %v, want %v", got.str, tt.want.str)
			}
		})
	}
}

func TestIRI_String(t *testing.T) {
	//nolint:gocritic
	is := is.New(t)
	iri, err := NewIRI("urn:123")
	is.NoErr(err)
	is.Equal(iri.RawValue(), iri.str) // RawValue() should return str

	iriStr := iri.String()

	is.True(strings.HasPrefix(iriStr, "<"))
	is.True(strings.HasSuffix(iriStr, ">"))
	is.Equal(iriStr, fmt.Sprintf("<%s>", iri.str))
}

func TestIRI_Equal(t *testing.T) {
	//nolint:gocritic
	is := is.New(t)

	iri, err := NewIRI("urn:123")
	is.NoErr(err)
	is.Equal(iri.RawValue(), iri.str) // RawValue() should return str
	is.Equal(iri.str, "urn:123")      // str should mirror input

	other, err := NewIRI("urn:123")
	is.NoErr(err)
	is.True(other.Equal(iri))
	is.True(iri.Equal(other))

	// test as pointers
	otherPtr := &other
	iriPtr := &iri
	is.True(otherPtr.Equal(iriPtr))
	is.True(iriPtr.Equal(otherPtr))

	noMatch, err := NewIRI("urn:1234")
	is.NoErr(err)
	is.True(!noMatch.Equal(iri))
	is.True(!iri.Equal(noMatch))

	// equal returns false if other is not an IRI
	is.True(!iri.Equal(nonIRI{str: "urn123"}))
}

type nonIRI struct {
	str string
}

func (n nonIRI) String() string                 { return n.str }
func (n nonIRI) RawValue() string               { return n.str }
func (n nonIRI) Equal(Term) bool                { return false }
func (n nonIRI) Type() TermType                 { return TermIRI }
func (n nonIRI) Validate() *validator.Validator { return nil }

func TestIRI_Split(t *testing.T) {
	type fields struct {
		str string
	}

	tests := []struct {
		name       string
		fields     fields
		wantPrefix string
		wantSuffix string
	}{
		{
			name:       "trailing slash",
			fields:     fields{str: "http://example.com/dc/123"},
			wantPrefix: "http://example.com/dc/",
			wantSuffix: "123",
		},
		{
			name:       "trailing #",
			fields:     fields{str: "http://www.w3.org/2004/02/skos/core#Concept"},
			wantPrefix: "http://www.w3.org/2004/02/skos/core#",
			wantSuffix: "Concept",
		},
		{
			name:       "unable to split",
			fields:     fields{str: "urn:123"},
			wantPrefix: "",
			wantSuffix: "",
		},
		{
			name:       "split urn",
			fields:     fields{str: "urn:123/person/4"},
			wantPrefix: "urn:123/person/",
			wantSuffix: "4",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			u := IRI{
				str: tt.fields.str,
			}
			gotPrefix, gotSuffix := u.Split()
			if gotPrefix != tt.wantPrefix {
				t.Errorf("IRI.Split() gotPrefix = %v, want %v", gotPrefix, tt.wantPrefix)
			}
			if gotSuffix != tt.wantSuffix {
				t.Errorf("IRI.Split() gotSuffix = %v, want %v", gotSuffix, tt.wantSuffix)
			}
		})
	}
}
