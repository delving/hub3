package pid

import (
	"fmt"
	"time"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/rdf"
)

const (
	pidPredicateIRI = "http://schemas.delving.eu/nave/terms/pid"
	isShownAtIRI    = "http://www.europeana.eu/schemas/edm/isShownAt"
)

func Extract(hubID domain.HubID, g *rdf.Graph) (*PID, error) {
	p, err := rdf.NewIRI(pidPredicateIRI)
	if err != nil {
		return nil, fmt.Errorf("unable to create predicate: %w", err)
	}

	isShownAt, err := rdf.NewIRI(isShownAtIRI)
	if err != nil {
		return nil, fmt.Errorf("unable to create predicate: %w", err)
	}

	for _, t := range g.Triples() {
		if !t.Predicate.Equal(p) {
			continue
		}
		pid := &PID{
			ID:         t.Subject.RawValue(),
			ExternalID: t.Object.RawValue(),
			ModifiedAt: time.Now(),
			Meta: Meta{
				OrgID:   hubID.OrgID(),
				DataSet: hubID.DatasetID.String(),
				HubID:   hubID.String(),
			},
		}

		if err := pid.cleanID(); err != nil {
			return nil, fmt.Errorf("unable to clean ID: %w", err)
		}

		pid.inferType()

		rsc, ok := g.Get(t.Subject)
		if !ok {
			return nil, fmt.Errorf("unable to get resource: %s", t.Subject)
		}

		a, ok := rsc.Predicates()[rdf.Predicate(isShownAt)]
		if !ok {
			return pid, fmt.Errorf("unable to get isShownAt predicate")
		}

		for _, obj := range a.Objects() {
			if obj.RawValue() == "" {
				continue
			}
			pid.IsShownAt = obj.RawValue()
		}

		return pid, nil
	}

	return nil, nil
}
