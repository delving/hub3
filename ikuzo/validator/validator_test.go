package validator_test

import (
	"regexp"
	"testing"

	"github.com/delving/hub3/ikuzo/validator"
	"github.com/matryer/is"
)

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func TestValidator(t *testing.T) {
	// nolint:gocritic
	is := is.New(t)

	v := validator.New()
	is.True(v != nil) // a new validator should not be nil

	is.True(v.Valid()) // an empty validator should always be valid

	v.AddError("triple", "is not valid")
	is.Equal(v.Valid(), false) // an error is added so it should be invalid
	is.Equal(len(v.Errors), 1) // only one error should have been added

	v.AddError("triple", "new error")
	is.Equal(len(v.Errors), 1) // only one error should have been added
	errMsg, ok := v.Errors["triple"]
	is.True(ok)
	is.Equal(errMsg, "is not valid")

	v.Check(1 > 0, "page", "must be greater than zero")
	is.Equal(len(v.Errors), 1) // no error should have been added

	v.Check(-1 > 0, "page", "must be greater than zero")
	errMsg, ok = v.Errors["page"]
	is.True(ok)
	is.Equal(errMsg, "must be greater than zero")
}

func TestMatches(t *testing.T) {
	type args struct {
		value string
		rx    *regexp.Regexp
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid match",
			args: args{value: "me@example.com", rx: EmailRX},
			want: true,
		},
		{
			name: "invalid match",
			args: args{value: "(me)@example.com", rx: EmailRX},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := validator.Matches(tt.args.value, tt.args.rx); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIn(t *testing.T) {
	type args struct {
		value string
		list  []string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty list",
			args: args{value: "1"},
			want: false,
		},
		{
			name: "no match",
			args: args{value: "1", list: []string{"2", "3"}},
			want: false,
		},
		{
			name: "match",
			args: args{value: "3", list: []string{"2", "3"}},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := validator.In(tt.args.value, tt.args.list...); got != tt.want {
				t.Errorf("In() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnique(t *testing.T) {
	type args struct {
		values []string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty list",
			args: args{},
			want: true,
		},
		{
			name: "unique list",
			args: args{values: []string{"1", "2", "3"}},
			want: true,
		},
		{
			name: "non-unique list",
			args: args{values: []string{"1", "2", "3", "1"}},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := validator.Unique(tt.args.values); got != tt.want {
				t.Errorf("Unique() = %v, want %v", got, tt.want)
			}
		})
	}
}
