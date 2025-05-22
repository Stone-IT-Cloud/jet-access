package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
)

// Helper function to create a new mock Vault client and server
func newTestVaultClient(t *testing.T, handler http.HandlerFunc) (*api.Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)

	conf := api.DefaultConfig()
	conf.Address = server.URL
	conf.MaxRetries = 0 // No retries for tests for predictability
	conf.Timeout = 1 * time.Second // Short timeout for tests

	client, err := api.NewClient(conf)
	if err != nil {
		server.Close()
		t.Fatalf("Failed to create Vault client for test: %v", err)
	}
	client.SetToken("testtoken") // Dummy token for tests

	return client, server
}

func TestListPath_Success(t *testing.T) {
	expectedKeys := []string{"key1", "key2", "subdir/"}
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Header.Get("X-Vault-Request") == "true" { // Vault CLI uses GET with X-Vault-Request, SDK uses LIST
			// Fallback for older vault versions or different client behaviors if needed
		}
		if !( (r.Method == http.MethodGet && r.Header.Get("X-Vault-Request") == "true") || r.Method == "LIST" ) {
			t.Errorf("Expected 'LIST' or 'GET with X-Vault-Request' request, got '%s'", r.Method)
			http.Error(w, "bad method", http.StatusBadRequest)
			return
		}
		if !strings.HasSuffix(r.URL.Path, "/test/list/path") {
			t.Errorf("Expected request to '/test/list/path', got '%s'", r.URL.Path)
			http.Error(w, "bad path", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"keys": expectedKeys,
			},
		})
	}

	client, server := newTestVaultClient(t, handler)
	defer server.Close()

	keys, err := ListPath(client, "test/list/path")
	if err != nil {
		t.Fatalf("ListPath failed: %v", err)
	}
	if !reflect.DeepEqual(keys, expectedKeys) {
		t.Errorf("Expected keys %v, got %v", expectedKeys, keys)
	}
}

func TestListPath_Empty(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"keys": []string{}, // Empty list of keys
			},
		})
	}
	client, server := newTestVaultClient(t, handler)
	defer server.Close()

	keys, err := ListPath(client, "test/empty/path")
	if err != nil {
		t.Fatalf("ListPath failed for empty path: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, got %d", len(keys))
	}
}

func TestListPath_NotFoundOrError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "path not found", http.StatusNotFound)
	}
	client, server := newTestVaultClient(t, handler)
	defer server.Close()

	_, err := ListPath(client, "test/nonexistent/path")
	if err == nil {
		t.Fatal("ListPath should have failed but didn't")
	}
	// Check if the error message contains the path and the original error
	if !strings.Contains(err.Error(), "test/nonexistent/path") {
		t.Errorf("Error message does not contain the path: %s", err.Error())
	}
}


func TestListPath_MalformedResponse(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Malformed: "keys" is a string instead of a slice
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"keys": "not_a_slice",
			},
		})
	}
	client, server := newTestVaultClient(t, handler)
	defer server.Close()

	_, err := ListPath(client, "test/malformed/path")
	if err == nil {
		t.Fatal("ListPath should have failed for malformed response but didn't")
	}
    if !strings.Contains(err.Error(), "failed to parse keys from path") {
        t.Errorf("Unexpected error message for malformed response: %s", err.Error())
    }
}


func TestReadPath_Success_Full_KV2(t *testing.T) {
	expectedSecret := VaultSecret{
		Hostname:      "testhost",
		IP:            "1.2.3.4",
		Port:          "22",
		Username:      "testuser",
		Password:      "testpass",
		Key:           "testkeypath",
		KeyPassphrase: "testkeypass",
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected 'GET' request, got '%s'", r.Method)
			http.Error(w, "bad method", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{ // KV v2 nesting
				"data": map[string]interface{}{
					"hostname":       expectedSecret.Hostname,
					"ip":             expectedSecret.IP,
					"port":           expectedSecret.Port,
					"username":       expectedSecret.Username,
					"password":       expectedSecret.Password,
					"key":            expectedSecret.Key,
					"key_passphrase": expectedSecret.KeyPassphrase,
				},
				"metadata": map[string]interface{}{"version": 1},
			},
		})
	}
	client, server := newTestVaultClient(t, handler)
	defer server.Close()

	secret, err := ReadPath(client, "test/secret/full_kv2")
	if err != nil {
		t.Fatalf("ReadPath failed: %v", err)
	}
	if !reflect.DeepEqual(*secret, expectedSecret) {
		t.Errorf("Expected secret %+v, got %+v", expectedSecret, *secret)
	}
}

