package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	c "github.com/delving/rapid-saas/config"
	"github.com/delving/rapid-saas/hub3/fragments"
	"github.com/gammazero/workerpool"
	r "github.com/kiivihal/rdf2go"
	ld "github.com/linkeddata/gojsonld"
	"github.com/parnurzeal/gorequest"
)

// PostHookJob  holds the info for building a crea
type PostHookJob struct {
	Graph   *fragments.SortedGraph
	Spec    string
	Deleted bool
	Subject string
}

// PostHookJobFactory can be used to fire off PostHookJob jobs
type PostHookJobFactory struct {
	Spec string
	wp   *workerpool.WorkerPool
}

// NewPostHookJob creates a new PostHookJob and populates the rdf2go Graph
func NewPostHookJob(sg *fragments.SortedGraph, spec string, delete bool, subject, hubID string) *PostHookJob {
	ph := &PostHookJob{sg, spec, delete, subject}
	if !delete {
		ph.cleanPostHookGraph()
		ph.addNarthexDefaults(hubID)
	}
	return ph
}

func (ph *PostHookJob) addNarthexDefaults(hubID string) {
	//log.Printf("adding defaults for %s", ph.Subject)
	parts := strings.Split(hubID, "_")
	localID := parts[2]
	s := r.NewResource(ph.Subject + "/about")
	ph.Graph.AddTriple(
		s,
		r.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#"),
		r.NewResource("http://xmlns.com/foaf/0.1/Document"),
	)
	ph.Graph.AddTriple(
		s,
		r.NewResource("http://schemas.delving.eu/narthex/terms/hubId"),
		r.NewLiteral(hubID),
	)
	ph.Graph.AddTriple(
		s,
		r.NewResource("http://schemas.delving.eu/narthex/terms/spec"),
		r.NewLiteral(ph.Spec),
	)
	ph.Graph.AddTriple(
		s,
		r.NewResource("http://schemas.delving.eu/narthex/terms/localId"),
		r.NewLiteral(localID),
	)
}

// Valid determines if the posthok is valid to apply.
func (ph PostHookJob) Valid() bool {
	return ProcessSpec(ph.Spec)
}

// ProcessSpec determines if a PostHookJob should be applied for a specific spec
func ProcessSpec(spec string) bool {
	for _, e := range c.Config.PostHook.ExcludeSpec {
		if e == spec {
			return false
		}
	}
	return true
}

// ApplyPostHookJob applies the PostHookJob to all the configured URLs
func ApplyPostHookJob(ph *PostHookJob) {
	//time.Sleep(100 * time.Millisecond)
	for _, u := range c.Config.PostHook.URLs {
		err := ph.Post(u)
		if err != nil {
			log.Println(err)
			//} else {
			//log.Printf("stored: %s", ph.Subject)		log.Printf("Unable to send %s to %s", ph.Subject, u)
		}
	}
}

// Post sends json-ld to the specified endpointt
func (ph PostHookJob) Post(url string) error {
	request := gorequest.New()
	if ph.Deleted {
		log.Printf("Deleting via posthook: %s", ph.Subject)
		deleteURL := fmt.Sprintf("%s/delete", url)
		req := request.Delete(deleteURL).
			Query(fmt.Sprintf("id=%s", ph.Subject)).
			Retry(3, 5*time.Second, http.StatusBadRequest, http.StatusInternalServerError, http.StatusRequestTimeout)
		//log.Printf("%v", req)
		rsp, body, errs := req.End()
		if errs != nil || rsp.StatusCode != http.StatusNoContent {
			log.Printf("post-response: %#v -> %#v\n %#v", rsp, body, errs)
			log.Printf("Unable to delete: %#v", errs)
			return fmt.Errorf("Unable to save %s to endpoint %s", ph.Subject, url)
		}
		//log.Printf("Deleted %s\n", ph.Subject)
		return nil
	}
	json, err := ph.String()

	if err != nil {
		return err
	}

	rsp, body, errs := request.Post(url).
		Set("Content-Type", "application/json-ld; charset=utf-8").
		Type("text").
		Send(json).
		Retry(3, 5*time.Second, http.StatusBadRequest, http.StatusInternalServerError, http.StatusRequestTimeout).
		End()
	//fmt.Printf("jsonld: %s\n", json)
	if errs != nil || rsp.StatusCode != http.StatusOK {
		log.Printf("post-response: %#v -> %#v\n %#v", rsp, body, errs)
		log.Printf("Unable to store: %#v\n", errs)
		log.Printf("JSON-LD: %s\n", json)
		return fmt.Errorf("Unable to save %s to endpoint %s", ph.Subject, url)
	}
	//log.Printf("Stored %s\n", ph.Subject)
	return nil
}

var (
	ns = struct {
		rdf, rdfs, acl, cert, foaf, stat, dc, dcterms, nave, rdagr2, edm ld.NS
	}{
		rdf:     ld.NewNS("http://www.w3.org/1999/02/22-rdf-syntax-ns#"),
		rdfs:    ld.NewNS("http://www.w3.org/2000/01/rdf-schema#"),
		acl:     ld.NewNS("http://www.w3.org/ns/auth/acl#"),
		cert:    ld.NewNS("http://www.w3.org/ns/auth/cert#"),
		foaf:    ld.NewNS("http://xmlns.com/foaf/0.1/"),
		stat:    ld.NewNS("http://www.w3.org/ns/posix/stat#"),
		dc:      ld.NewNS("http://purl.org/dc/elements/1.1/"),
		dcterms: ld.NewNS("http://purl.org/dc/terms/"),
		nave:    ld.NewNS("http://schemas.delving.eu/nave/terms/"),
		rdagr2:  ld.NewNS("http://rdvocab.info/ElementsGr2/"),
		edm:     ld.NewNS("http://www.europeana.eu/schemas/edm/"),
	}
)

