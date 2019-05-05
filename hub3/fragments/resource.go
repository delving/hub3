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

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/index"
	r "github.com/kiivihal/rdf2go"
	elastic "github.com/olivere/elastic"
	"github.com/pkg/errors"
)

const (
	literal  = "Literal"
	resource = "Resource"
	bnode    = "Bnode"
)

var ctx context.Context

func init() {
	ctx = context.Background()
}

// NewContext returns the context for the current fragmentresource
func (fr *FragmentResource) NewContext(predicate, objectID string) *FragmentReferrerContext {
	searchLabel, err := c.Config.NameSpaceMap.GetSearchLabel(predicate)
	if err != nil {
		log.Printf("Unable to create search label for %s  due to %s\n", predicate, err)
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
}

// Tree holds all the core information for building Navigational Trees from RDF graphs
type Tree struct {
	Leaf             string   `json:"leaf,omitempty"`
	Parent           string   `json:"parent,omitempty"`
	Label            string   `json:"label"`
	CLevel           string   `json:"cLevel"`
	UnitID           string   `json:"unitID"`
	Type             string   `json:"type"`
	HubID            string   `json:"hubID"`
	ChildCount       int      `json:"childCount"`
	Depth            int      `json:"depth"`
	HasChildren      bool     `json:"hasChildren"`
	HasDigitalObject bool     `json:"hasDigitalObject"`
	DaoLink          string   `json:"daoLink,omitempty"`
	ManifestLink     string   `json:"manifestLink,omitempty"`
	MimeTypes        []string `json:"mimeType,omitempty"`
	DOCount          int      `json:"doCount"`
	Inline           []*Tree  `json:"inline,omitempty"`
	SortKey          uint64   `json:"sortKey"`
	Periods          []string `json:"periods"`
	Content          []string `json:"content,omitempty"`
	Title            string   `json:"title,omitempty"`
	Description      string   `json:"description,omitempty"`
	InventoryID      string   `json:"inventoryID,omitempty"`
	AgencyCode       string   `json:"agencyCode,omitempty"`
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
	return tq.Label != "" || tq.UnitID != ""
}

// IsNavigatedQuery returns if there is both a query and active ID
func (tq *TreeQuery) IsNavigatedQuery() bool {
	return tq.Label != "" && tq.UnitID != ""
}

// GetPreviousScrollIDs returns scrollIDs up to the cLevel
// This information can be used to construct the previous search results when
// both the UnitID and the Label are being queried
func (tq *TreeQuery) GetPreviousScrollIDs(cLevel string, sr *SearchRequest, pager *ScrollPager) ([]string, error) {
	previous := []string{}
	query := elastic.NewBoolQuery()

	matchSuffix := fmt.Sprintf("_%s", strings.TrimLeft(cLevel, "@"))

	q := elastic.NewQueryStringQuery(sr.Tree.GetLabel())
	q = q.DefaultField("tree.label")
	if !isAdvancedSearch(sr.Tree.GetLabel()) {
		q = q.MinimumShouldMatch(c.Config.ElasticSearch.MinimumShouldMatch)
	}
	query = query.Must(q)
	query = query.Must(elastic.NewTermQuery(c.Config.ElasticSearch.SpecKey, tq.Spec))

	idSort := elastic.NewFieldSort("meta.hubID")
	fieldSort := elastic.NewFieldSort("tree.sortKey")

	scroll := index.ESClient().Scroll(c.Config.ElasticSearch.IndexName).
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
			//sr.CalculatedTotal = results.TotalHits()

			nextSearchAfter, err := sr.CreateBinKey(hit.Sort)
			if err != nil {
				return nil, errors.Wrap(err, "unable to create bytes for search after key")
			}

			sr.Start = int32(cursor)
			sr.SearchAfter = nextSearchAfter
			hexRequest, err := sr.SearchRequestToHex()
			if err != nil {
				return nil, errors.Wrap(err, "unable to create bytes for search after key")
			}

			if strings.HasSuffix(hit.Id, matchSuffix) {
				//log.Printf("found it: %s ", matchSuffix)
				pager.Cursor = int32(cursor)
				pager.ScrollID = hexRequest
				pager.Total = results.TotalHits()
				return previous, nil // all results retrieved
			}
			previous = append(previous, hexRequest)
			cursor++
		}
	}
}

func (tq *TreeQuery) expandedIDs(lastNode *Tree) map[string]bool {
	expandedIDs := make(map[string]bool)
	parents := strings.Split(tq.GetLeaf(), "~")

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
	if !lastNode.HasChildren {
		expandedIDs[lastNode.CLevel] = false
	}
	return expandedIDs
}

