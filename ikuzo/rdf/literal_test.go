package rdf

import (
	"strings"
	"testing"
	"time"

	"github.com/delving/hub3/ikuzo/validator"
	"github.com/matryer/is"
)

func Test_validateLanguageTag(t *testing.T) {
	type args struct {
		lang string
	}

	tests := []struct {
		name           string
		args           args
		isValid        bool
		errMsgContains string
	}{
		{
			name: "valid single tag",
			args: args{
				lang: "nl",
			},
			isValid:        true,
			errMsgContains: "",
		},
		{
			name: "valid double tag",
			args: args{
				lang: "nl-NL",
			},
			isValid:        true,
			errMsgContains: "",
		},
		{
			name: "valid number in tag",
			args: args{
				lang: "nl-NL1",
			},
			isValid:        true,
			errMsgContains: "",
		},
		{
			name: "invalid number in tag",
			args: args{
				lang: "nl1-NL",
			},
			isValid:        false,
			errMsgContains: "unexpected character",
		},
		{
			name: "leading '-'",
			args: args{
				lang: "-nl",
			},
			isValid:        false,
			errMsgContains: "must start with a letter",
		},
		{
			name: "trailing '-'",
			args: args{
				lang: "nl-",
			},
			isValid:        false,
			errMsgContains: "trailing '-' disallowed",
		},
		{
			name: "trailing '-'",
			args: args{
				lang: "nl-",
			},
			isValid:        false,
			errMsgContains: "trailing '-' disallowed",
		},
		{
			name: "only one '-' allowed",
			args: args{
				lang: "nl-NL-ab",
			},
			isValid:        false,
			errMsgContains: "only one '-' allowed",
		},
		{
			name: "invalid character",
			args: args{
				lang: "nl-行く",
			},
			isValid:        false,
			errMsgContains: "unexpected character",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			validateLanguageTag(v, tt.args.lang)

			if tt.isValid != v.Valid() {
				t.Errorf("validateLanguageTag() got = %v, want %v", v.Valid(), tt.isValid)
			}

			if !v.Valid() && !strings.Contains(v.ErrorOrNil().Error(), tt.errMsgContains) {
				t.Errorf("validateLanguageTag() got = %s, want to contain %v", v.ErrorOrNil().Error(), tt.errMsgContains)
			}
		})
	}
}

