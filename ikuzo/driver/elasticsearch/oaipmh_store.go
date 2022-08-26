package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/mappingxml"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/delving/hub3/ikuzo/service/x/oaipmh"
	"github.com/olivere/elastic/v7"
)

var _ oaipmh.Store = (*OAIPMHStore)(nil)

type OAIPMHStore struct {
	c            *Client
	ResponseSize int
}

func (c *Client) NewOAIPMHStore() (*OAIPMHStore, error) {
	return &OAIPMHStore{
		c:            c,
		ResponseSize: 100, // default 100
	}, nil
}

func (o *OAIPMHStore) ListSets(ctx context.Context, q *oaipmh.RequestConfig) (res oaipmh.Resumable, err error) {
	query := elastic.NewBoolQuery().
		Must(elastic.NewTermQuery("meta.orgID", q.OrgID)).
		Must(elastic.NewTermQuery("meta.tags", "narthex"))

	specCountAgg := elastic.NewCardinalityAggregation().
		Field("meta.spec")

	agg := elastic.NewCompositeAggregation().
		Sources(
			elastic.NewCompositeAggregationTermsValuesSource("datasets").Field("meta.spec"),
		).Size(1000)

	search := o.c.search.Search().
		Index(IndexNames{}.GetIndexName(q.OrgID)).
		TrackTotalHits(false).
		Query(query).
		Size(0).
		Aggregation("datasets", agg).
		Aggregation("specCount", specCountAgg)

	logSearchService(search)

	resp, err := search.Do(ctx)
	if err != nil {
		o.c.log.Error().Err(err).Msg("unable to list sets")
		return res, err
	}

	datasets, ok := resp.Aggregations.Composite("datasets") // change with composite later
	if !ok {
		res.Errors = append(res.Errors, oaipmh.ErrNoSetHierachy)
		return res, nil
	}

	for _, spec := range datasets.Buckets {
		specLabel := spec.Key["datasets"].(string)

		res.Sets = append(res.Sets, oaipmh.Set{
			SetSpec: specLabel,
			SetDescription: oaipmh.Description{
				Body: []byte(fmt.Sprintf("<totalRecords>%d</totalRecords>", int(spec.DocCount))),
			},
		})
	}

	if len(res.Sets) == 0 {
		res.Errors = append(res.Errors, oaipmh.ErrNoSetHierachy)
	}

	res.StorePayload, err = encodeCompositeSearchAfter(datasets.AfterKey)
	if err != nil {
		return res, err
	}

	specCount, ok := resp.Aggregations.Cardinality("specCount")
	if ok {
		res.Total = int(*specCount.Value)
	}

	return res, err
}

type resumableResponse struct {
	records    []json.RawMessage
	total      int64
	pitPayload string // payload for point in time parsing
}

func (o *OAIPMHStore) getRecords(ctx context.Context, q *oaipmh.RequestConfig, headersOnly bool) (resp resumableResponse, err error) {
	query := elastic.NewBoolQuery().
		Must(elastic.NewTermQuery("meta.orgID", q.OrgID)).
		Must(elastic.NewTermQuery("meta.tags", "narthex"))

	if q.DatasetID != "" {
		query = query.Must(elastic.NewTermQuery("meta.spec", q.DatasetID))
	}

	if q.FirstRequest.From != "" || q.FirstRequest.Until != "" {

		timeRange := elastic.NewRangeQuery("meta.modified")
		if q.FirstRequest.From != "" {
			timeRange = timeRange.Gte(q.FirstRequest.From)
		}

		if q.FirstRequest.Until != "" {
			timeRange = timeRange.Lte(q.FirstRequest.Until)
		}

		query = query.Must(timeRange)
	}

	if q.CurrentRequest.HarvestID == "" {
		openResp, openErr := o.c.search.OpenPointInTime(IndexNames{}.GetIndexName(q.OrgID)).
			KeepAlive("1m").
			Pretty(true).
			Do(context.Background())
		if openErr != nil {
			return resp, openErr
		}

		q.StoreCursor = openResp.Id
	}

	search := o.c.search.Search().
		// Index(IndexNames{}.GetIndexName(q.OrgID)).
		PointInTime(elastic.NewPointInTimeWithKeepAlive(q.StoreCursor, "1m")).
		Sort("_shard_doc", true).
		Size(o.ResponseSize).
		Query(query)

	if headersOnly {
		fsc := elastic.NewFetchSourceContext(true)
		fsc.Include("meta")
		search = search.FetchSourceContext(fsc)
	}

	if !q.IsResumedRequest() {
		search = search.TrackTotalHits(true)
	} else {
		searchAfter, decodeErr := decodeSearchAfter(q.CurrentRequest.StorePayload)
		if decodeErr != nil {
			return resp, decodeErr
		}

		search = search.SearchAfter(searchAfter...)
	}

	res, err := search.Do(ctx)
	if err != nil {
		o.c.log.Error().Err(err).Msg("unable to get record")
		return resp, err
	}

	if !q.IsResumedRequest() {
		resp.total = res.TotalHits()
	}

	var hit *elastic.SearchHit
	for _, hit = range res.Hits.Hits {
		resp.records = append(resp.records, hit.Source)
	}

	if hit != nil && len(hit.Sort) > 0 {
		nextSearchAfter, encodeErr := encodeSearchAfter(hit.Sort)
		if encodeErr != nil {
			return resp, encodeErr
		}

		resp.pitPayload = nextSearchAfter
	}

	return resp, err
}

