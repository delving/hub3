// Copyright 2017 Delving B.V.
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

package config

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/delving/hub3/ikuzo/logger"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var (
	// Config is the general configuration object
	Config RawConfig

	// CfgFile is the path to the config file
	CfgFile string
)

// RawConfig holds all the configuration blocks.
// These are bound from cli, Environment variables or configuration files by
// Viper.
type RawConfig struct {
	HTTP          `json:"http"`
	ElasticSearch `json:"elasticsearch"`
	Logging       `json:"logging"`
	RDF           `json:"rdf"`
	OAIPMH        `json:"oaipmh"`
	WebResource   `json:"webresource"`
	ImageProxy    `json:"imageproxy"`
	LOD           `json:"lod"`
	NameSpaces    []NameSpace   `json:"namespaces"`
	NameSpaceMap  *NameSpaceMap `json:"nameSpaceMap"`
	RDFTag        `json:"rdftag"`
	PostHook      []PostHook `json:"postHook"`
	Cache         `json:"cache"`
	RDFTagMap     *RDFTagMap `json:"rdfTagMap"`
	SiteMap       `json:"siteMap"`
	EAD           `json:"ead"`
	DataSetTag    map[string]DataSets `json:"dataSetTag"`
	DatasetTagMap *RDFTagMap
	Logger        *zerolog.Logger `json:"logger"`
}

type DataSets struct {
	Specs []string
}

// PostHook contains the configuration for the JSON-LD posthook configuration
type PostHook struct {
	Name        string   `json:"name"`
	ExcludeSpec []string `json:"excludeSpec"`
	URL         string   `json:"url"`
	OrgID       string   `json:"orgID"`
	APIKey      string   `json:"apiKey"`
	UserName    string   `json:"userName"`
	Password    string   `json:"password"`
	CustomWait  int      `json:"customWait"`
}

// ElasticSearch holds all the configuration values
// It is bound by Viper.
type ElasticSearch struct {
	Urls                []string `json:"urls"`
	Enabled             bool     `json:"enabled"`
	IndexName           string   `json:"indexName"`
	DigitalObjectSuffix string   `json:"digitalObjectSuffix"`
	Proxy               bool     `json:"proxy"`
	Fragments           bool     `json:"fragments"`
	IndexV1             bool     `json:"indexV1"` // exclusive with v2 indexing
	EnableTrace         bool     `json:"enableTrace"`
	EnableInfo          bool     `json:"enableInfo"`
	SpecKey             string   `json:"specKey"`
	RevisionKey         string   `json:"revisionKey"`
	OrgIDKey            string   `json:"orgIDKey"`
	UserName            string   `json:"userName"`
	Password            string   `json:"password"`
	FacetSize           int      `json:"facetSize"`
	MinimumShouldMatch  string   `json:"minimumShouldMatch"`
	Workers             int      `json:"workers"`
	Shards              int      `json:"shards"`
	Replicas            int      `json:"replicas"`
	RequestTimeout      int      `json:"requestTimeout"`
	EnableSearchAfter   bool     `json:"enableSearchAfter"`
	TrackTotalHits      bool     `json:"trackTotalHits"`
	IndexTypes          []string `json:"indexTypes"`
	MaxTreeSize         int      `json:"maxTreeSize"`
	OrphanWait          int      `json:"orphanWait"`
}

func (es *ElasticSearch) FragmentIndexName(orgID string) string {
	return fmt.Sprintf("%s_frag", es.GetIndexName(orgID))
}

func (es *ElasticSearch) GetSuggestIndexName(orgID string) string {
	return fmt.Sprintf("%s_suggest", es.GetIndexName(orgID))
}

// HasAuthentication returns if ElasticSearch has authentication enabled.
func (es *ElasticSearch) HasAuthentication() bool {
	return len(es.UserName) > 0 && len(es.Password) > 0
}

// GetIndexName returns the lowercased indexname.
// This inforced correct behavior when creating an index in ElasticSearch.
func (es *ElasticSearch) GetIndexName(orgID string) string {
	return strings.ToLower(orgID) + "v2"
}

func (es *ElasticSearch) GetV1IndexName(orgID string) string {
	return strings.ToLower(orgID) + "v1"
}

func (es *ElasticSearch) GetDigitalObjectIndexName(orgID string) string {
	if es.DigitalObjectSuffix == "" {
		return es.GetIndexName(orgID)
	}

	return strings.ToLower(orgID) + "v2-" + strings.ToLower(es.DigitalObjectSuffix)
}

