package adlib

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/delving/hub3/ikuzo/service/x/adlib/internal"
)

type Record = internal.Crecord

type harvestError struct {
	PageURL string
	Err     string
}

func (he harvestError) Error() string {
	return fmt.Sprintf("error on page %s: %s", he.PageURL, he.Err)
}

func (he harvestError) Is(target error) bool {
	return reflect.TypeOf(target) == reflect.TypeOf(he)
}

func (he harvestError) createSingleRecordPages() (pages []string, err error) {
	baseURL, err := url.Parse(he.PageURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse error url: %w", err)
	}

	params := baseURL.Query()
	limit, convErr := strconv.Atoi(params.Get("limit"))
	if convErr != nil {
		return nil, convErr
	}

	offset, convErr := strconv.Atoi(params.Get("startFrom"))
	if convErr != nil {
		return nil, convErr
	}

	slog.Debug("creating error pages", "offset", offset, "limit", limit, "pageURL", he.PageURL)

	for i := 0; i < limit; i++ {
		pageURL := baseURL
		params.Set("limit", "1")
		params.Set("startFrom", fmt.Sprintf("%d", offset+i))
		pageURL.RawQuery = params.Encode()
		pages = append(pages, pageURL.String())
	}

	return pages, nil
}

func (he harvestError) harvestPageWithErrors(ctx context.Context, c *Client, cfg *HarvestConfig, records chan *internal.Crecord) error {
	errorPages, err := he.createSingleRecordPages()
	if err != nil {
		slog.Error("unable to create single record pages: %w", err)
		return err
	}

	for _, pageURL := range errorPages {
		atomic.AddUint64(&cfg.ErrorPagesSubmitted, 1)
		processErr := c.processPage(ctx, pageURL, cfg, records)
		if processErr != nil {
			cfg.addError(pageURL, processErr)
		}
	}

	return nil
}

type HarvestConfig struct {
	TotalCount          int
	TotalPages          int
	Offset              int
	HarvestFrom         time.Time
	HarvestErrors       map[string]string
	PagesProcessed      uint64
	RecordsProcessed    uint64
	Database            string
	Search              string
	Limit               int
	rw                  sync.RWMutex
	ErrorPagesSubmitted uint64
	ErrorPagesProcessed uint64
}

func (cfg *HarvestConfig) addError(pageURL string, err error) {
	cfg.rw.Lock()
	if len(cfg.HarvestErrors) == 0 {
		cfg.HarvestErrors = map[string]string{}
	}
	cfg.HarvestErrors[pageURL] = err.Error()
	cfg.rw.Unlock()
	slog.Info("adding page with error", "error", err.Error(), "pageURL", pageURL)
}

func (cfg *HarvestConfig) removeError(pageURL string) {
	cfg.rw.Lock()
	if len(cfg.HarvestErrors) == 0 {
		cfg.HarvestErrors = map[string]string{}
	}
	delete(cfg.HarvestErrors, pageURL)
	cfg.rw.Unlock()
}

func (cfg *HarvestConfig) pageURL(baseURL string) (string, error) {
	if cfg.Limit == 0 {
		cfg.Limit = 50
	}
	if cfg.Offset == 0 {
		cfg.Offset = 1
	}

	if cfg.Search == "" {
		cfg.Search = "all"
	}

	if !cfg.HarvestFrom.IsZero() {
		isoFormat := "2006-01-02T15:04:05.000Z07:00"
		cfg.Search = fmt.Sprintf("modification greater '%s'", cfg.HarvestFrom.Format(isoFormat))
	}

	pageURL, parseErr := url.Parse(baseURL)
	if parseErr != nil {
		return "", parseErr
	}

	params := url.Values{}
	params.Set("database", cfg.Database)
	params.Set("search", cfg.Search)
	params.Set("xmlType", "grouped")
	params.Set("limit", fmt.Sprintf("%d", cfg.Limit))
	params.Set("startFrom", fmt.Sprintf("%d", cfg.Offset))

	pageURL.RawQuery = params.Encode()

	return pageURL.String(), nil
}

