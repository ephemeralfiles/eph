// Package config provides configuration management for the ephemeralfiles CLI.
// It supports loading configuration from both YAML files and environment variables,
// with environment variables taking precedence over file-based configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	// ConfigurationDirPerm is the permission (0700) for the configuration directory.
	ConfigurationDirPerm  = 0700
	// ConfigurationFilePerm is the permission (0600) for configuration files.
	ConfigurationFilePerm = 0600
)

var (
	// ErrConfigurationNotFound is returned when no valid configuration is found.
	ErrConfigurationNotFound = errors.New("configuration not found")
	// ErrInvalidToken is returned when the provided token is invalid.
	ErrInvalidToken          = errors.New("token is invalid")
	// ErrInvalidEndpoint is returned when the provided endpoint is invalid.
	ErrInvalidEndpoint       = errors.New("endpoint is invalid")
)

// Config is the configuration for the application.
type Config struct {
	Token               string `yaml:"token"`
	Endpoint            string `yaml:"endpoint"`
	DefaultOrganization string `yaml:"default_organization,omitempty"`
	homedir             string
}

// NewConfig creates a new configuration for the application.
func NewConfig() *Config {
	cfg := &Config{
		Token:               "",
		Endpoint:            "",
		DefaultOrganization: "",
		homedir:             "",
	}
	cfg.initHomedir()

	return cfg
}


// SetHomedir sets the homedir variable
// It is used for testing purposes.
func (c *Config) SetHomedir(homedir string) {
	c.homedir = homedir
}

// LoadConfigFromFile loads the configuration from a file.
func (c *Config) LoadConfigFromFile(filename string) error {
	// #nosec G304 -- filename is provided by user for config file reading
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error loading configuration file: %w", err)
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return fmt.Errorf("error parsing YAML file: %w", err)
	}

	return nil
}

// LoadConfigFromEnvVar loads the configuration from the environment variables.
func (c *Config) LoadConfigFromEnvVar() {
	c.Token = os.Getenv("EPHEMERALFILES_TOKEN")
	c.Endpoint = os.Getenv("EPHEMERALFILES_ENDPOINT")
}

// IsConfigValid checks if the configuration is valid.
func (c *Config) IsConfigValid() bool {
	if c.Token == "" || c.Endpoint == "" {
		return false
	}
	return true
}

// LoadConfiguration loads the configuration from the environment variables first.
// If the configuration is not valid, it tries to load the configuration from a file.
func (c *Config) LoadConfiguration(cfgFilePath string) error {
	c.LoadConfigFromEnvVar()

	if c.IsConfigValid() {
		return nil
	}

	err := c.LoadConfigFromFile(cfgFilePath)
	if err != nil {
		return err
	}

	if c.IsConfigValid() {
		return nil
	}
	return ErrConfigurationNotFound
}

// SaveConfiguration saves the configuration to a file
// If the parameter is empty, it saves the configuration to the default file.
func (c *Config) SaveConfiguration(cfgFilePath string) error {
	var (
		yamlData []byte
		err      error
	)

	if c.Token == "" {
		return ErrInvalidToken
	}
	if c.Endpoint == "" {
		return ErrInvalidEndpoint
	}
	if cfgFilePath == "" {
		cfgFilePath = DefaultConfigFilePath()
		if err = os.MkdirAll(DefautConfigDir(), ConfigurationDirPerm); err != nil {
			return fmt.Errorf("error creating configuration directory: %w", err)
		}
	}
	if yamlData, err = yaml.Marshal(c); err != nil {
		return fmt.Errorf("error marshalling configuration: %w", err)
	}
	if err = os.WriteFile(cfgFilePath, yamlData, ConfigurationFilePerm); err != nil {
		return fmt.Errorf("error writing configuration file: %w", err)
	}
	return nil
}

// initHomedir initializes the homedir variable
// It tries to get the user's home directory using os.UserHomeDir()
// If it fails, it tries to get the home directory from the HOME environment variable.
func (c *Config) initHomedir() {
	var err error
	c.homedir, err = os.UserHomeDir()
	if err != nil {
		c.homedir = os.Getenv("HOME")
	}
}

// DefautConfigDir returns the default configuration directory path.
func DefautConfigDir() string {
	return filepath.Join(os.Getenv("HOME"), ".config", "eph")
}

// DefaultConfigFilePath returns the default configuration file path.
func DefaultConfigFilePath() string {
	return filepath.Join(DefautConfigDir(), "default.yml")
}

// ResolveConfigPath resolves a configuration name or path to a full file path.
// If the input contains a path separator or .yml extension, it's treated as a full path.
// Otherwise, it's treated as a config name and resolved to $HOME/.config/eph/<name>.yml.
func ResolveConfigPath(nameOrPath string) string {
	// Empty string returns default config path
	if nameOrPath == "" {
		return DefaultConfigFilePath()
	}

	// If it contains a path separator or .yml extension, treat as full path
	if filepath.IsAbs(nameOrPath) || filepath.Base(nameOrPath) != nameOrPath || filepath.Ext(nameOrPath) == ".yml" {
		return nameOrPath
	}

	// Otherwise, treat as a config name and resolve to config directory
	return filepath.Join(DefautConfigDir(), nameOrPath+".yml")
}
