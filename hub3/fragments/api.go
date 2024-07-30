// Copyright © 2017 Delving B.V. <info@delving.eu>
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

package fragments

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	fmt "fmt"
	"log"
	"log/slog"
	"math/rand"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	elastic "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	proto "google.golang.org/protobuf/proto"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/rdf"
)

const (
	qfKey              = "qf"
	qfKeyList          = "qf[]"
	qfIDKey            = "qf.id"
	qfIDKeyList        = "qf.id[]"
	qfExistList        = "qf.exist[]"
	qfExist            = "qf.exist"
	qfDateRangeKey     = "qf.dateRange"
	responseSize       = int32(16)
	metaTags           = "meta.tags"
	metaSpec           = "meta.spec"
	treeDepth          = "tree.depth"
	treeLeaf           = "tree.leaf"
	treeCLevel         = "tree.cLevel"
	resourcesEntries   = "resources.entries"
	entriesSearchLabel = "resources.entries.searchLabel"
	entriesValue       = "resources.entries.@value"
	facetDisplayLabel  = "%s (%d)"
)

func logConvErr(p string, v []string, err error) {
	sanitized := []string{}
	for _, p := range v {
		sanitized = append(sanitized, strings.ReplaceAll(p, "\n|\r", ""))
	}

	log.Printf("unable to convert %v to int for %s; %+v", sanitized, p, err)
}

// DefaultSearchRequest takes an Config Objects and sets the defaults
func DefaultSearchRequest(cfg *c.RawConfig) *SearchRequest {
	id := ksuid.New()
	sr := &SearchRequest{
		ResponseSize: responseSize,
		SessionID:    id.String(),
		OrgIDKey:     cfg.ElasticSearch.OrgIDKey,
	}

	return sr
}

// SearchRequestFromHex creates a SearchRequest object from a string
func SearchRequestFromHex(s string) (*SearchRequest, error) {
	decoded, err := hex.DecodeString(s)
	newSr := &SearchRequest{}

	if err != nil {
		return newSr, err
	}

	err = proto.Unmarshal(decoded, newSr)

	// TODO(kiivihal): remove facet fields from search request when start > 0 or page > 1

	return newSr, err
}

// NewFacetField parses the QueryString and creates a FacetField
func NewFacetField(field string) (*FacetField, error) {
	ff := FacetField{Size: int32(c.Config.ElasticSearch.FacetSize)}

	var err error

	switch {
	case strings.HasPrefix(field, "{"):
		err = json.Unmarshal([]byte(field), &ff)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to unmarshal facetfield")
		}
	default:
		ff.Field = field
		ff.Name = field
	}

	if ff.Field == "" {
		return nil, errors.Wrap(err, "Unable to unmarshal facetfield: field cannot be empty")
	}

	if strings.Contains(field, "@") {
		before, after, found := strings.Cut(field, "@")
		if found {
			ff.Field = before
			ff.Language = after
		}
	}

	if ff.Name == "" {
		ff.Name = ff.Field
	}

	switch {
	case strings.HasPrefix(ff.Field, "tree."):
		ff.Type = FacetType_TREEFACET
	case strings.HasPrefix(ff.Field, "meta.tag"):
		ff.Type = FacetType_METATAGS
	case strings.HasPrefix(ff.Field, "tag"):
		ff.Type = FacetType_TAGS
	case strings.EqualFold(ff.Field, "searchLabel"):
		ff.Type = FacetType_FIELDS
	}

	return &ff, nil
}

// NewSearchRequest builds a search request object from URL Parameters
func NewSearchRequest(orgID string, params url.Values) (*SearchRequest, error) {
	hexRequest := params.Get("scrollID")
	if hexRequest == "" {
		hexRequest = params.Get("qs")
	}

	if hexRequest != "" {
		sr, err := SearchRequestFromHex(hexRequest)
		sr.Paging = true

		if err != nil {
			log.Printf("Unable to parse search request from scrollID: %q", hexRequest)
			return nil, err
		}

		return sr, nil
	}

	tree := &TreeQuery{
		PageSize: 250,
		OrgID:    orgID,
	}

	sr := DefaultSearchRequest(&c.Config)
	sr.OrgID = orgID

	for p, v := range params {
		if len(v) == 0 {
			continue
		}

		qfCfg := QueryFilterConfig{}
		if strings.HasPrefix(p, "hqf") {
			p = strings.TrimPrefix(p, "h")
			qfCfg.Hidden = true
		}

		switch p {
		case "q", "query":
			sr.Query = params.Get(p)
		case "rq":
			for _, rq := range v {
				sr.QueryRefinement = append(sr.QueryRefinement, strings.TrimSpace(rq))
			}
		case qfKey, qfKeyList:
			for _, qf := range v {
				if qf == "" {
					continue
				}

				err := sr.AddQueryFilter(qf, qfCfg)
				if err != nil {
					return sr, err
				}
			}
		case qfIDKey, qfIDKeyList:
			for _, qf := range v {
				if qf == "" {
					continue
				}

				qfCfg.IsIDFilter = true

				err := sr.AddQueryFilter(qf, qfCfg)
				if err != nil {
					return sr, err
				}
			}
		case qfDateRangeKey, "qf.dateRange[]":
			for _, qf := range v {
				if qf == "" {
					continue
				}

				err := sr.AddDateRangeFilter(qf, qfCfg)
				if err != nil {
					return sr, err
				}
			}
		case "qf.tree", "qf.tree[]":
			for _, qf := range v {
				if qf == "" {
					continue
				}

				err := sr.AddTreeFilter(qf)
				if err != nil {
					return sr, err
				}
			}
		case "qf.date", "qf.date[]":
			for _, qf := range v {
				if qf == "" {
					continue
				}

				err := sr.AddDateFilter(qf, qfCfg)
				if err != nil {
					return sr, err
				}
			}

		case "qf.exist", qfExistList:
			for _, qf := range v {
				if qf == "" {
					continue
				}

				err := sr.AddFieldExistFilter(qf, qfCfg)
				if err != nil {
					return sr, err
				}
			}
		case "contextIndex":
			sr.ContextIndex = v
		case "searchFields":
			sr.SearchFields = params.Get(p)
		case "facet.field":
			for _, ff := range v {
				if ff == "" {
					continue
				}

				facet, err := NewFacetField(ff)
				if err != nil {
					return nil, err
				}

				sr.FacetField = append(sr.FacetField, facet)
			}
		case "facet.size", "facet.limit":
			size, err := strconv.ParseInt(params.Get(p), 10, 32)
			if err != nil {
				logConvErr(p, []string{params.Get(p)}, err)
				return sr, err
			}

			if size > 2000 {
				size = 2000
			}

			sr.FacetLimit = int32(size)
		case "facetBoolType", "facet.boolType":
			fbt := params.Get(p)
			if fbt != "" {
				sr.FacetAndBoolType = strings.EqualFold(fbt, "and")
			}
		case "facet.mergeFilter":
			sr.FacetMergeFilter = append(sr.FacetMergeFilter, v...)
		case "facetOrBetween", "facet.orBetween":
			slog.Info("facet or between", "p", p, "v", v)
			fbt := params.Get(p)
			if fbt != "" {
				sr.ORBetweenFacets = strings.EqualFold(fbt, "true")
			}
		case "facet.expand":
			sr.FacetExpand = params.Get(p)
		case "facet.filter":
			sr.FacetFilter = params.Get(p)
		case "facet.cursor":
			sr.FacetCursor = params.Get(p)
		case "format":
			switch params.Get(p) {
			case "protobuf":
				sr.ResponseFormatType = ResponseFormatType_PROTOBUF
			case "jsonld":
				sr.ResponseFormatType = ResponseFormatType_LDJSON
			case "bulkaction":
				sr.ResponseFormatType = ResponseFormatType_BULKACTION
			}
		case "rows", "limit", "size":
			size, err := strconv.ParseInt(params.Get(p), 10, 32)
			if err != nil {
				logConvErr(p, []string{params.Get(p)}, err)
				return sr, err
			}

			if size > 1000 {
				size = 1000
			}

			sr.ResponseSize = int32(size)
		case "itemFormat", "item.format":
			format := params.Get(p)
			switch format {
			case "fragmentGraph", "resource":
				sr.ItemFormat = ItemFormatType_FRAGMENTGRAPH
			case "grouped":
				sr.ItemFormat = ItemFormatType_GROUPED
			case "jsonld":
				sr.ItemFormat = ItemFormatType_JSONLD
			case "flat":
				sr.ItemFormat = ItemFormatType_FLAT
			case "tree":
				sr.ItemFormat = ItemFormatType_TREE
			default:
				sr.ItemFormat = ItemFormatType_SUMMARY
			}
		case "sortBy":
			sortKey := params.Get(p)
			if strings.HasPrefix(sortKey, "^") {
				sr.SortAsc = true
				sortKey = strings.TrimPrefix(sortKey, "^")
			}
			sr.SortBy = sortKey

		case "sortAsc":
			if strings.EqualFold(params.Get(p), "true") {
				sr.SortAsc = true
			}
		case "sortOrder":
			if strings.EqualFold(params.Get(p), "asc") {
				sr.SortAsc = true
			}
		case "collapseFormat":
			sr.CollapseFormat = params.Get(p)
		case "collapseOn":
			sr.CollapseOn = params.Get(p)
		case "collapseSort":
			sr.CollapseSort = params.Get(p)
		case "collapseSize":
			size, err := strconv.ParseInt(params.Get(p), 10, 32)
			if err != nil {
				logConvErr(p, v, err)
				return sr, err
			}

			sr.CollapseSize = int32(size)
		case "peek":
			sr.Peek = params.Get(p)
		case "byLeaf":
			sr.Tree = tree
			tree.Leaf = params.Get(p)
			tree.FillTree = strings.EqualFold(params.Get("fillTree"), "true")
		case "byDepth":
			sr.Tree = tree
			tree.Depth = v
		case "byChildCount":
			sr.Tree = tree
			tree.ChildCount = params.Get(p)
		case "byParent":
			sr.Tree = tree
			tree.Parent = params.Get(p)
		case "byType":
			sr.Tree = tree
			tree.Type = v
		case "byLabel":
			sr.Tree = tree
			tree.IsSearch = true
			tree.Label = params.Get(p)
			sr.ResponseSize = 1
		case "byQuery":
			sr.Tree = tree
			tree.IsSearch = true
			tree.Query = params.Get(p)
			sr.ResponseSize = 1
		case "withFields":
			sr.Tree = tree
			tree.WithFields = strings.EqualFold(params.Get(p), "true")
		case "hasDigitalObject":
			sr.Tree = tree
			tree.HasDigitalObject = strings.EqualFold(params.Get(p), "true")
		case "paging":
			if strings.EqualFold(params.Get("paging"), "true") {
				sr.Tree = tree
				tree.IsPaging = true
			}
		case "pageMode":
			sr.Tree = tree
			tree.PageMode = params.Get(p)
		case "hasRestriction":
			sr.Tree = tree
			tree.HasRestriction = strings.EqualFold(params.Get(p), "true")
		case "byUnitID":
			sr.Tree = tree
			tree.UnitID = params.Get(p)
			tree.IsSearch = true
			tree.AllParents = strings.EqualFold(params.Get("allParents"), "true")
		case "byMimeType":
			sr.Tree = tree
			tree.MimeType = v
		case "cursorHint":
			sr.Tree = tree

			hint, err := strconv.ParseInt(params.Get(p), 10, 32)
			if err != nil {
				logConvErr(p, v, err)
				return sr, err
			}

			tree.CursorHint = int32(hint)
		case "page":
			page := params.Get(p)

			if page == "" {
				continue
			}

			pageInt, err := strconv.ParseInt(page, 10, 32)
			if err != nil {
				logConvErr(p, v, err)
				return sr, err
			}

			sr.Page = int32(pageInt)
		case "v1.mode":
			sr.V1Mode = strings.EqualFold(params.Get(p), "true")
		case "treePage":
			sr.Tree = tree
			tree.Page = []int32{}

			for _, page := range v {
				hint, err := strconv.ParseInt(page, 10, 32)
				if err != nil {
					logConvErr(p, v, err)
					return sr, err
				}

				tree.Page = append(tree.Page, int32(hint))
			}

			tree.IsPaging = true
		case "pageSize":
			sr.Tree = tree

			hint, err := strconv.ParseInt(params.Get(p), 10, 32)
			if err != nil {
				logConvErr(p, v, err)
				return sr, err
			}

			tree.PageSize = int32(hint)
		case "start":
			start, err := strconv.ParseInt(params.Get(p), 10, 32)
			if err != nil {
				logConvErr(p, v, err)
				return sr, err
			}
			sr.Start = int32(start)
		case "searchAfter":
			sa := make([]interface{}, 0)
			parts := strings.SplitN(params.Get(p), ",", 2)
			sortKey, _ := strconv.Atoi(parts[0])
			cLevel := parts[1]
			sa = append(sa, sortKey, cLevel)

			sb, err := getInterfaceBytes(sa)
			if err != nil {
				log.Printf(
					"unable to create bytes from interface %v; %s",
					domain.LogUserInput(fmt.Sprintf("%v", sa)),
					err,
				)
				return sr, err
			}
			sr.SearchAfter = sb
		}
	}

	if len(sr.GetQueryRefinement()) > 0 {
		for _, rq := range sr.GetQueryRefinement() {
			sr.Query += " AND (" + rq + ")"
		}
		sr.Query = strings.TrimPrefix(sr.Query, " AND ")
	}

	if sr.Tree != nil && sr.GetResponseSize() != int32(1) && sr.Page != 0 {
		rows := params.Get("rows")
		if rows == "" {
			// set hard max to number of nodes of 250
			sr.ResponseSize = int32(250)
		}
	}

	if sr.Page != 0 {
		sr.Start = getCursorFromPage(sr.GetPage(), sr.GetResponseSize())
	}

	if sr.SearchFields == "" {
		sr.SearchFields = "full_text"
	}

	return sr, nil
}

