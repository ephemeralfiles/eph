package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	// 0700 is the permission for the configuration directory
	ConfigurationDirPerm  = 0700
	ConfigurationFilePerm = 0600
)

var (
	ErrConfigurationNotFound = errors.New("configuration not found")
	ErrInvalidToken          = errors.New("token is invalid")
	ErrInvalidEndpoint       = errors.New("endpoint is invalid")
)

// Config is the configuration for the application.
type Config struct {
	Token    string `yaml:"token"`
	Endpoint string `yaml:"endpoint"`
	homedir  string
}

// NewConfig creates a new configuration for the application
func NewConfig() *Config {
	cfg := &Config{
		Token:    "",
		Endpoint: "",
		homedir:  "",
	}
	cfg.initHomedir()

	return cfg
}

// initHomedir initializes the homedir variable
// It tries to get the user's home directory using os.UserHomeDir()
// If it fails, it tries to get the home directory from the HOME environment variable
func (c *Config) initHomedir() {
	var err error
	c.homedir, err = os.UserHomeDir()
	if err != nil {
		c.homedir = os.Getenv("HOME")
	}
}

// SetHomedir sets the homedir variable
// It is used for testing purposes
func (c *Config) SetHomedir(homedir string) {
	c.homedir = homedir
}

// LoadConfigFromFile loads the configuration from a file
func (c *Config) LoadConfigFromFile(filename string) error {
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

// LoadConfigFromEnvVar loads the configuration from the environment variables
func (c *Config) LoadConfigFromEnvVar() {
	c.Token = os.Getenv("EPHEMERALFILES_TOKEN")
	c.Endpoint = os.Getenv("EPHEMERALFILES_ENDPOINT")
}

// IsConfigValid checks if the configuration is valid
func (c *Config) IsConfigValid() bool {
	if c.Token == "" || c.Endpoint == "" {
		return false
	}
	return true
}

// LoadConfiguration loads the configuration from the environment variables first
// If the configuration is not valid, it tries to load the configuration from the $HOME/.eph.yml file
func (c *Config) LoadConfiguration() error {
	c.LoadConfigFromEnvVar()

	if c.IsConfigValid() {
		return nil
	}

	err := c.LoadConfigFromFile(filepath.Join(c.homedir, ".config", "eph", "default.yml"))
	if err != nil {
		return err
	}

	if c.IsConfigValid() {
		return nil
	}
	return ErrConfigurationNotFound
}

func (c *Config) SaveConfiguration() error {
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
	if err = os.MkdirAll(filepath.Join(c.homedir, ".config", "eph"), ConfigurationDirPerm); err != nil {
		return fmt.Errorf("error creating configuration directory: %w", err)
	}
	if yamlData, err = yaml.Marshal(c); err != nil {
		return fmt.Errorf("error marshalling configuration: %w", err)
	}
	if err = os.WriteFile(filepath.Join(c.homedir, ".config", "eph", "default.yml"),
		yamlData, ConfigurationFilePerm); err != nil {
		return fmt.Errorf("error writing configuration file: %w", err)
	}
	return nil
}
