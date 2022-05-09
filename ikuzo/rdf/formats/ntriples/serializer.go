package ntriples

import (
	"fmt"
	"io"
	"strings"

	"github.com/delving/hub3/ikuzo/rdf"
)

func Serialize(g *rdf.Graph, w io.Writer) error {
	var err error

	for _, triple := range g.Triples() {
		_, err = fmt.Fprintf(w, "%s\n", triple.String())
		if err != nil {
			return err
		}
	}

	return nil
}

func SerializeFiltered(g *rdf.Graph, w io.Writer, uriPrefixFilter string) error {
	var err error

	for _, triple := range g.Triples() {
		s := triple.Subject.String()
		if strings.HasPrefix(s, uriPrefixFilter) {
			continue
		}

		p := triple.Predicate.String()
		o := triple.Object.String()

		if strings.HasPrefix(o, uriPrefixFilter) {
			continue
		}

		_, err = fmt.Fprintf(w, "%s %s %s .\n", s, p, o)
		if err != nil {
			return err
		}
	}

	return nil
}
