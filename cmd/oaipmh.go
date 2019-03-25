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
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/kiivihal/goharvest/oai"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	pb "gopkg.in/cheggaaa/pb.v1"
)

var (
	// oaipmhCmd represents the oaipmh command
	oaipmhCmd = &cobra.Command{
		Use:   "oaipmh",
		Short: "Harvesting an OAI-PMH endpoint.",
	}

	// identifyCmd subcommand to identify remote OAI-PMH target
	identifyCmd = &cobra.Command{
		Hidden: false,

		Use:   "identify",
		Short: "Identify OAI-PMH response",

		Run: identify,
	}

	// listDataSetsCmd subcommand to list all datasets remote OAI-PMH target
	listDataSetsCmd = &cobra.Command{
		Hidden: false,

		Use:   "datasets",
		Short: "list all available datasets",

		Run: listDatasets,
	}
	// listMetadataFormatsCmd subcommand to list all datasets remote OAI-PMH target
	listMetadataFormatsCmd = &cobra.Command{
		Hidden: false,

		Use:   "formats",
		Short: "list all available metadataformats",

		Run: listMetadataFormats,
	}
	// listIdentifiersCmd subcommand harvest all identifiers to a file
	listIdentifiersCmd = &cobra.Command{
		Hidden: false,

		Use:   "identifiers",
		Short: "harvest all identifiers for a spec and MetadataPrefix",

		Run: listIdentifiers,
	}

	// listRecordsCmd subcommand harvest all Records to a file
	listRecordsCmd = &cobra.Command{
		Hidden: false,

		Use:   "records",
		Short: "harvest all Records for a spec and MetadataPrefix",

		Run: listRecords,
	}

	// listGetRecordCmd subcommand harvest all Records to a file
	listGetRecordCmd = &cobra.Command{
		Hidden: false,

		Use:   "listget",
		Short: "store records listed with the listIdentifiers command and store them individually",

		Run: listGetRecords,
	}

	// getRecordCmd subcommand gets a single records and saves it to a file
	getRecordCmd = &cobra.Command{
		Hidden: false,

		Use:   "record",
		Short: "get a single record for an identifier and a MetadataPrefix",

		Run: getRecord,
	}

	url        string
	verbose    bool
	spec       string
	prefix     string
	identifier string
	outputPath string
)

func init() {
	RootCmd.AddCommand(oaipmhCmd)

	listIdentifiersCmd.Flags().StringVarP(&spec, "spec", "s", "", "The spec of the dataset to be harvested")
	listIdentifiersCmd.Flags().StringVarP(&prefix, "prefix", "p", "", "The metadataPrefix of the dataset to be harvested")
	listRecordsCmd.Flags().StringVarP(&spec, "spec", "s", "", "The spec of the dataset to be harvested")
	listRecordsCmd.Flags().StringVarP(&prefix, "prefix", "p", "", "The metadataPrefix of the dataset to be harvested")

	listGetRecordCmd.Flags().StringVarP(&spec, "spec", "s", "", "The spec of the dataset to be harvested")
	listGetRecordCmd.Flags().StringVarP(&prefix, "prefix", "p", "", "The metadataPrefix of the dataset to be harvested")

	getRecordCmd.Flags().StringVarP(&prefix, "prefix", "p", "", "The metadataPrefix of the record to be harvested")
	getRecordCmd.Flags().StringVarP(&identifier, "identifier", "i", "", "The metadataPrefix of the dataset to be harvested")

	oaipmhCmd.PersistentFlags().StringVarP(&url, "url", "u", "", "URL of the OAI-PMH endpoint (required)")
	oaipmhCmd.PersistentFlags().StringVarP(&outputPath, "output", "o", "", "Output path of the harvested content. Default: current directory")
	oaipmhCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose")

	oaipmhCmd.AddCommand(identifyCmd)
	oaipmhCmd.AddCommand(listDataSetsCmd)
	oaipmhCmd.AddCommand(listMetadataFormatsCmd)
	oaipmhCmd.AddCommand(listIdentifiersCmd)
	oaipmhCmd.AddCommand(listRecordsCmd)
	oaipmhCmd.AddCommand(listGetRecordCmd)
	oaipmhCmd.AddCommand(getRecordCmd)

}

// Print the OAI Response object to stdout
func dump(resp *oai.Response) {
	fmt.Printf("%#v\n", resp.Identify)

}

// identify returns the XML response from a remote OAI-PMH endpoint
func identify(ccmd *cobra.Command, args []string) {
	if url == "" {
		fmt.Println("Error: -u or --url is required and must be a valid URL.")
		return
	}
	req := (&oai.Request{
		BaseURL: url,
		Verb:    "Identify",
	})
	req.Harvest(func(resp *oai.Response) {
		fmt.Printf("%#v\n\n", resp.Identify)
	})
}

// listDataSets returns the datasets from a remote OAI-PMH endpoint
func listDatasets(ccmd *cobra.Command, args []string) {
	req := (&oai.Request{
		BaseURL: url,
		Verb:    "ListSets",
	})
	req.Harvest(func(resp *oai.Response) {
		for idx, set := range resp.ListSets.Set {
			fmt.Printf("\n========= %d =========\n", idx)
			fmt.Printf("Spec\t\t%s\n", set.SetSpec)
			if set.SetName != "None" {
				fmt.Printf("Name:\t\t%s\n", set.SetName)
			}
			if len(set.SetDescription.Body) > 0 && verbose {
				fmt.Printf("Description:\n%s\n", set.SetDescription)
			}
		}
	})
}

