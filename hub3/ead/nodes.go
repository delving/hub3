package ead

import (
	"fmt"

	"github.com/delving/rapid-saas/config"
	"github.com/delving/rapid-saas/hub3/fragments"
	r "github.com/kiivihal/rdf2go"
)

const FragmentGraphDocType = "ead"

func newSubject(cfg *NodeConfig, id string) string {
	return fmt.Sprintf("%s/archive/%s/%s", config.Config.RDF.BaseURL, cfg.Spec, id)
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
	fg.SetResources(rm)
	return fg, rm, nil
}

// Triples create a list of RDF triples from a NodeID
func (ni *NodeID) Triples(referrer r.Term, cfg *NodeConfig) []*r.Triple {
	s := r.NewAnonNode()
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

	s := r.NewAnonNode()
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
	for _, nodeID := range h.GetID() {
		triples = append(triples, nodeID.Triples(s, cfg)...)
	}

	return triples
}

// Triples returns all the triples for a NodeDate
func (nd *NodeDate) Triples(referrer r.Term, cfg *NodeConfig) []*r.Triple {
	s := r.NewAnonNode()
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

	parentSubject := s
	for i := len(n.ParentIDs) - 1; i >= 0; i-- {
		objectSubject := r.NewResource(newSubject(cfg, n.ParentIDs[i]))
		parent := r.NewTriple(
			parentSubject,
			r.NewResource(fmt.Sprintf("http://archief.nl/def/ead/%s", "hasParent")),
			objectSubject,
		)

		parentSubject = objectSubject
		label, ok := cfg.labels[n.ParentIDs[i]]
		if ok {
			objectLabel := r.NewTriple(
				parentSubject,
				r.NewResource("http://www.w3.org/2000/01/rdf-schema#label"),
				r.NewLiteral(label),
			)
			triples = append(triples, parent, objectLabel)
			continue
		}
		triples = append(triples, parent)
	}

	return triples
}

// store recursive function on nodelist for fragments and fragment graph

// create fragmentGraph

// node to triples

// create subject URI function: base, tentant, spec, inventoryID
