package rdf

// GetResourceID is a legacy function to return a string from of the resource
// identifier.
func GetResourceID(t Term) string {
	return t.RawValue()
}
