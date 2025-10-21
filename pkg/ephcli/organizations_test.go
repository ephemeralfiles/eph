package ephcli_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ephemeralfiles/eph/pkg/dto"
	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListOrganizations(t *testing.T) {
	t.Parallel()

	t.Run("successful list", func(t *testing.T) {
		t.Parallel()

		expectedOrgs := []dto.Organization{
			{
				ID:                   "org-1",
				Name:                 "Test Org 1",
				DefaultRetentionDays: 30,
				CreatedAt:            "2024-01-01T00:00:00Z",
				UpdatedAt:            "2024-01-01T00:00:00Z",
				SubscriptionActive:   true,
				UserRole:             "admin",
				StorageLimitGB:       100.0,
				UsedStorageGB:        25.5,
			},
			{
				ID:                   "org-2",
				Name:                 "Test Org 2",
				DefaultRetentionDays: 60,
				CreatedAt:            "2024-01-02T00:00:00Z",
				UpdatedAt:            "2024-01-02T00:00:00Z",
				SubscriptionActive:   false,
				UserRole:             "member",
				StorageLimitGB:       50.0,
				UsedStorageGB:        10.0,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/api/v1/organizations")
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedOrgs)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		orgs, err := client.ListOrganizations()
		require.NoError(t, err)
		assert.Equal(t, expectedOrgs, orgs)
		assert.Len(t, orgs, 2)
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]dto.Organization{})
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		orgs, err := client.ListOrganizations()
		require.NoError(t, err)
		assert.Empty(t, orgs)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "unauthorized",
			})
		}))
		defer server.Close()

		client := ephcli.NewClient("invalid-token")
		client.SetEndpoint(server.URL)

		orgs, err := client.ListOrganizations()
		require.Error(t, err)
		assert.Nil(t, orgs)
		assert.Contains(t, err.Error(), "401")
	})

	t.Run("invalid json response", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`invalid json`))
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		orgs, err := client.ListOrganizations()
		require.Error(t, err)
		assert.Nil(t, orgs)
		assert.Contains(t, err.Error(), "failed to decode")
	})
}

func TestGetOrganization(t *testing.T) {
	t.Parallel()

	t.Run("successful get", func(t *testing.T) {
		t.Parallel()

		expectedOrg := dto.Organization{
			ID:                   "org-123",
			Name:                 "Test Organization",
			DefaultRetentionDays: 45,
			CreatedAt:            "2024-01-01T00:00:00Z",
			UpdatedAt:            "2024-01-15T00:00:00Z",
			SubscriptionActive:   true,
			UserRole:             "owner",
			StorageLimitGB:       200.0,
			UsedStorageGB:        75.5,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/api/v1/organizations/org-123")
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedOrg)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		org, err := client.GetOrganization("org-123")
		require.NoError(t, err)
		require.NotNil(t, org)
		assert.Equal(t, expectedOrg.ID, org.ID)
		assert.Equal(t, expectedOrg.Name, org.Name)
		assert.Equal(t, expectedOrg.StorageLimitGB, org.StorageLimitGB)
	})

	t.Run("organization not found", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "organization not found",
			})
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		org, err := client.GetOrganization("nonexistent-org")
		require.Error(t, err)
		assert.Nil(t, org)
		assert.Contains(t, err.Error(), "404")
	})
}

