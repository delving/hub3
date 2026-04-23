// Copyright © 2026 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package models

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	c "github.com/delving/hub3/config"
	"github.com/matryer/is"
	elastic "github.com/olivere/elastic/v7"
)

func withStoragesDisabled(t *testing.T) func() {
	t.Helper()
	c.InitConfig()
	prevRDF := c.Config.RDF.RDFStoreEnabled
	prevES := c.Config.ElasticSearch.Enabled
	c.Config.RDF.RDFStoreEnabled = false
	c.Config.ElasticSearch.Enabled = false
	return func() {
		c.Config.RDF.RDFStoreEnabled = prevRDF
		c.Config.ElasticSearch.Enabled = prevES
	}
}

func TestDropRecordsByHubIDsEmpty(t *testing.T) {
	is := is.New(t)
	restore := withStoragesDisabled(t)
	defer restore()

	ds := DataSet{OrgID: "org1", Spec: "ds1"}
	count, err := ds.DropRecordsByHubIDs(context.Background(), nil)
	is.NoErr(err)
	is.Equal(count, 0)

	count, err = ds.DropRecordsByHubIDs(context.Background(), []string{})
	is.NoErr(err)
	is.Equal(count, 0)
}

func TestDropRecordsByHubIDsNoOpWhenBothStoragesOff(t *testing.T) {
	is := is.New(t)
	restore := withStoragesDisabled(t)
	defer restore()

	ds := DataSet{OrgID: "org1", Spec: "ds1"}
	count, err := ds.DropRecordsByHubIDs(
		context.Background(),
		[]string{"org1_ds1_a", "org1_ds1_b"},
	)
	is.NoErr(err)
	is.Equal(count, 0) // nothing deleted because nothing enabled
}

func TestDropGraphsByHubIDsBuildsDropSilentQuery(t *testing.T) {
	is := is.New(t)
	var captured string
	prev := sparqlUpdateSender
	sparqlUpdateSender = func(orgID, update string) []error {
		captured = update
		return nil
	}
	defer func() { sparqlUpdateSender = prev }()

	ds := DataSet{OrgID: "org1", Spec: "ds1"}
	err := ds.dropGraphsByHubIDs([]string{"org1_ds1_a", "org1_ds1_b"})
	is.NoErr(err)

	is.True(strings.Contains(captured, "DROP SILENT GRAPH <urn:org1_ds1_a/graph>"))
	is.True(strings.Contains(captured, "DROP SILENT GRAPH <urn:org1_ds1_b/graph>"))
}

func TestDropGraphsByHubIDsPropagatesError(t *testing.T) {
	is := is.New(t)
	prev := sparqlUpdateSender
	sparqlUpdateSender = func(orgID, update string) []error {
		return []error{fmt.Errorf("boom")}
	}
	defer func() { sparqlUpdateSender = prev }()

	ds := DataSet{OrgID: "org1", Spec: "ds1"}
	err := ds.dropGraphsByHubIDs([]string{"org1_ds1_a"})
	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "boom"))
}

func TestDeleteIndexRecordsByHubIDsBuildsTermsQuery(t *testing.T) {
	is := is.New(t)

	var capturedIndices []string
	var capturedHubIDs [][]interface{}

	prev := esDeleteByQuerySender
	esDeleteByQuerySender = func(
		ctx context.Context,
		index string,
		q elastic.Query,
	) (int, error) {
		capturedIndices = append(capturedIndices, index)
		// Extract the terms clause so we can assert hubIDs shipped through
		src, err := q.Source()
		if err != nil {
			return 0, err
		}
		b, _ := json.Marshal(src)
		var shaped struct {
			Bool struct {
				Must []map[string]interface{} `json:"must"`
			} `json:"bool"`
		}
		_ = json.Unmarshal(b, &shaped)
		for _, clause := range shaped.Bool.Must {
			if terms, ok := clause["terms"].(map[string]interface{}); ok {
				if ids, ok := terms["meta.hubID"].([]interface{}); ok {
					capturedHubIDs = append(capturedHubIDs, ids)
				}
			}
		}
		return 2, nil
	}
	defer func() { esDeleteByQuerySender = prev }()

	c.InitConfig()
	prevTypes := c.Config.ElasticSearch.IndexTypes
	c.Config.ElasticSearch.IndexTypes = []string{"v2"}
	defer func() { c.Config.ElasticSearch.IndexTypes = prevTypes }()

	ds := DataSet{OrgID: "org1", Spec: "ds1"}
	count, err := ds.deleteIndexRecordsByHubIDs(
		context.Background(),
		[]string{"org1_ds1_a", "org1_ds1_b"},
	)
	is.NoErr(err)
	is.Equal(count, 2)
	is.Equal(len(capturedIndices), 1) // v2 only

	is.Equal(len(capturedHubIDs), 1)
	is.Equal(len(capturedHubIDs[0]), 2)
	is.Equal(capturedHubIDs[0][0].(string), "org1_ds1_a")
	is.Equal(capturedHubIDs[0][1].(string), "org1_ds1_b")
}
