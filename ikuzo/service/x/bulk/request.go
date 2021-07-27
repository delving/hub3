// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	DatasetID     string `json:"dataset"`
	LocalID       string `json:"localID"`
	NamedGraphURI string `json:"graphUri"`
	RecordType    string `json:"type"`
	Action        string `json:"action"`
	ContentHash   string `json:"contentHash"`
	Graph         string `json:"graph"`
	GraphMimeType string `json:"graphMimeType"`
	SubjectType   string `json:"subjectType"`
	Revision      int    `json:"revision"`
	Tags          string `json:"tags,omitempty"`
}

func (req *Request) valid() error {
	if req.Graph == "" {
		return fmt.Errorf("empty graph during indexing is not allowed")
	}

	if req.OrgID == "" || req.HubID == "" || req.DatasetID == "" {
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
	fg.Meta.Spec = req.DatasetID
	fg.Meta.Revision = int32(revision)
	fg.Meta.NamedGraphURI = req.NamedGraphURI
	fg.Meta.EntryURI = fg.GetAboutURI()
	fg.Meta.Tags = []string{"narthex", "mdr"}

	if req.Tags != "" {
		tags := strings.Split(req.Tags, ",")
		for _, tag := range tags {
			if strings.HasPrefix(tag, "ck_") {
				if fg.Tree == nil {
					fg.Tree = &fragments.Tree{}
				}

				fg.Tree.Type = strings.TrimPrefix(tag, "ck_")

				continue
			}

			fg.Meta.Tags = append(fg.Meta.Tags, strings.TrimSpace(tag))
		}
	}

	if tags, ok := config.Config.DatasetTagMap.Get(req.DatasetID); ok {
		fg.Meta.Tags = append(fg.Meta.Tags, tags...)
	}

	fb := fragments.NewFragmentBuilder(fg)

	err := fb.ParseGraph(strings.NewReader(req.Graph), req.GraphMimeType)
	if err != nil {
		return fb, fmt.Errorf("source RDF is not in format: %s", req.GraphMimeType)
	}

	return fb, nil
}

func (req *Request) processV1(ctx context.Context, fb *fragments.FragmentBuilder, bi index.BulkIndex) error {
	return processV1(ctx, fb, bi)
}

func processV1(ctx context.Context, fb *fragments.FragmentBuilder, bi index.BulkIndex) error {
	fb.GetSortedWebResources(ctx)

	indexDoc, err := fragments.CreateV1IndexDoc(fb)
	if err != nil {
		log.Info().Msgf("Unable to create index doc: %s", err)
		return err
	}

	b, err := json.Marshal(indexDoc)
	if err != nil {
		return err
	}

	fg := fb.FragmentGraph()

	m := &domainpb.IndexMessage{
		OrganisationID: fg.Meta.OrgID,
		DatasetID:      fg.Meta.Spec,
		RecordID:       fg.Meta.HubID,
		IndexName:      config.Config.ElasticSearch.GetV1IndexName(), // TODO(kiivihal): remove config later
		Source:         b,
	}

	if err := bi.Publish(ctx, m); err != nil {
		return err
	}

	return nil
}

func (req *Request) processV2(ctx context.Context, fb *fragments.FragmentBuilder, bi index.BulkIndex) error {
	return processV2(ctx, fb, bi)
}

func processV2(ctx context.Context, fb *fragments.FragmentBuilder, bi index.BulkIndex) error {
	m, err := fb.Doc().IndexMessage()
	if err != nil {
		return err
	}

	if err := bi.Publish(ctx, m); err != nil {
		return err
	}

	return nil
}

func (req *Request) processFragments(fb *fragments.FragmentBuilder, bi index.BulkIndex) error {
	return fb.IndexFragments(bi)
}
