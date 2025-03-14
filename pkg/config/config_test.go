package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/ephemeralfiles/eph/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	DefaultToken    = "sdf"
	DefaultEndpoint = "http://localhost:8080"
)

// createTempConfigFile creates a temporary file with a valid configuration.
func createValidConfigFile() (string, error) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("/tmp", "example")
	if err != nil {
		return "", fmt.Errorf("error creating temporary file: %w", err)
	}
	// Write to the file
	_, err = tmpfile.WriteString("token: sdf\nendpoint: http://localhost:8080")
	if err != nil {
		return "", fmt.Errorf("error writing to temporary file: %w", err)
	}
	// Close the file
	if err := tmpfile.Close(); err != nil {
		return "", fmt.Errorf("error closing temporary file: %w", err)
	}

	return tmpfile.Name(), nil
}
func TestReadyamlConfigFile(t *testing.T) {
	t.Parallel()
	t.Run("valid file", func(t *testing.T) {
		t.Parallel()
		// Create a temporary file
		tmpfile, err := createValidConfigFile()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile) // clean up

		// Test the function
		cfg := config.NewConfig()
		err = cfg.LoadConfigFromFile(tmpfile)
		require.NoError(t, err)
		assert.Equal(t, "sdf", cfg.Token)
		assert.Equal(t, "http://localhost:8080", cfg.Endpoint)
	})
	t.Run("invalid file", func(t *testing.T) {
		t.Parallel()
		// Create a temporary file
		tmpfile, err := os.CreateTemp("/tmp", "example")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name()) // clean up

		// Write to the file
		_, err = tmpfile.WriteString("token:\n  - test")
		if err != nil {
			t.Fatal(err)
		}
		// Close the file
		if err := tmpfile.Close(); err != nil {
			t.Fatal(err)
		}

		// Test the function
		cfg := config.NewConfig()
		err = cfg.LoadConfigFromFile(tmpfile.Name())
		require.Error(t, err)
	})
}

func TestReadyamlConfigFileReturnErrorIfFileDoesNotExist(t *testing.T) {
	t.Parallel()
	err := config.NewConfig().LoadConfigFromFile("/tmp/test.yml")
	require.Error(t, err)
}

func TestGetConfigFromEnvVar(t *testing.T) {
	t.Setenv("EPHEMERALFILES_TOKEN", DefaultToken)
	t.Setenv("EPHEMERALFILES_ENDPOINT", DefaultEndpoint)

	// Test the function
	cfg := config.NewConfig()
	cfgExpected := config.NewConfig()
	cfgExpected.Endpoint = DefaultEndpoint
	cfgExpected.Token = DefaultToken

	cfg.LoadConfigFromEnvVar()
	assert.Equal(t, cfgExpected, cfg)
}

