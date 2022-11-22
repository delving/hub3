package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/delving/hub3/ikuzo/rdf/formats/hextuples"
	"github.com/delving/hub3/ikuzo/rdf/formats/rdfxml"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

type NarthexStats struct {
	Lines     int
	Records   int
	OrgID     string
	DatasetID string
}

var (
	linePrefix = []byte("<!--<")
	lineSuffix = []byte(">-->")
)

func extractID(b []byte) (hubID, hash string, err error) {
	parts := bytes.SplitN(b, []byte("__"), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid line separator: %q", string(b))
	}

	hubID = string(bytes.TrimPrefix(parts[0], linePrefix))
	hash = string(bytes.TrimSuffix(parts[1], lineSuffix))

	return hubID, hash, nil
}

func ParseNarthexToParquet(r io.Reader, w io.Writer) (NarthexStats, error) {
	var stats NarthexStats

	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 5*1024*1024)

	var record bytes.Buffer

	// write
	pw, err := writer.NewParquetWriterFromWriter(w, new(hextuples.HexTuple), 4)
	if err != nil {
		return stats, err
	}

	pw.RowGroupSize = 128 * 1024 * 1024 // 128M
	pw.CompressionType = parquet.CompressionCodec_SNAPPY

	for scanner.Scan() {
		b := scanner.Bytes()

		if bytes.HasPrefix(b, linePrefix) {
			stats.Records++

			hubID, _, err := extractID(b)
			if err != nil {
				return stats, err
			}

			// TODO(kiivihal): add support for graph name
			// if stats.OrgID == "" {
			// parts := strings.SplitN(hubID, "_", 3)
			// stats.OrgID = parts[0]
			// stats.DatasetID = parts[1]
			// }

			g, err := rdfxml.Parse(&record, nil)
			if err != nil {
				return stats, err
			}

			g.GraphName = "urn:" + hubID

			if err := hextuples.SerializeParquet(g, pw); err != nil {
				return stats, nil
			}

			record.Reset()
			continue
		}

		_, err := record.Write(b)
		if err != nil {
			return stats, err
		}

		stats.Lines++
	}

	if err := scanner.Err(); err != nil {
		return stats, err
	}

	if err = pw.WriteStop(); err != nil {
		return stats, err
	}

	return stats, nil
}

func ParseNarthex(r io.Reader, w io.Writer) (NarthexStats, error) {
	var stats NarthexStats

	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 5*1024*1024)

	var record bytes.Buffer

	for scanner.Scan() {
		b := scanner.Bytes()

		if bytes.HasPrefix(b, linePrefix) {
			stats.Records++

			hubID, _, err := extractID(b)
			if err != nil {
				return stats, err
			}

			// TODO(kiivihal): add support for graph name
			// if stats.OrgID == "" {
			// parts := strings.SplitN(hubID, "_", 3)
			// stats.OrgID = parts[0]
			// stats.DatasetID = parts[1]
			// }

			g, err := rdfxml.Parse(&record, nil)
			if err != nil {
				return stats, err
			}

			g.GraphName = "urn:" + hubID

			if err := hextuples.Serialize(g, w); err != nil {
				return stats, nil
			}

			record.Reset()
			continue
		}

		_, err := record.Write(b)
		if err != nil {
			return stats, err
		}

		stats.Lines++
	}

	if err := scanner.Err(); err != nil {
		return stats, err
	}

	return stats, nil
}