// listMetadataFormats returns the available metadataformats from a remote OAI-PMH endpoint
func listMetadataFormats(ccmd *cobra.Command, args []string) {
	req := (&oai.Request{
		BaseURL: url,
		Verb:    "ListMetadataFormats",
	})
	req.Harvest(func(resp *oai.Response) {
		for idx, format := range resp.ListMetadataFormats.MetadataFormat {
			fmt.Printf("\n========= %d =========\n", idx)
			fmt.Printf("prefix:\t\t%s\n", format.MetadataPrefix)
			if verbose {
				fmt.Printf("schema:\t\t%s\n", format.Schema)
				fmt.Printf("namespace:\t%s\n", format.MetadataNamespace)
			}
		}
	})
}

func getPath(fname string) string {
	if outputPath != "" {
		sep := string(os.PathSeparator)
		return fmt.Sprintf("%s%s%s", strings.TrimSuffix(outputPath, sep), sep, fname)
	}
	return fname
}

func getIDs() []string {
	req := (&oai.Request{
		BaseURL:        url,
		Verb:           "ListIdentifiers",
		Set:            spec,
		MetadataPrefix: prefix,
	})
	ids := []string{}
	fname := getPath(fmt.Sprintf("%s_%s_ids.txt", spec, prefix))
	file, err := os.Create(fname)
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()
	seen := 0
	req.HarvestIdentifiers(func(header *oai.Header) {
		seen++
		if seen%250 == 0 {
			fmt.Printf("\rharvested: %d\n", seen)
		}
		fmt.Fprintln(file, header.Identifier)
		ids = append(ids, header.Identifier)
	})
	return ids

}

// listidentifiers writes all identifiers to a file
func listIdentifiers(ccmd *cobra.Command, args []string) {
	getIDs()
	return
}

func getRecord(ccmd *cobra.Command, args []string) {
	storeRecord(identifier, prefix)
}

func storeRecord(identifier string, prefix string) string {
	req := (&oai.Request{
		BaseURL:        url,
		Verb:           "GetRecord",
		MetadataPrefix: prefix,
		Identifier:     identifier,
	})
	file, err := os.Create(getPath(fmt.Sprintf("%s_%s_record.xml", identifier, prefix)))
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	var record string
	req.Harvest(func(r *oai.Response) {
		record = r.GetRecord.Record.Metadata.GoString()
		fmt.Fprintln(file, record)
	})
	return record
}

func listGetRecords(ccmd *cobra.Command, args []string) {
	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)
	ids := make(chan string)

	bar := pb.New(0)
	bar.Start()

	g.Go(func() error {
		defer close(ids)
		seen := 0

		req := (&oai.Request{
			BaseURL:        url,
			Verb:           "ListIdentifiers",
			Set:            spec,
			MetadataPrefix: prefix,
		})
		req.HarvestIdentifiers(func(header *oai.Header) {
			seen++
			//if seen%250 == 0 {
			//fmt.Printf("\rharvested: %d\n", seen)
			//}
			//log.Printf("seen %d: %s\n", seen, header.Identifier)
			select {
			case ids <- header.Identifier:
				//case <-ctx.Done():
				//return ctx.Err()
			}
		})
		//log.Printf("total seen: %d", seen)
		return nil
	})

	//c := make(chan string)
	const numDigesters = 5
	for i := 0; i < numDigesters; i++ {
		g.Go(func() error {
			for id := range ids {
				id := id
				//log.Printf("consuming id: %s", id)
				//time.Sleep(100 * time.Millisecond)
				storeRecord(id, prefix)
				bar.Increment()
				//select {
				//case c <- id:

				//case <-ctx.Done():
				//return ctx.Err()
				//}
			}
			return nil
		})
	}
	go func() {
		g.Wait()
		//close(c)
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
	return
}

// listRecords writes all Records to a file
func listRecords(ccmd *cobra.Command, args []string) {
	req := (&oai.Request{
		BaseURL:        url,
		Verb:           "ListRecords",
		Set:            spec,
		MetadataPrefix: prefix,
	})
	file, err := os.Create(getPath(fmt.Sprintf("%s_%s_records.xml", spec, prefix)))
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Caught Interrupt. Closing the file as valid XML.")
		fmt.Fprintln(file, "</pockets>")
		defer file.Close()
		os.Exit(1)
	}()

	seen := 0
	completeListSize := 0
	bar := pb.New(1000000)

	fmt.Fprintln(file, `<?xml version="1.0" encoding="UTF-8" ?>`)
	fmt.Fprintln(file, "<pockets>")
	req.HarvestRecords(func(r *oai.Record) {
		seen++
		if req.CompleteListSize != 0 && completeListSize == 0 {
			completeListSize = req.CompleteListSize
			bar = pb.New(completeListSize)
			bar.Start()
		}
		bar.Increment()
		fmt.Fprintf(file, `<pocket id="%s">\n`, r.Header.Identifier)
		fmt.Fprintln(file, r.Metadata.GoString())
		fmt.Fprintln(file, "</pocket>")
	})
}
