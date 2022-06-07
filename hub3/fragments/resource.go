// Copyright © 2017 Delving B.V. <info@delving.eu>
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

package fragments

import (
	"context"
	fmt "fmt"
	"io"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/index"
	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/search"
	"github.com/delving/hub3/ikuzo/storage/x/memory"
	r "github.com/kiivihal/rdf2go"
	elastic "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

const (
	literal      = "Literal"
	resourceType = "Resource"
	bnode        = "Bnode"
)

var ctx context.Context

func init() {
	ctx = context.Background()
}

func logLabelErr(predicate string, err error) {
	log.Printf("Unable to create search label for %s  due to %s\n", predicate, err)
}

// NewContext returns the context for the current fragmentresource
func (fr *FragmentResource) NewContext(predicate, objectID string) *FragmentReferrerContext {
	searchLabel, err := c.Config.NameSpaceMap.GetSearchLabel(predicate)
	if err != nil {
		logLabelErr(predicate, err)
		searchLabel = ""
	}

	label, _ := fr.GetLabel()

	return &FragmentReferrerContext{
		Subject:      fr.ID,
		SubjectClass: fr.Types,
		Predicate:    predicate,
		Level:        fr.GetLevel(),
		ObjectID:     objectID,
		SearchLabel:  searchLabel,
		Label:        label,
	}
}

// ResourceMap is a convenience structure to hold the resourceMap data and functions
type ResourceMap struct {
	resources map[string]*FragmentResource
	orgID     string
}

// Tree holds all the core information for building Navigational Trees from RDF graphs
type Tree struct {
	Leaf             string              `json:"leaf,omitempty"`
	Parent           string              `json:"parent,omitempty"`
	Label            string              `json:"label"`
	CLevel           string              `json:"cLevel"`
	UnitID           string              `json:"unitID"`
	Type             string              `json:"type"`
	HubID            string              `json:"hubID"`
	ChildCount       int                 `json:"childCount"`
	Depth            int                 `json:"depth"`
	HasChildren      bool                `json:"hasChildren"`
	HasDigitalObject bool                `json:"hasDigitalObject"`
	HasRestriction   bool                `json:"hasRestriction"`
	DaoLink          string              `json:"daoLink,omitempty"`
	ManifestLink     string              `json:"manifestLink,omitempty"`
	MimeTypes        []string            `json:"mimeType,omitempty"`
	DOCount          int                 `json:"doCount"`
	Inline           []*Tree             `json:"inline,omitempty"`
	SortKey          uint64              `json:"sortKey"`
	Periods          []string            `json:"periods"`
	Content          []string            `json:"content,omitempty"`
	RawContent       []string            `json:"rawContent,omitempty"`
	Access           string              `json:"access,omitempty"`
	Title            string              `json:"title,omitempty"`
	Description      []string            `json:"description,omitempty"`
	InventoryID      string              `json:"inventoryID,omitempty"`
	AgencyCode       string              `json:"agencyCode,omitempty"`
	PeriodDesc       []string            `json:"periodDesc,omitempty"`
	Material         string              `json:"material,omitempty"`
	PhysDesc         string              `json:"physDesc,omitempty"`
	Fields           map[string][]string `json:"fields,omitempty"`
}

// DeepCopy creates a deep-copy of a Tree.
func (t *Tree) DeepCopy() *Tree {
	target := &Tree{
		Leaf:             t.Leaf,
		Parent:           t.Parent,
		Label:            t.Label,
		CLevel:           t.CLevel,
		UnitID:           t.UnitID,
		Type:             t.Type,
		HubID:            t.HubID,
		ChildCount:       t.ChildCount,
		Depth:            t.Depth,
		HasChildren:      t.HasChildren,
		HasDigitalObject: t.HasDigitalObject,
		HasRestriction:   t.HasRestriction,
		DaoLink:          t.DaoLink,
		ManifestLink:     t.ManifestLink,
		MimeTypes:        t.MimeTypes,
		DOCount:          t.DOCount,
		Inline:           t.Inline,
		SortKey:          t.SortKey,
		Periods:          t.Periods,
		Content:          t.Content,
		RawContent:       t.RawContent,
		Access:           t.Access,
		Title:            t.Title,
		Description:      t.Description,
		InventoryID:      t.InventoryID,
		AgencyCode:       t.AgencyCode,
		PeriodDesc:       t.PeriodDesc,
		Material:         t.Material,
		PhysDesc:         t.PhysDesc,
		Fields:           t.Fields,
	}
	return target
}

// PageEntry creates a paging entry for a tree element.
func (t *Tree) PageEntry() *TreePageEntry {
	return &TreePageEntry{
		CLevel:      t.CLevel,
		SortKey:     int32(t.SortKey),
		Depth:       int32(t.Depth),
		ExpandedIDs: ExpandedIDs(t),
	}
}

// TreeNavigator possible remove
type TreeNavigator struct {
	Cursor  int               `json:"cursor"`
	Total   int               `json:"total"`
	CLevel  string            `json:"cLevel"`
	Entries map[string]string `json:"entries"` // key is the CLevel
}

// IsExpanded returns if the tree query contains a query that puts the active ID
// expanded in the tree
func (tq *TreeQuery) IsExpanded() bool {
	return (tq.Label != "" || tq.UnitID != "" || tq.Query != "") || tq.IsPaging
}

// IsNavigatedQuery returns if there is both a query and active ID
func (tq *TreeQuery) IsNavigatedQuery() bool {
	return (tq.Label != "" || tq.Query != "") && tq.UnitID != ""
}

// PreviousCurrentNextPage returns the previous and next page based on the TreeQuery.
//
// This does not take max boundaries based on number of records returned into account.
func (tq *TreeQuery) PreviousCurrentNextPage() (int32, int32, int32) {
	sort.Slice(tq.Page, func(i, j int) bool { return tq.Page[i] < tq.Page[j] })
	if len(tq.Page) == 0 {
		return int32(0), int32(1), int32(2)
	}
	max := tq.Page[0]
	min := tq.Page[0]
	for _, value := range tq.Page {
		if max < value {
			max = value
		}
		if min > value {
			min = value
		}
	}
	return min - 1, min, max + 1
}

// TreePagingSize returns the relative size of the paging window based on the number of pages.
// This is used to set the ElasticSearch responseSize.
func (tq *TreeQuery) TreePagingSize() int32 {
	nrPages := len(tq.GetPage())
	if nrPages == 0 {
		return tq.GetPageSize()
	}
	return int32(nrPages) * tq.GetPageSize()
}

// SearchPages returns the active search pages for a given sortKey
func (tq *TreeQuery) SearchPages(sortKey int32) ([]int32, error) {
	pages := []int32{}
	if sortKey == 0 {
		return pages, fmt.Errorf("can't set search page for 0")
	}
	pageNr := (sortKey / tq.GetPageSize()) + 1
	pages = append(pages, pageNr)
	relativePlace := (sortKey % tq.GetPageSize())
	closeToNext := (tq.GetPageSize() - relativePlace) < 25
	if closeToNext {
		extraPages := (relativePlace / tq.GetPageSize()) + 1
		for i := int32(1); i <= extraPages; i++ {
			pages = append(pages, pageNr+int32(i))
		}
	}
	return pages, nil
}

// GetPreviousScrollIDs returns scrollIDs up to the cLevel
// This information can be used to construct the previous search results when
// both the UnitID and the Label are being queried
func (tq *TreeQuery) GetPreviousScrollIDs(cLevel string, sr *SearchRequest, pager *ScrollPager) ([]string, error) {
	previous := []string{}
	query := elastic.NewBoolQuery()

	matchSuffix := fmt.Sprintf("_%s", strings.TrimLeft(cLevel, "@"))

	q := elastic.NewQueryStringQuery(escapeRawQuery(sr.Tree.GetLabel()))
	q.DefaultOperator("and")

	q = q.Field("tree.label")
	if !isAdvancedSearch(sr.Tree.GetLabel()) {
		q = q.MinimumShouldMatch(c.Config.ElasticSearch.MinimumShouldMatch)
	}
	query = query.Must(q)
	query = query.Must(elastic.NewTermQuery(c.Config.ElasticSearch.SpecKey, tq.Spec))

	idSort := elastic.NewFieldSort("meta.hubID")
	fieldSort := elastic.NewFieldSort("tree.sortKey")

	scroll := index.ESClient().Scroll(c.Config.ElasticSearch.GetIndexName(tq.OrgID)).
		TrackTotalHits(c.Config.ElasticSearch.TrackTotalHits).
		SortBy(fieldSort, idSort).
		Size(100).
		FetchSource(false).
		Query(query)

	cursor := 0

	sr.Tree.UnitID = ""
	sr.Tree.Leaf = ""
	sr.SearchAfter = nil
	sr.Tree.Depth = []string{}
	sr.Tree.FillTree = false

	for {
		results, err := scroll.Do(context.Background())
		if err == io.EOF {
			return previous, nil // all results retrieved
		}
		if err != nil {
			return nil, err // something went wrong
		}

		for _, hit := range results.Hits.Hits {
			nextSearchAfter, err := sr.CreateBinKey(hit.Sort)
			if err != nil {
				return nil, errors.Wrap(err, "unable to create bytes for search after key")
			}
			// sr.CalculatedTotal = results.TotalHits()

			sr.Start = int32(cursor)
			sr.SearchAfter = nextSearchAfter
			hexRequest, err := sr.SearchRequestToHex()
			if err != nil {
				return nil, errors.Wrap(err, "unable to create bytes for search after key")
			}

			if strings.HasSuffix(hit.Id, matchSuffix) {
				// log.Printf("found it: %s ", matchSuffix)
				pager.Cursor = int32(cursor)
				pager.NextScrollID = hexRequest
				pager.Total = results.TotalHits()
				return previous, nil // all results retrieved
			}
			previous = append(previous, hexRequest)
			cursor++
		}
	}
}

// ExpandedIDs expands all the parent identifiers in a CLevel path and returns it as a map.
func ExpandedIDs(node *Tree) map[string]bool {
	expandedIDs := make(map[string]bool)
	parents := strings.Split(node.CLevel, "~")

	var path string
	for idx, leaf := range parents {
		if idx == 0 {
			path = leaf
			expandedIDs[path] = true
			continue
		}
		path = fmt.Sprintf("%s~%s", path, leaf)
		expandedIDs[path] = true
	}
	if !node.HasChildren {
		expandedIDs[node.CLevel] = false
	}
	return expandedIDs
}

// InlineTree creates a nested tree from an Array of *Tree
func InlineTree(nodes []*Tree, tq *TreeQuery, total int64) ([]*Tree, map[string]*Tree, error) {
	rootNodes := []*Tree{}
	nodeMap := make(map[string]*Tree)

	for _, n := range nodes {
		if n.Depth == 1 {
			rootNodes = append(rootNodes, n)
		}
		nodeMap[n.CLevel] = n
	}

	for _, n := range nodes {
		target, ok := nodeMap[n.Leaf]
		n.HasChildren = len(n.Inline) != 0
		if ok {
			target.HasChildren = true
			target.Inline = append(target.Inline, n)
		}
	}

	return rootNodes, nodeMap, nil
}

// ResourceEntryHighlight holds the values of the ElasticSearch highlight fiel
type ResourceEntryHighlight struct {
	SearchLabel string   `json:"searchLabel"`
	MarkDown    []string `json:"markdown"`
}

// Deprecated: use FragmentResource.AddTo to add them to a graph
func (fr *FragmentResource) GenerateTriples() []*r.Triple {
	triples := []*r.Triple{}
	subject := r.NewResource(fr.ID)
	for _, rdfType := range fr.Types {
		triples = append(
			triples,
			r.NewTriple(
				subject,
				r.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"),
				r.NewResource(rdfType),
			),
		)
	}

	for _, entry := range fr.Entries {
		triples = append(triples, entry.GetTriple(subject))
	}

	return triples
}

func (fr *FragmentResource) AddTo(g *rdf.Graph) error {
	subject, err := rdf.NewIRI(fr.ID)
	if err != nil {
		return err
	}

	for _, rdfType := range fr.Types {
		rdfType, err := rdf.NewIRI(rdfType)
		if err != nil {
			return err
		}

		g.AddTriple(subject, rdf.IsA, rdfType)
	}

	for _, entry := range fr.Entries {
		triple, err := entry.AsTriple(subject)
		if err != nil {
			return err
		}

		g.Add(triple)
	}

	return nil
}

// GenerateJSONLD converts a FragmenResource into a JSON-LD entry
func (fr *FragmentResource) GenerateJSONLD() map[string]interface{} {
	m := map[string]interface{}{}
	m["@id"] = fr.ID
	if len(fr.Types) > 0 {
		m["@type"] = fr.Types
	}
	entries := map[string][]*ResourceEntry{}
	for _, p := range fr.Entries {
		entries[p.Predicate] = append(entries[p.Predicate], p)
	}
	for k, v := range entries {
		objects := []*r.LdObject{}
		for _, p := range v {
			objects = append(objects, p.AsLdObject())
		}
		if len(objects) != 0 {
			m[k] = objects
		}
	}
	return m
}

// Collapsed holds each entry of a FieldCollapse elasticsearch result
type Collapsed struct {
	Field    string           `json:"field"`
	Title    string           `json:"title"`
	HitCount int64            `json:"hitCount"`
	Items    []*FragmentGraph `json:"items"`
}

// ScrollPager holds all paging information for a search result.
type ScrollPager struct {
	// scrollID is serialized version SearchRequest
	PreviousScrollID string `json:"previousScrollID"`
	NextScrollID     string `json:"nextScrollID"`
	Cursor           int32  `json:"cursor"`
	Total            int64  `json:"total"`
	Rows             int32  `json:"rows"`
}

// ProtoBuf holds a protobuf encode version of the messageType.
type ProtoBuf struct {
	MessageType string `json:"messageType,omitempty"`
	Data        string `json:"data,omitempty"`
}

// ScrollResultV4 intermediate non-protobuf search results
type ScrollResultV4 struct {
	Pager      *ScrollPager       `json:"pager"`
	Pagination *search.Paginator  `json:"pagination,omitempty"`
	Query      *Query             `json:"query"`
	Items      []*FragmentGraph   `json:"items,omitempty"`
	Collapsed  []*Collapsed       `json:"collapse,omitempty"`
	Peek       map[string]int64   `json:"peek,omitempty"`
	Facets     []*QueryFacet      `json:"facets,omitempty"`
	TreeHeader *TreeHeader        `json:"treeHeader,omitempty"`
	Tree       []*Tree            `json:"tree,omitempty"`
	TreePage   map[string][]*Tree `json:"treePage,omitempty"`
	ProtoBuf   *ProtoBuf          `json:"-"`
}

// TreeHeader contains rendering hints for the consumer of the TreeView API.
type TreeHeader struct {
	ActiveID          string          `json:"activeID"`
	ExpandedIDs       map[string]bool `json:"expandedIDs,omitempty"`
	PreviousScrollIDs []string        `json:"previousScrollIDs,omitempty"`
	Paging            *TreePaging     `json:"paging,omitempty"`
	Searching         *TreeSearching  `json:"searching,omitempty"`
}

// TreeSearching contains rendering hints for the search results in the TreeView API.
type TreeSearching struct {
	IsSearch    bool   `json:"isSearch"`
	HitsTotal   int32  `json:"hitsTotal"`
	CurrentHit  int32  `json:"currentHit"`
	HasNext     bool   `json:"hasNext"`
	HasPrevious bool   `json:"hasPrevious"`
	ByLabel     string `json:"byLabel,omitempty"`
	ByUnitID    string `json:"byUnitID,omitempty"`
	ByQuery     string `json:"byQuery,omitempty"`
}

// SetPreviousNext calculate previous and next search paging
func (ts *TreeSearching) SetPreviousNext(start int32) {
	cursor := start + 1
	ts.CurrentHit = cursor
	ts.HasNext = ts.CurrentHit < ts.HitsTotal
	ts.HasPrevious = start > 0
}

// TreePaging contains rendering hints for Paging through a Tree and Tree search-results
type TreePaging struct {
	PageSize        int32                    `json:"pageSize,omitempty"`
	NrPages         int32                    `json:"nrPages,omitempty"`
	HasNext         bool                     `json:"hasNext"`
	HasPrevious     bool                     `json:"hasPrevious"`
	PageNext        int32                    `json:"pageNext,omitempty"`
	PagePrevious    int32                    `json:"pagePrevious,omitempty"`
	PageCurrent     []int32                  `json:"pageCurrent,omitempty"`
	PageFirst       int32                    `json:"pageFirst"`
	PageLast        int32                    `json:"pageLast"`
	ResultFirst     *TreePageEntry           `json:"resultFirst,omitempty"`
	ResultLast      *TreePageEntry           `json:"resultLast,omitempty"`
	ResultActive    *TreePageEntry           `json:"resultActive,omitempty"`
	HitsOnPage      map[int32]*TreePageEntry `json:"hitsOnPage,omitempty"`
	HitsOnPageCount int32                    `json:"hitsOnPageCount,omitempty"`
	HitsTotalCount  int32                    `json:"hitsTotalCount,omitempty"`
	ActiveHit       int32                    `json:"activeHit,omitempty"`
	SameLeaf        bool                     `json:"sameLeaf"`
	IsSearch        bool                     `json:"isSearch"`
}

// CalculatePaging calculates all the paging information.
// This applies to searching and normal paging.
func (tp *TreePaging) CalculatePaging() {
	if tp.HitsTotalCount == 0 {
		tp.NrPages = 0
		return
	}
	if tp.HitsTotalCount < tp.PageSize {
		tp.PageCurrent = []int32{1}
		tp.NrPages = 1
		return
	}
	pages := tp.HitsTotalCount / tp.PageSize
	if tp.HitsTotalCount%tp.PageSize != 0 {
		pages++
	}
	tp.NrPages = pages
	tp.setFirstLastPage()
	if tp.PageFirst != int32(1) {
		tp.HasPrevious = true
		tp.PagePrevious = tp.PageFirst - 1
	}
	if tp.PageLast != tp.NrPages {
		tp.HasNext = true
		tp.PageNext = tp.PageLast + 1
	}
	return
}

func (tp *TreePaging) setFirstLastPage() {
	if len(tp.PageCurrent) == 0 {
		tp.PageFirst = int32(1)
		tp.PageLast = int32(1)
		return
	}

	sort.Slice(tp.PageCurrent, func(i, j int) bool { return tp.PageCurrent[i] < tp.PageCurrent[j] })

	min := tp.PageCurrent[0]
	max := tp.PageCurrent[0]

	for _, value := range tp.PageCurrent {
		if max < value {
			max = value
		}
		if min > value {
			min = value
		}
	}
	tp.PageFirst = min
	tp.PageLast = max
	return
}

// TreePageEntry contains information how to merge pages from different responses.
type TreePageEntry struct {
	CLevel      string          `json:"cLevel"`
	SortKey     int32           `json:"sortKey"`
	ExpandedIDs map[string]bool `json:"expandedIDs,omitempty"`
	Depth       int32           `json:"depth"`
}

// CreateTreePage creates a paging entry that can be used to merge the EAD tree between
// different paging request.
func (tpe *TreePageEntry) CreateTreePage(
	nodeMap map[string]*Tree,
	rootNodes []*Tree,
	appending bool,
	sortFrom int32,
) map[string][]*Tree {
	page := make(map[string][]*Tree)

	var rootLevelNodes []*Tree

	switch appending {
	case true:
		for _, rootNode := range rootNodes {
			if int32(rootNode.SortKey) >= tpe.SortKey && !strings.HasPrefix(tpe.CLevel, rootNode.CLevel) {
				rootLevelNodes = append(rootLevelNodes, rootNode)
			}
		}
	case false:
		for _, rootNode := range rootNodes {
			// Return more root nodes in prepend mode because it is possible they fall out of sort key range making
			// it impossible to prepend new nodes when they do not have a parent.
			if int32(rootNode.SortKey) <= tpe.SortKey {
				rootNode = rootNode.DeepCopy()
				rootNode.Inline = []*Tree{}
				rootLevelNodes = append(rootLevelNodes, rootNode)
			}
		}
	}

	if len(rootLevelNodes) != 0 {
		page["root"] = rootLevelNodes
	}

	switch appending {
	case true:
		for levelID := range tpe.ExpandedIDs {
			if levelID != tpe.CLevel {
				node, ok := nodeMap[levelID]
				levelNodes := []*Tree{}
				if ok {
					for _, subNode := range node.Inline {
						if int32(subNode.SortKey) >= tpe.SortKey {
							levelNodes = append(levelNodes, subNode)
						}
					}
				}
				page[levelID] = levelNodes
			}
		}
	case false:
		for _, rootNode := range rootNodes {
			tpe.recurseNodes(rootNode, page, sortFrom)
		}

	}

	return page
}

func (tpe *TreePageEntry) recurseNodes(node *Tree, page map[string][]*Tree, sortFrom int32) {
	for _, subNode := range node.Inline {
		children, ok := page[subNode.Leaf]
		if !ok {
			children = []*Tree{}
		}
		_, ok = tpe.ExpandedIDs[subNode.CLevel]
		if !ok {
			children = append(children, subNode)
			page[subNode.Leaf] = children
			continue
		}

		pagingNode := subNode.DeepCopy()
		pagingNode.Inline = []*Tree{}
		children = append(children, pagingNode)
		page[subNode.Leaf] = children
		tpe.recurseNodes(subNode, page, sortFrom)
	}
}

// SameLeaf determines if two TreePageEntry are in the same tree leaf.
func (tpe *TreePageEntry) SameLeaf(other *TreePageEntry) bool {
	first := strings.Split(tpe.CLevel, "~")
	second := strings.Split(other.CLevel, "~")
	return strings.Join(first[:len(first)-1], "~") == strings.Join(second[:len(second)-1], "~")
}

// QueryFacet contains all the information for an ElasticSearch Aggregation
type QueryFacet struct {
	Name        string       `json:"name"`
	Field       string       `json:"field"`
	IsSelected  bool         `json:"isSelected"`
	I18n        string       `json:"i18N,omitempty"`
	Total       int64        `json:"total"`
	MissingDocs int64        `json:"missingDocs"`
	OtherDocs   int64        `json:"otherDocs"`
	Min         string       `json:"min,omitempty"`
	Max         string       `json:"max,omitempty"`
	Type        string       `json:"type,omitempty"`
	Links       []*FacetLink `json:"links"`
}

// FacetLink contains all the information for creating a filter for this facet
type FacetLink struct {
	URL           string `json:"url"`
	IsSelected    bool   `json:"isSelected"`
	Value         string `json:"value"`
	DisplayString string `json:"displayString"`
	Count         int64  `json:"count"`
}

// FragmentResource holds all the conttext information for a resource
// It works together with the FragmentBuilder to create the linked fragments
type FragmentResource struct {
	ID                   string                     `json:"id"`
	Types                []string                   `json:"types"`
	GraphExternalContext []*FragmentReferrerContext `json:"graphExternalContext"`
	Context              []*FragmentReferrerContext `json:"context"`
	Entries              []*ResourceEntry           `json:"entries"`
	Tags                 []string                   `json:"tags,omitempty"`
	predicates           map[string][]*FragmentEntry
	objectIDs            []*FragmentReferrerContext
}

// BySortOrder implements sort.Interface for []*FragmentResource based on
// the Order field in the first FragmentEntry.
type BySortOrder []*FragmentResource

func (a BySortOrder) Len() int      { return len(a) }
func (a BySortOrder) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a BySortOrder) Less(i, j int) bool {
	one := a[i]
	two := a[j]
	if len(one.Entries) == 0 {
		return true
	}

	if len(two.Entries) == 0 {
		return false
	}

	return one.Entries[0].Order < two.Entries[0].Order
}

