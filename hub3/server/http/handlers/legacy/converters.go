package legacy

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/delving/hub3/ikuzo/domain"
)

// ConverterType represents the different converter types available
type ConverterType string

const (
	ESE       ConverterType = "ese"
	EDM       ConverterType = "edm"
	EDMStrict ConverterType = "edm-strict"
	TIB       ConverterType = "tib"
	ICN       ConverterType = "icn"
	ABM       ConverterType = "abm"
	V2        ConverterType = "v2"
)

// FieldMapping represents a mapping from one field to another
type FieldMapping struct {
	Source string `toml:"source"`
	Target string `toml:"target"`
}

// ConverterConfig represents the configuration for a specific converter
type ConverterConfig struct {
	Key            string            `toml:"key"`
	Namespace      string            `toml:"namespace"`
	Mappings       []FieldMapping    `toml:"mappings"`
	QueryReplacers map[string]string `toml:"query_replacers"`
	V1Config       *domain.V1Config
}

// DefaultsConfig represents default values used in conversion
type DefaultsConfig struct {
	RecordType            string `toml:"record_type"`
	HasDigitalObjectField string `toml:"has_digital_object_field"`
	HasGeoHashField       string `toml:"has_geo_hash_field"`
	HasLandingPageField   string `toml:"has_landing_page_field"`
	FilePathPrefix        string `toml:"file_path_prefix"`
}

// Config represents the entire configuration structure
type Config struct {
	Base struct {
		Mappings []FieldMapping `toml:"mappings"`
	} `toml:"base"`
	Delving struct {
		Mappings []FieldMapping `toml:"mappings"`
	} `toml:"delving"`
	Converters map[string]ConverterConfig `toml:"converters"`
	Defaults   DefaultsConfig             `toml:"defaults"`
}

// Converter defines the interface that all format converters should implement
type Converter interface {
	GetFieldMapping() map[string]string
	GetNamespace() string
	GetConverterKey() ConverterType
	Convert(input map[string][]string, addDelvingFields bool) map[string][]string
	GetQueryReplacers(reverse bool) map[string]string
	ReplaceQueryString(queryString string, reverse bool) string
	Configuration() *domain.V1Config
	Layout() map[string]string
}

// BaseConverter implements common functionality for all converters
type BaseConverter struct {
	AboutURI        string
	OrgID           string
	ConverterKey    ConverterType
	Namespace       string
	FieldMappings   map[string]string
	DelvingMappings map[string]string
	QueryReplacers  map[string]string
	Defaults        DefaultsConfig
}

// Configuration returns the domain.V1Config object
func (bc *BaseConverter) Configuration() *domain.V1Config {
	cfg, ok := defaultRegistry.registry[bc.OrgID]
	if !ok {
		return nil
	}
	return cfg.V1Config
}

func (bc *BaseConverter) Layout() (layout map[string]string) {
	layout = map[string]string{}
	for k := range bc.DelvingMappings {
		layout[k] = k
	}

	for k := range bc.FieldMappings {
		layout[k] = k
	}

	return layout
}

// GetConverterKey returns the converter's key identifier
func (bc *BaseConverter) GetConverterKey() ConverterType {
	return bc.ConverterKey
}

// GetNamespace returns the converter's namespace
func (bc *BaseConverter) GetNamespace() string {
	return bc.Namespace
}

// GetFieldMapping returns the converter's field mappings
func (bc *BaseConverter) GetFieldMapping() map[string]string {
	return bc.FieldMappings
}

// GetQueryReplacers returns the query string replacement mappings
// If reverse is true, the keys and values are swapped
func (bc *BaseConverter) GetQueryReplacers(reverse bool) map[string]string {
	replacers := make(map[string]string)

	// If no custom replacers are set, use defaults
	if len(bc.QueryReplacers) == 0 {
		replacers["europeana_"] = "edm_"
		replacers["tib_"] = "nave_"
		replacers["icn_"] = "nave_"
		replacers["abm_"] = "nave_"
		replacers["delving_"] = "nave_"
	} else {
		// Use custom replacers from config
		for k, v := range bc.QueryReplacers {
			replacers[k] = v
		}
	}

	// If reverse requested, swap keys and values
	if reverse {
		reversed := make(map[string]string)
		for k, v := range replacers {
			reversed[v] = k
		}
		return reversed
	}

	return replacers
}

