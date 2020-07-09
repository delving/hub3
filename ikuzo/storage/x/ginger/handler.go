package ginger

import (
	"sync"

	"github.com/go-chi/chi"
)

type Hit struct {
	Source `json:"_source"`
}

type Source struct {
	Revision int `json:"revision"`
	System   `json:"system"`
}

type System struct {
	HubID       string `json:"slug"`
	Spec        string `json:"spec"`
	Subject     string `json:"source_uri"`
	SourceGraph string `json:"source_graph"`
}

type AuthKey struct {
	Key string
	sync.Mutex
}

// PostHookResource is a struct for the Search routes
type postHookResource struct{}

// Routes returns the chi.Router
func (rs postHookResource) Routes() chi.Router {
	r := chi.NewRouter()

	// r.Get("/datasets", listDatasets)
	// r.Delete("/datasets/{spec}", deleteDataset)
	// r.Get("/input/{id}", showInput)
	// r.Get("/output/{id}", showOutput)
	// r.Get("/counters", showCounter)

	return r
}

// func NewESPostHook(ctx context.Context, hubID string) (*PostHookJob, error) {
// source, err := getSource(ctx, hubID)
// if err != nil {
// return nil, err
// }

// ph := &PostHookJob{
// Graph:    source.SourceGraph,
// Spec:     source.Spec,
// Deleted:  false,
// Subject:  source.Subject,
// Revision: 1,
// }
// err = ph.parseJsonLD()
// if err != nil {
// return nil, err
// }

// ph.addNarthexDefaults(hubID)
// ph.cleanPostHookGraph()
// //log.Printf("%#v", ph.jsonld)

// err = ph.updateJsonLD()
// if err != nil {
// return nil, err
// }

// //g := r.NewGraph("")
// //err = g.Parse(strings.NewReader(source.SourceGraph), "application/ld+json")
// //if err != nil {
// //return nil, err
// //}

// //for t := range g.IterTriples() {
// //if !cleanDates(ph.Graph, t) && !cleanEbuCore(ph.Graph, t) {
// //ph.Graph.Add(t)
// //}
// //}

// return ph, nil
// }

func GetRoutes() chi.Router {
	return postHookResource{}.Routes()
}

var authKey AuthKey

// func getSource(ctx context.Context, hubID string) (*System, error) {

// record, err := index.ESClient().Get().
// // TODO(kiivihal): use injected name
// Index(c.Config.ElasticSearch.IndexName).
// Type("void_edmrecord").
// Id(hubID).
// Do(ctx)
// if err != nil {
// return nil, err
// }

// var sourceGraph Source
// err = json.Unmarshal(*record.Source, &sourceGraph)
// if err != nil {
// return nil, err
// }
// //log.Printf("%#v", string(sourceGraph.SourceGraph))
// return &sourceGraph.System, nil
// }

// func showInput(w http.ResponseWriter, r *http.Request) {
// hudID := chi.URLParam(r, "id")
// source, err := getSource(r.Context(), hudID)
// if err != nil {
// if err.(*elastic.Error).Status == 404 {
// http.Error(w, "Not Found", http.StatusNotFound)
// return
// }
// log.Printf("%#v", err)
// http.Error(w, err.Error(), http.StatusInternalServerError)
// return
// }
// w.Header().Set("Content-Type", "application/ld+json")
// w.Write([]byte(source.SourceGraph))
// return
// }

// func showCounter(w http.ResponseWriter, r *http.Request) {
// filterActive := r.URL.Query().Get("active") == "true"
// filterError := r.URL.Query().Get("error") == "true"
// gauge.ActiveDatasets = len(gauge.Counters)
// if filterActive || filterError {
// filteredGauge := PostHookGauge{
// Created:   gauge.Created,
// QueueSize: gauge.QueueSize,
// Counters:  make(map[string]*PostHookCounter),
// }

// for k, v := range gauge.Counters {
// switch {
// case v.IsActive && filterActive:
// filteredGauge.Counters[k] = v
// continue
// case v.InError != 0 && filterError:
// filteredGauge.Counters[k] = v
// continue
// }
// }