func TestReadPath_Success_KV1_Fallback(t *testing.T) {
    expectedSecret := VaultSecret{
        Hostname: "kv1host",
        IP:       "4.3.2.1",
        Port:     "2222",
        Username: "kv1user",
        Password: "kv1password",
    }
    handler := func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        // No "data" nesting for KV v1 style
        json.NewEncoder(w).Encode(map[string]interface{}{
            "data": map[string]interface{}{ // This outer "data" is part of the Vault response structure itself
                "hostname": expectedSecret.Hostname,
                "ip":       expectedSecret.IP,
                "port":     expectedSecret.Port,
                "username": expectedSecret.Username,
                "password": expectedSecret.Password,
            },
        })
    }
    client, server := newTestVaultClient(t, handler)
    defer server.Close()

    secret, err := ReadPath(client, "test/secret/kv1")
    if err != nil {
        t.Fatalf("ReadPath for KV1 failed: %v", err)
    }
    if !reflect.DeepEqual(*secret, expectedSecret) {
        t.Errorf("Expected KV1 secret %+v, got %+v", expectedSecret, *secret)
    }
}


func TestReadPath_Success_PasswordOnly(t *testing.T) {
	expectedSecret := VaultSecret{
		Hostname: "testhost_pw_only",
		IP:       "1.2.3.5",
		Port:     "23",
		Username: "testuser_pw_only",
		Password: "testpass_only",
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"data": map[string]interface{}{
					"hostname": expectedSecret.Hostname,
					"ip":       expectedSecret.IP,
					"port":     expectedSecret.Port,
					"username": expectedSecret.Username,
					"password": expectedSecret.Password,
				},
			},
		})
	}
	client, server := newTestVaultClient(t, handler)
	defer server.Close()

	secret, err := ReadPath(client, "test/secret/pw_only")
	if err != nil {
		t.Fatalf("ReadPath failed: %v", err)
	}
	if !reflect.DeepEqual(*secret, expectedSecret) {
		t.Errorf("Expected secret %+v, got %+v", expectedSecret, *secret)
	}
}

func TestReadPath_Success_KeyOnly(t *testing.T) {
	expectedSecret := VaultSecret{
		Hostname:      "testhost_key_only",
		IP:            "1.2.3.6",
		Port:          "24",
		Username:      "testuser_key_only",
		Key:           "testkeypath_only",
		KeyPassphrase: "testkeypass_optional",
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"data": map[string]interface{}{
					"hostname":       expectedSecret.Hostname,
					"ip":             expectedSecret.IP,
					"port":           expectedSecret.Port,
					"username":       expectedSecret.Username,
					"key":            expectedSecret.Key,
					"key_passphrase": expectedSecret.KeyPassphrase,
				},
			},
		})
	}
	client, server := newTestVaultClient(t, handler)
	defer server.Close()

	secret, err := ReadPath(client, "test/secret/key_only")
	if err != nil {
		t.Fatalf("ReadPath failed: %v", err)
	}
	if !reflect.DeepEqual(*secret, expectedSecret) {
		t.Errorf("Expected secret %+v, got %+v", expectedSecret, *secret)
	}
}

func TestReadPath_MissingCredentials(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"data": map[string]interface{}{ // Missing password and key
					"hostname": "testhost_no_creds",
					"ip":       "1.2.3.7",
					"port":     "25",
					"username": "testuser_no_creds",
				},
			},
		})
	}
	client, server := newTestVaultClient(t, handler)
	defer server.Close()

	_, err := ReadPath(client, "test/secret/no_creds")
	if err == nil {
		t.Fatal("ReadPath should have failed due to missing credentials but didn't")
	}
	if !strings.Contains(err.Error(), "either 'password' or 'key' must be provided") {
		t.Errorf("Unexpected error message for missing credentials: %s", err.Error())
	}
}

