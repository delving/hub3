package legacy

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/matryer/is"
)

// getTestConfigManager creates a ConfigManager with test configuration
func getTestConfigManager(t *testing.T) *ConfigManager {
	// Create a temporary config file for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.toml")

	// Write test configuration to the file
	err := os.WriteFile(configPath, []byte(testConfigToml), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create and return the config manager
	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	return cm
}

func TestConfigManagerLoading(t *testing.T) {
	is := is.New(t)

	cm := getTestConfigManager(t)

	// Test that base mappings were loaded
	baseMappings := cm.GetBaseFieldMappings()
	is.Equal("dc_title", baseMappings["dc_title"])
	is.Equal("edm_type", baseMappings["europeana_type"])

	// Test that converter configs were loaded
	tibConfig, ok := cm.GetConverterConfig(TIB)
	is.True(ok)
	is.Equal("tib", tibConfig.Key)
	is.Equal("http://schemas.delving.eu/resource/ns/tib/", tibConfig.Namespace)

	// Test that query replacers were loaded
	is.Equal("edm_", tibConfig.QueryReplacers["europeana_"])
	is.Equal("nave_", tibConfig.QueryReplacers["tib_"])

	// Test that other converter types also have their query replacers
	edmConfig, ok := cm.GetConverterConfig(EDM)
	is.True(ok)
	is.Equal("edm_", edmConfig.QueryReplacers["europeana_"])
	is.Equal(1, len(edmConfig.QueryReplacers)) // Should only have one replacer

	v2Config, ok := cm.GetConverterConfig(V2)
	is.True(ok)
	is.Equal("edm_", v2Config.QueryReplacers["europeana_"])
	is.Equal("nave_", v2Config.QueryReplacers["delving_"])
	is.Equal(2, len(v2Config.QueryReplacers)) // Should have two replacers

	// Test that delving mappings were loaded
	delvingMappings := cm.GetDelvingMappings()
	is.Equal("nave_thumbnail", delvingMappings["delving_thumbnail"])

	// Test that defaults were loaded
	defaults := cm.GetDefaults()
	is.Equal("mdr", defaults.RecordType)
}

func TestNewConverterWithConfig(t *testing.T) {
	is := is.New(t)

	cm := getTestConfigManager(t)

	// Test different converter types
	cases := []struct {
		converterType ConverterType
		expectedKey   ConverterType
		expectedNS    string
	}{
		{ESE, ESE, "http://www.europeana.eu/schemas/ese/"},
		{EDM, EDM, "http://www.europeana.eu/schemas/edm/"},
		{EDMStrict, EDMStrict, "http://www.europeana.eu/schemas/edm/"},
		{TIB, TIB, "http://schemas.delving.eu/resource/ns/tib/"},
		{ICN, ICN, "http://www.icn.nl/schemas/icn/"},
		{ABM, ABM, "http://purl.org/abm/sen"},
		{V2, V2, ""},
		{"unknown", V2, ""}, // Default to V2 for unknown types
	}

	for _, tc := range cases {
		t.Run(string(tc.converterType), func(t *testing.T) {
			converter, err := NewConverterWithConfig(
				cm,
				tc.converterType,
				"http://example.org/resource/test/123",
				"test_org",
			)
			is.NoErr(err)
			is.Equal(tc.expectedKey, converter.GetConverterKey())
			is.Equal(tc.expectedNS, converter.GetNamespace())
		})
	}
}

func TestESEConverterWithConfig(t *testing.T) {
	is := is.New(t)

	cm := getTestConfigManager(t)

	// Create test input
	input := map[string][]string{
		"dc_title":   {"Test Title"},
		"dc_creator": {"Test Creator"},
		"edm_type":   {"IMAGE"},
		"edm_rights": {"Public Domain"},
	}

	converter, err := NewConverterWithConfig(
		cm,
		ESE,
		"http://example.org/resource/test/123",
		"test_org",
	)
	is.NoErr(err)

	result := converter.Convert(input, true)

	// Check that mapped fields exist in the result
	is.Equal([]string{"Test Title"}, result["dc_title"])
	is.Equal([]string{"Test Creator"}, result["dc_creator"])

	// Check that Europeana fields are properly mapped
	is.Equal([]string{"IMAGE"}, result["europeana_type"])
	is.Equal([]string{"Public Domain"}, result["europeana_rights"])

	// Check that default fields were added
	is.Equal([]string{"mdr"}, result["delving_recordType"])
	is.Equal([]string{"test_org_test_123"}, result["delving_hubId"])
	is.Equal([]string{"test"}, result["delving_spec"])
}

func TestTIBConverterWithConfig(t *testing.T) {
	is := is.New(t)

	cm := getTestConfigManager(t)

	// Create test input with TIB-specific fields
	input := map[string][]string{
		"dc_title":         {"Test Title"},
		"nave_material":    {"Canvas"},
		"nave_technique":   {"Oil painting"},
		"nave_color":       {"Blue", "Red"},
		"nave_objectSoort": {"Painting"},
	}

	converter, err := NewConverterWithConfig(
		cm,
		TIB,
		"http://example.org/resource/test/123",
		"test_org",
	)
	is.NoErr(err)

	result := converter.Convert(input, true)

	// Check that base fields are mapped correctly
	is.Equal([]string{"Test Title"}, result["dc_title"])

	// Check that TIB-specific fields are mapped correctly
	is.Equal([]string{"Canvas"}, result["tib_material"])
	is.Equal([]string{"Oil painting"}, result["tib_technique"])
	is.Equal([]string{"Blue", "Red"}, result["tib_color"])
	is.Equal([]string{"Painting"}, result["tib_objectSoort"])

	// Check default fields
	is.Equal([]string{"test"}, result["europeana_collectionName"])
}

func TestICNConverterWithConfig(t *testing.T) {
	is := is.New(t)

	cm := getTestConfigManager(t)

	// Create test input with ICN-specific fields
	input := map[string][]string{
		"dc_title":             {"Test Title"},
		"nave_material":        {"Stone"},
		"nave_productionPlace": {"Amsterdam"},
		"nave_dimension":       {"100 x 200 cm"},
	}

	converter, err := NewConverterWithConfig(
		cm,
		ICN,
		"http://example.org/resource/test/123",
		"test_org",
	)
	is.NoErr(err)

	result := converter.Convert(input, true)

	// Check that ICN-specific fields are mapped correctly
	is.Equal([]string{"Stone"}, result["icn_material"])
	is.Equal([]string{"Amsterdam"}, result["icn_productionPlace"])
	is.Equal([]string{"100 x 200 cm"}, result["icn_dimension"])

	// Check that default fields were added
	is.Equal([]string{"test_org"}, result["delving_orgId"])
}

func TestABMConverterWithConfig(t *testing.T) {
	is := is.New(t)

	cm := getTestConfigManager(t)

	// Create test input with ABM-specific fields
	input := map[string][]string{
		"dc_title":     {"Test Title"},
		"nave_county":  {"Oslo"},
		"nave_country": {"Norway"},
		"nave_latLong": {"59.9139,10.7522"},
	}

	converter, err := NewConverterWithConfig(
		cm,
		ABM,
		"http://example.org/resource/test/123",
		"test_org",
	)
	is.NoErr(err)

	result := converter.Convert(input, true)

	// Check that ABM-specific fields are mapped correctly
	is.Equal([]string{"Oslo"}, result["abm_county"])
	is.Equal([]string{"Norway"}, result["abm_country"])
	is.Equal([]string{"59.9139,10.7522"}, result["abm_latLong"])

	// Check that default fields were added
	is.Equal([]string{"false"}, result["delving_hasGeoHash"])
}

func TestV2ConverterWithConfig(t *testing.T) {
	is := is.New(t)

	cm := getTestConfigManager(t)

	// Create test input with some narthex fields (which should be filtered)
	input := map[string][]string{
		"dc_title":       {"Test Title"},
		"dc_creator":     {"Test Creator"},
		"narthex_record": {"Some internal data"},
		"narthex_meta":   {"More internal data"},
	}

	converter, err := NewConverterWithConfig(
		cm,
		V2,
		"http://example.org/resource/test/123",
		"test_org",
	)
	is.NoErr(err)

	// For V2, we implement special filtering of narthex_ fields in the Convert method
	result := converter.Convert(input, false)

	// Check that normal fields are preserved
	is.Equal([]string{"Test Title"}, result["dc_title"])
	is.Equal([]string{"Test Creator"}, result["dc_creator"])

	// Check that narthex fields are filtered out
	_, hasNarthexRecord := result["narthex_record"]
	is.Equal(false, hasNarthexRecord)

	_, hasNarthexMeta := result["narthex_meta"]
	is.Equal(false, hasNarthexMeta)
}

func TestTranslateField(t *testing.T) {
	is := is.New(t)

	testCases := []struct {
		input    string
		expected string
	}{
		{"europeana_type", "edm_type"},
		{"abm_country", "nave_country"},
		{"tib_material", "nave_material"},
		{"icn_technique", "nave_technique"},
		{"delving_color", "nave_color"},
		{"nave_person_facet", "nave_person"},
		{"dc_title_string", "dc_title"},
	}

	for _, tc := range testCases {
		result := GetTranslatedField(tc.input, false)
		is.Equal(tc.expected, result)
	}
}

func TestBaseConverterAddDefaults(t *testing.T) {
	is := is.New(t)

	// Create default config
	defaults := DefaultsConfig{
		RecordType:            "mdr",
		HasDigitalObjectField: "delving_hasDigitalObject",
		HasGeoHashField:       "delving_hasGeoHash",
		HasLandingPageField:   "delving_hasLandingPage",
		FilePathPrefix:        "file:///opt",
	}

	// Create a BaseConverter with defaults
	bc := &BaseConverter{
		AboutURI: "http://example.org/resource/collection/item/123",
		OrgID:    "test-org",
		Defaults: defaults,
	}

	outputDoc := make(map[string][]string)
	inputDoc := make(map[string]string)

	// Add default fields
	bc.addDefaults(outputDoc, inputDoc)

	// Verify expected default fields
	is.Equal([]string{"mdr"}, outputDoc["delving_recordType"])
	is.Equal([]string{"test-org_item_123"}, outputDoc["delving_hubId"])
	is.Equal([]string{"item"}, outputDoc["delving_spec"])
	is.Equal([]string{"item/123"}, outputDoc["europeana_uri"])
	is.Equal([]string{"false"}, outputDoc["delving_hasDigitalObject"])
	is.Equal([]string{"false"}, outputDoc["delving_hasGeoHash"])
	is.Equal([]string{"false"}, outputDoc["delving_hasLandingPage"])
	is.Equal([]string{"item"}, outputDoc["europeana_collectionName"])
	is.Equal([]string{"test-org"}, outputDoc["delving_orgId"])
}

// Test backwards compatibility with the original API
func TestBackwardsCompatibility(t *testing.T) {
	is := is.New(t)

	// Test that the old API still works
	converterTypes := []ConverterType{ESE, EDM, EDMStrict, TIB, ICN, ABM, V2}

	for _, converterType := range converterTypes {
		t.Run(string(converterType), func(t *testing.T) {
			// Use the old factory function
			converter := NewConverter(converterType, "http://example.org/resource/test/123", "test-org")

			// Verify it returns the right type
			is.Equal(converterType, converter.GetConverterKey())

			// Verify it can convert something basic
			input := map[string][]string{
				"dc_title": {"Test Title"},
			}

			result := converter.Convert(input, true)
			is.Equal([]string{"Test Title"}, result["dc_title"])
		})
	}
}

// Test the organization-specific default converter registry
func TestOrgDefaultConverters(t *testing.T) {
	is := is.New(t)

	// Create temporary config files for different orgs
	tempDir := t.TempDir()

	// Config for org1 - uses ESE converter
	org1ConfigPath := filepath.Join(tempDir, "org1-config.toml")
	err := os.WriteFile(org1ConfigPath, []byte(testConfigToml), 0o644)
	is.NoErr(err)

	// Config for org2 - uses TIB converter
	org2ConfigPath := filepath.Join(tempDir, "org2-config.toml")
	err = os.WriteFile(org2ConfigPath, []byte(testConfigToml), 0o644)
	is.NoErr(err)

	// Create mock domain.V1Config for testing
	v1Config := &domain.V1Config{}
	v1Config.Convertor.Path = org1ConfigPath
	v1Config.Convertor.Type = "ese"

	v1ConfigTIB := &domain.V1Config{}
	v1ConfigTIB.Convertor.Path = org2ConfigPath
	v1ConfigTIB.Convertor.Type = "tib"

	// Register converters for each organization
	err = RegisterOrgConverter("org1", org1ConfigPath, ESE, v1Config)
	is.NoErr(err)

	err = RegisterOrgConverter("org2", org2ConfigPath, TIB, v1ConfigTIB)
	is.NoErr(err)

	// Test that we can retrieve the correct converter for each org
	org1Converter, err := DefaultConverter("http://example.org/resource/test/123", "org1")
	is.NoErr(err)
	is.Equal(ESE, org1Converter.GetConverterKey())

	org2Converter, err := DefaultConverter("http://example.org/resource/test/123", "org2")
	is.NoErr(err)
	is.Equal(TIB, org2Converter.GetConverterKey())

	// Test with an unregistered org - should use fallback
	SetDefaultFallbackType(EDM)
	unknownOrgConverter, err := DefaultConverter("http://example.org/resource/test/123", "unknown-org")
	is.NoErr(err)
	is.Equal(EDM, unknownOrgConverter.GetConverterKey())

	// Test updating converter type
	updated := UpdateOrgConverterType("org1", ABM)
	is.True(updated)

	org1UpdatedConverter, err := DefaultConverter("http://example.org/resource/test/123", "org1")
	is.NoErr(err)
	is.Equal(ABM, org1UpdatedConverter.GetConverterKey())

	// Check registered orgs
	orgs := ListRegisteredOrgs()
	is.True(len(orgs) == 2)
	is.True(contains(orgs, "org1"))
	is.True(contains(orgs, "org2"))
}

// Test that query replacers from TOML configuration are used correctly
func TestQueryReplacersFromConfig(t *testing.T) {
	is := is.New(t)

	// Get the test config
	cm := getTestConfigManager(t)

	// Create a converter with the TOML config for TIB
	converter, err := NewConverterWithConfig(
		cm,
		TIB,
		"http://example.org/resource/test/123",
		"test_org",
	)
	is.NoErr(err)

	// Test replacements using the query replacers from TOML
	// The TOML defines: query_replacers = { "europeana_" = "edm_", "tib_" = "nave_" }
	result := converter.ReplaceQueryString("europeana_type=IMAGE&tib_material=Canvas", false)
	is.Equal("edm_type=IMAGE&nave_material=Canvas", result)

	// Test a different converter type (V2)
	v2Converter, err := NewConverterWithConfig(
		cm,
		V2,
		"http://example.org/resource/test/123",
		"test_org",
	)
	is.NoErr(err)

	// Test replacements using the query replacers from TOML for V2
	// The TOML defines: query_replacers = { "europeana_" = "edm_", "delving_" = "nave_" }
	result = v2Converter.ReplaceQueryString("europeana_type=IMAGE&delving_title=Test", false)
	is.Equal("edm_type=IMAGE&nave_title=Test", result)

	// Verify that each converter type has its own specific replacers
	tibResult := converter.ReplaceQueryString("tib_material=Canvas", false)
	v2Result := v2Converter.ReplaceQueryString("tib_material=Canvas", false)

	// TIB converter should replace "tib_" with "nave_"
	is.Equal("nave_material=Canvas", tibResult)

	// V2 converter should NOT replace "tib_" since it's not in its replacers map
	is.Equal("tib_material=Canvas", v2Result)
}

// Test the query string replacement functionality
func TestQueryStringReplacement(t *testing.T) {
	is := is.New(t)

	// Create a converter with default replacers
	converter := NewConverter(EDM, "http://example.org/resource/test/123", "test-org")

	// Test forward replacements
	forwardTests := []struct {
		input    string
		expected string
	}{
		{"europeana_type=IMAGE", "edm_type=IMAGE"},
		{"europeana_creator=Picasso&europeana_title=Guernica", "edm_creator=Picasso&edm_title=Guernica"},
		// {"tib_material=Canvas", "nave_material=Canvas"},
		// {"delving_title=Test", "nave_title=Test"},
		{"http://example.org?europeana_uri=123", "http://example.org?edm_uri=123"},
	}

	for _, test := range forwardTests {
		result := converter.ReplaceQueryString(test.input, false)
		is.Equal(test.expected, result)
	}

	// Test reverse replacements
	reverseTests := []struct {
		input    string
		expected string
	}{
		{"edm_type=IMAGE", "europeana_type=IMAGE"},
		{"edm_creator=Picasso&edm_title=Guernica", "europeana_creator=Picasso&europeana_title=Guernica"},
		// Note: For "nave_" prefix, we can't predict exactly which original prefix will be chosen
		// because multiple prefixes map to "nave_". So we'll check separately.
	}

	for _, test := range reverseTests {
		result := converter.ReplaceQueryString(test.input, true)
		is.Equal(test.expected, result)
	}

	// Test nave_ prefix in reverse direction separately
	// Since it maps to multiple possible prefixes
	naveMaterialResult := converter.ReplaceQueryString("nave_material=Canvas", true)
	// Check that we replaced nave_ with some prefix
	is.True(naveMaterialResult != "nave_material=Canvas")
	// Check that the suffix part is preserved
	is.True(strings.HasSuffix(naveMaterialResult, "material=Canvas"))

	naveTitleResult := converter.ReplaceQueryString("nave_title=Test", true)
	// Check that we replaced nave_ with some prefix
	is.True(naveTitleResult != "nave_title=Test")
	// Check that the suffix part is preserved
	is.True(strings.HasSuffix(naveTitleResult, "title=Test"))

	// Test with empty string
	is.Equal("", converter.ReplaceQueryString("", false))

	// Test with a converter that has custom replacers
	// cm := getTestConfigManager(t)

	// Create a temporary config file with custom query replacers
	tempDir := t.TempDir()
	customConfigPath := filepath.Join(tempDir, "custom-config.toml")

	customConfig := `
[base]
mappings = []

[delving]
mappings = []

[converters.custom]
key = "custom"
namespace = "http://example.org/"
mappings = []
query_replacers = { "foo_" = "bar_", "test_" = "prod_" }

[defaults]
record_type = "mdr"
`

	err := os.WriteFile(customConfigPath, []byte(customConfig), 0o644)
	is.NoErr(err)

	// Load the custom config
	customCM, err := NewConfigManager(customConfigPath)
	is.NoErr(err)

	// Create a converter with custom replacers
	customConverterConfig, ok := customCM.GetConverterConfig("custom")
	is.True(ok)

	// Test that we can access the custom replacers
	is.Equal("bar_", customConverterConfig.QueryReplacers["foo_"])
	is.Equal("prod_", customConverterConfig.QueryReplacers["test_"])

	customConverter, err := NewConverterWithConfig(
		customCM,
		"custom",
		"http://example.org/resource/test/123",
		"test-org",
	)
	is.NoErr(err)

	// Test custom replacements
	is.Equal("bar_field=value", customConverter.ReplaceQueryString("foo_field=value", false))
	is.Equal("prod_env=staging", customConverter.ReplaceQueryString("test_env=staging", false))
}

// Test the ParseConverterType and RegisterWithTypeString functions
func TestConverterTypeFromString(t *testing.T) {
	is := is.New(t)

	// Test parsing known types
	is.Equal(ESE, ParseConverterType("ese"))
	is.Equal(EDM, ParseConverterType("edm"))
	is.Equal(TIB, ParseConverterType("tib"))
	is.Equal(ICN, ParseConverterType("icn"))
	is.Equal(ABM, ParseConverterType("abm"))
	is.Equal(V2, ParseConverterType("v2"))

	// Test parsing unknown type (defaults to V2)
	is.Equal(V2, ParseConverterType("unknown"))
}

// Test RegisterOrgConverterWithTypeString specifically
func TestRegisterOrgConverterWithTypeString(t *testing.T) {
	is := is.New(t)

	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")
	err := os.WriteFile(configPath, []byte(testConfigToml), 0o644)
	is.NoErr(err)

	// Create domain.V1Config objects with different converter types

	v1ConfigEDM := &domain.V1Config{}
	v1ConfigEDM.Convertor.Path = configPath
	v1ConfigEDM.Convertor.Type = "edm"

	v1ConfigUnknown := &domain.V1Config{}
	v1ConfigUnknown.Convertor.Path = configPath
	v1ConfigUnknown.Convertor.Type = "unknown"

	// Test with edm type
	err = RegisterOrgConverterWithTypeString("org3", v1ConfigEDM)
	is.NoErr(err)

	converter, err := DefaultConverter("http://example.org/resource/test/123", "org3")
	is.NoErr(err)
	is.Equal(EDM, converter.GetConverterKey())

	// Test with unknown type (should default to V2)
	err = RegisterOrgConverterWithTypeString("org4", v1ConfigUnknown)
	is.NoErr(err)

	converter, err = DefaultConverter("http://example.org/resource/test/123", "org4")
	is.NoErr(err)
	is.Equal(V2, converter.GetConverterKey())
}

// Helper function to check if a slice contains a string
func contains(s []string, e string) bool {
	return slices.Contains(s, e)
}

// Test configuration with minimal sample for test purposes
const testConfigToml = `
[base]
mappings = [
  { source = "dc_title", target = "dc_title" },
  { source = "dc_creator", target = "dc_creator" },
  { source = "europeana_type", target = "edm_type" },
  { source = "europeana_rights", target = "edm_rights" }
]

[delving]
mappings = [
  { source = "delving_thumbnail", target = "nave_thumbnail" },
  { source = "delving_year", target = "delving_year" }
]

[converters.ese]
key = "ese"
namespace = "http://www.europeana.eu/schemas/ese/"
mappings = []
query_replacers = { "europeana_" = "edm_" }

[converters.edm]
key = "edm"
namespace = "http://www.europeana.eu/schemas/edm/"
mappings = []
query_replacers = { "europeana_" = "edm_" }

[converters.edm-strict]
key = "edm-strict"
namespace = "http://www.europeana.eu/schemas/edm/"
mappings = []
query_replacers = { "europeana_" = "edm_" }

[converters.v2]
key = "v2"
namespace = ""
mappings = []
query_replacers = { "europeana_" = "edm_", "delving_" = "nave_" }

[converters.tib]
key = "tib"
namespace = "http://schemas.delving.eu/resource/ns/tib/"
mappings = [
  { source = "tib_material", target = "nave_material" },
  { source = "tib_technique", target = "nave_technique" },
  { source = "tib_color", target = "nave_color" },
  { source = "tib_objectSoort", target = "nave_objectSoort" }
]
query_replacers = { "europeana_" = "edm_", "tib_" = "nave_" }

[converters.icn]
key = "icn"
namespace = "http://www.icn.nl/schemas/icn/"
mappings = [
  { source = "icn_material", target = "nave_material" },
  { source = "icn_productionPlace", target = "nave_productionPlace" },
  { source = "icn_dimension", target = "nave_dimension" }
]
query_replacers = { "europeana_" = "edm_", "icn_" = "nave_" }

[converters.abm]
key = "abm"
namespace = "http://purl.org/abm/sen"
mappings = [
  { source = "abm_county", target = "nave_county" },
  { source = "abm_country", target = "nave_country" },
  { source = "abm_latLong", target = "nave_latLong" }
]
query_replacers = { "europeana_" = "edm_", "abm_" = "nave_" }

[defaults]
record_type = "mdr"
has_digital_object_field = "delving_hasDigitalObject"
has_geo_hash_field = "delving_hasGeoHash"
has_landing_page_field = "delving_hasLandingPage"
file_path_prefix = "file:///opt"
`