// Logging holds all the logging and path configuration
type Logging struct {
	DevMode       bool   `json:"devmode"`
	SentryDSN     string `json:"sentrydsn"`
	Level         string `json:"level"`
	WithCaller    bool   `json:"withCaller"`
	ConsoleLogger bool   `json:"consoleLogger"`
}

// HTTP holds all the configuration for the http server subcommand
type HTTP struct {
	Port      int    `json:"port" mapstructure:"port"`
	StaticDir string `json:"staticDir"` // the relative path to the static directory to serve documentation.
	ProxyTLS  bool   `json:"proxyTLS"`
}

// RDF holds all the configuration for SPARQL queries and RDF conversions
type RDF struct {
	SparqlEnabled    bool   `json:"sparqlEnabled"`    // Enable the SPARQL proxy
	SparqlHost       string `json:"sparqlHost"`       // the base-url to the SPARQL endpoint including the scheme and the port
	SparqlPath       string `json:"sparqlPath"`       // the relative path of the endpoint. This can should contain the database name that is injected when the sparql endpoint is build
	SparqlUpdatePath string `json:"sparqlUpdatePath"` // the relative path of the update endpoint. This can should contain the database name that is injected when the sparql endpoint is build
	GraphStorePath   string `json:"dataPath"`         // the relative GraphStore path of the endpoint. This can should contain the database name that is injected when the sparql endpoint is build
	BaseURL          string `json:"baseUrl"`          // the RDF baseUrl used for minting new URIs (should not include scheme)
	BaseScheme       string `json:"baseScheme"`       // the scheme (http or https) used in the baseURL
	RDFStoreEnabled  bool   `json:"rdfStoreEnabled"`  // Store to Triple Store while saving RDF
	// the RDF entryPoints. Lookups are made on the fully qualified URIs. It is sometimes needed to support other baseUrls as well.
	// The entry-points need to be fully qualified, i.e. with their scheme.
	RoutedEntryPoints []string `json:"RoutedEntryPoints"`
	Tags              string   `json:"tags" mapstructure:"tags"`
	DefaultFormat     string   `json:"defaultFormat"`
	RDFStoreTags      []string `json:"rdfStoreTags"` // the tags that trigger storage in the triple-store
}

func (rdf *RDF) HasStoreTag(tags []string) bool {
	for _, tag := range tags {
		for _, storeTag := range rdf.RDFStoreTags {
			if strings.EqualFold(tag, storeTag) {
				return true
			}
		}
	}

	return false
}

// Cache is the configuration of the BigCache implementation
type Cache struct {
	Enabled           bool   `json:"enabled"`
	LifeWindowMinutes int    `json:"lifeWindowMinutes"`
	APIPrefix         string `json:"apiPrefix"`
	StripPrefix       bool   `json:"stripPrefix"`
	CacheDomain       string `json:"cacheDomain"`
	HardMaxCacheSize  int    `json:"hardMaxCacheSize"`
	MaxEntrySize      int    `json:"maxEntrySize"`
}

// OAIPMH holds all the configuration options for the OAI-PMH endpoint
type OAIPMH struct {
	// Make the oai-pmh endpoint available
	Enabled bool `json:"enabled"`
	// AdminEmails has a list of the admin emails of this endpoint
	AdminEmails []string `json:"adminEmails"`
	// RepositoryName is the name of the OAI-PMH repossitory
	RepositoryName  string `json:"repositoryName"`
	HarvestDelay    int    `json:"harvestDelay"`
	EadHarvestURL   string `json:"eadHarvestURL"`
	MetsHarvestURL  string `json:"metsHarvestURL"`
	MetsDownloadFmt string `json:"metsDownloadFmt"`
}

// WebResource holds all the configuration options for the WebResource endpoint
type WebResource struct {
	Enabled                   bool   `json:"enabled"`          // Make the webresource endpoint available
	WebResourceDir            string `json:"sourceDir"`        // Target directory for the webresources
	CacheResourceDir          string `json:"cacheDir"`         // cache directory for the webresources
	enableSearch              bool   `json:"enableSearch"`     // enable searching for images in ElasticSearch
	MediaManagerHost          string `json:"mediaManagerHost"` // the domain to build the derivatives
	SmallDefault              string `json:"smallDefault"`
	MediumDefault             string `json:"mediumDefault"`
	LargeDefault              string `json:"largeDefault"`
	MaxSize                   int32  `json:"maxSize"`
	ResolveRemoteWebResources bool   `json:"resolveRemoteWebResources"`
}

