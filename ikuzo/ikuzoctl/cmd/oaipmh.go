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
	"strings"
	"syscall"
	"time"

	pb "github.com/cheggaaa/pb/v3"
	"github.com/gocarina/gocsv"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/delving/hub3/ikuzo/service/x/oaipmh"
)

type pmhCfg struct {
	outputPath string
	spec       string
	prefix     string
	identifier string
	url        string
	from       string
	until      string
	idCSV      string
	userName   string
	password   string
	configPath string
	verbose    bool
	storeEAD   bool
}

func addPmhCommonFlags(cmd *cobra.Command, cfg *pmhCfg) {
	cmd.Flags().StringVarP(&cfg.spec, "spec", "s", "", "The spec of the dataset to be harvested")
	cmd.Flags().StringVarP(&cfg.prefix, "prefix", "p", "", "The metadataPrefix of the dataset to be harvested")
	cmd.Flags().StringVarP(&cfg.from, "from", "", "", "from date to be harvested")
	cmd.Flags().StringVarP(&cfg.until, "until", "", "", "until date to be harvested")
}

func NewOaiPmhCmd() *cobra.Command {
	// oaiPmhCmd represents the oaipmh command
	oaiPmhCmd := &cobra.Command{
		Use:   "oaipmh",
		Short: "Harvesting an OAI-PMH endpoint.",
	}

	cfg := &pmhCfg{}

	oaiPmhCmd.PersistentFlags().StringVarP(&cfg.url, "url", "u", "", "URL of the OAI-PMH endpoint (required)")
	oaiPmhCmd.PersistentFlags().StringVarP(&cfg.outputPath, "output", "o", "", "Output path of the harvested content. Default: current directory")
	oaiPmhCmd.PersistentFlags().BoolVarP(&cfg.verbose, "verbose", "v", false, "Verbose")
	oaiPmhCmd.PersistentFlags().StringVarP(&cfg.userName, "username", "", "", "BasicAuth username")
	oaiPmhCmd.PersistentFlags().StringVarP(&cfg.password, "password", "", "", "BasicAuth password")
	oaiPmhCmd.PersistentFlags().StringVarP(&cfg.configPath, "config", "c", "", "config file (default is $HOME/.app.yaml)")

	oaiPmhCmd.AddCommand(identifyCmd(cfg))
	oaiPmhCmd.AddCommand(listIdentifiersCmd(cfg))
	oaiPmhCmd.AddCommand(listRecordsCmd(cfg))
	oaiPmhCmd.AddCommand(listDataSetsCmd(cfg))
	oaiPmhCmd.AddCommand(listGetRecordCmd(cfg))
	oaiPmhCmd.AddCommand(listMetadataFormatsCmd(cfg))
	oaiPmhCmd.AddCommand(getRecordCmd(cfg))

	return oaiPmhCmd
}

// listDataSets returns the datasets from a remote OAI-PMH endpoint
func listDatasets(cfg *pmhCfg) {
	req := (&oaipmh.Request{
		BaseURL:  cfg.url,
		Verb:     "ListSets",
		UserName: cfg.userName,
		Password: cfg.password,
	})
	req.Harvest(func(resp *oaipmh.Response) {
		for idx, set := range resp.ListSets.Set {
			fmt.Printf("\n========= %d =========\n", idx)
			fmt.Printf("Spec\t\t%s\n", set.SetSpec)
			if set.SetName != "None" {
				fmt.Printf("Name:\t\t%s\n", set.SetName)
			}
			if len(set.SetDescription.Body) > 0 && cfg.verbose {
				fmt.Printf("Description:\n%s\n", set.SetDescription)
			}
		}
	})
}

// listMetadataFormats returns the available metadataformats from a remote OAI-PMH endpoint
func listMetadataFormats(cfg *pmhCfg) {
	req := (&oaipmh.Request{
		BaseURL:  cfg.url,
		Verb:     "ListMetadataFormats",
		UserName: cfg.userName,
		Password: cfg.password,
	})
	req.Harvest(func(resp *oaipmh.Response) {
		for idx, format := range resp.ListMetadataFormats.MetadataFormat {
			fmt.Printf("\n========= %d =========\n", idx)
			fmt.Printf("prefix:\t\t%s\n", format.MetadataPrefix)
			if cfg.verbose {
				fmt.Printf("schema:\t\t%s\n", format.Schema)
				fmt.Printf("namespace:\t%s\n", format.MetadataNamespace)
			}
		}
	})
}

