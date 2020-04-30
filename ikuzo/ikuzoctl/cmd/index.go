/*
Copyright © 2020 Delving B.V. <info@delving.eu>

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
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "build index from disk",
	Long: `This command allows you to target a local directory with source json
	records and submit them for indexing.

	This command uses the default hub3 configuration file`,
	Run: func(cmd *cobra.Command, args []string) {
		err := indexRecords()
		if err != nil {
			log.Fatal().Err(err).Msg("error synchronizing records")
		}
	},
}

var (
	indexMode string
	dataPath  string
	offline   bool
	orgID     string
	dataset   string
)

func init() {
	rootCmd.AddCommand(indexCmd)

	indexCmd.Flags().StringVarP(&indexMode, "indexMode", "m", "v2", "which mode of indexing is used")
	indexCmd.Flags().StringVarP(&dataPath, "path", "p", "", "which directory contains the source records")
	indexCmd.Flags().StringVarP(&orgID, "orgID", "", "", "orgID of the records")
	indexCmd.Flags().StringVarP(&dataset, "dataset", "", "", "dataset spec of the records")
	indexCmd.Flags().BoolVarP(&offline, "offline", "o", false, "build a new index but not set the default alias")
}

func indexRecords() error {
	logger := logger.NewLogger(logger.Config{})

	// TODO(kiivihal): hard-code the index name for now
	indexName := "hub3v2"
	if indexMode != "v2" {
		indexName = "hub3v1"
	}

	logger.Info().Str("index_name", indexName).Msg("selected index")

	// create BulkIndexer
	ncfg, err := cfg.Nats.GetConfig()
	if err != nil {
		return err
	}

	// turn off remote index
	cfg.ElasticSearch.UseRemoteIndexer = false

	svc, err := cfg.ElasticSearch.IndexService(&logger, ncfg)
	if err != nil {
		return err
	}

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		svc.Shutdown(ctx)
	}

	// func to loop through all records
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan bool)
	ticker := time.NewTicker(3 * time.Second)

	baseDir := filepath.Join(
		dataPath,
		orgID,
		fmt.Sprintf("%s.git", dataset),
		"index",
		"void_edmrecord",
	)

	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return err
	}

	indexRecords := uint64(len(files))

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			// case <-ctx.Done():
			// done <- true

			// log.Info().Msg("time out expired")

			// return
			case <-sigs:
				log.Info().Msg("caught shutdown signal")
				cancel()
				time.Sleep(1 * time.Second)
				shutdown()
				done <- true
			case <-ticker.C:
				// log.Printf("bi stats: %+v", bi.Stats())
				log.Printf("consumer stats: %+v", svc.Metrics())

				if svc.BulkIndexStats().NumFlushed >= indexRecords {
					cancel()
					time.Sleep(1 * time.Second)
					shutdown()

					done <- true

					log.Info().Msg("all records are indexed")

					return
				}
			}
		}
	}()

	err = svc.Start(ctx, 1)
	if err != nil {
		return err
	}

	for _, fInfo := range files {
		if fInfo.IsDir() || !strings.HasSuffix(fInfo.Name(), ".json") {
			continue
		}

		b, readErr := ioutil.ReadFile(filepath.Join(baseDir, fInfo.Name()))
		if readErr != nil {
			return readErr
		}

		id := strings.TrimSuffix(fInfo.Name(), ".json")
		hubID := fmt.Sprintf("%s_%s_%s", orgID, dataset, id)

		readErr = svc.Publish(
			ctx,
			&domainpb.IndexMessage{
				OrganisationID: orgID,
				DatasetID:      dataset,
				RecordID:       hubID,
				IndexName:      indexName,
				Deleted:        false,
				Revision: &domainpb.Revision{
					SHA:  "",
					Path: "",
				},
				Source: b,
			},
		)
		if readErr != nil {
			return readErr
		}
	}

	<-done

	return nil
}
