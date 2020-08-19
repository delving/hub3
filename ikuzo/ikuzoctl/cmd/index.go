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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/delving/hub3/ikuzo/service/x/bulk"
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
	is, isErr := cfg.GetIndexService()
	if isErr != nil {
		return fmt.Errorf("unable to create index service; %w", isErr)
	}

	bulkSvc, bulkErr := bulk.NewService(
		bulk.SetIndexService(is),
		bulk.SetIndexTypes(cfg.ElasticSearch.IndexTypes...),
	)
	if bulkErr != nil {
		return fmt.Errorf("unable to create bulk service; %w", isErr)
	}

	parser := bulkSvc.NewParser()

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := is.Shutdown(ctx); err != nil {
			log.Fatal().Err(err).Msg("shutdown failed")
		}
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
		dataset,
	)

	var badFiles uint64

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-sigs:
				log.Info().Msg("caught shutdown signal")
				cancel()
				time.Sleep(1 * time.Second)
				shutdown()
				done <- true
			case <-ticker.C:
				log.Info().Msgf("consumer stats: %+v", is.Metrics())

				if is.BulkIndexStats().NumFlushed >= is.Metrics().Nats.Consumed {
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

	walkErr := filepath.Walk(baseDir, func(path string, fInfo os.FileInfo, err error) error {
		if fInfo.IsDir() || !strings.HasSuffix(fInfo.Name(), ".json") {
			return nil
		}

		if fInfo.Size() == int64(0) {
			badFiles++
			return nil
		}

		f, readErr := os.Open(path)
		if readErr != nil {
			return readErr
		}

		v1rec, convErr := newV1(f)
		if convErr != nil {
			log.Error().Err(err).Str("fname", fInfo.Name()).Msg("read error")
			return nil
		}

		f.Close()

		if errors.Is(ctx.Err(), context.Canceled) {
			return ctx.Err()
		}

		if err := parser.Publish(ctx, v1rec.bulkRequest()); err != nil {
			log.Error().Err(err).Msgf("bulk request: %#v", v1rec.bulkRequest())
			return err
		}

		return nil
	},
	)

	if walkErr != nil && !errors.Is(walkErr, context.Canceled) {
		return walkErr
	}

	<-done

	return nil
}

type v1 struct {
	OrgID  string `json:"orgID"`
	Spec   string `json:"spec"`
	HubID  string `json:"hubID"`
	System struct {
		Graph         string `json:"source_graph"`
		NamedGraphURI string `json:"graph_name"`
	} `json:"system"`
}

func newV1(r io.Reader) (*v1, error) {
	var rec v1
	if err := json.NewDecoder(r).Decode(&rec); err != nil {
		return nil, err
	}

	return &rec, nil
}

func (rec *v1) bulkRequest() *bulk.Request {
	return &bulk.Request{
		HubID:     rec.HubID,
		OrgID:     rec.OrgID,
		DatasetID: rec.Spec,
		// LocalID:       "",
		NamedGraphURI: rec.System.NamedGraphURI,
		// RecordType:    "",
		Action: "index",
		// ContentHash:   "",
		Graph:         rec.System.Graph,
		GraphMimeType: "application/ld+json",
		SubjectType:   "",
	}
}
