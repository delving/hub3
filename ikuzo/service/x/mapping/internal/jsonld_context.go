package internal

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
)

type ResourceLookup interface {
	Get(label string) (*Resource, bool)
}

type JSONLDSchema struct {
	Context      map[string]any `json:"@context"`
	Resources    map[string]*Resource
	Predicates   map[string]*Predicate
	RootResource string
}

type Predicate struct {
	Container  string `json:"@container"`
	ID         string `json:"@id"`
	Type       string `json:"@type"`
	Resource   string
	Label      string
	nsID       string
	TargetNode Node
}

type Node struct {
	ID         int
	ClassLabel string
	ClassURI   string
	ClassNS    string
}

type Resource struct {
	Label      string
	ClassURI   string
	Predicates map[string]*Predicate
	DomainNode Node
}

func (rsc *Resource) Elems(parent *Celem, lookup ResourceLookup) error {
	child := &Celem{
		Attrattrs: "rdf:about",
		Attrtag:   rsc.DomainNode.ClassNS,
		Attrlabel: rsc.Label,
		Cattr:     []*Cattr{},
		Celem:     []*Celem{},
	}

	for _, pred := range rsc.Predicates {
		pElem := &Celem{
			XMLName:   xml.Name{},
			Attrtag:   pred.nsID,
			Attrlabel: pred.Label,
			Cattr:     []*Cattr{},
			Celem:     []*Celem{},
		}

		switch pred.Type {
		case "@id":
			pElem.Attrattrs = "rdf:resource"

			targetRsc, ok := lookup.Get(pred.TargetNode.ClassLabel)
			if ok {
				if err := targetRsc.Elems(pElem, lookup); err != nil {
					return err
				}
			} else if pred.TargetNode.ClassLabel != "" {
				return fmt.Errorf("unknown resource: %#v", pred)
			}
		default:
			pElem.Attrattrs = "xml:lang,rdf:resource"
		}

		child.Celem = append(child.Celem, pElem)
	}

	sort.Slice(
		child.Celem,
		func(i, j int) bool { return child.Celem[i].Attrlabel < child.Celem[j].Attrlabel },
	)

	parent.Celem = append(parent.Celem, child)

	return nil
}

func (schema JSONLDSchema) Get(label string) (*Resource, bool) {
	rsc, ok := schema.Resources[label]
	return rsc, ok
}

func (schema JSONLDSchema) Elems() ([]*Celem, error) {
	var elems []*Celem
	elem := &Celem{
		Attrattrs: "rdf:about",
		Attrtag:   schema.RootResource,
		Cattr:     []*Cattr{},
		Celem:     []*Celem{},
	}

	elems = append(elems, elem)

	for _, rsc := range schema.Resources {
		err := rsc.Elems(elem, schema)
		if err != nil {
			return elems, err
		}
	}

	return elems, nil
}

func ParseJSONLD(r io.Reader) (*JSONLDSchema, error) {
	schema := JSONLDSchema{
		Resources:  map[string]*Resource{},
		Predicates: map[string]*Predicate{},
	}

	if err := json.NewDecoder(r).Decode(&schema); err != nil {
		return nil, err
	}

	for label, v := range schema.Context {
		switch data := v.(type) {
		case string:
			rsc, ok := schema.Resources[label]
			if !ok {
				rsc = &Resource{
					Label:      label,
					Predicates: map[string]*Predicate{},
				}
			}
			rsc.ClassURI = data
			schema.Resources[label] = rsc

		case map[string]interface{}:
			pred := Predicate{
				Label: label,
			}

			if strings.Contains(label, ".") {
				parts := strings.SplitN(label, ".", 2)
				pred.Resource = parts[0]
				pred.Label = parts[1]
			}

			for k, v := range data {
				switch k {
				case "@id":
					pred.ID = v.(string)
				case "@type":
					pred.Type = v.(string)
				case "@container":
					pred.Container = v.(string)
				}
			}

			if pred.Resource != "" {
				rsc, ok := schema.Resources[pred.Resource]
				if !ok {
					rsc = &Resource{
						Label:      pred.Resource,
						Predicates: map[string]*Predicate{},
					}
				}

				rsc.Predicates[pred.Label] = &pred

				schema.Resources[pred.Resource] = rsc

				// continue
			}

			schema.Predicates[label] = &pred
		default:
			return nil, fmt.Errorf("unknown type in jsonld: %#v", v)
		}
	}

	return &schema, nil
}