// ContextPath returns a string that can be used to reconstruct the path hierarchy
// for statistics. The values are separated by a forward slash.
func (fr *FragmentResource) ContextPath() string {
	var path []string
	for _, context := range fr.Context {

		// just take the first rdf:type the rest are shown in @rdf:type
		rdfType := "rdf_Description"
		if len(context.GetSubjectClass()) != 0 {
			rdfType = context.GetSubjectClass()[0]
			searchLabel, err := c.Config.NameSpaceMap.GetSearchLabel(rdfType)
			if err != nil {
				log.Printf("Unable to create search label for %s  due to %s\n", rdfType, err)
			}
			if searchLabel != "" {
				rdfType = searchLabel
			}
		}
		path = append(path, rdfType, context.GetSearchLabel())
	}
	return strings.Join(path, "/")
}

// ObjectIDs returns an array of FragmentReferrerContext
func (fr *FragmentResource) ObjectIDs() []*FragmentReferrerContext {
	return fr.objectIDs
}

// Predicates returns a map of FragmentEntry
func (fr *FragmentResource) Predicates() map[string][]*FragmentEntry {
	return fr.predicates
}

// SetEntries sets the ResourceEntries for indexing
func (fr *FragmentResource) SetEntries(rm *ResourceMap) error {
	fr.Entries = []*ResourceEntry{}
	for predicate, entries := range fr.predicates {
		for _, entry := range entries {
			re, err := entry.NewResourceEntry(predicate, fr.GetLevel(), rm)
			if err != nil {
				return err
			}
			fr.Entries = append(fr.Entries, re)
		}
	}
	// sort entries by order
	sort.Slice(fr.Entries[:], func(i, j int) bool {
		return fr.Entries[i].Order < fr.Entries[j].Order
	})
	return nil
}

