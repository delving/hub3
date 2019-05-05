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

func newSubject(cfg *NodeConfig, id string) string {
	return fmt.Sprintf("%s/NL-HaNA/archive/%s/%s", config.Config.RDF.BaseURL, cfg.Spec, id)
}

// getFirstBranch returs the first parent of the current node
func (n *Node) getFirstBranch() string {
	parents := strings.Split(n.GetPath(), pathSep)
	if len(parents) < 2 {
		return ""
	}
	return fmt.Sprintf("%s%s", CLevelLeader, strings.Join(parents[:len(parents)-1], pathSep))
}

// getSecondBranch returs the second parent of the current node
func (n *Node) getSecondBranch() string {
	parents := strings.Split(n.GetPath(), pathSep)
	if len(parents) < 3 {
		return ""
	}
	return fmt.Sprintf("%s%s", CLevelLeader, strings.Join(parents[:len(parents)-2], pathSep))
}

// FragmentGraph returns the archival node as a FragmentGraph
func (n *Node) FragmentGraph(cfg *NodeConfig) (*fragments.FragmentGraph, *fragments.ResourceMap, error) {
	rm := fragments.NewEmptyResourceMap()
	id := n.GetPath()
	subject := n.GetSubject(cfg)
	header := &fragments.Header{
		OrgID:         cfg.OrgID,
		Spec:          cfg.Spec,
		Revision:      cfg.Revision,
		HubID:         fmt.Sprintf("%s_%s_%s", cfg.OrgID, cfg.Spec, id),
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
	tree.Type = n.GetType()
	tree.CLevel = fmt.Sprintf("%s%s", CLevelLeader, id)
	tree.Label = n.GetHeader().GetTreeLabel()
	tree.UnitID = n.GetHeader().GetInventoryNumber()
	tree.Leaf = n.getFirstBranch()
	tree.Parent = n.getSecondBranch()
	tree.Depth = len(n.ParentIDs) + 1
	tree.HasDigitalObject = n.GetHeader().GetHasDigitalObject()
	tree.DaoLink = n.GetHeader().GetDaoLink()
	tree.SortKey = n.GetOrder()
	tree.Periods = n.GetHeader().GetPeriods()
	tree.MimeTypes = []string{}
	tree.ManifestLink = ""
	tree.Content = []string{html.UnescapeString(n.HTML)}

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
	t(s, "typeID", ni.GetTypeID(), r.NewLiteral)
	t(s, "type", ni.GetType(), r.NewLiteral)
	t(s, "audience", ni.GetAudience(), r.NewLiteral)
	t(s, "identifier", ni.GetID(), r.NewLiteral)
	return triples
}

// GetSubject creates subject URI for the parent Node
// the header itself is an anonymous BlankNode
func (n *Node) GetSubject(cfg *NodeConfig) string {
	id := n.GetPath()
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

	t(s, "idUnittype", h.GetType(), r.NewLiteral)
	t(s, "idDateAsLabel", fmt.Sprintf("%t", h.GetDateAsLabel()), r.NewLiteral)
	t(s, "idInventorynr", h.GetInventoryNumber(), r.NewLiteral)
	t(s, "physdesc", h.GetPhysdesc(), r.NewLiteral)

	for _, label := range h.GetLabel() {
		t(s, "idUnittitle", label, r.NewLiteral)
	}
	for idx, nodeID := range h.GetID() {
		triples = append(triples, nodeID.Triples(s, idx, cfg)...)
	}

	for idx, nodeDate := range h.GetDate() {
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

	t(s, "dateCalendar", nd.GetCalendar(), r.NewLiteral)
	t(s, "dateEra", nd.GetEra(), r.NewLiteral)
	t(s, "dateNormal", nd.GetNormal(), r.NewLiteral)
	t(s, "dateType", nd.GetType(), r.NewLiteral)
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
			r.NewLiteral(n.GetHeader().GetTreeLabel()),
		),
	}
	t := func(s r.Term, p, o string, oType convert) {
		t := addNonEmptyTriple(s, p, o, oType)
		if t != nil {
			triples = append(triples, t)
		}
		return
	}

	t(s, "cLevel", n.GetCTag(), r.NewLiteral)
	t(s, "branchID", n.GetBranchID(), r.NewLiteral)
	t(s, "cType", n.GetType(), r.NewLiteral)
	t(s, "cSubtype", n.GetSubType(), r.NewLiteral)
	t(s, "scopecontent", n.GetHTML(), r.NewLiteral)

	triples = append(triples, n.GetHeader().Triples(subject, cfg)...)

	return triples
}
