// Code generated by protoc-gen-go. DO NOT EDIT.
// source: hub3/api/api.proto

/*
Package api is a generated protocol buffer package.

It is generated from these files:
	hub3/api/api.proto

It has these top-level messages:
	FilterValue
	SearchRequest
	DetailRequest
	BreadCrumb
	PaginationLink
	Pagination
	Query
	Facet
	FaceLink
	MetadataFieldV1
	MetadataItemV1
	SearchResultWrapperV1
	SearchResultV1
	DetailResult
*/
package api

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type ResponseFormatType int32

const (
	ResponseFormatType_PROTOBUF ResponseFormatType = 0
	ResponseFormatType_JSON     ResponseFormatType = 1
	// not supported
	ResponseFormatType_XML ResponseFormatType = 2
	// not supported
	ResponseFormatType_JSONP ResponseFormatType = 3
	// not supported
	ResponseFormatType_KML ResponseFormatType = 4
	// not supported
	ResponseFormatType_GEOCLUSTER ResponseFormatType = 5
	// not supported
	ResponseFormatType_GEOJSON ResponseFormatType = 6
)

var ResponseFormatType_name = map[int32]string{
	0: "PROTOBUF",
	1: "JSON",
	2: "XML",
	3: "JSONP",
	4: "KML",
	5: "GEOCLUSTER",
	6: "GEOJSON",
}
var ResponseFormatType_value = map[string]int32{
	"PROTOBUF":   0,
	"JSON":       1,
	"XML":        2,
	"JSONP":      3,
	"KML":        4,
	"GEOCLUSTER": 5,
	"GEOJSON":    6,
}