// cursor is zero based
func getCursorFromPage(page, responseSize int32) int32 {
	return (page * responseSize) - responseSize
}

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandSeq returns a random string of letters with the size of 'n'
func RandSeq(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

// FacetURIBuilder is used for creating facet filter fields
// TODO implement pop and push for creating facets links
type FacetURIBuilder struct {
	query             string
	filters           map[string]map[string]*QueryFilter
	ORBetweenFacets   bool
	facetMergeFilters map[string][]string
}

// NewFacetURIBuilder creates a builder for Facet links
func NewFacetURIBuilder(query string, filters []*QueryFilter) (*FacetURIBuilder, error) {
	fub := &FacetURIBuilder{
		query:             query,
		filters:           make(map[string]map[string]*QueryFilter),
		facetMergeFilters: map[string][]string{},
	}

	for _, f := range filters {
		if err := fub.AddFilter(f); err != nil {
			return nil, err
		}
	}

	return fub, nil
}

func (fub *FacetURIBuilder) AddFacetMergeFilters(filters []string) {
	for _, filter := range filters {
		var mergeFields []string
		parts := strings.Split(filter, ",")
		if len(parts) < 2 {
			continue
		}
		for _, p := range parts {
			mergeFields = append(mergeFields, strings.TrimSpace(p))
		}

		known, ok := fub.facetMergeFilters[filter]
		if ok {
			mergeFields = append(mergeFields, known...)
		}
		fub.facetMergeFilters[filter] = mergeFields
	}
}

func (fub *FacetURIBuilder) hasQueryFilter(field, value string) bool {
	if len(fub.filters) == 0 {
		return false
	}

	byField, ok := fub.filters[field]
	if !ok {
		return false
	}

	_, ok = byField[value]

	return ok
}

// AddFilter adds a QueryFilter to a multi dimensional map
func (fub *FacetURIBuilder) AddFilter(f *QueryFilter) error {
	child, ok := fub.filters[f.GetSearchLabel()]
	if !ok {
		child = map[string]*QueryFilter{}
		fub.filters[f.GetSearchLabel()] = child
	}

	child[f.GetValue()] = f

	return nil
}

// CreateFacetFilterURI generates a facetquery for each FacetLink and determines if it is selected
func (fub FacetURIBuilder) CreateFacetFilterURI(field, value string) (string, bool) {
	fields := []string{}
	var selected bool

	if fub.query != "" {
		fields = append(fields, fmt.Sprintf("q=%s", fub.query))
	}

	// todo replace with sort at builder level
	var filters []string
	for k := range fub.filters {
		filters = append(filters, k)
	}

	sort.Slice(filters, func(i, j int) bool { return filters[i] < filters[j] })

	for _, f := range filters {
		var filterValues []string

		values := fub.filters[f]
		for k := range values {
			filterValues = append(filterValues, k)
		}

		sort.Slice(filterValues, func(i, j int) bool { return filterValues[i] < filterValues[j] })

		for _, k := range filterValues {
			qf := values[k]

			if f == field && k == value {
				selected = true
				continue
			}

			// set tree filter type
			if strings.HasPrefix(f, "tree.") || strings.HasPrefix(qf.GetSearchLabel(), "tree.") {
				qf.Type = QueryFilterType_TREEITEM
			}

			filterKey := qfKey
			switch qf.GetType() {
			case QueryFilterType_EXISTS:
				fields = append(fields, fmt.Sprintf("qf.exist[]=%s", f))
				continue
			case QueryFilterType_ID:
				filterKey = qfIDKey
			case QueryFilterType_DATERANGE:
				filterKey = "qf.dateRange"
			case QueryFilterType_ISODATE:
				filterKey = "qf.date"
			case QueryFilterType_TREEITEM:
				filterKey = "qf.tree"
				f = strings.TrimPrefix(f, "tree.")
				if strings.HasPrefix(field, "tree.") {
					matchField := strings.TrimPrefix(field, "tree.")
					if f == matchField && k == value {
						selected = true
						continue
					}
				}
			}
			fields = append(fields, fmt.Sprintf("%s[]=%s:%s", filterKey, f, k))
		}
	}
	if !selected {
		key := qfKey
		switch {
		case strings.HasSuffix(field, ".id"):
			key = qfIDKey
			field = strings.TrimSuffix(field, ".id")
		}
		fields = append(fields, fmt.Sprintf("%s[]=%s:%s", key, field, value))
	}

	return strings.Join(fields, "&"), selected
}

// CreateFacetFilterQuery creates an elasticsearch Query to filter facets
// for the Facet Aggregation specified by 'filterfield'.
func (fub *FacetURIBuilder) CreateFacetFilterQuery(filterField string, andQuery bool) (*elastic.BoolQuery, error) {
	for k, filters := range fub.facetMergeFilters {
		_, known := fub.filters[k]
		if known {
			continue
		}
		merged := map[string]*QueryFilter{}
		for _, searchLabel := range filters {
			qfs := fub.filters[searchLabel]
			for nestedk, nestedv := range qfs {
				merged["m:"+nestedk+nestedv.GetSearchLabel()] = nestedv
			}
			delete(fub.filters, searchLabel)
		}
		fub.filters[k] = merged
	}

	q := elastic.NewBoolQuery()
	var fieldFilters []string
	for k := range fub.filters {
		fieldFilters = append(fieldFilters, k)
	}

	slog.Debug("fub output", "fieldFilters", fieldFilters, "filters", fub.filters, "mergeFilters", fub.facetMergeFilters)

	sort.Slice(fieldFilters, func(i, j int) bool { return fieldFilters[i] < fieldFilters[j] })

	for _, field := range fieldFilters {
		qfs := fub.filters[field]
		// skip filter field. this allows for all available options to be shown
		if filterField == field {
			continue
		}

		fieldQ := elastic.NewBoolQuery()

		var (
			active  bool
			filters []string
		)

		for k := range qfs {
			filters = append(filters, k)
		}

		sort.Slice(filters, func(i, j int) bool { return filters[i] < filters[j] })

		for _, k := range filters {
			qf := qfs[k]

			filterQuery, err := qf.ElasticFilter()
			if err != nil {
				return q, errors.Wrap(err, "Unable to build filter query")
			}

			switch andQuery {
			case true:
				fieldQ = fieldQ.Must(filterQuery)
			case false:
				fieldQ = fieldQ.Should(filterQuery)
			}
			active = true
		}

		if !active {
			continue
		}

		switch fub.ORBetweenFacets {
		case true:
			q = q.Should(fieldQ)
		default:
			q = q.Must(fieldQ)
		}
	}

	return q, nil
}

// BreadCrumbBuilder is a struct that holds all the information to build a BreadCrumb trail
type BreadCrumbBuilder struct {
	hrefPath []string
	crumbs   []*BreadCrumb
}

func (bcb *BreadCrumbBuilder) BreadCrumbs() []*BreadCrumb {
	return bcb.crumbs
}

// AppendBreadCrumb creates a BreadCrumb
func (bcb *BreadCrumbBuilder) AppendBreadCrumb(param string, qf *QueryFilter) {
	if qf.Hidden {
		// don't append hidden queries
		return
	}
	bc := &BreadCrumb{IsLast: true}

	switch param {
	case "query", "q":
		if qf.GetValue() != "" {
			bc.Display = qf.GetValue()
			bc.Href = fmt.Sprintf("q=%s", qf.GetValue())
			bc.Value = qf.GetValue()
			bcb.hrefPath = append(bcb.hrefPath, bc.Href)
		}
	case qfKeyList, qfKey, qfIDKey, qfIDKeyList:
		if !strings.HasSuffix(param, "[]") {
			param = fmt.Sprintf("%s[]", param)
		}

		qfs := fmt.Sprintf("%s:%s", qf.GetSearchLabel(), qf.GetValue())

		if qf.Exclude {
			qfs = fmt.Sprintf("-%s", qfs)
		}

		href := fmt.Sprintf("%s=%s", param, qfs)
		bc.Href = href
		if bcb.GetPath() != "" {
			bc.Href = bcb.GetPath() + "&" + bc.Href
		}
		bcb.hrefPath = append(bcb.hrefPath, href)
		bc.Display = qfs
		bc.Field = qf.GetSearchLabel()
		bc.Value = qf.GetValue()
	case qfExistList, "qf.exist":
		if !strings.HasSuffix(param, "[]") {
			param = fmt.Sprintf("%s[]", param)
		}
		qfs := qf.GetSearchLabel()
		href := fmt.Sprintf("%s=%s", param, qfs)
		bc.Href = href
		if bcb.GetPath() != "" {
			bc.Href = bcb.GetPath() + "&" + bc.Href
		}
		bcb.hrefPath = append(bcb.hrefPath, href)
		bc.Display = qfs
		// bc.Value = qf.GetValue()
		bc.Field = qf.GetSearchLabel()
	}

	last := bcb.GetLast()
	if last != nil {
		last.IsLast = false
	}

	bcb.crumbs = append(bcb.crumbs, bc)
}

// GetPath returns the path for the BreadCrumb
func (bcb *BreadCrumbBuilder) GetPath() string {
	return strings.Join(bcb.hrefPath, "&")
}

// GetLast returns the last BreadCrumb from the trail
func (bcb *BreadCrumbBuilder) GetLast() *BreadCrumb {
	if len(bcb.crumbs) == 0 {
		return nil
	}

	return bcb.crumbs[len(bcb.crumbs)-1]
}

// NewUserQuery creates an object with the user Query and the breadcrumbs
func (sr *SearchRequest) NewUserQuery() (*Query, *BreadCrumbBuilder, error) {
	q := &Query{}
	bcb := &BreadCrumbBuilder{}

	if sr.GetQuery() != "" {
		q.Terms = sr.GetQuery()
		bcb.AppendBreadCrumb("query", &QueryFilter{Value: sr.GetQuery()})
	}
	for _, qf := range sr.GetQueryFilter() {
		fieldKey := "qf[]"
		if qf.GetID() {
			fieldKey = qfIDKeyList
		}
		if qf.Exists {
			fieldKey = qfExistList
		}
		bcb.AppendBreadCrumb(fieldKey, qf)
	}
	q.BreadCrumbs = bcb.crumbs
	return q, bcb, nil
}

// ElasticQuery creates an ElasticSearch query from the Search Request
// This query can be passed into an elastic Search Object.
func (sr *SearchRequest) ElasticQuery() (elastic.Query, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("meta.docType", FragmentGraphDocType))
	orgQuery := elastic.NewBoolQuery().Should(elastic.NewTermQuery(sr.OrgIDKey, sr.OrgID))
	if len(sr.ContextIndex) != 0 {
		for _, index := range sr.ContextIndex {
			if index != "" {
				orgQuery = orgQuery.Should(elastic.NewTermQuery(sr.OrgIDKey, strings.TrimSuffix(index, "v2")))
			}
		}
	}

	query = query.Must(orgQuery)

	metaSpecPrefix := "meta.spec:"

	if sr.GetQuery() != "" {
		rawQuery := strings.Replace(sr.GetQuery(), "delving_spec:", metaSpecPrefix, 1)

		if strings.Contains(rawQuery, metaSpec) {
			all := []string{}
			for _, part := range strings.Split(rawQuery, " ") {
				if strings.HasPrefix(part, metaSpecPrefix) {
					spec := strings.TrimPrefix(part, metaSpecPrefix)
					query = query.Must(elastic.NewTermQuery(metaSpec, spec))
					continue
				}
				all = append(all, part)
			}
			rawQuery = strings.Join(all, " ")
		}
		if rawQuery != "" {
			fields := strings.Split(sr.SearchFields, ",")
			if len(fields) == 0 {
				fields = c.Config.ElasticSearch.SearchFields
			}
			qs, err := QueryFromSearchFields(rawQuery, fields...)
			if err != nil {
				return query, err
			}

			query = query.Must(qs)
		}
	}

	// add support for highlighting
	// add support for suggest queries

	if strings.HasPrefix(sr.GetSortBy(), "random") {
		randomFunc := elastic.NewRandomFunction()

		seeds := strings.Split(sr.GetSortBy(), "_")
		if len(seeds) == 2 {
			seed := seeds[1]
			randomFunc.Seed(seed)
		} else {
			seed := RandSeq(10)
			sr.SortBy = fmt.Sprintf("random_%s", seed)
			randomFunc.Seed(seed)
		}

		query := elastic.NewFunctionScoreQuery().
			AddScoreFunc(randomFunc).
			Query(query)
		return query, nil
	}

	if sr.Tree != nil && sr.Tree.GetFillTree() {
		parents := strings.Split(sr.Tree.GetLeaf(), "~")
		treeQuery := elastic.NewBoolQuery()
		var path string
		for idx, leaf := range parents {
			if idx == 0 {
				treeQuery = treeQuery.Should(elastic.NewMatchQuery(treeDepth, 1))
				path = leaf
				treeQuery = treeQuery.Should(elastic.NewTermQuery(treeLeaf, path))
				continue
			}
			path = fmt.Sprintf("%s~%s", path, leaf)
			treeQuery = treeQuery.Should(elastic.NewTermQuery(treeLeaf, path))
		}
		query = query.Must(treeQuery)

	}

	if sr.Tree != nil && sr.Tree.GetAllParents() {
		parents := strings.Split(sr.Tree.GetUnitID(), "~")
		treeQuery := elastic.NewBoolQuery()
		var path string
		for idx, leaf := range parents {
			if idx == 0 {
				path = leaf
				treeQuery = treeQuery.Should(elastic.NewTermQuery(treeCLevel, path))
				continue
			}

			path = fmt.Sprintf("%s~%s", path, leaf)
			treeQuery = treeQuery.Should(elastic.NewTermQuery(treeCLevel, path))
		}
		query = query.Must(treeQuery)
	}

	// todo move this into a separate function
	if sr.Tree != nil && !sr.Tree.GetFillTree() && !sr.Tree.GetAllParents() {
		// exclude description
		query = query.Must(elastic.NewMatchQuery(metaTags, "ead"))
		if sr.Tree.GetLeaf() != "" {
			query = query.Must(elastic.NewTermQuery(treeLeaf, sr.Tree.GetLeaf()))
		}
		if sr.Tree.GetParent() != "" {
			query = query.Must(elastic.NewTermQuery("tree.parent", sr.Tree.GetParent()))
		}
		if sr.Tree.GetChildCount() != "" {
			query = query.Must(elastic.NewMatchQuery("tree.childCount", sr.Tree.GetChildCount()))
		}
		// todo add filtering for hasRestriction and HasDigitalObject
		if sr.Tree.HasRestriction {
			query = query.Must(elastic.NewMatchQuery("tree.hasRestriction", "true"))
		}
		if sr.Tree.HasDigitalObject {
			query = query.Must(elastic.NewMatchQuery("tree.hasDigitalObject", "true"))
		}

		if sr.Tree.GetQuery() != "" {
			q, err := QueryFromSearchFields(sr.Tree.GetQuery(), c.Config.EAD.SearchFields...)
			if err != nil {
				return query, err
			}

			query = query.Must(q)
		}
		if sr.Tree.GetLabel() != "" {
			q := elastic.NewQueryStringQuery(escapeRawQuery(sr.Tree.GetLabel()))
			q.DefaultOperator("and")

			q = q.Field("tree.label")
			if !isAdvancedSearch(sr.Tree.GetLabel()) {
				q = q.MinimumShouldMatch(c.Config.ElasticSearch.MinimumShouldMatch)
			}

			query = query.Must(q)
		}
		if sr.Tree.GetUnitID() != "" {
			if strings.HasPrefix(sr.Tree.GetUnitID(), "@") {
				query = query.Must(elastic.NewTermQuery(treeCLevel, sr.Tree.GetUnitID()))
			} else {
				query = query.Must(elastic.NewTermQuery("tree.unitID", sr.Tree.GetUnitID()))
			}
		}
		switch len(sr.Tree.GetDepth()) {
		case 1:
			query = query.Must(elastic.NewMatchQuery(treeDepth, sr.Tree.GetDepth()[0]))
		case 0:
		default:
			q := elastic.NewBoolQuery()
			for _, d := range sr.Tree.GetDepth() {
				q = q.Should(elastic.NewTermQuery(treeDepth, d))
			}
			query = query.Must(q)
			sr.Tree.FillTree = true
		}
		switch len(sr.Tree.GetType()) {
		case 1:
			query = query.Must(elastic.NewMatchQuery("tree.type", sr.Tree.GetType()[0]))
		case 0:
		default:
			q := elastic.NewBoolQuery()
			for _, d := range sr.Tree.GetType() {
				q = q.Should(elastic.NewTermQuery("tree.type", d))
			}
			query = query.Must(q)
			sr.Tree.FillTree = true
		}
		switch len(sr.Tree.GetMimeType()) {
		case 1:
			query = query.Must(elastic.NewMatchQuery("tree.mimeType", sr.Tree.GetMimeType()[0]))
		case 0:
		default:
			q := elastic.NewBoolQuery()
			for _, d := range sr.Tree.GetMimeType() {
				q = q.Should(elastic.NewTermQuery("tree.mimeType", d))
			}
			query = query.Must(q)
			sr.Tree.FillTree = true
		}
	}

	// TODO: add support for hidden query filters
	if len(sr.HiddenQueryFilter) > 0 {
		fub, err := NewFacetURIBuilder("", sr.HiddenQueryFilter)
		if err != nil {
			return nil, err
		}
		fub.ORBetweenFacets = sr.ORBetweenFacets
		fub.AddFacetMergeFilters(sr.FacetMergeFilter)

		hiddenFilterQuery, err := fub.CreateFacetFilterQuery("", sr.FacetAndBoolType)
		if err != nil {
			return nil, err
		}

		query = query.Filter(hiddenFilterQuery)
	}

	return query, nil
}

