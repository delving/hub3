/*
Copyright © 2021 Delving B.V. <info@delving.eu>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/delving/hub3/ikuzo/service/x/bulk"
	"github.com/spf13/cobra"
)

var (
	requestPath string
	publishHost string
	chunkSize   int
)

// bulkCmd represents the bulk command
var bulkCmd = &cobra.Command{
	Use:   "bulk",
	Short: "Load bulk.Request from disk",
	Long: `Loading bulk.Requests serialized as line delimited json.

	We assume that the structure is 'orgID/datasetID/hubID.jsonl' and that each bulk.Request
	is stored in a single file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("bulk called")
		log.Fatal(publish(cmd.Context(), publishHost, requestPath))
	},
}

func init() {
	rootCmd.AddCommand(bulkCmd)

	bulkCmd.Flags().StringVarP(&requestPath, "dataPath", "p", ".", "Full path to orgIDs for the bulk.Requests.")
	bulkCmd.Flags().StringVarP(&publishHost, "host", "", "http://localhost:3001", "network host of where target hub3 is running")
	bulkCmd.Flags().IntVarP(&chunkSize, "chunkSize", "", 500, "size of number of records send per batch")
}

func publish(ctx context.Context, host, dataPath string) error {
	p := bulk.NewPublisher(publishHost, requestPath)
	p.BulkSize = chunkSize

	err := p.Do(ctx)
	if err != nil {
		return fmt.Errorf("unable to publish data: %w", err)
	}

	return nil
}
