package tool

import (
	"fmt"

	"github.com/e-minguez/eib-mcp/schema"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// GenerateConfig validates the input map against the EIB schema and returns the YAML representation.
func GenerateConfig(input map[string]interface{}) (string, error) {
	// 1. Load Schema
	s, err := schema.LoadSchema()
	if err != nil {
		return "", fmt.Errorf("failed to load schema: %w", err)
	}

	// 2. Validate Input
	inputLoader := gojsonschema.NewGoLoader(input)
	result, err := s.Validate(inputLoader)
	if err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	if !result.Valid() {
		var errMsgs string
		for _, desc := range result.Errors() {
			errMsgs += fmt.Sprintf("- %s\n", desc)
		}
		return "", fmt.Errorf("configuration is invalid:\n%s", errMsgs)
	}

	// 3. Convert to YAML
	yamlBytes, err := yaml.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	return string(yamlBytes), nil
}