// isAdvancedSearch checks if the query contains Lucene QueryString
// advanced search query syntax.
func isAdvancedSearch(query string) bool {
	parts := strings.Fields(query)
	for _, p := range parts {
		switch {
		case p == "AND":
			return true
		case p == "OR":
			return true
		case strings.HasPrefix(p, "-"):
			return true
		case strings.HasPrefix(p, "+"):
			return true
		}
	}
	return false
}

// Aggregations returns the aggregations for the SearchRequest
func (sr *SearchRequest) Aggregations(fub *FacetURIBuilder) (map[string]elastic.Aggregation, error) {
	aggs := map[string]elastic.Aggregation{}

	for _, facetField := range sr.FacetField {
		if sr.FacetLimit != 0 {
			facetField.Size = sr.FacetLimit
		}

		agg, err := sr.CreateAggregationBySearchLabel(resourcesEntries, facetField, fub)
		if err != nil {
			return nil, err
		}
		fieldName := facetField.GetField()
		if facetField.ById {
			fieldName = fmt.Sprintf("%s.id", fieldName)
		}
		aggs[fieldName] = agg
	}
	return aggs, nil
}

func createFieldedSubQuery(field, userQuery string, boost float64) elastic.Query {
	fieldTermQuery := elastic.NewTermQuery(entriesSearchLabel, field)

	uq := elastic.NewQueryStringQuery(escapeRawQuery(userQuery))
	uq = uq.DefaultOperator("and")
	switch {
	case boost < 0, boost > 0:
		uq = uq.FieldWithBoost(entriesValue, boost)
	default:
		uq = uq.Field(entriesValue)
	}

	return elastic.NewBoolQuery().Must(fieldTermQuery, uq)
}

