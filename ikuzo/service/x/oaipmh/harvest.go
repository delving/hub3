package oaipmh

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/delving/hub3/config"
	"github.com/kiivihal/goharvest/oai"
	"github.com/rs/zerolog/log"
)

const (
	VerbListRecords     = "ListRecords"
	VerbListIdentifiers = "ListIdentifiers"
	DateFormat          = "2006-01-02T15:04:05Z"
	UnixStart           = "1970-01-01T12:00:00Z"
)

var ErrNoRecordsMatch = errors.New("no records match OAI-PMH request")

type HarvestInfo struct {
	LastCheck    time.Time
	LastModified time.Time
	Error        string
}

type HarvestCallback func(r *oai.Response) error

type HarvestMetrics struct {
	Errors           []error
	Processed        int
	Deleted          int
	From             string
	Until            string
	Aborted          bool
	NoRecordsMatch   bool
	Pages            int
	CompleteListSize int
}

type HarvestTask struct {
	OrgID       string
	Name        string
	CheckEvery  time.Duration
	HarvestInfo *HarvestInfo
	Request     oai.Request
	CallbackFn  HarvestCallback
	rw          sync.Mutex
	running     bool
	m           HarvestMetrics
	Client      *http.Client
}

func (ht *HarvestTask) getOrCreateHarvestInfo() error {
	path := ht.getHarvestInfoPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		ht.HarvestInfo = &HarvestInfo{}
		err := ht.writeHarvestInfo()
		if err != nil {
			return err
		}
		return err
	}

	r, err := os.Open(path)
	if err != nil {
		return err
	}

	defer r.Close()

	var info HarvestInfo

	err = json.NewDecoder(r).Decode(&info)
	if err != nil {
		return err
	}

	ht.HarvestInfo = &info
	if ht.HarvestInfo == nil {
		log.Warn().Msg("unable to get harvest info and starting from scratch")

		ht.HarvestInfo = &HarvestInfo{}
	}

	return nil
}

func (ht *HarvestTask) writeHarvestInfo() error {
	b, err := json.Marshal(ht.HarvestInfo)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(
		ht.getHarvestInfoPath(),
		b,
		os.ModePerm,
	)
}

func (ht *HarvestTask) getHarvestInfoPath() string {
	return filepath.Join(config.Config.EAD.CacheDir, ht.Name+"_harvest.json")
}

func (ht *HarvestTask) BuildRequest() *oai.Request {
	if err := ht.getOrCreateHarvestInfo(); err != nil {
		log.Error().Err(err).Msg("unable to get harvest info")
	}

	req := ht.Request
	if !ht.HarvestInfo.LastModified.IsZero() {
		req.From = ht.HarvestInfo.LastModified.Format(DateFormat)
	}

	return &req
}

func (ht *HarvestTask) GetPage(ctx context.Context, request *oai.Request) (*oai.Response, error) {
	if ht.Client == nil {
		timeout := 60 * time.Second
		ht.Client = &http.Client{
			Timeout: timeout,
		}
	}

	var oaiResponse *oai.Response

	err := retry(10, time.Second, func() error {
		resp, err := ht.Client.Get(request.GetFullURL())
		if err != nil {
			return err
		}

		// Make sure the response body object will be closed after
		// reading all the content body's data
		defer resp.Body.Close()

		s := resp.StatusCode
		switch {
		case s >= 500:
			// Retry
			return fmt.Errorf("server error: %v", s)
		case s == 408:
			// Retry
			return fmt.Errorf("timeout error: %v", s)
		case s >= 400:
			// Don't retry, it was client's fault
			return stop{fmt.Errorf("client error: %v", s)}
		default:
			// Happy
			// Read all the data
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return stop{err}
			}

			// Unmarshall all the data
			err = xml.Unmarshal(body, &oaiResponse)
			if err != nil {
				return stop{err}
			}

			return nil
		}
	})
	if err != nil {
		log.Error().Err(err).Msgf("problem url: %s", request.GetFullURL())
		return nil, err
	}

	return oaiResponse, nil
}