// ReplaceQueryString applies the query string replacements to the input string
// If reverse is true, the replacement is done in reverse direction
func (bc *BaseConverter) ReplaceQueryString(queryString string, reverse bool) string {
	if queryString == "" {
		return queryString
	}

	replacers := bc.GetQueryReplacers(reverse)
	result := queryString

	// Apply all replacements
	for search, replace := range replacers {
		result = strings.ReplaceAll(result, search, replace)
	}

	return result
}

// getHubID extracts the hub ID and specification from the AboutURI
func (bc *BaseConverter) getHubID() (string, string) {
	parts := strings.Split(bc.AboutURI, "/")
	if len(parts) < 2 {
		return "", ""
	}

	spec := parts[len(parts)-2]
	localID := parts[len(parts)-1]

	// Clean localID similar to RDFRecord.clean_local_id
	re := regexp.MustCompile(`[^\w-]`)
	cleanLocalID := re.ReplaceAllString(localID, "")

	hubID := fmt.Sprintf("%s_%s_%s", bc.OrgID, spec, cleanLocalID)
	return hubID, spec
}

// addDefaults adds standard fields to the output document
func (bc *BaseConverter) addDefaults(outputDoc map[string][]string, inputDoc map[string]string) {
	outputDoc["delving_recordType"] = []string{bc.Defaults.RecordType}

	hubID, spec := bc.getHubID()
	outputDoc["delving_hubId"] = []string{hubID}
	outputDoc["delving_pmhId"] = []string{hubID}
	outputDoc["delving_spec"] = []string{spec}
	outputDoc["europeana_uri"] = []string{strings.Join(strings.Split(hubID, "_")[1:], "/")}

	// Add resourceUri as europeana_object if available
	if _, ok := outputDoc["delving_resourceUri"]; ok {
		outputDoc["europeana_object"] = outputDoc["delving_resourceUri"]
	}

	// Set hasDigitalObject flag
	_, hasObject := outputDoc["europeana_object"]
	outputDoc[bc.Defaults.HasDigitalObjectField] = []string{fmt.Sprintf("%t", hasObject)}

	// Add thumbnail
	if _, ok := outputDoc["europeana_object"]; ok {
		if _, hasThumbnail := outputDoc["delving_thumbnail"]; hasThumbnail {
			outputDoc["delving_thumbnail"] = append(outputDoc["delving_thumbnail"], outputDoc["europeana_object"]...)
		} else {
			outputDoc["delving_thumbnail"] = outputDoc["europeana_object"]
		}
	}

	// Handle geography
	if _, ok := outputDoc["delving_locationLatLong"]; ok {
		outputDoc["delving_geoHash"] = outputDoc["delving_locationLatLong"]
	}
	_, hasGeoHash := outputDoc["delving_geoHash"]
	outputDoc[bc.Defaults.HasGeoHashField] = []string{fmt.Sprintf("%t", hasGeoHash)}

	// Handle isShownAt
	if shownAt, ok := outputDoc["europeana_isShownAt"]; ok {
		shouldDelete := false

		for _, path := range shownAt {
			if strings.HasPrefix(path, bc.Defaults.FilePathPrefix) {
				shouldDelete = true
				break
			}
		}

		if shouldDelete {
			delete(outputDoc, "europeana_isShownAt")
		} else {
			outputDoc["delving_landingpage"] = shownAt
		}
	}

	// Add title
	if title, ok := outputDoc["dc_title"]; ok {
		outputDoc["delving_title"] = title
	}

	// Set hasLandingPage flag
	_, hasLandingPage := outputDoc["europeana_isShownAt"]
	outputDoc[bc.Defaults.HasLandingPageField] = []string{fmt.Sprintf("%t", hasLandingPage)}

	// Add collection name
	outputDoc["europeana_collectionName"] = []string{spec}

	// Set organization ID
	outputDoc["delving_orgId"] = []string{bc.OrgID}
}

