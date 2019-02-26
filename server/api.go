// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strconv"
	"strings"

	c "github.com/delving/rapid-saas/config"
	"github.com/delving/rapid-saas/hub3"
	"github.com/delving/rapid-saas/hub3/ead"
	"github.com/delving/rapid-saas/hub3/fragments"
	"github.com/delving/rapid-saas/hub3/harvesting"
	"github.com/delving/rapid-saas/hub3/index"
	"github.com/delving/rapid-saas/hub3/models"
	"github.com/gammazero/workerpool"
	"github.com/gorilla/schema"
	"github.com/kiivihal/rdf2go"

	elastic "github.com/olivere/elastic"

	"github.com/asdine/storm"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/kiivihal/goharvest/oai"
)

var bp *elastic.BulkProcessor
var wp *workerpool.WorkerPool
var ctx context.Context

func init() {
	var err error
	ctx = context.Background()
	bps := index.CreateBulkProcessorService()
	bp, err = bps.Do(ctx)
	if err != nil {
		log.Fatalf("Unable to start BulkProcessor: %#v", err)
	}
	wp = workerpool.New(10)
}

// APIErrorMessage contains the default API error messages
type APIErrorMessage struct {
	HTTPStatus int    `json:"code"`
	Message    string `json:"type"`
	Error      error  `json:"error"`
}

// NewSingleFinalPathHostReverseProxy proxies QueryString of the request url to the target url
func NewSingleFinalPathHostReverseProxy(target *url.URL, relPath string) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path + relPath
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		log.Printf("proxy request: %#v", req)
		log.Printf("proxy request: %#v", req.URL.String())
		log.Printf("proxy request: %#v", req.Body)
	}
	return &httputil.ReverseProxy{Director: director}
}

func searchLabelStatsValues(w http.ResponseWriter, r *http.Request) {
	return
}

func getResourceEntryStats(field string, r *http.Request) (*elastic.SearchResult, error) {

	fieldPath := fmt.Sprintf("resources.entries.%s", field)

	labelAgg := elastic.NewTermsAggregation().Field(fieldPath).Size(100)

	order := r.URL.Query().Get("order")
	switch order {
	case "term":
		labelAgg = labelAgg.OrderByTermAsc()
	default:
		labelAgg = labelAgg.OrderByCountDesc()
	}
	searchLabelAgg := elastic.NewNestedAggregation().Path("resources.entries")
	searchLabelAgg = searchLabelAgg.SubAggregation(field, labelAgg)

	q := elastic.NewBoolQuery()
	q = q.Must(
		elastic.NewTermQuery("meta.docType", fragments.FragmentGraphDocType),
		elastic.NewTermQuery(c.Config.ElasticSearch.OrgIDKey, c.Config.OrgID),
	)
	spec := r.URL.Query().Get("spec")
	if spec != "" {
		q = q.Must(elastic.NewTermQuery(c.Config.ElasticSearch.SpecKey, spec))
	}
	res, err := index.ESClient().Search().
		Index(c.Config.ElasticSearch.IndexName).
		Query(q).
		Size(0).
		Aggregation(field, searchLabelAgg).
		Do(ctx)
	return res, err
}

func searchLabelStats(w http.ResponseWriter, r *http.Request) {

	res, err := getResourceEntryStats("searchLabel", r)
	if err != nil {
		log.Print("Unable to get statistics for searchLabels")
		render.PlainText(w, r, err.Error())
		render.Status(r, http.StatusBadRequest)
		return
	}
	fmt.Printf("total hits: %d\n", res.Hits.TotalHits)
	render.JSON(w, r, res)
	return
}
func predicateStats(w http.ResponseWriter, r *http.Request) {

	res, err := getResourceEntryStats("predicate", r)
	if err != nil {
		log.Print("Unable to get statistics for predicate")
		render.PlainText(w, r, err.Error())
		render.Status(r, http.StatusBadRequest)
		return
	}
	fmt.Printf("total hits: %d\n", res.Hits.TotalHits)
	render.JSON(w, r, res)
	return
}

