package sparql

import (
	"bytes"
	"context"
	fmt "fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/knakk/sparql"
	// "github.com/knakk/rdf"
)

type RepoConfig struct {
	Host           string `json:"host"`       // the base-url to the SPARQL endpoint including the scheme and the port
	QueryPath      string `json:"queryPath"`  // the relative path of the endpoint. This can should contain the database name that is injected when the sparql endpoint is build
	UpdatePath     string `json:"updatePath"` // the relative path of the update endpoint. This can should contain the database name that is injected when the sparql endpoint is build
	GraphStorePath string `json:"dataPath"`   // the relative GraphStore path of the endpoint. This can should contain the database name that is injected when the sparql endpoint is build
	Transport      struct {
		Retry    int
		Timeout  int
		UserName string `json:"userName"`
		Password string `json:"password"`
	}
	Bank *sparql.Bank
}

// Repo represent a RDF repository, assumed to be
// queryable via the SPARQL protocol over HTTP.
type Repo struct {
	cfg                RepoConfig
	client             *retryablehttp.Client
	queryEndpoint      string
	updateEndpoint     string
	graphStoreEndpoint string
	updateTmpl         *template.Template
}

// NewRepo creates a new representation of a RDF repository that can be
// queried through SPARQL.
func NewRepo(cfg RepoConfig) (*Repo, error) {
	tmpl, err := template.New("bulkUpdate").Parse(sparqlUpdateTemplate)
	if err != nil {
		return nil, err
	}

	r := Repo{
		cfg:        cfg,
		updateTmpl: tmpl,
	}

	if cfg.Transport.Timeout == 0 {
		cfg.Transport.Timeout = 10
	}

	r.setClient()
	if err := r.setEndpoints(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *Repo) setClient() {
	c := retryablehttp.NewClient()
	c.RetryMax = r.cfg.Transport.Retry
	c.HTTPClient.Timeout = time.Duration(r.cfg.Transport.Timeout) * time.Second
	r.client = c
}

func (r *Repo) setEndpoints() error {
	u, err := url.Parse(r.cfg.Host)
	if err != nil {
		return fmt.Errorf("invalid sparql host: %w", err)
	}

	if r.cfg.QueryPath == "" {
		return fmt.Errorf("RepoConfig.QueryPath cannot be empty")
	}
	u.Path = r.cfg.QueryPath
	r.queryEndpoint = u.String()

	if r.cfg.UpdatePath != "" {
		u.Path = r.cfg.UpdatePath
		r.updateEndpoint = u.String()
	}

	if r.cfg.GraphStorePath != "" {
		u.Path = r.cfg.GraphStorePath
		r.graphStoreEndpoint = u.String()
	}

	return nil
}

func (r *Repo) hasBasicAuth() bool {
	return r.cfg.Transport.UserName != "" && r.cfg.Transport.Password != ""
}

// Query performs a SPARQL Get HTTP request to the Repo, and returns the
// parsed application/sparql-results+json response.
func (r *Repo) Query(q string) (*Results, error) {
	return r.query(q, http.MethodGet)
}

// Query performs a SPARQL POST HTTP request to the Repo, and returns the
// parsed application/sparql-results+json response.
func (r *Repo) QueryPost(q string) (*Results, error) {
	return r.query(q, http.MethodPost)
}

func (r *Repo) queryRaw(q string, method, accept string) (*http.Response, error) {
	form := url.Values{}
	form.Set("query", q)
	b := form.Encode()

	req, err := retryablehttp.NewRequest(
		method,
		r.queryEndpoint,
		bytes.NewBufferString(b))
	if err != nil {
		return nil, err
	}

	if accept == "" {
		accept = "application/sparql-results+json"
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(b)))
	req.Header.Set("Accept", accept)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		var msg string
		if err != nil {
			msg = "Failed to read response body"
		} else {
			if strings.TrimSpace(string(b)) != "" {
				msg = "Response body: \n" + string(b)
			}
		}
		resp.Body.Close()

		return resp, fmt.Errorf("Query: SPARQL request failed (code: %s): \n%s ", resp.Status, msg)
	}

	return resp, nil
}

func (r *Repo) query(q string, method string) (*Results, error) {
	resp, err := r.queryRaw(q, method, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	results, err := parseJSON(resp.Body)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// Construct performs a SPARQL HTTP request to the Repo, and returns the
// result as a rdf.Graph.
func (r *Repo) Construct(q string) (*rdf.Graph, error) {
	resp, err := r.queryRaw(q, http.MethodPost, "text/turtle")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	g, err := ntriples.Parse(resp.Body, nil)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (r *Repo) Resolve(ctx context.Context, subj rdf.Subject) (*rdf.Graph, error) {
	query := fmt.Sprintf("describe %s", subj.String())
	g, err := r.Construct(query)
	if err != nil {
		return nil, err
	}

	return g, nil
}

// Update performs a SPARQL HTTP update request
func (r *Repo) Update(q string) error {
	if r.updateEndpoint == "" {
		return fmt.Errorf("update not supported by this Repo")
	}

	form := url.Values{}
	form.Set("update", q)
	b := form.Encode()

	req, err := retryablehttp.NewRequest(
		http.MethodPost,
		r.updateEndpoint,
		bytes.NewBufferString(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(b)))

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		b, err := ioutil.ReadAll(resp.Body)
		var msg string
		if err != nil {
			msg = "Failed to read response body"
		} else {
			if strings.TrimSpace(string(b)) != "" {
				msg = "Response body: \n" + string(b)
			}
		}
		return fmt.Errorf("Update: SPARQL request failed: %s. "+msg, resp.Status)
	}
	return nil
}
