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

package elasticsearch

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

var (
	ErrAliasNotFound        = errors.New("alias not found")
	ErrAliasAlreadyCreated  = errors.New("alias is already created")
	ErrIndexNotFound        = errors.New("index not found")
	ErrIndexAlreadyCreated  = errors.New("index already created")
	ErrIndexMappingNotValid = errors.New("parsing error in mapping")
)

type ErrorType struct {
	Index       string
	Type        string
	Reason      string
	CauseType   string
	CauseReason string
}

func GetErrorType(r io.Reader) ErrorType {
	json := read(r)
	res := gjson.GetMany(
		json,
		"error.index",
		"error.type",
		"error.reason",
		"error.caused_by.type",
		"error.caused_by.reason",
	)

	et := ErrorType{
		Index:       res[0].String(),
		Type:        res[1].String(),
		Reason:      res[2].String(),
		CauseType:   res[3].String(),
		CauseReason: res[4].String(),
	}

	if et == (ErrorType{}) {
		et.Reason = gjson.Get(json, "error").String()
	}

	return et
}

func (et ErrorType) Error() error {
	// first switch on type
	switch et.Type {
	case "index_not_found_exception":
		return ErrIndexNotFound
	case "aliases_not_found_exception":
		return ErrAliasNotFound
	case "mapper_parsing_exception", "parse_exception":
		log.Error().
			RawJSON("reason", []byte(et.Reason)).
			Str("error_type", et.Type).
			Str("cause_type", et.CauseType).
			Str("cause_reason", et.CauseReason).
			Msg("mapper parsing exception")

		return ErrIndexMappingNotValid
	}

	// switch on raw error as fallback
	if strings.Contains(et.Reason, "alias [") {
		return ErrAliasNotFound
	}

	return fmt.Errorf("unknown error type: %s", et.Type)
}

func read(r io.Reader) string {
	var b bytes.Buffer

	b.ReadFrom(r)

	return b.String()
}
