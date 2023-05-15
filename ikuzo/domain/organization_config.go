package domain

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
)

type OrgConfigRetriever interface {
	RetrieveConfig(orgID string) (cfg OrganizationConfig, ok bool)
	Configs(ctx context.Context) (cfgs []OrganizationConfig, err error)
}

type ArchesConfig struct {
	Enabled           bool   `json:"enabled"`
	URL               string `json:"url"`
	OAuthClientID     string `json:"oAuthClientID"`
	OAuthClientSecret string `json:"oAuthClientSecret"`
	UserName          string `json:"userName"`
	Password          string `json:"password"`
	DSN               string `json:"dsn"` // arches postgresql
}

type OAIPMHConfig struct {
	Enabled        bool     `json:"enabled"`
	AdminEmails    []string `json:"adminEmails"`
	RepositoryName string   `json:"repositoryName"`
	Deleted        string   `json:"deleted"` // no, persistent, transient
	// HarvestPath    string   `json:"harvestPath"`
}

type OrganizationConfig struct {
	// domain is a list of all valid domains (including subdomains) for an domain.Organization
	// the domain ID will be injected in each request by the organization middleware.
	id          string
	Domains     []string      `json:"domains,omitempty"`
	Default     bool          `json:"default"`
	CustomID    string        `json:"customID"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	IndexTypes  []string      `json:"indexTypes,omitempty"`
	Arches      *ArchesConfig `json:"arches"`
	// archivespace config
	ArchivesSpace struct {
		Enabled      bool   `json:"enabled"`
		URL          string `json:"url"`
		RepositoryID string `json:"repositoryID"`
	} `json:"archivesspace"`
	// Sitemaps       []SitemapConfig `json:"sitemaps,omitempty"`
	Sitemaps []struct {
		ID            string   `json:"id"`
		BaseURL       string   `json:"baseURL"`
		Query         string   `json:"query"`
		Filters       []string `json:"filters"` // qf and q URL params
		OrgID         string   `json:"-"`
		ExcludedSpecs []string `json:"excludedSpecs"`
	} `json:"sitemaps"`
	Config struct {
		Identifiers struct {
			ArkNAAN      string `json:"arkNAAN"`
			IsilCode     string `json:"isilCode"`
			BrocadeOrgID string `json:"brocadeOrgID"`
		} `json:"identifiers"`
		ShortName   string `json:"shortName"`
		SubTitle    string `json:"subTitle"`
		Description string `json:"description"`
		Location    string `json:"location"`
	} `json:"config"`
	ElasticSearch struct {
		// base of the index aliases
		IndexName string `json:"indexName,omitempty"`
		// if non-empty digital objects will be indexed in a dedicated v2 index
		DigitalObjectSuffix string `json:"digitalObjectSuffix,omitempty"`
		// IndexTypes options are v1, v2, fragment
		IndexTypes []string `json:"indexTypes,omitempty"`
		// Shards is the number of index shards created
		Shards int `json:"shards"`
		// Replicas is the number of replicas created
		Replicas int `json:"replicas"`
		// TrackTotalHits determins if elasticsearch counts results over 10.000
		TrackTotalHits bool `json:"trackTotalHits"`
		// Defaults are the search defaults to be used in the API
		Defaults struct {
			// maxTreeSize is the maximum size of the number of nodes in the tree navigation API
			MaxTreeSize int `json:"maxTreeSize"`
			// FacetSize is the default number of facet items returned from the v2 API
			FacetSize int `json:"facetSize"`
			// Limit is the default number of results returned from from the v2 API
			Limit int `json:"limit"`
			// MaxLimit is the maximum number of results returned from the v2 API
			MaxLimit int `json:"maxLimit"`
		} `json:"defaults"`
	} `json:"elasticSearch,omitempty"`
	OAIPMH struct {
		Enabled        bool     `json:"enabled"`
		AdminEmails    []string `json:"adminEmails"`
		RepositoryName string   `json:"repositoryName"`
		ResponseSize   int      `json:"responseSize"`
	} `json:"oaipmh,omitempty"`
	RDF struct {
		RDFBaseURL     string `json:"rdfBaseURL"`
		MintDatasetURL string `json:"mintDatasetURL"`
		MintOrgIDURL   string `json:"mintOrgIDURL"`
	} `json:"rdf,omitempty"`
	SPARQL struct {
		Enabled        bool   `json:"enabled"`    // Enable the SPARQL proxy
		Host           string `json:"host"`       // the base-url to the SPARQL endpoint including the scheme and the port
		QueryPath      string `json:"queryPath"`  // the relative path of the endpoint. This can should contain the database name that is injected when the sparql endpoint is build
		UpdatePath     string `json:"updatePath"` // the relative path of the update endpoint. This can should contain the database name that is injected when the sparql endpoint is build
		GraphStorePath string `json:"dataPath"`   // the relative GraphStore path of the endpoint. This can should contain the database name that is injected when the sparql endpoint is build
		UserName       string `json:"userName"`
		Password       string `json:"password"`
	} `json:"sparql,omitempty"`
	Namespaces []Namespace `json:"namespaces"`
}

func (cfg *OrganizationConfig) OrgID() string {
	if cfg.CustomID != "" {
		return cfg.CustomID
	}

	return cfg.id
}

func (cfg *OrganizationConfig) SetOrgID(id string) {
	cfg.id = id
}

func (cfg *OrganizationConfig) indexName() string {
	if cfg.ElasticSearch.IndexName != "" {
		return cfg.ElasticSearch.IndexName
	}

	return cfg.OrgID()
}

// GetIndexName returns the lowercased indexname for the v2 index
// This inforced correct behavior when creating an index in ElasticSearch.
func (cfg *OrganizationConfig) GetIndexName() string {
	return strings.ToLower(cfg.indexName()) + "v2"
}

// GetV1IndexName returns the v1 index name
func (cfg *OrganizationConfig) GetV1IndexName() string {
	return strings.ToLower(cfg.indexName()) + "v1"
}

// GetFragmentsIndexName returns the indexname for lod-fragments
func (cfg *OrganizationConfig) GetFragmentsIndexName() string {
	return strings.ToLower(cfg.indexName()) + "v2_frag"
}

func (cfg *OrganizationConfig) GetSuggestIndexName() string {
	return strings.ToLower(cfg.indexName()) + "v2_suggest"
}

// GetDigitalObjectIndexName returns the names for the digitalobject index.
// In some cases the indices for records and digitalobjects need to be split in
// separated indexes. The v2 indexname is returned when the digitalobject suffix
// is empty.
func (cfg *OrganizationConfig) GetDigitalObjectIndexName() string {
	if cfg.ElasticSearch.DigitalObjectSuffix == "" {
		return cfg.GetIndexName()
	}

	return strings.ToLower(cfg.indexName()) + "v2-" + strings.ToLower(cfg.ElasticSearch.DigitalObjectSuffix)
}

// GetSparqlEndpoint builds the SPARQL endpoint from the RDF Config object.
// When the dbName is empty the OrgId from the configuration is used.
func (cfg *OrganizationConfig) GetSparqlEndpoint(dbName string) string {
	if dbName == "" {
		dbName = cfg.OrgID()
	}

	u, err := url.Parse(cfg.SPARQL.Host)
	if err != nil {
		log.Fatal(err)
	}

	u.Path = fmt.Sprintf(cfg.SPARQL.QueryPath, dbName)

	return u.String()
}

// GetSparqlUpdateEndpoint builds the SPARQL Update endpoint from the RDF Config object.
// When the dbName is empty the OrgId from the configuration is used.
func (cfg *OrganizationConfig) GetSparqlUpdateEndpoint(dbName string) string {
	if dbName == "" {
		dbName = cfg.OrgID()
	}

	u, err := url.Parse(cfg.SPARQL.Host)
	if err != nil {
		log.Fatal(err)
	}

	u.Path = fmt.Sprintf(cfg.SPARQL.UpdatePath, dbName)

	return u.String()
}

// GetGraphStoreEndpoint builds the GraphStore endpoint from the RDF Config object.
// When the dbName is empty the OrgId from the configuration is used.
func (cfg *OrganizationConfig) GetGraphStoreEndpoint(dbName string) string {
	if dbName == "" {
		dbName = cfg.OrgID()
	}

	u, err := url.Parse(cfg.SPARQL.Host)
	if err != nil {
		log.Fatal(err)
	}

	u.Path = fmt.Sprintf(cfg.SPARQL.GraphStorePath, dbName)

	return u.String()
}
