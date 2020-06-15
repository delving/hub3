package ginger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/service/x/bulk"
	r "github.com/kiivihal/rdf2go"
	ld "github.com/linkeddata/gojsonld"
	"github.com/parnurzeal/gorequest"
)

// compile time check to see if full interface is implemented
// var _ bulk.PostHookService = (*PostHook)(nil)

// PostHookJob  holds the info for building a crea
type PostHookJob struct {
	item   *bulk.PostHookItem
	jsonld []map[string]interface{}
}

type PostHook struct {
	orgID            string
	endpoint         string
	excludedDataSets []string
	apiKey           string
	gauge            PostHookGauge
}

func NewPostHook(orgID, endpoint, apiKey string, excludedDataSets ...string) *PostHook {
	return &PostHook{
		orgID:            orgID,
		endpoint:         endpoint,
		excludedDataSets: excludedDataSets,
		apiKey:           apiKey,
		gauge: PostHookGauge{
			Created:  time.Now(),
			Counters: make(map[string]*PostHookCounter),
		},
	}
}

func (ph *PostHook) OrgID() string {
	return ph.orgID
}

func (ph *PostHook) Publish(items ...*bulk.PostHookItem) error {
	jobs := []*PostHookJob{}
	for _, item := range items {
		jobs = append(jobs, NewPostHookJob(item))
	}

	request := gorequest.New()

	bulk := []interface{}{}
	for _, job := range jobs {
		bulk = append(bulk, job.jsonld)
		gauge.Queue(ph)
	}
	json, err := json.Marshal(bulk)
	if err != nil {
		return err
	}

	rsp, body, errs := request.Post(ph.endpoint).
		Set("Content-Type", "application/json-ld; charset=utf-8").
		Query(fmt.Sprintf("api_key=%s", ph.apiKey)).
		Type("text").
		Send(string(json)).
		End()

	//fmt.Printf("jsonld: %s\n", json)
	// log.Printf("post-response: %#v -> %#v\n %#v", rsp, body, errs)
	if errs != nil || rsp.StatusCode != http.StatusOK {
		// log.Printf("post-response: %#v -> %#v\n %#v", rsp, body, errs)
		// log.Printf("Unable to store: %#v\n", errs)
		// log.Printf("JSON-LD: %s\n", json)
		for _, job := range jobs {
			err := gauge.Error(job)
			if err != nil {
				return err
			}
		}

		return fmt.Errorf("Unable to save to endpoint %s;\n %s", ph.endpoint, body)
	}
	// log.Printf("Stored %d bulk items \n", len(bulk))
	for _, job := range jobs {
		err := gauge.Done(job)
		if err != nil {
			return err
		}
	}

	return nil
}

// NewPostHookJob creates a new PostHookJob and populates the rdf2go Graph
func NewPostHookJob(item *bulk.PostHookItem) *PostHookJob {

	ph := &PostHookJob{
		item: item,
	}

	if !ph.item.Deleted {
		ph.cleanPostHookGraph()
		ph.addNarthexDefaults(item.HubID)
	}

	return ph
}

// -------------------- old code -------------------
func (ph *PostHookJob) parseJsonLD() error {
	var jsonld []map[string]interface{}

	err := json.Unmarshal([]byte(ph.item.Graph), &jsonld)
	if err != nil {
		return err
	}

	ph.jsonld = jsonld

	return nil
}

func (ph *PostHookJob) updateJsonLD() error {
	b, err := json.Marshal(ph.jsonld)
	if err != nil {
		return err
	}

	ph.Graph = string(b)

	return nil
}

func containsString(s []interface{}, e interface{}) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (ph *PostHookJob) addNarthexDefaults(hubID string) {
	//log.Printf("adding defaults for %s", ph.Subject)
	parts := strings.Split(hubID, "_")
	localID := parts[2]
	subject := ph.item.Subject + "/about"

	var defaults map[string]interface{}
	var found bool

	for _, resource := range ph.jsonld {
		ttype, ok := resource["@type"]
		if ok && containsString(ttype.([]interface{}), "http://xmlns.com/foaf/0.1/Document") {
			defaults = resource
			found = true
			break
		}
	}

	if defaults == nil {
		defaults = make(map[string]interface{})
		defaults["@id"] = subject
		defaults["@type"] = []string{"http://xmlns.com/foaf/0.1/Document"}
	}
	checkUpdate(defaults, "localId", localID)
	checkUpdate(defaults, "hubID", hubID)
	checkUpdate(defaults, "spec", ph.Spec)
	checkUpdate(defaults, "belongsTo", createDatasetURI(ph.Subject))
	checkUpdate(defaults, "revision", ph.Revision)
	checkUpdate(defaults, "http://creativecommons.org/ns#attributionName", ph.Spec)
	checkUpdate(defaults, "http://xmlns.com/foaf/0.1/primaryTopic", ph.Subject)

	if !found {
		ph.jsonld = append(ph.jsonld, defaults)
	}
}

func createDatasetURI(subject string) string {
	parts := strings.Split(subject, "/")
	base := strings.Join(
		parts[0:len(parts)-1],
		"/",
	)
	return strings.Replace(base, "/aggregation/", "/dataset/", 1)
}