func eadUpload(w http.ResponseWriter, r *http.Request) {
	spec := r.FormValue("spec")

	_, err := ead.ProcessUpload(r, w, spec, bp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

type rdfUploadForm struct {
	Spec          string `json:"spec"`
	RecordType    string `json:"recordType"`
	TypePredicate string `json:"typePredicate"`
	IDSplitter    string `json:"idSplitter"`
}

func (ruf *rdfUploadForm) isValid() error {
	if ruf.Spec == "" {
		return fmt.Errorf("spec param is required")
	}
	if ruf.RecordType == "" {
		return fmt.Errorf("recordType param is required")
	}
	if ruf.TypePredicate == "" {
		return fmt.Errorf("typePredicate param is required")
	}
	if ruf.IDSplitter == "" {
		return fmt.Errorf("idSplitter param is required")
	}
	return nil
}

var decoder = schema.NewDecoder()

func rdfUpload(w http.ResponseWriter, r *http.Request) {
	in, _, err := r.FormFile("turtle")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var form rdfUploadForm
	err = decoder.Decode(&form, r.PostForm)
	if err != nil {
		log.Printf("Unable to decode form %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = form.isValid()
	if err != nil {
		log.Printf("form is not valid; %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ds, created, err := models.GetOrCreateDataSet(form.Spec)
	if err != nil {
		log.Printf("Unable to get DataSet for %s\n", form.Spec)
		render.PlainText(w, r, err.Error())
		return
	}
	if created {
		err = fragments.SaveDataSet(form.Spec, bp)
		if err != nil {
			log.Printf("Unable to Save DataSet Fragment for %s\n", form.Spec)
			if err != nil {
				render.PlainText(w, r, err.Error())
				return
			}
		}
	}

	ds, err = ds.IncrementRevision()
	if err != nil {
		render.PlainText(w, r, err.Error())
		return
	}

	upl := fragments.NewRDFUploader(
		c.Config.OrgID,
		form.Spec,
		form.RecordType,
		form.TypePredicate,
		form.IDSplitter,
		ds.Revision,
	)

	go func() {
		defer in.Close()
		log.Print("Start creating resource map")
		_, err := upl.Parse(in)
		if err != nil {
			log.Printf("Can't read turtle file: %v", err)
			return
		}
		log.Printf("Start saving fragments.")
		processedFragments, err := upl.IndexFragments(bp)
		if err != nil {
			log.Printf("Can't save fragments: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Saved %d fragments for %s", processedFragments, upl.Spec)
		processed, err := upl.SaveFragmentGraphs(bp)
		if err != nil {
			log.Printf("Can't save records: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Saved %d records for %s", processed, upl.Spec)
		ds.DropOrphans(r.Context(), bp, nil)
	}()

	render.Status(r, http.StatusCreated)
	render.PlainText(w, r, "ok")
	return
}

func skosUpload(w http.ResponseWriter, r *http.Request) {

	// create byte.buffer from the input file
	// get dataset param
	// get dataset object
	// get subjectClass
	// create map subject map[string]bool
	// matcher string for rdf:Type
	// create resourceMap
	// use n-triple / turtle parse to build line by line, see rdf2go libraries for Graph
	// addTriple per line
	// gather subject per type
	// check subject map
	// next

	// alternative approach
	// store all fragments
	// get scanner for spec and rdfType to get subjects back
	// make nested call for elasticsearch: get all objects, do mget on fragments,
	// parse into resource map
	// do next level mget on resource objects
	// parse into resource map
	// add to elastic bulk processor
	in, _, err := r.FormFile("skos")
	if err != nil {
		render.PlainText(w, r, err.Error())
		return
	}
	//io.Copy(w, in)
	var buff bytes.Buffer
	fileSize, err := buff.ReadFrom(in)
	//fmt.Println(fileSize) // this will return you a file size.
	//if err != nil {
	//render.PlainText(w, r, err.Error())
	//return
	//}
	render.PlainText(w, r, fmt.Sprintf("The file is %d bytes long", fileSize))

	jsonld := []map[string]interface{}{}
	err = json.Unmarshal(buff.Bytes(), &jsonld)
	if err != nil {
		render.PlainText(w, r, err.Error())
		return
	}

	log.Printf("found %#v resources", jsonld[0])
	log.Printf("found %d resources", len(jsonld))

	defer in.Close()
	//g := rdf2go.NewGraph("")
	//err = g.Parse(in, "application/ld+json")
	//if err != nil {
	//render.PlainText(w, r, err.Error())
	//return
	//}

	//render.PlainText(w, r, fmt.Sprintf("processed triples: %d", g.Len()))
	return
}

func skosSync(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("uri")
	spec := r.URL.Query().Get("spec")

	ds, created, err := models.GetOrCreateDataSet(spec)
	if err != nil {
		log.Printf("Unable to get DataSet for %s\n", spec)
		render.PlainText(w, r, err.Error())
		return
	}
	if created {
		err = fragments.SaveDataSet(spec, bp)
		if err != nil {
			log.Printf("Unable to Save DataSet Fragment for %s\n", spec)
			if err != nil {
				render.PlainText(w, r, err.Error())
				return
			}
		}
	}

	ds, err = ds.IncrementRevision()
	if err != nil {
		render.PlainText(w, r, err.Error())
		return
	}

	g := rdf2go.NewGraph("")
	err = g.LoadURI(targetURL)
	if err != nil {
		log.Printf("Unable to get skos for %s\n", targetURL)
		render.PlainText(w, r, err.Error())
		return
	}

	render.PlainText(w, r, fmt.Sprintf("processed triples: %d", g.Len()))
	return
}

func csvDelete(w http.ResponseWriter, r *http.Request) {
	conv := fragments.NewCSVConvertor()
	conv.DefaultSpec = r.FormValue("defaultSpec")

	if conv.DefaultSpec == "" {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "defaultSpec is a required field")
		return
	}

	ds, _, err := models.GetOrCreateDataSet(conv.DefaultSpec)
	if err != nil {
		log.Printf("Unable to get DataSet for %s\n", conv.DefaultSpec)
		render.PlainText(w, r, err.Error())
		return
	}
	_, err = ds.DropRecords(ctx, wp)
	if err != nil {
		log.Printf("Unable to delete all fragments for %s: %s", conv.DefaultSpec, err.Error())
		render.Status(r, http.StatusBadRequest)
		return
	}

	render.Status(r, http.StatusNoContent)
	return
}

func csvUpload(w http.ResponseWriter, r *http.Request) {
	in, _, err := r.FormFile("csv")
	if err != nil {
		render.PlainText(w, r, err.Error())
		return
	}

	conv := fragments.NewCSVConvertor()
	conv.InputFile = in
	conv.SubjectColumn = r.FormValue("subjectColumn")
	conv.SubjectClass = r.FormValue("subjectClass")
	conv.SubjectURIBase = r.FormValue("subjectURIBase")
	conv.Separator = r.FormValue("separator")
	conv.PredicateURIBase = r.FormValue("predicateURIBase")
	conv.SubjectColumn = r.FormValue("subjectColumn")
	conv.ObjectResourceColumns = strings.Split(r.FormValue("objectResourceColumns"), ",")
	conv.ObjectIntegerColumns = strings.Split(r.FormValue("objectIntegerColumns"), ",")
	conv.ObjectURIFormat = r.FormValue("objectURIFormat")
	conv.DefaultSpec = r.FormValue("defaultSpec")
	conv.ThumbnailURIBase = r.FormValue("thumbnailURIBase")
	conv.ThumbnailColumn = r.FormValue("thumbnailColumn")
	conv.ManifestColumn = r.FormValue("manifestColumn")
	conv.ManifestURIBase = r.FormValue("manifestURIBase")
	conv.ManifestLocale = r.FormValue("manifestLocale")

	if conv.Separator == "" {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "Separator is a required field. When ';' is the separator you can escape it as '%3B'")
		return
	}

	ds, created, err := models.GetOrCreateDataSet(conv.DefaultSpec)
	if err != nil {
		log.Printf("Unable to get DataSet for %s\n", conv.DefaultSpec)
		render.PlainText(w, r, err.Error())
		return
	}
	if created {
		err = fragments.SaveDataSet(conv.DefaultSpec, bp)
		if err != nil {
			log.Printf("Unable to Save DataSet Fragment for %s\n", conv.DefaultSpec)
			if err != nil {
				render.PlainText(w, r, err.Error())
				return
			}
		}
	}

	ds, err = ds.IncrementRevision()
	if err != nil {
		render.PlainText(w, r, err.Error())
		return
	}

	triplesCreated, rowsSeen, err := conv.IndexFragments(bp, ds.Revision)
	conv.RowsProcessed = rowsSeen
	conv.TriplesCreated = triplesCreated
	log.Printf("Processed %d csv rows\n", rowsSeen)
	if err != nil {
		render.PlainText(w, r, err.Error())
		return
	}

	_, err = ds.DropOrphans(ctx, bp, wp)
	if err != nil {
		render.PlainText(w, r, err.Error())
		return
	}

	render.Status(r, http.StatusCreated)
	//render.PlainText(w, r, "ok")
	render.JSON(w, r, conv)
	return
}

func bulkSyncStart(w http.ResponseWriter, r *http.Request) {

	//host := r.URL.Query().Get("host")
	//index := r.URL.Query().Get("index")

}

func bulkSyncList(w http.ResponseWriter, r *http.Request) {

	//host := r.URL.Query().Get("host")
	//index := r.URL.Query().Get("index")

}

func bulkSyncProgress(w http.ResponseWriter, r *http.Request) {

}

func bulkSyncCancel(w http.ResponseWriter, r *http.Request) {

}

// bulkApi receives bulkActions in JSON form (1 per line) and processes them in
// ingestion pipeline.
func bulkAPI(w http.ResponseWriter, r *http.Request) {
	response, err := hub3.ReadActions(ctx, r.Body, bp, wp)
	if err != nil {
		log.Println("Unable to read actions")
		errR := ErrRender(err)
		// todo fix errr renderer for better narthex consumption.
		_ = errR.Render(w, r)
		render.Render(w, r, errR)
		return
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
	return
}

// bindPMHRequest the query parameters to the OAI-Request
func bindPMHRequest(r *http.Request) oai.Request {
	baseURL := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	q := r.URL.Query()
	req := oai.Request{
		Verb:            q.Get("verb"),
		MetadataPrefix:  q.Get("metadataPrefix"),
		Set:             q.Get("set"),
		From:            q.Get("from"),
		Until:           q.Get("until"),
		Identifier:      q.Get("identifier"),
		ResumptionToken: q.Get("resumptionToken"),
		BaseURL:         baseURL,
	}
	return req
}

// oaiPmhEndpoint processed OAI-PMH request and returns the results
func oaiPmhEndpoint(w http.ResponseWriter, r *http.Request) {
	req := bindPMHRequest(r)
	log.Println(req)
	resp := harvesting.ProcessVerb(&req)
	render.XML(w, r, resp)
}

// RenderLODResource returns a list of matching fragments
// for a LOD resource. This mimicks a SPARQL describe request
func RenderLODResource(w http.ResponseWriter, r *http.Request) {

	lodKey := r.URL.Path
	if c.Config.LOD.SingleEndpoint == "" {
		resourcePrefix := fmt.Sprintf("/%s", c.Config.LOD.Resource)
		if strings.HasPrefix(lodKey, resourcePrefix) {
			// todo for  now only support  RDF data
			lodKey = strings.Replace(lodKey, c.Config.LOD.Resource, c.Config.LOD.RDF, 1)
			http.Redirect(w, r, lodKey, 302)
			return
		}

		lodKey = strings.Replace(lodKey, c.Config.LOD.RDF, c.Config.LOD.Resource, 1)
		lodKey = strings.Replace(lodKey, c.Config.LOD.HTML, c.Config.LOD.Resource, 1)
	} else {
		// for now only support nt as format
		if !strings.HasSuffix(lodKey, ".nt") {
			lodKey = fmt.Sprintf("%s.nt", strings.TrimSuffix(lodKey, "/"))
			log.Printf("Redirecting to %s", lodKey)
			http.Redirect(w, r, lodKey, 302)
			return
		}
		lodKey = strings.TrimSuffix(lodKey, ".nt")

	}

	fr := fragments.NewFragmentRequest()
	fr.LodKey = lodKey
	frags, _, err := fr.Find(ctx, index.ESClient())
	if err != nil || len(frags) == 0 {
		w.WriteHeader(http.StatusNotFound)
		if err != nil {
			log.Printf("Unable to list fragments because of: %s", err)
			return
		}

		log.Printf("Unable to find fragments")
		return
	}

	var buffer bytes.Buffer
	for _, frag := range frags {
		buffer.WriteString(fmt.Sprintln(frag.Triple))
	}
	if strings.Contains(r.Header.Get("Accept"), "n-triples") {
		w.Header().Add("Content-Type", "application/n-triples")
	} else {
		w.Header().Add("Content-Type", "text/plain")
	}
	w.Write(buffer.Bytes())
	return

}

// listFragments returns a list of matching fragments
// See for more info: http://linkeddatafragments.org/
func listFragments(w http.ResponseWriter, r *http.Request) {
	fr := fragments.NewFragmentRequest()
	spec := chi.URLParam(r, "spec")
	if spec != "" {
		fr.Spec = spec
	}
	err := fr.ParseQueryString(r.URL.Query())
	if err != nil {
		log.Printf("Unable to list fragments because of: %s", err)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusBadRequest,
			Message:    fmt.Sprint("Unable to list fragments was not found"),
			Error:      err,
		})
		return
	}

	frags, totalFrags, err := fr.Find(ctx, index.ESClient())
	if err != nil || len(frags) == 0 {
		log.Printf("Unable to list fragments because of: %s", err)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusNotFound,
			Message:    fmt.Sprint("No fragments for query were found."),
			Error:      err,
		})
		return
	}
	switch fr.Echo {
	case "raw":
		render.JSON(w, r, frags)
		return
	case "es":
		src, err := fr.BuildQuery().Source()
		if err != nil {
			msg := "Unable to get the query source"
			log.Printf(msg)
			render.JSON(w, r, APIErrorMessage{
				HTTPStatus: http.StatusBadRequest,
				Message:    fmt.Sprint(msg),
				Error:      err,
			})
			return
		}
		render.JSON(w, r, src)
		return
	case "searchResponse":
		res, err := fr.Do(ctx, index.ESClient())
		if err != nil {
			msg := fmt.Sprintf("Unable to dump request: %s", err)
			log.Print(msg)
			render.JSON(w, r, APIErrorMessage{
				HTTPStatus: http.StatusBadRequest,
				Message:    fmt.Sprint(msg),
				Error:      err,
			})
			return
		}
		render.JSON(w, r, res)
		return
	case "request":
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			msg := fmt.Sprintf("Unable to dump request: %s", err)
			log.Print(msg)
			render.JSON(w, r, APIErrorMessage{
				HTTPStatus: http.StatusBadRequest,
				Message:    fmt.Sprint(msg),
				Error:      err,
			})
			return
		}

		render.PlainText(w, r, string(dump))
		return
	}

	var buffer bytes.Buffer
	for _, frag := range frags {
		buffer.WriteString(fmt.Sprintln(frag.Triple))
	}
	w.Header().Add("FRAG_COUNT", strconv.Itoa(int(totalFrags)))

	// Add hyperMediaControls
	hmd := fragments.NewHyperMediaDataSet(r, totalFrags, fr)
	controls, err := hmd.CreateControls()
	if err != nil {
		msg := fmt.Sprintf("Unable to create media controls: %s", err)
		log.Print(msg)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusBadRequest,
			Message:    fmt.Sprint(msg),
			Error:      err,
		})
		return
	}

	if strings.Contains(r.Header.Get("Accept"), "n-triples") {
		w.Header().Add("Content-Type", "application/n-triples")
	} else {
		w.Header().Add("Content-Type", "text/plain")
	}

	w.Write(controls)
	w.Write(buffer.Bytes())

	return
}

