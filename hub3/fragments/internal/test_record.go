package internal

import (
	"sort"
	"strings"

	"github.com/delving/hub3/ikuzo/rdf"
)

type BaseRecord struct {
	Type                 []string                `json:"@type,omitempty" rdf:"@types"`
	ID                   string                  `json:"@id,omitempty" rdf:"@id"`
	BaseID               string                  `json:"baseID,omitempty" rdf:"nk_baseID"`
	Context              Context                 `json:"@context,omitempty" rdf:"-"`
	BaseType             string                  `json:"baseType,omitempty" rdf:"nk_baseType"`
	EdmIsShownBy         string                  `json:"edm:isShownBy,omitempty" rdf:"edm_isShownBy"`
	EdmObject            string                  `json:"edm:object,omitempty" rdf:"edm_object"`
	CleanID              string                  `json:"cleanID,omitempty" rdf:"nk_cleanID"`
	DcIdentifier         string                  `json:"dc:identifier,omitempty" rdf:"dc_identifier"`
	DcTitle              string                  `json:"dc:title,omitempty" rdf:"dc_title"`
	RestitutionState     string                  `json:"currentRestitutionState,omitempty" rdf:"nk_currentRestitutionState"`
	RestitutionStateDate []string                `json:"currentRestitutionStateDate,omitempty" rdf:"nk_currentRestitutionDate"`
	ObjectNameFilter     []string                `json:"object_name_filter,omitempty" rdf:"object_name_filter"`
	LocationFilter       []rdf.LiteralOrResource `json:"location_filter,omitempty" rdf:"nk_location_filter"`
	EdmHasView           []HasView               `json:"edm:hasView,omitempty" rdf:"edm_hasView"`
	RestitutionCases     []Case                  `json:"sourceCases,omitempty"`
	Timeline             []Timeline              `json:"timeline,omitempty" rdf:"nk_timeline"`
	Cho                  []CHO                   `json:"cho,omitempty" rdf:"nk_cho"`
	SNKDeclaration       []SNKDeclaration        `json:"snkDeclaration,omitempty" rdf:"nk_snkDeclaration"`
	dzi                  []string
}

type Context struct {
	Base    string `json:"@base"`
	Vocab   string `json:"@vocab"`
	Rcbase  string `json:"rcbase"`
	Edm     string `json:"edm"`
	Dc      string `json:"dc"`
	Dcterms string `json:"dcterms"`
	Ore     string `json:"ore"`
	Nave    string `json:"nave"`
	Rdf     string `json:"rdf"`
	Rdfs    string `json:"rdfs"`
	Skos    string `json:"skos"`
	Cho     struct {
		Type string `json:"@type"`
	} `json:"cho"`
	Creator struct {
		Type string `json:"@type"`
	} `json:"creator"`
	Date struct {
		Type string `json:"@type"`
	} `json:"date"`
	DateStart struct {
		Type string `json:"@type"`
	} `json:"dateStart"`
	DateEnd struct {
		Type string `json:"@type"`
	} `json:"dateEnd"`
	DateRejected struct {
		Type string `json:"@type"`
	} `json:"dateRejected"`
	DateGranted struct {
		Type string `json:"@type"`
	} `json:"dateGranted"`
	Language string `json:"@language"`
}

type HasView struct {
	Type            []string `json:"@type" rdf:"@types"`
	ID              string   `json:"@id" rdf:"@id"`
	RdfLabel        string   `json:"rdf:label" rdf:"rdf_label"`
	NaveThumbSmall  string   `json:"nave:thumbSmall" rdf:"nave_thumbSmall"`
	NaveThumbLarge  string   `json:"nave:thumbLarge" rdf:"nave_thumbLarge"`
	NaveDeepZoomURI string   `json:"nave:deepZoomUri" rdf:"nave_deepZoomUri"`
}

type Timeline struct {
	TimelineType      string         `json:"timeline_type" rdf:"nk_timeline_type"`
	TimelineHeader    string         `json:"timeline_header" rdf:"nk_timeline_header"`
	TimelineValue     string         `json:"timeline_value" rdf:"nk_timeline_value"`
	TimelineCaseID    string         `json:"timeline_caseID" rdf:"nk_timeline_caseID"`
	TimelineCaseNote  string         `json:"timeline_caseNote" rdf:"nk_timeline_caseNote"`
	TimelineDate      string         `json:"timeline_date,omitempty" rdf:"nk_timeline_date"`
	TimelineOwner     string         `json:"timeline_owner" rdf:"nk_timeline_owner"`
	TimelineSortOrder int            `json:"timeline_sort_order" rdf:"nk_timeline_sort_order"`
	TimelineLocation  string         `json:"timeline_location,omitempty" rdf:"nk_timeline_location"`
	TimelineSource    TimelineSource `json:"timeline_source,omitempty" rdf:"nk_timeline_source"`
	TimelineRole      string         `json:"timeline_role,omitempty" rdf:"nk_timeline_role"`
	TimelineNotes     []string       `json:"timeline_notes,omitempty" rdf:"nk_timeline_notes"`
	TimelineStartDate string         `json:"timeline_start_date,omitempty" rdf:"nk_timeline_start_date"`
	TimelineEndDate   string         `json:"timeline_end_date,omitempty" rdf:"nk_timeline_end_data"`
	TimelineStatus    string         `json:"timeline_status,omitempty" rdf:"nk_timeline_status"`
	TimelineLink      string         `json:"timeline_link,omitempty" rdf:"nk_timeline_link"`
}

type TimelineSource []string

func (t TimelineSource) String() string {
	return strings.Join(t, ", ")
}

