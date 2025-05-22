package vault

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/vault/api"
)

// VaultSecret holds the data read from a Vault path.
type VaultSecret struct {
	Hostname      string `json:"hostname"`
	IP            string `json:"ip"`
	Port          string `json:"port"`
	Username      string `json:"username"`
	Password      string `json:"password,omitempty"`
	Key           string `json:"key,omitempty"`
	KeyPassphrase string `json:"key_passphrase,omitempty"`
}

// ListPath lists secrets at the given path in Vault.
// It takes an initialized Vault client and the path string as input.
// It returns a slice of strings (keys) and an error if any occurs.
func ListPath(client *api.Client, path string) ([]string, error) {
	if client == nil {
		return nil, fmt.Errorf("Vault client is not initialized")
	}

	logicalClient := client.Logical()
	secret, err := logicalClient.List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list path '%s': %w", path, err)
	}

	if secret == nil || secret.Data == nil || secret.Data["keys"] == nil {
		// Path might be valid but contain no keys or is not a listable path
		return []string{}, nil
	}

	keysInterface, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to parse keys from path '%s': 'keys' field is not a []interface{}", path)
	}

	var keys []string
	for _, keyInterface := range keysInterface {
		key, ok := keyInterface.(string)
		if !ok {
			return nil, fmt.Errorf("failed to parse key: item in 'keys' is not a string")
		}
		keys = append(keys, key)
	}

	return keys, nil
}

// ReadPath reads a secret from the given path in Vault and parses it into a VaultSecret struct.
// It takes an initialized Vault client and the path string as input.
// It returns a pointer to a VaultSecret struct and an error if any occurs.
// The secret is expected to be a KV v2 secret, where actual data is nested under secret.Data["data"].
func ReadPath(client *api.Client, path string) (*VaultSecret, error) {
	if client == nil {
		return nil, fmt.Errorf("Vault client is not initialized")
	}

	logicalClient := client.Logical()
	secret, err := logicalClient.Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read path '%s': %w", path, err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no secret found at path '%s'", path)
	}

	// For KV v2 secrets, the actual data is nested under "data"
	secretData, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		// Attempt to parse as KV v1 if "data" map is not present or not in expected format
		// KV v1 stores data directly in secret.Data
		secretData = secret.Data
		// Check if this direct data can be reasonably mapped to our struct.
		// This is a simple check; more complex validation might be needed based on actual KV v1 structure.
		if _, PathOk := secretData["hostname"].(string); !PathOk {
			// If hostname is not directly available, it's unlikely to be a compatible KV v1 secret or any known format.
			return nil, fmt.Errorf("secret data at path '%s' is not in expected KV v2 format (missing 'data' map) or compatible KV v1 format", path)
		}
	}
	
	parsedSecret := &VaultSecret{}

	// Use reflection to map data to struct fields
	structVal := reflect.ValueOf(parsedSecret).Elem()
	structType := structVal.Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		jsonTag := field.Tag.Get("json") // Assumes json tags match map keys
		// Strip ",omitempty" if present in the tag
		if tagName, _, found := strings.Cut(jsonTag, ","); found {
			jsonTag = tagName
		}


		if val, exists := secretData[jsonTag]; exists {
			if valStr, IsString := val.(string); IsString {
				if structVal.Field(i).CanSet() {
					structVal.Field(i).SetString(valStr)
				}
			} else {
				// Handle cases where field is present but not a string (e.g. if data types are mixed in Vault)
				// For now, we skip non-string fields if they don't match our struct's string fields.
				// More sophisticated type conversion could be added here if needed.
				// Or return an error if strict type matching is required.
			}
		}
	}
	
	// Validate optional fields logic:
	// password is optional if key has been set, required otherwise
	// key is optional if password has been set, required otherwise
	if parsedSecret.Password == "" && parsedSecret.Key == "" {
		return nil, fmt.Errorf("either 'password' or 'key' must be provided in secret at path '%s'", path)
	}

	return parsedSecret, nil
}