// CreateAggregationBySearchLabel creates Elastic aggregations for the nested fragment resources
func (sr *SearchRequest) CreateAggregationBySearchLabel(path string, facet *FacetField, fub *FacetURIBuilder) (elastic.Aggregation, error) {
	return CreateAggregationBySearchLabel(path, facet, sr.FacetAndBoolType, fub)
}

// CreateAggregationBySearchLabel creates Elastic aggregations for the nested fragment resources
func CreateAggregationBySearchLabel(path string, facet *FacetField, facetAndBoolType bool, fub *FacetURIBuilder) (elastic.Aggregation, error) {
	nestedPath := fmt.Sprintf("%s.searchLabel", path)
	fieldTermQuery := elastic.NewTermQuery(nestedPath, facet.GetField())

	entryKey := "@value.keyword"
	if facet.GetById() {
		entryKey = "@id"
	}

	termAggPath := fmt.Sprintf("%s.%s", path, entryKey)

	labelAgg := elastic.NewTermsAggregation().Field(termAggPath).Size(int(facet.GetSize()))

	if facet.GetByName() {
		labelAgg = labelAgg.OrderByKey(facet.GetAsc())
	} else {
		labelAgg = labelAgg.OrderByCount(facet.GetAsc())
	}

	// Add Filters as nested path
	facetFilters, err := fub.CreateFacetFilterQuery(facet.GetField(), facetAndBoolType)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create FacetFilterQuery")
	}

	facetFilterAgg := elastic.NewFilterAggregation().
		Filter(facetFilters)

	switch facet.GetType() {
	case FacetType_METATAGS:
		tagAgg := elastic.NewTermsAggregation().
			Field(metaTags).
			Size(int(facet.GetSize()))

		facetFilterAgg = facetFilterAgg.SubAggregation("object", tagAgg)

	case FacetType_TREEFACET:
		treeAgg := elastic.NewTermsAggregation().
			Field(facet.Field).
			Size(int(facet.GetSize()))

		facetFilterAgg = facetFilterAgg.SubAggregation("object", treeAgg)
	case FacetType_TAGS:
		tagAgg := elastic.NewTermsAggregation().
			Field(fmt.Sprintf("%s.tags", resourcesEntries)).
			Size(int(facet.GetSize()))

		filterAgg := elastic.NewFilterAggregation().
			Filter(elastic.NewBoolQuery()).
			SubAggregation("value", tagAgg)

		innerAgg := elastic.NewNestedAggregation().
			Path(path).
			SubAggregation("inner", filterAgg)
		facetFilterAgg = facetFilterAgg.SubAggregation("filter", innerAgg)
	case FacetType_FIELDS:
		tagAgg := elastic.NewTermsAggregation().
			Field(fmt.Sprintf("%s.searchLabel", resourcesEntries)).
			Size(int(facet.GetSize()))

		filterAgg := elastic.NewFilterAggregation().
			Filter(elastic.NewBoolQuery()).
			SubAggregation("value", tagAgg)

		innerAgg := elastic.NewNestedAggregation().
			Path(path).
			SubAggregation("inner", filterAgg)
		facetFilterAgg = facetFilterAgg.SubAggregation("filter", innerAgg)

	case FacetType_MINMAX:
		field := fmt.Sprintf("resources.entries.%s", facet.GetAggField())
		minAgg := elastic.NewMinAggregation().Field(field)
		maxAgg := elastic.NewMaxAggregation().Field(field)

		filterAgg := elastic.NewFilterAggregation().
			Filter(fieldTermQuery).
			SubAggregation("minval", minAgg).
			SubAggregation("maxval", maxAgg)

		innerAgg := elastic.NewNestedAggregation().
			Path(path).
			SubAggregation("inner", filterAgg)
		facetFilterAgg = facetFilterAgg.SubAggregation("filter", innerAgg)
	case FacetType_HISTOGRAM:
		if facet.DateInterval == "" {
			facet.DateInterval = "1y"
		}

		field := fmt.Sprintf("resources.entries.%s", "isoDate")
		minAgg := elastic.NewMinAggregation().Field(field)
		maxAgg := elastic.NewMaxAggregation().Field(field)
		histAgg := elastic.NewDateHistogramAggregation().
			Field(field).
			FixedInterval(facet.DateInterval)

		filterAgg := elastic.NewFilterAggregation().
			Filter(fieldTermQuery).
			SubAggregation("minval", minAgg).
			SubAggregation("maxval", maxAgg).
			SubAggregation("histogram", histAgg)

		innerAgg := elastic.NewNestedAggregation().
			Path(path).
			SubAggregation("inner", filterAgg)
		facetFilterAgg = facetFilterAgg.SubAggregation("filter", innerAgg)
	default:
		var pathFilter elastic.Query
		pathFilter = fieldTermQuery
		if facet.GetLanguage() != "" {
			langPath := fmt.Sprintf("%s.@language", path)
			langTermQuery := elastic.NewTermQuery(langPath, facet.GetLanguage())
			pathFilter = elastic.NewBoolQuery().Must(
				fieldTermQuery, langTermQuery,
			)
		}

		filterAgg := elastic.NewFilterAggregation().
			Filter(pathFilter).
			SubAggregation("value", labelAgg)
		testAgg := elastic.NewNestedAggregation().Path(path)
		testAgg = testAgg.SubAggregation("inner", filterAgg)
		facetFilterAgg = facetFilterAgg.SubAggregation("filter", testAgg)
	}

	return facetFilterAgg, nil
}