func (ht *HarvestTask) harvest(ctx context.Context, request *oai.Request) error {
	log.Debug().
		Str("name", ht.Name).
		Str("url", request.GetFullURL()).
		Msg("retrieving page")

	// Use Perform to get the OAI response
	resp, err := ht.GetPage(ctx, request)
	if err != nil {
		ht.m.Errors = append(ht.m.Errors, err)
		ht.m.Aborted = true

		return err
	}

	switch resp.Error.Code {
	case "noRecordMatch":
		ht.m.NoRecordsMatch = true
		return ErrNoRecordsMatch
	case "error":
		pmhErr := resp.Error
		err := fmt.Errorf(
			"OAI-PMH response returns an error %s: %s", pmhErr.Code, pmhErr.Message,
		)
		log.Error().Err(err).Str("verb", resp.Request.Verb).
			Str("error.code", resp.Error.Code).
			Str("error.message", resp.Error.Message).
			Str("url", request.GetFullURL()).
			Msg("response returns an error")

		ht.m.Errors = append(ht.m.Errors, err)
		ht.m.Aborted = true

		return err
	default:
	}

	ht.m.Pages++

	// Execute the callback function with the response
	if callBackErr := ht.CallbackFn(resp); callBackErr != nil {
		ht.m.Errors = append(ht.m.Errors, callBackErr)
		ht.m.Aborted = true

		return callBackErr
	}

	// Check for a resumptionToken
	hasResumptionToken, token, completeListSize := resp.ResumptionToken()

	switch ht.Request.Verb {
	case VerbListIdentifiers:
		ht.m.Processed += len(resp.ListIdentifiers.Headers)

		for _, header := range resp.ListIdentifiers.Headers {
			if header.Status == "deleted" {
				ht.m.Deleted++
			}
		}
	case VerbListRecords:
		ht.m.Processed += len(resp.ListRecords.Records)

		for _, record := range resp.ListRecords.Records {
			if record.Header.Status == "deleted" {
				ht.m.Deleted++
			}
		}
	}

	if completeListSize != 0 {
		ht.m.CompleteListSize = completeListSize
	}

	select {
	case <-ctx.Done():
		log.Info().Msg("context canceled for harvesting")
		return ctx.Err()
	default:
	}

	// Harvest further if there is a resumption token
	if hasResumptionToken {
		request.Set = ""
		request.MetadataPrefix = ""
		request.From = ""
		request.Until = ""
		request.ResumptionToken = token
		request.CompleteListSize = completeListSize

		if err := ht.harvest(ctx, request); err != nil {
			return err
		}
	}

	return nil
}

func (ht *HarvestTask) Harvest(ctx context.Context) error {
	if ht.running {
		log.Warn().Str("name", ht.Name).Msg("harvest task already running")
		return nil
	}

	ht.rw.Lock()
	ht.running = true
	start := time.Now()
	ht.rw.Unlock()

	defer func() {
		ht.rw.Lock()
		ht.running = false
		ht.rw.Unlock()
	}()

	// start gathering metrics
	ht.m = HarvestMetrics{}

	// copy because the original needs to be available for next harvest run
	req := ht.BuildRequest()
	// alway set the until to now
	req.Until = start.Format(DateFormat)

	ht.m.Until = req.Until
	ht.m.From = req.From

	log.Info().
		Str("name", ht.Name).
		Str("url", req.GetFullURL()).
		Msg("starting harvest task")

	err := ht.harvest(ctx, req)
	if err != nil {
		if !errors.Is(err, ErrNoRecordsMatch) {
			log.Error().
				Err(err).
				Str("name", ht.Name).
				Str("url", req.GetFullURL()).
				Msg("harvest returned with error")

			return err
		}
	}

	ht.HarvestInfo.LastCheck = start
	if !ht.m.Aborted && !ht.m.NoRecordsMatch {
		ht.HarvestInfo.LastModified = start
	}

	if err := ht.writeHarvestInfo(); err != nil {
		log.Error().
			Err(err).
			Str("name", ht.Name).
			Msg("unable to write harvest info to disk")
	}

	log.Info().
		Str("name", ht.Name).
		Int("processed", ht.m.Processed).
		Int("deleted", ht.m.Deleted).
		Int("pages", ht.m.Pages).
		Str("from", ht.m.From).
		Str("until", ht.m.Until).
		Int("completeListSize", ht.m.CompleteListSize).
		Bool("aborted", ht.m.Aborted).
		Bool("noRecordsMatch", ht.m.NoRecordsMatch).
		Str("finalURL", req.GetFullURL()).
		Msg("finished harvest task")

	return nil
}

func (ht *HarvestTask) updateHarvestInfo() {
	if err := ht.getOrCreateHarvestInfo(); err != nil {
		log.Error().Err(err).Msg("cannot get last harvest check")
	}
}

// GetLastCheck returns last time the task has run.
func (ht *HarvestTask) GetLastCheck() time.Time {
	if ht.HarvestInfo == nil {
		ht.updateHarvestInfo()
	}

	return ht.HarvestInfo.LastCheck
}

// SetLastCheck sets time the task has run.
func (ht *HarvestTask) SetLastCheck(t time.Time) {
	if ht.HarvestInfo == nil {
		ht.updateHarvestInfo()
	}

	ht.HarvestInfo.LastCheck = t
}

// SetUnixStartFrom sets the From param to unix start datetime.
func (ht *HarvestTask) SetUnixStartFrom() {
	ht.Request.From = UnixStart
}

// SetRelativeFrom sets the From param based on the last check minus the duration check.
func (ht *HarvestTask) SetRelativeFrom() {
	lt := ht.GetLastCheck()
	if lt.IsZero() {
		lt = time.Now()
	}

	ht.Request.From = lt.Add(ht.CheckEvery * -1).Format(DateFormat)
}
