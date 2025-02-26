// github.com/delving/hub3/tools/cmd/source_analyzer/processor/sorter.go
package processor

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"sync"

	compression "github.com/delving/hub3/tools/cmd/source_analyzer/io"
	"github.com/delving/hub3/tools/cmd/source_analyzer/models"
)

// SortType defines different sorting strategies
type SortType int

const (
	ValueSort SortType = iota
	HistogramSort
)

// Sorter handles the sorting of values
type Sorter struct {
	nodeRepo   *models.NodeRepo
	sortFiles  []string
	mergeQueue chan *MergeJob
	wg         sync.WaitGroup
}

// MergeJob represents a merge operation between two sorted files
type MergeJob struct {
	FileA    string
	FileB    string
	OutFile  string
	SortType SortType
}

const LinesToSort = 10000

// NewSorter creates a new sorter instance
func NewSorter(repo *models.NodeRepo) *Sorter {
	return &Sorter{
		nodeRepo:   repo,
		mergeQueue: make(chan *MergeJob, 100),
	}
}

// github.com/delving/hub3/tools/cmd/source_analyzer/processor/sorter.go

func (s *Sorter) Sort(sortType SortType) error {
	valuesPath := s.nodeRepo.GetValuesPath()
	slog.Debug("starting sort",
		"path", valuesPath,
		"sort_type", sortType,
		"compressed", s.nodeRepo.DatasetCtx.CompressOutput)

	// Check if values file exists
	if _, err := os.Stat(valuesPath); err != nil {
		slog.Error("values file not found",
			"path", valuesPath,
			"error", err)
		return fmt.Errorf("values file not found: %v", err)
	}

	// Sort in chunks
	if err := s.sortInChunks(sortType); err != nil {
		slog.Error("chunk sorting failed", "error", err)
		return fmt.Errorf("chunk sorting failed: %v", err)
	}

	// Determine output file
	var finalFile string
	if sortType == ValueSort {
		finalFile = s.nodeRepo.GetSortedPath()
	} else {
		finalFile = s.nodeRepo.GetCountedPath()
	}

	slog.Debug("merging sorted chunks",
		"output_file", finalFile,
		"chunk_count", len(s.sortFiles))

	// Merge sorted chunks
	if err := s.mergeSortedFiles(finalFile, sortType); err != nil {
		slog.Error("merging failed", "error", err)
		return fmt.Errorf("merging failed: %v", err)
	}

	return nil
}

func (s *Sorter) sortInChunks(sortType SortType) error {
	// Open values file
	reader, closer, err := s.nodeRepo.CreateValuesReader()
	if err != nil {
		slog.Error("failed to open values file", "error", err)
		return fmt.Errorf("failed to open values file: %v", err)
	}
	defer closer.Close()

	var lines []string
	scanner := bufio.NewScanner(reader)
	chunkNum := 0
	totalLines := 0

	slog.Debug("reading values for sorting")

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		lines = append(lines, line)
		totalLines++

		if len(lines) >= LinesToSort {
			slog.Debug("sorting chunk",
				"chunk", chunkNum,
				"lines", len(lines))
			if err := s.sortAndWriteChunk(lines, chunkNum, sortType); err != nil {
				return err
			}
			lines = lines[:0]
			chunkNum++
		}
	}

	if len(lines) > 0 {
		slog.Debug("sorting final chunk",
			"chunk", chunkNum,
			"lines", len(lines))
		if err := s.sortAndWriteChunk(lines, chunkNum, sortType); err != nil {
			return err
		}
	}

	slog.Debug("finished chunk sorting",
		"total_lines", totalLines,
		"chunks", chunkNum+1)

	if err := scanner.Err(); err != nil {
		slog.Error("error reading values", "error", err)
		return fmt.Errorf("error reading values: %v", err)
	}

	return nil
}

