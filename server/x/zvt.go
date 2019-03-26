package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	c "github.com/delving/hub3/config"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/labstack/gommon/log"
)

// SparqlResource is a struct for the Search routes
type ZVTResource struct{}

// Routes returns the chi.Router
func (rs ZVTResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/search/json", gafApeProxy)
	r.Post("/search/descendants/*", gafApeProxy)
	r.Post("/search/descendantsWithAncestors/*", gafApeProxy)
	r.Post("/search/children/*", gafApeProxy)
	r.Post("/search/ead/*", gafApeProxy)
	r.Post("/urlrewrite/getapeid", gafApeProxy)
	// todo enable later again
	//r.Get("/api/search/v1/hub", getScrollResult)
	//r.Get("/api/search/v1/tree/{spec}/desc", treeDescription)
	//r.Get("/api//search/v1/tree/{spec}/desc", treeDescription)
	//r.Get("/api//search/v1/tree/{spec}", treeList)
	//r.Get("/api/search/v1/tree/{spec}", treeList)
	//r.Get("/api//search/v1/tree/{spec}/{nodeID:.*$}", treeList)
	return r
}

var labelsEN = map[string]interface{}{
	"lang":                          "en",
	"label":                         "Inventories",
	"search":                        "Search",
	"emptyFieldMessage":             "Field is empty",
	"searchPlaceHolder":             "Search by keyword, archive, inventory or index id.",
	"currentSearchResultsLabel":     "Search results:",
	"totalSearchResultsLabel":       "%amount% results for:",
	"noSearchResultsLabel":          "No results found.",
	"showDescriptiveUnitButtonText": "Show number of descriptive units:",
	"showEADButtonText":             "Show EAD",
	"next":                          "Next",
	"last":                          "Last",
	"previous":                      "Previous",
	"first":                         "First",
	"resultsPerPage":                "Results per page",
	"filter":                        "Filter",
	"searchresulttitles": map[string]string{
		"id":             "Id",
		"fondsUnitId":    "Archival Inventory",
		"unitTitle":      "Archival Inventory name",
		"fondsUnitTitle": "Archive name",
		"scopeContent":   "Description",
		"unitDate":       "Period",
		"repository":     "RHC",
		"unitId":         "Inventory",
	},
	"facetFieldTitles": map[string]string{
		"country":           "Country:",
		"subject":           "Theme:",
		"repository":        "Archival Institute:",
		"docType":           "Document Type:",
		"level":             "Level:",
		"hasDigitalObject":  "Has digital object:",
		"digitalObjectType": "Digital Object Type:",
		"unitDateType":      "Date Type:",
		"fromDate":          "Startdate:",
		"toDate":            "Enddate:",
	},
	"facetFieldNames": map[string]string{
		"NETHERLANDS":                "Netherlands",
		"GERMANY":                    "Germany",
		"fa":                         "Finding Aid",
		"hg":                         "Archival Overview",
		"clevel":                     "Other description",
		"archdesc":                   "Archive description",
		"false":                      "No digital object",
		"true":                       "Has digital object",
		"UNSPECIFIED":                "Unknown type",
		"IMAGE":                      "Image",
		"TEXT":                       "Text",
		"normal":                     "Full date",
		"nodate":                     "No date specified",
		"otherdate":                  "Textual date",
		"german.democratic.republic": "DDR (German Democratic Republic )",
		"arts":                       "Arts",
		"germany.sed.fdgb":           "DDR members and unions",
	},
}

var labelsNL = map[string]interface{}{
	"lang":                          "nl",
	"label":                         "Inventarissen",
	"search":                        "Zoeken",
	"emptyFieldMessage":             "Veld is leeg",
	"searchPlaceHolder":             "Zoek op trefwoord, archiefinventaris, inventarisnummer of index.",
	"currentSearchResultsLabel":     "Zoekresultaten:",
	"totalSearchResultsLabel":       "%amount% resultaten voor:",
	"noSearchResultsLabel":          "Geen resultaten gevonden.",
	"showDescriptiveUnitButtonText": "Toon aantal gevonden inventarissen:",
	"showEADButtonText":             "Toon EAD",
	"next":                          "Volgende",
	"last":                          "Laatste",
	"previous":                      "Vorige",
	"first":                         "Eerste",
	"resultsPerPage":                "Resultaten per pagina",
	"filter":                        "Filter",
	"searchresulttitles": map[string]interface{}{
		"id":             "Id",
		"fondsUnitId":    "Archiefinventaris",
		"unitTitle":      "Archiefinventarisnaam",
		"fondsUnitTitle": "Archiefnaam",
		"scopeContent":   "Beschrijving",
		"unitDate":       "Periode",
		"repository":     "RHC",
		"unitId":         "Inventaris",
	},
	"facetFieldTitles": map[string]interface{}{
		"country":           "Land:",
		"subject":           "Thema:",
		"repository":        "Archiefinstelling:",
		"level":             "Niveau:",
		"hasDigitalObject":  "Bevat digitale representatie:",
		"digitalObjectType": "Digitaal Object Type:",
		"unitDateType":      "Datum Typen:",
		"fromDate":          "Startdatum:",
		"toDate":            "Einddatum:",
	},
	"facetFieldNames": map[string]interface{}{
		"NETHERLANDS":                "Nederland",
		"GERMANY":                    "Duitsland",
		"fa":                         "Toegang",
		"hg":                         "Archievenoverzicht",
		"clevel":                     "Andere beschrijvingen",
		"archdesc":                   "Archiefbeschrijving",
		"false":                      "Geen digitale representatie",
		"true":                       "Bevat digitale representatie",
		"UNSPECIFIED":                "Onbekend type",
		"IMAGE":                      "Afbeelding",
		"TEXT":                       "Tekst",
		"normal":                     "Volledige datum",
		"nodate":                     "Geen datum gespecificeerd",
		"otherdate":                  "Alleen tekstuele datum",
		"german.democratic.republic": "DDR (Duitse Democratische Republiek )",
		"arts":                       "Kunsten",
		"germany.sed.fdgb":           "DDR Partijen en vakbonden",
	},
}

