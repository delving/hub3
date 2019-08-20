package ead

import (
	"fmt"
	"html"
	"strings"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	r "github.com/kiivihal/rdf2go"
)

const FragmentGraphDocType = "ead"

const CLevelLeader = "@"

// Node holds all the clevel information.
type Node struct {
	CTag      string
	Depth     int32
	Type      string
	SubType   string
	Header    *Header
	HTML      []string
	Nodes     []*Node
	Order     uint64
	ParentIDs []string
	Path      string
	BranchID  string
	Access    string
	Material  string
	CLevel    CLevel
}

type NodeList struct {
	Type  string
	Label []string
	Nodes []*Node
}

type Header struct {
	Type             string
	InventoryNumber  string
	ID               []*NodeID
	Label            []string
	Date             []*NodeDate
	Physdesc         string
	DateAsLabel      bool
	HasDigitalObject bool
	DaoLink          string
}
type NodeDate struct {
	Calendar string
	Era      string
	Normal   string
	Label    string
	Type     string
}
type NodeID struct {
	TypeID   string
	Type     string
	Audience string
	ID       string
}

func newSubject(cfg *NodeConfig, id string) string {
	return fmt.Sprintf("%s/NL-HaNA/archive/%s/%s", config.Config.RDF.BaseURL, cfg.Spec, id)
}

// getFirstBranch returs the first parent of the current node
func (n *Node) getFirstBranch() string {
	parents := strings.Split(n.Path, pathSep)
	if len(parents) < 2 {
		return ""
	}
	return fmt.Sprintf("%s%s", CLevelLeader, strings.Join(parents[:len(parents)-1], pathSep))
}

// getSecondBranch returs the second parent of the current node
func (n *Node) getSecondBranch() string {
	parents := strings.Split(n.Path, pathSep)
	if len(parents) < 3 {
		return ""
	}
	return fmt.Sprintf("%s%s", CLevelLeader, strings.Join(parents[:len(parents)-2], pathSep))
}

// FragmentGraph returns the archival node as a FragmentGraph
func (n *Node) FragmentGraph(cfg *NodeConfig) (*fragments.FragmentGraph, *fragments.ResourceMap, error) {
	rm := fragments.NewEmptyResourceMap()
	id := n.Path
	subject := n.GetSubject(cfg)
	header := &fragments.Header{
		OrgID:    cfg.OrgID,
		Spec:     cfg.Spec,
		Revision: cfg.Revision,
		HubID: fmt.Sprintf(
			"%s_%s_%s",
			cfg.OrgID,
			cfg.Spec,
			strings.Replace(id, "/", "-", 0),
		),
		DocType:       fragments.FragmentGraphDocType,
		EntryURI:      subject,
		NamedGraphURI: fmt.Sprintf("%s/graph", subject),
		Modified:      fragments.NowInMillis(),
		Tags:          []string{"ead"},
	}

	for idx, t := range n.Triples(subject, cfg) {
		if err := rm.AppendOrderedTriple(t, false, idx); err != nil {
			return nil, nil, err
		}
	}

	fg := fragments.NewFragmentGraph()
	fg.Meta = header
	fg.Tree = cfg.CreateTree(cfg, n, header.HubID, id)
	fg.SetResources(rm)
	return fg, rm, nil
}

func CreateTree(cfg *NodeConfig, n *Node, hubID string, id string) *fragments.Tree {

	tree := &fragments.Tree{}
	tree.HubID = hubID
	tree.ChildCount = len(n.Nodes)
	tree.Type = n.Type
	tree.CLevel = fmt.Sprintf("%s%s", CLevelLeader, id)
	tree.Label = n.Header.GetTreeLabel()
	tree.UnitID = n.Header.InventoryNumber
	tree.Leaf = n.getFirstBranch()
	tree.Parent = n.getSecondBranch()
	tree.Depth = len(n.ParentIDs) + 1
	tree.HasDigitalObject = n.Header.HasDigitalObject
	tree.DaoLink = n.Header.DaoLink
	tree.SortKey = n.Order
	tree.Periods = n.Header.GetPeriods()
	tree.MimeTypes = []string{}
	tree.ManifestLink = ""
	tree.Content = []string{}
	for _, n := range n.HTML {
		tree.Content = append(tree.Content, html.UnescapeString(n))
	}
	tree.Access = n.Access
	tree.HasRestriction = n.Access != ""
	tree.PhysDesc = n.Header.Physdesc
	return tree
}

// Triples create a list of RDF triples from a NodeID
func (ni *NodeID) Triples(referrer r.Term, order int, cfg *NodeConfig) []*r.Triple {
	s := r.NewResource(fmt.Sprintf("%s/%d", referrer.RawValue(), order))
	triples := []*r.Triple{
		r.NewTriple(
			referrer,
			r.NewResource(fmt.Sprintf("http://archief.nl/def/ead/%s", "unitid")),
			s,
		),
		r.NewTriple(
			s,
			r.NewResource(fragments.RDFType),
			r.NewResource(fmt.Sprintf("http://archief.nl/def/ead/%s", "Unitid")),
		),
	}

	t := func(s r.Term, p, o string, oType convert) {
		t := addNonEmptyTriple(s, p, o, oType)
		if t != nil {
			triples = append(triples, t)
		}
		return
	}
	t(s, "typeID", ni.TypeID, r.NewLiteral)
	t(s, "type", ni.Type, r.NewLiteral)
	t(s, "audience", ni.Audience, r.NewLiteral)
	t(s, "identifier", ni.ID, r.NewLiteral)
	return triples
}

