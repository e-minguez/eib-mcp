// Package tool contains the logic for EIB configuration tools.
//
// It provides functionality to validate and generate Edge Image Builder configuration
// files, including handling password encryption.
package tool

import (
	"fmt"
	"strings"

	"github.com/e-minguez/eib-mcp/schema"
	"github.com/xeipuuv/gojsonschema"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

// GenerateConfig validates the input map against the EIB schema and returns the YAML representation.
//
// It performs the following steps:
// 1. Encrypts any plaintext passwords found in the input.
// 2. Validates the input against the EIB JSON schema.
// 3. Marshals the valid input into a YAML string.
//
// Parameters:
//   - input: A map representing the configuration data.
//
// Returns:
//   - string: The generated YAML configuration.
//   - error: An error if validation or generation fails.
func GenerateConfig(input map[string]interface{}) (string, error) {
	// 1. Process Passwords (encrypt plaintext 'password' fields)
	// We do this BEFORE validation so that 'password' is replaced by 'encryptedPassword',
	// which complies with the strict schema.
	if err := processPasswords(input); err != nil {
		return "", fmt.Errorf("failed to encrypt passwords: %w", err)
	}

	// 2. Load Schema
	s, err := schema.LoadSchema()
	if err != nil {
		return "", fmt.Errorf("failed to load schema: %w", err)
	}

	// 3. Validate Input
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

	// 4. Convert to YAML
	yamlBytes, err := yaml.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	return string(yamlBytes), nil
}

// processPasswords iterates through the configuration and encrypts plaintext passwords.
//
// It looks for "password" fields in the "operatingSystem.users" list and replaces them
// with "encryptedPassword" fields containing the bcrypt hash. It also ensures that
// existing "encryptedPassword" fields are not double-encrypted if they appear to be hashes.
//
// Parameters:
//   - input: The configuration map to process.
//
// Returns:
//   - error: An error if encryption fails.
func processPasswords(input map[string]interface{}) error {
	osVal, ok := input["operatingSystem"]
	if !ok {
		return nil
	}
	osMap, ok := osVal.(map[string]interface{})
	if !ok {
		return nil
	}

	usersVal, ok := osMap["users"]
	if !ok {
		return nil
	}
	usersSlice, ok := usersVal.([]interface{})
	if !ok {
		return nil
	}

	for _, u := range usersSlice {
		userMap, ok := u.(map[string]interface{})
		if !ok {
			continue
		}
		// Check for 'password' field (virtual field for plaintext)
		if pwd, ok := userMap["password"].(string); ok && pwd != "" {
			hash, err := encryptPassword(pwd)
			if err != nil {
				return fmt.Errorf("encryption failed: %w", err)
			}
			userMap["encryptedPassword"] = hash
			delete(userMap, "password")
		} else if encPwd, ok := userMap["encryptedPassword"].(string); ok && encPwd != "" {
			// Check if 'encryptedPassword' is actually plaintext (doesn't start with $)
			if !strings.HasPrefix(encPwd, "$") {
				hash, err := encryptPassword(encPwd)
				if err != nil {
					return fmt.Errorf("encryption failed: %w", err)
				}
				userMap["encryptedPassword"] = hash
			}
		}
	}
	return nil
}

// encryptPassword generates a bcrypt hash for the given password.
//
// It uses a default cost of 10.
//
// Parameters:
//   - password: The plaintext password to encrypt.
//
// Returns:
//   - string: The bcrypt hash of the password.
//   - error: An error if hashing fails.
func encryptPassword(password string) (string, error) {
	// Use bcrypt (native Go) instead of shelling out to openssl.
	// Cost 10 is a reasonable default.
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
