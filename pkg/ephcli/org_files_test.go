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

func TestListOrganizationFiles(t *testing.T) {
	t.Parallel()

	t.Run("successful list with pagination", func(t *testing.T) {
		t.Parallel()

		expectedFiles := []dto.OrganizationFile{
			{
				ID:              "file-1",
				Filename:        "document1.pdf",
				Size:            1024,
				OrganizationID:  "org-123",
				Tags:            []string{"important", "document"},
				UploadDateBegin: "2024-01-01T00:00:00Z",
				ExpirationDate:  "2024-02-01T00:00:00Z",
				OwnerID:         "user-1",
				OwnerEmail:      "user1@example.com",
			},
			{
				ID:              "file-2",
				Filename:        "image.jpg",
				Size:            2048,
				OrganizationID:  "org-123",
				Tags:            []string{"image"},
				UploadDateBegin: "2024-01-02T00:00:00Z",
				ExpirationDate:  "2024-02-02T00:00:00Z",
				OwnerID:         "user-2",
				OwnerEmail:      "user2@example.com",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/api/v1/organizations/org-123/files")
			assert.Equal(t, "10", r.URL.Query().Get("limit"))
			assert.Equal(t, "0", r.URL.Query().Get("offset"))
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedFiles)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		files, err := client.ListOrganizationFiles("org-123", 10, 0)
		require.NoError(t, err)
		assert.Equal(t, expectedFiles, files)
		assert.Len(t, files, 2)
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]dto.OrganizationFile{})
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		files, err := client.ListOrganizationFiles("org-empty", 10, 0)
		require.NoError(t, err)
		assert.Empty(t, files)
	})

	t.Run("pagination with offset", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "50", r.URL.Query().Get("limit"))
			assert.Equal(t, "100", r.URL.Query().Get("offset"))
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]dto.OrganizationFile{})
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		_, err := client.ListOrganizationFiles("org-123", 50, 100)
		require.NoError(t, err)
	})
}

func TestGetOrganizationFilesByTags(t *testing.T) {
	t.Parallel()

	t.Run("filter by single tag", func(t *testing.T) {
		t.Parallel()

		expectedFiles := []dto.OrganizationFile{
			{
				ID:             "file-1",
				Filename:       "tagged.pdf",
				Tags:           []string{"important"},
				OrganizationID: "org-123",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/api/v1/organizations/org-123/files/tags")
			assert.Contains(t, r.URL.RawQuery, "tags=important")

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedFiles)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		files, err := client.GetOrganizationFilesByTags("org-123", []string{"important"}, 10, 0)
		require.NoError(t, err)
		assert.Len(t, files, 1)
		assert.Contains(t, files[0].Tags, "important")
	})

	t.Run("filter by multiple tags", func(t *testing.T) {
		t.Parallel()

		expectedFiles := []dto.OrganizationFile{
			{
				ID:             "file-1",
				Filename:       "multi-tagged.pdf",
				Tags:           []string{"important", "urgent", "review"},
				OrganizationID: "org-123",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Contains(t, r.URL.RawQuery, "tags=")
			// Tags should be comma-separated and URL encoded
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedFiles)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		files, err := client.GetOrganizationFilesByTags("org-123", []string{"important", "urgent", "review"}, 10, 0)
		require.NoError(t, err)
		assert.Len(t, files, 1)
	})

	t.Run("no files match tags", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]dto.OrganizationFile{})
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		files, err := client.GetOrganizationFilesByTags("org-123", []string{"nonexistent"}, 10, 0)
		require.NoError(t, err)
		assert.Empty(t, files)
	})
}

func TestListRecentOrganizationFiles(t *testing.T) {
	t.Parallel()

	t.Run("successful list recent", func(t *testing.T) {
		t.Parallel()

		expectedFiles := []dto.OrganizationFile{
			{
				ID:              "file-recent-1",
				Filename:        "recent1.pdf",
				UploadDateBegin: "2024-01-15T10:00:00Z",
				OrganizationID:  "org-123",
			},
			{
				ID:              "file-recent-2",
				Filename:        "recent2.pdf",
				UploadDateBegin: "2024-01-15T09:00:00Z",
				OrganizationID:  "org-123",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/api/v1/organizations/org-123/files/recent")
			assert.Equal(t, "20", r.URL.Query().Get("limit"))

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedFiles)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		files, err := client.ListRecentOrganizationFiles("org-123", 20)
		require.NoError(t, err)
		assert.Len(t, files, 2)
	})
}

func TestListExpiredOrganizationFiles(t *testing.T) {
	t.Parallel()

	t.Run("successful list expired", func(t *testing.T) {
		t.Parallel()

		expectedFiles := []dto.OrganizationFile{
			{
				ID:             "file-expired-1",
				Filename:       "expired1.pdf",
				ExpirationDate: "2024-01-01T00:00:00Z",
				OrganizationID: "org-123",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/api/v1/organizations/org-123/files/expired")
			assert.Equal(t, "10", r.URL.Query().Get("limit"))

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedFiles)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		files, err := client.ListExpiredOrganizationFiles("org-123", 10)
		require.NoError(t, err)
		assert.Len(t, files, 1)
	})
}

func TestGetPopularTags(t *testing.T) {
	t.Parallel()

	t.Run("successful get popular tags", func(t *testing.T) {
		t.Parallel()

		expectedTags := []dto.TagCount{
			{
				Tag:   "important",
				Count: 50,
			},
			{
				Tag:   "document",
				Count: 30,
			},
			{
				Tag:   "image",
				Count: 20,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/api/v1/organizations/org-123/tags")
			assert.Equal(t, "10", r.URL.Query().Get("limit"))

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedTags)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		tags, err := client.GetPopularTags("org-123", 10)
		require.NoError(t, err)
		assert.Len(t, tags, 3)
		assert.Equal(t, "important", tags[0].Tag)
		assert.Equal(t, int64(50), tags[0].Count)
	})

	t.Run("no tags", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]dto.TagCount{})
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		tags, err := client.GetPopularTags("org-empty", 10)
		require.NoError(t, err)
		assert.Empty(t, tags)
	})
}