// AsLdObject generates an rdf2go.LdObject for JSON-LD generation
func (fe *FragmentEntry) AsLdObject() *r.LdObject {
	return &r.LdObject{
		ID:       fe.ID,
		Value:    fe.Value,
		Language: fe.Language,
		Datatype: fe.DataType,
	}
}

// NewResourceEntry creates a resource entry for indexing
func (fe *FragmentEntry) NewResourceEntry(predicate string, level int32, rm *ResourceMap) (*ResourceEntry, error) {
	label, err := c.Config.NameSpaceMap.GetSearchLabel(predicate)
	if err != nil {
		logLabelErr(predicate, err)
		label = ""
	}
	re := &ResourceEntry{
		ID:          fe.ID,
		Value:       fe.Value,
		Language:    fe.Language,
		DataType:    fe.DataType,
		EntryType:   fe.EntryType,
		Predicate:   predicate,
		Level:       level,
		SearchLabel: label,
		Order:       fe.Order,
	}

	if re.ID != "" {
		r, ok := rm.GetResource(re.ID)
		if ok {
			re.Value, _ = r.GetLabel()
		}
	}

	// add label for resolved
	if fe.Resolved {
		re.AddTags("resolved")
	}

	switch re.DataType {
	case "http://www.w3.org/2001/XMLSchema#integer":
		i, err := strconv.Atoi(re.Value)
		if err != nil {
			log.Printf("unable to convert to int: %#v", err)
			return re, err
		}
		re.Integer = i
	case "http://www.w3.org/2001/XMLSchema#float":
		i, err := strconv.ParseFloat(re.Value, 32)
		if err != nil {
			log.Printf("unable to convert to float: %#v", err)
			return re, err
		}
		re.Float = i
	}

	labels, ok := c.Config.RDFTagMap.Get(predicate)
	if ok {
		re.AddTags(labels...)
		if re.Value != "" {
			// TODO add validation for the values here
			for _, label := range labels {
				switch label {
				case "isoDate":
					re.Date = append(re.Date, re.Value)
				case "dateRange":
					indexRange, err := CreateDateRange(re.Value)
					if err != nil {
						log.Printf("Unable to create dateRange for: %#v", re.Value)
						continue
					}
					re.DateRange = &indexRange
					if indexRange.Greater != "" {
						re.Date = append(re.Date, indexRange.Greater)
					}
					if indexRange.Less != "" {
						re.Date = append(re.Date, indexRange.Less)
					}
				case "latLong":
					re.LatLong = re.Value
				case "integer":
					i, err := strconv.Atoi(re.Value)
					if err != nil {
						log.Printf("Unable to create integer for: %#v", re.Value)
						continue
					}
					log.Printf("extracting integer from tag: %s", re.Value)
					re.Integer = i
				}
			}
		}
	}
	return re, nil
}

