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
	"context"
	"fmt"
	"log/slog"
	"time"
)

// SlogAdapter provides a zerolog-compatible API backed by slog.
// This allows gradual migration from zerolog to slog while maintaining
// backward compatibility with existing code.
type SlogAdapter struct {
	slog *slog.Logger
	ctx  context.Context
}

// NewSlogAdapter creates a new zerolog-compatible logger backed by slog.
func NewSlogAdapter(logger *slog.Logger) *SlogAdapter {
	return &SlogAdapter{
		slog: logger,
		ctx:  context.Background(),
	}
}

// WithContext returns a new adapter with the given context.
func (l *SlogAdapter) WithContext(ctx context.Context) *SlogAdapter {
	return &SlogAdapter{
		slog: l.slog,
		ctx:  ctx,
	}
}

// With creates a child logger with additional fields.
func (l *SlogAdapter) With() *SlogContext {
	return &SlogContext{
		logger: l.slog,
		ctx:    l.ctx,
		attrs:  make([]slog.Attr, 0),
	}
}

// Info starts a new info level message.
func (l *SlogAdapter) Info() *SlogEvent {
	return &SlogEvent{
		logger: l.slog,
		ctx:    l.ctx,
		level:  slog.LevelInfo,
		attrs:  make([]slog.Attr, 0),
	}
}

// Error starts a new error level message.
func (l *SlogAdapter) Error() *SlogEvent {
	return &SlogEvent{
		logger: l.slog,
		ctx:    l.ctx,
		level:  slog.LevelError,
		attrs:  make([]slog.Attr, 0),
	}
}

// Warn starts a new warning level message.
func (l *SlogAdapter) Warn() *SlogEvent {
	return &SlogEvent{
		logger: l.slog,
		ctx:    l.ctx,
		level:  slog.LevelWarn,
		attrs:  make([]slog.Attr, 0),
	}
}

// Debug starts a new debug level message.
func (l *SlogAdapter) Debug() *SlogEvent {
	return &SlogEvent{
		logger: l.slog,
		ctx:    l.ctx,
		level:  slog.LevelDebug,
		attrs:  make([]slog.Attr, 0),
	}
}

// Fatal starts a new fatal level message.
func (l *SlogAdapter) Fatal() *SlogEvent {
	return &SlogEvent{
		logger: l.slog,
		ctx:    l.ctx,
		level:  slog.LevelError + 4, // Higher than error
		attrs:  make([]slog.Attr, 0),
		fatal:  true,
	}
}

// Panic starts a new panic level message.
func (l *SlogAdapter) Panic() *SlogEvent {
	return &SlogEvent{
		logger: l.slog,
		ctx:    l.ctx,
		level:  slog.LevelError + 8, // Even higher than fatal
		attrs:  make([]slog.Attr, 0),
	}
}

// SlogContext builds up logger context with additional fields.
type SlogContext struct {
	logger *slog.Logger
	ctx    context.Context
	attrs  []slog.Attr
}

// Str adds a string field to the context.
func (c *SlogContext) Str(key, val string) *SlogContext {
	c.attrs = append(c.attrs, slog.String(key, val))
	return c
}

// Int adds an integer field to the context.
func (c *SlogContext) Int(key string, val int) *SlogContext {
	c.attrs = append(c.attrs, slog.Int(key, val))
	return c
}

// Logger creates a new logger with the accumulated context.
func (c *SlogContext) Logger() *SlogAdapter {
	return &SlogAdapter{
		slog: c.logger.With(convertAttrsToAny(c.attrs)...),
		ctx:  c.ctx,
	}
}

// SlogEvent builds up a log event with fields before emitting it.
type SlogEvent struct {
	logger *slog.Logger
	ctx    context.Context
	level  slog.Level
	attrs  []slog.Attr
	fatal  bool
}

// Str adds a string field to the event.
func (e *SlogEvent) Str(key, val string) *SlogEvent {
	e.attrs = append(e.attrs, slog.String(key, val))
	return e
}

// Int adds an integer field to the event.
func (e *SlogEvent) Int(key string, val int) *SlogEvent {
	e.attrs = append(e.attrs, slog.Int(key, val))
	return e
}

// Int64 adds an int64 field to the event.
func (e *SlogEvent) Int64(key string, val int64) *SlogEvent {
	e.attrs = append(e.attrs, slog.Int64(key, val))
	return e
}

// Bool adds a boolean field to the event.
func (e *SlogEvent) Bool(key string, val bool) *SlogEvent {
	e.attrs = append(e.attrs, slog.Bool(key, val))
	return e
}

// Err adds an error field to the event.
func (e *SlogEvent) Err(err error) *SlogEvent {
	if err != nil {
		e.attrs = append(e.attrs, slog.Any("error", err))
	}
	return e
}

// Dur adds a duration field to the event.
func (e *SlogEvent) Dur(key string, val time.Duration) *SlogEvent {
	e.attrs = append(e.attrs, slog.Duration(key, val))
	return e
}

// Time adds a time field to the event.
func (e *SlogEvent) Time(key string, val time.Time) *SlogEvent {
	e.attrs = append(e.attrs, slog.Time(key, val))
	return e
}

// Interface adds an arbitrary field to the event.
func (e *SlogEvent) Interface(key string, val interface{}) *SlogEvent {
	e.attrs = append(e.attrs, slog.Any(key, val))
	return e
}

// Msg emits the log event with the given message.
func (e *SlogEvent) Msg(msg string) {
	e.logger.LogAttrs(e.ctx, e.level, msg, e.attrs...)
}

// Msgf emits the log event with a formatted message.
func (e *SlogEvent) Msgf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	e.logger.LogAttrs(e.ctx, e.level, msg, e.attrs...)
}

// Send emits the log event with no message.
func (e *SlogEvent) Send() {
	e.Msg("")
}

// convertAttrsToAny converts []slog.Attr to []any for With() method
func convertAttrsToAny(attrs []slog.Attr) []any {
	result := make([]any, len(attrs))
	for i, attr := range attrs {
		result[i] = attr
	}
	return result
}