func TestDeleteOrganizationFile(t *testing.T) {
	t.Parallel()

	t.Run("successful delete", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "DELETE", r.Method)
			assert.Contains(t, r.URL.Path, "/api/v1/files/file-123")
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		err := client.DeleteOrganizationFile("file-123")
		require.NoError(t, err)
	})

	t.Run("file not found", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "file not found",
			})
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		err := client.DeleteOrganizationFile("nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})

	t.Run("unauthorized delete", func(t *testing.T) {
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

		err := client.DeleteOrganizationFile("file-123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "403")
	})
}

func TestUpdateFileTags(t *testing.T) {
	t.Parallel()

	t.Run("successful update", func(t *testing.T) {
		t.Parallel()

		updatedFile := dto.OrganizationFile{
			ID:       "file-123",
			Filename: "document.pdf",
			Tags:     []string{"updated", "new-tag"},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PUT", r.Method)
			assert.Contains(t, r.URL.Path, "/api/v1/files/file-123/tags")
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			// Verify request body contains tags
			var body map[string][]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			assert.Contains(t, body, "tags")
			assert.ElementsMatch(t, []string{"updated", "new-tag"}, body["tags"])

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(updatedFile)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		file, err := client.UpdateFileTags("file-123", []string{"updated", "new-tag"})
		require.NoError(t, err)
		require.NotNil(t, file)
		assert.Equal(t, "file-123", file.ID)
		assert.ElementsMatch(t, []string{"updated", "new-tag"}, file.Tags)
	})

	t.Run("update with empty tags", func(t *testing.T) {
		t.Parallel()

		updatedFile := dto.OrganizationFile{
			ID:       "file-123",
			Filename: "document.pdf",
			Tags:     []string{},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string][]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			assert.Contains(t, body, "tags")
			assert.Empty(t, body["tags"])

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(updatedFile)
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		file, err := client.UpdateFileTags("file-123", []string{})
		require.NoError(t, err)
		require.NotNil(t, file)
		assert.Empty(t, file.Tags)
	})

	t.Run("file not found", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "file not found",
			})
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		file, err := client.UpdateFileTags("nonexistent", []string{"tag"})
		require.Error(t, err)
		assert.Nil(t, file)
		assert.Contains(t, err.Error(), "404")
	})
}

func TestOrgFilesEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("special characters in tags", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Tags with special chars should be URL encoded
			assert.Contains(t, r.URL.RawQuery, "tags=")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]dto.OrganizationFile{})
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		_, err := client.GetOrganizationFilesByTags("org-123", []string{"tag with spaces", "tag-with-dashes"}, 10, 0)
		require.NoError(t, err)
	})

	t.Run("very large limit", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "1000", r.URL.Query().Get("limit"))
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]dto.OrganizationFile{})
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		_, err := client.ListOrganizationFiles("org-123", 1000, 0)
		require.NoError(t, err)
	})

	t.Run("negative pagination values", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Should still pass negative values to API
			// API should handle validation
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "invalid pagination",
			})
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		_, err := client.ListOrganizationFiles("org-123", -1, -1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "400")
	})

	t.Run("malformed JSON in list response", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{invalid json`))
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		_, err := client.ListOrganizationFiles("org-123", 10, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode")
	})

	t.Run("malformed JSON in tags response", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{malformed`))
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		_, err := client.GetPopularTags("org-123", 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode")
	})

	t.Run("malformed JSON in update tags response", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`not valid json`))
		}))
		defer server.Close()

		client := ephcli.NewClient("test-token")
		client.SetEndpoint(server.URL)

		_, err := client.UpdateFileTags("file-123", []string{"tag"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode")
	})
}