func TestReadPath_NotFoundOrError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "secret not found", http.StatusNotFound)
	}
	client, server := newTestVaultClient(t, handler)
	defer server.Close()

	_, err := ReadPath(client, "test/secret/nonexistent")
	if err == nil {
		t.Fatal("ReadPath should have failed but didn't")
	}
    if !strings.Contains(err.Error(), "failed to read path") {
        t.Errorf("Unexpected error message for not found: %s", err.Error())
    }
}

func TestReadPath_MalformedSecretData(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"data": "not_a_map", // Malformed: data is a string instead of a map
			},
		})
	}
	client, server := newTestVaultClient(t, handler)
	defer server.Close()

	_, err := ReadPath(client, "test/secret/malformed_data")
	if err == nil {
		t.Fatal("ReadPath should have failed for malformed secret data but didn't")
	}
    if !strings.Contains(err.Error(), "is not in expected KV v2 format") {
         t.Errorf("Unexpected error message for malformed data: %s", err.Error())
    }
}

func TestReadPath_NoSecretData(t *testing.T) {
    handler := func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        // Vault response for a path that exists but has no data (e.g., a "folder" in KV v2)
        // or a truly empty secret.
        json.NewEncoder(w).Encode(map[string]interface{}{
            // "data" key is completely missing or null
        })
    }
    client, server := newTestVaultClient(t, handler)
    defer server.Close()

    _, err := ReadPath(client, "test/secret/no_data_field")
    if err == nil {
        t.Fatal("ReadPath should have failed when 'data' field is missing/null but didn't")
    }
    if !strings.Contains(err.Error(), "no secret found at path") {
        t.Errorf("Unexpected error message for missing 'data' field: %s", err.Error())
    }
}

// TestNewVaultClient_Basic tests the NewVaultClient function.
// It focuses on ensuring the client can be created if a valid configuration is present.
func TestNewVaultClient_Basic(t *testing.T) {
	// Create a temporary config directory relative to the current working dir
	// GetConfigPath() looks for "config/jet-ssh-config.yaml"
	tempConfigDirPath := "./config" // Relative to where 'go test' is run for this package
	tempConfigFilePath := filepath.Join(tempConfigDirPath, "jet-ssh-config.yaml")

	// Ensure the directory exists
	if err := os.MkdirAll(tempConfigDirPath, 0755); err != nil {
		t.Fatalf("Failed to create temporary config directory %s: %v", tempConfigDirPath, err)
	}
	// Cleanup: remove the entire temp config_test directory
	t.Cleanup(func() {
		if err := os.RemoveAll(tempConfigDirPath); err != nil {
			// Log the error but don't fail the test at cleanup, as the main test logic is more important.
			t.Logf("Warning: failed to remove temporary config directory %s: %v", tempConfigDirPath, err)
		}
	})
	
	dummyConfigContent := `
default:
  address: http://127.0.0.1:8200 # Standard Vault dev address
  token: s.unittestActualToken
  timeout: 5
  retry:
    max_attempts: 1
    initial_interval: 1
    max_interval: 1
  tls:
    verify: false # Typically false for local dev/test
`
	if err := os.WriteFile(tempConfigFilePath, []byte(dummyConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write temporary config file %s: %v", tempConfigFilePath, err)
	}

	client, err := NewVaultClient()

	// We expect NewVaultClient to successfully load the config.
	// It might still return an error if it tries to connect to Vault and Vault is not running,
	// but it should not be a config loading error.
	if err != nil {
		if strings.Contains(err.Error(), "failed to load Vault configuration") || 
		   strings.Contains(err.Error(), "failed to get Vault config path") ||
		   strings.Contains(err.Error(), "no such file or directory") { // Error related to finding/reading config
			t.Fatalf("NewVaultClient() failed due to config loading issues, but should have used the temp config: %v", err)
		}
		// If the error is about connecting to Vault (e.g., connection refused), that's acceptable for this test.
		t.Logf("NewVaultClient() returned an error as expected (Vault not running or other runtime issue): %v", err)
	} else {
		if client == nil {
			t.Fatal("NewVaultClient() returned nil client without error")
		}
		// Verify client configuration if no error occurred
		if client.Address() != "http://127.0.0.1:8200" {
			t.Errorf("Expected client address http://127.0.0.1:8200, got %s", client.Address())
		}
		if client.Token() != "s.unittestActualToken" {
			t.Errorf("Expected client token s.unittestActualToken, got %s", client.Token())
		}
		t.Log("NewVaultClient() created and configured client successfully.")
	}
}
