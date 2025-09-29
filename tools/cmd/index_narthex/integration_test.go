package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/rdf/formats/rdfxml"
	"github.com/delving/hub3/ikuzo/rdf/index"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

const testRDF = `<rdf:RDF xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:ebucore="urn:ebu:metadata-schema:ebuCore_2014/" xmlns:edm="http://www.europeana.eu/schemas/edm/" xmlns:nave="http://schemas.delving.eu/nave/terms/" xmlns:ore="http://www.openarchives.org/ore/terms/" xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <ore:Aggregation rdf:about="http://data.brabantcloud.nl/resource/aggregation/enb-10-beeldmateriaal/enb-10.beeldmateriaal-db129c8f-50bc-930c-5f0e-90bc80ecbe30-ab16f200-8232-11e5-b3dd-0741013467d3">
        <edm:aggregatedCHO rdf:resource="http://data.brabantcloud.nl/resource/document/enb-10-beeldmateriaal/enb-10.beeldmateriaal-db129c8f-50bc-930c-5f0e-90bc80ecbe30-ab16f200-8232-11e5-b3dd-0741013467d3"/>
        <edm:dataProvider>Heemkundekring 'De Plaets'</edm:dataProvider>
        <edm:hasView rdf:resource="http://images.memorix.nl/enb_10/thumb/500x500/77c2b3c4-5297-0605-7f59-b5f752b8ae29.jpg"/>
        <edm:isShownAt rdf:resource="http://data.brabantcloud.nl/resource/aggregation/enb-10-beeldmateriaal/enb-10.beeldmateriaal-db129c8f-50bc-930c-5f0e-90bc80ecbe30-ab16f200-8232-11e5-b3dd-0741013467d3"/>
        <edm:isShownBy rdf:resource="http://images.memorix.nl/enb_10/thumb/500x500/77c2b3c4-5297-0605-7f59-b5f752b8ae29.jpg"/>
        <edm:object rdf:resource="http://images.memorix.nl/enb_10/thumb/220x220/77c2b3c4-5297-0605-7f59-b5f752b8ae29.jpg"/>
        <edm:provider>Erfgoed Brabant</edm:provider>
        <edm:rights rdf:resource="http://www.europeana.eu/rights/rr-f/"/>
    </ore:Aggregation>
    <edm:ProvidedCHO rdf:about="http://data.brabantcloud.nl/resource/document/enb-10-beeldmateriaal/enb-10.beeldmateriaal-db129c8f-50bc-930c-5f0e-90bc80ecbe30-ab16f200-8232-11e5-b3dd-0741013467d3">
        <dc:description>Mr. Johan Marcelus van der Meer en zijn vrouw Anna Marie Bertels. Hij was burgemeester van de gemeente Berlicum van 1947 tot 1958. Hier is hij zojuist geridderd voor zijn aandeel in de wederopbouw na de oorlog van 1940-1945.</dc:description>
        <dc:format>foto</dc:format>
        <dc:identifier>1470</dc:identifier>
        <dc:subject>Burgemeester</dc:subject>
        <dc:subject>Ridderorde</dc:subject>
        <dc:subject>van der Meer</dc:subject>
        <dc:subject>Bertels</dc:subject>
        <dc:title>personen</dc:title>
        <dcterms:created>1947-1958</dcterms:created>
        <dcterms:spatial rdf:resource="http://sws.geonames.org/2759059"/>
        <edm:type>IMAGE</edm:type>
    </edm:ProvidedCHO>
    <edm:WebResource rdf:about="http://images.memorix.nl/enb_10/thumb/500x500/77c2b3c4-5297-0605-7f59-b5f752b8ae29.jpg">
        <ebucore:hasMimeType>image/jpeg</ebucore:hasMimeType>
        <nave:deepZoomUrl>http://images.memorix.nl/enb_10/deepzoom/77c2b3c4-5297-0605-7f59-b5f752b8ae29.dzi</nave:deepZoomUrl>
        <nave:resourceSortOrder>1</nave:resourceSortOrder>
        <nave:thumbSmall>http://images.memorix.nl/enb_10/thumb/220x220/77c2b3c4-5297-0605-7f59-b5f752b8ae29.jpg</nave:thumbSmall>
        <nave:thumbLarge>http://images.memorix.nl/enb_10/thumb/500x500/77c2b3c4-5297-0605-7f59-b5f752b8ae29.jpg</nave:thumbLarge>
    </edm:WebResource>
    <edm:Place rdf:about="http://sws.geonames.org/2759059">
        <nave:country>Nederland</nave:country>
        <nave:province>Gemeente Sint-Michielsgestel</nave:province>
        <nave:municipality>Gemeente Sint-Michielsgestel</nave:municipality>
        <nave:city>Berlicum</nave:city>
    </edm:Place>
    <nave:BrabantCloudResource>
        <nave:collection>Heemkundekring 'De Plaets'</nave:collection>
        <nave:collectionType>Beeldmateriaal</nave:collectionType>
        <nave:collectionPart>personen</nave:collectionPart>
        <nave:date>1947-1958</nave:date>
        <nave:dimensionNote>5044-14</nave:dimensionNote>
        <nave:place>Berlicum</nave:place>
        <nave:productionEnd>1958</nave:productionEnd>
        <nave:productionPeriod>1947-1958</nave:productionPeriod>
        <nave:productionStart>1947</nave:productionStart>
    </nave:BrabantCloudResource>
    <nave:DelvingResource>
        <nave:featured>false</nave:featured>
        <nave:allowDeepZoom>true</nave:allowDeepZoom>
        <nave:allowLinkedOpenData>true</nave:allowLinkedOpenData>
        <nave:allowSourceDownload>false</nave:allowSourceDownload>
        <nave:public>true</nave:public>
    </nave:DelvingResource>
</rdf:RDF>`

