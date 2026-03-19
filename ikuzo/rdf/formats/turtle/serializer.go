package turtle

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/delving/hub3/ikuzo/rdf"
)

// Option configures the turtle serializer.
type Option func(*serializer)

// WithPrefixes sets custom prefix-to-namespace mappings for compact output.
func WithPrefixes(prefixes map[string]string) Option {
	return func(s *serializer) {
		for k, v := range prefixes {
			s.prefixes[k] = v
		}
	}
}

// WithBaseURI sets the @base directive for the turtle output.
func WithBaseURI(base string) Option {
	return func(s *serializer) {
		s.baseURI = base
	}
}

type serializer struct {
	prefixes   map[string]string // prefix -> namespace URI
	baseURI    string
	nsToPrefix map[string]string // namespace URI -> prefix
}

// Serialize writes an rdf.Graph as Turtle to the writer.
// The graph should have UseResource and UseIndex set to true for optimal output.
func Serialize(g *rdf.Graph, w io.Writer, opts ...Option) error {
	s := &serializer{
		prefixes:   make(map[string]string),
		nsToPrefix: make(map[string]string),
	}

	for _, opt := range opts {
		opt(s)
	}

	// Build reverse lookup
	for prefix, ns := range s.prefixes {
		s.nsToPrefix[ns] = prefix
	}

	resources := g.ResourcesOrdered()
	if len(resources) == 0 {
		return nil
	}

	// Write prefix declarations (sorted)
	if err := s.writePrefixes(w); err != nil {
		return err
	}

	// Write resources
	for i, rsc := range resources {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if err := s.writeResource(w, rsc); err != nil {
			return err
		}
	}

	return nil
}

func (s *serializer) writePrefixes(w io.Writer) error {
	if len(s.prefixes) == 0 {
		return nil
	}

	// Sort prefixes alphabetically
	keys := make([]string, 0, len(s.prefixes))
	for k := range s.prefixes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, prefix := range keys {
		ns := s.prefixes[prefix]
		if _, err := fmt.Fprintf(w, "@prefix %s: <%s> .\n", prefix, ns); err != nil {
			return err
		}
	}

	// Blank line after prefixes
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	return nil
}

func (s *serializer) writeResource(w io.Writer, rsc *rdf.Resource) error {
	subject := s.formatTerm(rsc.Subject())
	predicates := rsc.SortedPredicates()

	if len(predicates) == 0 {
		return nil
	}

	// Write first predicate on same line as subject
	for i, pred := range predicates {
		predStr := s.formatPredicate(pred.IRI())
		objects := pred.Objects()

		if i == 0 {
			if _, err := fmt.Fprintf(w, "%s %s ", subject, predStr); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(w, "    %s ", predStr); err != nil {
				return err
			}
		}

		// Write objects (comma-separated for same predicate)
		for j, obj := range objects {
			if j > 0 {
				if _, err := fmt.Fprint(w, ",\n        "); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprint(w, s.formatTerm(obj)); err != nil {
				return err
			}
		}

		// Semicolon or period
		if i < len(predicates)-1 {
			if _, err := fmt.Fprint(w, " ;\n"); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprint(w, " .\n"); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *serializer) formatPredicate(iri rdf.IRI) string {
	// Special case: rdf:type -> "a"
	if iri.RawValue() == rdf.RDFType {
		return "a"
	}
	return s.formatIRI(iri)
}

func (s *serializer) formatTerm(term rdf.Term) string {
	switch v := term.(type) {
	case rdf.IRI:
		return s.formatIRI(v)
	case rdf.Literal:
		return s.formatLiteral(v)
	case rdf.BlankNode:
		return v.String()
	default:
		return fmt.Sprintf("<%s>", term.RawValue())
	}
}

func (s *serializer) formatIRI(iri rdf.IRI) string {
	raw := iri.RawValue()

	// Try prefix compression
	for ns, prefix := range s.nsToPrefix {
		if strings.HasPrefix(raw, ns) {
			localName := raw[len(ns):]
			if localName != "" && !strings.ContainsAny(localName, "/#") {
				return prefix + ":" + localName
			}
		}
	}

	return "<" + raw + ">"
}

func (s *serializer) formatLiteral(lit rdf.Literal) string {
	value := lit.RawValue()
	escaped := escapeTurtleString(value)

	if lang := lit.Lang(); lang != "" {
		return fmt.Sprintf(`"%s"@%s`, escaped, lang)
	}

	dt := lit.DataType
	if dt.RawValue() != "" && dt.RawValue() != "http://www.w3.org/2001/XMLSchema#string" {
		dtStr := s.formatIRI(dt)
		return fmt.Sprintf(`"%s"^^%s`, escaped, dtStr)
	}

	return fmt.Sprintf(`"%s"`, escaped)
}

func escapeTurtleString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}