func (t Timeline) Contains(target string) bool {
	return strings.Contains(t.TimelineValue, target)
}

type Dimension struct {
	HeightLenght  []string `json:"heightLenght" rdf:"nk_heightLength"`
	DepthDiameter []string `json:"depthDiameter" rdf:"nk_depthDiameter"`
	Weight        []string `json:"weight" rdf:"nk_weight"`
	Width         []string `json:"width" rdf:"nk_width"`
	Type          []string `json:"@type" rdf:"@types"`
	Badge         string   `json:"badge" rdf:"nk_badge"`
}

type CHO struct {
	ObjectNumber       string                  `json:"objectNumber" rdf:"nk_objectNumber"`
	Badge              string                  `json:"badge" rdf:"nk_badge"`
	DcSubject          []string                `json:"dc:subject,omitempty" rdf:"dc_subject"`
	DcDescription      []string                `json:"dc:description" rdf:"dc_description"`
	Dimension          Dimension               `json:"dimension,omitempty" rdf:"nk_dimension"`
	Barcode            []string                `json:"barcode" rdf:"nk_barcode"`
	BarcodeLabel       []string                `json:"barcodeLabel" rdf:"nk_barcodeLabel"`
	NrOfParts          []string                `json:"nrOfParts,omitempty" rdf:"nk_nrOfParts"`
	DcTitle            []rdf.LiteralOrResource `json:"dc:title" rdf:"dc_title"`
	NaveMaterial       []string                `json:"nave:material" rdf:"nave_material"`
	NaveTechnique      []string                `json:"nave:technique" rdf:"nave_technique"`
	NaveObjectName     []string                `json:"nave:objectName,omitempty" rdf:"nave_objectName"`
	NaveObjectCategory []string                `json:"nave:objectCategory,omitempty" rdf:"nave_objectCategory"`
	UserCode           string                  `json:"user_code" rdf:"nk_userCode"`
	Creator            []Creator               `json:"creator" rdf:"nk_creator"`
	ProductionPlace    []string                `json:"productionPlace,omitempty" rdf:"nk_productionPlace"`
	ProductionDate     []ProductionDate        `json:"productionDate" rdf:"nk_productionDate"`
	Type               []string                `json:"@type" rdf:"@types"`
	Thumbnail          string                  `json:"thumbnail,omitempty" rdf:"nk_thumbnail"`
	ID                 string                  `json:"@id" rdf:"@id"`
	ObjectNameFilter   []string                `json:"objectNameFilter,omitempty" rdf:"nk_objectNameFilter"`
	RestitutionState   string                  `json:"restitutionState" rdf:"nk_restitutionState"`
	RestitutionDate    string                  `json:"restitutionDate" rdf:"nk_restitutionDate"`
}

// Creators returns a deduplicated sorted list of creators
func (cho CHO) CreatorNames() []string {
	unique := map[string]bool{}
	creators := []string{}
	for _, creator := range cho.Creator {
		for _, name := range creator.CreatorName {
			if _, ok := unique[name]; !ok {
				unique[name] = true
				creators = append(creators, name)
			}
		}
	}

	sort.Strings(creators)

	return creators
}

type ProductionDate struct {
	DateStart  []string `json:"dateStart" rdf:"nk_dateStart"`
	DateEnd    []string `json:"dateEnd" rdf:"nk_dateEnd"`
	DatePeriod []string `json:"datePeriod" rdf:"nk_datePeriod"`
	Type       []string `json:"@type" rdf:"@types"`
	Badge      string   `json:"badge" rdf:"nk_badge"`
}

type Creator struct {
	CreatorName  []string `json:"creatorName" rdf:"nk_creatorName"`
	CreationRole []string `json:"creationRole" rdf:"nk_creationRole"`
	DateOfBirth  []string `json:"dateOfBirth" rdf:"nk_dateOfBirth"`
	DateOfDeath  []string `json:"dateOfDeath" rdf:"nk_dateOfDeath"`
	Type         []string `json:"@type" rdf:"@types"`
	Badge        string   `json:"badge" rdf:"nk_badge"`
}

type Source struct {
	ID   string `json:"@id"`
	Type string `json:"@type"`
	// DctermsHasParts []struct {
	// DcIdentifier string `json:"dc:identifier"`
	// ID           string `json:"@id"`
	// Type         string `json:"@type"`
	// IsShownAt    string `json:"isShownAt"`
	// DcTitle      string `json:"dc:title"`
	// } `json:"dcterms:hasParts"`
	// RdfLabel    string `json:"rdf:label"`
	// ProviderURL string `json:"providerURL"`
	// DataSource  string `json:"dataSource"`
}

type SNKDeclaration struct {
	Name   string `json:"declarationName,omitempty" rdf:"nk_snkName"`
	Date   string `json:"declarationDate,omitempty" rdf:"nk_snkDate"`
	Source string `json:"declarationSource,omitempty" rdf:"nk_snkSource"`
	Place  string `json:"declarationPlace,omitempty" rdf:"nk_snkPlace"`
	Number string `json:"declarationNumber,omitempty" rdf:"nk_snkNumber"`
	Status string `json:"declarationStatus,omitempty" rdf:"nk_snkStatus"`
}

func (br *BaseRecord) DeepZoomImages() []string {
	if br.dzi != nil {
		return br.dzi
	}

	dzi := []string{}
	for _, view := range br.EdmHasView {
		if view.NaveDeepZoomURI != "" {
			dzi = append(dzi, view.NaveDeepZoomURI)
		}
	}

	br.dzi = dzi

	return dzi
}

type Case struct {
	ID    string `json:"caseID"`
	Title string `json:"caseTitle"`
	Link  string `json:"caseLink"`
}