// processLegacyBulkAPI processes RDF through the legacy bulk API system
func processLegacyBulkAPI(rdfData, hubID string) (map[string]interface{}, error) {
	// Create FragmentGraph and FragmentBuilder as done in bulk API
	fg := fragments.NewFragmentGraph()
	
	// Set meta information like bulk API does
	id, err := domain.NewHubID(hubID)
	if err != nil {
		return nil, fmt.Errorf("unable to create hubID: %w", err)
	}
	
	fg.Meta.OrgID = id.OrgID().String()
	fg.Meta.HubID = id.String()
	fg.Meta.Spec = id.DatasetID.ID
	fg.Meta.Revision = 1
	fg.Meta.NamedGraphURI = fmt.Sprintf("http://data.example.org/resource/%s/graph", id.String())
	fg.Meta.Modified = fragments.NowInMillis()
	fg.Meta.Tags = []string{"narthex"}
	// Set AboutTypeURI to Aggregation as the root type
	fg.Meta.AboutTypeURI = []string{"http://www.openarchives.org/ore/terms/Aggregation"}
	
	fb := fragments.NewFragmentBuilder(fg)
	
	// Parse RDF data like bulk API does
	err = fb.ParseResolvedGraph(strings.NewReader(rdfData), "application/rdf+xml")
	if err != nil {
		return nil, fmt.Errorf("unable to parse RDF: %w", err)
	}
	
	// Create the document as bulk API does
	_, err = fb.ResourceMap()
	if err != nil {
		return nil, fmt.Errorf("unable to build resource map: %w", err)
	}
	
	doc, err := fb.Doc()
	if err != nil {
		return nil, fmt.Errorf("unable to build doc: %w", err)
	}
	
	// Get IndexMessage as bulk API does
	msg, err := doc.IndexMessage()
	if err != nil {
		return nil, fmt.Errorf("unable to create index message: %w", err)
	}
	
	// Parse the source JSON
	var result map[string]interface{}
	if err := json.Unmarshal(msg.Source, &result); err != nil {
		return nil, fmt.Errorf("unable to unmarshal legacy result: %w", err)
	}
	
	return result, nil
}