func checkUpdate(defaults map[string]interface{}, uri string, value interface{}) {
	if !strings.HasPrefix(uri, "http") {
		uri = fmt.Sprintf("http://schemas.delving.eu/narthex/terms/%s", uri)
	}
	if _, ok := defaults[uri]; !ok {
		switch value.(type) {
		case string:
			defaults[uri] = addLiteralValue(value)
		case int:
			defaults[uri] = addLiteralInt(value)
		}
	}
}

// addLiteralValue returns an Array of RDF Literal values in the JSON-LD format
func addLiteralValue(v interface{}) []map[string]interface{} {
	vmap := make(map[string]interface{})
	vmap["@value"] = v
	return []map[string]interface{}{vmap}
}

func addLiteralInt(v interface{}) []map[string]interface{} {
	vmap := make(map[string]interface{})
	vmap["@value"] = v
	vmap["@type"] = "http://www.w3.org/2001/XMLSchema#integer"
	return []map[string]interface{}{vmap}
}

// Valid determines if the posthok is valid to apply.
func (ph PostHookJob) Valid() bool {
	return ProcessSpec(ph.Spec)
}

// ProcessSpec determines if a PostHookJob should be applied for a specific spec
func ProcessSpec(spec string) bool {
	if c.Config.PostHook.URL == "" {
		return false
	}
	for _, e := range c.Config.PostHook.ExcludeSpec {
		if e == spec {
			return false
		}
	}
	return true
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

func cleanDates(g *fragments.SortedGraph, t *r.Triple) bool {
	for _, date := range dateFields {
		if t.Predicate.RawValue() == date.RawValue() {
			newTriple := r.NewTriple(
				t.Subject,
				r.NewResource(fmt.Sprintf("%sRaw", t.Predicate.RawValue())),
				t.Object,
			)
			g.Add(newTriple)

			return true
		}
	}
	return false
}

func cleanEbuCore(g *fragments.SortedGraph, t *r.Triple) bool {
	uri := t.Predicate.RawValue()
	if strings.HasPrefix(uri, "urn:ebu:metadata-schema:ebuCore_2014") {
		uri := strings.TrimLeft(uri, "urn:ebu:metadata-schema:ebuCore_2014")
		uri = strings.TrimLeft(uri, "/")
		uri = fmt.Sprintf("http://www.ebu.ch/metadata/ontologies/ebucore/ebucore#%s", uri)
		g.AddTriple(
			t.Subject,
			r.NewResource(uri),
			t.Object,
		)
		return true
	}
	return false
}

func cleanDateURI(uri string) string {
	return fmt.Sprintf("%sRaw", uri)
}

// cleanPostHookGraph applies post hook clean actions to the graph
func (ph *PostHookJob) cleanPostHookGraph() {
	cleanMap := []map[string]interface{}{}
	for _, rsc := range ph.jsonld {
		cleanEntry := make(map[string]interface{})
		for uri, v := range rsc {
			if strings.HasPrefix(uri, "urn:ebu:metadata-schema:ebuCore_2014") {
				uri = strings.TrimLeft(uri, "urn:ebu:metadata-schema:ebuCore_2014")
				uri = strings.TrimLeft(uri, "/")
				uri = fmt.Sprintf("http://www.ebu.ch/metadata/ontologies/ebucore/ebucore#%s", uri)
			}

			var dateUri string
			switch uri {
			case ns.dcterms.Get("created").RawValue():
				dateUri = cleanDateURI(uri)
			case ns.dcterms.Get("issued").RawValue():
				dateUri = cleanDateURI(uri)
			case ns.nave.Get("creatorBirthYear").RawValue():
				dateUri = cleanDateURI(uri)
			case ns.nave.Get("creatorDeathYear").RawValue():
				dateUri = cleanDateURI(uri)
			case ns.nave.Get("date").RawValue():
				dateUri = cleanDateURI(uri)
			case ns.dc.Get("date").RawValue():
				dateUri = cleanDateURI(uri)
			case ns.nave.Get("dateOfBurial").RawValue():
				dateUri = cleanDateURI(uri)
			case ns.nave.Get("dateOfDeath").RawValue():
				dateUri = cleanDateURI(uri)
			case ns.nave.Get("productionEnd").RawValue():
				dateUri = cleanDateURI(uri)
			case ns.nave.Get("productionStart").RawValue():
				dateUri = cleanDateURI(uri)
			case ns.nave.Get("productionPeriod").RawValue():
				dateUri = cleanDateURI(uri)
			case ns.rdagr2.Get("dateOfBirth").RawValue():
				dateUri = cleanDateURI(uri)
			case ns.rdagr2.Get("dateOfDeath").RawValue():
				dateUri = cleanDateURI(uri)

			}
			if dateUri != "" {
				// todo add code to cleanup the date formatting
				// TODO also add the original
				cleanEntry[dateUri] = v
			} else {
				// insert the uri original URI and raw value
				cleanEntry[uri] = v

			}
		}
		cleanMap = append(cleanMap, cleanEntry)
	}
	ph.jsonld = cleanMap
}

// Bytes returns the PostHookJob as an JSON-LD bytes.Buffer
func (ph PostHookJob) Bytes() (bytes.Buffer, error) {
	var b bytes.Buffer
	b.WriteString(ph.Graph)
	return b, nil
}

// Bytes returns the PostHookJob as an JSON-LD string
func (ph PostHookJob) String() (string, error) {

	return ph.Graph, nil
}
