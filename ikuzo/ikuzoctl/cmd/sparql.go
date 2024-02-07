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
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/mappingxml"
	"github.com/delving/hub3/ikuzo/service/x/sparql"
)

var (
	// sparqlCmd represents the sparql command
	sparqlCmd = &cobra.Command{
		Use:   "sparql",
		Short: "Harvesting an SPARQL endpoint.",
	}

	harvestCmd = &cobra.Command{
		Hidden: false,

		Use:   "harvest",
		Short: "harvest sparql endpoint as XML",

		Run: harvestXML,
	}

	harvestFrom string
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	rootCmd.AddCommand(sparqlCmd)

	harvestCmd.Flags().StringVarP(&harvestFrom, "from", "f", "", "timestamp to harvest from")

	sparqlCmd.AddCommand(harvestCmd)
}

// listRecords writes all Records to a file
func harvestXML(ccmd *cobra.Command, args []string) {
	cfg := sparql.HarvestConfig{
		URL: "https://eu.api.kleksi.com/apps/pqx31b/datasets/default/sparql",
		Queries: struct {
			NamespacePrefix        string
			WhereClause            string
			SubjectVar             string
			IncrementalWhereClause string
			GetGraphQuery          string
		}{
			NamespacePrefix: `
			PREFIX schema: <https://schema.org/>
			PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>
			`,
			WhereClause: "?s schema:identifier ?identifier .",
			SubjectVar:  "identifier",
			IncrementalWhereClause: `
				?s schema:identifier ?identifier ;
				schema:dateModified ?dateModified .
				FILTER(?dateModified > "~~DATE~~"^^xsd:dateTime)
			`,
			GetGraphQuery: "",
		},
		GraphMimeType: "application/n-triples",
		// MaxSubjects:   200,
		PageSize: 500,
	}

	if harvestFrom != "" {
		layout := "2006-01-02T15:04:05.999Z"
		parsedTime, err := time.Parse(layout, harvestFrom)
		if err != nil {
			slog.Error("unable to parse timestamp", "error", err)
			return
		}

		cfg.From = parsedTime
	}

	slog.Info("starting sparql harvest")

	timeStart := time.Now()

	fname := "/tmp/kleksi"
	if !cfg.From.IsZero() {
		fname += "_incremental"
	}
	file, createErr := os.Create(fname + ".xml")
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
		fmt.Fprintln(file, "</records>")
		defer file.Close()
		os.Exit(1)
	}()

	var seen int
	fmt.Fprintln(file, "<records>")

	cb := func(g *rdf.Graph) error {
		fmt.Fprintf(file, "<record id=\"%s\">\n", g.Subject.RawValue())

		filterCfg := &mappingxml.FilterConfig{Subject: g.Subject}
		_ = filterCfg
		err := mappingxml.Serialize(g, file, filterCfg)
		if err != nil {
			return err
		}
		seen++
		if seen%100 == 0 {
			slog.Info(
				"harvesting progress",
				"seen", seen, "total", cfg.TotalSizeSubjects,
				"max", cfg.MaxSubjects, "errors", len(cfg.HarvestErrors),
				"duration", prettyDuration(time.Since(timeStart)),
			)
		}
		fmt.Fprintln(file, "</record>")
		return nil
	}

	harvestErr := sparql.HarvestGraphs(ctx, &cfg, cb)
	if harvestErr != nil {
		slog.Error("unable to harvest all graphs", "error", harvestErr)
		fmt.Fprintln(file, "</records>")
		return
	}

	totalHarvested := cfg.TotalSizeSubjects

	if len(cfg.HarvestErrors) > 0 {
		cfg.TotalSizeSubjects = len(cfg.HarvestErrors)
		slog.Info("retrying errors", "total", cfg.TotalSizeSubjects)
		harvestErr := sparql.HarvestGraphs(ctx, &cfg, cb)
		if harvestErr != nil {
			slog.Error("unable to harvest all graphs", "error", harvestErr)
			fmt.Fprintln(file, "</records>")
			return
		}
		totalHarvested += cfg.TotalSizeSubjects
	}

	slog.Info("finished harvesting the sparql endpoint", "totalHarvested", totalHarvested, "errors", len(cfg.HarvestErrors))

	fmt.Fprintln(file, "</records>")
}

func prettyDuration(d time.Duration) string {
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour

	hours := d / time.Hour
	d -= hours * time.Hour

	minutes := d / time.Minute
	d -= minutes * time.Minute

	seconds := d / time.Second

	var result string
	if days > 0 {
		result += fmt.Sprintf("%dd ", days)
	}
	if hours > 0 {
		result += fmt.Sprintf("%dh ", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dm ", minutes)
	}
	if seconds > 0 || result == "" {
		result += fmt.Sprintf("%ds", seconds)
	}

	return result
}