// GetNonNull adds a non-null value from input to output
func (bc *BaseConverter) getNonNull(key string, input map[string]string, output map[string][]string) {
	if value, ok := input[key]; ok && value != "" && value != "null" {
		output[key] = []string{value}
	}
}

// Convert converts input fields according to the converter's mappings
func (bc *BaseConverter) Convert(input map[string][]string, addDelvingFields bool) map[string][]string {
	// Create a flattened input map for processing
	flatInput := make(map[string]string)
	for k, v := range input {
		if len(v) > 0 {
			flatInput[k] = v[0]
		}
	}

	mapping := bc.GetFieldMapping()
	if addDelvingFields {
		maps.Copy(mapping, bc.DelvingMappings)
	}

	outputDoc := make(map[string][]string)

	// Process each field according to the mapping
	for key, indexDocKey := range mapping {
		if values, ok := input[indexDocKey]; ok && len(values) > 0 {
			outputDoc[key] = values
		}
	}

	// Add default fields if requested
	if addDelvingFields {
		bc.addDefaults(outputDoc, flatInput)
	}

	return outputDoc
}

// GetTranslatedField normalizes field names
func GetTranslatedField(key string, translate bool) string {
	// First normalize prefixes
	normalizedKey := key
	normalizedKey = regexp.MustCompile(`abm_|tib_|icn_|delving_`).ReplaceAllString(normalizedKey, "nave_")
	normalizedKey = regexp.MustCompile(`europeana_`).ReplaceAllString(normalizedKey, "edm_")

	// Then remove legacy suffixes
	for _, suffix := range []string{"_string", "_facet", "_text"} {
		if strings.HasSuffix(normalizedKey, suffix) {
			parts := strings.Split(normalizedKey, "_")
			normalizedKey = strings.Join(parts[:len(parts)-1], "_")
			break
		}
	}

	// TODO: Add translation functionality if needed
	// For now, just return the normalized key
	return normalizedKey
}

// ConfigManager handles loading and access to configuration
type ConfigManager struct {
	config *Config
}

// NewConfigManager creates a new ConfigManager with the given config file
func NewConfigManager(configPath string) (*ConfigManager, error) {
	// Ensure the config file exists
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", absPath)
	}

	var config Config
	if _, err := toml.DecodeFile(absPath, &config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &ConfigManager{config: &config}, nil
}

// GetBaseFieldMappings converts the base field mappings to a map
func (cm *ConfigManager) GetBaseFieldMappings() map[string]string {
	result := make(map[string]string)
	for _, mapping := range cm.config.Base.Mappings {
		result[mapping.Source] = mapping.Target
	}
	return result
}

// GetDelvingMappings converts the delving field mappings to a map
func (cm *ConfigManager) GetDelvingMappings() map[string]string {
	result := make(map[string]string)
	for _, mapping := range cm.config.Delving.Mappings {
		result[mapping.Source] = mapping.Target
	}
	return result
}

// GetConverterConfig gets the configuration for a specific converter
func (cm *ConfigManager) GetConverterConfig(converterType ConverterType) (ConverterConfig, bool) {
	config, ok := cm.config.Converters[string(converterType)]
	return config, ok
}

// GetDefaults returns the default values configuration
func (cm *ConfigManager) GetDefaults() DefaultsConfig {
	return cm.config.Defaults
}