func generateFuzzed(w http.ResponseWriter, r *http.Request) {
	in, _, err := r.FormFile("file")
	if err != nil {
		render.PlainText(w, r, err.Error())
		return
	}
	spec := r.FormValue("spec")
	number := r.FormValue("number")
	baseURI := r.FormValue("baseURI")
	subjectType := r.FormValue("rootType")
	n, err := strconv.Atoi(number)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	recDef, err := fragments.NewRecDef(in)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fz, err := fragments.NewFuzzer(recDef)
	fz.BaseURL = baseURI
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	records, err := fz.CreateRecords(n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	typeLabel, err := c.Config.NameSpaceMap.GetSearchLabel(subjectType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	actions := []string{}
	for idx, rec := range records {
		hubID := fmt.Sprintf("%s_%s_%d", c.Config.OrgID, spec, idx)
		action := &hub3.BulkAction{
			HubID:         hubID,
			OrgID:         c.Config.OrgID,
			LocalID:       fmt.Sprintf("%d", idx),
			Spec:          spec,
			NamedGraphURI: fmt.Sprintf("%s/graph", fz.NewURI(typeLabel, idx)),
			Action:        "index",
			GraphMimeType: "application/ld+json",
			SubjectType:   subjectType,
			RecordType:    "mdr",
			Graph:         rec,
		}
		bytes, err := json.Marshal(action)
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			log.Printf("Unable to create Bulkactions: %s\n", err.Error())
			render.PlainText(w, r, err.Error())
			return
		}
		actions = append(actions, string(bytes))
	}
	render.PlainText(w, r, strings.Join(actions, "\n"))
	//w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	return
}

func listDataSetHistogram(w http.ResponseWriter, r *http.Request) {
	buckets, err := models.NewDataSetHistogram()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, buckets)
}