var dateFields = []ld.Term{
	ns.dcterms.Get("created"),
	ns.dcterms.Get("issued"),
	ns.nave.Get("creatorBirthYear"),
	ns.nave.Get("creatorDeathYear"),
	ns.nave.Get("date"),
	ns.dc.Get("date"),
	ns.nave.Get("dateOfBurial"),
	ns.nave.Get("dateOfDeath"),
	ns.nave.Get("productionEnd"),
	ns.nave.Get("productionStart"),
	ns.nave.Get("productionPeriod"),
	ns.rdagr2.Get("dateOfBirth"),
	ns.rdagr2.Get("dateOfDeath"),
}

func cleanDates(sg *fragments.SortedGraph, t *r.Triple) bool {
	for _, date := range dateFields {
		if t.Predicate.RawValue() == date.RawValue() {
			newTriple := r.NewTriple(
				t.Subject,
				r.NewResource(fmt.Sprintf("%sRaw", t.Predicate.RawValue())),
				t.Object,
			)
			sg.Add(newTriple)
			return true
		}
	}
	return false
}

func cleanEbuCore(sg *fragments.SortedGraph, t *r.Triple) bool {
	uri := t.Predicate.RawValue()
	if strings.HasPrefix(uri, "urn:ebu:metadata-schema:ebuCore_2014") {
		uri := strings.TrimLeft(uri, "urn:ebu:metadata-schema:ebuCore_2014")
		uri = strings.TrimLeft(uri, "/")
		uri = fmt.Sprintf("http://www.ebu.ch/metadata/ontologies/ebucore/ebucore#%s", uri)
		sg.AddTriple(
			t.Subject,
			r.NewResource(uri),
			t.Object,
		)
		return true
	}
	return false
}

// ResourceSortOrder holds information to sort RDF:type webresources based on
// their nave:resourceSortOrder key
type ResourceSortOrder struct {
	Resource map[string]interface{}
	SortKey  int
}

func sortMapArray(m []map[string]interface{}) []map[string]interface{} {

	var ss []ResourceSortOrder
	for _, wr := range m {
		sortKey, ok := wr["http://schemas.delving.eu/nave/terms/resourceSortOrder"]
		var sortOrder int
		if ok {
			sortKeyValue := sortKey.([]*r.LdObject)[0]
			sortInt, err := strconv.Atoi(sortKeyValue.Value)
			if err == nil {
				sortOrder = sortInt
			}
		}
		ss = append(ss, ResourceSortOrder{wr, sortOrder})
	}

	// sort by key
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].SortKey < ss[j].SortKey
	})

	var entries []map[string]interface{}
	for _, entry := range ss {
		entries = append(entries, entry.Resource)
	}
	return entries
}

// sortWebResources sorts the webresources in order last
func (ph *PostHookJob) sortWebResources() (bytes.Buffer, error) {
	var b bytes.Buffer

	entries := []map[string]interface{}{}
	wr := []map[string]interface{}{}

	jsonld, err := ph.Graph.GenerateJSONLD()
	if err != nil {
		return b, err
	}

	for _, resource := range jsonld {
		rdfTypes, ok := resource["@type"]
		if !ok {
			return b, fmt.Errorf("JSONLD entry does not contain @type definition")
		}
		for _, t := range rdfTypes.([]string) {
			switch t {
			case "http://www.europeana.eu/schemas/edm/WebResource":
				wr = append(wr, resource)
			default:
				entries = append(entries, resource)
			}
		}
	}

	entries = append(entries, sortMapArray(wr)...)

	// write bytes
	bytes, err := json.Marshal(entries)
	if err != nil {
		return b, err
	}
	fmt.Fprint(&b, string(bytes))

	return b, nil
}

// cleanPostHookGraph applies post hook clean actions to the graph
func (ph *PostHookJob) cleanPostHookGraph() {
	newGraph := &fragments.SortedGraph{}
	for _, t := range ph.Graph.Triples() {
		if !cleanDates(newGraph, t) && !cleanEbuCore(newGraph, t) {
			newGraph.Add(t)
		}
	}
	ph.Graph = newGraph
}

// Bytes returns the PostHookJob as an JSON-LD bytes.Buffer
func (ph PostHookJob) Bytes() (bytes.Buffer, error) {
	var b bytes.Buffer
	err := ph.Graph.SerializeFlatJSONLD(&b)
	return b, err
}

// Bytes returns the PostHookJob as an JSON-LD string
func (ph PostHookJob) String() (string, error) {
	var b bytes.Buffer

	entries, err := ph.Graph.GenerateJSONLD()
	if err != nil {
		return "", err
	}

	// write bytes
	bytes, err := json.Marshal(entries)
	if err != nil {
		return "", err
	}
	fmt.Fprint(&b, string(bytes))
	return b.String(), nil
}
