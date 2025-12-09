// Package schema handles loading and access to the EIB configuration schema.
//
// It provides functionality to load the embedded JSON schema and retrieve it
// for validation purposes.
package schema

import (
	_ "embed"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema.json
var schemaJSON []byte

// LoadSchema loads the EIB configuration schema from the embedded file.
//
// It parses the embedded JSON schema and returns a compiled schema object
// ready for validation.
//
// Returns:
//   - *gojsonschema.Schema: The compiled JSON schema.
//   - error: An error if the schema cannot be parsed.
func LoadSchema() (*gojsonschema.Schema, error) {
	loader := gojsonschema.NewBytesLoader(schemaJSON)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema: %w", err)
	}

	return schema, nil
}

// GetRawSchema returns the raw JSON bytes of the schema.
//
// This is useful for clients that need to inspect the schema directly
// or embed it in other tools.
//
// Returns:
//   - []byte: The raw JSON schema bytes.
func GetRawSchema() []byte {
	return schemaJSON
}