func TestConfig_IsConfigValid(t *testing.T) {
	t.Parallel()

	type fields struct {
		Token    string
		Endpoint string
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "empty",
			fields: fields{Token: "", Endpoint: ""},
			want:   false,
		},
		{
			name:   "empty token",
			fields: fields{Token: "", Endpoint: DefaultEndpoint},
			want:   false,
		},
		{
			name:   "empty endpoint",
			fields: fields{Token: DefaultToken, Endpoint: ""},
			want:   false,
		},
		{
			name:   "valid",
			fields: fields{Token: DefaultToken, Endpoint: DefaultEndpoint},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := config.NewConfig()
			c.Token = tt.fields.Token
			c.Endpoint = tt.fields.Endpoint

			if got := c.IsConfigValid(); got != tt.want {
				t.Errorf("Config.IsConfigValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadConfigurationLoadEnvVarByDefault(t *testing.T) {
	// Test that env var overrides file
	t.Setenv("EPHEMERALFILES_TOKEN", "envvar")
	t.Setenv("EPHEMERALFILES_ENDPOINT", DefaultEndpoint)

	// Test the function
	cfg := config.NewConfig()
	cfg.SetHomedir("/tmp")
	err := cfg.LoadConfiguration("")
	require.NoError(t, err)
	assert.Equal(t, "envvar", cfg.Token)
	assert.Equal(t, DefaultEndpoint, cfg.Endpoint)
}

func TestLoadConfigurationWithNoEnv(t *testing.T) {
	t.Setenv("EPHEMERALFILES_TOKEN", "")
	t.Setenv("EPHEMERALFILES_ENDPOINT", "")
	err := os.MkdirAll("/tmp/.config/eph", 0755)
	require.NoError(t, err)

	defer os.RemoveAll("/tmp/.config")
	// Create file /tmp/.eph.yml
	tmpfile, err := os.Create("/tmp/.config/eph/default.yml") //nolint:gosec
	require.NoError(t, err)
	// Write to the file
	_, err = tmpfile.WriteString("token: sdf\nendpoint: http://localhost:8080")
	if err != nil {
		require.NoError(t, err)
	}
	// Close the file
	if err := tmpfile.Close(); err != nil {
		require.NoError(t, err)
	}

	// Test the function
	cfg := config.NewConfig()
	cfg.SetHomedir("/tmp")
	err = cfg.LoadConfiguration("/tmp/.config/eph/default.yml")
	require.NoError(t, err)
	assert.Equal(t, "sdf", cfg.Token)
	assert.Equal(t, "http://localhost:8080", cfg.Endpoint)
}

func TestLoadConfigurationNotFound(t *testing.T) {
	t.Setenv("EPHEMERALFILES_TOKEN", "")
	t.Setenv("EPHEMERALFILES_ENDPOINT", "")

	cfg := config.NewConfig()
	cfg.SetHomedir("/tmp")
	err := cfg.LoadConfiguration("/tmp/.config/eph/default.yml")
	require.Error(t, err)
}

func TestLoadConfigurationReturnsErrIfNotConfigured(t *testing.T) {
	t.Setenv("EPHEMERALFILES_TOKEN", "")
	t.Setenv("EPHEMERALFILES_ENDPOINT", "")
	err := os.MkdirAll("/tmp/.config/eph", 0755)
	require.NoError(t, err)

	defer os.RemoveAll("/tmp/.config")
	// Create file /tmp/.eph.yml
	tmpfile, err := os.Create("/tmp/.config/eph/default.yml") //nolint:gosec
	require.NoError(t, err)
	// Write to the file
	_, err = tmpfile.WriteString("token: \nendpoint: ")
	require.NoError(t, err)

	// Close the file
	if err := tmpfile.Close(); err != nil {
		require.NoError(t, err)
	}

	// Test the function
	cfg := config.NewConfig()
	cfg.SetHomedir("/tmp")
	err = cfg.LoadConfiguration("/tmp/.config/eph/default.yml")
	require.Error(t, err)
}

func TestSaveConfiguration(t *testing.T) {
	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		cfg := config.NewConfig()
		cfg.Endpoint = "http://localhost:8080"
		cfg.Token = "sdf"
		cfg.SetHomedir("/tmp")
		err := cfg.SaveConfiguration("/tmp/default.yml")
		require.NoError(t, err)
		// Load the configuration to check the values
		cfg = config.NewConfig()
		cfg.SetHomedir("/tmp")
		err = cfg.LoadConfiguration("/tmp/default.yml")
		require.NoError(t, err)
		assert.Equal(t, "sdf", cfg.Token)
		assert.Equal(t, "http://localhost:8080", cfg.Endpoint)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()
		cfg := config.NewConfig()
		cfg.SetHomedir("/tmp")
		err := cfg.SaveConfiguration("/tmp/.config/eph/default.yml")
		require.Error(t, err)
	})

	t.Run("invalid directory", func(t *testing.T) {
		t.Parallel()
		cfg := config.NewConfig()
		cfg.Endpoint = "http://localhost:8080"
		cfg.Token = "sdf"
		cfg.SetHomedir("/dir/does/not/exist")
		err := cfg.SaveConfiguration("/dir/does/not/exist/.config/eph/default.yml")
		require.Error(t, err)
	})
}
