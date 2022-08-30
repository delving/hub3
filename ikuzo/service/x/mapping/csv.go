package mapping

// Definition is the root struct for all mappings
type Definition struct {
	Resource string `json:"resource" csv:"resource"`
	Label    string `json:"label" csv:"label"`
}
