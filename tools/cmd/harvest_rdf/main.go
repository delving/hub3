package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/rdf"
)

const privatePrefix = "<urn:private"

// main is the entry point of the application.
func main() {
	var apiURL, outputDir, fileLabel, format string
	flag.StringVar(&apiURL, "url", "", "v2 search API URL to fetch data from, e.g. https://api.example.org/api/search/v2?qf=meta.spec:x")
	flag.StringVar(&outputDir, "output", "output", "Output directory to save files")
	flag.StringVar(&fileLabel, "label", "", "Descriptive label to be used in the filename")
	flag.StringVar(&format, "format", "nq", "Output format: nq (N-Quads, one named graph per record) or nt (N-Triples)")

	flag.Parse()

	if apiURL == "" {
		fmt.Println("API URL is required. Usage: harvest_rdf -url=<API_URL>")
		os.Exit(2)
	}

	if format != "nq" && format != "nt" {
		fmt.Println("format must be nq or nt")
		os.Exit(2)
	}

	if err := harvestV2Data(apiURL, outputDir, fileLabel, format); err != nil {
		slog.Error("could not harvest data", "error", err)
		os.Exit(1)
	}
}

// harvestV2Data scrolls through the v2 API and writes every record graph to the output directory.
func harvestV2Data(apiURL, outputDir, fileLabel, format string) error {
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	f, err := os.Create(outputFname(outputDir, fileLabel, format))
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	output := bufio.NewWriter(f)

	var seen, pages int
	var total int64
	var scrollID string
	for {
		resp, err := getV2APIResponse(apiURL, scrollID)
		if err != nil {
			return fmt.Errorf("failed to get API response: %w", err)
		}
		pages++

		for _, fg := range resp.Items {
			seen++
			graph, err := fg.Graph()
			if err != nil {
				return fmt.Errorf("failed to get graph: %w", err)
			}

			// namedGraphURI is urn:<orgID>_<hubID>/graph; the underscore is not valid in a
			// URN namespace id, so the graph is named after the record's entry URI instead.
			var graphName string
			if format == "nq" {
				graphName = "<" + fg.Meta.GetEntryURI() + "/graph>"
			}

			if err := writeRecord(output, graph, fg.Meta.GetHubID(), graphName); err != nil {
				return fmt.Errorf("failed to serialize graph: %w", err)
			}
		}

		scrollID = ""
		if resp.Pager != nil {
			total = resp.Pager.Total
			scrollID = resp.Pager.NextScrollID
		}

		slog.Info("progress update", "seen", seen, "pages", pages, "total", total)

		if scrollID == "" {
			break
		}
	}

	if err := output.Flush(); err != nil {
		return fmt.Errorf("failed to flush output: %w", err)
	}

	if int64(seen) != total {
		return fmt.Errorf("harvested %d records, but API reported %d", seen, total)
	}

	return f.Close()
}

// writeRecord writes one record graph as N-Triples, or as N-Quads when graphName is set.
// Blank node labels are only unique within a record, so they are prefixed with a hash
// of the record's hubID to keep records from sharing blank nodes in the combined dump.
func writeRecord(w io.Writer, g *rdf.Graph, hubID, graphName string) error {
	h := fnv.New64a()
	h.Write([]byte(hubID))
	bnodePrefix := fmt.Sprintf("_:r%x_", h.Sum64())

	term := func(t rdf.Term) string {
		s := t.String()
		if strings.HasPrefix(s, "_:") {
			return bnodePrefix + s[2:]
		}
		return s
	}

	end := " .\n"
	if graphName != "" {
		end = " " + graphName + " .\n"
	}

	for _, triple := range g.Triples() {
		s := term(triple.Subject)
		o := term(triple.Object)
		if strings.HasPrefix(s, privatePrefix) || strings.HasPrefix(o, privatePrefix) {
			continue
		}

		if _, err := io.WriteString(w, s+" "+triple.Predicate.String()+" "+o+end); err != nil {
			return err
		}
	}

	return nil
}

// getV2APIResponse makes an HTTP GET request to the API and returns the response.
func getV2APIResponse(uri, scrollID string) (*fragments.ScrollResultV4, error) {
	apiURI, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}
	params := apiURI.Query()
	params.Set("itemFormat", "fragmentGraph")
	if scrollID != "" {
		params.Set("scrollID", scrollID)
	}
	apiURI.RawQuery = params.Encode()
	uri = apiURI.String()

	slog.Debug("retrieving page", "uri", uri)

	resp, err := http.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to make the API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error response body: %w", err)
		}
		slog.Error("error response from API", "uri", uri, "response", string(b))
		return nil, fmt.Errorf("unable to retrieve API: %s", uri)
	}

	var apiResp fragments.ScrollResultV4
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("unable to decode API response %q; %w", uri, err)
	}

	return &apiResp, nil
}

// outputFname generates a filename for the output file based on the output directory and file label.
func outputFname(outputDir, fileLabel, format string) string {
	return filepath.Join(
		outputDir,
		fmt.Sprintf("rdf-export_%s_%s.%s", fileLabel, time.Now().Format("2006-01-02_150405"), format),
	)
}
