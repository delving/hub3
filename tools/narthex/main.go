package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/delving/hub3/ikuzo/rdf/formats/hextuples"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/delving/hub3/ikuzo/rdf/formats/rdfxml"
)

func main() {
	var outputDir, inputDir string
	flag.StringVar(&outputDir, "output", "output", "Output directory to save files")
	flag.StringVar(&inputDir, "input", "input", "input directory to save files")
	flag.Parse()

	if err := findAndProcessRDFfiles(inputDir); err != nil {
		log.Fatal(err)
	}
}

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

func findAndProcessRDFfiles(rootDir string) error {
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return nil
		}

		if info.IsDir() && info.Name() == "processed" {
			// Found the "processed" directory, call your processing function here
			if err := processDir(path); err != nil {
				return err
			}
			return filepath.SkipDir // Skip processing subdirectories of "processed"
		}

		return nil
	})
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		return err
	}
	return nil
}

func processDir(processedPath string) error {
	files, err := os.ReadDir(processedPath)
	if err != nil {
		return err
	}

	for _, entry := range files {
		if strings.HasSuffix(entry.Name(), ".xml") {
			dataFname := filepath.Join(processedPath, entry.Name())
			log.Printf("processing: %s", dataFname)

			f, err := os.Open(dataFname)
			if err != nil {
				return err
			}
			defer f.Close()
			specName := filepath.Base(filepath.Dir(processedPath))
			ntriples, err := os.Create(filepath.Join("/tmp", specName+".ntriples"))
			if err != nil {
				return err
			}
			defer ntriples.Close()
			stats, err := ParseNarthexToNtriples(f, ntriples)
			if err != nil {
				return err
			}
			_ = stats
		}
	}

	return nil
}

func ParseNarthexToNtriples(r io.Reader, w io.Writer) (NarthexStats, error) {
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
				return stats, fmt.Errorf("rdfxml parse erorr: %w", err)
			}

			g.GraphName = "urn:" + hubID

			if err := ntriples.SerializeFiltered(g, w, "<urn:private"); err != nil {
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
