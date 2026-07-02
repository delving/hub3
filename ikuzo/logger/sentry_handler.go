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
	"log/slog"

	"github.com/getsentry/sentry-go"
)

// SentryHandler wraps an slog.Handler and sends errors/warnings/panics to Sentry/GlitchTip.
// It captures Error, Warn, and higher level logs while adding breadcrumbs for Info/Debug logs.
type SentryHandler struct {
	handler slog.Handler
	hub     *sentry.Hub
	attrs   []slog.Attr
	groups  []string
}

// NewSentryHandler creates a new handler that forwards errors to Sentry/GlitchTip.
// The base handler is used for normal logging, while hub is used for error tracking.
func NewSentryHandler(base slog.Handler, hub *sentry.Hub) *SentryHandler {
	if hub == nil {
		hub = sentry.CurrentHub()
	}
	return &SentryHandler{
		handler: base,
		hub:     hub,
		attrs:   make([]slog.Attr, 0),
		groups:  make([]string, 0),
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *SentryHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle processes a log record, forwarding errors to Sentry and all logs to the base handler.
func (h *SentryHandler) Handle(ctx context.Context, r slog.Record) error {
	// First, forward to base handler for normal logging
	if err := h.handler.Handle(ctx, r); err != nil {
		return err
	}

	// Get hub from context if available, otherwise use default
	hub := h.hub
	if ctxHub := sentry.GetHubFromContext(ctx); ctxHub != nil {
		hub = ctxHub
	}

	// Build extra data from attributes
	extra := make(map[string]interface{})
	tags := make(map[string]string)

	// Add handler-level attrs
	for _, attr := range h.attrs {
		h.addAttrToMaps(attr, extra, tags)
	}

	// Add record-level attrs
	r.Attrs(func(a slog.Attr) bool {
		h.addAttrToMaps(a, extra, tags)
		return true
	})

	// Handle based on log level
	switch {
	case r.Level >= slog.LevelError:
		// Capture errors and panics
		event := sentry.NewEvent()
		event.Level = h.slogLevelToSentry(r.Level)
		event.Message = r.Message
		event.Timestamp = r.Time
		event.Extra = extra
		event.Tags = tags

		// Add source location if available
		if r.PC != 0 {
			frames := sentry.NewStacktrace()
			if frames != nil && len(frames.Frames) > 0 {
				event.Exception = []sentry.Exception{{
					Value:      r.Message,
					Type:       "LogError",
					Stacktrace: frames,
				}}
			}
		}

		hub.CaptureEvent(event)

	case r.Level >= slog.LevelWarn:
		// Capture warnings
		event := sentry.NewEvent()
		event.Level = sentry.LevelWarning
		event.Message = r.Message
		event.Timestamp = r.Time
		event.Extra = extra
		event.Tags = tags

		hub.CaptureEvent(event)

	default:
		// Add breadcrumbs for info/debug logs
		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Type:      "log",
			Category:  "log",
			Message:   r.Message,
			Level:     h.slogLevelToSentry(r.Level),
			Timestamp: r.Time,
			Data:      extra,
		}, nil)
	}

	return nil
}

// WithAttrs returns a new handler with additional attributes.
func (h *SentryHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	newAttrs = append(newAttrs, h.attrs...)
	newAttrs = append(newAttrs, attrs...)

	return &SentryHandler{
		handler: h.handler.WithAttrs(attrs),
		hub:     h.hub,
		attrs:   newAttrs,
		groups:  h.groups,
	}
}

// WithGroup returns a new handler with a group name.
func (h *SentryHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	return &SentryHandler{
		handler: h.handler.WithGroup(name),
		hub:     h.hub,
		attrs:   h.attrs,
		groups:  newGroups,
	}
}

// addAttrToMaps adds an attribute to both extra data and tags maps
func (h *SentryHandler) addAttrToMaps(attr slog.Attr, extra map[string]interface{}, tags map[string]string) {
	key := attr.Key

	// Prefix with groups if any
	for _, group := range h.groups {
		key = group + "." + key
	}

	// Add to extra data
	extra[key] = attr.Value.Any()

	// Add to tags if it's a string (Sentry tags must be strings)
	if str, ok := attr.Value.Any().(string); ok {
		tags[key] = str
	}
}

// slogLevelToSentry converts slog.Level to sentry.Level
func (h *SentryHandler) slogLevelToSentry(level slog.Level) sentry.Level {
	switch {
	case level >= slog.LevelError+4: // Panic-like
		return sentry.LevelFatal
	case level >= slog.LevelError:
		return sentry.LevelError
	case level >= slog.LevelWarn:
		return sentry.LevelWarning
	case level >= slog.LevelInfo:
		return sentry.LevelInfo
	default:
		return sentry.LevelDebug
	}
}