// CreateDateRange creates a date indexRange
func CreateDateRange(period string) (IndexRange, error) {
	ir := IndexRange{}
	parts := strings.FieldsFunc(period, splitPeriod)
	switch len(parts) {
	case 1:
		// start and end year
		ir.Greater, _ = padYears(parts[0], true)
		ir.Less, _ = padYears(parts[0], false)
	case 2:
		ir.Greater, _ = padYears(parts[0], true)
		ir.Less, _ = padYears(parts[1], false)
	default:
		return ir, fmt.Errorf("Unable to create data range for: %#v", parts)
	}

	if err := ir.Valid(); err != nil {
		return ir, err
	}

	return ir, nil
}

func padYears(year string, start bool) (string, error) {
	parts := strings.Split(year, "-")
	switch len(parts) {
	case 3:
		return year, nil
	case 2:
		year := parts[0]
		month := parts[1]
		switch start {
		case true:
			return fmt.Sprintf("%s-%s-01", year, month), nil
		case false:
			switch parts[1] {
			case "01", "03", "05", "07", "08", "10", "12":
				return fmt.Sprintf("%s-%s-31", year, month), nil
			case "02":
				return fmt.Sprintf("%s-%s-28", year, month), nil
			default:
				return fmt.Sprintf("%s-%s-30", year, month), nil
			}
		}
	case 1:
		year := parts[0]
		switch len(year) {
		case 4:
			switch start {
			case true:
				return fmt.Sprintf("%s-01-01", year), nil
			case false:
				return fmt.Sprintf("%s-12-31", year), nil
			}
		default:
			// try to hyphenate the date
			date, err := hyphenateDate(year)
			if err != nil {
				return "", err
			}
			return padYears(date, start)
		}
	}
	return "", fmt.Errorf("unsupported case for padding: %s", year)
}

