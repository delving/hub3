package bulk

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/service/x/index"
	"github.com/rs/zerolog/log"
)

type Request struct {
	HubID         string `json:"hubId"`
	OrgID         string `json:"orgID"`
	Spec          string `json:"dataset"`
	LocalID       string `json:"localID"`
	NamedGraphURI string `json:"graphUri"`
	RecordType    string `json:"type"`
	Action        string `json:"action"`
	ContentHash   string `json:"contentHash"`
	Graph         string `json:"graph"`
	GraphMimeType string `json:"graphMimeType"`
	SubjectType   string `json:"subjectType"`
}

func (req *Request) valid() error {
	if req.Graph == "" {
		return fmt.Errorf("empty graph during indexing is not allowed")
	}

	if req.OrgID == "" || req.HubID == "" || req.Spec == "" {
		return fmt.Errorf("orgID, hubID and spec cannot be empty in bulk request")
	}

	if req.GraphMimeType == "" {
		log.Warn().Str("svc", "bulk").Msgf("reverting to default. graphMimeType must be set when bulk action is 'index'")

		req.GraphMimeType = "application/ld+json"
	}

	return nil
}

func (req *Request) createFragmentBuilder(revision int) (*fragments.FragmentBuilder, error) {
	fg := fragments.NewFragmentGraph()
	fg.Meta.OrgID = req.OrgID
	fg.Meta.HubID = req.HubID
	fg.Meta.Spec = req.Spec
	fg.Meta.Revision = int32(revision)
	fg.Meta.NamedGraphURI = req.NamedGraphURI
	fg.Meta.EntryURI = fg.GetAboutURI()
	fg.Meta.Tags = []string{"narthex", "mdr"}

	fb := fragments.NewFragmentBuilder(fg)
	err := fb.ParseGraph(strings.NewReader(req.Graph), req.GraphMimeType)
	// log.Printf("Unable to parse the graph: %s", err)
	if err != nil {
		return fb, fmt.Errorf("source RDF is not in format: %s", req.GraphMimeType)
	}

	return fb, nil
}

func (req *Request) processV1(fb *fragments.FragmentBuilder, bi index.BulkIndex) error {
	fb.GetSortedWebResources()

	indexDoc, err := fragments.CreateV1IndexDoc(fb)
	if err != nil {
		log.Info().Msgf("Unable to create index doc: %s", err)
		return err
	}

	b, err := json.Marshal(indexDoc)
	if err != nil {
		return err
	}

	m := &domainpb.IndexMessage{
		OrganisationID: req.OrgID,
		DatasetID:      req.Spec,
		RecordID:       req.HubID,
		IndexName:      config.Config.ElasticSearch.GetV1IndexName(), // TODO(kiivihal): remove config later
		Source:         b,
	}

	if err := bi.Publish(context.Background(), m); err != nil {
		return err
	}

	// TODO(kiivihal): add posthook later

	return nil
}

func (req *Request) processV2(fb *fragments.FragmentBuilder, bi index.BulkIndex) error {
	m, err := fb.Doc().IndexMessage()
	if err != nil {
		return err
	}

	if err := bi.Publish(context.Background(), m); err != nil {
		return err
	}

	return nil
}

func (req *Request) processFragments(fb *fragments.FragmentBuilder, bi index.BulkIndex) error {
	return fb.IndexFragments(bi)
}