// processNewIndexSystem processes RDF through the new index system
func processNewIndexSystem(rdfData, hubID string) (map[string]interface{}, error) {
	// Parse RDF like new system does
	rdfGraph, err := rdfxml.Parse(strings.NewReader(rdfData), nil, hubID)
	if err != nil {
		return nil, fmt.Errorf("unable to parse RDF: %w", err)
	}
	
	// Create header like new system does
	s := strings.TrimSuffix(strings.TrimPrefix(hubID, "urn:"), "/graph")
	id, err := domain.NewHubID(s)
	if err != nil {
		return nil, fmt.Errorf("unable to create hubID: %w", err)
	}
	
	// Use the same aboutTypes as the legacy system to find the root
	aboutTypes := []string{"http://www.openarchives.org/ore/terms/Aggregation"}
	entryURI, err := rdfGraph.GetAboutURI(aboutTypes)
	if err != nil {
		// Fallback to SubjectSafe if no aboutType found
		entryURI = rdfGraph.SubjectSafe().RawValue()
	}
	
	header := index.Header{
		OrgID:    id.OrgID().String(),
		Spec:     id.DatasetID.ID,
		EntryURI: entryURI,
		Tags:     []string{"narthex"},
		HubID:    id.String(),
	}
	
	// Create index graph like new system does
	fg, err := index.NewGraph(header)
	if err != nil {
		return nil, fmt.Errorf("unable to create graph: %w", err)
	}
	
	if err := fg.AddGraph(rdfGraph); err != nil {
		return nil, fmt.Errorf("unable to add graph: %w", err)
	}
	
	// Create IndexMessage like new system does
	msg, err := fg.IndexMessage(nil) // No TagMap for now
	if err != nil {
		return nil, fmt.Errorf("unable to create index message: %w", err)
	}
	
	// Parse the source JSON
	var result map[string]interface{}
	if err := json.Unmarshal(msg.Source, &result); err != nil {
		return nil, fmt.Errorf("unable to unmarshal new result: %w", err)
	}
	
	return result, nil
}

// TestLegacyVsNewRDFProcessing compares legacy and new RDF processing outputs
func TestLegacyVsNewRDFProcessing(t *testing.T) {
	// Set up test logging
	logHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelWarn, // Reduce noise for detailed comparison
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)
	
	hubID := "test-org_test-dataset_test-record"
	
	t.Run("basic comparison", func(t *testing.T) {
		// Process through legacy system
		legacyResult, err := processLegacyBulkAPI(testRDF, hubID)
		if err != nil {
			t.Fatalf("Legacy processing failed: %v", err)
		}
		
		// Process through new system
		newResult, err := processNewIndexSystem(testRDF, hubID)
		if err != nil {
			t.Fatalf("New processing failed: %v", err)
		}
		
		// Compare results with detailed diff
		compareResultsDetailed(t, legacyResult, newResult)
	})
}

// compareResultsDetailed uses google/go-cmp for detailed comparison
func compareResultsDetailed(t *testing.T, legacy, new map[string]interface{}) {
	t.Helper()
	
	// Skip printing full results to focus on differences
	
	// Use cmp.Diff for detailed comparison
	diff := cmp.Diff(legacy, new, 
		cmpopts.SortSlices(func(a, b string) bool { return a < b }), // Sort string slices
		cmpopts.SortMaps(func(a, b string) bool { return a < b }),   // Sort map keys
		cmp.FilterPath(func(p cmp.Path) bool {
			// Skip timestamp fields that will always be different
			return strings.Contains(p.String(), ".modified")
		}, cmp.Ignore()),
	)
	
	if diff == "" {
		t.Log("✅ Results are identical!")
		return
	}
	
	// Results differ - let's analyze the differences
	t.Errorf("❌ Results differ:\n%s", diff)
	
	// Also provide a summary of differences
	analyzeDifferences(t, diff)
}

