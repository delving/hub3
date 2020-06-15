package ginger

import "testing"

func Test_reverseDates(t *testing.T) {
	type args struct {
		date string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"goede volgorde",
			args{"1970-08-13"},
			"1970-08-13",
			false,
		},
		{
			"omgekeerde volgorde",
			args{"13/08/1977"},
			"1970-08-13",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := reverseDates(tt.args.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("reverseDates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("reverseDates() = %v, want %v", got, tt.want)
			}
		})
	}
}
