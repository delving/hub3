package hextuples

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestNew(t *testing.T) {
	type args struct {
		b []byte
	}

	// ["https://www.w3.org/People/Berners-Lee/", "http://schema.org/birthDate", "1955-06-08", "http://www.w3.org/2001/XMLSchema#date", "", ""]

	// ["https://www.w3.org/People/Berners-Lee/", "http://schema.org/birthPlace", "http://dbpedia.org/resource/London", "globalId", "", ""]

	tests := []struct {
		name    string
		args    args
		want    HexTuple
		wantErr bool
	}{
		{
			name:    "literal",
			args:    args{b: []byte(`["https://www.w3.org/People/Berners-Lee/", "http://schema.org/birthDate", "1955-06-08", "http://www.w3.org/2001/XMLSchema#date", "", ""]`)},
			want:    HexTuple{hextuple: [6]string{"https://www.w3.org/People/Berners-Lee/", "http://schema.org/birthDate", "1955-06-08", "http://www.w3.org/2001/XMLSchema#date", "", ""}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			opt := cmp.AllowUnexported(HexTuple{})
			if diff := cmp.Diff(tt.want, got, opt); diff != "" {
				t.Errorf("New() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHexTupleFields(t *testing.T) {
	t.Run("date literal", func(t *testing.T) {
		is := is.New(t)
		b := []byte(
			`["https://www.w3.org/People/Berners-Lee/", "http://schema.org/birthDate", "1955-06-08", "http://www.w3.org/2001/XMLSchema#date", "", ""]`,
		)

		ht, err := New(b)
		is.NoErr(err)
		is.Equal(ht.Subject(), "https://www.w3.org/People/Berners-Lee/")
		is.Equal(ht.Predicate(), "http://schema.org/birthDate")
		is.Equal(ht.Value(), "1955-06-08")
		is.Equal(ht.DataType(), "http://www.w3.org/2001/XMLSchema#date")
		is.Equal(ht.Language(), "")
		is.Equal(ht.Graph(), "")
	})
}