// render.JSON(w, r, filteredGauge)
// return
// }

// render.JSON(w, r, gauge)
// return
// }

// func showOutput(w http.ResponseWriter, r *http.Request) {
// hubID := chi.URLParam(r, "id")
// ph, err := NewESPostHook(r.Context(), hubID)
// if err != nil {
// if err.(*elastic.Error).Status == 404 {
// http.Error(w, "Not Found", http.StatusNotFound)
// return
// }
// log.Printf("%#v", err)
// http.Error(w, err.Error(), http.StatusInternalServerError)
// return
// }
// log.Printf("%#v", r.URL.Query())
// if r.URL.Query().Get("store") == "true" {
// log.Println("storing the posthook")
// Submit(ph)
// }
// w.Header().Set("Content-Type", "application/ld+json")
// fmt.Fprint(w, ph.Graph)
// return
// }

// func deleteDataset(w http.ResponseWriter, r *http.Request) {
// dataset := chi.URLParam(r, "spec")
// resp, err := DropPosthookDataset(dataset, "")
// if err != nil {
// http.Error(w, err.Error(), http.StatusInternalServerError)
// return
// }
// defer resp.Body.Close()
// w.Header().Set("Content-Type", "text/plain")
// _, err = io.Copy(w, resp.Body)
// if err != nil {
// http.Error(w, err.Error(), http.StatusInternalServerError)
// return
// }
// return
// }

// ListDatasets returns a list of indexed datasets from the PostHook endpoint.
// It renews the authorisation key when this not valid.
// func listDatasets(w http.ResponseWriter, r *http.Request) {
// //key, err := getAuthKey()
// //if err != nil {
// //http.Error(w, err.Error(), http.StatusInternalServerError)
// //return
// //}
// url := fmt.Sprintf("%s/api/erfgoedbrabant/brabantcloud", strings.TrimSuffix(c.Config.PostHook.URL, "/"))
// req, err := http.NewRequest("GET", url, nil)
// if err != nil {
// http.Error(w, err.Error(), http.StatusInternalServerError)
// return
// }
// //req.Header.Set("Cookie", key.Key)
// q := req.URL.Query()
// q.Add("api_key", c.Config.PostHook.APIKey)
// req.URL.RawQuery = q.Encode()

// req.Header.Set("Content-Type", "application/json")

// var netClient = &http.Client{
// Timeout: time.Second * 5,
// }
// resp, err := netClient.Do(req)
// if err != nil {
// http.Error(w, err.Error(), http.StatusInternalServerError)
// return
// }
// defer resp.Body.Close()
// w.Header().Set("Content-Type", "application/json")
// _, err = io.Copy(w, resp.Body)
// if err != nil {
// http.Error(w, err.Error(), http.StatusInternalServerError)
// return
// }

// //render.PlainText(w, r, key.Key)
// return
// }

// func getAuthKey() (AuthKey, error) {
// if authKey.Key != "" {
// return authKey, nil
// }
// return renewAuthKey()
// }

// func renewAuthKey() (AuthKey, error) {
// authEndpoint := fmt.Sprintf("%s/data/auth/login", strings.TrimSuffix(c.Config.PostHook.URL, "/"))
// payload := strings.NewReader(fmt.Sprintf(`{"username": "%s", "password": "%s"}`, c.Config.PostHook.UserName, c.Config.PostHook.Password))
// req, err := http.NewRequest("POST", authEndpoint, payload)
// if err != nil {
// return authKey, errors.Wrapf(err, "unable to build authentication request for posthook")
// }
// req.Header.Set("Content-Type", "application/json")

// var netClient = &http.Client{
// Timeout: time.Second * 5,
// }
// resp, err := netClient.Do(req)
// if err != nil {
// log.Printf("Error in posthook auth request: %s", err)
// return authKey, err
// }

// zSid := resp.Header.Get("Set-Cookie")
// if zSid == "" || resp.StatusCode != 200 {
// log.Printf("err = %#v", resp)
// return authKey, errors.New("unable to get auth key from response")
// }
// authKey.Lock()
// authKey.Key = zSid
// authKey.Unlock()

// return authKey, nil
// }
