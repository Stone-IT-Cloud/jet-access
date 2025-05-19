package vault

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// TLSConfig contains TLS/SSL configuration parameters for connecting to Vault
type TLSConfig struct {
	// Verify enables TLS certificate verification
	Verify bool `yaml:"verify"`
	// CACert is the path to the CA certificate file used for TLS verification
	CACert string `yaml:"ca_cert"`
	// ClientCert is the path to the client certificate file for mutual TLS authentication
	ClientCert string `yaml:"client_cert"`
	// ClientKey is the path to the client private key file for mutual TLS authentication
	ClientKey string `yaml:"client_key"`
}

// RetryConfig contains configuration parameters for retry behavior when connecting to Vault
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts that will be made
	MaxAttempts int `yaml:"max_attempts"`
	// InitialInterval is the initial delay in seconds between retry attempts
	InitialInterval int `yaml:"initial_interval"`
	// MaxInterval is the maximum delay in seconds between retry attempts
	MaxInterval int `yaml:"max_interval"`
}

// EnvironmentConfig contains configuration parameters for a specific Vault environment
type EnvironmentConfig struct {
	// Address is the URL of the Vault server (must start with http:// or https://)
	Address string `yaml:"address"`
	// Token is the authentication token used to access Vault
	Token string `yaml:"token"`
	// TLS contains the TLS/SSL configuration for connecting to Vault
	TLS TLSConfig `yaml:"tls"`
	// Timeout specifies the request timeout in seconds
	Timeout int `yaml:"timeout"`
	// Retry contains the configuration for retry behavior on failed requests
	Retry RetryConfig `yaml:"retry"`
}

// JetSSHConfig represents the top-level configuration structure for the JetSSH application.
// It contains a default environment configuration and a map of named environment-specific configurations.
type JetSSHConfig struct {
	// Default contains the base configuration that all environments inherit from
	Default EnvironmentConfig `yaml:"default"`
	// Environments is a map of named environment configurations that can override the default settings
	Environments map[string]EnvironmentConfig `yaml:"environments,omitempty"`
}

func (c *TLSConfig) validate() error {
	if c.Verify {
		if c.CACert == "" {
			return fmt.Errorf("TLS verification is enabled but no CA certificate is provided")
		}
		if c.ClientCert != "" && c.ClientKey == "" {
			return fmt.Errorf("client certificate is provided but no client key")
		}
		if c.ClientKey != "" && c.ClientCert == "" {
			return fmt.Errorf("client key is provided but no client certificate")
		}
	}
	return nil
}

func (c *RetryConfig) validate() error {
	if c.MaxAttempts < 1 {
		return fmt.Errorf("max_attempts must be greater than 0")
	}
	if c.InitialInterval < 1 {
		return fmt.Errorf("initial_interval must be greater than 0")
	}
	if c.MaxInterval < c.InitialInterval {
		return fmt.Errorf("max_interval must be greater than or equal to initial_interval")
	}
	return nil
}

func (c *EnvironmentConfig) validate() error {
	if c.Address == "" {
		return fmt.Errorf("address is required")
	}
	if !strings.HasPrefix(c.Address, "http://") && !strings.HasPrefix(c.Address, "https://") {
		return fmt.Errorf("address must start with http:// or https://")
	}
	if c.Token == "" {
		return fmt.Errorf("token is required")
	}
	if c.Timeout < 1 {
		return fmt.Errorf("timeout must be greater than 0")
	}

	if err := c.TLS.validate(); err != nil {
		return fmt.Errorf("TLS validation failed: %w", err)
	}
	if err := c.Retry.validate(); err != nil {
		return fmt.Errorf("retry validation failed: %w", err)
	}
	return nil
}

func (c *JetSSHConfig) validate() error {
	if err := c.Default.validate(); err != nil {
		return fmt.Errorf("default configuration validation failed: %w", err)
	}

	for env, config := range c.Environments {
		if err := config.validate(); err != nil {
			return fmt.Errorf("environment '%s' validation failed: %w", env, err)
		}
	}
	return nil
}

// GetConfigPath returns the path to the configuration file for the application.
// It first checks for a configuration file in the development environment at
// "config/jet-ssh-config.yaml". If not found, it falls back to creating and using
// a configuration file in the user's configuration directory.
//
// The function follows this search order:
//  1. Looks for config/jet-ssh-config.yaml in the current directory
//  2. Creates and uses $XDG_CONFIG_HOME/jet-ssh/jet-ssh-config.yaml (Unix)
//     or %AppData%/jet-ssh/jet-ssh-config.yaml (Windows)
//
// Returns:
//   - string: The path to the configuration file
//   - error: An error if unable to determine or create the config directory
func GetConfigPath() (string, error) {
	// Check the config directory. This will be used on developer's development environment
	devConfigDir := "config/jet-ssh-config.yaml"
	_, errDevConfig := os.Stat(devConfigDir)
	if errDevConfig == nil {
		return devConfigDir, nil
	}

	slog.Debug("Developer config directory not found, using user config directory", "error", errDevConfig)

	configDir, errUsrConfig := os.UserConfigDir()
	if errUsrConfig != nil {
		return "", fmt.Errorf("error getting user config directory: %s", errUsrConfig)
	}

	appConfigDir := filepath.Join(configDir, "jet-ssh")
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", fmt.Errorf("error creating config directory: %s", err)
	}
	return filepath.Join(appConfigDir, "jet-ssh-config.yaml"), nil
}

// GetConfig reads and parses the configuration file at the given path.
// It unmarshals the YAML content into a JetSSHConfig struct and validates
// the configuration.
//
// Parameters:
//   - configPath: The path to the configuration file to read
//
// Returns:
//   - *JetSSHConfig: The parsed and validated configuration
//   - error: An error if reading, parsing or validation fails
func GetConfig(configPath string) (*JetSSHConfig, error) {
	config := &JetSSHConfig{}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}
