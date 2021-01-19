package ead

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/delving/hub3/hub3/ead"
	"github.com/delving/hub3/ikuzo/service/x/oaipmh"
	"github.com/kiivihal/goharvest/oai"
	"github.com/rs/zerolog/log"
)

type EADHarvester struct {
	s     *Service
	OrgID string
}

func NewEADHarvester(s *Service) (EADHarvester, error) {
	if s == nil {
		return EADHarvester{}, fmt.Errorf("nil value not allowed for ead.Service")
	}

	return EADHarvester{s: s}, nil
}

func (e *EADHarvester) ProcessEadFromOai(r *oai.Response) {
	if r.Request.Verb != oaipmh.VerbListRecords {
		log.Warn().Str("verb", r.Request.Verb).Msg("verb is not supported for getting ead records")
		return
	}

	for _, record := range r.ListRecords.Records {
		record := record

		log.Info().
			Str("identifier", record.Header.Identifier).
			Strs("set", record.Header.SetSpec).
			Str("lastModified", record.Header.DateStamp).
			Str("status", record.Header.Status).
			Msg("ead file entry")

		if err := e.processRecord(&record); err != nil {
			log.Error().
				Str("identifier", record.Header.Identifier).
				Err(err).
				Msg("unable to process ead record")
		}
	}
}

func (e *EADHarvester) processRecord(record *oai.Record) error {
	archiveID := record.Header.Identifier
	body := record.Metadata.Body

	if len(body) == 0 {
		return fmt.Errorf("metadata.Body cannot be empty")
	}

	r := bytes.NewReader(body)

	_, meta, err := e.s.SaveEAD(r, int64(len(body)), archiveID, e.OrgID)
	if err != nil {
		if errors.Is(err, ErrTaskAlreadySubmitted) {
			e.s.M.IncAlreadyQueued()
			return nil
		}

		e.s.M.IncFailed()

		return err
	}

	t, err := e.s.NewTask(&meta)
	if err != nil {
		e.s.M.IncAlreadyQueued()
		return err
	}

	log.Info().
		Str("orgID", t.Meta.OrgID).
		Str("datasetID", t.Meta.OrgID).
		Str("taskID", t.ID).
		Msg("submitted task for indexing via OAI-PMH")

	return nil
}

type MetsHarvester struct {
	c     *ead.DaoClient
	OrgID string
}

func NewMetsHarvest(c *ead.DaoClient) (MetsHarvester, error) {
	if c == nil {
		return MetsHarvester{}, fmt.Errorf("nil value not allowed for DaoClient")
	}

	return MetsHarvester{
		c: c,
	}, nil
}

func (m *MetsHarvester) ProcessMetsFromOai(r *oai.Response) {
	if r.Request.Verb != oaipmh.VerbListIdentifiers {
		log.Warn().Str("verb", r.Request.Verb).Msg("verb is not supported for getting mets headers")
		return
	}

	for _, header := range r.ListIdentifiers.Headers {
		log.Info().
			Str("identifier", header.Identifier).
			Strs("set", header.SetSpec).
			Str("lastModified", header.DateStamp).
			Str("status", header.Status).
			Msg("mets entry")

		if err := m.processHeader(header); err != nil {
			log.Error().
				Str("identifier", header.Identifier).
				Err(err).
				Msg("unable to process mets header")
		}
	}
}

func (m MetsHarvester) processHeader(header oai.Header) error {
	uuid := header.Identifier

	archiveID, _ := extractSpecs(header.SetSpec)
	if archiveID == "" {
		return fmt.Errorf("archiveID cannot be empty")
	}

	cfg, err := m.c.GetDaoConfig(archiveID, uuid)
	if err != nil {
		return err
	}

	if err := m.c.PublishFindingAid(&cfg); err != nil {
		return err
	}

	return nil
}

func extractSpecs(specs []string) (archiveID, inventoryID string) {
	for _, spec := range specs {
		parts := strings.SplitN(spec, ":", 2)

		switch prefix := parts[0]; prefix {
		case "TOE":
			archiveID = parts[1]
		case "INV":
			inventoryID = parts[1]
		}
	}

	return archiveID, inventoryID
}