func getPath(cfg *pmhCfg, fname string) string {
	if cfg.outputPath != "" {
		sep := string(os.PathSeparator)
		return fmt.Sprintf("%s%s%s", strings.TrimSuffix(cfg.outputPath, sep), sep, fname)
	}

	return fname
}

func getRecord(cfg *pmhCfg) {
	os.MkdirAll(cfg.outputPath, os.ModePerm)
	storeRecord(cfg.identifier, cfg)
}

func storeRecord(id string, cfg *pmhCfg) string {
	req := (&oaipmh.Request{
		BaseURL:        cfg.url,
		Verb:           "GetRecord",
		MetadataPrefix: cfg.prefix,
		Identifier:     id,
		UserName:       cfg.userName,
		Password:       cfg.password,
	})
	var record string
	req.Harvest(func(r *oaipmh.Response) {
		if len(r.Error) > 0 {
			errs := []string{}
			for _, e := range r.Error {
				errs = append(errs, e.Error())
			}
			slog.Error("error harvesting record", "id", id, "errors", errs)
			return
		}

		record = r.GetRecord.Record.Metadata.String()
		file, err := os.Create(getPath(cfg, fmt.Sprintf("%s_%s_record.xml", id, cfg.prefix)))
		if err != nil {
			log.Fatal("Cannot create file", err)
		}
		fmt.Fprintf(file, "<record id=\"%s\">\n", id)
		fmt.Fprintln(file, record)
		fmt.Fprintln(file, "</record>")
	})

	return record
}

type dataIDS struct {
	Identifier string `csv:"identifier"`
}

func idsFromCSV(fname string) ([]dataIDS, error) {
	var ids []dataIDS
	f, err := os.Open(fname)
	if err != nil {
		return ids, err
	}
	defer f.Close()

	if err := gocsv.UnmarshalFile(f, &ids); err != nil { // Load clients from file
		return ids, err
	}

	return ids, nil
}

func listGetRecords(cfg *pmhCfg) {
	ctx := context.Background()
	g, _ := errgroup.WithContext(ctx)
	ids := make(chan string)

	var completeListSize int
	bar := pb.New(completeListSize)
	bar.Start()

	g.Go(func() error {
		defer close(ids)
		seen := 0

		if cfg.idCSV != "" {
			identifiers, err := idsFromCSV(cfg.idCSV)
			if err != nil {
				return err
			}
			bar.SetTotal(int64(len(identifiers)))
			for _, id := range identifiers {
				seen++
				if id.Identifier == "" {
					continue
				}
				ids <- id.Identifier
			}

			return nil
		}

		req := (&oaipmh.Request{
			BaseURL:        cfg.url,
			Verb:           "ListIdentifiers",
			Set:            cfg.spec,
			MetadataPrefix: cfg.prefix,
			From:           cfg.from,
			Until:          cfg.until,
			UserName:       cfg.userName,
			Password:       cfg.password,
		})
		req.HarvestIdentifiers(func(header *oaipmh.Header) {
			if req.CompleteListSize != 0 && completeListSize == 0 {
				completeListSize = req.CompleteListSize
				bar.SetTotal(int64(completeListSize))
			}
			seen++
			// if seen%250 == 0 {
			// fmt.Printf("\rharvested: %d\n", seen)
			// }
			// log.Printf("seen %d: %s\n", seen, header.Identifier)
			ids <- header.Identifier
		})
		bar.SetTotal(int64(seen))
		// log.Printf("total seen: %d", seen)
		return nil
	})

	os.MkdirAll(cfg.outputPath, os.ModePerm)

	// c := make(chan string)
	const numDigesters = 5
	for i := 0; i < numDigesters; i++ {
		g.Go(func() error {
			for id := range ids {
				id := id
				storeRecord(id, cfg)
				bar.Increment()
			}
			return nil
		})
	}
	go func() {
		g.Wait()
	}()

	// Check whether any of the goroutines failed. Since g is accumulating the
	// errors, we don't need to send them (or check for them) in the individual
	// results sent on the channel.
	if err := g.Wait(); err != nil {
		log.Println(err)
		bar.Finish()
		return
	}
	bar.Finish()
}

