package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/olivere/elastic/v7"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/mappingxml"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/delving/hub3/ikuzo/service/x/oaipmh"
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
		Must(elastic.NewTermQuery("meta.orgID", q.OrgID))

	query = addFilters(q, query)

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
				Body: fmt.Appendf(nil, "<totalRecords>%d</totalRecords>", int(spec.DocCount)),
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

type recordWrapper struct {
	HubID string
	Data  json.RawMessage
}

type resumableResponse struct {
	records    []recordWrapper
	total      int64
	pitPayload string // payload for point in time parsing
}

func addFilters(q *oaipmh.RequestConfig, query *elastic.BoolQuery) *elastic.BoolQuery {
	if len(q.Filters) > 0 {
		fq := elastic.NewBoolQuery()
		for _, filt := range q.Filters {
			fq = fq.Should(elastic.NewTermQuery("meta.tags", filt))
		}
		query = query.Must(fq)
	}

	return query
}

func (o *OAIPMHStore) getRecords(ctx context.Context, q *oaipmh.RequestConfig, headersOnly bool) (resp resumableResponse, err error) {
	query := elastic.NewBoolQuery().
		Must(elastic.NewTermQuery("meta.orgID", q.OrgID))

	query = addFilters(q, query)

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
		pitID, openErr := o.c.pit.CreatePIT(ctx, IndexNames{}.GetIndexName(q.OrgID), "1m")
		if openErr != nil {
			slog.Error("error opening point in time API", "error", openErr)
			return resp, fmt.Errorf("open point in time error; %w", openErr)
		}

		q.StoreCursor = pitID
	}

	search := o.c.search.Search().
		PointInTime(elastic.NewPointInTimeWithKeepAlive(q.StoreCursor, "1m")).
		Sort("_id", true).
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
		slog.Error("raw result", "resp", res, "query", q)
		return resp, fmt.Errorf("error during search; %w", err)
	}

	if !q.IsResumedRequest() {
		resp.total = res.TotalHits()
	}

	var hit *elastic.SearchHit
	for _, hit = range res.Hits.Hits {
		wrapper := recordWrapper{
			HubID: hit.Id,
			Data:  hit.Source,
		}
		resp.records = append(resp.records, wrapper)
	}

	if hit != nil && len(hit.Sort) > 0 {
		nextSearchAfter, encodeErr := encodeSearchAfter(hit.Sort)
		if encodeErr != nil {
			return resp, encodeErr
		}

		resp.pitPayload = nextSearchAfter
	}

	return resp, nil
}

func (o *OAIPMHStore) ListIdentifiers(ctx context.Context, q *oaipmh.RequestConfig) (res oaipmh.Resumable, err error) {
	resp, err := o.getRecords(ctx, q, true)
	if err != nil {
		return
	}

	for _, wrapper := range resp.records {
		rec, getErr := o.getOAIPMHRecord(wrapper, q.FirstRequest.MetadataPrefix, true)
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
	if err != nil {
		slog.Error("unable to get records", "error", err, "config", q)
		return res, fmt.Errorf("es store ListRecords error; %w", err)
	}

	for _, raw := range resp.records {
		rec, getErr := o.getOAIPMHRecord(raw, q.FirstRequest.MetadataPrefix, false)
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
		errors = append(errors, oaipmh.ErrIDDoesNotExist)
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

	wrapper := recordWrapper{
		HubID: res.Id,
		Data:  res.Source,
	}

	record, err = o.getOAIPMHRecord(wrapper, q.FirstRequest.MetadataPrefix, false)
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
	case "rdfxml", "oai_dc", "edm":
		slog.Info("record graph", "len", g.Len(), "uri", fg.Meta.GetEntryURI())

		err = mappingxml.Serialize(g, w, nil)
		if err != nil {
			o.c.log.Error().Err(err).Msg("unable to get serialize mappingxml")
			return err
		}

		return nil
	case "edm-strict":
		iri, err := rdf.NewIRI(fg.Meta.GetEntryURI())
		if err != nil {
			return err
		}

		cfg := &mappingxml.FilterConfig{
			Subject:             iri,
			URIPrefixFilter:     "urn:private",
			ExcludePrefixes:     []string{"http://schemas.delving.eu/nave/terms/"},
			ExcludeTypePrefixes: []string{"http://schemas.delving.eu/nave/terms/"},
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

func (o *OAIPMHStore) getOAIPMHRecord(wrapper recordWrapper, format string, onlyHeader bool) (record oaipmh.Record, err error) {
	fg, err := decodeFragmentGraph(wrapper.Data)
	if err != nil {
		o.c.log.Error().Err(err).Str("hubID", wrapper.HubID).RawJSON("rawMessage", wrapper.Data).Msg("unable to get FragmentGraph")
		return record, fmt.Errorf("unable to decode oai-pmh record %q; %w", wrapper.HubID, err)
	}

	var buf bytes.Buffer

	if err := o.serialize(format, fg, &buf); err != nil {
		return record, err
	}

	record.Header.Identifier = fg.Meta.HubID
	record.Header.DateStamp = fg.Meta.LastModified().UTC().Format(oaipmh.TimeFormat)
	record.Header.SetSpec = []string{fg.Meta.Spec}
	if !onlyHeader {
		record.Metadata = oaipmh.Metadata{Body: buf.Bytes()}
	}

	return record, nil
}

func (o *OAIPMHStore) ListMetadataFormats(ctx context.Context, q *oaipmh.RequestConfig) (formats []oaipmh.MetadataFormat, err error) {
	formats = []oaipmh.MetadataFormat{
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
		{
			MetadataPrefix:    "edm",
			Schema:            "",
			MetadataNamespace: "http://www.europeana.eu/schemas/edm/",
		},
		{
			MetadataPrefix:    "edm-strict",
			Schema:            "",
			MetadataNamespace: "http://www.europeana.eu/schemas/edm/",
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
