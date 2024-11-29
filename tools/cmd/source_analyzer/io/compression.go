// github.com/delving/hub3/tools/cmd/source_analyzer/io/compression.go
package io

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
)

// CompressionType represents supported compression formats
type CompressionType int

const (
	NoCompression CompressionType = iota
	ZstdCompression
	GzipCompression
)

// GetCompressionType determines compression type from file extension
func GetCompressionType(filename string) CompressionType {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".zst", ".zstd":
		return ZstdCompression
	case ".gz", ".gzip":
		return GzipCompression
	default:
		return NoCompression
	}
}

// OpenCompressedFile opens a file and returns an appropriate reader based on compression
func OpenCompressedFile(filename string) (io.ReadCloser, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	compressionType := GetCompressionType(filename)
	switch compressionType {
	case ZstdCompression:
		decoder, err := zstd.NewReader(file)
		if err != nil {
			file.Close()
			return nil, err
		}
		return &zstdReadCloser{
			decoder: decoder,
			file:    file,
		}, nil
	case GzipCompression:
		reader, err := gzip.NewReader(file)
		if err != nil {
			file.Close()
			return nil, err
		}
		return &gzipReadCloser{
			reader: reader,
			file:   file,
		}, nil
	default:
		return file, nil
	}
}

// CreateCompressedWriter creates a writer with optional compression
func CreateCompressedWriter(filename string) (io.WriteCloser, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	compressionType := GetCompressionType(filename)
	switch compressionType {
	case ZstdCompression:
		encoder, err := zstd.NewWriter(file)
		if err != nil {
			file.Close()
			return nil, err
		}
		return &zstdWriteCloser{
			encoder: encoder,
			file:    file,
		}, nil
	case GzipCompression:
		writer := gzip.NewWriter(file)
		return &gzipWriteCloser{
			writer: writer,
			file:   file,
		}, nil
	default:
		return file, nil
	}
}

// zstdReadCloser wraps zstd.Decoder to properly close both decoder and underlying file
type zstdReadCloser struct {
	decoder *zstd.Decoder
	file    *os.File
}

func (z *zstdReadCloser) Read(p []byte) (int, error) {
	return z.decoder.Read(p)
}

func (z *zstdReadCloser) Close() error {
	z.decoder.Close()
	return z.file.Close()
}

// zstdWriteCloser wraps zstd.Encoder to properly close both encoder and underlying file
type zstdWriteCloser struct {
	encoder *zstd.Encoder
	file    *os.File
}

func (z *zstdWriteCloser) Write(p []byte) (int, error) {
	return z.encoder.Write(p)
}

func (z *zstdWriteCloser) Close() error {
	if err := z.encoder.Close(); err != nil {
		z.file.Close()
		return err
	}
	return z.file.Close()
}

// gzipReadCloser wraps gzip.Reader to properly close both reader and underlying file
type gzipReadCloser struct {
	reader *gzip.Reader
	file   *os.File
}

func (g *gzipReadCloser) Read(p []byte) (int, error) {
	return g.reader.Read(p)
}

func (g *gzipReadCloser) Close() error {
	if err := g.reader.Close(); err != nil {
		g.file.Close()
		return err
	}
	return g.file.Close()
}

// gzipWriteCloser wraps gzip.Writer to properly close both writer and underlying file
type gzipWriteCloser struct {
	writer *gzip.Writer
	file   *os.File
}

func (g *gzipWriteCloser) Write(p []byte) (int, error) {
	return g.writer.Write(p)
}

func (g *gzipWriteCloser) Close() error {
	if err := g.writer.Close(); err != nil {
		g.file.Close()
		return err
	}
	return g.file.Close()
}

// CreateBufferedReader creates a buffered reader for a file with optional compression
func CreateBufferedReader(filename string) (*bufio.Reader, io.Closer, error) {
	reader, err := OpenCompressedFile(filename)
	if err != nil {
		return nil, nil, err
	}
	return bufio.NewReader(reader), reader, nil
}

// CreateBufferedWriter creates a buffered writer for a file with optional compression
func CreateBufferedWriter(filename string) (*bufio.Writer, io.Closer, error) {
	writer, err := CreateCompressedWriter(filename)
	if err != nil {
		return nil, nil, err
	}
	return bufio.NewWriter(writer), writer, nil
}
