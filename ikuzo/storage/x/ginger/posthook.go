package ginger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/delving/hub3/ikuzo/service/x/bulk"
	"github.com/parnurzeal/gorequest"
	"github.com/rs/zerolog/log"
)

// compile time check to see if full interface is implemented
var _ bulk.PostHookService = (*PostHook)(nil)

// PostHookJob  holds the info for building a crea
type PostHookJob struct {
	item   *bulk.PostHookItem
	jsonld []map[string]interface{}
	Graph  string
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

func (ph *PostHook) DropDataset(dataset string, revision int) (resp *http.Response, err error) {
	req, err := http.NewRequest("DELETE", ph.endpoint, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("api_key", ph.apiKey)
	q.Add("collection", dataset)

	if revision > 0 {
		q.Add("rev", fmt.Sprintf("%d", revision))
	}

	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")

	var netClient = &http.Client{
		Timeout: time.Second * 15,
	}

	return netClient.Do(req)
}

func (ph *PostHook) Valid(datasetID string) bool {
	if ph.endpoint == "" {
		return false
	}

	for _, e := range ph.excludedDataSets {
		if strings.EqualFold(e, datasetID) {
			return false
		}
	}

	return true
}

func (ph *PostHook) Publish(items ...*bulk.PostHookItem) error {
	jobs := []*PostHookJob{}

	for _, item := range items {
		if item.Deleted {
			resp, err := ph.DropDataset(item.DatasetID, item.Revision)
			if err != nil {
				log.Error().Err(err).Str("datasetID", item.DatasetID).Msg("unable to drop posthook dataset")
				return err
			}

			if resp.StatusCode > 299 {
				defer resp.Body.Close()
				body, readErr := ioutil.ReadAll(resp.Body)

				if readErr != nil {
					log.Error().Err(err).Str("datasetID", item.DatasetID).
						Msg("unable to read posthook body")
				}

				log.Error().Err(err).
					Str("body", string(body)).
					Int("revision", item.Revision).
					Int("status_code", resp.StatusCode).
					Str("datasetID", item.DatasetID).
					Msg("unable to drop posthook dataset")
			}

			continue
		}

		ph, err := NewPostHookJob(item)
		if err != nil {
			return err
		}

		jobs = append(jobs, ph)
	}

	if len(jobs) == 0 {
		return nil
	}

	request := gorequest.New()

	bulkGraphs := []interface{}{}
	for _, job := range jobs {
		// gauge.Queue(ph)
		bulkGraphs = append(bulkGraphs, job.jsonld)
	}

	graphsAsJSON, err := json.Marshal(bulkGraphs)
	if err != nil {
		return err
	}

	rsp, body, errs := request.Post(ph.endpoint).
		Set("Content-Type", "application/json-ld; charset=utf-8").
		Query(fmt.Sprintf("api_key=%s", ph.apiKey)).
		Type("text").
		Send(string(graphsAsJSON)).
		End()

	//fmt.Printf("jsonld: %s\n", json)
	// log.Printf("post-response: %#v -> %#v\n %#v", rsp, body, errs)
	if errs != nil || rsp.StatusCode != http.StatusOK {
		// log.Error().Str("apiKey", ph.apiKey).Msgf("post-response: %#v -> %#v\n %#v", rsp, body, errs)
		// log.Error().Msgf("Unable to store: %#v\n", errs)
		log.Error().Msgf("JSON-LD: %s\n", graphsAsJSON)
		// log.Error().Msgf("bulk: %s\n", bulk)
		// for _, job := range jobs {
		// err := gauge.Error(job)
		// if err != nil {
		// return err
		// }
		// }

		return fmt.Errorf("unable to save to endpoint %s;\n %s", ph.endpoint, body)
	}

	log.Info().Str("svc", "posthook").Int("bulkItems", len(bulkGraphs)).Msg("Stored posthook items for ginger")
	// for _, job := range jobs {
	// err := gauge.Done(job)
	// if err != nil {
	// return err
	// }
	// }

	return nil
}

// NewPostHookJob creates a new PostHookJob and populates the rdf2go Graph
func NewPostHookJob(item *bulk.PostHookItem) (*PostHookJob, error) {
	ph := &PostHookJob{
		item: item,
	}

	if !ph.item.Deleted {
		// setup the cleanup
		err := ph.parseJSONLD()
		if err != nil {
			return nil, err
		}

		ph.addNarthexDefaults(ph.item.HubID)
		ph.cleanPostHookGraph()
		// log.Info().Msgf("ph.jsonld %#v", ph.jsonld)

		err = ph.updateJSONLD()
		if err != nil {
			return nil, err
		}
	}

	return ph, nil
}

func (ph *PostHookJob) updateJSONLD() error {
	b, err := json.Marshal(ph.jsonld)
	if err != nil {
		return err
	}

	ph.Graph = string(b)

	return nil
}

func (ph *PostHookJob) parseJSONLD() error {
	jsonld, err := ph.item.Graph.GenerateJSONLD()
	if err != nil {
		return err
	}

	ph.jsonld = jsonld

	return nil
}

// cleanPostHookGraph applies post hook clean actions to the graph
func (ph *PostHookJob) cleanPostHookGraph() {
	cleanMap := []map[string]interface{}{}

	for _, rsc := range ph.jsonld {
		cleanEntry := make(map[string]interface{})

		ebuCore := "urn:ebu:metadata-schema:ebuCore_2014"

		for uri, v := range rsc {
			if strings.HasPrefix(uri, ebuCore) {
				uri = strings.TrimLeft(uri, ebuCore)
				uri = strings.TrimLeft(uri, "/")
				uri = fmt.Sprintf("http://www.ebu.ch/metadata/ontologies/ebucore/ebucore#%s", uri)
			}

			var dateURI string

			if _, ok := dateFields[uri]; ok {
				dateURI = cleanDateURI(uri)
			}

			if dateURI != "" {
				// todo add code to cleanup the date formatting
				// TODO also add the original
				cleanEntry[dateURI] = v
			} else {
				// insert the uri original URI and raw value
				cleanEntry[uri] = v
			}
		}

		cleanMap = append(cleanMap, cleanEntry)
	}

	ph.jsonld = cleanMap
}

func containsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

func (ph *PostHookJob) addNarthexDefaults(hubID string) {
	parts := strings.Split(hubID, "_")
	localID := parts[2]
	subject := ph.item.Subject + "/about"

	var (
		defaults map[string]interface{}
		found    bool
	)

	for _, resource := range ph.jsonld {
		ttype, ok := resource["@type"]
		if ok {
			// nolint:gocritic // type check must use switch
			switch ttype := ttype.(type) {
			case []string:
				// log.Debug().Msgf("ttype: %s", ttype)
				if containsString(ttype, "http://xmlns.com/foaf/0.1/Document") {
					defaults = resource
					found = true

					break
				}
			}
		}
	}

	if !found {
		defaults = make(map[string]interface{})
		defaults["@id"] = subject
		defaults["@type"] = []string{"http://xmlns.com/foaf/0.1/Document"}
	}

	checkUpdate(defaults, "localId", localID)
	checkUpdate(defaults, "hubID", hubID)
	checkUpdate(defaults, "spec", ph.item.DatasetID)
	checkUpdate(defaults, "belongsTo", createDatasetURI(ph.item.Subject))
	// checkUpdate(defaults, "revision", ph.item.Revision)
	checkUpdate(defaults, "revision", 10)
	checkUpdate(defaults, "http://creativecommons.org/ns#attributionName", ph.item.DatasetID)
	checkUpdate(defaults, "http://xmlns.com/foaf/0.1/primaryTopic", ph.item.Subject)

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
