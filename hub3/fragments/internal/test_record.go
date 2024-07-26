package internal

import (
	"fmt"
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
	ObjectNameFilter     []rdf.LiteralOrResource `json:"object_name_filter,omitempty" rdf:"object_name_filter"`
	CreatorFilter        []rdf.LiteralOrResource `json:"creator_filter,omitempty" rdf:"nk_creator_filter"`
	EdmHasView           []HasView               `json:"edm:hasView,omitempty" rdf:"edm_hasView"`
	RestitutionCases     []Case                  `json:"sourceCases,omitempty"`
	Timeline             []Timeline              `json:"timeline,omitempty" rdf:"nk_timeline"`
	Cho                  []CHO                   `json:"cho,omitempty" rdf:"nk_cho"`
	SNKDeclaration       []SNKDeclaration        `json:"snkDeclaration,omitempty" rdf:"nk_snkDeclaration"`
	SNKRegistration      []SNKRegistration       `rdf:"nk_snkRegistration"`
	ResearchInformation  ResearchInformation     `rdf:"nk_researchInformation"`
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
	Type            []string              `json:"@type" rdf:"@types"`
	ID              string                `json:"@id" rdf:"@id"`
	RdfLabel        rdf.LiteralOrResource `json:"rdf:label" rdf:"rdf_label"`
	NaveThumbSmall  string                `json:"nave:thumbSmall" rdf:"nave_thumbSmall"`
	NaveThumbLarge  string                `json:"nave:thumbLarge" rdf:"nave_thumbLarge"`
	NaveDeepZoomURI string                `json:"nave:deepZoomUri" rdf:"nave_deepZoomUri"`
}

type Timeline struct {
	TimelineType      rdf.LiteralOrResource `json:"timeline_type" rdf:"nk_timeline_type"`
	TimelineHeader    rdf.LiteralOrResource `json:"timeline_header" rdf:"nk_timeline_header"`
	TimelineValue     rdf.LiteralOrResource `json:"timeline_value" rdf:"nk_timeline_value"`
	TimelineCaseID    rdf.LiteralOrResource `json:"timeline_caseID" rdf:"nk_timeline_caseID"`
	TimelineCaseNote  rdf.LiteralOrResource `json:"timeline_caseNote" rdf:"nk_timeline_caseNote"`
	TimelineDate      rdf.LiteralOrResource `json:"timeline_date,omitempty" rdf:"nk_timeline_date"`
	TimelineOwner     rdf.LiteralOrResource `json:"timeline_owner" rdf:"nk_timeline_owner"`
	TimelineSortOrder int                   `json:"timeline_sort_order" rdf:"nk_timeline_sort_order"`
	TimelineLocation  rdf.LiteralOrResource `json:"timeline_location,omitempty" rdf:"nk_timeline_location"`
	TimelineSource    TimelineSource        `json:"timeline_source,omitempty" rdf:"nk_timeline_source"`
	TimelineRole      rdf.LiteralOrResource `json:"timeline_role,omitempty" rdf:"nk_timeline_role"`
	TimelineNotes     []string              `json:"timeline_notes,omitempty" rdf:"nk_timeline_notes"`
	TimelineStartDate rdf.LiteralOrResource `json:"timeline_start_date,omitempty" rdf:"nk_timeline_start_date"`
	TimelineEndDate   rdf.LiteralOrResource `json:"timeline_end_date,omitempty" rdf:"nk_timeline_end_date"`
	TimelineStatus    rdf.LiteralOrResource `json:"timeline_status,omitempty" rdf:"nk_timeline_status"`
	TimelineLink      rdf.LiteralOrResource `json:"timeline_link,omitempty" rdf:"nk_timeline_link"`
}

type TimelineSource []string

func (t TimelineSource) String() string {
	return strings.Join(t, ", ")
}