func getInterface(bts []byte, data interface{}) error {
	buf := bytes.NewBuffer(bts)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(data)
	return err
}

func getInterfaceBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SearchRequestToHex converts the SearchRequest to a hex string
func (sr *SearchRequest) SearchRequestToHex() (string, error) {
	output, err := proto.Marshal(sr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", output), nil
}

// DeepCopy create a deepCopy of the SearchRequest.
// This is used to calculate next NextScrollID values without change the current values of the request.
func (sr *SearchRequest) DeepCopy() (*SearchRequest, error) {
	output, err := proto.Marshal(sr)
	if err != nil {
		return nil, err
	}
	newSr := &SearchRequest{}
	err = proto.Unmarshal(output, newSr)
	if err != nil {
		return nil, err
	}

	return newSr, nil
}

func (sr *SearchRequest) CreateBinKey(key interface{}) ([]byte, error) {
	return getInterfaceBytes(key)
}

// DecodeSearchAfter returns an interface array decoded from []byte
func (sr *SearchRequest) DecodeSearchAfter() ([]interface{}, error) {
	var sa []interface{}
	err := getInterface(sr.SearchAfter, &sa)
	if err != nil {
		log.Printf("Unable to decode interface: %s", err)
		return sa, errors.Wrap(err, "Unable to decode interface")
	}
	return sa, nil
}

// ElasticSearchService creates the elastic SearchService for execution
func (sr *SearchRequest) ElasticSearchService(ec *elastic.Client) (*elastic.SearchService, *FacetURIBuilder, error) {
	idSort := elastic.NewFieldSort("meta.hubID")
	var fieldSort *elastic.FieldSort

	switch {
	case sr.Tree != nil && sr.GetSortBy() == "":
		fieldSort = elastic.NewFieldSort("tree.sortKey")
	case sr.GetSortBy() == "_score":
		fieldSort = elastic.NewFieldSort("_score")
	case strings.HasPrefix(sr.GetSortBy(), "random"), sr.GetSortBy() == "":
		fieldSort = elastic.NewFieldSort("_score").Desc()
	case strings.HasPrefix(sr.GetSortBy(), "tree."), strings.HasPrefix(sr.GetSortBy(), "meta."):
		fieldSort = elastic.NewFieldSort(sr.GetSortBy())
	case strings.HasSuffix(sr.GetSortBy(), "_int"):
		field := strings.TrimSuffix(sr.GetSortBy(), "_int")
		sortNestedQuery := elastic.NewTermQuery(entriesSearchLabel, field)
		fieldSort = elastic.NewFieldSort("resources.entries.integer").
			NestedPath(resourcesEntries).
			NestedFilter(sortNestedQuery)
	default:
		sortNestedQuery := elastic.NewTermQuery(entriesSearchLabel, sr.GetSortBy())
		nestedSort := elastic.NewNestedSort(resourcesEntries).
			// Path(resourcesEntries).
			Filter(sortNestedQuery)

		fieldSort = elastic.NewFieldSort("resources.entries.@value.keyword").NestedSort(nestedSort)
	}

	if sr.SortAsc {
		fieldSort = fieldSort.Asc()
	} else {
		fieldSort = fieldSort.Desc()
	}

	if sr.Tree != nil && sr.GetResponseSize() != 1 {
		sr.ResponseSize = int32(1000)
		if sr.Tree.IsPaging {
			sr.ResponseSize = sr.Tree.TreePagingSize()
		}

		if len(sr.Tree.Type) > 0 {
			sr.ResponseSize = int32(c.Config.ElasticSearch.MaxTreeSize)
		}
	}

	indices := []string{c.Config.ElasticSearch.GetIndexName(sr.OrgID)}
	if len(sr.ContextIndex) != 0 {
		indices = append(indices, sr.ContextIndex...)
	}

	log.Printf("configured indices: %#v", indices)

	s := ec.Search().
		Index(indices...).
		TrackTotalHits(c.Config.ElasticSearch.TrackTotalHits).
		Preference(sr.GetSessionID()).
		Size(int(sr.GetResponseSize()))

	// This section is used to return the tree page section of 250 nodes starting the current.
	if sr.Tree != nil && sr.Tree.IsPaging && !sr.Tree.IsSearch {
		s = s.SortBy(fieldSort)
		_, current, _ := sr.Tree.PreviousCurrentNextPage()
		searchAfterPage := current - int32(1)
		searchAfterCursor := (searchAfterPage * sr.Tree.GetPageSize())
		if searchAfterPage > 0 {
			s = s.SearchAfter(searchAfterCursor)
		}
	} else {
		// This section is used to get the search hit with size 1.
		s = s.SortBy(fieldSort, idSort)
		if len(sr.SearchAfter) != 0 && sr.CollapseOn == "" {
			sa, err := sr.DecodeSearchAfter()
			if err != nil {
				return nil, nil, err
			}
			if c.Config.ElasticSearch.EnableSearchAfter {
				s = s.SearchAfter(sa...)
			} else {
				s = s.From(int(sr.GetStart()))
			}
		}
	}

	// if sr.GetPage() > int32(0) && sr.GetStart() > int32(0) {
	s = s.From(int(sr.GetStart()))
	// }

	query, err := sr.ElasticQuery()
	if err != nil {
		log.Println("Unable to build the query result.")
		return s, nil, err
	}

	s = s.Query(query)

	fub, err := NewFacetURIBuilder(sr.GetQuery(), sr.GetQueryFilter())
	if err != nil {
		log.Println("Unable to FacetURIBuilder")
		return s, nil, err
	}
	fub.ORBetweenFacets = sr.ORBetweenFacets
	fub.AddFacetMergeFilters(sr.FacetMergeFilter)

	// Add post filters
	postFilter, err := fub.CreateFacetFilterQuery("", sr.FacetAndBoolType)
	if err != nil {
		log.Printf("unable to create postfilter: %#v", err)
		return s, nil, err
	}
	s = s.PostFilter(postFilter)

	if sr.CollapseOn != "" {
		collapseSize := 5
		if sr.CollapseSize != 0 {
			collapseSize = int(sr.CollapseSize)
		}
		b := elastic.NewCollapseBuilder(sr.CollapseOn).
			InnerHit(elastic.NewInnerHit().Name("collapse").Size(collapseSize)).
			MaxConcurrentGroupRequests(4)
		s = s.Collapse(b)
		s = s.FetchSource(false)

		collapseCountAgg := elastic.NewCardinalityAggregation().
			Field(sr.CollapseOn)

		countFilterAgg := elastic.NewFilterAggregation().
			Filter(postFilter).
			SubAggregation("collapseCount", collapseCountAgg)

		s = s.Aggregation("counts", countFilterAgg)
	}

	if sr.Peek != "" {
		facetField := &FacetField{Field: sr.Peek, Size: int32(100)}
		if sr.Peek == metaTags {
			facetField.Type = FacetType_METATAGS
		}
		agg, aggErr := sr.CreateAggregationBySearchLabel(resourcesEntries, facetField, fub)
		if err != nil {
			return nil, nil, aggErr
		}

		s = s.Size(0)
		s = s.Aggregation(sr.Peek, agg)
		sr.FacetField = append(sr.FacetField, facetField)
		return s.Query(query), nil, nil
	}

	if sr.Tree != nil {
		fsc := elastic.NewFetchSourceContext(true)
		fsc.Include("tree")
		fsc.Exclude("tree.rawContent")

		if sr.Tree.WithFields {
			fsc.Include("resources")
		}
		s = s.FetchSourceContext(fsc)
	}

	// Add aggregations
	if !sr.V1Mode {
		if sr.Paging {
			return s.Query(query), nil, err
		}
	}

	aggs, err := sr.Aggregations(fub)
	if err != nil {
		log.Println("Unable to build the Aggregations.")
		return s, nil, err
	}
	for facetField, agg := range aggs {
		s = s.Aggregation(facetField, agg)
	}

	return s.Query(query), fub, err
}

// NewScrollPager returns a ScrollPager with defaults set
func NewScrollPager() *ScrollPager {
	sp := &ScrollPager{}
	sp.Total = 0
	sp.Cursor = 0

	return sp
}

type ScrollType int

const (
	ScrollNext ScrollType = iota
	ScrollPrev
)

func newSearchRequestScrollPage(sr *SearchRequest, position ScrollType, total int64) (*SearchRequest, error) {
	copySr, err := sr.DeepCopy()
	if err != nil {
		return nil, err
	}

	if total == 0 {
		return sr, nil
	}

	// set the next or prev cursor
	if position == ScrollNext {
		copySr.Start = copySr.GetStart() + copySr.GetResponseSize()
	}

	if position == ScrollPrev {
		copySr.Start = copySr.GetStart() - copySr.GetResponseSize()
	}

	// if paging set next page
	if copySr.Tree != nil && copySr.Tree.IsPaging && !copySr.Tree.IsSearch {
		prev, _, next := sr.Tree.PreviousCurrentNextPage()

		if position == ScrollNext {
			copySr.Tree.Page = []int32{next}
		}

		if position == ScrollPrev {
			copySr.Tree.Page = []int32{prev}
		}
	}

	return copySr, nil
}

func (sr *SearchRequest) ScrollPagers(total int64) (*ScrollPager, error) {
	sp := NewScrollPager()
	sp.Cursor = sr.GetStart()
	sp.Rows = sr.GetResponseSize()
	sp.Total = total
	if sr.CalculatedTotal != 0 {
		sp.Total = sr.CalculatedTotal
	}
	prev, err := newSearchRequestScrollPage(sr, ScrollPrev, total)
	if err != nil {
		return nil, err
	}
	next, err := newSearchRequestScrollPage(sr, ScrollNext, total)
	if err != nil {
		return nil, err
	}

	if int64(next.GetStart()) < total {
		sp.NextScrollID, err = next.SearchRequestToHex()
		if err != nil {
			return nil, err
		}
	}

	if prev.GetStart() >= 0 {
		sp.PreviousScrollID, err = prev.SearchRequestToHex()
		if err != nil {
			return nil, err
		}
	}

	return sp, nil
}

func qfSplit(r rune) bool {
	return r == ']' || r == '['
}

func validateTypeClass(tc string) string {
	if tc == "a" {
		return ""
	}
	return tc
}

// NewQueryFilter parses the filter string and creates a QueryFilter object
func NewQueryFilter(filter string) (*QueryFilter, error) {
	qf := &QueryFilter{}

	// TODO serialize
	if strings.HasPrefix(filter, "{") {
		err := json.Unmarshal([]byte(filter), &qf)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to unmarshal query filter")
		}
		return qf, nil
	}

	if strings.HasPrefix(filter, "-") {
		qf.Exclude = true
		filter = strings.TrimPrefix(filter, "-")
	}

	if strings.HasPrefix(filter, "~") {
		qf.Hidden = true
		filter = strings.TrimPrefix(filter, "~")
	}

	// fill empty type classes
	filter = strings.Replace(filter, "[]", `[a]`, -1)

	parts := strings.SplitN(filter, ":", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("no query field specified in: %s", filter)
	}
	qf.Value = parts[1]

	filterKey := parts[0]
	if strings.HasSuffix(filterKey, ".id") {
		qf.ID = true
		filterKey = strings.TrimSuffix(filterKey, ".id")
	}

	parts = strings.FieldsFunc(filterKey, qfSplit)
	switch len(parts) {
	case 1:
		qf.SearchLabel = filterKey
	case 2:
		qf.SearchLabel = parts[1]
		qf.TypeClass = validateTypeClass(parts[0])
	case 3:
		qf.SearchLabel = parts[2]
		qf.TypeClass = validateTypeClass(parts[1])
		qf.Level2 = &ContextQueryFilter{SearchLabel: parts[0]}
	case 4:
		qf.SearchLabel = parts[3]
		qf.TypeClass = validateTypeClass(parts[2])
		qf.Level2 = &ContextQueryFilter{SearchLabel: parts[1], TypeClass: validateTypeClass(parts[0])}
	case 5:
		qf.SearchLabel = parts[4]
		qf.TypeClass = validateTypeClass(parts[3])
		qf.Level2 = &ContextQueryFilter{SearchLabel: parts[2], TypeClass: validateTypeClass(parts[1])}
		qf.Level1 = &ContextQueryFilter{SearchLabel: parts[0]}
	case 6:
		qf.SearchLabel = parts[5]
		qf.TypeClass = validateTypeClass(parts[4])
		qf.Level2 = &ContextQueryFilter{SearchLabel: parts[3], TypeClass: validateTypeClass(parts[2])}
		qf.Level1 = &ContextQueryFilter{SearchLabel: parts[1], TypeClass: validateTypeClass(parts[0])}
	}

	return qf, nil
}

