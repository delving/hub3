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
	"testing"

	c "github.com/delving/hub3/config"
	"github.com/matryer/is"
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
