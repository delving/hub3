// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pelletier/go-toml"

	"github.com/spf13/cobra"

	"github.com/delving/hub3/ikuzo/service/x/adlib"
)

var (
	// adlibCmd represents the adlib command
	adlibCmd = &cobra.Command{
		Use:   "adlib",
		Short: "Harvesting an adlib endpoint.",
	}

	adlibHarvestCmd = &cobra.Command{
		Hidden: false,

		Use:   "harvest",
		Short: "harvest adlib endpoint as XML",

		Run: harvestAdlibXML,
	}

	adlibFrom       string
	adlibURL        string
	adlibDatabase   string
	adlibSearch     string
	adlibHarvestCfg string
	adlibFileName   string
	adlibOutputDir  string
	adlibConfig     string
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	rootCmd.AddCommand(adlibCmd)

	adlibHarvestCmd.Flags().StringVarP(&adlibFrom, "from", "f", "", "timestamp to harvest from")
	adlibHarvestCmd.Flags().StringVarP(&adlibURL, "url", "u", "", "url to harvest from")
	adlibHarvestCmd.Flags().StringVarP(&adlibDatabase, "database", "d", "", "database to harvest from")
	adlibHarvestCmd.Flags().StringVarP(&adlibSearch, "search", "s", "all", "search parameters")
	adlibHarvestCmd.Flags().StringVarP(&adlibConfig, "cfg", "c", "", "path to the harvest toml configuration")
	adlibHarvestCmd.Flags().StringVarP(&adlibFileName, "fname", "", "", "filename to write the output to")
	adlibHarvestCmd.Flags().StringVarP(&adlibOutputDir, "outputDir", "o", "/tmp", "directory to write the filename to")

	adlibCmd.AddCommand(adlibHarvestCmd)
}

// listRecords writes all Records to a file
func harvestAdlibXML(ccmd *cobra.Command, args []string) {
	cfg, err := decodeAdlibConfig(adlibConfig)
	if err != nil {
		slog.Error("unable to decode config", "error", err, "path", harvestCfg)
		return
	}

	if adlibFrom != "" {
		layout := "2006-01-02T15:04:05.999Z"
		parsedTime, err := time.Parse(layout, adlibFrom)
		if err != nil {
			slog.Error("unable to parse timestamp", "error", err)
			return
		}

		cfg.HarvestFrom = parsedTime
	}

	slog.Info("starting adlib harvest", "cfg", cfg)

	timeStart := time.Now()
	if adlibFileName == "" {
		slog.Error("A filename must be set to be able to harvest")
		return
	}

	fname := filepath.Join(adlibOutputDir, adlibFileName)
	if !cfg.HarvestFrom.IsZero() {
		fname += "_incremental"
	}
	outputFname := fname + ".xml"

	file, createErr := os.Create(outputFname)
	if createErr != nil {
		slog.Error("Cannot create file", "error", createErr)
		return
	}
	defer file.Close()

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Caught Interrupt. Closing the file as valid XML.")
		cancel()
		fmt.Fprintln(file, "</pockets>")
		defer file.Close()
		os.Exit(1)
	}()

	client := adlib.New(adlibURL)

	var seen int
	var duplicates int
	fmt.Fprintln(file, "<pockets>")

	unique := map[string]int{}

	cb := func(record *adlib.Record) error {
		if record == nil {
			return fmt.Errorf("cannot process nil *internal.Crecord")
		}
		fmt.Fprintf(file, "<pocket id=\"%s\">\n", record.Attrpriref)
		fmt.Fprintf(file, "<record id=\"%s\">\n", record.Attrpriref)

		_, writeErr := file.Write(record.Raw)
		if writeErr != nil {
			return writeErr
		}

		count, ok := unique[record.Attrpriref]
		if ok {
			count++
			duplicates++
		}

		unique[record.Attrpriref] = count

		seen++
		if seen%500 == 0 {
			slog.Info(
				"harvesting progress",
				"seen", seen, "total", cfg.TotalCount,
				"totalPages", cfg.TotalPages, "pagesSeen", cfg.PagesProcessed,
				"errors", len(cfg.HarvestErrors), "duplicates", duplicates,
				"submittedErrorPages", cfg.ErrorPagesSubmitted,
				"processedErrorPages", cfg.ErrorPagesProcessed,
				"duration", prettyDuration(time.Since(timeStart)),
				"timeRemaining", timeRemaining(timeStart, seen, cfg.TotalCount),
			)
		}
		fmt.Fprintln(file, "</record>")
		fmt.Fprintln(file, "</pocket>")
		return nil
	}

	harvestErr := client.Harvest(ctx, cfg, cb)
	if harvestErr != nil {
		slog.Error("unable to harvest all graphs", "error", harvestErr)
		fmt.Fprintln(file, "</pockets>")
		return
	}

	fmt.Fprintln(file, "</pockets>")

	if len(cfg.HarvestErrors) > 0 {
		var buf bytes.Buffer
		for pageURL, errStr := range cfg.HarvestErrors {
			buf.WriteString(errStr + "\n")
			slog.Error("retrieve errors", "page", pageURL, "error", errStr)
		}
		if writeErr := os.WriteFile(fname+".errors.txt", buf.Bytes(), os.ModePerm); writeErr != nil {
			slog.Error("unable to write error file", "error", writeErr)
		}
	}

	slog.Info(
		"finished harvesting the adlib endpoint",
		"totalHarvested", cfg.RecordsProcessed, "errors", len(cfg.HarvestErrors),
		"apiTotal", cfg.TotalCount, "pagesRetrieved", cfg.PagesProcessed,
		"filename", outputFname, "duplicates", duplicates, "duration", prettyDuration(time.Since(timeStart)),
	)
}

func decodeAdlibConfig(path string) (cfg *adlib.HarvestConfig, err error) {
	if path != "" {
		f, err := os.Open(path)
		if err != nil {
			return cfg, fmt.Errorf("unable to find configuration; %w", err)
		}

		var config adlib.HarvestConfig
		decodeErr := toml.NewDecoder(f).Decode(&config)
		if decodeErr != nil {
			return cfg, fmt.Errorf("unable to decode %s; %w", path, decodeErr)
		}
		return &config, nil
	}

	config := adlib.HarvestConfig{
		TotalCount:       0,
		TotalPages:       0,
		Offset:           0,
		HarvestFrom:      time.Time{},
		HarvestErrors:    map[string]string{},
		PagesProcessed:   0,
		RecordsProcessed: 0,
		Database:         adlibDatabase,
		Search:           adlibSearch,
		Limit:            0,
	}

	return &config, nil
}

func timeRemaining(startTime time.Time, recordsProcessed, totalRecords int) string {
	elapsedTime := time.Since(startTime)

	rate := float64(recordsProcessed) / elapsedTime.Seconds()

	timeRequired := float64(totalRecords) / rate

	timeRequiredDuration := time.Duration(timeRequired) * time.Second

	hours := int(timeRequiredDuration.Hours())
	minutes := int(timeRequiredDuration.Minutes()) % 60
	seconds := int(timeRequiredDuration.Seconds()) % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}
