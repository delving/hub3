// github.com/delving/hub3/tools/cmd/source_analyzer/processor/collator.go
package processor

import (
	"bufio"
	"fmt"
	"log/slog"
	"sort"

	"github.com/delving/hub3/tools/cmd/source_analyzer/models"
)

// CollationResult contains the results of the collation process
type CollationResult struct {
	NodeRepo    *models.NodeRepo
	UniqueCount int
	SampleSizes []int
}

// Collator processes sorted values to generate statistics
type Collator struct {
	nodeRepo *models.NodeRepo
	samples  map[int]*RandomSample
}

// NewCollator creates a new collator instance
func NewCollator(repo *models.NodeRepo) *Collator {
	return &Collator{
		nodeRepo: repo,
		samples:  make(map[int]*RandomSample),
	}
}

func (c *Collator) writeValue(writer *bufio.Writer, value string, count int) error {
	_, err := fmt.Fprintf(writer, "%7d %s\n", count, value)
	return err
}

func (c *Collator) writeUsefulSamples(uniqueCount int) ([]int, error) {
	var usefulSizes []int
	for size, sample := range c.samples {
		// Always write samples, even if we have fewer unique values
		if err := c.nodeRepo.WriteSample(size, sample.Values()); err != nil {
			return nil, err
		}
		usefulSizes = append(usefulSizes, size)
	}
	return usefulSizes, nil
}

func (c *Collator) Process() (*CollationResult, error) {
	slog.Debug("starting collation",
		"node_dir", c.nodeRepo.BaseDir,
		"compressed", c.nodeRepo.DatasetCtx.CompressOutput)

	// Initialize samples for different sizes
	for _, size := range c.nodeRepo.GetSampleSizes() {
		c.samples[size] = NewRandomSample(size)
	}

	// Create output writers
	countedWriter, countedCloser, err := c.nodeRepo.CreateCountedWriter()
	if err != nil {
		return nil, fmt.Errorf("failed to create counted writer: %v", err)
	}
	defer countedCloser.Close()

	// Process values to count occurrences
	occurrenceMap := make(map[string]int)
	reader, closer, err := c.nodeRepo.CreateValuesReader()
	if err != nil {
		return nil, fmt.Errorf("failed to create values reader: %v", err)
	}
	defer closer.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		value := scanner.Text()
		if value == "" {
			continue
		}
		occurrenceMap[value]++

		// Record in samples
		for _, sample := range c.samples {
			sample.Record(value)
		}
	}

	// Write counted values and build histogram entries
	var entries []models.HistogramEntry
	for value, count := range occurrenceMap {
		// Write to counted file
		if err := c.writeValue(countedWriter, value, count); err != nil {
			return nil, err
		}

		// Add to histogram entries
		entries = append(entries, models.HistogramEntry{
			Value: value,
			Count: count,
		})
	}

	// Sort entries by count descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Count > entries[j].Count
	})

	uniqueCount := len(entries)
	slog.Debug("finished counting values",
		"total_unique", uniqueCount,
		"total_entries", len(entries))

	// Flush counted writer
	if err := countedWriter.Flush(); err != nil {
		return nil, fmt.Errorf("error flushing counted writer: %v", err)
	}

	// Write the full histogram to histogram.txt (possibly compressed)
	if err := c.writeFullHistogram(entries); err != nil {
		return nil, fmt.Errorf("failed to write full histogram: %v", err)
	}

	// Write histogram files based on number of unique values
	standardSizes := []int{100, 500, 2500, 12500}
	for _, size := range standardSizes {
		// Only create histogram if we have values and they fit in the range
		if size >= len(entries) {
			slog.Debug("writing histogram",
				"size", size,
				"actual_entries", len(entries))

			if err := c.nodeRepo.WriteHistogram(size, entries); err != nil {
				return nil, fmt.Errorf("failed to write histogram-%d: %v", size, err)
			}
			break
		}
	}

	// If we have very few entries and none of the standard sizes fit,
	// create a histogram with the actual size
	if len(entries) > 0 && len(entries) < standardSizes[0] {
		slog.Debug("writing small histogram",
			"actual_size", len(entries))

		if err := c.nodeRepo.WriteHistogram(len(entries), entries); err != nil {
			return nil, fmt.Errorf("failed to write small histogram: %v", err)
		}
	}

	// Write samples
	usefulSizes, err := c.writeUsefulSamples(uniqueCount)
	if err != nil {
		return nil, err
	}

	return &CollationResult{
		NodeRepo:    c.nodeRepo,
		UniqueCount: uniqueCount,
		SampleSizes: usefulSizes,
	}, nil
}

// writeFullHistogram writes all histogram entries to a file
func (c *Collator) writeFullHistogram(entries []models.HistogramEntry) error {
	// Create writer for histogram.txt
	histWriter, histCloser, err := c.nodeRepo.CreateHistogramWriter()
	if err != nil {
		return fmt.Errorf("failed to create histogram writer: %v", err)
	}
	defer histCloser.Close()

	// Write all entries
	for _, entry := range entries {
		if _, err := fmt.Fprintf(histWriter, "%7d %s\n", entry.Count, entry.Value); err != nil {
			return fmt.Errorf("failed to write histogram entry: %v", err)
		}
	}

	return histWriter.Flush()
}