// GetSubject creates subject URI for the parent Node
// the header itself is an anonymous BlankNode
func (n *Node) GetSubject(cfg *NodeConfig) string {
	id := n.Path
	return newSubject(cfg, id)
}

// Triples converts the EAD Did to RDF triples
func (h *Header) Triples(subject string, cfg *NodeConfig) []*r.Triple {

	s := r.NewResource(subject + "/did")
	triples := []*r.Triple{
		r.NewTriple(
			r.NewResource(subject),
			r.NewResource(fmt.Sprintf("http://archief.nl/def/ead/%s", "hasDid")),
			s,
		),
		r.NewTriple(
			s,
			r.NewResource(fragments.RDFType),
			r.NewResource(fmt.Sprintf("http://archief.nl/def/ead/%s", "Did")),
		),
	}
	t := func(s r.Term, p, o string, oType convert) {
		t := addNonEmptyTriple(s, p, o, oType)
		if t != nil {
			triples = append(triples, t)
		}
		return
	}

	t(s, "idUnittype", h.Type, r.NewLiteral)
	t(s, "idDateAsLabel", fmt.Sprintf("%t", h.DateAsLabel), r.NewLiteral)
	t(s, "idInventorynr", h.InventoryNumber, r.NewLiteral)
	t(s, "physdesc", h.Physdesc, r.NewLiteral)

	for _, label := range h.Label {
		t(s, "idUnittitle", label, r.NewLiteral)
	}
	for idx, nodeID := range h.ID {
		triples = append(triples, nodeID.Triples(s, idx, cfg)...)
	}

	for idx, nodeDate := range h.Date {
		triples = append(triples, nodeDate.Triples(s, idx, cfg)...)
	}

	return triples
}

// Triples returns all the triples for a NodeDate
func (nd *NodeDate) Triples(referrer r.Term, order int, cfg *NodeConfig) []*r.Triple {
	s := r.NewResource(fmt.Sprintf("%s/%d", referrer.RawValue(), order))
	triples := []*r.Triple{
		r.NewTriple(
			referrer,
			r.NewResource(fmt.Sprintf("http://archief.nl/def/ead/%s", "unitdate")),
			s,
		),
		r.NewTriple(
			s,
			r.NewResource(fragments.RDFType),
			r.NewResource(fmt.Sprintf("http://archief.nl/def/ead/%s", "Unitdate")),
		),
	}

	t := func(s r.Term, p, o string, oType convert) {
		t := addNonEmptyTriple(s, p, o, oType)
		if t != nil {
			triples = append(triples, t)
		}
		return
	}

	t(s, "dateCalendar", nd.Calendar, r.NewLiteral)
	t(s, "dateEra", nd.Era, r.NewLiteral)
	t(s, "dateNormal", nd.Normal, r.NewLiteral)
	t(s, "dateType", nd.Type, r.NewLiteral)
	return triples
}

type convert func(string) r.Term

func addNonEmptyTriple(s r.Term, p, o string, oType convert) *r.Triple {
	if o == "" {
		return nil
	}
	return r.NewTriple(
		s,
		r.NewResource(fmt.Sprintf("http://archief.nl/def/ead/%s", p)),
		oType(o),
	)
}

// Triples returns a list of triples created from an Archive Node
// Nested elements are linked as object references
func (n *Node) Triples(subject string, cfg *NodeConfig) []*r.Triple {
	s := r.NewResource(subject)
	triples := []*r.Triple{
		r.NewTriple(
			s,
			r.NewResource(fragments.RDFType),
			r.NewResource(fmt.Sprintf("http://archief.nl/def/ead/%s", "Clevel")),
		),
		r.NewTriple(
			s,
			r.NewResource("http://www.w3.org/2000/01/rdf-schema#label"),
			r.NewLiteral(n.Header.GetTreeLabel()),
		),
	}
	t := func(s r.Term, p, o string, oType convert) {
		t := addNonEmptyTriple(s, p, o, oType)
		if t != nil {
			triples = append(triples, t)
		}
		return
	}

	t(s, "cLevel", n.CTag, r.NewLiteral)
	t(s, "branchID", n.BranchID, r.NewLiteral)
	t(s, "cType", n.Type, r.NewLiteral)
	t(s, "cSubtype", n.SubType, r.NewLiteral)
	for _, html := range n.HTML {
		t(s, "scopecontent", html, r.NewLiteral)

	}

	triples = append(triples, n.Header.Triples(subject, cfg)...)

	return triples
}