// InlineTree creates a nested tree from an Array of *Tree
func InlineTree(nodes []*Tree, tq *TreeQuery) ([]*Tree, *TreeHeader, error) {
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
		if ok {
			target.Inline = append(target.Inline, n)
		}
	}
	if tq.GetLeaf() == "" {
		return rootNodes, nil, nil
	}

	lastNode, ok := nodeMap[tq.GetLeaf()]
	if !ok {
		return nil, nil, fmt.Errorf("Unable to find node %s in map", tq.GetLeaf())
	}
	header := &TreeHeader{
		ExpandedIDs: tq.expandedIDs(lastNode),
		ActiveID:    tq.GetLeaf(),
		UnitID:      tq.GetUnitID(),
	}
	return rootNodes, header, nil
}

// FragmentGraph is a container for all entries of an RDF Named Graph
type FragmentGraph struct {
	Meta       *Header                   `json:"meta,omitempty"`
	Tree       *Tree                     `json:"tree,omitempty"`
	Resources  []*FragmentResource       `json:"resources,omitempty"`
	Summary    *ResultSummary            `json:"summary,omitempty"`
	JSONLD     []map[string]interface{}  `json:"jsonld,omitempty"`
	Fields     map[string][]string       `json:"fields,omitempty"`
	Highlights []*ResourceEntryHighlight `json:"highlights,omitempty"`
}