func (t Timeline) Contains(target string) bool {
	return strings.Contains(t.TimelineValue.Value, target)
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
	DcSubject          []rdf.LiteralOrResource `json:"dc:subject,omitempty" rdf:"dc_subject"`
	DcDescription      []rdf.LiteralOrResource `json:"dc:description" rdf:"dc_description"`
	Dimension          Dimension               `json:"dimension,omitempty" rdf:"nk_dimension"`
	Barcode            []rdf.LiteralOrResource `json:"barcode" rdf:"nk_barcode"`
	BarcodeLabel       []rdf.LiteralOrResource `json:"barcodeLabel" rdf:"nk_barcodeLabel"`
	NrOfParts          []rdf.LiteralOrResource `json:"nrOfParts,omitempty" rdf:"nk_nrOfParts"`
	DcTitle            []rdf.LiteralOrResource `json:"dc:title" rdf:"dc_title"`
	NaveMaterial       []rdf.LiteralOrResource `json:"nave:material" rdf:"nave_material"`
	NaveTechnique      []rdf.LiteralOrResource `json:"nave:technique" rdf:"nave_technique"`
	NaveObjectName     []rdf.LiteralOrResource `json:"nave:objectName,omitempty" rdf:"nave_objectName"`
	NaveObjectCategory []rdf.LiteralOrResource `json:"nave:objectCategory,omitempty" rdf:"nave_objectCategory"`
	UserCode           rdf.LiteralOrResource   `json:"user_code" rdf:"nk_userCode"`
	Creator            []Creator               `json:"creator" rdf:"nk_creator"`
	ProductionPlace    []rdf.LiteralOrResource `json:"productionPlace,omitempty" rdf:"nk_productionPlace"`
	ProductionDate     []ProductionDate        `json:"productionDate" rdf:"nk_productionDate"`
	Type               []string                `json:"@type" rdf:"@types"`
	Thumbnail          rdf.LiteralOrResource   `json:"thumbnail,omitempty" rdf:"nk_thumbnail"`
	ID                 string                  `json:"@id" rdf:"@id"`
	ObjectNameFilter   []rdf.LiteralOrResource `json:"objectNameFilter,omitempty" rdf:"nk_objectNameFilter"`
	RestitutionState   rdf.LiteralOrResource   `json:"restitutionState" rdf:"nk_restitutionState"`
	RestitutionDate    rdf.LiteralOrResource   `json:"restitutionDate" rdf:"nk_restitutionDate"`
}

func (cho CHO) CreatorLinks() []string {
	unique := map[string]bool{}
	creators := []string{}
	for _, creator := range cho.Creator {
		for _, name := range creator.CreatorName {
			if _, ok := unique[name.String()]; !ok {
				unique[name.String()] = true
				if len(creator.RKDArtistLink) > 0 {
					name.Value = fmt.Sprintf(`<a href="%s">%s</a>`, creator.RKDArtistLink[0].String(), name.String())
				}
				creators = append(creators, name.String())
			}
		}
	}

	sort.Strings(creators)

	return creators
}

// Creators returns a deduplicated sorted list of creators
func (cho CHO) CreatorNames() []string {
	unique := map[string]bool{}
	creators := []string{}
	for _, creator := range cho.Creator {
		for _, name := range creator.CreatorName {
			if _, ok := unique[name.Value]; !ok {
				unique[name.Value] = true
				creators = append(creators, name.Value)
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
	CreatorName   []rdf.LiteralOrResource `json:"creatorName" rdf:"nk_creatorName"`
	CreationRole  []rdf.LiteralOrResource `json:"creationRole" rdf:"nk_creationRole"`
	DateOfBirth   []rdf.LiteralOrResource `json:"dateOfBirth" rdf:"nk_dateOfBirth"`
	DateOfDeath   []rdf.LiteralOrResource `json:"dateOfDeath" rdf:"nk_dateOfDeath"`
	Type          []string                `json:"@type" rdf:"@types"`
	Badge         string                  `json:"badge" rdf:"nk_badge"`
	RKDArtistLink []rdf.LiteralOrResource `rdf:"nk_rkdArtistLink"`
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
	Name     rdf.LiteralOrResource   `json:"declarationName,omitempty" rdf:"nk_snkName"`
	Date     rdf.LiteralOrResource   `json:"declarationDate,omitempty" rdf:"nk_snkDate"`
	Source   rdf.LiteralOrResource   `json:"declarationSource,omitempty" rdf:"nk_snkSource"`
	Place    rdf.LiteralOrResource   `json:"declarationPlace,omitempty" rdf:"nk_snkPlace"`
	Number   rdf.LiteralOrResource   `json:"declarationNumber,omitempty" rdf:"nk_snkNumber"`
	Status   rdf.LiteralOrResource   `json:"declarationStatus,omitempty" rdf:"nk_snkStatus"`
	Claimant []rdf.LiteralOrResource `rdf:"nk_snkClaimant"`
}

type SNKRegistration struct {
	ID     string                `rdf:"@id"`
	Type   []string              `rdf:"@types"`
	Term   rdf.LiteralOrResource `rdf:"nk_snkRegistrationTerm"`
	Number rdf.LiteralOrResource `rdf:"nk_snkRegistrationNumber"`
	VV     rdf.LiteralOrResource `rdf:"nk_snkRegistrationVV"`
}

type ResearchInformation struct {
	Conclusion rdf.LiteralOrResource `rdf:"nk_researchConclusion"`
	Remark     rdf.LiteralOrResource `rdf:"nk_researchRemark"`
}

func (ri *ResearchInformation) IsEmpty() bool {
	return ri.Conclusion.Value == ""
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
	ID    string                `json:"caseID"`
	Title rdf.LiteralOrResource `json:"caseTitle"`
	Link  rdf.LiteralOrResource `json:"caseLink"`
}