func TestLiteral(t *testing.T) {
	t.Run("NewLiteral", func(t *testing.T) {
		// nolint:gocritic
		is := is.New(t)

		l, err := NewLiteral("")
		is.True(err != nil) // empty literal is not allowed
		is.Equal(l, Literal{})

		l, err = NewLiteral("some text")
		is.NoErr(err)
		is.Equal(l.String(), "\"some text\"")
		is.Equal(l.RawValue(), "some text")
	})

	t.Run("Literal.Equal", func(t *testing.T) {
		// nolint:gocritic
		is := is.New(t)

		l, err := NewLiteralInferred("123")
		is.NoErr(err)

		is.True(l.Equal(Literal{str: "123", DataType: xsdString}))

		is.True(!l.Equal(Literal{str: "1234", DataType: xsdString}))

		is.True(!l.Equal(&Literal{str: "123", lang: "nl-NL", DataType: xsdString}))

		is.True(!l.Equal(NewAnonNode())) // different types can never be equal

		is.True(!l.Equal(Literal{str: "123", DataType: xsdBoolean}))

		is.Equal(l.Type(), TermLiteral)
	})

	t.Run("Literal.String", func(t *testing.T) {
		// nolint:gocritic
		is := is.New(t)

		l, err := NewLiteralInferred("123")
		is.NoErr(err)
		is.Equal(l.String(), "\"123\"")

		tests := []struct {
			value string
			lang  string
			dt    *IRI
			want  string
		}{
			{value: "123", want: "\"123\""},
			{value: "123", lang: "nl-NL", want: "\"123\"@nl-NL"},
			{value: "true", dt: xsdBoolean, want: "\"true\"^^<http://www.w3.org/2001/XMLSchema#boolean>"},
		}

		for _, tt := range tests {
			tt := tt

			var l Literal
			var err error

			switch {
			case tt.lang != "":
				l, err = NewLiteralWithLang(tt.value, tt.lang)
				is.NoErr(err) //
			case tt.dt != nil:
				l, err = NewLiteralWithType(tt.value, tt.dt)
				is.NoErr(err)
			default:
				l, err = NewLiteral(tt.value)
				is.NoErr(err)
			}

			is.Equal(l.String(), tt.want)
			is.Equal(l.RawValue(), tt.value)
		}
	})

	t.Run("NewLiteralInferred", func(t *testing.T) {
		tests := []struct {
			input     interface{}
			dt        *IRI
			errString string
		}{
			{1, xsdInteger, ""},
			{int64(1), xsdInteger, ""},
			{int32(1), xsdInteger, ""},
			{3.14, xsdDouble, ""},
			{float32(3.14), xsdDouble, ""},
			{float64(3.14), xsdDouble, ""},
			{time.Now(), xsdDateTime, ""},
			{true, xsdBoolean, ""},
			{false, xsdBoolean, ""},
			{"a", xsdString, ""},
			{[]byte("123"), xsdByte, ""},
			{struct{ a, b string }{"1", "2"}, &IRI{}, `cannot infer XSD datatype from struct { a string; b string }{a:"1", b:"2"}`},
			{"", xsdString, "invalid literal value: cannot be empty"},
		}

		for _, tt := range tests {
			tt := tt

			l, err := NewLiteralInferred(tt.input)
			if err != nil {
				if tt.errString == "" {
					t.Errorf("NewLiteral(%#v) failed with %v; want no error", tt.input, err)
					continue
				}
				if tt.errString != err.Error() {
					t.Errorf("NewLiteral(%#v) failed with %v; want %v", tt.input, err, tt.errString)
					continue
				}
			}
			if err == nil && tt.errString != "" {
				t.Errorf("NewLiteral(%#v) => <no error>; want error %v", tt.input, tt.errString)
				continue
			}

			if err == nil && l.DataType != tt.dt {
				t.Errorf("NewLiteral(%#v).DataType => got %v; want %v", tt.input, l.DataType, tt.dt)
			}
		}
	})

	t.Run("NewLiteralWithType", func(t *testing.T) {
		tests := []struct {
			dataType *IRI
			errWant  string
		}{
			{nil, "cannot be nil"},
			{xsdBoolean, ""},
			{&IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#Unknown"}, "unsupported Literal.DataType IRI"},
		}

		for _, tt := range tests {
			value := "some text"
			_, err := NewLiteralWithType(value, tt.dataType)
			if err != nil {
				if tt.errWant == "" {
					t.Errorf("NewLiteralWithType(%s, %#v) failed with %v; want no error", value, tt.dataType, err)
					continue
				}
				if !strings.Contains(err.Error(), tt.errWant) {
					t.Errorf("NewLiteralWithType(%s, %#v) failed with %v; want %v", value, tt.dataType, err, tt.errWant)
					continue
				}
			}
			if err == nil && tt.errWant != "" {
				t.Errorf("NewLiteralWithType(%s, %#v) => <no error>; want error %v", value, tt.dataType, tt.errWant)
				continue
			}
		}
	})

	t.Run("NewWithLanguage", func(t *testing.T) {
		tests := []struct {
			value   string
			tag     string
			errWant string
		}{
			{"some text", "", ""},
			{"", "en", ""},
			{"", "en-GB", ""},
			{"", "nb-no2", ""},
			{"", "no-no-a", "invalid language tag: only one '-' allowed"},
			{"", "1", "invalid language tag: unexpected character: '1'"},
			{"", "fr-ø", "invalid language tag: unexpected character: 'ø'"},
			{"", "en-", "invalid language tag: trailing '-' disallowed"},
			{"", "-en", "invalid language tag: must start with a letter"},
		}
		for _, tt := range tests {
			if tt.value == "" {
				tt.value = "string"
			}
			l, err := NewLiteralWithLang(tt.value, tt.tag)
			if err != nil {
				if tt.errWant == "" {
					t.Errorf("NewLiteralWithLang(%s, %#v) failed with %v; want no error", tt.value, tt.tag, err)
					continue
				}
				if !strings.Contains(err.Error(), tt.errWant) {
					t.Errorf("NewLiteralWithLang(%s, %#v) failed with %v; want %v", tt.value, tt.tag, err, tt.errWant)
					continue
				}
			}
			if err == nil && tt.errWant != "" {
				t.Errorf("NewLiteralWithLang(%s, %#v) => <no error>; want error %v", tt.value, tt.tag, tt.errWant)
				continue
			}

			if err == nil && tt.tag != l.Lang() {
				t.Errorf("NewLiteralWithLang(%s, %#v) => got %s; want %v", tt.value, tt.tag, l.Lang(), tt.tag)
			}
		}
	})
}

func TestIsValidDataType(t *testing.T) {
	// nolint:gocritic
	is := is.New(t)

	is.True(!isValidDataType(nil)) // nil is not a valid dataType

	is.True(isValidDataType(xsdString))
}

func TestLiteral_String(t *testing.T) {
	type fields struct {
		str      string
		val      interface{}
		lang     string
		DataType *IRI
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"simple", fields{str: "hello"}, "\"hello\""},
		{"with datatype", fields{str: "true", DataType: &IRI{str: "http://www.w3.org/2001/XMLSchema#boolean"}}, "\"true\"^^<http://www.w3.org/2001/XMLSchema#boolean>"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			l := Literal{
				str:      tt.fields.str,
				val:      tt.fields.val,
				lang:     tt.fields.lang,
				DataType: tt.fields.DataType,
			}
			if got := l.String(); got != tt.want {
				t.Errorf("Literal.String() %s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
