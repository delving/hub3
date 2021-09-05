package domain

import "strings"

type OrgConfigRetriever interface {
	RetrieveConfig(orgID string) (cfg OrganizationConfig, ok bool)
}

type OrganizationConfig struct {
	// domain is a list of all valid domains (including subdomains) for an domain.Organization
	// the domain ID will be injected in each request by the organization middleware.
	id             string
	Domains        []string `json:"domains,omitempty"`
	Default        bool     `json:"default"`
	CustomID       string   `json:"customID"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	RDFBaseURL     string   `json:"rdfBaseURL"`
	MintDatasetURL string   `json:"mintDatasetURL"`
	MintOrgIDURL   string   `json:"mintOrgIDURL"`
	// Sitemaps       []SitemapConfig `json:"sitemaps,omitempty"`
	Sitemaps []struct {
		ID      string `json:"id"`
		BaseURL string `json:"baseURL"`
		Filters string `json:"filters"` // qf and q URL params
		OrgID   string `json:"-"`
	} `json:"sitemaps"`
	ElasticSearch struct {
		// base of the index aliases
		IndexName string `json:"indexName,omitempty"`
		// if non-empty digital objects will be indexed in a dedicated v2 index
		DigitalObjectSuffix string `json:"digitalObjectSuffix,omitempty"`
		// IndexTypes options are v1, v2, fragment
		IndexTypes []string `json:"indexTypes,omitempty"`
		Defaults   struct {
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