// analyzeDifferences provides a summary of what types of differences we found
func analyzeDifferences(t *testing.T, diff string) {
	t.Helper()
	
	lines := strings.Split(diff, "\n")
	var summaryCategories []string
	
	// Categorize the types of differences
	hasMetaDiff := false
	hasResourceDiff := false
	hasContextDiff := false
	hasFieldDiff := false
	hasEntryDiff := false
	
	for _, line := range lines {
		if strings.Contains(line, "meta.") {
			hasMetaDiff = true
		}
		if strings.Contains(line, "resources[") {
			hasResourceDiff = true
		}
		if strings.Contains(line, "context") || strings.Contains(line, "Context") {
			hasContextDiff = true
		}
		if strings.Contains(line, "fields.") {
			hasFieldDiff = true
		}
		if strings.Contains(line, "entries[") {
			hasEntryDiff = true
		}
	}
	
	t.Log("\n📊 Difference Summary:")
	if hasMetaDiff {
		summaryCategories = append(summaryCategories, "Meta/Header differences")
	}
	if hasResourceDiff {
		summaryCategories = append(summaryCategories, "Resource structure differences")
	}
	if hasContextDiff {
		summaryCategories = append(summaryCategories, "Context/hierarchy differences")
	}
	if hasFieldDiff {
		summaryCategories = append(summaryCategories, "Fields aggregation differences")
	}
	if hasEntryDiff {
		summaryCategories = append(summaryCategories, "Entry count/content differences")
	}
	
	for i, category := range summaryCategories {
		t.Logf("  %d. %s", i+1, category)
	}
	
	// Determine criticality
	criticalDiffs := hasMetaDiff || hasResourceDiff || hasContextDiff
	if criticalDiffs {
		t.Log("\n🔴 ASSESSMENT: These differences may affect functionality")
	} else {
		t.Log("\n🟡 ASSESSMENT: These differences appear to be minor (entry counts, field variations)")
	}
}

// compareResults performs detailed comparison of legacy vs new outputs
func compareResults(t *testing.T, legacy, new map[string]interface{}) {
	t.Helper()
	
	// Print both results for debugging
	legacyJSON, _ := json.MarshalIndent(legacy, "", "  ")
	newJSON, _ := json.MarshalIndent(new, "", "  ")
	
	t.Logf("Legacy result:\n%s", string(legacyJSON))
	t.Logf("New result:\n%s", string(newJSON))
	
	// Compare high-level structure
	compareMeta(t, legacy, new)
	compareResources(t, legacy, new)
	compareFields(t, legacy, new)
}

// compareMeta compares the meta/header sections
func compareMeta(t *testing.T, legacy, new map[string]interface{}) {
	t.Helper()
	
	legacyMeta, legacyOk := legacy["meta"].(map[string]interface{})
	newMeta, newOk := new["meta"].(map[string]interface{})
	
	if !legacyOk {
		t.Error("Legacy result missing meta section")
		return
	}
	if !newOk {
		t.Error("New result missing meta section")
		return
	}
	
	// Compare key meta fields
	compareField(t, "orgID", legacyMeta, newMeta)
	compareField(t, "spec", legacyMeta, newMeta)
	compareField(t, "hubID", legacyMeta, newMeta)
	compareField(t, "entryURI", legacyMeta, newMeta)
	
	// Tags comparison
	legacyTags, _ := legacyMeta["tags"].([]interface{})
	newTags, _ := newMeta["tags"].([]interface{})
	
	if !reflect.DeepEqual(legacyTags, newTags) {
		t.Errorf("Tags mismatch: legacy=%v, new=%v", legacyTags, newTags)
	}
}

// compareResources compares the resources sections
func compareResources(t *testing.T, legacy, new map[string]interface{}) {
	t.Helper()
	
	legacyResources, legacyOk := legacy["resources"].([]interface{})
	newResources, newOk := new["resources"].([]interface{})
	
	if !legacyOk {
		t.Error("Legacy result missing resources section")
		return
	}
	if !newOk {
		t.Error("New result missing resources section")
		return
	}
	
	t.Logf("Resources count: legacy=%d, new=%d", len(legacyResources), len(newResources))
	
	if len(legacyResources) != len(newResources) {
		t.Errorf("Resources count mismatch: legacy=%d, new=%d", len(legacyResources), len(newResources))
	}
	
	// Create maps by ID for easier comparison
	legacyResourceMap := make(map[string]interface{})
	newResourceMap := make(map[string]interface{})
	
	for _, r := range legacyResources {
		if res, ok := r.(map[string]interface{}); ok {
			if id, ok := res["id"].(string); ok {
				legacyResourceMap[id] = res
			}
		}
	}
	
	for _, r := range newResources {
		if res, ok := r.(map[string]interface{}); ok {
			if id, ok := res["id"].(string); ok {
				newResourceMap[id] = res
			}
		}
	}
	
	// Compare each resource
	for id, legacyRes := range legacyResourceMap {
		newRes, exists := newResourceMap[id]
		if !exists {
			t.Errorf("Resource %s missing in new result", id)
			continue
		}
		
		compareResource(t, id, legacyRes.(map[string]interface{}), newRes.(map[string]interface{}))
	}
	
	// Check for extra resources in new result
	for id := range newResourceMap {
		if _, exists := legacyResourceMap[id]; !exists {
			t.Errorf("Extra resource %s in new result", id)
		}
	}
}

