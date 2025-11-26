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
	"strings"
	"time"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/rdf/formats/jsonld"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/tidwall/gjson"
)

// main is the entry point of the application.
func main() {
	var apiURL, outputDir, fileLabel string
	var v1Mode bool
	flag.StringVar(&apiURL, "url", "", "API URL to fetch data from")
	flag.StringVar(&outputDir, "output", "output", "Output directory to save files")
	flag.StringVar(&fileLabel, "label", "", "Descriptive label to be used in the filename")
	flag.BoolVar(&v1Mode, "v1", false, "Harvest in v1.mode")

	flag.Parse()

	if apiURL == "" {
		fmt.Println("API URL is required. Usage: harvest_rdf -url=<API_URL>")
		return
	}

	if v1Mode {
		err := harvestV1Data(apiURL, outputDir, fileLabel)
		if err != nil {
			slog.Error("could not harvest data", "error", err)
		}
		return
	}

	err := harvestV2Data(apiURL, outputDir, fileLabel)
	if err != nil {
		slog.Error("could not harvest data", "error", err)
	}
}

// harvestV1Data fetches data from the v1 API and saves it to the specified output directory.
func harvestV1Data(apiURL, outputDir, fileLabel string) error {
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		slog.Error("failed to create output directory", "error", err)
		return err
	}

	output, err := os.Create(ntriplesFname(outputDir, fileLabel))
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	var nextPage, seen, pages int
	nextPage = 1

	for {
		resp, err := getV1APIResponse(apiURL, nextPage)
		if err != nil {
			return fmt.Errorf("error retrieving v1 response: %w", err)
		}
		pages++

		nextPage = resp.NextPage

		for _, graph := range resp.Graphs {
			seen++
			g, err := jsonld.Parse(strings.NewReader(graph), nil)
			if err != nil {
				return fmt.Errorf("unable to parse json-ld sting: %w \n %s", err, graph)
			}

			if err := ntriples.SerializeFiltered(g, output, "<urn:private"); err != nil {
				return fmt.Errorf("failed to serialize graph: %w", err)
			}
		}
		if seen%500 == 0 {
			slog.Info("progress update", "seen", seen, "pages", pages, "total", resp.NumFound)
		}

		if resp.EOF {
			break
		}
	}

	return nil
}

type v1Response struct {
	NumFound int
	NextPage int
	Graphs   []string
	EOF      bool
}

// getV1APIResponse makes an HTTP GET request to the API and returns the response.
func getV1APIResponse(uri string, page int) (data *v1Response, err error) {
	data = &v1Response{}

	// Always parse the URI to ensure format=json is set
	apiURI, err := url.Parse(uri)
	if err != nil {
		return data, fmt.Errorf("failed to parse URL: %w", err)
	}
	params := apiURI.Query()

	// Ensure format=json for v1 API
	if params.Get("format") == "" {
		params.Set("format", "json")
	}

	// Set page parameter if provided
	if page != 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}

	apiURI.RawQuery = params.Encode()
	uri = apiURI.String()

	slog.Info("retrieving page", "uri", uri)

	resp, err := http.Get(uri)
	if err != nil {
		return data, fmt.Errorf("failed to make the API request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, fmt.Errorf("failed to read error response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("error response from API", "uri", uri, "response", string(body))
		return data, fmt.Errorf("unable to retrieve API: %s", uri)
	}

	numFoundRaw := gjson.GetBytes(body, "@dig:pagination|@flatten.0.numFound")
	if !numFoundRaw.Exists() {
		return data, fmt.Errorf("unable to find numfound in API response: %s", uri)
	}

	slog.Info("numFound", "numFound", numFoundRaw.Int(), "raw", numFoundRaw.String())

	data.NumFound = int(numFoundRaw.Int())

	graphsRaw := gjson.GetBytes(body, "@dig:system.source_graph|@flatten")
	if !graphsRaw.Exists() {
		return data, fmt.Errorf("unable to find source_graph in API response: %s", uri)
	}

	for _, graph := range graphsRaw.Array() {
		data.Graphs = append(data.Graphs, graph.String())
	}

	hasNext := gjson.GetBytes(body, "@dig:pagination|@flatten.0.hasNext")
	if !hasNext.Exists() {
		return data, fmt.Errorf("unable to find hasNext in API response: %s", uri)
	}

	if !hasNext.Bool() {
		data.EOF = true
		return data, nil
	}

	nextPage := gjson.GetBytes(body, "@dig:pagination|@flatten.0.nextPageNumber")
	if !nextPage.Exists() {
		return data, fmt.Errorf("unable to find nextPage in API response: %s", uri)
	}

	data.NextPage = int(nextPage.Int())

	return data, nil
}

// harvestV2Data fetches data from the API and saves it to the specified output directory.
func harvestV2Data(apiURL, outputDir, fileLabel string) error {
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
		resp, err := getV2APIResponse(apiURL, scrollID)
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

// getV2APIResponse makes an HTTP GET request to the API and returns the response.
func getV2APIResponse(uri, scrollID string) (*fragments.ScrollResultV4, error) {
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
