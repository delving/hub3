package components

import _ "embed"

//go:embed api.swagger.json
var SwaggerJSON []byte

//go:embed api.swagger.yaml
var SwaggerYAML []byte