func (s *Sorter) sortAndWriteChunk(lines []string, chunkNum int, sortType SortType) error {
	// Sort based on type
	if sortType == ValueSort {
		sort.Strings(lines)
	} else {
		sort.Sort(sort.Reverse(sort.StringSlice(lines)))
	}

	// Create temporary file
	tmpFile, err := s.nodeRepo.CreateTempFile(fmt.Sprintf("chunk_%d_", chunkNum))
	if err != nil {
		slog.Error("failed to create temp file", "error", err)
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()

	slog.Debug("writing sorted chunk",
		"chunk", chunkNum,
		"path", tmpPath,
		"lines", len(lines))

	// Write sorted lines with compression
	writer, err := compression.CreateCompressedWriter(tmpPath)
	if err != nil {
		os.Remove(tmpPath)
		slog.Error("failed to create compressed writer", "error", err)
		return fmt.Errorf("failed to create compressed writer: %v", err)
	}

	buffWriter := bufio.NewWriter(writer)
	for _, line := range lines {
		if _, err := buffWriter.WriteString(line + "\n"); err != nil {
			writer.Close()
			os.Remove(tmpPath)
			slog.Error("failed to write line", "error", err)
			return fmt.Errorf("failed to write line: %v", err)
		}
	}

	if err := buffWriter.Flush(); err != nil {
		writer.Close()
		os.Remove(tmpPath)
		slog.Error("failed to flush writer", "error", err)
		return fmt.Errorf("failed to flush writer: %v", err)
	}

	if err := writer.Close(); err != nil {
		os.Remove(tmpPath)
		slog.Error("failed to close writer", "error", err)
		return fmt.Errorf("failed to close writer: %v", err)
	}

	s.sortFiles = append(s.sortFiles, tmpPath)
	return nil
}

func (s *Sorter) mergeSortedFiles(finalFile string, sortType SortType) error {
	// If only one file, rename it
	if len(s.sortFiles) == 1 {
		if err := os.Rename(s.sortFiles[0], finalFile); err != nil {
			return err
		}
		s.sortFiles = nil
		return nil
	}

	// Start merger workers
	const numWorkers = 4
	s.wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go s.mergeWorker()
	}

	// Queue merge jobs
	for len(s.sortFiles) > 1 {
		files := s.sortFiles[:2]
		s.sortFiles = s.sortFiles[2:]

		outFile := fmt.Sprintf("%s.merge_%d", finalFile, len(s.sortFiles))
		s.mergeQueue <- &MergeJob{
			FileA:    files[0],
			FileB:    files[1],
			OutFile:  outFile,
			SortType: sortType,
		}

		s.sortFiles = append(s.sortFiles, outFile)
	}

	close(s.mergeQueue)
	s.wg.Wait()

	// Move final merged file
	if len(s.sortFiles) == 1 {
		if err := os.Rename(s.sortFiles[0], finalFile); err != nil {
			return err
		}
		s.sortFiles = nil
		return nil
	}

	return fmt.Errorf("no sorted files remaining")
}

func (s *Sorter) mergeWorker() {
	defer s.wg.Done()

	for job := range s.mergeQueue {
		if err := s.mergeTwoFiles(job); err != nil {
			fmt.Printf("Error merging files: %v\n", err)
			continue
		}

		// Cleanup merged files
		os.Remove(job.FileA)
		os.Remove(job.FileB)
	}
}

// mergeTwoFiles merges two sorted files into one
func (s *Sorter) mergeTwoFiles(job *MergeJob) error {
	// Open input files using compression
	readerA, closerA, err := compression.CreateBufferedReader(job.FileA)
	if err != nil {
		return fmt.Errorf("failed to open file A: %v", err)
	}
	defer closerA.Close()

	readerB, closerB, err := compression.CreateBufferedReader(job.FileB)
	if err != nil {
		return fmt.Errorf("failed to open file B: %v", err)
	}
	defer closerB.Close()

	// Create output file with compression
	writer, err := compression.CreateCompressedWriter(job.OutFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer writer.Close()

	buffWriter := bufio.NewWriter(writer)

	scannerA := bufio.NewScanner(readerA)
	scannerB := bufio.NewScanner(readerB)

	hasA := scannerA.Scan()
	hasB := scannerB.Scan()

	for hasA && hasB {
		lineA := scannerA.Text()
		lineB := scannerB.Text()

		var writeLineA bool
		if job.SortType == ValueSort {
			writeLineA = lineA < lineB
		} else {
			writeLineA = lineA > lineB
		}

		if writeLineA {
			if _, err := buffWriter.WriteString(lineA + "\n"); err != nil {
				return err
			}
			hasA = scannerA.Scan()
		} else {
			if _, err := buffWriter.WriteString(lineB + "\n"); err != nil {
				return err
			}
			hasB = scannerB.Scan()
		}
	}

	// Write remaining lines
	for hasA {
		if _, err := buffWriter.WriteString(scannerA.Text() + "\n"); err != nil {
			return err
		}
		hasA = scannerA.Scan()
	}

	for hasB {
		if _, err := buffWriter.WriteString(scannerB.Text() + "\n"); err != nil {
			return err
		}
		hasB = scannerB.Scan()
	}

	return buffWriter.Flush()
}