// AsString returns the QueryFilter formatted as a string
func (qf *QueryFilter) AsString() string {
	base := fmt.Sprintf("[%s]%s:%s", qf.GetTypeClass(), qf.GetSearchLabel(), qf.GetValue())
	level2 := ""
	if qf.GetLevel2() != nil {
		level2 = fmt.Sprintf("[%s]%s", qf.Level2.GetTypeClass(), qf.Level2.GetSearchLabel())
	}
	level1 := ""
	if qf.GetLevel1() != nil {
		level1 = fmt.Sprintf("[%s]%s", qf.Level1.GetTypeClass(), qf.Level1.GetSearchLabel())
	}
	return fmt.Sprintf("%s%s%s", level1, level2, base)
}

// TypeClassAsURI resolves the type class formatted as "prefix_label" as fully qualified URI
func TypeClassAsURI(uri string) (string, error) {
	parts := strings.SplitN(uri, "_", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("wrong shorhand for TypeClass is defined; got %s", uri)
	}

	label := parts[1]
	base, err := rdf.DefaultNamespaceManager.GetWithBase(parts[0])
	if err != nil {
		return "", fmt.Errorf("namespace for prefix %s is unknown; %w", parts[0], err)
	}

	if strings.HasSuffix(base.URI, "#") || strings.HasSuffix(base.URI, "/") {
		return fmt.Sprintf("%s%s", base, label), nil
	}

	return fmt.Sprintf("%s/%s", base, label), nil
}

