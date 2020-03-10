// nolint:gocritic,scopelint
package logger

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
	"github.com/rs/zerolog"
)

func TestNewTestLogger(t *testing.T) {
	is := is.New(t)

	var buf bytes.Buffer

	logger := NewLogger(
		Config{
			testMode: true,
			Output:   &buf,
		},
	)

	want := `{"level":"info","message":"test"}` + "\n"

	logger.Info().Msg("test")
	is.Equal(buf.String(), want)

	buf.Reset()

	logger = NewLogger(Config{})
	logger = logger.Output(&buf)
	logger.Info().Msg("test 2")
	is.True(buf.String() != `{"level":"info","message":"test 2"}`+"\n")
}

func TestNewLogger(t *testing.T) {
	type args struct {
		cfg   Config
		msg   string
		level zerolog.Level
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"info msg",
			args{
				cfg: Config{
					LogLevel:            InfoLevel,
					EnableConsoleLogger: false,
					Output:              nil,
					testMode:            true,
				},
				msg:   "info msg",
				level: zerolog.InfoLevel,
			},
			`{"level":"info","message":"info msg"}`,
		},
		{
			"info console msg",
			args{
				cfg: Config{
					LogLevel:            InfoLevel,
					EnableConsoleLogger: true,
					testMode:            true,
				},
				msg:   "info msg",
				level: zerolog.InfoLevel,
			},
			`INF info msg`,
		},
		{
			"info console with caller",
			args{
				cfg: Config{
					LogLevel:            InfoLevel,
					EnableConsoleLogger: true,
					testMode:            true,
					WithCaller:          true,
				},
				msg:   "caller msg",
				level: zerolog.InfoLevel,
			},
			`INF logger_test.go:`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if tt.args.cfg.Output == nil {
				tt.args.cfg.Output = &buf
			}
			logger := NewLogger(tt.args.cfg)

			logger.WithLevel(tt.args.level).Msg(tt.args.msg)

			if got := buf.String(); !strings.HasPrefix(got, tt.want) {
				t.Errorf("NewLogger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_timeFormat(t *testing.T) {
	is := is.New(t)
	want := "2019-12-26T18:29:36+01:00"
	got := timeFormat(want)
	is.Equal(got, want)

	// no time
	emptyTimestamp := timeFormat(nil)
	is.Equal(emptyTimestamp, "")
}

func TestLevel_toZeroLog(t *testing.T) {
	tests := []struct {
		name string
		l    Level
		want zerolog.Level
	}{
		{
			"debug",
			DebugLevel,
			zerolog.DebugLevel,
		},
		{
			"info",
			InfoLevel,
			zerolog.InfoLevel,
		},
		{
			"warn",
			WarnLevel,
			zerolog.WarnLevel,
		},
		{
			"error",
			ErrorLevel,
			zerolog.ErrorLevel,
		},
		{
			"fatal",
			FatalLevel,
			zerolog.FatalLevel,
		},
		{
			"panic",
			PanicLevel,
			zerolog.PanicLevel,
		},
		{
			"nolevel",
			NoLevel,
			zerolog.NoLevel,
		},
		{
			"disabled",
			Disabled,
			zerolog.Disabled,
		},
		{
			"trace",
			TraceLevel,
			zerolog.TraceLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.toZeroLog(); !cmp.Equal(got, tt.want) {
				t.Errorf("Level.toZeroLog() = %v, want %v", got, tt.want)
			}
		})
	}
}