func TestGetOrganizationByName(t *testing.T) {
	t.Parallel()

	t.Run("found by name", func(t *testing.T) {
		t.Parallel()

		orgs := []dto.Organization{
			{
				ID:                 "org-1",
				Name:               "Production",
				SubscriptionActive: true,
			},
			{
				ID:                 "org-2",
				Name:               "Staging",
				SubscriptionActive: true,
			},
			{
				ID:                 "org-3",
				Name:               "Development",
				SubscriptionActive: false,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(orgs)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		org, err := client.GetOrganizationByName("Staging")
		require.NoError(t, err)
		require.NotNil(t, org)
		assert.Equal(t, "org-2", org.ID)
		assert.Equal(t, "Staging", org.Name)
	})

	t.Run("not found by name", func(t *testing.T) {
		t.Parallel()

		orgs := []dto.Organization{
			{
				ID:   "org-1",
				Name: "Production",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(orgs)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		org, err := client.GetOrganizationByName("NonExistent")
		require.Error(t, err)
		assert.Nil(t, org)
		assert.ErrorIs(t, err, ephcli.ErrOrganizationNotFound)
	})

	t.Run("empty organization list", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]dto.Organization{})
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		org, err := client.GetOrganizationByName("AnyOrg")
		require.Error(t, err)
		assert.Nil(t, org)
		assert.ErrorIs(t, err, ephcli.ErrOrganizationNotFound)
	})
}

func TestGetOrganizationStorage(t *testing.T) {
	t.Parallel()

	t.Run("successful get storage", func(t *testing.T) {
		t.Parallel()

		expectedStorage := dto.OrganizationStorage{
			ID:             "org-123",
			Name:           "Test Org",
			StorageLimitGB: 100.0,
			UsedStorageGB:  45.5,
			UsagePercent:   45.5,
			IsFull:         false,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/api/v1/organizations/org-123/storage")
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedStorage)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		storage, err := client.GetOrganizationStorage("org-123")
		require.NoError(t, err)
		require.NotNil(t, storage)
		assert.Equal(t, expectedStorage.ID, storage.ID)
		assert.Equal(t, expectedStorage.StorageLimitGB, storage.StorageLimitGB)
		assert.Equal(t, expectedStorage.UsedStorageGB, storage.UsedStorageGB)
		assert.Equal(t, expectedStorage.IsFull, storage.IsFull)
	})

	t.Run("storage full", func(t *testing.T) {
		t.Parallel()

		expectedStorage := dto.OrganizationStorage{
			ID:             "org-456",
			Name:           "Full Org",
			StorageLimitGB: 50.0,
			UsedStorageGB:  50.0,
			UsagePercent:   100.0,
			IsFull:         true,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedStorage)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		storage, err := client.GetOrganizationStorage("org-456")
		require.NoError(t, err)
		require.NotNil(t, storage)
		assert.True(t, storage.IsFull)
		assert.Equal(t, 100.0, storage.UsagePercent)
	})
}

func TestGetOrganizationStats(t *testing.T) {
	t.Parallel()

	t.Run("successful get stats", func(t *testing.T) {
		t.Parallel()

		expectedStats := dto.OrganizationStats{
			OrganizationID: "org-123",
			FileCount:      150,
			TotalSizeGB:    45.5,
			MemberCount:    5,
			ActiveFiles:    140,
			ExpiredFiles:   10,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/api/v1/organizations/org-123/stats")
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedStats)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		stats, err := client.GetOrganizationStats("org-123")
		require.NoError(t, err)
		require.NotNil(t, stats)
		assert.Equal(t, expectedStats.OrganizationID, stats.OrganizationID)
		assert.Equal(t, expectedStats.FileCount, stats.FileCount)
		assert.Equal(t, expectedStats.TotalSizeGB, stats.TotalSizeGB)
		assert.Equal(t, expectedStats.MemberCount, stats.MemberCount)
		assert.Equal(t, expectedStats.ActiveFiles, stats.ActiveFiles)
		assert.Equal(t, expectedStats.ExpiredFiles, stats.ExpiredFiles)
	})

	t.Run("empty stats", func(t *testing.T) {
		t.Parallel()

		expectedStats := dto.OrganizationStats{
			OrganizationID: "org-empty",
			FileCount:      0,
			TotalSizeGB:    0.0,
			MemberCount:    1,
			ActiveFiles:    0,
			ExpiredFiles:   0,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedStats)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		stats, err := client.GetOrganizationStats("org-empty")
		require.NoError(t, err)
		require.NotNil(t, stats)
		assert.Equal(t, int64(0), stats.FileCount)
		assert.Equal(t, 0.0, stats.TotalSizeGB)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "forbidden",
			})
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		stats, err := client.GetOrganizationStats("org-forbidden")
		require.Error(t, err)
		assert.Nil(t, stats)
		assert.Contains(t, err.Error(), "403")
	})
}

func TestOrganizationEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("list organizations with server error", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "internal server error",
			})
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		orgs, err := client.ListOrganizations()
		require.Error(t, err)
		assert.Nil(t, orgs)
		assert.Contains(t, err.Error(), "500")
	})

	t.Run("get organization with malformed response", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{invalid json}`))
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		org, err := client.GetOrganization("org-123")
		require.Error(t, err)
		assert.Nil(t, org)
	})
}