func (c *Client) GetPage(pageURL string) (*internal.CadlibXML, error) {
	resp, err := c.c.Get(pageURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to get page: status code %d (%s)", resp.StatusCode, resp.Status)
	}

	defer resp.Body.Close()

	apiResponse, err := internal.ParseResponse(resp.Body)
	if err != nil {
		return nil, err
	}

	if apiResponse.Cdiagnostic.Cerror != nil {
		cerror := apiResponse.Cdiagnostic.Cerror
		slog.Debug("retrieve page error", "page", pageURL, "info", cerror.Info, "message", cerror.Message)
	}

	return apiResponse, nil
}

func (c *Client) processPage(ctx context.Context, pageURL string, cfg *HarvestConfig, records chan *internal.Crecord) error {
	switch {
	case strings.Contains(pageURL, "&limit=1&"):
		atomic.AddUint64(&cfg.ErrorPagesProcessed, 1)
	default:
		atomic.AddUint64(&cfg.PagesProcessed, 1)
	}

	apiResponse, err := c.GetPage(pageURL)
	if err != nil {
		slog.Debug("unable to harvest page", "page", pageURL, "error", err)
		if apiResponse == nil {
			cfg.addError(pageURL, fmt.Errorf("unable to process page: %w", err))
			return nil
		}
	}

	if len(apiResponse.CrecordList.Crecord) == 0 || apiResponse.GetError() != "" {
		err := harvestError{
			PageURL: pageURL,
			Err:     apiResponse.GetError(),
		}
		slog.Debug("unable to harvest page", "page",
			pageURL, "error", err, "response", apiResponse.Cdiagnostic.Cerror)
		return err
	}

	for _, record := range apiResponse.CrecordList.Crecord {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case records <- record:
		}
	}

	return nil
}

func (c *Client) Harvest(ctx context.Context, cfg *HarvestConfig, cb func(record *Record) error) error {
	pages := make(chan string, 500)
	g, _ := errgroup.WithContext(ctx)

	// Produce
	g.Go(func() error {
		defer func() {
			close(pages)
		}()

		// produces pages to save
		firstPageURL, err := cfg.pageURL(c.url)
		if err != nil {
			return err
		}

		firstPage, err := c.GetPage(firstPageURL)
		if err != nil {
			return err
		}
		slog.Info("downloading first page", "page", firstPage.Cdiagnostic, "pageURL", firstPageURL)

		cfg.TotalCount, err = strconv.Atoi(firstPage.Cdiagnostic.Chits.Text)
		if err != nil {
			return err
		}

		cfg.TotalPages = (cfg.TotalCount / cfg.Limit) + 1
		pages <- firstPageURL

		for i := 1; i < cfg.TotalPages; i++ {
			cfg.Offset += cfg.Limit
			pageURL, pageErr := cfg.pageURL(c.url)
			if pageErr != nil {
				return pageErr
			}

			if cfg.Offset > cfg.TotalCount {
				continue
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case pages <- pageURL:
			}
		}

		return nil
	})

	records := make(chan *internal.Crecord, 1000)
	harvestErrors := make(chan *harvestError, 100)

	// Map
	nWorkers := 4
	workers := int32(nWorkers)
	for i := 0; i < nWorkers; i++ {
		g.Go(func() error {
			defer func() {
				// Last one out closes shop
				if atomic.AddInt32(&workers, -1) == 0 {
					close(records)
				}
			}()

			for pageURL := range pages {
				processErr := c.processPage(ctx, pageURL, cfg, records)
				if processErr != nil {
					if !errors.Is(processErr, harvestError{}) {
						slog.Info("processing error during page processing", "error", processErr, "harvestErr", errors.Is(processErr, harvestError{}))
						return processErr
					}

					harvestErr, ok := processErr.(harvestError)
					if !ok {
						return fmt.Errorf("can't cast harvest err; %w", processErr)
					}

					if err := harvestErr.harvestPageWithErrors(ctx, c, cfg, records); err != nil {
						return fmt.Errorf("unable to reharvest page with errors; %w", err)
					}
				}
			}

			return nil
		})
	}

	// Reduce
	g.Go(func() error {
		defer close(harvestErrors)

		for record := range records {
			if record != nil {
				atomic.AddUint64(&cfg.RecordsProcessed, 1)
				if cfg.RecordsProcessed%500 == 0 {
					slog.Debug("processing records", "totalRecords", cfg.TotalCount, "processed", cfg.RecordsProcessed)
				}
				if err := cb(record); err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
