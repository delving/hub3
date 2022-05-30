package oaipmh

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/render"
)

func (s *Service) handleVerb() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := NewRequest(r)

		org, _ := domain.GetOrganization(r)
		req.orgConfig = &org.Config

		if !req.orgConfig.OAIPMH.Enabled {
			render.Error(w, r, domain.ErrServiceNotEnabled, &render.ErrorConfig{
				StatusCode:    http.StatusNotFound,
				PreventBubble: true,
			})
			return
		}

		resp, err := s.Do(r.Context(), &req)
		if err != nil {
			render.Error(w, r, err, nil)
			return
		}

		render.XML(w, r, resp)
	}
}

func (s *Service) Do(ctx context.Context, req *Request) (*Response, error) {
	resp := &Response{
		AttrXmlns:          "http://www.openarchives.org/OAI/2.0/",
		AttrXmlnsxsi:       "http://www.w3.org/2001/XMLSchema-instance",
		AttrSchemaLocation: "http://www.openarchives.org/OAI/2.0/ http://www.openarchives.org/OAI/2.0/OAI-PMH.xsd",
		ResponseDate:       time.Now().Format(TimeFormat),
		Request:            req,
	}

	var err error

	switch Verb(req.Verb) {
	case VerbIdentify:
		s.handleIdentify(resp)
	case VerbListMetadataFormats:
		s.handleListMetadataFormats(resp)
	case VerbGetRecord:
		err = s.handleGetRecord(resp)
	case VerbListSets:
		err = s.handleListSets(resp)
	case VerbListIdentifiers:
		err = s.handleListIdentifiers(resp)
	case VerbListRecords:
		err = s.handleListRecords(resp)
	default:
		resp.Error = append(resp.Error, ErrBadVerb)
	}

	if len(resp.Error) != 0 {
		resp.Request = &Request{BaseURL: req.BaseURL}
		return resp, nil
	}

	if err != nil {
		return nil, fmt.Errorf("unexpected error processing oai-pmh; %w", err)
	}

	return resp, nil
}

func (s *Service) handleIdentify(resp *Response) {
	resp.Identify = &Identify{
		RepositoryName:    resp.Request.orgConfig.OAIPMH.RepositoryName,
		BaseURL:           resp.Request.BaseURL,
		ProtocolVersion:   "2.0",
		AdminEmail:        resp.Request.orgConfig.OAIPMH.AdminEmails,
		DeletedRecord:     "no", // TODO(kiivihal): change later to persistent
		EarliestDatestamp: "1970-01-01T00:00:00Z",
		Granularity:       "YYYY-MM-DDThh:mm:ssZ",
	}
}

func (s *Service) handleListMetadataFormats(resp *Response) error {
	// TODO(kiivihal): implement check for identifier

	cfg := resp.Request.RequestConfig()

	formats, err := s.store.ListMetadataFormats(context.TODO(), &cfg)
	if err != nil {
		return err
	}

	resp.ListMetadataFormats = &ListMetadataFormats{
		MetadataFormat: formats,
	}

	return nil
}

func (s *Service) handleListSets(resp *Response) error {
	ctx := context.TODO()

	cfg, err := s.requestConfig(resp.Request)
	if err != nil {
		if errors.Is(err, ErrBadResumptionToken) {
			resp.Error = append(resp.Error, ErrBadResumptionToken)
			return nil
		}

		return fmt.Errorf("cannot get request config: %w", err)
	}

	res, err := s.store.ListSets(ctx, &cfg)
	if err != nil {
		return fmt.Errorf("error during listsets: %w", err)
	}

	if len(res.Errors) != 0 {
		resp.Error = res.Errors
		return nil
	}

	if cfg.TotalSize == 0 && res.Total > 0 {
		s.m.Lock()
		cfg.TotalSize = res.Total
		s.steps[cfg.ID] = cfg
		s.m.Unlock()
	}

	resp.ListSets = &ListSets{
		Set:             res.Sets,
		ResumptionToken: cfg.NextResumptionToken(&res),
	}

	return nil
}

func (s *Service) requestConfig(req *Request) (cfg RequestConfig, err error) {
	if req.ResumptionToken == "" {
		cfg = req.RequestConfig()
		cfg.ID, err = s.sid.Generate()
		if err != nil {
			return cfg, err
		}

		return cfg, nil
	}

	token, err := parseToken(req.ResumptionToken)
	if err != nil {
		return cfg, err
	}

	s.m.Lock()
	defer s.m.Unlock()

	cfg, ok := s.steps[token.HarvestID]
	if !ok {
		return cfg, ErrBadResumptionToken
	}

	cfg.CurrentRequest = token

	return cfg, nil
}

func (s *Service) handleListIdentifiers(resp *Response) error {
	ctx := context.TODO()

	cfg, err := s.requestConfig(resp.Request)
	if err != nil {
		if errors.Is(err, ErrBadResumptionToken) {
			resp.Error = append(resp.Error, ErrBadResumptionToken)
			return nil
		}
		return fmt.Errorf("cannot get request config: %w", err)
	}

	if cfg.DatasetID == "" {
		resp.Error = append(resp.Error, ErrBadArgument)
		return nil
	}

	res, err := s.store.ListIdentifiers(ctx, &cfg)
	if err != nil {
		return fmt.Errorf("error during listIdentifiers: %w", err)
	}

	if len(res.Errors) != 0 {
		resp.Error = res.Errors
		return nil
	}

	if len(res.Headers) == 0 {
		resp.Error = append(resp.Error, ErrNoRecordsMatch)
		return nil
	}

	if cfg.TotalSize == 0 && res.Total > 0 {
		s.m.Lock()
		cfg.TotalSize = res.Total
		s.steps[cfg.ID] = cfg
		s.m.Unlock()
	}

	resp.ListIdentifiers = &ListIdentifiers{
		Headers:         res.Headers,
		ResumptionToken: cfg.NextResumptionToken(&res),
	}

	return nil
}

func (s *Service) handleListRecords(resp *Response) error {
	ctx := context.TODO()

	cfg, err := s.requestConfig(resp.Request)
	if err != nil {
		if errors.Is(err, ErrBadResumptionToken) {
			resp.Error = append(resp.Error, ErrBadResumptionToken)
			return nil
		}
		return fmt.Errorf("cannot get request config: %w", err)
	}

	if cfg.DatasetID == "" {
		resp.Error = append(resp.Error, ErrBadArgument)
		return nil
	}

	res, err := s.store.ListRecords(ctx, &cfg)
	if err != nil {
		return fmt.Errorf("error during listRecords: %w", err)
	}

	if len(res.Errors) != 0 {
		resp.Error = res.Errors
		return nil
	}

	if len(res.Records) == 0 {
		resp.Error = append(resp.Error, ErrNoRecordsMatch)
		return nil
	}

	if cfg.TotalSize == 0 && res.Total > 0 {
		s.m.Lock()
		cfg.TotalSize = res.Total
		s.steps[cfg.ID] = cfg
		s.m.Unlock()
	}

	resp.ListRecords = &ListRecords{
		Records:         res.Records,
		ResumptionToken: cfg.NextResumptionToken(&res),
	}

	return nil
}

func (s *Service) handleGetRecord(resp *Response) error {
	ctx := context.TODO()

	q := resp.Request.RequestConfig()

	record, errors, err := s.store.GetRecord(ctx, &q)
	if err != nil {
		return fmt.Errorf("error during listRecords: %w", err)
	}

	if len(errors) != 0 {
		resp.Error = errors
		return nil
	}

	resp.GetRecord = &GetRecord{
		Record: record,
	}

	return nil
}