// listDataSets returns a list of all public datasets
func listDataSets(w http.ResponseWriter, r *http.Request) {
	sets, err := models.ListDataSets()
	if err != nil {
		log.Printf("Unable to list datasets because of: %s", err)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusInternalServerError,
			Message:    fmt.Sprint("Unable to list datasets was not found"),
			Error:      err,
		})
		return
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, sets)
	return
}

// getDataSetStats returns a dataset when found or a 404
func getDataSetStats(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	log.Printf("Get stats for spec %s", spec)
	stats, err := models.CreateDataSetStats(ctx, spec)
	if err != nil {
		if err == storm.ErrNotFound {
			log.Printf("Unable to retrieve a dataset: %s", err)
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, APIErrorMessage{
				HTTPStatus: http.StatusNotFound,
				Message:    fmt.Sprintf("%s was not found", chi.URLParam(r, "spec")),
				Error:      err,
			})
			return
		}
		status := http.StatusInternalServerError
		render.Status(r, status)
		log.Printf("Unable to create dataset stats: %#v", err)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: status,
			Message:    fmt.Sprintf("Can't create stats for %s", spec),
			Error:      err,
		})
		return
	}
	render.JSON(w, r, stats)
	return

}

// getDataSet returns a dataset when found or a 404
func getDataSet(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	ds, err := models.GetDataSet(spec)
	if err != nil {
		if err == storm.ErrNotFound {
			log.Printf("Unable to retrieve a dataset: %s", err)
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, APIErrorMessage{
				HTTPStatus: http.StatusNotFound,
				Message:    fmt.Sprintf("%s was not found", spec),
				Error:      err,
			})
			return
		}
		status := http.StatusInternalServerError
		render.Status(r, status)
		log.Printf("Unable to get dataset: %s", spec)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: status,
			Message:    fmt.Sprintf("Can't create stats for %s", spec),
			Error:      err,
		})
		return

	}
	render.JSON(w, r, ds)
	return
}

