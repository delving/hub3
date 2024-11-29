package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
)

// main is the entry point of the application.
func main() {
	var apiURL, outputDir, fileLabel string
	flag.StringVar(&apiURL, "url", "", "API URL to fetch data from")
	flag.StringVar(&outputDir, "output", "output", "Output directory to save files")
	flag.StringVar(&fileLabel, "label", "", "Descriptive label to be used in the filename")

	flag.Parse()

	if apiURL == "" {
		fmt.Println("API URL is required. Usage: harvest_rdf -url=<API_URL>")
		return
	}

	err := harvestData(apiURL, outputDir, fileLabel)
	if err != nil {
		slog.Error("could not harvest data", "error", err)
	}
}

// harvestData fetches data from the API and saves it to the specified output directory.
func harvestData(apiURL, outputDir, fileLabel string) error {
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		slog.Error("failed to create output directory", "error", err)
		return err
	}

	output, err := os.Create(ntriplesFname(outputDir, fileLabel))
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	var seen int
	var pages int
	var scrollID string
	for {
		resp, err := getAPIResponse(apiURL, scrollID)
		if err != nil {
			return fmt.Errorf("failed to get API response: %w", err)
		}
		scrollID = resp.Pager.NextScrollID
		pages++

		for _, fg := range resp.Items {
			seen++
			graph, err := fg.Graph()
			if err != nil {
				return fmt.Errorf("failed to get graph: %w", err)
			}

			if err := ntriples.SerializeFiltered(graph, output, "<urn:private"); err != nil {
				return fmt.Errorf("failed to serialize graph: %w", err)
			}
		}

		if seen%500 == 0 {
			slog.Info("progress update", "seen", seen, "pages", pages, "total", resp.Pager.Total)
		}

		if scrollID == "" {
			break
		}
	}

	return nil
}

// getAPIResponse makes an HTTP GET request to the API and returns the response.
func getAPIResponse(uri, scrollID string) (*fragments.ScrollResultV4, error) {
	if scrollID != "" {
		apiURI, err := url.Parse(uri)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %w", err)
		}
		params := apiURI.Query()
		params.Set("scrollID", scrollID)
		apiURI.RawQuery = params.Encode()
		uri = apiURI.String()
	}

	slog.Info("retrieving page", "uri", uri)

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

// ntriplesFname generates a filename for the output file based on the output directory and file label.
func ntriplesFname(outputDir, fileLabel string) string {
	return filepath.Join(
		outputDir,
		fmt.Sprintf("rdf-export_%s_%s.nt", fileLabel, time.Now().Format("2006-01-02_150405")),
	)
}
