package ead

import (
	"fmt"
	"strings"

	"github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	r "github.com/kiivihal/rdf2go"
)

const FragmentGraphDocType = "ead"

const CLevelLeader = "@"

// Node holds all the clevel information.
type Node struct {
	CTag               string
	Depth              int32
	Type               string
	SubType            string
	Header             *Header
	Nodes              []*Node
	Order              uint64
	ParentIDs          []string
	Path               string
	BranchID           string
	AccessRestrict     string
	AccessRestrictYear string
	Material           string
	Phystech           []string
	triples            []*r.Triple
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
	Physloc          string
	DateAsLabel      bool
	HasDigitalObject bool
	DaoLink          string
	AltRender        string
	Genreform        string
	Attridentifier   string
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
	return fmt.Sprintf("%s/%s/archive/%s/%s",
		config.Config.RDF.BaseURL, config.Config.OrgID, cfg.Spec, id)
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
			strings.Replace(id, "/", "-", -1),
		),
		DocType:       fragments.FragmentGraphDocType,
		EntryURI:      subject,
		NamedGraphURI: fmt.Sprintf("%s/graph", subject),
		Modified:      fragments.NowInMillis(),
		Tags:          []string{"ead"},
	}

	for idx, t := range n.Triples(cfg) {
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
	tree.RawContent = []string{}
	for _, t := range n.triples {
		switch t.Predicate.RawValue() {
		case NewResource("unitTitle").RawValue():
		case NewResource("geogname").RawValue():
		case NewResource("persname").RawValue():
		case NewResource("datetext").RawValue():
		case NewResource("dateiso").RawValue():
		default:
			tree.RawContent = append(tree.RawContent, t.Object.RawValue())
		}
	}
	tree.Access = n.AccessRestrict
	tree.HasRestriction = n.AccessRestrict != ""
	tree.PhysDesc = n.Header.Physdesc

	return tree
}

// GetSubject creates subject URI for the parent Node
// the header itself is an anonymous BlankNode
func (n *Node) GetSubject(cfg *NodeConfig) string {
	id := n.Path
	return newSubject(cfg, id)
}

type convert func(string) r.Term

func addNonEmptyTriple(s r.Term, p, o string, oType convert) *r.Triple {
	if o == "" {
		return nil
	}
	return r.NewTriple(
		s,
		NewResource(p),
		oType(o),
	)
}

// Triples returns a list of triples created from an Archive Node
// Nested elements are linked as object references
func (n *Node) Triples(cfg *NodeConfig) []*r.Triple {
	subject := n.GetSubject(cfg)
	s := r.NewResource(subject)
	triples := n.triples

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
	t(s, "genreform", n.Header.Genreform, r.NewLiteral)

	for _, p := range cfg.PeriodDesc {
		t(s, "periodDesc", p, r.NewLiteral)
	}

	return triples
}