func deleteDataset(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	fmt.Printf("spec is %s", spec)
	ds, err := models.GetDataSet(spec)
	if err == storm.ErrNotFound {
		render.Status(r, http.StatusNotFound)
		log.Printf("Dataset is not found: %s", spec)
		return
	}
	ok, err := ds.DropAll(ctx, wp)
	if !ok || err != nil {
		render.Status(r, http.StatusBadRequest)
		log.Printf("Unable to delete request because: %s", err)
		return
	}
	log.Printf("Dataset is deleted: %s", spec)
	render.Status(r, http.StatusAccepted)
	return
}

// createDataSet creates a new dataset.
func createDataSet(w http.ResponseWriter, r *http.Request) {
	spec := r.FormValue("spec")
	if spec == "" {
		spec = chi.URLParam(r, "spec")
	}
	if spec == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusBadRequest,
			Message:    fmt.Sprintln("spec can't be empty."),
			Error:      nil,
		})
		return
	}
	fmt.Printf("spec is %s", spec)
	ds, err := models.GetDataSet(spec)
	if err == storm.ErrNotFound {
		var created bool
		ds, created, err = models.CreateDataSet(spec)
		if created {
			err = fragments.SaveDataSet(spec, bp)
		}
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, APIErrorMessage{
				HTTPStatus: http.StatusBadRequest,
				Message:    fmt.Sprintf("Unable to create dataset for %s", spec),
				Error:      nil,
			})
			log.Printf("Unable to create dataset for %s.\n", spec)
			return
		}
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, ds)
		return
	}
	render.Status(r, http.StatusNotModified)
	render.JSON(w, r, ds)
	return
}