func (x ResponseFormatType) String() string {
	return proto.EnumName(ResponseFormatType_name, int32(x))
}
func (ResponseFormatType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type ResponseBlockType int32

const (
	ResponseBlockType_QUERY  ResponseBlockType = 0
	ResponseBlockType_ITEMS  ResponseBlockType = 1
	ResponseBlockType_FACETS ResponseBlockType = 2
	ResponseBlockType_LAYOUT ResponseBlockType = 3
)

var ResponseBlockType_name = map[int32]string{
	0: "QUERY",
	1: "ITEMS",
	2: "FACETS",
	3: "LAYOUT",
}
var ResponseBlockType_value = map[string]int32{
	"QUERY":  0,
	"ITEMS":  1,
	"FACETS": 2,
	"LAYOUT": 3,
}

func (x ResponseBlockType) String() string {
	return proto.EnumName(ResponseBlockType_name, int32(x))
}
func (ResponseBlockType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type GeoType int32

const (
	GeoType_BBOX    GeoType = 0
	GeoType_GEOFILT GeoType = 1
)

var GeoType_name = map[int32]string{
	0: "BBOX",
	1: "GEOFILT",
}
var GeoType_value = map[string]int32{
	"BBOX":    0,
	"GEOFILT": 1,
}

func (x GeoType) String() string {
	return proto.EnumName(GeoType_name, int32(x))
}
func (GeoType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type IdType int32

const (
	// same as ES doc_id
	IdType_HUDID IdType = 0
	// case insensitive id search
	IdType_IDCI IdType = 1
	// named graph
	IdType_NAMEDGRAPH IdType = 2
)

var IdType_name = map[int32]string{
	0: "HUDID",
	1: "IDCI",
	2: "NAMEDGRAPH",
}
var IdType_value = map[string]int32{
	"HUDID":      0,
	"IDCI":       1,
	"NAMEDGRAPH": 2,
}

func (x IdType) String() string {
	return proto.EnumName(IdType_name, int32(x))
}
func (IdType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type FilterValue struct {
	Value []string `protobuf:"bytes,1,rep,name=value" json:"value,omitempty"`
}

func (m *FilterValue) Reset()                    { *m = FilterValue{} }
func (m *FilterValue) String() string            { return proto.CompactTextString(m) }
func (*FilterValue) ProtoMessage()               {}
func (*FilterValue) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *FilterValue) GetValue() []string {
	if m != nil {
		return m.Value
	}
	return nil
}

type SearchRequest struct {
	// Will output a summary result set. Any valid Lucene or Solr Query syntax will work.
	Query  string             `protobuf:"bytes,1,opt,name=query" json:"query,omitempty"`
	Format ResponseFormatType `protobuf:"varint,2,opt,name=format,enum=api.ResponseFormatType" json:"format,omitempty"`
	// number of results returned
	// rows
	ResponseSize      int32                   `protobuf:"varint,3,opt,name=responseSize" json:"responseSize,omitempty"`
	Start             int32                   `protobuf:"varint,4,opt,name=start" json:"start,omitempty"`
	Page              int32                   `protobuf:"varint,5,opt,name=page" json:"page,omitempty"`
	QueryFilter       map[string]*FilterValue `protobuf:"bytes,6,rep,name=QueryFilter" json:"QueryFilter,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	HiddenQueryFilter map[string]*FilterValue `protobuf:"bytes,7,rep,name=HiddenQueryFilter" json:"HiddenQueryFilter,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Disable           []ResponseBlockType     `protobuf:"varint,8,rep,packed,name=disable,enum=api.ResponseBlockType" json:"disable,omitempty"`
	Enable            []ResponseFormatType    `protobuf:"varint,9,rep,packed,name=enable,enum=api.ResponseFormatType" json:"enable,omitempty"`
	FacetField        []string                `protobuf:"bytes,10,rep,name=FacetField" json:"FacetField,omitempty"`
	FacetLimit        int32                   `protobuf:"varint,11,opt,name=FacetLimit" json:"FacetLimit,omitempty"`
	FacetBoolType     bool                    `protobuf:"varint,12,opt,name=FacetBoolType" json:"FacetBoolType,omitempty"`
	SortBy            string                  `protobuf:"bytes,13,opt,name=sortBy" json:"sortBy,omitempty"`
	// geo options
	LatLong  string `protobuf:"bytes,14,opt,name=LatLong" json:"LatLong,omitempty"`
	Distance string `protobuf:"bytes,15,opt,name=Distance" json:"Distance,omitempty"`
	// min_* and max_* are the bounding box parameters
	MinX float32 `protobuf:"fixed32,16,opt,name=min_x,json=minX" json:"min_x,omitempty"`
	MinY float32 `protobuf:"fixed32,17,opt,name=min_y,json=minY" json:"min_y,omitempty"`
	MaxX float32 `protobuf:"fixed32,18,opt,name=max_x,json=maxX" json:"max_x,omitempty"`
	MaxY float32 `protobuf:"fixed32,19,opt,name=max_y,json=maxY" json:"max_y,omitempty"`
	// add support for polygon
	Field   []string `protobuf:"bytes,20,rep,name=field" json:"field,omitempty"`
	GeoType GeoType  `protobuf:"varint,21,opt,name=geoType,enum=api.GeoType" json:"geoType,omitempty"`
	// qr
	QueryRefinement string `protobuf:"bytes,22,opt,name=QueryRefinement" json:"QueryRefinement,omitempty"`
}

func (m *SearchRequest) Reset()                    { *m = SearchRequest{} }
func (m *SearchRequest) String() string            { return proto.CompactTextString(m) }
func (*SearchRequest) ProtoMessage()               {}
func (*SearchRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *SearchRequest) GetQuery() string {
	if m != nil {
		return m.Query
	}
	return ""
}

func (m *SearchRequest) GetFormat() ResponseFormatType {
	if m != nil {
		return m.Format
	}
	return ResponseFormatType_PROTOBUF
}

func (m *SearchRequest) GetResponseSize() int32 {
	if m != nil {
		return m.ResponseSize
	}
	return 0
}

func (m *SearchRequest) GetStart() int32 {
	if m != nil {
		return m.Start
	}
	return 0
}

func (m *SearchRequest) GetPage() int32 {
	if m != nil {
		return m.Page
	}
	return 0
}

func (m *SearchRequest) GetQueryFilter() map[string]*FilterValue {
	if m != nil {
		return m.QueryFilter
	}
	return nil
}

func (m *SearchRequest) GetHiddenQueryFilter() map[string]*FilterValue {
	if m != nil {
		return m.HiddenQueryFilter
	}
	return nil
}

func (m *SearchRequest) GetDisable() []ResponseBlockType {
	if m != nil {
		return m.Disable
	}
	return nil
}

func (m *SearchRequest) GetEnable() []ResponseFormatType {
	if m != nil {
		return m.Enable
	}
	return nil
}

func (m *SearchRequest) GetFacetField() []string {
	if m != nil {
		return m.FacetField
	}
	return nil
}

func (m *SearchRequest) GetFacetLimit() int32 {
	if m != nil {
		return m.FacetLimit
	}
	return 0
}

func (m *SearchRequest) GetFacetBoolType() bool {
	if m != nil {
		return m.FacetBoolType
	}
	return false
}

func (m *SearchRequest) GetSortBy() string {
	if m != nil {
		return m.SortBy
	}
	return ""
}

func (m *SearchRequest) GetLatLong() string {
	if m != nil {
		return m.LatLong
	}
	return ""
}

func (m *SearchRequest) GetDistance() string {
	if m != nil {
		return m.Distance
	}
	return ""
}

func (m *SearchRequest) GetMinX() float32 {
	if m != nil {
		return m.MinX
	}
	return 0
}

func (m *SearchRequest) GetMinY() float32 {
	if m != nil {
		return m.MinY
	}
	return 0
}

func (m *SearchRequest) GetMaxX() float32 {
	if m != nil {
		return m.MaxX
	}
	return 0
}

func (m *SearchRequest) GetMaxY() float32 {
	if m != nil {
		return m.MaxY
	}
	return 0
}

func (m *SearchRequest) GetField() []string {
	if m != nil {
		return m.Field
	}
	return nil
}

func (m *SearchRequest) GetGeoType() GeoType {
	if m != nil {
		return m.GeoType
	}
	return GeoType_BBOX
}

func (m *SearchRequest) GetQueryRefinement() string {
	if m != nil {
		return m.QueryRefinement
	}
	return ""
}

type DetailRequest struct {
	// option: any valid identifier specified by the idType
	// description: Will output a full-view. Default idType is hubId taken from the delving_hubId field.
	Id             string             `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Mlt            bool               `protobuf:"varint,2,opt,name=mlt" json:"mlt,omitempty"`
	Format         ResponseFormatType `protobuf:"varint,3,opt,name=format,enum=api.ResponseFormatType" json:"format,omitempty"`
	MltCount       int32              `protobuf:"varint,4,opt,name=mltCount" json:"mltCount,omitempty"`
	MltQueryFilter string             `protobuf:"bytes,5,opt,name=mltQueryFilter" json:"mltQueryFilter,omitempty"`
	MltFilterKey   string             `protobuf:"bytes,6,opt,name=mltFilterKey" json:"mltFilterKey,omitempty"`
	// searchRequest is a serialised form of the search result and is the return
	// to results link
	SearchRequest string `protobuf:"bytes,7,opt,name=searchRequest" json:"searchRequest,omitempty"`
	// resultIndex is the point where this detail object is in the search result order
	ResultIndex int32 `protobuf:"varint,8,opt,name=resultIndex" json:"resultIndex,omitempty"`
	// converter for result fields
	Converter string `protobuf:"bytes,9,opt,name=converter" json:"converter,omitempty"`
	// the type of id used in the ?id field
	IdType IdType `protobuf:"varint,10,opt,name=idType,enum=api.IdType" json:"idType,omitempty"`
}

func (m *DetailRequest) Reset()                    { *m = DetailRequest{} }
func (m *DetailRequest) String() string            { return proto.CompactTextString(m) }
func (*DetailRequest) ProtoMessage()               {}
func (*DetailRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *DetailRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *DetailRequest) GetMlt() bool {
	if m != nil {
		return m.Mlt
	}
	return false
}

func (m *DetailRequest) GetFormat() ResponseFormatType {
	if m != nil {
		return m.Format
	}
	return ResponseFormatType_PROTOBUF
}

func (m *DetailRequest) GetMltCount() int32 {
	if m != nil {
		return m.MltCount
	}
	return 0
}

func (m *DetailRequest) GetMltQueryFilter() string {
	if m != nil {
		return m.MltQueryFilter
	}
	return ""
}

func (m *DetailRequest) GetMltFilterKey() string {
	if m != nil {
		return m.MltFilterKey
	}
	return ""
}

func (m *DetailRequest) GetSearchRequest() string {
	if m != nil {
		return m.SearchRequest
	}
	return ""
}

func (m *DetailRequest) GetResultIndex() int32 {
	if m != nil {
		return m.ResultIndex
	}
	return 0
}

func (m *DetailRequest) GetConverter() string {
	if m != nil {
		return m.Converter
	}
	return ""
}

func (m *DetailRequest) GetIdType() IdType {
	if m != nil {
		return m.IdType
	}
	return IdType_HUDID
}

type BreadCrumb struct {
	Href           string `protobuf:"bytes,1,opt,name=href" json:"href,omitempty"`
	Display        string `protobuf:"bytes,2,opt,name=display" json:"display,omitempty"`
	Field          string `protobuf:"bytes,3,opt,name=field" json:"field,omitempty"`
	LocalisedField string `protobuf:"bytes,4,opt,name=localised_field,json=localisedField" json:"localised_field,omitempty"`
	Value          string `protobuf:"bytes,5,opt,name=value" json:"value,omitempty"`
	IsLast         bool   `protobuf:"varint,6,opt,name=is_last,json=isLast" json:"is_last,omitempty"`
}

func (m *BreadCrumb) Reset()                    { *m = BreadCrumb{} }
func (m *BreadCrumb) String() string            { return proto.CompactTextString(m) }
func (*BreadCrumb) ProtoMessage()               {}
func (*BreadCrumb) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *BreadCrumb) GetHref() string {
	if m != nil {
		return m.Href
	}
	return ""
}

func (m *BreadCrumb) GetDisplay() string {
	if m != nil {
		return m.Display
	}
	return ""
}

func (m *BreadCrumb) GetField() string {
	if m != nil {
		return m.Field
	}
	return ""
}

func (m *BreadCrumb) GetLocalisedField() string {
	if m != nil {
		return m.LocalisedField
	}
	return ""
}

func (m *BreadCrumb) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

func (m *BreadCrumb) GetIsLast() bool {
	if m != nil {
		return m.IsLast
	}
	return false
}

type PaginationLink struct {
	Start      int32 `protobuf:"varint,1,opt,name=start" json:"start,omitempty"`
	IsLinked   bool  `protobuf:"varint,2,opt,name=isLinked" json:"isLinked,omitempty"`
	PageNumber int32 `protobuf:"varint,3,opt,name=pageNumber" json:"pageNumber,omitempty"`
}

func (m *PaginationLink) Reset()                    { *m = PaginationLink{} }
func (m *PaginationLink) String() string            { return proto.CompactTextString(m) }
func (*PaginationLink) ProtoMessage()               {}
func (*PaginationLink) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *PaginationLink) GetStart() int32 {
	if m != nil {
		return m.Start
	}
	return 0
}

func (m *PaginationLink) GetIsLinked() bool {
	if m != nil {
		return m.IsLinked
	}
	return false
}

func (m *PaginationLink) GetPageNumber() int32 {
	if m != nil {
		return m.PageNumber
	}
	return 0
}

type Pagination struct {
	Start        int32             `protobuf:"varint,1,opt,name=start" json:"start,omitempty"`
	Rows         int32             `protobuf:"varint,2,opt,name=rows" json:"rows,omitempty"`
	NumFound     int32             `protobuf:"varint,3,opt,name=numFound" json:"numFound,omitempty"`
	HasNext      bool              `protobuf:"varint,4,opt,name=hasNext" json:"hasNext,omitempty"`
	NextPage     int32             `protobuf:"varint,5,opt,name=nextPage" json:"nextPage,omitempty"`
	HasPrevious  bool              `protobuf:"varint,6,opt,name=hasPrevious" json:"hasPrevious,omitempty"`
	PreviousPage int32             `protobuf:"varint,7,opt,name=previousPage" json:"previousPage,omitempty"`
	CurrentPage  int32             `protobuf:"varint,8,opt,name=currentPage" json:"currentPage,omitempty"`
	Links        []*PaginationLink `protobuf:"bytes,9,rep,name=links" json:"links,omitempty"`
}

func (m *Pagination) Reset()                    { *m = Pagination{} }
func (m *Pagination) String() string            { return proto.CompactTextString(m) }
func (*Pagination) ProtoMessage()               {}
func (*Pagination) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *Pagination) GetStart() int32 {
	if m != nil {
		return m.Start
	}
	return 0
}

func (m *Pagination) GetRows() int32 {
	if m != nil {
		return m.Rows
	}
	return 0
}

func (m *Pagination) GetNumFound() int32 {
	if m != nil {
		return m.NumFound
	}
	return 0
}

func (m *Pagination) GetHasNext() bool {
	if m != nil {
		return m.HasNext
	}
	return false
}

func (m *Pagination) GetNextPage() int32 {
	if m != nil {
		return m.NextPage
	}
	return 0
}

func (m *Pagination) GetHasPrevious() bool {
	if m != nil {
		return m.HasPrevious
	}
	return false
}

func (m *Pagination) GetPreviousPage() int32 {
	if m != nil {
		return m.PreviousPage
	}
	return 0
}

func (m *Pagination) GetCurrentPage() int32 {
	if m != nil {
		return m.CurrentPage
	}
	return 0
}

func (m *Pagination) GetLinks() []*PaginationLink {
	if m != nil {
		return m.Links
	}
	return nil
}

type Query struct {
	Numfound    int32         `protobuf:"varint,1,opt,name=numfound" json:"numfound,omitempty"`
	Terms       string        `protobuf:"bytes,2,opt,name=terms" json:"terms,omitempty"`
	BreadCrumbs []*BreadCrumb `protobuf:"bytes,3,rep,name=breadCrumbs" json:"breadCrumbs,omitempty"`
}

func (m *Query) Reset()                    { *m = Query{} }
func (m *Query) String() string            { return proto.CompactTextString(m) }
func (*Query) ProtoMessage()               {}
func (*Query) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *Query) GetNumfound() int32 {
	if m != nil {
		return m.Numfound
	}
	return 0
}

func (m *Query) GetTerms() string {
	if m != nil {
		return m.Terms
	}
	return ""
}

func (m *Query) GetBreadCrumbs() []*BreadCrumb {
	if m != nil {
		return m.BreadCrumbs
	}
	return nil
}

type Facet struct {
	Name        string      `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	IsSelected  bool        `protobuf:"varint,2,opt,name=isSelected" json:"isSelected,omitempty"`
	I18N        string      `protobuf:"bytes,3,opt,name=i18n" json:"i18n,omitempty"`
	Total       int32       `protobuf:"varint,4,opt,name=total" json:"total,omitempty"`
	MissingDocs int32       `protobuf:"varint,5,opt,name=missingDocs" json:"missingDocs,omitempty"`
	OtherDocs   int32       `protobuf:"varint,6,opt,name=otherDocs" json:"otherDocs,omitempty"`
	Links       []*FaceLink `protobuf:"bytes,7,rep,name=links" json:"links,omitempty"`
}

func (m *Facet) Reset()                    { *m = Facet{} }
func (m *Facet) String() string            { return proto.CompactTextString(m) }
func (*Facet) ProtoMessage()               {}
func (*Facet) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *Facet) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Facet) GetIsSelected() bool {
	if m != nil {
		return m.IsSelected
	}
	return false
}

func (m *Facet) GetI18N() string {
	if m != nil {
		return m.I18N
	}
	return ""
}

func (m *Facet) GetTotal() int32 {
	if m != nil {
		return m.Total
	}
	return 0
}

func (m *Facet) GetMissingDocs() int32 {
	if m != nil {
		return m.MissingDocs
	}
	return 0
}

func (m *Facet) GetOtherDocs() int32 {
	if m != nil {
		return m.OtherDocs
	}
	return 0
}

func (m *Facet) GetLinks() []*FaceLink {
	if m != nil {
		return m.Links
	}
	return nil
}

type FaceLink struct {
	Url           string `protobuf:"bytes,1,opt,name=url" json:"url,omitempty"`
	IsSelected    bool   `protobuf:"varint,2,opt,name=isSelected" json:"isSelected,omitempty"`
	Value         string `protobuf:"bytes,3,opt,name=value" json:"value,omitempty"`
	Count         int32  `protobuf:"varint,4,opt,name=count" json:"count,omitempty"`
	DisplayString string `protobuf:"bytes,5,opt,name=displayString" json:"displayString,omitempty"`
}

func (m *FaceLink) Reset()                    { *m = FaceLink{} }
func (m *FaceLink) String() string            { return proto.CompactTextString(m) }
func (*FaceLink) ProtoMessage()               {}
func (*FaceLink) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *FaceLink) GetUrl() string {
	if m != nil {
		return m.Url
	}
	return ""
}

