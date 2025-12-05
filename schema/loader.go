package schema

import (
	_ "embed"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema.json
var schemaJSON []byte

// LoadSchema loads the EIB configuration schema from the embedded file.
func LoadSchema() (*gojsonschema.Schema, error) {
	loader := gojsonschema.NewBytesLoader(schemaJSON)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema: %w", err)
	}

	return schema, nil
}

// GetRawSchema returns the raw JSON bytes of the schema.
func GetRawSchema() []byte {
	return schemaJSON
}