// listNameSpaces list all currently defined NameSpace object
func listNameSpaces(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, c.Config.NameSpaceMap.ByPrefix())
	//render.JSON(w, r, c.Config.NameSpaces)
	return
}

func treeList(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	if spec == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusBadRequest,
			Message:    fmt.Sprintln("spec can't be empty."),
			Error:      nil,
		})
		return
	}
	nodeID := chi.URLParam(r, "nodeID")
	if nodeID != "" {
		id, err := url.QueryUnescape(nodeID)
		if err != nil {
			log.Println("Unable to unescape QueryParameters.")
			render.Status(r, http.StatusBadRequest)
			render.PlainText(w, r, err.Error())
			return
		}
		q := r.URL.Query()
		q.Add("byLeaf", id)
		r.URL.RawQuery = q.Encode()
	}
	searchRequest, err := fragments.NewSearchRequest(r.URL.Query())
	if err != nil {
		log.Println("Unable to create Search request")
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, err.Error())
		return
	}
	searchRequest.ItemFormat = fragments.ItemFormatType_TREE
	searchRequest.AddQueryFilter(fmt.Sprintf("%s:%s", c.Config.ElasticSearch.SpecKey, spec), false)
	switch searchRequest.Tree {
	case nil:
		searchRequest.Tree = &fragments.TreeQuery{
			Depth: []string{"1", "2"},
			Spec:  spec,
		}
	default:
		searchRequest.Tree.Spec = spec
	}
	processSearchRequest(w, r, searchRequest)
	return
}

