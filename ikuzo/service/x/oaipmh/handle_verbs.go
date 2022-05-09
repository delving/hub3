package oaipmh

import (
	"context"
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
	case VerbListSets:
		err = s.handleListSets(resp)
	case VerbListIdentifiers:
		err = s.handleListIdentifiers(resp)
	case VerbListRecords:
		err = s.handleListRecords(resp)
	case VerbGetRecord:
		err = s.handleGetRecord(resp)
	default:
		resp.Error = append(resp.Error, ErrBadVerb)
	}

	if err != nil {
		return nil, fmt.Errorf("unexpected error processing OAI-PMH")
	}

	if len(resp.Error) != 0 {
		resp.Request = &Request{BaseURL: req.BaseURL}
	}

	return resp, nil
}

func (s *Service) handleIdentify(resp *Response) {
	resp.Identify = &Identify{
		RepositoryName:    resp.Request.orgConfig.OAIPMH.RepositoryName,
		BaseURL:           resp.Request.BaseURL,
		ProtocolVersion:   "2.0",
		AdminEmail:        resp.Request.orgConfig.OAIPMH.AdminEmails,
		DeletedRecord:     "persistent",
		EarliestDatestamp: "1970-01-01T00:00:00Z",
		Granularity:       "YYYY-MM-DDThh:mm:ssZ",
	}
}

func (s *Service) handleListMetadataFormats(resp *Response) {
	// TODO(kiivihal): implement check for identifier
	formats := []MetadataFormat{
		{
			MetadataPrefix:    "edm",
			Schema:            "",
			MetadataNamespace: "http://www.europeana.eu/schemas/edm/",
		},
		{
			MetadataPrefix:    "rdf",
			Schema:            "",
			MetadataNamespace: "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
		},
	}
	resp.ListMetadataFormats = &ListMetadataFormats{
		MetadataFormat: formats,
	}
}

func (s *Service) handleListSets(resp *Response) error {
	ctx := context.TODO()

	q := resp.Request.QueryConfig()

	sets, errors, err := s.store.ListSets(ctx, &q)
	if err != nil {
		return fmt.Errorf("error during listsets: %w", err)
	}

	if len(errors) != 0 {
		resp.Error = errors
		return nil
	}

	resp.ListSets = &ListSets{
		Set: sets,
	}

	return nil
}

func (s *Service) handleListIdentifiers(resp *Response) error {
	ctx := context.TODO()

	q := resp.Request.QueryConfig()

	headers, errors, err := s.store.ListIdentifiers(ctx, &q)
	if err != nil {
		return fmt.Errorf("error during listIdentifiers: %w", err)
	}

	if len(errors) != 0 {
		resp.Error = errors
		return nil
	}

	resp.ListIdentifiers = &ListIdentifiers{
		Headers:         headers,
		ResumptionToken: q.NextResumptionToken(),
	}

	return nil
}

func (s *Service) handleListRecords(resp *Response) error {
	ctx := context.TODO()

	q := resp.Request.QueryConfig()

	records, errors, err := s.store.ListRecords(ctx, &q)
	if err != nil {
		return fmt.Errorf("error during listRecords: %w", err)
	}

	if len(errors) != 0 {
		resp.Error = errors
		return nil
	}

	resp.ListRecords = &ListRecords{
		Records:         records,
		ResumptionToken: q.NextResumptionToken(),
	}

	return nil
}

func (s *Service) handleGetRecord(resp *Response) error {
	ctx := context.TODO()

	q := resp.Request.QueryConfig()

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