// hyphenateDate converts a string of date string into the hyphenated form.
// Only YYYYMMDD and YYYYMM are supported.
func hyphenateDate(date string) (string, error) {
	switch len(date) {
	case 4:
		return date, nil
	case 6:
		return fmt.Sprintf("%s-%s", date[:4], date[4:]), nil
	case 8:
		return fmt.Sprintf("%s-%s-%s", date[:4], date[4:6], date[6:]), nil
	}
	return "", fmt.Errorf("Unable to hyphenate date string: %#v", date)
}

func splitPeriod(c rune) bool {
	return !unicode.IsNumber(c) && c != '-'
}

// GetLabel returns the label and language for a resource
// This is used to present a label for a link in the interface
func (fr *FragmentResource) GetLabel() (label, language string) {
	if fr.ID == "" {
		return "", ""
	}
	for _, labelPredicate := range c.Config.RDFTag.Label {
		o, ok := fr.predicates[labelPredicate]
		if ok && len(o) != 0 {
			return o[0].Value, o[0].Language
		}
	}
	return "", ""
}

// SetContextLevels sets FragmentReferrerContext to each level from the root
func (rm *ResourceMap) SetContextLevels(subjectURI string) (map[string]*FragmentResource, error) {
	if len(rm.resources) == 0 {
		return nil, fmt.Errorf("ResourceMap cannot be empty for subjecURI: %s", subjectURI)
	}

	subject, ok := rm.GetResource(subjectURI)
	if !ok {
		return nil, fmt.Errorf("Subject %s is not part of the graph", subjectURI)
	}

	linkedObjects := map[string]*FragmentResource{}
	linkedObjects[subjectURI] = subject
	for _, level1 := range subject.objectIDs {
		level2Resource, ok := rm.GetResource(level1.ObjectID)
		if !ok {
			// log.Printf("unknown target URI: %s", level1.ObjectID)
			continue
		}
		linkedObjects[level1.ObjectID] = level2Resource
		level1.Level = 1
		if len(level1.GetSubjectClass()) == 0 {
			level1.SubjectClass = subject.Types
		}
		// validate context
		level2Resource.AppendContext(level1)

		// loop into the next level, i.e. level 3
		for _, level2 := range level2Resource.objectIDs {
			level2.Level = 2
			level3Resource, ok := rm.GetResource(level2.ObjectID)
			if !ok {
				log.Printf("unknown target URI: %s", level2.ObjectID)
				continue
			}

			linkedObjects[level2.ObjectID] = level3Resource
			if len(level2.GetSubjectClass()) == 0 {
				level2.SubjectClass = level2Resource.Types
			}
			level3Resource.AppendContext(level1, level2)

			for _, level3 := range level3Resource.objectIDs {
				level3.Level = 3
				level4Resource, ok := rm.GetResource(level3.ObjectID)
				if !ok {
					log.Printf("unknown target URI: %s", level3.ObjectID)
					continue
				}
				linkedObjects[level3.ObjectID] = level3Resource
				if len(level3.GetSubjectClass()) == 0 {
					level3.SubjectClass = level3Resource.Types
				}
				level4Resource.AppendContext(level1, level2, level3)
			}
		}
	}

	return linkedObjects, nil
}

// AppendContext adds the referrerContext to the FragmentResource
// This action increments nilthe level count
func (fr *FragmentResource) AppendContext(ctxs ...*FragmentReferrerContext) {
	for _, ctx := range ctxs {
		if !containsContext(fr.Context, ctx) {
			fr.Context = append(fr.Context, ctx)
		}
	}
}

// FragmentEntry holds all the information for the object of a rdf2go.Triple
type FragmentEntry struct {
	ID        string `json:"@id,omitempty"`
	Value     string `json:"@value,omitempty"`
	Language  string `json:"@language,omitempty"`
	DataType  string `json:"@type,omitempty"`
	EntryType string `json:"entrytype"`
	Triple    string `json:"triple"`
	Resolved  bool   `json:"resolved"`
	Order     int    `json:"order"`
}

// ResourceEntry contains all the indexed entries for FragmentResources
type ResourceEntry struct {
	ID          string            `json:"@id,omitempty"`
	Value       string            `json:"@value,omitempty"`
	Language    string            `json:"@language,omitempty"`
	DataType    string            `json:"@type,omitempty"`
	EntryType   string            `json:"entrytype,omitempty"`
	Predicate   string            `json:"predicate,omitempty"`
	SearchLabel string            `json:"searchLabel,omitempty"`
	Level       int32             `json:"level"`
	Tags        []string          `json:"tags,omitempty"`
	Date        []string          `json:"isoDate,omitempty"`
	DateRange   *IndexRange       `json:"dateRange,omitempty"`
	Integer     int               `json:"integer,omitempty"`
	Float       float64           `json:"float,omitempty"`
	IntRange    *IndexRange       `json:"intRange,omitempty"`
	LatLong     string            `json:"latLong,omitempty"`
	Inline      *FragmentResource `json:"inline,omitempty"`
	Order       int               `json:"order"`
}

// IndexRange is used for indexing ranges.
type IndexRange struct {
	Greater string `json:"gte"`
	Less    string `json:"lte"`
}

// Valid checks if Less is smaller than Greater.
func (ir IndexRange) Valid() error {
	if ir.Greater > ir.Less {
		return fmt.Errorf("%s should not be greater than %s", ir.Less, ir.Greater)
	}
	return nil
}

func (re *ResourceEntry) AsTriple(subject rdf.Subject) (*rdf.Triple, error) {
	var err error
	predicate, err := rdf.NewIRI(re.Predicate)
	if err != nil {
		return nil, err
	}

	var object rdf.Object

	switch re.EntryType {
	case bnode:
		object, err = rdf.NewBlankNode(re.ID)
	case resourceType:
		object, err = rdf.NewIRI(re.ID)
	case literal:
		switch {
		case re.Language != "":
			object, err = rdf.NewLiteralWithLang(re.Value, re.Language)
		case re.DataType != "":
			dt, err := rdf.NewIRI(re.DataType)
			if err != nil {
				return nil, err
			}
			object, err = rdf.NewLiteralWithType(re.Value, dt)
		default:
			object, err = rdf.NewLiteral(re.Value)
		}
	default:
		log.Printf("bad datatype: '%#v'", re)
	}

	if err != nil {
		return nil, err
	}

	return rdf.NewTriple(
		subject,
		predicate,
		object,
	), nil
}

func (re *ResourceEntry) GetTriple(subject r.Term) *r.Triple {
	predicate := r.NewResource(re.Predicate)
	var object r.Term

	switch re.EntryType {
	case bnode:
		object = r.NewBlankNode(re.ID)
	case resourceType:
		object = r.NewResource(re.ID)
	case literal:
		switch {
		case re.Language != "":
			object = r.NewLiteralWithLanguage(re.Value, re.Language)
		case re.DataType != "":
			object = r.NewLiteralWithDatatype(re.Value, r.NewResource(re.DataType))
		default:
			object = r.NewLiteral(re.Value)
		}
	default:
		log.Printf("bad datatype: '%#v'", re)
	}

	return r.NewTriple(
		subject,
		predicate,
		object,
	)
}

// AsLdObject generates an rdf2go.LdObject for JSON-LD generation
func (re *ResourceEntry) AsLdObject() *r.LdObject {
	o := &r.LdObject{
		ID:       re.ID,
		Language: re.Language,
		Datatype: re.DataType,
	}
	if re.ID == "" {
		o.Value = re.Value
	}
	return o
}

