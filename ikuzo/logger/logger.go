// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// Level defines log levels.
type Level int8

const (
	// DebugLevel defines debug log level.
	DebugLevel Level = iota
	// InfoLevel defines info log level.
	InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel
	// PanicLevel defines panic log level.
	PanicLevel
	// NoLevel defines an absent log level.
	NoLevel
	// Disabled disables the logger.
	Disabled

	// TraceLevel defines trace log level.
	TraceLevel Level = -1
)

func (l Level) toZeroLog() zerolog.Level {
	var level zerolog.Level

	switch l {
	case DebugLevel:
		level = zerolog.DebugLevel
	case InfoLevel:
		level = zerolog.InfoLevel
	case WarnLevel:
		level = zerolog.WarnLevel
	case ErrorLevel:
		level = zerolog.ErrorLevel
	case FatalLevel:
		level = zerolog.FatalLevel
	case PanicLevel:
		level = zerolog.PanicLevel
	case NoLevel:
		level = zerolog.NoLevel
	case Disabled:
		level = zerolog.Disabled
	case TraceLevel:
		level = zerolog.TraceLevel
	}

	return level
}

func ParseLogLevel(level string) Level {
	switch strings.ToLower(level) {
	case "info":
		return InfoLevel
	case "disabled":
		return Disabled
	case "trace":
		return TraceLevel
	case "debug":
		return DebugLevel
	case "warn":
		return WarnLevel
	case "panic":
		return PanicLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// Config configured the logging.
type Config struct {
	// output overrides default logging to os.Stderr
	Output io.Writer
	// setting with log-level should be written to output. (default: Info)
	LogLevel Level
	// enable colored logging to console, otherwise use JSON logger. (default: false)
	EnableConsoleLogger bool
	// testMode is for testing the log messages.
	// It disables timestamp and hostname logging
	testMode bool
	// WithCaller logs the path and linenumber of the caller.
	// This should not be used in production
	WithCaller bool
	// ErrorFieldName is the label for Go errors in the JSON output. (default: error)
	ErrorFieldName string
}

func Nop() CustomLogger {
	return CustomLogger{zerolog.Nop()}
}

// NewLogger creates zerolog.Logger with sensible defaults
func NewLogger(cfg Config) CustomLogger {
	var output io.Writer
	output = cfg.Output

	if output == nil {
		output = os.Stderr
	}

	if cfg.EnableConsoleLogger {
		output = zerolog.ConsoleWriter{
			Out:             output,
			TimeFormat:      time.RFC3339Nano,
			FormatTimestamp: timeFormat,
			NoColor:         cfg.testMode,
		}
	}

	loggerContext := zerolog.New(output).With()
	if !cfg.testMode {
		loggerContext = loggerContext.Timestamp()
		host, err := os.Hostname()

		if err == nil {
			loggerContext = loggerContext.Str("host", host)
		}
	}

	if cfg.WithCaller {
		loggerContext = loggerContext.Caller()
	}

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.ErrorFieldName = "error.message"

	if cfg.ErrorFieldName != "" {
		zerolog.ErrorFieldName = cfg.ErrorFieldName
	}

	return CustomLogger{loggerContext.Logger().Level(cfg.LogLevel.toZeroLog())}
}

// timeFormat formats the timestamp for the console logger.
func timeFormat(i interface{}) string {
	if i == nil {
		return ""
	}

	return fmt.Sprintf("%s", i)
}
