package ead_test

import (
	"context"
	"testing"

	"github.com/delving/hub3/hub3/ead"
	"github.com/matryer/is"
)

// nolint:gocritic
func TestNumbered(t *testing.T) {
	is := is.New(t)

	dsc := new(ead.Cdsc)
	err := parseUtil(dsc, "ead.0x.xml")
	is.NoErr(err)

	cfg := ead.NewNodeConfig(context.Background())
	_, nodeCount, err := dsc.NewNodeList(cfg)
	is.NoErr(err)

	is.Equal(int(nodeCount), 7)
}