// GafSearchResponse returns the APE search result format
type GafSearchResponse struct {
	TotalResults     int                    `json:"totalResults"`
	StartIndex       int                    `json:"startIndex"`
	TotalPages       int                    `json:"totalPages"`
	TotalDocs        int                    `json:"totalDocs"`
	EadDocList       []string               `json:"eadDocList"`
	FacetDateFields  map[string][]GafFacet  `json:"facetDateFields"`
	FacetFields      map[string][]GafFacet  `json:"facetFields"`
	TranslatedLabels map[string]interface{} `json:"translatedLabels"`
}

// EadDoc holds the information for each individual Archive returned
type EadDoc struct {
	ID                                  string `json:"id"`
	FindingAidTitle                     string `json:"findingAidTitle"`
	NumberOfSearchResults               int    `json:"numberOfSearchResults"`
	Repository                          string `json:"repository"`
	Country                             string `json:"country"`
	Language                            string `json:"language"`
	RepositoryCode                      string `json:"repositoryCode"`
	FindingAidNo                        string `json:"findingAidNo"`
	UnitDate                            string `json:"unitDate"`
	ScopeContent                        string `json:"scopeContent"`
	NumberOfDigitalObjects              int    `json:"numberOfDigitalObjects"`
	numberOfDigitalObjectsInDescendents int    `json:"numberOfDigitalObjectsInDescendents"`
	numberOfDescendents                 int    `json:"numberOfDescendents"`
	documentUrl                         string `json:"documentUrl"`
}

// GafFacet is a generic structure to hold information about facets
type GafFacet struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Frequency int    `json:"frequency"`
}

// GafSearchRequest holds the body of the APE search request
type GafSearchRequest struct {
	Path                 string `json:"path"`
	SearchTerm           string `json:"searchTerm"`
	ResultsPerPage       int    `json:"resultsPerPage"`
	Language             string `json:"language"`
	Page                 int    `json:"page"`
	StartIndex           int    `json:"startIndex"`
	FeatureToggle        string `json:"featureToggle"`
	GafFacetQueryFilters `json:"gafFacetQueryFilters"`
}

type GafFacetQueryFilters struct {
	FromDate         string `json:"fromDate"`
	ToDate           string `json:"toDate"`
	Repository       string `json:"repository"`
	UnitDateType     string `json:"unitDateType"`
	HasDigitalObject string `json:"hasDigitalObject"`
}

func gafApeProxy(w http.ResponseWriter, r *http.Request) {
	resp, statusCode, contentType, err := runGafApeQuery(r)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, err.Error())
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}
	render.Status(r, statusCode)
	return
}

// runGafApeQuery sends a gaf query to the gaf-endpoint specified in the configuration
func runGafApeQuery(r *http.Request) (body []byte, statusCode int, contentType string, err error) {
	gafBaseURL := c.Config.EAD.SearchURL
	if gafBaseURL == "" {
		err = fmt.Errorf("ead proxy url is not configured")
		return
	}
	fullPath := fmt.Sprintf("%s%s", gafBaseURL, r.URL.Path)
	log.Printf("path %#v", fullPath)
	req, err := http.NewRequest("POST", fullPath, r.Body)
	if err != nil {
		log.Errorf("Unable to create gaf request %s", err)
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := netClient.Do(req)
	if err != nil {
		log.Errorf("Error in gaf query: %s", err)
		return
	}
	body, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Errorf("Unable to read the response body with error: %s", err)
		return
	}
	statusCode = resp.StatusCode
	contentType = resp.Header.Get("Content-Type")
	return
}
