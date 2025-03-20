// cmd/analyzer/main.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/delving/hub3/tools/cmd/source_analyzer/models"
	"github.com/delving/hub3/tools/cmd/source_analyzer/processor"
	"github.com/delving/hub3/tools/cmd/source_analyzer/sip"
	"github.com/spf13/cobra"
)

var (
	// analysis command flags
	inputFile      string
	outputDir      string
	compressOutput bool
	// legacy command flags
	maxUniqueLength int
	elementStep     int
	datasetName     string
)

var rootCmd = &cobra.Command{
	Use:   "analyzer",
	Short: "XML document analyzer and processor",
	Long: `A document analysis tool that processes XML files,
           builds tree structures, and generates statistics about the content.
           Supports both plain and zstd compressed files.`,
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze an XML document",
	Long:  `Analyze an XML document and generate statistical information about its structure and content`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAnalysis()
	},
}

var legacyCmd = &cobra.Command{
	Use:   "legacy",
	Short: "Analyze XML using legacy SIP format",
	Long: `Analyze XML documents using the legacy SIP format.
           Generates detailed statistics about paths, values, and namespaces.
           Compatible with the original Java SIP analyzer.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create configuration
		config := &sip.Config{
			InputFile:            inputFile,
			OutputDir:            outputDir,
			DatasetName:          datasetName,
			MaxUniqueValueLength: maxUniqueLength,
			ElementStep:          elementStep,
			CompressOutput:       compressOutput,
		}

		// Use input filename as dataset name if not specified
		if config.DatasetName == "" {
			config.DatasetName = filepath.Base(inputFile)
		}

		// Create progress listener
		progress := &progressListener{}

		// Create analysis listener
		listener := &analysisListener{
			startTime: time.Now(),
		}

		// Create and run analyzer
		analyzer := sip.NewAnalyzer(config, listener)
		analyzer.SetProgressListener(progress)

		return analyzer.Process()
	},
}

func init() {
	analyzeCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input XML file to analyze (can be .zst compressed)")
	analyzeCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for analysis results")
	analyzeCmd.Flags().BoolVarP(&compressOutput, "compress-output", "c", false, "Compress output files using zstd")

	analyzeCmd.MarkFlagRequired("input")
	analyzeCmd.MarkFlagRequired("output")

	rootCmd.AddCommand(analyzeCmd)

	legacyCmd.Flags().IntVar(&maxUniqueLength, "max-unique-length", 1000, "Maximum length for unique values")
	legacyCmd.Flags().IntVar(&elementStep, "element-step", 10000, "Progress reporting interval")
	legacyCmd.Flags().StringVar(&datasetName, "dataset-name", "", "Dataset name (defaults to input filename)")
	legacyCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input XML file to analyze (can be .zst compressed)")
	legacyCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for analysis results")
	legacyCmd.Flags().BoolVarP(&compressOutput, "compress-output", "c", false, "Compress output files using zstd")

	rootCmd.AddCommand(legacyCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runAnalysis() error {
	// Check if output directory exists and remove it
	if _, err := os.Stat(outputDir); err == nil {
		fmt.Printf("Removing existing output directory: %s\n", outputDir)
		if err := os.RemoveAll(outputDir); err != nil {
			return fmt.Errorf("failed to remove existing output directory: %v", err)
		}
	}

	// Create dataset context with compression settings
	ctx := &models.DatasetContext{
		BaseDir:        outputDir,
		TreeRoot:       filepath.Join(outputDir, "tree"),
		CompressOutput: compressOutput,
	}

	// Ensure output directories exist
	if err := os.MkdirAll(ctx.TreeRoot, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Process the XML file
	_, err := processor.ProcessXML(inputFile, ctx)
	if err != nil {
		return fmt.Errorf("analysis failed: %v", err)
	}

	return nil
}

// progressListener implements sip.ProgressListener
type progressListener struct{}

func (p *progressListener) SetProgress(count int) {
	fmt.Printf("Processed %d elements\n", count)
}

func (p *progressListener) SetProgressMessage(message string) {
	fmt.Printf("Status: %s\n", message)
}

// analysisListener implements sip.AnalysisListener
type analysisListener struct {
	startTime time.Time
}

func (l *analysisListener) Success(stats *sip.Stats) {
	elapsed := time.Since(l.startTime)
	fmt.Printf("\nAnalysis complete:\n")
	fmt.Printf("Dataset: %s\n", stats.Name)
	fmt.Printf("Paths analyzed: %d\n", stats.GetPathCount())
	fmt.Printf("Namespaces found: %d\n", len(stats.Namespaces))
	fmt.Printf("Time taken: %s\n", elapsed.Round(time.Second))
}

func (l *analysisListener) Failure(message string, err error) {
	fmt.Printf("Analysis failed: %s\nError: %v\n", message, err)
}
