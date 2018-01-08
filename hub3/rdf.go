package hub3

import (
	"fmt"
	"net/url"
	"regexp"
)

// ntriples2nquads is a utility function to convert turtle strings to nquads
// This is used to store RDF data in bulk via the RDFGraphStore protocol
func Ntriples2Nquads(i string, graphUri *url.URL) string {
	graphName := fmt.Sprintf(" <%s> .", graphUri.String())
	re := regexp.MustCompile(" .\n")
	return re.ReplaceAllString(i, graphName)
}