// ImageProxy holds all the configuration for the ImageProxy functionality
type ImageProxy struct {
	Enabled          bool     `json:"enabled"`          // Make the imageproxy endpoint available
	CacheDir         string   `json:"cacheDir"`         // The path to the imageCache
	Referrer         []string `json:"referrer"`         // A list of allowed referrers. If empty allow all.
	Whitelist        []string `json:"whitelist"`        // A list of allowed remote hosts. If empty allow all.
	ScaleUp          bool     `json:"scaleUp"`          // Allow images to scale beyond their original dimensions.
	TimeOut          int      `json:"timeout"`          // timelimit for request served by this proxy. 0 is for no timeout
	Deepzoom         bool     `json:"deepzoom"`         // Enable deepzoom of remote images.
	ProxyPrefix      string   `json:"proxyPrefix"`      // The prefix where we mount the imageproxy. default: imageproxy. default: imageproxy.
	DefaultImagePath string   `json:"defaultImagePath"` // Default image in case the image cannot be found.
}

// LOD holds all the configuration for the Linked Open Data (LOD) functionality
type LOD struct {
	Enabled           bool   `json:"enabled"`        // Make the lod endpoint available
	Resource          string `json:"resource"`       // the 303 redirect entry point. This is where the content negotiation happens
	HTML              string `json:"html"`           // the endpoint that renders the data as formatted HTML
	RDF               string `json:"rdf"`            // the endpoint that renders the RDF data in the requested RDF format. Currently, JSON-LD and N-triples are supported
	SingleEndpoint    string `json:"singleEndpoint"` // when this is set it overrides the other endpoints
	HTMLRedirectRegex string `json:"redirectregex"`  // the regular expression to convert the subject uri to the uri for the external Page view
}

// SiteMap holds all the configuration for the sitemap generation
type SiteMap struct {
	Enabled bool   `json:"enabled"`
	BaseDir string `json:"baseDir"`
	BaseURL string `json:"baseURL"`
	Gzip    bool   `json:"gzip"`
}

// EAD holds all the configuration for the EAD endpoint
type EAD struct {
	CacheDir         string   `json:"cacheDir"`
	SearchURL        string   `json:"searchURL"`
	GenreFormDefault string   `json:"genreFormDefault"`
	TreeFields       []string `json:"treeFields"`
	SearchFields     []string `json:"searchFields"`
}

