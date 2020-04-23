package logger

import "github.com/rs/zerolog"

type WrapInfo struct {
	logger *zerolog.Logger
}

func NewWrapInfo(logger *zerolog.Logger) *WrapInfo {
	return &WrapInfo{
		logger: logger,
	}
}

func (w *WrapInfo) Printf(format string, vars ...interface{}) {
	if e := w.logger.Info(); e.Enabled() {
		e.Str("component", "elasticsearch").Msgf(format, vars...)
	}
}

type WrapTrace struct {
	logger *zerolog.Logger
}

func NewWrapTrace(logger *zerolog.Logger) *WrapTrace {
	return &WrapTrace{
		logger: logger,
	}
}

func (w *WrapTrace) Printf(format string, vars ...interface{}) {
	if e := w.logger.Trace(); e.Enabled() {
		e.Str("component", "elasticsearch").Msgf(format, vars...)
	}
}

type WrapDebug struct {
	logger *zerolog.Logger
}

func NewWrapDebug(logger *zerolog.Logger) *WrapDebug {
	return &WrapDebug{
		logger: logger,
	}
}

func (w *WrapDebug) Printf(format string, vars ...interface{}) {
	if e := w.logger.Debug(); e.Enabled() {
		e.Str("component", "elasticsearch").Msgf(format, vars...)
	}
}

type WrapError struct {
	logger *zerolog.Logger
}

func NewWrapError(logger *zerolog.Logger) *WrapError {
	return &WrapError{
		logger: logger,
	}
}

func (w *WrapError) Printf(format string, vars ...interface{}) {
	if e := w.logger.Error(); e.Enabled() {
		e.Str("component", "elasticsearch").Msgf(format, vars...)
	}
}
