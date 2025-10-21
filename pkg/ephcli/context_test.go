package ephcli_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ephemeralfiles/eph/pkg/config"
	"github.com/ephemeralfiles/eph/pkg/dto"
	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrgContext(t *testing.T) {
	t.Parallel()

	client := ephcli.NewClient("test-token")
	cfg := config.NewConfig()

	ctx := ephcli.NewOrgContext(client, cfg)

	assert.NotNil(t, ctx)
}

func TestResolveOrganization(t *testing.T) {
	t.Parallel()

	t.Run("priority 1: explicit org-id flag", func(t *testing.T) {
		t.Parallel()

		expectedOrg := dto.Organization{
			ID:   "org-id-123",
			Name: "Org By ID",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Should request by ID
			assert.Contains(t, r.URL.Path, "/api/v1/organizations/org-id-123")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedOrg)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		cfg := config.NewConfig()
		cfg.DefaultOrganization = "should-be-ignored"

		ctx := ephcli.NewOrgContext(client, cfg)

		org, err := ctx.ResolveOrganization("should-also-be-ignored", "org-id-123")
		require.NoError(t, err)
		require.NotNil(t, org)
		assert.Equal(t, "org-id-123", org.ID)
	})

	t.Run("priority 2: explicit org name flag", func(t *testing.T) {
		t.Parallel()

		orgs := []dto.Organization{
			{
				ID:   "org-1",
				Name: "Production",
			},
			{
				ID:   "org-2",
				Name: "Staging",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Should list organizations to find by name
			assert.Contains(t, r.URL.Path, "/api/v1/organizations")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(orgs)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		cfg := config.NewConfig()
		cfg.DefaultOrganization = "should-be-ignored"

		ctx := ephcli.NewOrgContext(client, cfg)

		org, err := ctx.ResolveOrganization("Staging", "")
		require.NoError(t, err)
		require.NotNil(t, org)
		assert.Equal(t, "org-2", org.ID)
		assert.Equal(t, "Staging", org.Name)
	})

	t.Run("priority 3: config default organization by name", func(t *testing.T) {
		t.Parallel()

		orgs := []dto.Organization{
			{
				ID:   "org-default",
				Name: "DefaultOrg",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Should list organizations to find default
			assert.Contains(t, r.URL.Path, "/api/v1/organizations")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(orgs)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		cfg := config.NewConfig()
		cfg.DefaultOrganization = "DefaultOrg"

		ctx := ephcli.NewOrgContext(client, cfg)

		org, err := ctx.ResolveOrganization("", "")
		require.NoError(t, err)
		require.NotNil(t, org)
		assert.Equal(t, "org-default", org.ID)
		assert.Equal(t, "DefaultOrg", org.Name)
	})

	t.Run("priority 3: config default organization by ID", func(t *testing.T) {
		t.Parallel()

		expectedOrg := dto.Organization{
			ID:   "org-default-id",
			Name: "DefaultOrgByID",
		}

		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				// First call: try to find by name (returns empty list)
				assert.Contains(t, r.URL.Path, "/api/v1/organizations")
				assert.NotContains(t, r.URL.Path, "org-default-id")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode([]dto.Organization{})
			} else {
				// Second call: get by ID
				assert.Contains(t, r.URL.Path, "/api/v1/organizations/org-default-id")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(expectedOrg)
			}
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		cfg := config.NewConfig()
		cfg.DefaultOrganization = "org-default-id"

		ctx := ephcli.NewOrgContext(client, cfg)

		org, err := ctx.ResolveOrganization("", "")
		require.NoError(t, err)
		require.NotNil(t, org)
		assert.Equal(t, "org-default-id", org.ID)
	})

	t.Run("no organization specified", func(t *testing.T) {
		t.Parallel()

		client := ephcli.NewClient("test-token")
		cfg := config.NewConfig()
		cfg.DefaultOrganization = ""

		ctx := ephcli.NewOrgContext(client, cfg)

		org, err := ctx.ResolveOrganization("", "")
		require.Error(t, err)
		assert.Nil(t, org)
		assert.ErrorIs(t, err, ephcli.ErrNoOrganizationSpecified)
	})

	t.Run("default organization not found", func(t *testing.T) {
		t.Parallel()

		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				// First call: list organizations (empty)
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode([]dto.Organization{})
			} else {
				// Second call: get by ID (not found)
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "organization not found",
				})
			}
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		cfg := config.NewConfig()
		cfg.DefaultOrganization = "nonexistent"

		ctx := ephcli.NewOrgContext(client, cfg)

		org, err := ctx.ResolveOrganization("", "")
		require.Error(t, err)
		assert.Nil(t, org)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestResolveOrganizationID(t *testing.T) {
	t.Parallel()

	t.Run("resolve by name", func(t *testing.T) {
		t.Parallel()

		orgs := []dto.Organization{
			{
				ID:   "org-123",
				Name: "MyOrg",
			},
			{
				ID:   "org-456",
				Name: "OtherOrg",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Should list organizations to find by name
			assert.Contains(t, r.URL.Path, "/api/v1/organizations")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(orgs)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		cfg := config.NewConfig()
		ctx := ephcli.NewOrgContext(client, cfg)

		orgID, err := ctx.ResolveOrganizationID("MyOrg")
		require.NoError(t, err)
		assert.Equal(t, "org-123", orgID)
	})

	t.Run("resolve by ID when name fails", func(t *testing.T) {
		t.Parallel()

		expectedOrg := dto.Organization{
			ID:   "org-direct-id",
			Name: "DirectOrg",
		}

		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				// First call: list organizations (no match)
				assert.Contains(t, r.URL.Path, "/api/v1/organizations")
				assert.NotContains(t, r.URL.Path, "org-direct-id")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode([]dto.Organization{})
			} else {
				// Second call: get by ID
				assert.Contains(t, r.URL.Path, "/api/v1/organizations/org-direct-id")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(expectedOrg)
			}
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		cfg := config.NewConfig()
		ctx := ephcli.NewOrgContext(client, cfg)

		orgID, err := ctx.ResolveOrganizationID("org-direct-id")
		require.NoError(t, err)
		assert.Equal(t, "org-direct-id", orgID)
	})

	t.Run("organization not found by name or ID", func(t *testing.T) {
		t.Parallel()

		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				// First call: list organizations (empty)
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode([]dto.Organization{})
			} else {
				// Second call: get by ID (not found)
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "not found",
				})
			}
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		cfg := config.NewConfig()
		ctx := ephcli.NewOrgContext(client, cfg)

		orgID, err := ctx.ResolveOrganizationID("nonexistent")
		require.Error(t, err)
		assert.Empty(t, orgID)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("empty name or ID", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Should still try to list
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]dto.Organization{})
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		cfg := config.NewConfig()
		ctx := ephcli.NewOrgContext(client, cfg)

		orgID, err := ctx.ResolveOrganizationID("")
		require.Error(t, err)
		assert.Empty(t, orgID)
	})
}

func TestOrgContextEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("nil client", func(t *testing.T) {
		t.Parallel()

		cfg := config.NewConfig()

		// Should not panic with nil client
		assert.NotPanics(t, func() {
			ephcli.NewOrgContext(nil, cfg)
		})
	})

	t.Run("nil config", func(t *testing.T) {
		t.Parallel()

		client := ephcli.NewClient("test-token")

		// Should not panic with nil config
		assert.NotPanics(t, func() {
			ephcli.NewOrgContext(client, nil)
		})
	})

	t.Run("special characters in organization name", func(t *testing.T) {
		t.Parallel()

		orgs := []dto.Organization{
			{
				ID:   "org-special",
				Name: "Org-with-dashes_and_underscores",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(orgs)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		cfg := config.NewConfig()
		ctx := ephcli.NewOrgContext(client, cfg)

		org, err := ctx.ResolveOrganization("Org-with-dashes_and_underscores", "")
		require.NoError(t, err)
		assert.Equal(t, "org-special", org.ID)
	})

	t.Run("multiple organizations with similar names", func(t *testing.T) {
		t.Parallel()

		orgs := []dto.Organization{
			{
				ID:   "org-1",
				Name: "Production",
			},
			{
				ID:   "org-2",
				Name: "Production-Backup",
			},
			{
				ID:   "org-3",
				Name: "Pre-Production",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(orgs)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		cfg := config.NewConfig()
		ctx := ephcli.NewOrgContext(client, cfg)

		// Should match exactly, not partial match
		org, err := ctx.ResolveOrganization("Production", "")
		require.NoError(t, err)
		assert.Equal(t, "org-1", org.ID)
		assert.Equal(t, "Production", org.Name)
	})
}
