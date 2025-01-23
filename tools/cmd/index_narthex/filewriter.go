package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/klauspost/compress/zstd"
)

type FileWriter struct {
	nTripleWriter io.WriteCloser
	bulkWriter    io.WriteCloser
	nTripleFile   *os.File
	bulkFile      *os.File
	mu            sync.Mutex
}

func NewFileWriter(outputDir, prefix string) (*FileWriter, error) {
	nTriplePath := fmt.Sprintf("%s__rdf.nt.zst", prefix)
	bulkPath := fmt.Sprintf("%s__index.jsonl.zst", prefix)

	slog.Info("output files", "ntriples", nTriplePath, "bulk", bulkPath, "prefix", prefix)

	ntFile, err := os.Create(filepath.Join(outputDir, nTriplePath))
	if err != nil {
		return nil, fmt.Errorf("create ntriples file: %w", err)
	}

	bFile, err := os.Create(filepath.Join(outputDir, bulkPath))
	if err != nil {
		ntFile.Close()
		return nil, fmt.Errorf("create bulk file: %w", err)
	}

	ntWriter, err := zstd.NewWriter(ntFile)
	if err != nil {
		ntFile.Close()
		bFile.Close()
		return nil, fmt.Errorf("create ntriples zstd writer: %w", err)
	}

	bWriter, err := zstd.NewWriter(bFile)
	if err != nil {
		ntWriter.Close()
		ntFile.Close()
		bFile.Close()
		return nil, fmt.Errorf("create bulk zstd writer: %w", err)
	}

	return &FileWriter{
		nTripleWriter: ntWriter,
		bulkWriter:    bWriter,
		nTripleFile:   ntFile,
		bulkFile:      bFile,
	}, nil
}

func (w *FileWriter) WriteNTriples(g *rdf.Graph) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := ntriples.Serialize(g, w.nTripleWriter); err != nil {
		return fmt.Errorf("unable to serialize ntriples: %w", err)
	}

	return nil
}

func (w *FileWriter) WriteBulk(record []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	_, err := w.bulkWriter.Write(record)
	if err != nil {
		return fmt.Errorf("write bulk: %w", err)
	}

	return nil
}

func (w *FileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.nTripleWriter.Close(); err != nil {
		return fmt.Errorf("close ntriples writer: %w", err)
	}
	if err := w.bulkWriter.Close(); err != nil {
		return fmt.Errorf("close bulk writer: %w", err)
	}
	if err := w.nTripleFile.Close(); err != nil {
		return fmt.Errorf("close ntriples file: %w", err)
	}
	if err := w.bulkFile.Close(); err != nil {
		return fmt.Errorf("close bulk file: %w", err)
	}
	return nil
}
