package elasticsearch

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
)

func encodeCompositeSearchAfter(sa map[string]interface{}) (string, error) {
	b, err := getInterfaceBytes(sa)
	if err != nil {
		return "", fmt.Errorf("search after encoding error: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func decodeCompositeSearchAfter(input string) (map[string]interface{}, error) {
	var sa map[string]interface{}

	b, err := base64.URLEncoding.DecodeString(input)
	if err != nil {
		return sa, fmt.Errorf("search after decoding error: %w", err)
	}

	if err := getInterface(b, &sa); err != nil {
		return sa, err
	}

	return sa, nil
}

func encodeSearchAfter(sa []interface{}) (string, error) {
	b, err := getInterfaceBytes(sa)
	if err != nil {
		return "", fmt.Errorf("search after encodig error: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func decodeSearchAfter(input string) ([]interface{}, error) {
	var sa []interface{}

	b, err := base64.URLEncoding.DecodeString(input)
	if err != nil {
		return sa, fmt.Errorf("search after decoding error: %w", err)
	}

	if err := getInterface(b, &sa); err != nil {
		return sa, err
	}

	return sa, nil
}

func getInterface(bts []byte, data interface{}) error {
	buf := bytes.NewBuffer(bts)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(data)
	return err
}

func getInterfaceBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
