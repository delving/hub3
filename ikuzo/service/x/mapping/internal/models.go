package internal

import (
	"fmt"
	"io"
	"log"
	"sort"
	"strings"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/service/x/namespace"
	"github.com/gocarina/gocsv"
)

type ModelRow struct {
	NodeSource    string `json:"nodeSource" csv:"node_source"`
	ClassDomain   string `json:"classDomain" csv:"class_domain"`
	ClassDomainID string `json:"classDomainID" csv:"class_domain.class_id"`
	Property      string `json:"property" csv:"property"`
	PropertyID    string `json:"propertyID" csv:"property_id"`
	NodeTarget    string `json:"nodeTarget" csv:"node_target"`
	ClassRange    string `json:"classRange" csv:"class_range"`
	ClassRangeID  string `json:"classRangeID" csv:"class_range.class_id"`
	// Inline        *ModelRow `json:"inline,omitempty" csv:"-"`
}

func (row *ModelRow) newResource(m *Model) *Resource {
	rsc := &Resource{
		Label:      row.ClassDomain,
		ClassURI:   row.ClassDomainID,
		Predicates: map[string]*Predicate{},
		DomainNode: Node{
			ID:         m.getNextNodeID(),
			ClassLabel: row.ClassDomain,
			ClassURI:   row.ClassDomainID,
		},
	}

	ns, ok := m.getNS(row.ClassDomainID)
	if ok {
		rsc.DomainNode.ClassNS = ns
	}

	return rsc
}

type Model struct {
	RootResource string
	nextNodeID   int
	rows         []ModelRow
	resources    map[string]*Resource
	ns           *namespace.Service
}

func newModel() (*Model, error) {
	svc, err := namespace.NewService(namespace.WithDefaults())
	if err != nil {
		return nil, err
	}

	model := &Model{
		nextNodeID: 0,
		rows:       []ModelRow{},
		resources:  map[string]*Resource{},
		ns:         svc,
	}

	return model, nil
}

func (m *Model) Get(label string) (*Resource, bool) {
	rsc, ok := m.resources[label]
	return rsc, ok
}

func (m *Model) getNS(uri string) (string, bool) {
	base, label := domain.SplitURI(uri)

	predNS, err := m.ns.GetWithBase(base)
	if err != nil && base != "" {
		log.Printf("not found %q; %s", base, err)
		return "", false
	}

	if predNS == nil {
		return "", false
	}

	return fmt.Sprintf("%s:%s", predNS.Prefix, label), true
}

func (m *Model) inline() error {
	if len(m.resources) != 0 {
		return fmt.Errorf("you can only call inline once")
	}

	m.resources = map[string]*Resource{}

	for _, row := range m.rows {
		rsc, ok := m.resources[row.ClassDomain]
		if !ok {
			rsc = row.newResource(m)
		}

		m.resources[row.ClassDomain] = rsc

		pred := &Predicate{
			ID:    row.PropertyID,
			Label: row.Property,
			TargetNode: Node{
				ID:         0,
				ClassLabel: row.ClassRange,
				ClassURI:   row.ClassRangeID,
			},
		}

		predNS, ok := m.getNS(pred.ID)
		if ok {
			pred.nsID = predNS
		}

		ns, ok := m.getNS(row.ClassRangeID)
		if ok {
			pred.TargetNode.ClassNS = ns
		}

		switch {
		case strings.EqualFold(row.ClassRange, "TaalString"):
			pred.TargetNode.ID = m.getNextNodeID()
			pred.Type = ""
		case strings.EqualFold(row.ClassRange, "xsd:Decimal"):
			pred.TargetNode.ID = m.getNextNodeID()
			pred.Type = row.ClassRange
		case strings.EqualFold(row.ClassRange, "GetypeerdeString"):
			pred.TargetNode.ID = m.getNextNodeID()
			pred.Type = ""
		default:
			pred.Type = "@id"
			targetRsc, ok := m.resources[row.ClassDomain]
			if !ok {
				targetRsc = row.newResource(m)
				m.resources[row.ClassDomain] = targetRsc
			}

			pred.TargetNode.ID = targetRsc.DomainNode.ID
		}

		rsc.Predicates[row.Property] = pred
	}

	// TODO(kiivihal): inline the data or maybe not

	return nil
}

func (m *Model) Elems() ([]*Celem, error) {
	var elems []*Celem

	root, ok := m.resources[m.RootResource]
	if !ok {
		return elems, fmt.Errorf("unknown root resource %q", m.RootResource)
	}

	elem := &Celem{
		Attrattrs: "rdf:RDF",
		Cattr:     []*Cattr{},
		Celem:     []*Celem{},
	}

	elems = append(elems, elem)

	err := root.Elems(elem, m)
	if err != nil {
		return elems, err
	}

	sort.Slice(
		elem.Celem,
		func(i, j int) bool { return elem.Celem[i].Attrlabel < elem.Celem[j].Attrlabel },
	)

	return elems, nil
}

func (m *Model) getNextNodeID() int {
	if len(m.resources) == 0 {
		return m.nextNodeID
	}

	m.nextNodeID++

	return m.nextNodeID
}

func ParseModel(r io.Reader) (*Model, error) {
	model, err := newModel()
	if err != nil {
		return nil, err
	}
	if err := gocsv.Unmarshal(r, &model.rows); err != nil {
		return model, err
	}

	return model, nil
}