func setDefaults() {
	// setting defaults
	viper.SetDefault("HTTP.port", 3001)
	viper.SetDefault("HTTP.staticDir", "public")
	viper.SetDefault("orgId", "hub3")

	// elastic
	viper.SetDefault("ElasticSearch.urls", []string{"http://localhost:9200"})
	viper.SetDefault("ElasticSearch.enabled", true)
	viper.SetDefault("ElasticSearch.IndexName", viper.GetString("OrgId"))
	viper.SetDefault("ElasticSearch.Proxy", true)
	viper.SetDefault("ElasticSearch.Fragments", true)
	viper.SetDefault("ElasticSearch.IndexV1", false)
	viper.SetDefault("ElasticSearch.EnableTrace", false)
	viper.SetDefault("ElasticSearch.EnableInfo", false)
	viper.SetDefault("ElasticSearch.SpecKey", "meta.spec")
	viper.SetDefault("ElasticSearch.RevisionKey", "meta.revision")
	viper.SetDefault("ElasticSearch.OrgIDKey", "meta.orgID")
	viper.SetDefault("ElasticSearch.FacetSize", 50)
	viper.SetDefault("ElasticSearch.MimimumShouldMatch", "2<70%")
	viper.SetDefault("ElasticSearch.workers", 1)
	viper.SetDefault("ElasticSearch.EnableSearchAfter", false)
	viper.SetDefault("ElasticSearch.Shards", 1)
	viper.SetDefault("ElasticSearch.Replicas", 0)
	viper.SetDefault("ElasticSearch.RequestTimeout", 15)
	viper.SetDefault("ElasticSearch.MaxTreeSize", 250)
	viper.SetDefault("ElasticSearch.OrphanWait", 15)
	viper.SetDefault("ElasticSearch.TrackTotalHits", true)
	viper.SetDefault("ElasticSearch.IndexTypes", []string{"v2"})

	// logging
	viper.SetDefault("Logging.DevMode", false)
	viper.SetDefault("Logging.level", "info")

	// cache
	viper.SetDefault("Cache.enabled", false)
	viper.SetDefault("Cache.lifeWindow", 10)
	viper.SetDefault("Cache.apiPrefix", "/api/http/cache")
	viper.SetDefault("Cache.stripPrefix", true)
	viper.SetDefault("Cache.cacheDomain", "")
	viper.SetDefault("Cache.HardMaxCacheSize", 512)
	viper.SetDefault("Cache.MaxEntrySize", 10240)

	// rdf with defaults for Apache Fuseki
	viper.SetDefault("RDF.SparqlEnabled", true)
	viper.SetDefault("RDF.SparqlHost", "http://localhost:3030")
	viper.SetDefault("RDF.SparqlPath", "/%s/sparql")
	viper.SetDefault("RDF.SparqlUpdatePath", "/%s/update")
	viper.SetDefault("RDF.GraphStorePath", "/%s/data")
	viper.SetDefault("RDF.BaseUrl", "http://data.hub3.org")
	viper.SetDefault("RDF.BaseScheme", "http")
	viper.SetDefault("RDF.RoutedEntryPoints", []string{"http://localhost:3000", "http://localhost:3001"})
	viper.SetDefault("RDF.RDFStoreEnabled", false)
	viper.SetDefault("RDF.DefaultFormat", "application/ld+json")
	viper.SetDefault("RDF.RDFStoreTags", []string{"narthex", "mdr"})

	// rdftags
	viper.SetDefault("RDFTag.Label", []string{
		"http://www.w3.org/2004/02/skos/core#prefLabel",
		"http://purl.org/dc/elements/1.1/title",
	})

	// posthook
	viper.SetDefault("PostHook.urls", []string{})
	viper.SetDefault("PostHook.ExcludeSpec", []string{})

	// oai-pmh
	viper.SetDefault("OAIPMH.enabled", true)
	viper.SetDefault("OAIPMH.RepositoryName", "hub3")
	viper.SetDefault("OAIPMH.AdminEmails", "info@delving.eu")
	viper.SetDefault("OAIPMH.HarvestDelay", 1)
	viper.SetDefault("OAIPMH.EadHarvestUrl", "")
	viper.SetDefault("OAIPMH.MetsHarvestUrl", "")

	// image proxy
	viper.SetDefault("ImageProxy.enabled", true)
	viper.SetDefault("ImageProxy.CacheDir", "webresource_cache/cache")
	viper.SetDefault("ImageProxy.referrer", []string{})
	viper.SetDefault("ImageProxy.whitelist", []string{})
	viper.SetDefault("ImageProxy.scaleUp", false)
	viper.SetDefault("ImageProxy.timeout", 0)
	viper.SetDefault("ImageProxy.deepzoom", true)
	viper.SetDefault("ImageProxy.proxyPrefix", "imageproxy")
	viper.SetDefault("ImageProxy.DefaultImagePath", "")

	// webresource
	viper.SetDefault("WebResource.enabled", true)
	viper.SetDefault("WebResource.WebResourceDir", "/tmp/webresource")
	viper.SetDefault("WebResource.CacheResourceDir", "/tmp/webresource_cache")
	viper.SetDefault("WebResource.enableSearch", false)
	viper.SetDefault("WebResource.resolveRemoteWebResources", false)
	viper.SetDefault("WebResource.mediaManagerHost", "https://media.delving.org")
	viper.SetDefault("WebResource.SmallDefault", "220")
	viper.SetDefault("WebResource.MediumDefault", "350")
	viper.SetDefault("WebResource.LargeDefault", "500")
	viper.SetDefault("WebResource.MaxSize", int32(2000))

	// lod
	viper.SetDefault("LOD.enabled", true)
	viper.SetDefault("LOD.html", "page")
	viper.SetDefault("LOD.rdf", "data")
	viper.SetDefault("LOD.singleEndpoint", "")
	viper.SetDefault("LOD.resource", "resource")
	viper.SetDefault("LOD.redirectregex", "")

	// sitemap
	viper.SetDefault("SiteMap.Enabled", false)

	// ead
	viper.SetDefault("EAD.CacheDir", "/tmp/ead")
}

func cleanConfig() {
	Config.RDF.BaseURL = strings.TrimSuffix(Config.RDF.BaseURL, "/")
	if !strings.HasPrefix(Config.RDF.BaseScheme, "http") {
		Config.Logger.Fatal().
			Str("baseURL", Config.RDF.BaseURL).
			Msg("RDF.BaseUrl config value '%s' should start with 'http' or 'https'.")
	}
}