func (m *FaceLink) GetIsSelected() bool {
	if m != nil {
		return m.IsSelected
	}
	return false
}

func (m *FaceLink) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

func (m *FaceLink) GetCount() int32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *FaceLink) GetDisplayString() string {
	if m != nil {
		return m.DisplayString
	}
	return ""
}

type MetadataFieldV1 struct {
	Field []string `protobuf:"bytes,1,rep,name=field" json:"field,omitempty"`
}

func (m *MetadataFieldV1) Reset()                    { *m = MetadataFieldV1{} }
func (m *MetadataFieldV1) String() string            { return proto.CompactTextString(m) }
func (*MetadataFieldV1) ProtoMessage()               {}
func (*MetadataFieldV1) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *MetadataFieldV1) GetField() []string {
	if m != nil {
		return m.Field
	}
	return nil
}

type MetadataItemV1 struct {
	DocId   string                      `protobuf:"bytes,1,opt,name=doc_id,json=docId" json:"doc_id,omitempty"`
	DocType string                      `protobuf:"bytes,2,opt,name=doc_type,json=docType" json:"doc_type,omitempty"`
	Fields  map[string]*MetadataFieldV1 `protobuf:"bytes,3,rep,name=fields" json:"fields,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *MetadataItemV1) Reset()                    { *m = MetadataItemV1{} }
func (m *MetadataItemV1) String() string            { return proto.CompactTextString(m) }
func (*MetadataItemV1) ProtoMessage()               {}
func (*MetadataItemV1) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *MetadataItemV1) GetDocId() string {
	if m != nil {
		return m.DocId
	}
	return ""
}

func (m *MetadataItemV1) GetDocType() string {
	if m != nil {
		return m.DocType
	}
	return ""
}

func (m *MetadataItemV1) GetFields() map[string]*MetadataFieldV1 {
	if m != nil {
		return m.Fields
	}
	return nil
}

type SearchResultWrapperV1 struct {
	Result *SearchResultV1 `protobuf:"bytes,1,opt,name=result" json:"result,omitempty"`
}

func (m *SearchResultWrapperV1) Reset()                    { *m = SearchResultWrapperV1{} }
func (m *SearchResultWrapperV1) String() string            { return proto.CompactTextString(m) }
func (*SearchResultWrapperV1) ProtoMessage()               {}
func (*SearchResultWrapperV1) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *SearchResultWrapperV1) GetResult() *SearchResultV1 {
	if m != nil {
		return m.Result
	}
	return nil
}

// Full SearchResult
type SearchResultV1 struct {
	Query      *Query            `protobuf:"bytes,1,opt,name=query" json:"query,omitempty"`
	Pagination *Pagination       `protobuf:"bytes,2,opt,name=pagination" json:"pagination,omitempty"`
	Items      []*MetadataItemV1 `protobuf:"bytes,3,rep,name=items" json:"items,omitempty"`
	Facets     []*Facet          `protobuf:"bytes,4,rep,name=facets" json:"facets,omitempty"`
}

func (m *SearchResultV1) Reset()                    { *m = SearchResultV1{} }
func (m *SearchResultV1) String() string            { return proto.CompactTextString(m) }
func (*SearchResultV1) ProtoMessage()               {}
func (*SearchResultV1) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

func (m *SearchResultV1) GetQuery() *Query {
	if m != nil {
		return m.Query
	}
	return nil
}

func (m *SearchResultV1) GetPagination() *Pagination {
	if m != nil {
		return m.Pagination
	}
	return nil
}

func (m *SearchResultV1) GetItems() []*MetadataItemV1 {
	if m != nil {
		return m.Items
	}
	return nil
}

func (m *SearchResultV1) GetFacets() []*Facet {
	if m != nil {
		return m.Facets
	}
	return nil
}

// The structure of the detail page
type DetailResult struct {
	Item *MetadataItemV1 `protobuf:"bytes,1,opt,name=item" json:"item,omitempty"`
}

func (m *DetailResult) Reset()                    { *m = DetailResult{} }
func (m *DetailResult) String() string            { return proto.CompactTextString(m) }
func (*DetailResult) ProtoMessage()               {}
func (*DetailResult) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

func (m *DetailResult) GetItem() *MetadataItemV1 {
	if m != nil {
		return m.Item
	}
	return nil
}

func init() {
	proto.RegisterType((*FilterValue)(nil), "api.FilterValue")
	proto.RegisterType((*SearchRequest)(nil), "api.SearchRequest")
	proto.RegisterType((*DetailRequest)(nil), "api.DetailRequest")
	proto.RegisterType((*BreadCrumb)(nil), "api.BreadCrumb")
	proto.RegisterType((*PaginationLink)(nil), "api.PaginationLink")
	proto.RegisterType((*Pagination)(nil), "api.Pagination")
	proto.RegisterType((*Query)(nil), "api.Query")
	proto.RegisterType((*Facet)(nil), "api.Facet")
	proto.RegisterType((*FaceLink)(nil), "api.FaceLink")
	proto.RegisterType((*MetadataFieldV1)(nil), "api.MetadataFieldV1")
	proto.RegisterType((*MetadataItemV1)(nil), "api.MetadataItemV1")
	proto.RegisterType((*SearchResultWrapperV1)(nil), "api.SearchResultWrapperV1")
	proto.RegisterType((*SearchResultV1)(nil), "api.SearchResultV1")
	proto.RegisterType((*DetailResult)(nil), "api.DetailResult")
	proto.RegisterEnum("api.ResponseFormatType", ResponseFormatType_name, ResponseFormatType_value)
	proto.RegisterEnum("api.ResponseBlockType", ResponseBlockType_name, ResponseBlockType_value)
	proto.RegisterEnum("api.GeoType", GeoType_name, GeoType_value)
	proto.RegisterEnum("api.IdType", IdType_name, IdType_value)
}

func init() { proto.RegisterFile("hub3/api/api.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 1436 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x57, 0xcd, 0x52, 0x1b, 0x47,
	0x10, 0xf6, 0xea, 0x67, 0x25, 0x5a, 0x20, 0xd6, 0x03, 0xc6, 0x1b, 0x2a, 0x95, 0xa8, 0x44, 0xca,
	0x96, 0x49, 0x05, 0x07, 0x7c, 0xb0, 0x2b, 0x37, 0x40, 0x12, 0x56, 0x2c, 0x10, 0x5e, 0x01, 0x86,
	0x13, 0x35, 0x68, 0x07, 0x98, 0x62, 0x7f, 0xe4, 0x9d, 0x59, 0x07, 0xe5, 0x21, 0xf2, 0x02, 0x79,
	0x81, 0x9c, 0xf3, 0x06, 0xb9, 0xa6, 0x72, 0xcd, 0xfb, 0xa4, 0xa6, 0x67, 0x77, 0xb5, 0x02, 0x5c,
	0x49, 0x55, 0x0e, 0xaa, 0x9a, 0xfe, 0xba, 0xa7, 0xa7, 0xa7, 0xa7, 0xbf, 0xee, 0x15, 0x90, 0xeb,
	0xf8, 0xe2, 0xd5, 0x4b, 0x3a, 0xe6, 0xea, 0xb7, 0x31, 0x8e, 0x42, 0x19, 0x92, 0x22, 0x1d, 0xf3,
	0xe6, 0x1a, 0xd4, 0xba, 0xdc, 0x93, 0x2c, 0x3a, 0xa1, 0x5e, 0xcc, 0xc8, 0x32, 0x94, 0x3f, 0xa9,
	0x85, 0x6d, 0x34, 0x8a, 0xad, 0x39, 0x47, 0x0b, 0xcd, 0x3f, 0x2a, 0xb0, 0x30, 0x64, 0x34, 0x1a,
	0x5d, 0x3b, 0xec, 0x63, 0xcc, 0x84, 0x54, 0x76, 0x1f, 0x63, 0x16, 0x4d, 0x6c, 0xa3, 0x61, 0x28,
	0x3b, 0x14, 0xc8, 0x4b, 0x30, 0x2f, 0xc3, 0xc8, 0xa7, 0xd2, 0x2e, 0x34, 0x8c, 0x56, 0x7d, 0xeb,
	0xe9, 0x86, 0x3a, 0xcd, 0x61, 0x62, 0x1c, 0x06, 0x82, 0x75, 0x51, 0x75, 0x34, 0x19, 0x33, 0x27,
	0x31, 0x23, 0x4d, 0x98, 0x8f, 0x12, 0xed, 0x90, 0xff, 0xcc, 0xec, 0x62, 0xc3, 0x68, 0x95, 0x9d,
	0x19, 0x4c, 0x1d, 0x25, 0x24, 0x8d, 0xa4, 0x5d, 0x42, 0xa5, 0x16, 0x08, 0x81, 0xd2, 0x98, 0x5e,
	0x31, 0xbb, 0x8c, 0x20, 0xae, 0x49, 0x07, 0x6a, 0xef, 0x55, 0x1c, 0xfa, 0x42, 0xb6, 0xd9, 0x28,
	0xb6, 0x6a, 0x5b, 0x6b, 0x18, 0xc3, 0x4c, 0xf4, 0x1b, 0x39, 0xab, 0x4e, 0x20, 0xa3, 0x89, 0x93,
	0xdf, 0x47, 0x3e, 0xc0, 0xe3, 0xb7, 0xdc, 0x75, 0x59, 0x90, 0x77, 0x56, 0x41, 0x67, 0x2f, 0x1e,
	0x70, 0x76, 0xcf, 0x56, 0xbb, 0xbc, 0xef, 0x83, 0x7c, 0x0f, 0x15, 0x97, 0x0b, 0x7a, 0xe1, 0x31,
	0xbb, 0xda, 0x28, 0xb6, 0xea, 0x5b, 0x2b, 0x33, 0xf9, 0xd9, 0xf1, 0xc2, 0xd1, 0x0d, 0xa6, 0x27,
	0x35, 0x53, 0x09, 0x65, 0x01, 0x6e, 0x98, 0xc3, 0x0d, 0x9f, 0x4f, 0xa8, 0x36, 0x23, 0x5f, 0x01,
	0x74, 0xe9, 0x88, 0xc9, 0x2e, 0x67, 0x9e, 0x6b, 0x03, 0x3e, 0x62, 0x0e, 0xc9, 0xf4, 0x7d, 0xee,
	0x73, 0x69, 0xd7, 0x30, 0x79, 0x39, 0x84, 0x7c, 0x03, 0x0b, 0x28, 0xed, 0x84, 0xa1, 0xa7, 0x1c,
	0xdb, 0xf3, 0x0d, 0xa3, 0x55, 0x75, 0x66, 0x41, 0xb2, 0x02, 0xa6, 0x08, 0x23, 0xb9, 0x33, 0xb1,
	0x17, 0xf0, 0xf9, 0x13, 0x89, 0xd8, 0x50, 0xe9, 0x53, 0xd9, 0x0f, 0x83, 0x2b, 0xbb, 0x8e, 0x8a,
	0x54, 0x24, 0xab, 0x50, 0x6d, 0x73, 0x21, 0x69, 0x30, 0x62, 0xf6, 0x22, 0xaa, 0x32, 0x99, 0x2c,
	0x41, 0xd9, 0xe7, 0xc1, 0xf9, 0xad, 0x6d, 0x35, 0x8c, 0x56, 0xc1, 0x29, 0xf9, 0x3c, 0x38, 0x4d,
	0xc1, 0x89, 0xfd, 0x38, 0x03, 0xcf, 0x10, 0xa4, 0xb7, 0xe7, 0xb7, 0x36, 0x49, 0x40, 0x7a, 0x7b,
	0x9a, 0x82, 0x13, 0x7b, 0x29, 0x03, 0xcf, 0x54, 0xd1, 0x5c, 0x62, 0x0a, 0x96, 0x75, 0x1d, 0xa3,
	0x40, 0x9e, 0x41, 0xe5, 0x8a, 0x85, 0x78, 0xaf, 0x27, 0x58, 0xa0, 0xf3, 0x98, 0xcf, 0x3d, 0x8d,
	0x39, 0xa9, 0x92, 0xb4, 0x60, 0x11, 0xdf, 0xcd, 0x61, 0x97, 0x3c, 0x60, 0x3e, 0x0b, 0xa4, 0xbd,
	0x82, 0x41, 0xdf, 0x85, 0x57, 0x0f, 0xc1, 0xba, 0xfb, 0xf2, 0xc4, 0x82, 0xe2, 0x0d, 0x4b, 0x99,
	0xa1, 0x96, 0xe4, 0x59, 0xca, 0x2a, 0x45, 0x8b, 0xda, 0x96, 0x85, 0xa7, 0xe6, 0x68, 0x97, 0xf0,
	0xec, 0x87, 0xc2, 0x1b, 0x63, 0xf5, 0x04, 0x56, 0x1e, 0xae, 0xa8, 0xff, 0xe7, 0xb7, 0xf9, 0x77,
	0x01, 0x16, 0xda, 0x4c, 0x52, 0xee, 0xa5, 0x1c, 0xae, 0x43, 0x81, 0xbb, 0x89, 0xbb, 0x02, 0x77,
	0x95, 0x7f, 0xdf, 0xd3, 0xd4, 0xad, 0x3a, 0x6a, 0x99, 0xe3, 0x73, 0xf1, 0xbf, 0xf1, 0x79, 0x15,
	0xaa, 0xbe, 0x27, 0x77, 0xc3, 0x38, 0x48, 0xe9, 0x9a, 0xc9, 0xe4, 0x19, 0xd4, 0x7d, 0x4f, 0xe6,
	0x39, 0x55, 0xc6, 0xa3, 0xef, 0xa0, 0xaa, 0x27, 0xf8, 0x9e, 0xd4, 0xc2, 0x3b, 0x36, 0xb1, 0x4d,
	0xb4, 0x9a, 0xc1, 0x54, 0x99, 0x8a, 0x3c, 0x09, 0xed, 0x0a, 0x1a, 0xcd, 0x82, 0xa4, 0x01, 0xb5,
	0x88, 0x89, 0xd8, 0x93, 0xbd, 0xc0, 0x65, 0xb7, 0x76, 0x15, 0x03, 0xca, 0x43, 0xe4, 0x4b, 0x98,
	0x1b, 0x85, 0xc1, 0x27, 0x16, 0xa9, 0x70, 0xe6, 0xd0, 0xc7, 0x14, 0x20, 0x6b, 0x60, 0x72, 0x17,
	0xab, 0x05, 0xf0, 0xfa, 0x35, 0xbc, 0x7e, 0xcf, 0xd5, 0x57, 0xd6, 0xaa, 0xe6, 0x6f, 0x06, 0xc0,
	0x4e, 0xc4, 0xa8, 0xbb, 0x1b, 0xc5, 0xfe, 0x85, 0xea, 0x4b, 0xd7, 0x11, 0xbb, 0x4c, 0xd2, 0x8a,
	0x6b, 0x45, 0x0b, 0x97, 0x8b, 0xb1, 0x47, 0x27, 0x98, 0xdc, 0x39, 0x27, 0x15, 0xa7, 0x65, 0x5a,
	0xd4, 0x6d, 0x54, 0x97, 0xe9, 0x73, 0x58, 0xf4, 0xc2, 0x11, 0xf5, 0xb8, 0x60, 0xee, 0xb9, 0xd6,
	0x97, 0x74, 0xaa, 0x32, 0x58, 0xb3, 0x39, 0xeb, 0xd6, 0x3a, 0x93, 0x5a, 0x20, 0x4f, 0xa1, 0xc2,
	0xc5, 0xb9, 0x47, 0x85, 0xc4, 0xdc, 0x55, 0x1d, 0x93, 0x8b, 0x3e, 0x15, 0xb2, 0x79, 0x01, 0xf5,
	0x43, 0x7a, 0xc5, 0x03, 0x2a, 0x79, 0x18, 0xf4, 0x79, 0x70, 0x33, 0xed, 0xad, 0x46, 0xbe, 0xb7,
	0xae, 0x42, 0x95, 0x0b, 0xa5, 0x67, 0x6e, 0x52, 0x0d, 0x99, 0xac, 0x1a, 0x88, 0xea, 0xb5, 0x07,
	0xb1, 0x7f, 0xc1, 0xa2, 0xa4, 0x5f, 0xe7, 0x90, 0xe6, 0xaf, 0x05, 0x80, 0xe9, 0x21, 0x9f, 0x39,
	0x80, 0x40, 0x29, 0x0a, 0x7f, 0x12, 0xe8, 0xbc, 0xec, 0xe0, 0x5a, 0x1d, 0x1a, 0xc4, 0x7e, 0x37,
	0x8c, 0x03, 0x37, 0x71, 0x9b, 0xc9, 0x2a, 0x81, 0xd7, 0x54, 0x1c, 0xb0, 0x5b, 0x5d, 0x55, 0x55,
	0x27, 0x15, 0x71, 0x17, 0xbb, 0x95, 0x87, 0xd3, 0x51, 0x90, 0xc9, 0xea, 0xf9, 0xaf, 0xa9, 0x38,
	0x8c, 0xd8, 0x27, 0x1e, 0xc6, 0x22, 0xc9, 0x45, 0x1e, 0x52, 0xa5, 0x36, 0x4e, 0xd6, 0xe8, 0xa1,
	0xa2, 0xc7, 0x4f, 0x1e, 0x53, 0x5e, 0x46, 0x71, 0x14, 0xb1, 0x40, 0x1f, 0x92, 0x14, 0x51, 0x0e,
	0x22, 0x2f, 0xa0, 0xec, 0xf1, 0xe0, 0x46, 0x60, 0x8f, 0xae, 0x6d, 0x2d, 0x61, 0x95, 0xcc, 0x26,
	0xda, 0xd1, 0x16, 0x4d, 0x0f, 0xca, 0x58, 0xea, 0xc9, 0x6d, 0x2f, 0xf1, 0xb6, 0x46, 0x76, 0x5b,
	0x94, 0x55, 0xce, 0x24, 0x8b, 0x7c, 0x91, 0x14, 0x8b, 0x16, 0xc8, 0x26, 0xd4, 0x2e, 0xb2, 0x32,
	0x13, 0x76, 0x11, 0xcf, 0x5a, 0xc4, 0xb3, 0xa6, 0xe5, 0xe7, 0xe4, 0x6d, 0x9a, 0x7f, 0x1a, 0x50,
	0xc6, 0xc6, 0xad, 0x12, 0x1e, 0x50, 0x9f, 0xa5, 0x55, 0xa9, 0xd6, 0xea, 0x25, 0xb9, 0x18, 0x32,
	0x8f, 0x8d, 0x64, 0xf6, 0xce, 0x39, 0x44, 0xed, 0xe1, 0x9b, 0x6f, 0x82, 0xa4, 0x34, 0x71, 0x8d,
	0xa1, 0x85, 0x92, 0x7a, 0xe9, 0x2c, 0x46, 0x41, 0xa5, 0xc8, 0xe7, 0x42, 0xf0, 0xe0, 0xaa, 0x1d,
	0x8e, 0x44, 0xf2, 0x0e, 0x79, 0x48, 0xf1, 0x2c, 0x94, 0xd7, 0x2c, 0x42, 0xbd, 0x89, 0xfa, 0x29,
	0x40, 0xd6, 0xd2, 0x04, 0xea, 0x21, 0xbb, 0xa0, 0xdb, 0x18, 0x1d, 0xb1, 0x7c, 0xea, 0x7e, 0x31,
	0xa0, 0x9a, 0x62, 0xaa, 0x55, 0xc5, 0x91, 0x97, 0xb6, 0xc2, 0x38, 0xf2, 0xfe, 0xf5, 0x36, 0x19,
	0x55, 0x8a, 0x79, 0xaa, 0x2c, 0x43, 0x79, 0x94, 0x6b, 0x56, 0x5a, 0x50, 0xdd, 0x25, 0x21, 0xe8,
	0x50, 0x46, 0x3c, 0xb8, 0x4a, 0xe8, 0x35, 0x0b, 0x36, 0x9f, 0xc3, 0xe2, 0x3e, 0x93, 0xd4, 0xa5,
	0x92, 0x22, 0x1b, 0x4f, 0x36, 0xa7, 0x74, 0x36, 0x72, 0x53, 0xa7, 0xf9, 0x97, 0x01, 0xf5, 0xd4,
	0xb2, 0x27, 0x99, 0x7f, 0xb2, 0x49, 0x9e, 0x80, 0xe9, 0x86, 0xa3, 0xf3, 0xac, 0xfd, 0x96, 0xdd,
	0x70, 0xd4, 0x73, 0xc9, 0x17, 0x50, 0x55, 0xb0, 0x54, 0x2d, 0x27, 0xed, 0x14, 0xe1, 0x08, 0x47,
	0xd2, 0x6b, 0x30, 0xd1, 0x5b, 0xfa, 0xf2, 0x5f, 0x63, 0x92, 0x66, 0xdd, 0x6e, 0x60, 0x1c, 0x42,
	0x7f, 0x7f, 0x24, 0xe6, 0xab, 0x03, 0xf5, 0x81, 0x97, 0xc1, 0x0f, 0x0c, 0x91, 0xf5, 0xd9, 0x21,
	0xb2, 0x3c, 0xe3, 0x38, 0xb9, 0x59, 0x7e, 0x90, 0xb4, 0xe1, 0x49, 0xfa, 0x01, 0xa4, 0x1a, 0xe9,
	0x87, 0x88, 0x8e, 0xc7, 0x2c, 0x3a, 0xd9, 0x24, 0xdf, 0x82, 0xa9, 0x7b, 0x2b, 0x7a, 0x4f, 0x89,
	0x90, 0xb7, 0x3d, 0xd9, 0x74, 0x12, 0x93, 0xe6, 0xef, 0x06, 0xd4, 0x67, 0x55, 0xa4, 0x91, 0xff,
	0xa6, 0xac, 0x6d, 0x01, 0x6e, 0xd7, 0x03, 0x37, 0xfb, 0xbe, 0x54, 0xad, 0x26, 0xe1, 0x55, 0x12,
	0xef, 0xe2, 0x1d, 0xba, 0x39, 0x39, 0x13, 0x45, 0x4d, 0x2e, 0x99, 0x9f, 0x26, 0x6d, 0xe9, 0x81,
	0xa4, 0x39, 0xda, 0x82, 0x34, 0xc1, 0xbc, 0x54, 0x5c, 0x11, 0x76, 0x09, 0x6d, 0x21, 0xab, 0x42,
	0xe9, 0x24, 0x9a, 0xe6, 0x6b, 0x98, 0x4f, 0x47, 0xa8, 0x8a, 0x99, 0x3c, 0x87, 0x92, 0xda, 0x3c,
	0x73, 0xdf, 0x3b, 0xde, 0xd1, 0x60, 0xfd, 0x12, 0xc8, 0xfd, 0xa9, 0x49, 0xe6, 0xa1, 0x7a, 0xe8,
	0x0c, 0x8e, 0x06, 0x3b, 0xc7, 0x5d, 0xeb, 0x11, 0xa9, 0x42, 0xe9, 0xc7, 0xe1, 0xe0, 0xc0, 0x32,
	0x48, 0x05, 0x8a, 0xa7, 0xfb, 0x7d, 0xab, 0x40, 0xe6, 0xa0, 0xac, 0xa0, 0x43, 0xab, 0xa8, 0xb0,
	0x77, 0xfb, 0x7d, 0xab, 0x44, 0xea, 0x00, 0x7b, 0x9d, 0xc1, 0x6e, 0xff, 0x78, 0x78, 0xd4, 0x71,
	0xac, 0x32, 0xa9, 0x41, 0x65, 0xaf, 0x33, 0xc0, 0x9d, 0xe6, 0xfa, 0x36, 0x3c, 0xbe, 0xf7, 0x35,
	0xa9, 0xbc, 0xbc, 0x3f, 0xee, 0x38, 0x67, 0xd6, 0x23, 0xb5, 0xec, 0x1d, 0x75, 0xf6, 0x87, 0x96,
	0x41, 0x00, 0xcc, 0xee, 0xf6, 0x6e, 0xe7, 0x68, 0x68, 0x15, 0xd4, 0xba, 0xbf, 0x7d, 0x36, 0x38,
	0x3e, 0xb2, 0x8a, 0xeb, 0x0d, 0xa8, 0x24, 0xdf, 0x43, 0x2a, 0xa2, 0x9d, 0x9d, 0xc1, 0xa9, 0xf5,
	0x28, 0x39, 0xa4, 0xdb, 0xeb, 0x1f, 0x59, 0xc6, 0xfa, 0x77, 0x60, 0xea, 0x19, 0xa8, 0xdc, 0xbd,
	0x3d, 0x6e, 0xf7, 0xda, 0x3a, 0xfa, 0x5e, 0x7b, 0xb7, 0x67, 0x19, 0x2a, 0xc0, 0x83, 0xed, 0xfd,
	0x4e, 0x7b, 0xcf, 0xd9, 0x3e, 0x7c, 0x6b, 0x15, 0x2e, 0x4c, 0xfc, 0xb7, 0xf1, 0xea, 0x9f, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x66, 0x2b, 0xba, 0xed, 0x83, 0x0c, 0x00, 0x00,
}