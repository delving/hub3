package elasticsearch

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/tidwall/gjson"
)

var (
	ErrAliasNotFound       = errors.New("alias not found")
	ErrAliasAlreadyCreated = errors.New("alias is already created")
	ErrIndexNotFound       = errors.New("index not found")
	ErrIndexAlreadyCreated = errors.New("index already created")
)

type ErrorType struct {
	Index  string
	Type   string
	Reason string
}

func GetErrorType(r io.Reader) ErrorType {
	json := read(r)
	res := gjson.GetMany(
		json,
		"error.index",
		"error.type",
		"error.reason",
	)

	log.Printf("json error: %s", json)

	et := ErrorType{
		Index:  res[0].String(),
		Type:   res[1].String(),
		Reason: res[2].String(),
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
	}

	// switch on raw error as fallback
	switch {
	case strings.HasPrefix(et.Reason, "alias ["):
		return ErrAliasNotFound
	}

	return fmt.Errorf("unknown error type: %s", et.Type)
}

func read(r io.Reader) string {
	var b bytes.Buffer

	b.ReadFrom(r)

	return b.String()
}
