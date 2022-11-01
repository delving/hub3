package schema

type Schema struct {
	Name            string
	Description     string
	NamespacePrefix string
	NamespaceURI    string
	Classes         []Class
	Properties      []Property
}