// listRecords writes all Records to a file
func listRecords(cfg *pmhCfg) {
	req := (&oaipmh.Request{
		BaseURL:        cfg.url,
		Verb:           "ListRecords",
		Set:            cfg.spec,
		MetadataPrefix: cfg.prefix,
		From:           cfg.from,
		Until:          cfg.until,
		UserName:       cfg.userName,
		Password:       cfg.password,
	})

	file, err := os.Create(getPath(cfg, fmt.Sprintf("%s_%s_records.xml", cfg.spec, cfg.prefix)))
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Caught Interrupt. Closing the file as valid XML.")
		fmt.Fprintln(file, "</pockets>")
		defer file.Close()
		slog.Info("finished harvesting", "date", time.Now())
		os.Exit(1)
	}()

	seen := 0
	completeListSize := 0

	bar := progressbar.NewOptions(seen,
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetDescription("Progress:"),
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionShowCount(),
		progressbar.OptionFullWidth(),
	)

	var urls []string
	fmt.Fprintln(file, `<?xml version="1.0" encoding="UTF-8" ?>`)
	fmt.Fprintln(file, "<pockets>")

	req.HarvestRecords(func(r *oaipmh.Record) {
		if req.CompleteListSize != 0 && completeListSize == 0 {
			completeListSize = req.CompleteListSize
			bar.ChangeMax(completeListSize)
		}
		seen++

		var hasChanges bool
		urls, hasChanges = addHarvestURL(urls, req.GetFullURL())
		if hasChanges {
			clearScreen()
			moveCursor(1, 1)
			fmt.Println("Last 10 Urls:")
			printURLs(urls)

			moveCursor(13, 1)
			bar.Set(seen)
			barStr := bar.String()
			fmt.Println(barStr)
		}

		fmt.Fprintf(file, `<pocket id="%s">\n`, r.Header.Identifier)
		body := r.Metadata.String()
		if body == "" {
			fmt.Fprintln(file, r.GoString())
		} else {
			fmt.Fprintln(file, r.Metadata.String())
		}

		fmt.Fprintln(file, "</pocket>")
	})
	fmt.Fprintln(file, "</pockets>")
	bar.Finish()

	slog.Info("finished harvesting", "date", time.Now())
}

// identifyCmd subcommand to identify remote OAI-PMH target
func identifyCmd(cfg *pmhCfg) *cobra.Command {
	identifyCmd := &cobra.Command{
		Hidden: false,
		Use:    "identify",
		Short:  "Identify OAI-PMH response",
		Run: func(cmd *cobra.Command, args []string) {
			identify(cfg)
		},
	}

	return identifyCmd
}

// identify returns the XML response from a remote OAI-PMH endpoint
func identify(cfg *pmhCfg) {
	if cfg.url == "" {
		fmt.Println("Error: -u or --url is required and must be a valid URL.")
		return
	}

	req := (&oaipmh.Request{
		BaseURL:  cfg.url,
		Verb:     "Identify",
		UserName: cfg.userName,
		Password: cfg.password,
	})
	req.Harvest(func(resp *oaipmh.Response) {
		fmt.Printf("%#v\n\n", resp.Identify)
	})
}

// listIdentifiersCmd subcommand harvest all identifiers to a file
func listIdentifiersCmd(cfg *pmhCfg) *cobra.Command {
	listIdentifiersCmd := &cobra.Command{
		Hidden: false,
		Use:    "identifiers",
		Short:  "harvest all identifiers for a spec and MetadataPrefix",
		Run: func(cmd *cobra.Command, args []string) {
			listIdentifiers(cfg)
		},
	}

	addPmhCommonFlags(listIdentifiersCmd, cfg)

	return listIdentifiersCmd
}

// listidentifiers writes all identifiers to a file
func listIdentifiers(cfg *pmhCfg) {
	slog.Info("listIdentifiers", "cfg", cfg)
	getIDs(cfg)
}

