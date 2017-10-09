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
	"fmt"
	"log"
	"os"

	"github.com/renevanderark/goharvest/oai"
	"github.com/spf13/cobra"
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

	url     string
	verbose bool
	spec    string
	prefix  string
)

func init() {
	RootCmd.AddCommand(oaipmhCmd)

	listIdentifiersCmd.Flags().StringVarP(&spec, "spec", "s", "", "The spec of the dataset to be harvested")
	listIdentifiersCmd.Flags().StringVarP(&prefix, "prefix", "p", "", "The metadataPrefix of the dataset to be harvested")

	oaipmhCmd.PersistentFlags().StringVarP(&url, "url", "u", "", "URL of the OAI-PMH endpoint (required)")
	oaipmhCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose")

	oaipmhCmd.AddCommand(identifyCmd)
	oaipmhCmd.AddCommand(listDataSetsCmd)
	oaipmhCmd.AddCommand(listMetadataFormatsCmd)
	oaipmhCmd.AddCommand(listIdentifiersCmd)
}

// Print the OAI Response object to stdout
func dump(resp *oai.Response) {
	fmt.Printf("%#v\n", resp.Identify)

}

// identify returns the XML response from a remote OAI-PMH endpoint
func identify(ccmd *cobra.Command, args []string) {
	fmt.Println(url)
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

// listidentifiers writes all identifiers to a file
func listIdentifiers(ccmd *cobra.Command, args []string) {
	req := (&oai.Request{
		BaseURL:        url,
		Verb:           "ListIdentifiers",
		Set:            spec,
		MetadataPrefix: prefix,
	})

	file, err := os.Create(fmt.Sprintf("%s_%s_ids.txt", spec, prefix))
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
	})

}