func (qf *QueryFilter) SetExclude(q *elastic.BoolQuery, qs ...elastic.Query) *elastic.BoolQuery {
	switch qf.Exclude {
	case true:
		q = q.MustNot(qs...)
	case false:
		q = q.Must(qs...)
	}
	return q
}

// ElasticFilter creates an elasticsearch filter from the QueryFilter
func (qf *QueryFilter) ElasticFilter() (elastic.Query, error) {
	nestedBoolQuery := elastic.NewBoolQuery()

	// resource.entries queries
	labelQ := elastic.NewTermQuery(entriesSearchLabel, qf.SearchLabel)

	if qf.Exists {
		qs := elastic.NewBoolQuery()
		qs = qf.SetExclude(qs, labelQ)
		nq := elastic.NewNestedQuery(resourcesEntries, qs)
		return nq, nil
	}

	// object queries
	switch qf.SearchLabel {
	case "spec", "delving_spec", "delving_spec.raw", metaSpec:
		qf.SearchLabel = c.Config.ElasticSearch.SpecKey
		return qf.SetExclude(
			elastic.NewBoolQuery(),
			elastic.NewTermQuery(qf.SearchLabel, qf.Value),
		), nil
	case metaTags, "meta.tag":
		qf.SearchLabel = metaTags
		return qf.SetExclude(
			elastic.NewBoolQuery(),
			elastic.NewTermQuery(qf.SearchLabel, qf.Value),
		), nil
	}

	// support meta. and tree. filter queries
	switch {
	case strings.HasPrefix(qf.SearchLabel, "meta."), strings.HasPrefix(qf.SearchLabel, "tree."):
		return qf.SetExclude(
			elastic.NewBoolQuery(),
			elastic.NewTermQuery(qf.SearchLabel, qf.Value),
		), nil
	}

	// nested queries
	var fieldQuery elastic.Query

	switch qf.GetType() {
	case QueryFilterType_DATERANGE:
		fieldKey := "resources.entries.dateRange"
		rq := elastic.NewRangeQuery(fieldKey)
		if qf.Gte != "" {
			rq = rq.Gte(qf.Gte)
		}
		if qf.Lte != "" {
			rq = rq.Lte(qf.Lte)
		}
		fieldQuery = rq
	case QueryFilterType_ISODATE:
		fieldKey := "resources.entries.isoDate"
		fieldQuery = elastic.NewTermQuery(fieldKey, qf.Value)
	case QueryFilterType_TREEITEM:
		q := elastic.NewBoolQuery()
		return qf.SetExclude(
			q,
			elastic.NewTermQuery(qf.SearchLabel, qf.Value),
		), nil
	case QueryFilterType_ENTRYTAG:
		fieldQuery = elastic.NewTermQuery("resources.entries.tags", qf.Value)
		qs := elastic.NewBoolQuery()
		qs = qf.SetExclude(qs, fieldQuery)
		return elastic.NewNestedQuery(resourcesEntries, qs), nil
	case QueryFilterType_SEARCHLABEL:
		fieldQuery = elastic.NewTermQuery("resources.entries.searchLabel", qf.Value)
		qs := elastic.NewBoolQuery()
		qs = qf.SetExclude(qs, fieldQuery)
		return elastic.NewNestedQuery(resourcesEntries, qs), nil
	default:
		fieldKey := "resources.entries.@value.keyword"
		if qf.ID {
			fieldKey = "resources.entries.@id"
		}
		fieldQuery = elastic.NewTermQuery(fieldKey, qf.Value)
	}

	qs := elastic.NewBoolQuery()
	qs = qf.SetExclude(qs, labelQ, fieldQuery)
	nq := elastic.NewNestedQuery(resourcesEntries, qs)

	if qf.GetTypeClass() == "" && qf.GetLevel2() == nil {
		return nq, nil
	}

	nestedBoolQuery = nestedBoolQuery.Must(nq)
	mainQuery := elastic.NewNestedQuery("resources", nestedBoolQuery)

	// resource.types query
	if qf.GetTypeClass() != "" {
		tc, err := TypeClassAsURI(qf.GetTypeClass())
		if err != nil {
			return mainQuery, errors.Wrap(err, "Unable to convert TypeClass from shorthand to URI")
		}
		typeQuery := elastic.NewTermQuery("resources.types", tc)
		nestedBoolQuery = nestedBoolQuery.Must(typeQuery)
	}

	// TODO implement this with recursion later
	// resource.context queries
	if qf.GetLevel2() != nil {
		level2 := qf.GetLevel2()
		levelq := elastic.NewBoolQuery()
		if level2.GetTypeClass() != "" {
			tc, err := TypeClassAsURI(level2.GetTypeClass())
			if err != nil {
				return mainQuery, errors.Wrap(err, "Unable to convert TypeClass from shorthand to URI")
			}
			classQuery := elastic.NewTermQuery("resources.context.SubjectClass", tc)
			levelq = levelq.Must(classQuery)
		}
		labelQ := elastic.NewTermQuery("resources.context.SearchLabel", level2.SearchLabel)
		lq := elastic.NewNestedQuery("resources.context", levelq.Must(labelQ))
		nestedBoolQuery = nestedBoolQuery.Must(lq)
	}

	return mainQuery, nil
}

// Equal determines equality between Query Filters
func (qf *QueryFilter) Equal(oqf *QueryFilter) bool {
	// TODO replace with property by property comparison
	return qf.AsString() == oqf.AsString()
}

type QueryFilterConfig struct {
	Hidden     bool
	IsIDFilter bool
}

func (sr *SearchRequest) appendQueryFilter(qf *QueryFilter) {
	switch qf.Hidden {
	case true:
		sr.HiddenQueryFilter = append(sr.HiddenQueryFilter, qf)
	default:
		sr.QueryFilter = append(sr.QueryFilter, qf)
	}
}

// AddQueryFilter adds a QueryFilter to the SearchRequest
// The raw query from the QueryString are added here. This function converts
// this string to a QueryFilter.
func (sr *SearchRequest) AddQueryFilter(filter string, cfg QueryFilterConfig) error {
	if filter == "" {
		// continue if string is empty
		return nil
	}

	qf, err := NewQueryFilter(filter)
	if err != nil {
		return err
	}
	qf.Type = QueryFilterType_TEXT
	if cfg.IsIDFilter {
		qf.ID = true
		qf.Type = QueryFilterType_ID
	}

	if cfg.Hidden {
		qf.Hidden = cfg.Hidden
	}

	switch qf.SearchLabel {
	case "tag", "tags":
		qf.Type = QueryFilterType_ENTRYTAG
	case "searchLabel":
		qf.Type = QueryFilterType_SEARCHLABEL
	}

	// todo replace later with map lookup that can be reused
	for _, v := range sr.QueryFilter {
		if cmp.Equal(qf, v) {
			return nil
		}
	}

	sr.appendQueryFilter(qf)
	return nil
}

// NewTreeFilter creates QueryFilter for Tree
func NewTreeFilter(filter string) (*QueryFilter, error) {
	if !strings.HasPrefix(filter, "tree.") {
		filter = fmt.Sprintf("tree.%s", filter)
	}

	qf, err := NewQueryFilter(filter)
	if err != nil {
		return nil, err
	}

	qf.Type = QueryFilterType_TREEITEM
	return qf, nil
}

// AddTreeFilter extracts a start and end date from the QueryFilter.Value
// add appends it to the QueryFilter Array.
func (sr *SearchRequest) AddTreeFilter(filter string) error {
	qf, err := NewTreeFilter(filter)
	if err != nil {
		return err
	}

	sr.QueryFilter = append(sr.QueryFilter, qf)
	return nil
}

// AddDateFilter adds a filter for Date Querying.
func (sr *SearchRequest) AddDateFilter(filter string, cfg QueryFilterConfig) error {
	qf, err := NewQueryFilter(filter)
	if err != nil {
		return err
	}
	qf.Type = QueryFilterType_ISODATE
	qf.Hidden = cfg.Hidden

	sr.appendQueryFilter(qf)
	return nil
}

// AddDateRangeFilter extracts a start and end date from the QueryFilter.Value
// add appends it to the QueryFilter Array.
func (sr *SearchRequest) AddDateRangeFilter(filter string, cfg QueryFilterConfig) error {
	qf, err := NewDateRangeFilter(filter, cfg)
	if err != nil {
		return err
	}

	sr.appendQueryFilter(qf)
	return nil
}