// NewConverterWithConfig creates a new converter with the given configuration
func NewConverterWithConfig(cm *ConfigManager, converterType ConverterType, aboutURI, orgID string) (Converter, error) {
	// Get base mappings
	baseMappings := cm.GetBaseFieldMappings()
	delvingMappings := cm.GetDelvingMappings()
	defaults := cm.GetDefaults()

	// Get converter-specific configuration
	converterConfig, ok := cm.GetConverterConfig(converterType)
	if !ok {
		// If converter not found, use default V2 converter
		converterType = V2
		converterConfig, _ = cm.GetConverterConfig(converterType)
	}

	// Create the field mappings by starting with base mappings
	fieldMappings := make(map[string]string)
	maps.Copy(fieldMappings, baseMappings)

	// Add converter-specific mappings
	for _, mapping := range converterConfig.Mappings {
		fieldMappings[mapping.Source] = mapping.Target
	}

	// Create the base converter
	baseConverter := &BaseConverter{
		AboutURI:        aboutURI,
		OrgID:           orgID,
		ConverterKey:    ConverterType(converterConfig.Key),
		Namespace:       converterConfig.Namespace,
		FieldMappings:   fieldMappings,
		DelvingMappings: delvingMappings,
		QueryReplacers:  converterConfig.QueryReplacers,
		Defaults:        defaults,
	}

	return baseConverter, nil
}

// OrgConverterConfig holds the configuration for a specific organization's default converter
type OrgConverterConfig struct {
	ConfigManager *ConfigManager
	ConverterType ConverterType
	V1Config      *domain.V1Config
}

// ParseConverterType converts a string to a ConverterType
// Returns the parsed type or V2 if not found
func ParseConverterType(typeStr string) ConverterType {
	switch ConverterType(typeStr) {
	case ESE:
		return ESE
	case EDM:
		return EDM
	case EDMStrict:
		return EDMStrict
	case TIB:
		return TIB
	case ICN:
		return ICN
	case ABM:
		return ABM
	case V2:
		return V2
	default:
		return V2
	}
}

// DefaultConverterRegistry manages organization-specific default converters
type DefaultConverterRegistry struct {
	mu       sync.RWMutex
	registry map[string]OrgConverterConfig
}

// Global registry for default converters
var (
	defaultRegistry     = &DefaultConverterRegistry{registry: make(map[string]OrgConverterConfig)}
	defaultFallbackType = EDM
)

// RegisterOrgConverter registers a default converter configuration for an organization
// This should be called during application startup for each organization
func RegisterOrgConverter(orgID string, configPath string, converterType ConverterType, cfg *domain.V1Config) error {
	// Load config from file
	cm, err := NewConfigManager(configPath)
	if err != nil {
		return fmt.Errorf("failed to initialize converter for org %s: %w", orgID, err)
	}

	// Register the configuration
	defaultRegistry.mu.Lock()
	defaultRegistry.registry[orgID] = OrgConverterConfig{
		ConfigManager: cm,
		ConverterType: converterType,
		V1Config:      cfg,
	}
	defaultRegistry.mu.Unlock()

	return nil
}

// RegisterOrgConverterWithTypeString registers a default converter using a string converter type
// This is useful when the converter type comes from a configuration file as a string
func RegisterOrgConverterWithTypeString(orgID string, cfg *domain.V1Config) error {
	converterType := ParseConverterType(cfg.Convertor.Type)
	return RegisterOrgConverter(orgID, cfg.Convertor.Path, converterType, cfg)
}

// DefaultConverter returns a converter instance using the configuration for the specified organization
// If no configuration exists for the organization, it falls back to a default in-memory configuration
func DefaultConverter(aboutURI, orgID string) (Converter, error) {
	// Get org-specific config if it exists
	defaultRegistry.mu.RLock()
	orgConfig, exists := defaultRegistry.registry[orgID]
	defaultRegistry.mu.RUnlock()

	// If org-specific config exists, use it
	if exists {
		return NewConverterWithConfig(orgConfig.ConfigManager, orgConfig.ConverterType, aboutURI, orgID)
	}

	// Otherwise, use in-memory default
	return NewConverter(defaultFallbackType, aboutURI, orgID), nil
}