// compareResource compares individual resources
func compareResource(t *testing.T, id string, legacy, new map[string]interface{}) {
	t.Helper()
	
	// Compare types
	legacyTypes, _ := legacy["types"].([]interface{})
	newTypes, _ := new["types"].([]interface{})
	
	if !reflect.DeepEqual(legacyTypes, newTypes) {
		t.Errorf("Resource %s types mismatch: legacy=%v, new=%v", id, legacyTypes, newTypes)
	}
	
	// Compare entries
	legacyEntries, _ := legacy["entries"].([]interface{})
	newEntries, _ := new["entries"].([]interface{})
	
	if len(legacyEntries) != len(newEntries) {
		t.Errorf("Resource %s entries count mismatch: legacy=%d, new=%d", id, len(legacyEntries), len(newEntries))
	}
	
	// Compare context
	legacyContext, _ := legacy["context"].([]interface{})
	newContext, _ := new["context"].([]interface{})
	
	if len(legacyContext) != len(newContext) {
		t.Errorf("Resource %s context count mismatch: legacy=%d, new=%d", id, len(legacyContext), len(newContext))
	}
	
	// Compare external context
	legacyExtContext, _ := legacy["graphExternalContext"].([]interface{})
	newExtContext, _ := new["graphExternalContext"].([]interface{})
	
	if len(legacyExtContext) != len(newExtContext) {
		t.Errorf("Resource %s external context count mismatch: legacy=%d, new=%d", id, len(legacyExtContext), len(newExtContext))
	}
}

// compareFields compares the fields sections
func compareFields(t *testing.T, legacy, new map[string]interface{}) {
	t.Helper()
	
	legacyFields, legacyOk := legacy["fields"].(map[string]interface{})
	newFields, newOk := new["fields"].(map[string]interface{})
	
	if !legacyOk && !newOk {
		return // Both missing fields, that's ok
	}
	
	if legacyOk != newOk {
		t.Errorf("Fields presence mismatch: legacy has fields=%v, new has fields=%v", legacyOk, newOk)
		return
	}
	
	// Compare field keys
	legacyKeys := make([]string, 0, len(legacyFields))
	newKeys := make([]string, 0, len(newFields))
	
	for k := range legacyFields {
		legacyKeys = append(legacyKeys, k)
	}
	for k := range newFields {
		newKeys = append(newKeys, k)
	}
	
	if !cmp.Equal(legacyKeys, newKeys, cmpopts.SortSlices(func(a, b string) bool { return a < b })) {
		t.Errorf("Fields keys mismatch:\nlegacy: %v\nnew: %v", legacyKeys, newKeys)
	}
	
	// Compare field values
	for key, legacyVal := range legacyFields {
		newVal, exists := newFields[key]
		if !exists {
			t.Errorf("Field %s missing in new result", key)
			continue
		}
		
		if !reflect.DeepEqual(legacyVal, newVal) {
			t.Errorf("Field %s value mismatch: legacy=%v, new=%v", key, legacyVal, newVal)
		}
	}
}

// compareField compares a specific field between two maps
func compareField(t *testing.T, field string, legacy, new map[string]interface{}) {
	t.Helper()
	
	legacyVal, legacyOk := legacy[field]
	newVal, newOk := new[field]
	
	if legacyOk != newOk {
		t.Errorf("Field %s presence mismatch: legacy=%v, new=%v", field, legacyOk, newOk)
		return
	}
	
	if legacyOk && !reflect.DeepEqual(legacyVal, newVal) {
		t.Errorf("Field %s value mismatch: legacy=%v, new=%v", field, legacyVal, newVal)
	}
}