func SetCfgFile(cfgFile string) {
	CfgFile = cfgFile
}

// InitConfig reads in config file and ENV variables if set.
func InitConfig() {
	// InitConfig() must be idempotent
	if Config.Logger != nil {
		return
	}

	if CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".hub3" (without extension).
		viper.AddConfigPath("/etc/default/")
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName("hub3")
	}

	viper.SetEnvPrefix("HUB3")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	setDefaults()

	logCfg := logger.Config{
		LogLevel:            logger.ParseLogLevel(Config.Logging.Level),
		WithCaller:          Config.Logging.WithCaller,
		EnableConsoleLogger: Config.Logging.ConsoleLogger,
	}

	configLogger := logger.NewLogger(logCfg)
	Config.Logger = &configLogger.Logger

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err == nil {
		Config.Logger.Info().
			Str("configPath", viper.ConfigFileUsed()).
			Msg("starting up with config path (legacy)")
	} else {
		log.Printf("Unable to read config file %s", viper.ConfigFileUsed())
		switch v := err.(type) {
		case viper.ConfigParseError:
			log.Fatalf("config parse error: %#v", v.Error())
		default:
			log.Printf("config parse error: %#v", err)
		}
	}

	err = viper.Unmarshal(&Config)
	if err != nil {
		log.Fatal(
			fmt.Sprintf("unable to decode into struct, %v", err),
		)
	}

	Config.NameSpaceMap = NewConfigNameSpaceMap(&Config)
	Config.RDFTagMap = NewRDFTagMap(&Config)
	Config.DatasetTagMap = NewDataSetTagMap(&Config)

	cleanConfig()
}

// GetSparqlEndpoint builds the SPARQL endpoint from the RDF Config object.
// When the dbName is empty the OrgId from the configuration is used.
func (c *RawConfig) GetSparqlEndpoint(orgID, dbName string) string {
	if dbName == "" {
		dbName = orgID
	}

	u, err := url.Parse(c.RDF.SparqlHost)
	if err != nil {
		log.Fatal(err)
	}

	u.Path = fmt.Sprintf(c.RDF.SparqlPath, dbName)

	return u.String()
}

// GetSparqlUpdateEndpoint builds the SPARQL Update endpoint from the RDF Config object.
// When the dbName is empty the OrgId from the configuration is used.
func (c *RawConfig) GetSparqlUpdateEndpoint(orgID, dbName string) string {
	if dbName == "" {
		dbName = orgID
	}

	u, err := url.Parse(c.RDF.SparqlHost)
	if err != nil {
		log.Fatal(err)
	}

	u.Path = fmt.Sprintf(c.RDF.SparqlUpdatePath, dbName)

	return u.String()
}

// GetGraphStoreEndpoint builds the GraphStore endpoint from the RDF Config object.
// When the dbName is empty the OrgId from the configuration is used.
func (c *RawConfig) GetGraphStoreEndpoint(orgID, dbName string) string {
	if dbName == "" {
		dbName = orgID
	}

	u, err := url.Parse(c.RDF.SparqlHost)
	if err != nil {
		log.Fatal(err)
	}

	u.Path = fmt.Sprintf(c.RDF.GraphStorePath, dbName)

	log.Printf("GraphStore endpoint: %s", u)

	return u.String()
}

// Save saves the update version of the configuration file
// At the moment this is mostly used for persisting the namespaces
func (c *RawConfig) Save() error {
	return viper.SafeWriteConfig()
}

// BuildVersionInfo holds all the version information
type BuildVersionInfo struct {
	Version    string `json:"version"`
	Commit     string `json:"commit"`
	BuildAgent string `json:"buildAgent"`
	BuildDate  string `json:"buildDate"`
	BuildID    string `json:"buildID"`
}

// NewBuildVersionInfo creates a BuildVersionInfo struct
func NewBuildVersionInfo(version, commit, buildagent, builddate string) *BuildVersionInfo {
	if version == "" {
		version = "devBuild"
	}

	return &BuildVersionInfo{
		Version:    version,
		Commit:     commit,
		BuildAgent: buildagent,
		BuildDate:  builddate,
	}
}

// JSON returns a json version of the BuildVersionInfo
func (b *BuildVersionInfo) JSON(pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(b, "", "\t")
	}

	return json.Marshal(b)
}
