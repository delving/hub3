package internal

import (
	"os"
	"testing"

	"github.com/matryer/is"
)

func TestParseNarthex(t *testing.T) {
	t.Skip("skip")
	is := is.New(t)

	f, err := os.Open("./testdata/processed.xml")
	// f, err := os.Open("/tmp/00000.xml")
	is.NoErr(err)

	o, err := os.Create("/tmp/big.hext")
	is.NoErr(err)

	// var buf bytes.Buffer
	// stats, err := ParseNarthex(f, &buf)
	stats, err := ParseNarthex(f, o)
	is.NoErr(err)
	is.Equal(stats.Records, 2)
	is.Equal(stats.Lines, 137) // only count content lines

	o.Close()

	// t.Logf("hextuples: %s", buf.String())
	// os.WriteFile("/tmp/test.hext", buf.Bytes(), os.ModePerm)
	// is.True(false)
}

func TestParseNarthexBig(t *testing.T) {
	t.Skip("dev")
	is := is.New(t)

	// f, err := os.Open("/home/kiivihal/NarthexFiles/brabantcloud/datasets/brabantse-gebouwen/processed/00000.xml")
	// f, err := os.Open("/home/kiivihal/NarthexFiles/NL-HaNA/datasets/rijksmuseum/processed/00000.xml")
	f, err := os.Open("/tmp/00000.xml")
	is.NoErr(err)
	defer f.Close()

	o, err := os.Create("/tmp/big.parquet")
	is.NoErr(err)

	stats, err := ParseNarthexToParquet(f, o)
	is.NoErr(err)
	is.True(stats.Records > 2)
	// is.Equal(stats.Records, 2)
	// is.Equal(stats.Lines, 137) // only count content lines
	o.Close()
}

func TestExtractID(t *testing.T) {
	is := is.New(t)

	testLine := "<!--<brabantcloud_brabantse-gebouwen_Q2477__c9f58148e36d0f9e2f64231d4b09e438c418fd1f>-->"

	hubID, hash, err := extractID([]byte(testLine))
	is.NoErr(err)
	is.Equal(hubID, "brabantcloud_brabantse-gebouwen_Q2477")
	is.Equal(hash, "c9f58148e36d0f9e2f64231d4b09e438c418fd1f")
}