func getIDs(cfg *pmhCfg) []string {
	req := (&oaipmh.Request{
		BaseURL:        cfg.url,
		Verb:           "ListIdentifiers",
		Set:            cfg.spec,
		MetadataPrefix: cfg.prefix,
		From:           cfg.from,
		Until:          cfg.until,
		UserName:       cfg.userName,
		Password:       cfg.password,
	})
	ids := []string{}
	fname := getPath(cfg, fmt.Sprintf("%s_%s_ids.txt", cfg.spec, cfg.prefix))

	file, err := os.Create(fname)
	if err != nil {
		log.Fatal("Cannot create file", err)
	}

	defer file.Close()

	seen := 0
	completeListSize := 0
	bar := pb.New(0)
	bar.Start()

	req.HarvestIdentifiers(func(header *oaipmh.Header) {
		if req.CompleteListSize != 0 && completeListSize == 0 {
			completeListSize = req.CompleteListSize
			bar.SetTotal(int64(completeListSize))
		}
		seen++
		bar.Increment()
		fmt.Fprintln(file, header.Identifier)
		ids = append(ids, header.Identifier)
	})

	bar.Finish()

	return ids
}

// listRecordsCmd subcommand harvest all Records to a file
func listRecordsCmd(cfg *pmhCfg) *cobra.Command {
	listRecordsCmd := &cobra.Command{
		Hidden: false,
		Use:    "records",
		Short:  "harvest all Records for a spec and MetadataPrefix",
		Run: func(cmd *cobra.Command, args []string) {
			listRecords(cfg)
		},
	}

	addPmhCommonFlags(listRecordsCmd, cfg)

	return listRecordsCmd
}

// listDataSetsCmd subcommand to list all datasets remote OAI-PMH target
func listDataSetsCmd(cfg *pmhCfg) *cobra.Command {
	cmd := &cobra.Command{
		Hidden: false,
		Use:    "datasets",
		Short:  "list all available datasets",
		Run: func(cmd *cobra.Command, args []string) {
			listDatasets(cfg)
		},
	}

	return cmd
}

// listGetRecordCmd subcommand harvest all Records to a file
func listGetRecordCmd(cfg *pmhCfg) *cobra.Command {
	cmd := &cobra.Command{
		Hidden: false,
		Use:    "listget",
		Short:  "store records listed with the listIdentifiers command and store them individually",
		Run: func(cmd *cobra.Command, args []string) {
			listGetRecords(cfg)
		},
	}

	addPmhCommonFlags(cmd, cfg)
	cmd.Flags().BoolVarP(&cfg.storeEAD, "storeEAD", "", false, "Process and store EAD records")
	cmd.Flags().StringVarP(&cfg.idCSV, "idCSV", "", "", "csv with ids to be harvested")

	return cmd
}

func getRecordCmd(cfg *pmhCfg) *cobra.Command {
	// cmd subcommand gets a single records and saves it to a file
	cmd := &cobra.Command{
		Hidden: false,

		Use:   "record",
		Short: "get a single record for an identifier and a MetadataPrefix",
		Run: func(cmd *cobra.Command, args []string) {
			getRecord(cfg)
		},
	}
	cmd.Flags().StringVarP(&cfg.prefix, "prefix", "p", "", "The metadataPrefix of the record to be harvested")
	cmd.Flags().StringVarP(&cfg.identifier, "identifier", "i", "", "The metadataPrefix of the dataset to be harvested")

	return cmd
}

// listMetadataFormatsCmd subcommand to list all datasets remote OAI-PMH target
func listMetadataFormatsCmd(cfg *pmhCfg) *cobra.Command {
	cmd := &cobra.Command{
		Hidden: false,
		Use:    "formats",
		Short:  "list all available metadataformats",
		Run: func(cmd *cobra.Command, args []string) {
			listMetadataFormats(cfg)
		},
	}

	return cmd
}

// progress bar setup

// clearScreen clears the terminal screen
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// moveCursor moves the cursor to the given position (row, col)
func moveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

// printURLs prints the last n actions
func printURLs(urls []string) {
	for _, harvestURL := range urls {
		fmt.Println(harvestURL)
	}
}

// addHarvestURL adds a new action to the list and keeps only the last 10 actions
func addHarvestURL(urls []string, harvestURL string) (lastUrls []string, isChanged bool) {
	if len(urls) == 0 {
		return append(urls, harvestURL), true
	}

	if harvestURL == urls[len(urls)-1] {
		return urls, false
	}

	if len(urls) >= 10 {
		urls = urls[1:]
	}
	return append(urls, harvestURL), true
}