// ResourceEntryHighlight holds the values of the ElasticSearch highlight fiel
type ResourceEntryHighlight struct {
	SearchLabel string   `json:"searchLabel"`
	MarkDown    []string `json:"markdown"`
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
		for _, p := range v {
			m[k] = p.AsLdObject()
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

// ScrollResultV4 intermediate non-protobuf search results
type ScrollResultV4 struct {
	Pager      *ScrollPager     `json:"pager"`
	Query      *Query           `json:"query"`
	Items      []*FragmentGraph `json:"items,omitempty"`
	Collapsed  []*Collapsed     `json:"collapse,omitempty"`
	Peek       map[string]int64 `json:"peek,omitempty"`
	Facets     []*QueryFacet    `json:"facets,omitempty"`
	Tree       []*Tree          `json:"tree,omitempty"`
	TreeHeader *TreeHeader      `json:"treeHeader,omitempty"`
}

// TreeHeader contains rendering hints for the consumer of the TreeView API
type TreeHeader struct {
	ExpandedIDs       map[string]bool `json:"expandedIDs,omitempty"`
	ActiveID          string          `json:"activeID"`
	UnitID            string          `json:"unitID"`
	PreviousScrollIDs []string        `json:"previousScrollIDs"`
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
		log.Printf("Unable to create search label for %s  due to %s\n", predicate, err)
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
	}

	labels, ok := c.Config.RDFTagMap.Get(predicate)
	if ok {
		re.AddTags(labels...)
		if re.Value != "" {
			// TODO add validation for the values here
			for _, label := range labels {
				switch label {
				case "isoDate":
					re.Date = re.Value
					//log.Printf("Date value: %s", re.Date)
				case "dateRange":
					re.DateRange = re.Value
				case "latLong":
					re.LatLong = re.Value
				}
			}
		}
	}
	return re, nil
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
	subject, ok := rm.GetResource(subjectURI)
	if !ok {
		return nil, fmt.Errorf("Subject %s is not part of the graph", subjectURI)
	}

	linkedObjects := map[string]*FragmentResource{}
	linkedObjects[subjectURI] = subject
	for _, level1 := range subject.objectIDs {
		level2Resource, ok := rm.GetResource(level1.ObjectID)
		if !ok {
			log.Printf("unknown target URI: %s", level1.ObjectID)
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
	Date        string            `json:"date,omitempty"`
	DateRange   string            `json:"dateRange,omitempty"`
	Integer     int               `json:"integer,omitempty"`
	LatLong     string            `json:"latLong,omitempty"`
	Inline      *FragmentResource `json:"inline,omitempty"`
	Order       int               `json:"order"`
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
func NewResourceMap(g *r.Graph) (*ResourceMap, error) {
	rm := &ResourceMap{make(map[string]*FragmentResource)}

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
func NewEmptyResourceMap() *ResourceMap {
	return &ResourceMap{make(map[string]*FragmentResource)}
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
	//log.Printf("IDs to be resolved: %#v", objectIDs)

	req := NewFragmentRequest()
	req.Subject = objectIDs
	req.ExcludeHubID = excludeHubID
	frags, _, err := req.Find(ctx, index.ESClient())
	if err != nil {
		log.Printf("unable to find fragments: %s", err.Error())
		return err
	}
	for _, f := range frags {
		t := f.CreateTriple()
		//log.Printf("resolved triple: %#v", t)
		err = rm.AppendTriple(t, true)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateTriple creates a *rdf2go.Triple from a Fragment
func (f *Fragment) CreateTriple() *r.Triple {
	s := r.NewResource(f.Subject)
	p := r.NewResource(f.Predicate)
	var o r.Term

	switch f.ObjectType {
	case resource:
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
		entry.EntryType = resource
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
	return int32(highestLevel + 1)
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
func (fg *FragmentGraph) NewFields() map[string][]string {
	fieldMap := make(map[string]map[string]struct{})
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

			nd, ok := fieldMap[entry.SearchLabel]
			if !ok {
				fd := make(map[string]struct{})
				fd[entryKey] = struct{}{}
				fieldMap[entry.SearchLabel] = fd
				continue
			}
			_, ok = nd[entryKey]
			if !ok {
				nd[entryKey] = struct{}{}
			}
		}
	}
	fg.Fields = make(map[string][]string)
	for k, v := range fieldMap {
		fields := []string{}
		for vk := range v {
			if vk != "" {
				fields = append(fields, vk)
			}
		}
		if len(fields) > 0 {
			fg.Fields[k] = fields
		}
	}
	return fg.Fields
}

// NewTree returns the output as navigation tree
func (fg *FragmentGraph) NewTree() *Tree {
	return fg.Tree
}

// NewJSONLD creates a JSON-LD version of the FragmentGraph
func (fg *FragmentGraph) NewJSONLD() []map[string]interface{} {
	fg.JSONLD = []map[string]interface{}{}
	for _, rsc := range fg.Resources {
		fg.JSONLD = append(fg.JSONLD, rsc.GenerateJSONLD())
	}
	return fg.JSONLD
}

// NewGrouped returns an inlined version of the FragmentResources in the FragmentGraph
func (fg *FragmentGraph) NewGrouped() (*FragmentResource, error) {
	rm := &ResourceMap{make(map[string]*FragmentResource)}

	// create the resource map
	for _, fr := range fg.Resources {
		log.Printf("%#v", fr.ID)
		rm.resources[fr.ID] = fr
	}

	// inlining check

	// set the inlines
	for _, fr := range fg.Resources {
	Loop:
		for _, entry := range fr.Entries {
			if entry.ID != "" && fr.ID != entry.ID {

				target, ok := rm.GetResource(entry.ID)
				if ok {
					//log.Printf("\n\n%d.%d %#v %s", idx, idx2, fr.ID, target.ID)
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

	for _, tag := range entry.Tags {
		switch tag {
		case "title":
			if sum.Title == "" {
				sum.Title = entry.Value
			}
		case "thumbnail":
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
func (h *Header) AddTags(tags ...string) {
	for _, tag := range tags {
		if !contains(h.Tags, tag) {
			h.Tags = append(h.Tags, tag)
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
// unique BlankNodesThe named graph that this triple is part of
func (fg *FragmentGraph) NormalisedResource(uri string) string {
	if !strings.HasPrefix(uri, "_:") {
		return uri
	}
	return fmt.Sprintf("%s-%s", uri, CreateHash(fg.Meta.NamedGraphURI))
}

// CreateFragments creates ElasticSearch documents for each
// RDF triple in the FragmentResource
func (fr *FragmentResource) CreateFragments(fg *FragmentGraph) ([]*Fragment, error) {
	fragments := []*Fragment{}

	lodKey, _ := fr.CreateLodKey()

	// TODO add statistics path
	// type is searchLabel
	// @about is extra entry
	// add type links
	for _, ttype := range fr.Types {
		frag := &Fragment{
			Meta:       fg.CreateHeader(FragmentDocType),
			Subject:    fg.NormalisedResource(fr.ID),
			Predicate:  RDFType,
			Object:     ttype,
			ObjectType: resource,
		}
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
			frag := &Fragment{
				Meta:       fg.CreateHeader(FragmentDocType),
				Subject:    fg.NormalisedResource(fr.ID),
				Predicate:  predicate,
				DataType:   entry.DataType,
				Language:   entry.Language,
				ObjectType: entry.EntryType,
				Order:      int32(entry.Order),
			}
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
func (fb *FragmentBuilder) IndexFragments(p *elastic.BulkProcessor) error {
	rm, err := fb.ResourceMap()
	if err != nil {
		return err
	}
	return IndexFragments(rm, fb.FragmentGraph(), p)
}

// IndexFragments updates the Fragments for standalone indexing and adds them to the Elastic BulkProcessorService
func IndexFragments(rm *ResourceMap, fg *FragmentGraph, p *elastic.BulkProcessor) error {

	for _, fr := range rm.Resources() {
		fragments, err := fr.CreateFragments(fg)
		if err != nil {
			return err
		}
		for _, frag := range fragments {
			err := frag.AddTo(p)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// AddGraphExternalContext