func (o *OAIPMHStore) ListIdentifiers(ctx context.Context, q *oaipmh.RequestConfig) (res oaipmh.Resumable, err error) {
	resp, err := o.getRecords(ctx, q, true)
	if err != nil {
		return
	}

	for _, raw := range resp.records {
		rec, getErr := o.getOAIPMHRecord(raw, q.FirstRequest.MetadataPrefix, true)
		if getErr != nil {
			return res, getErr
		}

		res.Headers = append(res.Headers, rec.Header)
	}

	res.Total = int(resp.total)
	res.StorePayload = resp.pitPayload

	return res, err
}

func (o *OAIPMHStore) ListRecords(ctx context.Context, q *oaipmh.RequestConfig) (res oaipmh.Resumable, err error) {
	resp, err := o.getRecords(ctx, q, false)

	log.Printf("metadataPrefix: %#v", q.FirstRequest)

	for _, raw := range resp.records {
		rec, getErr := o.getOAIPMHRecord(raw, q.FirstRequest.MetadataPrefix, true)
		if getErr != nil {
			return res, getErr
		}

		res.Records = append(res.Records, rec)
	}

	res.Total = int(resp.total)
	res.StorePayload = resp.pitPayload

	return res, err
}

func (o *OAIPMHStore) GetRecord(ctx context.Context, q *oaipmh.RequestConfig) (record oaipmh.Record, errors []oaipmh.Error, err error) {
	if q.FirstRequest.Identifier == "" {
		errors = append(errors, oaipmh.ErrIdDoesNotExist)
		return
	}

	search := o.c.search.Get().
		Index(IndexNames{}.GetIndexName(q.OrgID)).
		Id(q.FirstRequest.Identifier)

	res, err := search.Do(ctx)
	if err != nil {
		o.c.log.Error().Err(err).Msg("unable to get record")
		return
	}

	record, err = o.getOAIPMHRecord(res.Source, q.FirstRequest.MetadataPrefix, false)
	if err != nil {
		o.c.log.Error().Err(err).Msg("unable to serialize record")
		return
	}

	return record, errors, err
}

func (o *OAIPMHStore) serialize(format string, fg *fragments.FragmentGraph, w io.Writer) error {
	g, err := fg.Graph()
	if err != nil {
		o.c.log.Error().Err(err).Msg("unable to get rdf.Graph")
		return err
	}

	switch format {
	case "ntriples":
		fmt.Fprintln(w, "<!CDATA[")

		if err := ntriples.Serialize(g, w); err != nil {
			return err
		}

		fmt.Fprintln(w, "]]>")

		return nil
	case "rdfxml", "oai_dc":
		iri, err := rdf.NewIRI(fg.Meta.GetEntryURI())
		if err != nil {
			return err
		}

		cfg := &mappingxml.FilterConfig{
			Subject:         iri,
			URIPrefixFilter: "urn:private",
		}

		err = mappingxml.Serialize(g, w, cfg)
		if err != nil {
			o.c.log.Error().Err(err).Msg("unable to get serialize mappingxml")
			return err
		}

		return nil
	}

	return oaipmh.ErrCannotDisseminateFormat
}

func (o *OAIPMHStore) getOAIPMHRecord(source json.RawMessage, format string, onlyHeader bool) (record oaipmh.Record, err error) {
	fg, err := decodeFragmentGraph(source)
	if err != nil {
		o.c.log.Error().Err(err).Msg("unable to get FragmentGraph")
		return
	}

	var buf bytes.Buffer

	if err := o.serialize(format, fg, &buf); err != nil {
		return record, err
	}

	record.Header.Identifier = fg.Meta.HubID
	record.Header.DateStamp = fg.Meta.LastModified().UTC().Format(oaipmh.TimeFormat)
	record.Header.SetSpec = []string{fg.Meta.Spec}
	record.Metadata = oaipmh.Metadata{Body: buf.Bytes()}

	return record, nil
}

func (o *OAIPMHStore) ListMetadataFormats(ctx context.Context, q *oaipmh.RequestConfig) (formats []oaipmh.MetadataFormat, err error) {
	formats = []oaipmh.MetadataFormat{
		// {
		// MetadataPrefix:    "ntriples",
		// Schema:            "",
		// MetadataNamespace: "http://www.europeana.eu/schemas/edm/",
		// },
		{
			MetadataPrefix:    "rdfxml",
			Schema:            "",
			MetadataNamespace: "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
		},
		{
			MetadataPrefix:    "oai_dc",
			Schema:            "http://www.openarchives.org/OAI/2.0/oai_dc.xsd",
			MetadataNamespace: "http://www.openarchives.org/OAI/2.0/oai_dc/",
		},
	}

	return formats, err
}

func decodeFragmentGraph(hit json.RawMessage) (*fragments.FragmentGraph, error) {
	r := new(fragments.FragmentGraph)
	if err := json.Unmarshal(hit, r); err != nil {
		return nil, err
	}

	return r, nil
}
