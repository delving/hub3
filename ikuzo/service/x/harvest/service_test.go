package harvest

import (
	"testing"
	"time"
)

func TestQuery_Valid(t *testing.T) {
	type fields struct {
		From  time.Time
		Until time.Time
	}

	type args struct {
		t time.Time
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"empty from until -> everything is valid",
			fields{},
			args{t: time.Date(2006, 1, 1, 12, 0, 0, 0, time.UTC)},
			true,
		},
		{
			"empty from valid",
			fields{
				Until: time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			args{t: time.Date(2006, 1, 1, 12, 0, 0, 0, time.UTC)},
			true,
		},
		{
			"empty from invalid",
			fields{
				Until: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			args{t: time.Date(2006, 1, 1, 12, 0, 0, 0, time.UTC)},
			false,
		},
		{
			"empty until valid",
			fields{
				From: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			args{t: time.Date(2006, 1, 1, 12, 0, 0, 0, time.UTC)},
			true,
		},
		{
			"empty until invalid",
			fields{
				From: time.Date(2010, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			args{t: time.Date(2006, 1, 1, 12, 0, 0, 0, time.UTC)},
			false,
		},
		{
			"range valid",
			fields{
				From:  time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
				Until: time.Date(2010, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			args{t: time.Date(2006, 1, 1, 12, 0, 0, 0, time.UTC)},
			true,
		},
		{
			"range valid",
			fields{
				From:  time.Date(2005, 1, 1, 12, 0, 0, 0, time.UTC),
				Until: time.Date(2010, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			args{t: time.Date(2006, 1, 1, 12, 0, 0, 0, time.UTC)},
			true,
		},
		{
			"range valid",
			fields{
				From:  time.Date(2007, 1, 1, 12, 0, 0, 0, time.UTC),
				Until: time.Date(2010, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			args{t: time.Date(2006, 1, 1, 12, 0, 0, 0, time.UTC)},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			q := Query{
				From:  tt.fields.From,
				Until: tt.fields.Until,
			}
			if got := q.Valid(tt.args.t); got != tt.want {
				t.Errorf("Query.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