func treeDownload(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	if spec == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusBadRequest,
			Message:    fmt.Sprintln("spec can't be empty."),
			Error:      nil,
		})
		return
	}
	eadPath := path.Join(c.Config.EAD.CacheDir, fmt.Sprintf("%s.xml", spec))
	http.ServeFile(w, r, eadPath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.xml", spec))
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	return
}

func treeDescriptionApi(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	eadPath := path.Join(c.Config.EAD.CacheDir, fmt.Sprintf("%s.xml", spec))
	cead, err := ead.ReadEAD(eadPath)
	if err != nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusNotFound,
			Message:    fmt.Sprintln("archive not found"),
			Error:      nil,
		})
		return
	}
	render.JSON(w, r, cead.Carchdesc.Cdescgrp)
	return
}

func treeDescription(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	ds, err := models.GetDataSet(spec)
	if err != nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusNotFound,
			Message:    fmt.Sprintln("archive not found"),
			Error:      nil,
		})
		return
	}

	desc := &fragments.TreeDescription{}
	desc.Name = ds.Label
	desc.Abstract = ds.Abstract
	desc.InventoryID = ds.Spec
	desc.Owner = ds.Owner
	desc.Period = ds.Period

	render.JSON(w, r, desc)

	return
}

func treeStats(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	if spec == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusBadRequest,
			Message:    fmt.Sprintln("spec can't be empty."),
			Error:      nil,
		})
		return
	}
	stats, err := fragments.CreateTreeStats(r.Context(), spec)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// todo return 404 if stats.Leafs == 0
	if stats.Leafs == 0 {
		render.Status(r, http.StatusNotFound)
		return
	}
	render.JSON(w, r, stats)
	return
}
