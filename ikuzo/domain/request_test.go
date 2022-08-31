package domain

import "testing"

func Test_SanitizeParam(t *testing.T) {
	type args struct {
		param string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"empty",
			args{},
			"",
		},
		{
			"valid string",
			args{"1.04.02"},
			"1.04.02",
		},
		{
			"path injection",
			args{"../../1.04.02"},
			"1.04.02",
		},
		{
			"leading path injection",
			args{"123/../../1.04.02"},
			"1.04.02",
		},
		{
			". path injection",
			args{"./1.04.02"},
			"1.04.02",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeParam(tt.args.param); got != tt.want {
				t.Errorf("sanitizeParam() %s = '%v', want '%v'", tt.name, got, tt.want)
			}
		})
	}
}
