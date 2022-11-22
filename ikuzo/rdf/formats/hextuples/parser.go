package hextuples

import (
	"bufio"
	"io"

	"github.com/delving/hub3/ikuzo/rdf"
)

func Parse(r io.Reader, g *rdf.Graph) (*rdf.Graph, error) {
	if g == nil {
		g = rdf.NewGraph()
	}

	// parse lines
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		ht, err := New(scanner.Bytes())
		if err != nil {
			return nil, err
		}

		t, err := ht.AsTriple()
		if err != nil {
			return nil, err
		}

		g.Add(t)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return g, nil
}