// NewDateRangeFilter creates a new QueryFilter from the input string.
func NewDateRangeFilter(filter string, cfg QueryFilterConfig) (*QueryFilter, error) {
	// sometimes javascript front-ends send null for empty filters, so these need to be removed.
	if strings.Contains(filter, "null") {
		filter = strings.ReplaceAll(filter, "null", "")
	}

	qf, err := NewQueryFilter(filter)
	if err != nil {
		return nil, err
	}
	qf.Hidden = cfg.Hidden

	if qf.Value == "~" || qf.Value == "" {
		return nil, fmt.Errorf(
			"date range %s cannot be without a start and end",
			filter,
		)
	}

	qf.Type = QueryFilterType_DATERANGE

	parts := strings.Split(qf.Value, "~")
	if len(parts) != 2 {
		return nil, fmt.Errorf(
			"the date range value %s must include ~ to separate start and end",
			qf.Value,
		)
	}

	if parts[0] != "" {
		qf.Gte = parts[0]
	}

	if parts[1] != "" {
		qf.Lte = parts[1]
	}

	return qf, nil
}

// AddFieldExistFilter adds a query to filter on records where this fields exists.
// This query for now works on any field level. It is not possible to specify
// context path.
func (sr *SearchRequest) AddFieldExistFilter(filter string, cfg QueryFilterConfig) error {
	qf := &QueryFilter{}
	qf.Exists = true
	qf.Hidden = cfg.Hidden
	qf.Type = QueryFilterType_EXISTS
	qf.SearchLabel = filter
	sr.appendQueryFilter(qf)
	return nil
}

// RemoveQueryFilter removes a QueryFilter from the SearchRequest
// The raw query from the QueryString are added here.
func (sr *SearchRequest) RemoveQueryFilter(filter string) error {
	return nil
}

func getKeyAsString(raw json.RawMessage) string {
	return strings.Trim(
		string(raw),
		"\"",
	)
}

// DecodeFacets decodes the elastic aggregations in the SearchResult to fragments.QueryFacets
// The QueryFacets are returned in the order of the SearchRequest.FacetField
func (sr *SearchRequest) DecodeFacets(res *elastic.SearchResult, fb *FacetURIBuilder) ([]*QueryFacet, error) {
	facets, err := DecodeFacets(res, fb)
	if err != nil {
		return facets, err
	}

	queryFacets := map[string]*QueryFacet{}
	for _, facet := range facets {
		queryFacets[facet.Field] = facet
	}

	orderedFacets := []*QueryFacet{}

	for _, field := range sr.FacetField {
		facet, ok := queryFacets[field.Field]
		if ok {
			orderedFacets = append(orderedFacets, facet)
		}
	}

	return orderedFacets, nil
}

// DecodeFacets decodes the elastic aggregations in the SearchResult to fragments.QueryFacets
func DecodeFacets(res *elastic.SearchResult, fb *FacetURIBuilder) ([]*QueryFacet, error) {
	if res == nil || res.TotalHits() == 0 {
		return nil, nil
	}

	var aggs []*QueryFacet
	for k := range res.Aggregations {
		facetFilter, ok := res.Aggregations.Nested(k)
		if ok {
			facet, ok := facetFilter.Filter("filter")
			if ok {
				inner, ok := facet.Filter("inner")
				if ok {
					var valid bool
					qf := &QueryFacet{
						Name:  k, // todo add get by name to fb
						Field: k,
						Total: inner.DocCount,
						Links: []*FacetLink{},
					}

					maxAgg, ok := inner.Max("maxval")
					if ok {

						qf.Max = getKeyAsString(maxAgg.Aggregations["value_as_string"])
						valid = true
					}
					minAgg, ok := inner.Max("minval")
					if ok {
						qf.Min = getKeyAsString(minAgg.Aggregations["value_as_string"])
						valid = true
					}
					value, ok := inner.Terms("value")
					if ok {
						valid = true
						qf.OtherDocs = value.SumOfOtherDocCount
						for _, b := range value.Buckets {
							key := KeyAsString(b)
							fl := &FacetLink{
								Value:         key,
								Count:         b.DocCount,
								DisplayString: fmt.Sprintf(facetDisplayLabel, key, b.DocCount),
							}
							SetFacetLink(key, qf, fl, fb)
							qf.Links = append(qf.Links, fl)
						}
					}
					histogram, ok := inner.Histogram("histogram")
					if ok {
						valid = true
						for _, b := range histogram.Buckets {
							key := *b.KeyAsString
							url, isSelected := fb.CreateFacetFilterURI(qf.Field, key)

							if isSelected && !qf.IsSelected {
								qf.IsSelected = true
							}
							fl := &FacetLink{
								URL:           url,
								IsSelected:    isSelected,
								Value:         key,
								Count:         b.DocCount,
								DisplayString: fmt.Sprintf(facetDisplayLabel, key, b.DocCount),
							}
							qf.Links = append(qf.Links, fl)
						}
					}
					if valid {
						aggs = append(aggs, qf)
					}
				}
			}
		}
		objectFilter, ok := res.Aggregations.Filter(k)
		if ok {
			value, ok := objectFilter.Terms("object")
			if ok {
				qf := &QueryFacet{
					Name:        k, // todo add get by name to fb
					Field:       k,
					Total:       objectFilter.DocCount,
					MissingDocs: value.SumOfOtherDocCount,
					Links:       []*FacetLink{},
				}

				for _, b := range value.Buckets {
					key := KeyAsString(b)

					fl := &FacetLink{
						Value:         key,
						Count:         b.DocCount,
						DisplayString: fmt.Sprintf(facetDisplayLabel, key, b.DocCount),
					}

					SetFacetLink(key, qf, fl, fb)

					qf.Links = append(qf.Links, fl)
				}

				aggs = append(aggs, qf)
			}
		}
	}
	return aggs, nil
}

func SetFacetLink(key string, qf *QueryFacet, fl *FacetLink, fb *FacetURIBuilder) {
	if fb != nil {
		url, isSelected := fb.CreateFacetFilterURI(qf.Field, key)

		if isSelected && !qf.IsSelected {
			qf.IsSelected = true
		}

		fl.URL = url
		fl.IsSelected = isSelected
	}
}

// KeyAsString extracts the key as string from the elastic.AggregationBucketKeyItem.
func KeyAsString(b *elastic.AggregationBucketKeyItem) string {
	var key string
	// first check value KeyAsString
	switch b.KeyAsString {
	case nil:
		switch b.Key.(type) {
		case float64:
			key = strconv.Itoa(int(b.Key.(float64)))
		case string:
			key = b.Key.(string)
		default:
			log.Printf("unable to format key %#v", b.Key)
		}
	default:
		key = *b.KeyAsString
	}
	return key
}

func escapeRawQuery(query string) string {
	if !strings.Contains(query, "/") {
		return query
	}

	var parts []string
	for _, v := range strings.Fields(query) {
		if !strings.Contains(v, "/") {
			parts = append(parts, v)
			continue
		}
		if strings.HasPrefix(v, "\"") || strings.HasSuffix(v, "\"") {
			parts = append(parts, v)
			continue
		}

		parts = append(parts, fmt.Sprintf("\"%s\"", v))
	}

	return strings.Join(parts, " ")
}

func getBoost(input string) (field string, boost float64) {
	parts := strings.Split(input, "^")
	if len(parts) < 2 {
		return parts[0], 0
	}

	boost, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		slog.Error("unable to parse boost from search field", "field", input, "error", err)
		return parts[0], 0
	}

	return parts[0], boost
}

func QueryFromSearchFields(query string, fields ...string) (elastic.Query, error) {
	bq := elastic.NewBoolQuery()

	var directFields, nestedFields []string

	// filter fields
	for _, field := range fields {
		if field == "full_text" || (strings.HasPrefix(field, "tree.") || strings.HasPrefix(field, "meta.")) {
			directFields = append(directFields, field)
			continue
		}
		nestedFields = append(nestedFields, field)
	}

	q := elastic.NewQueryStringQuery(escapeRawQuery(query))
	q.DefaultOperator("and")
	for _, field := range directFields {
		f, boost := getBoost(field)
		if boost > 0 || boost < 0 {
			q = q.FieldWithBoost(f, boost)
			continue
		}
		q = q.Field(f)
	}

	if len(directFields) > 0 {
		bq = bq.Should(q)
	}

	nbq := elastic.NewBoolQuery()
	for _, field := range nestedFields {
		f, boost := getBoost(field)
		nbq.Should(createFieldedSubQuery(f, query, boost))
	}

	if len(nestedFields) > 0 {
		nq := elastic.NewNestedQuery(resourcesEntries, nbq)
		bq = bq.Should(nq)
	}

	if !isAdvancedSearch(query) {
		bq = bq.MinimumShouldMatch(c.Config.MinimumShouldMatch)
	}

	return bq, nil
}

func (h *Header) LastModified() time.Time {
	return LastModified(h.Modified)
}