// NewResourceMap creates a map for all the resources in the rdf2go.Graph
func NewResourceMap(orgID string, g *r.Graph) (*ResourceMap, error) {
	rm := &ResourceMap{
		resources: make(map[string]*FragmentResource),
		orgID:     orgID,
	}

	if g.Len() == 0 {
		return rm, fmt.Errorf("The graph cannot be empty")
	}

	seen := 0
	for t := range g.IterTriples() {
		seen++
		err := rm.AppendOrderedTriple(t, false, seen)
		if err != nil {
			return rm, err
		}
	}
	return rm, nil
}

// NewEmptyResourceMap returns an initialised ResourceMap
func NewEmptyResourceMap(orgID string) *ResourceMap {
	return &ResourceMap{
		resources: make(map[string]*FragmentResource),
		orgID:     orgID,
	}
}

// ResolveObjectIDs queries the fragmentstore for additional context
func (rm *ResourceMap) ResolveObjectIDs(excludeHubID string) error {
	objectIDs := []string{}
	for _, fr := range rm.Resources() {
		if contains(fr.Types, "http://www.europeana.eu/schemas/edm/WebResource") {
			objectIDs = append(objectIDs, fr.ID)
		}
	}
	if len(objectIDs) == 0 {
		return nil
	}
	// log.Printf("IDs to be resolved: %#v", objectIDs)

	req := NewFragmentRequest(rm.orgID)
	req.Subject = objectIDs
	req.ExcludeHubID = excludeHubID
	frags, _, err := req.Find(ctx, index.ESClient())
	if err != nil {
		log.Printf("unable to find fragments: %s", err.Error())
		return err
	}
	for _, f := range frags {
		t := f.CreateTriple()
		switch t.Predicate.RawValue() {
		case "https://archief.nl/def/manifest", "https://archief.nl/def/manifests":
			link := strings.Replace(
				t.Object.RawValue(),
				"/hubID",
				fmt.Sprintf("/%s", excludeHubID),
				1,
			)
			t.Object = r.NewLiteral(link)
		}
		// log.Printf("resolved triple: %#v", t)
		err = rm.AppendTriple(t, true)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetPath sets the full context path for the Fragment that can be used
// for statistics aggregations.
func (f *Fragment) SetPath(contextPath string) {
	rdfType := "rdf_Description"
	if len(f.GetResourceType()) > 0 {
		rdfType = f.GetResourceType()[0]
		searchLabel, err := c.Config.NameSpaceMap.GetSearchLabel(rdfType)
		if err != nil {
			logLabelErr(rdfType, err)
		}
		if searchLabel != "" {
			rdfType = searchLabel
		}
	}
	typePath := fmt.Sprintf("%s/%s", contextPath, rdfType)
	path := fmt.Sprintf("%s/%s", typePath, f.SearchLabel)
	typedLabel := fmt.Sprintf("%s/%s", rdfType, f.SearchLabel)
	switch {
	case f.Predicate == RDFType:
		f.NestedPath = append(
			f.NestedPath,
			fmt.Sprintf("%s/@rdf:about", typePath),
			fmt.Sprintf("%s/@rdf:type", typePath),
		)
		f.Path = append(
			f.Path,
			fmt.Sprintf("%s/@rdf:about", rdfType),
			fmt.Sprintf("%s/@rdf:type", rdfType),
		)
	case f.ObjectType == resourceType:
		f.NestedPath = append(
			f.NestedPath,
			fmt.Sprintf("%s/@rdf:resource", path),
		)
		f.Path = append(
			f.Path,
			fmt.Sprintf("%s/@rdf:resource", typedLabel),
		)
	default:
		if f.Language != "" {
			f.NestedPath = append(
				f.NestedPath,
				fmt.Sprintf("%s/@xml:lang", path),
			)
			f.Path = append(
				f.Path,
				fmt.Sprintf("%s/@xml:lang", typedLabel),
			)
		}
		if f.DataType != "" {
			f.NestedPath = append(
				f.NestedPath,
				fmt.Sprintf("%s/@xsd:type", path),
			)
			f.Path = append(
				f.Path,
				fmt.Sprintf("%s/@xsd:type", typedLabel),
			)
		}
		f.NestedPath = append(f.NestedPath, path)
		f.Path = append(f.Path, typedLabel)
	}

	return
}

// CreateTriple creates a *rdf2go.Triple from a Fragment
func (f *Fragment) CreateTriple() *r.Triple {
	s := r.NewResource(f.Subject)
	p := r.NewResource(f.Predicate)
	var o r.Term

	switch f.ObjectType {
	case resourceType:
		o = r.NewResource(f.Object)
	case bnode:
		o = r.NewBlankNode(f.Object)
	default:
		if f.DataType != "" {
			o = r.NewLiteralWithDatatype(
				f.Object,
				r.NewResource(f.DataType),
			)
			t := r.NewTriple(s, p, o)
			return t
		}
		o = r.NewLiteralWithLanguage(f.Object, f.Language)
	}

	t := r.NewTriple(s, p, o)
	return t
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// containsContext determines if a FragmentReferrerContext is already part of list
// this deduplication is important to not provide false counts for context levels
func containsContext(s []*FragmentReferrerContext, e *FragmentReferrerContext) bool {
	for _, a := range s {
		if a.ObjectID == e.ObjectID && a.Predicate == e.Predicate {
			return true
		}
	}
	return false
}

// containtsEntry determines if a FragmentEntry is already part of a predicate list
func containsEntry(s []*FragmentEntry, e *FragmentEntry) bool {
	for _, a := range s {
		if a.ID == e.ID && a.Value == e.Value {
			return true
		}
	}
	return false
}

// debrack removes the brackets around a string representation of a triple
func debrack(s string) string {
	if len(s) < 2 {
		return s
	}
	if s[0] != '<' {
		return s
	}
	if s[len(s)-1] != '>' {
		return s
	}
	return s[1 : len(s)-1]
}

// CreateFragmentEntry creates a FragmentEntry from a triple
func CreateFragmentEntry(t *r.Triple, resolved bool, order int) (*FragmentEntry, string) {
	entry := &FragmentEntry{Triple: t.String()}
	entry.Order = order
	entry.Resolved = resolved

	switch o := t.Object.(type) {
	case *r.Resource:
		id := r.GetResourceID(o)
		entry.ID = r.GetResourceID(o)
		entry.EntryType = resourceType
		return entry, id
	case *r.BlankNode:
		id := r.GetResourceID(o)
		entry.ID = r.GetResourceID(o)
		entry.EntryType = bnode
		return entry, id
	case *r.Literal:
		entry.Value = o.Value
		entry.EntryType = literal
		if o.Datatype != nil && len(o.Datatype.String()) > 0 {
			if o.Datatype.String() != "<http://www.w3.org/2001/XMLSchema#string>" {
				entry.DataType = debrack(o.Datatype.String())
			}
			switch o.Datatype.String() {
			case "<http://www.w3.org/2001/XMLSchema#string>":
			default:
				entry.DataType = debrack(o.Datatype.String())
			}

		}
		if len(o.Language) > 0 {
			entry.Language = o.Language
		}
	}
	return entry, ""
}

// AppendTriple appends a triple to a subject map
func (rm *ResourceMap) AppendTriple(t *r.Triple, resolved bool) error {
	return rm.AppendOrderedTriple(t, resolved, 0)
}

// AppendOrderedTriple appends a triple to a subject map
func (rm *ResourceMap) AppendOrderedTriple(t *r.Triple, resolved bool, order int) error {
	id := t.GetSubjectID()
	fr, ok := rm.resources[id]
	if !ok {
		fr = &FragmentResource{}
		fr.ID = id
		rm.resources[id] = fr
		fr.predicates = make(map[string][]*FragmentEntry)
	}

	ttype, ok := t.GetRDFType()
	if ok {
		if !contains(fr.Types, ttype) {
			fr.Types = append(fr.Types, ttype)
		}
		return nil
	}

	p := r.GetResourceID(t.Predicate)
	predicates, ok := fr.predicates[p]
	if !ok {
		predicates = []*FragmentEntry{}
	}

	entry, fragID := CreateFragmentEntry(t, resolved, order)
	if fragID != "" {
		if fragID != id {
			ctx := fr.NewContext(p, fragID)
			if !containsContext(fr.objectIDs, ctx) {
				fr.objectIDs = append(fr.objectIDs, ctx)
			}
		}
	}
	if !containsEntry(predicates, entry) {
		fr.predicates[p] = append(predicates, entry)
	}

	return nil
}

// Resources returns the map
func (rm *ResourceMap) Resources() map[string]*FragmentResource {
	return rm.resources
}

// GetResource returns a Fragment resource from the ResourceMap
func (rm *ResourceMap) GetResource(subject string) (*FragmentResource, bool) {
	fr, ok := rm.resources[subject]
	return fr, ok
}

// GetLevel returns the relative level that this resource has from the root
// or parent resource
func (fr *FragmentResource) GetLevel() int32 {
	highestLevel := int32(0)
	for _, ctx := range fr.Context {
		if ctx.GetLevel() > highestLevel {
			highestLevel = ctx.GetLevel()
		}
	}

	return highestLevel + 1
}

// NewResultSummary creates a Summary from the FragmentGraph based on the
// RDFTag configuration.
func (fg *FragmentGraph) NewResultSummary() *ResultSummary {
	fg.Summary = &ResultSummary{}
	for _, rsc := range fg.Resources {
		for _, entry := range rsc.Entries {
			fg.Summary.AddEntry(entry)
		}
	}
	return fg.Summary
}

// NewFields returns a map of the triples sorted by their searchLabel
func (fg *FragmentGraph) NewFields(tq *memory.TextQuery, fields ...string) map[string][]string {
	if tq != nil {
		tq.Reset()
	}

	fieldMap := make(map[string]map[string]int)

	includeMap := make(map[string]bool)
	for _, field := range fields {
		includeMap[field] = true
	}

	for _, rsc := range fg.Resources {
		for _, entry := range rsc.Entries {
			var entryKey string
			switch entry.EntryType {
			case "Resource":
				entryKey = entry.ID
			default:
				entryKey = entry.Value
			}

			if entryKey == "" {
				continue
			}

			_, ok := includeMap[entry.SearchLabel]
			if !ok && tq != nil {
				continue
			}

			nd, ok := fieldMap[entry.SearchLabel]
			if !ok {
				fd := make(map[string]int)
				fd[entryKey] = entry.Order
				fieldMap[entry.SearchLabel] = fd
				continue
			}
			_, ok = nd[entryKey]
			if !ok {
				nd[entryKey] = entry.Order
			}
		}
	}

	fg.Fields = make(map[string][]string)

	if tq == nil {
		type posMap struct {
			Key   string
			Value int
		}
		for name, u := range fieldMap {
			// The map u from fieldMap is always unordered
			posItems := make([]posMap, 0)
			for key, position := range u {
				if key != "" {
					posItems = append(posItems, posMap{key, position})
				}
			}
			if len(posItems) > 0 {
				sort.Slice(posItems, func(i, j int) bool {
					return posItems[i].Value < posItems[j].Value
				})
				sortValues := make([]string, 0)
				for _, n := range posItems {
					sortValues = append(sortValues, n.Key)
				}
				fg.Fields[name] = sortValues
			}
		}

		// log.Printf("flat fields: %#v", fg.Fields)

		return fg.Fields
	}

	type hlEntry struct {
		searchLabel string
		docID       int
		text        string
	}

	hlFields := []hlEntry{}

	for searchLabel, rawFields := range fieldMap {
		for field, order := range rawFields {
			if field != "" {
				if tq != nil {
					indexErr := tq.AppendString(field, order)
					if indexErr != nil {
						log.Printf("index error: %#v", indexErr)
					}
				}
				hlFields = append(hlFields, hlEntry{
					searchLabel: searchLabel,
					docID:       order,
					text:        field,
				})
			}
		}
	}

	// keep fields by docID
	sort.Slice(hlFields, func(i, j int) bool {
		return hlFields[i].docID < hlFields[j].docID
	})

	if tq != nil {
		_, searchErr := tq.PerformSearch()
		if searchErr != nil {
			log.Printf("unable to do field search = %+v\n", searchErr)
		}
	}

	flatFields := map[string][]string{}
	for _, field := range hlFields {
		var text string
		if tq != nil {
			text, _ = tq.Highlight(field.text, field.docID)
		} else {
			text = field.text
		}

		fieldValue, ok := flatFields[field.searchLabel]
		if !ok {
			fieldValue = []string{}
		}
		flatFields[field.searchLabel] = append(fieldValue, text)

	}
	fg.Fields = flatFields

	return fg.Fields
}

// NewTree returns the output as navigation tree
func (fg *FragmentGraph) NewTree() *Tree {
	return fg.Tree
}

// NewJSONLD creates a JSON-LD version of the FragmentGraph
func (fg *FragmentGraph) NewJSONLD() []map[string]interface{} {
	fg.JSONLD = []map[string]interface{}{}
	ids := map[string]bool{}
	for _, rsc := range fg.Resources {
		if _, ok := ids[rsc.ID]; ok {
			continue
		}
		fg.JSONLD = append(fg.JSONLD, rsc.GenerateJSONLD())
		ids[rsc.ID] = true
	}
	return fg.JSONLD
}

// NewGrouped returns an inlined version of the FragmentResources in the FragmentGraph
func (fg *FragmentGraph) NewGrouped() (*FragmentResource, error) {
	rm := &ResourceMap{
		resources: make(map[string]*FragmentResource),
		orgID:     fg.Meta.OrgID,
	}

	// create the resource map
	for _, fr := range fg.Resources {
		// log.Printf("%#v", fr.ID)
		rm.resources[fr.ID] = fr
	}

	// set the inlines
	for _, fr := range fg.Resources {
	Loop:
		for _, entry := range fr.Entries {
			if entry.ID != "" && fr.ID != entry.ID {
				target, ok := rm.GetResource(entry.ID)
				if ok {
					// log.Printf("\n\n%d.%d %#v %s", idx, idx2, fr.ID, target.ID)
					for _, c := range fr.Context {
						if target.ID == c.GetObjectID() {
							continue Loop
						}
					}
					entry.Inline = target
				}
			}
		}
	}

	// only return the subject
	subject, ok := rm.GetResource(fg.GetAboutURI())

	if !ok {
		return nil, fmt.Errorf("unable to find root of the graph for %s", fg.GetAboutURI())
	}

	fg.Resources = []*FragmentResource{subject}
	return subject, nil
}

// AddEntry adds Summary fields based on the ResourceEntry tags
func (sum *ResultSummary) AddEntry(entry *ResourceEntry) {
	// TODO(kiivihal): decide on returning []string instead of string
	for _, tag := range entry.Tags {
		switch tag {
		case "title":
			if sum.Title == "" {
				sum.Title = entry.Value
			}
		case "thumbnail":
			// Always prefer edm:object for the thumbnail.
			// This also ensures that first webresource is used
			if entry.SearchLabel == "edm_object" {
				sum.Thumbnail = entry.Value
			}

			if sum.Thumbnail == "" {
				sum.Thumbnail = entry.Value
			}
		case "subject":
			if sum.Subject == "" {
				sum.Subject = entry.Value
			}
		case "creator":
			if sum.Creator == "" {
				sum.Creator = entry.Value
			}
		case "description":
			if sum.Description == "" {
				sum.Description = entry.Value
			}
		case "landingPage":
			if sum.LandingPage == "" {
				sum.LandingPage = entry.Value
			}
		case "collection":
			if sum.Collection == "" {
				sum.Collection = entry.Value
			}
		case "subCollection":
			if sum.SubCollection == "" {
				sum.SubCollection = entry.Value
			}
		case "objectType":
			if sum.ObjectType == "" {
				sum.ObjectType = entry.Value
			}
		case "objectID":
			if sum.ObjectID == "" {
				sum.ObjectID = entry.Value
			}
		case "owner":
			if sum.Owner == "" {
				sum.Owner = entry.Value
			}
		case "date":
			if sum.Date == "" {
				sum.Date = entry.Value
			}
		}
	}
}

// CreateHeader Linked Data Fragment entry for ElasticSearch
// as described here: http://linkeddatafragments.org/.
//
// The goal of this document is to support Linked Data Fragments based resolving
// for all stored RDF triples in the Hub3 system.
func (fg *FragmentGraph) CreateHeader(docType string) *Header {
	h := &Header{
		OrgID:    fg.Meta.OrgID,
		Spec:     fg.Meta.Spec,
		Revision: fg.Meta.Revision,
		HubID:    fg.Meta.HubID,
		DocType:  docType,
		Modified: NowInMillis(),
	}
	return h
}

// AddTags adds a tag string to the tags array of the Header
func (m *Header) AddTags(tags ...string) {
	for _, tag := range tags {
		if !contains(m.Tags, tag) {
			m.Tags = append(m.Tags, tag)
		}
	}
}

// AddTags adds a tag string to the tags array of the Header
func (re *ResourceEntry) AddTags(tags ...string) {
	for _, tag := range tags {
		if !contains(re.Tags, tag) {
			re.Tags = append(re.Tags, tag)
		}
	}
}

// CreateLodKey returns the path including the # fragments from the subject URL
// This is used for the Linked Open Data resolving
func (fr *FragmentResource) CreateLodKey() (string, error) {
	u, err := url.Parse(fr.ID)
	if err != nil {
		return "", err
	}
	lodKey := u.Path
	if c.Config.LOD.SingleEndpoint == "" {
		lodResourcePrefix := fmt.Sprintf("/%s", c.Config.LOD.Resource)
		if !strings.HasPrefix(u.Path, lodResourcePrefix) {
			return "", nil
		}
		lodKey = strings.TrimPrefix(u.Path, lodResourcePrefix)
	}
	if u.Fragment != "" {
		lodKey = fmt.Sprintf("%s#%s", lodKey, u.Fragment)
	}
	return lodKey, nil
}

// NormalisedResource creates a unique BlankNode key
// Normal resources are returned as is.
//
// This function is used so that you can query via the Fragment API for
// unique BlankNodes
func (fg *FragmentGraph) NormalisedResource(uri string) string {
	if !strings.HasPrefix(uri, "_:") {
		return uri
	}
	// TODO(kiivihal): investigate this
	// return fmt.Sprintf("%s-%s", uri, CreateHash(fg.Meta.NamedGraphURI))
	return strings.ToLower(uri)
}

// CreateFragments creates ElasticSearch documents for each
// RDF triple in the FragmentResource
func (fr *FragmentResource) CreateFragments(fg *FragmentGraph) ([]*Fragment, error) {
	fragments := []*Fragment{}

	lodKey, _ := fr.CreateLodKey()

	typeLabel, err := c.Config.NameSpaceMap.GetSearchLabel(RDFType)
	if err != nil {
		logLabelErr(RDFType, err)
		typeLabel = ""
	}
	path := fr.ContextPath()
	types := []string{}
	for _, ttype := range fr.Types {
		types = append(types, ttype)
		frag := &Fragment{
			Meta:         fg.CreateHeader(FragmentDocType),
			Subject:      fg.NormalisedResource(fr.ID),
			Predicate:    RDFType,
			Object:       ttype,
			ObjectType:   resourceType,
			ResourceType: types,
			SearchLabel:  typeLabel,
			Level:        fr.GetLevel(),
		}
		frag.SetPath(path)
		frag.Meta.NamedGraphURI = fg.Meta.NamedGraphURI
		if strings.HasPrefix(fr.ID, "_:") {
			frag.Triple = fmt.Sprintf("%s <%s> <%s> .", frag.Subject, RDFType, ttype)
		} else {
			frag.Triple = fmt.Sprintf("<%s> <%s> <%s> .", fr.ID, RDFType, ttype)
		}
		frag.Meta.AddTags("typelink", "Resource")
		if lodKey != "" {
			frag.LodKey = lodKey
		}
		fragments = append(fragments, frag)
	}

	// add entries
	for predicate, entries := range fr.predicates {
		for _, entry := range entries {

			label, err := c.Config.NameSpaceMap.GetSearchLabel(predicate)
			if err != nil {
				logLabelErr(predicate, err)
				label = ""
			}

			frag := &Fragment{
				Meta:         fg.CreateHeader(FragmentDocType),
				Subject:      fg.NormalisedResource(fr.ID),
				Predicate:    predicate,
				DataType:     entry.DataType,
				Language:     entry.Language,
				ObjectType:   entry.EntryType,
				Order:        int32(entry.Order),
				ResourceType: types,
				SearchLabel:  label,
				Level:        fr.GetLevel(),
			}
			frag.SetPath(path)
			frag.Meta.NamedGraphURI = fg.Meta.NamedGraphURI
			if entry.ID != "" {
				frag.Object = fg.NormalisedResource(entry.ID)
			} else {
				frag.Object = entry.Value
			}
			frag.Triple = strings.Replace(entry.Triple, entry.ID, fg.NormalisedResource(entry.ID), -1)
			frag.Triple = strings.Replace(frag.Triple, fr.ID, frag.Subject, -1)
			frag.Meta.AddTags(entry.EntryType)
			if lodKey != "" {
				frag.LodKey = lodKey
			}
			fragments = append(fragments, frag)
		}
	}
	return fragments, nil
}

// GetXSDLabel returns a namespaced label for the RDF datatype
func (fe *FragmentEntry) GetXSDLabel() string {
	return strings.Replace(fe.DataType, "http://www.w3.org/2001/XMLSchema#", "xsd:", 1)
}

// IndexFragments updates the Fragments for standalone indexing and adds them to the Elastic BulkProcessorService
func (fb *FragmentBuilder) IndexFragments(bi BulkIndex) error {
	rm, err := fb.ResourceMap()
	if err != nil {
		return err
	}

	return IndexFragments(rm, fb.FragmentGraph(), bi)
}

// IndexFragments updates the Fragments for standalone indexing and adds them to the Elastic BulkProcessorService
func IndexFragments(rm *ResourceMap, fg *FragmentGraph, bi BulkIndex) error {
	for _, fr := range rm.Resources() {
		fragments, err := fr.CreateFragments(fg)
		if err != nil {
			return err
		}
		for _, frag := range fragments {
			err := frag.AddTo(bi)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// NowInMillis returns time.Now() in miliseconds
func NowInMillis() int64 {
	return time.Now().UTC().UnixMilli()
}

// LastModified converts millis into time.Time
func LastModified(millis int64) time.Time {
	return time.Unix(0, millis*int64(time.Millisecond))
}
