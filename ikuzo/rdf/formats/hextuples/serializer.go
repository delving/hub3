package hextuples

import (
	"fmt"
	"io"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/xitongsys/parquet-go/writer"
)

func SerializeParquet(g *rdf.Graph, pw *writer.ParquetWriter) error {
	for _, t := range g.Triples() {
		ht := FromTriple(t, g.GraphName)
		if err := pw.Write(ht); err != nil {
			return err
		}
	}

	return nil
}

func Serialize(g *rdf.Graph, w io.Writer) error {
	for _, t := range g.Triples() {
		ht := FromTriple(t, g.GraphName)

		b, err := ht.MarshalJSON()
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(w, string(b))
		if err != nil {
			return err
		}
	}

	return nil
}
