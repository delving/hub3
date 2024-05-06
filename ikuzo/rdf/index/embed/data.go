package embed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/vmihailenco/msgpack/v5"

	proto "google.golang.org/protobuf/proto"
)

type CompressionType int

const (
	Uncompressed CompressionType = iota
	ZSTD
)

type SerializationType int

const (
	JSON SerializationType = iota
	Protobuf
	MsgPack
)

// Raw contains a serialized Data struct as []byte
type Raw []byte

// Data unmarshals Raw into the Data object
func (r Raw) Data() (d Data, err error) {
	err = msgpack.Unmarshal(r, &d)
	if err != nil {
		return d, err
	}

	return d, nil
}

// String returns Raw as a String
func (r Raw) String() string {
	return string(r)
}

// Data is used to embed custom metadata in the index.
// It should be stored as Raw
type Data struct {
	DataModel     string
	Data          []byte
	Compression   CompressionType
	Serialization SerializationType
}

func (d *Data) Raw() (Raw, error) {
	b, err := msgpack.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal Data; %w", err)
	}

	return Raw(b), nil
}

// Marshal v into d.Data
//
// This is based on the Compression and Serialization type
func (d *Data) marshal(v any) (err error) {
	switch d.Serialization {
	case JSON:
		d.Data, err = json.Marshal(&v)
		if err != nil {
			return fmt.Errorf("unable to json.marshal Data.Data; %w", err)
		}
	case MsgPack:
		d.Data, err = msgpack.Marshal(&v)
		if err != nil {
			return fmt.Errorf("unable to msgpack.Marshal Data.Data; %w", err)
		}
	case Protobuf:
		m, ok := v.(proto.Message)
		if !ok {
			return fmt.Errorf("unable to cast as proto.Message: %#v", v)
		}
		d.Data, err = proto.Marshal(m)
		if err != nil {
			return fmt.Errorf("unable to proto.Unmarshal Data.Data; %w", err)
		}

	}

	// compress last
	b, err := d.compress()
	if err != nil {
		return err
	}

	d.Data = b

	return nil
}

// Unmarshal d.Data into the target v.
// This is based on the Compression and Serialization type
func (d *Data) Unmarshal(v any) error {
	b, err := d.uncompress()
	if err != nil {
		return err
	}

	switch d.Serialization {
	case JSON:
		if err := json.Unmarshal(b, &v); err != nil {
			return fmt.Errorf("unable to json.Unmarshal Data.Data; %w", err)
		}
	case MsgPack:
		if err := msgpack.Unmarshal(b, &v); err != nil {
			return fmt.Errorf("unable to msgpack.Unmarshal Data.Data; %w", err)
		}
	case Protobuf:
		m, ok := v.(proto.Message)
		if !ok {
			return fmt.Errorf("unable to cast as proto.Message: %#v", v)
		}
		if err := proto.Unmarshal(b, m); err != nil {
			return fmt.Errorf("unable to proto.Unmarshal Data.Data; %w", err)
		}

	}

	return nil
}

func (d *Data) compress() ([]byte, error) {
	if d.Serialization == SerializationType(Uncompressed) {
		return d.Data, nil
	}

	var buf bytes.Buffer

	zstdWriter, err := zstd.NewWriter(&buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create zstd.Writer; %w", err)
	}

	_, err = zstdWriter.Write(d.Data)
	if err != nil {
		return nil, fmt.Errorf("unable to compress data: %w", err)
	}

	zstdWriter.Close()

	return buf.Bytes(), nil
}

func (d *Data) uncompress() ([]byte, error) {
	if d.Serialization == SerializationType(Uncompressed) {
		return d.Data, nil
	}

	zstdReader, err := zstd.NewReader(bytes.NewReader(d.Data))
	if err != nil {
		return nil, fmt.Errorf("unable to read Data.Data as zstd stream; %w", err)
	}
	defer zstdReader.Close()

	b, err := io.ReadAll(io.Reader(zstdReader))
	if err != nil {
		return nil, fmt.Errorf("unable to decompress Data.Data; %w", err)
	}

	return b, nil
}
