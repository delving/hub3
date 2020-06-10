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
