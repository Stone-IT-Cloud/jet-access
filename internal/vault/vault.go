package vault

import (
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
	// "log/slog" // Only if logging errors from config loading
)

// Vault prints "Hello, World!" to the standard output.
// This function can be kept or removed if not used.
func Vault() {
	fmt.Println("Hello, World!")
}

// NewVaultClient creates and configures a new Vault client using settings
// from the default environment in the application's configuration file.
func NewVaultClient() (*api.Client, error) {
	configPath, err := GetConfigPath() // Assumes GetConfigPath is defined in configs.go
	if err != nil {
		return nil, fmt.Errorf("failed to get Vault config path: %w", err)
	}

	appConfig, err := GetConfig(configPath) // Assumes GetConfig is defined in configs.go
	if err != nil {
		return nil, fmt.Errorf("failed to load Vault configuration: %w", err)
	}

	// Use the Default environment configuration. appConfig.Default is EnvironmentConfig (a struct)
	vaultConfig := appConfig.Default

	apiConfig := api.DefaultConfig() // Start with default config
	apiConfig.Address = vaultConfig.Address // Directly access fields

	// RetryConfig is a struct, directly access its fields
	apiConfig.MaxRetries = vaultConfig.Retry.MaxAttempts 
	apiConfig.Timeout = time.Duration(vaultConfig.Timeout) * time.Second

	// TLSConfig is a struct, directly access its fields
	// No nil check needed for vaultConfig.TLS itself
	tlsConfig := &api.TLSConfig{
		CACert:     vaultConfig.TLS.CACert,
		ClientCert: vaultConfig.TLS.ClientCert,
		ClientKey:  vaultConfig.TLS.ClientKey,
		Insecure:   !vaultConfig.TLS.Verify, // If Verify is true, Insecure is false
	}
	if err := apiConfig.ConfigureTLS(tlsConfig); err != nil {
		// This error is from apiConfig.ConfigureTLS, not a nil check on vaultConfig.TLS
		return nil, fmt.Errorf("failed to configure TLS for Vault client: %w", err)
	}

	client, err := api.NewClient(apiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	client.SetToken(vaultConfig.Token) // Directly access token

	return client, nil
}