// UpdateOrgConverterType changes the converter type for a specific organization
func UpdateOrgConverterType(orgID string, converterType ConverterType) bool {
	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()

	config, exists := defaultRegistry.registry[orgID]
	if !exists {
		return false
	}

	config.ConverterType = converterType
	defaultRegistry.registry[orgID] = config
	return true
}

// SetDefaultFallbackType sets the default converter type used when an organization-specific
// configuration is not found
func SetDefaultFallbackType(converterType ConverterType) {
	defaultFallbackType = converterType
}

// ListRegisteredOrgs returns a list of all organization IDs with registered converters
func ListRegisteredOrgs() []string {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()

	orgs := make([]string, 0, len(defaultRegistry.registry))
	for org := range defaultRegistry.registry {
		orgs = append(orgs, org)
	}

	return orgs
}

// NewConverter creates a new converter for compatibility with old code
// This is a convenience function that uses embedded default configuration
func NewConverter(converterType ConverterType, aboutURI, orgID string) Converter {
	// Create in-memory default config
	config := &Config{}

	// Set up base mappings similar to baseSharedFieldDict
	config.Base.Mappings = []FieldMapping{
		{Source: "dc_title", Target: "dc_title"},
		{Source: "dc_creator", Target: "dc_creator"},
		// Add all other base mappings from the TOML here
	}

	// Set up delving mappings
	config.Delving.Mappings = []FieldMapping{
		{Source: "delving_year", Target: "delving_year"},
		{Source: "delving_thumbnail", Target: "nave_thumbnail"},
		// Add all other delving mappings from the TOML here
	}

	// Set up converter configs for all supported types
	config.Converters = map[string]ConverterConfig{
		string(ESE): {
			Key:       string(ESE),
			Namespace: "http://www.europeana.eu/schemas/ese/",
			Mappings:  []FieldMapping{},
			QueryReplacers: map[string]string{
				"europeana_": "edm_",
			},
		},
		string(EDM): {
			Key:       string(EDM),
			Namespace: "http://www.europeana.eu/schemas/edm/",
			Mappings:  []FieldMapping{},
			QueryReplacers: map[string]string{
				"europeana_": "edm_",
				"tib_":       "nave_",
			},
		},
		string(EDMStrict): {
			Key:       string(EDMStrict),
			Namespace: "http://www.europeana.eu/schemas/edm/",
			Mappings:  []FieldMapping{},
			QueryReplacers: map[string]string{
				"europeana_": "edm_",
			},
		},
		string(TIB): {
			Key:       string(TIB),
			Namespace: "http://schemas.delving.eu/resource/ns/tib/",
			Mappings:  []FieldMapping{},
			QueryReplacers: map[string]string{
				"europeana_": "edm_",
				"tib_":       "nave_",
			},
		},
		string(ICN): {
			Key:       string(ICN),
			Namespace: "http://www.icn.nl/schemas/icn/",
			Mappings:  []FieldMapping{},
			QueryReplacers: map[string]string{
				"europeana_": "edm_",
				"icn_":       "nave_",
			},
		},
		string(ABM): {
			Key:       string(ABM),
			Namespace: "http://purl.org/abm/sen",
			Mappings:  []FieldMapping{},
			QueryReplacers: map[string]string{
				"europeana_": "edm_",
				"abm_":       "nave_",
			},
		},
		string(V2): {
			Key:       string(V2),
			Namespace: "",
			Mappings:  []FieldMapping{},
			QueryReplacers: map[string]string{
				"europeana_": "edm_",
				"delving_":   "nave_",
			},
		},
	}

	// Set up defaults
	config.Defaults = DefaultsConfig{
		RecordType:            "mdr",
		HasDigitalObjectField: "delving_hasDigitalObject",
		HasGeoHashField:       "delving_hasGeoHash",
		HasLandingPageField:   "delving_hasLandingPage",
		FilePathPrefix:        "file:///opt",
	}

	// Create config manager with in-memory config
	cm := &ConfigManager{config: config}

	// Create and return converter
	converter, _ := NewConverterWithConfig(cm, converterType, aboutURI, orgID)
	return converter
}
