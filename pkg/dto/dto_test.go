package dto_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ephemeralfiles/eph/pkg/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfoFile(t *testing.T) {
	t.Parallel()

	t.Run("JSON marshaling", func(t *testing.T) {
		t.Parallel()

		info := dto.InfoFile{
			Filename: "test.txt",
			Size:     1024,
			NbParts:  3,
		}

		data, err := json.Marshal(info)
		require.NoError(t, err)

		expected := `{"filename":"test.txt","size":1024,"nb_parts":3}`
		assert.JSONEq(t, expected, string(data))
	})

	t.Run("JSON unmarshaling", func(t *testing.T) {
		t.Parallel()

		jsonData := `{"filename":"document.pdf","size":2048,"nb_parts":5}`

		var info dto.InfoFile
		err := json.Unmarshal([]byte(jsonData), &info)
		require.NoError(t, err)

		assert.Equal(t, "document.pdf", info.Filename)
		assert.Equal(t, int64(2048), info.Size)
		assert.Equal(t, 5, info.NbParts)
	})

	t.Run("empty values", func(t *testing.T) {
		t.Parallel()

		info := dto.InfoFile{}

		data, err := json.Marshal(info)
		require.NoError(t, err)

		expected := `{"filename":"","size":0,"nb_parts":0}`
		assert.JSONEq(t, expected, string(data))
	})

	t.Run("special characters in filename", func(t *testing.T) {
		t.Parallel()

		info := dto.InfoFile{
			Filename: "test file with spaces & symbols.txt",
			Size:     512,
			NbParts:  1,
		}

		data, err := json.Marshal(info)
		require.NoError(t, err)

		var unmarshaled dto.InfoFile
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, info.Filename, unmarshaled.Filename)
		assert.Equal(t, info.Size, unmarshaled.Size)
		assert.Equal(t, info.NbParts, unmarshaled.NbParts)
	})
}

func TestRequestAESKey(t *testing.T) {
	t.Parallel()

	t.Run("JSON marshaling", func(t *testing.T) {
		t.Parallel()

		req := dto.RequestAESKey{
			AESKey: "abcd1234567890",
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		expected := `{"aeskey":"abcd1234567890"}`
		assert.JSONEq(t, expected, string(data))
	})

	t.Run("JSON unmarshaling", func(t *testing.T) {
		t.Parallel()

		jsonData := `{"aeskey":"xyz9876543210"}`

		var req dto.RequestAESKey
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "xyz9876543210", req.AESKey)
	})

	t.Run("empty AES key", func(t *testing.T) {
		t.Parallel()

		req := dto.RequestAESKey{}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		expected := `{"aeskey":""}`
		assert.JSONEq(t, expected, string(data))
	})

	t.Run("long AES key", func(t *testing.T) {
		t.Parallel()

		longKey := "this-is-a-very-long-aes-key-that-might-be-used-in-real-scenarios-1234567890"
		req := dto.RequestAESKey{AESKey: longKey}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var unmarshaled dto.RequestAESKey
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, longKey, unmarshaled.AESKey)
	})
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	t.Run("JSON marshaling", func(t *testing.T) {
		t.Parallel()

		apiErr := dto.APIError{
			Err:     true,
			Message: "File not found",
		}

		data, err := json.Marshal(apiErr)
		require.NoError(t, err)

		expected := `{"error":true,"msg":"File not found"}`
		assert.JSONEq(t, expected, string(data))
	})

	t.Run("JSON unmarshaling", func(t *testing.T) {
		t.Parallel()

		jsonData := `{"error":false,"msg":"Success"}`

		var apiErr dto.APIError
		err := json.Unmarshal([]byte(jsonData), &apiErr)
		require.NoError(t, err)

		assert.False(t, apiErr.Err)
		assert.Equal(t, "Success", apiErr.Message)
	})

	t.Run("error states", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			err     bool
			message string
		}{
			{"success state", false, "Operation completed"},
			{"error state", true, "Something went wrong"},
			{"empty message", true, ""},
			{"false with message", false, "No error but has message"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				apiErr := dto.APIError{
					Err:     tt.err,
					Message: tt.message,
				}

				data, err := json.Marshal(apiErr)
				require.NoError(t, err)

				var unmarshaled dto.APIError
				err = json.Unmarshal(data, &unmarshaled)
				require.NoError(t, err)

				assert.Equal(t, tt.err, unmarshaled.Err)
				assert.Equal(t, tt.message, unmarshaled.Message)
			})
		}
	})
}

func TestFile(t *testing.T) {
	t.Parallel()

	t.Run("JSON marshaling", func(t *testing.T) {
		t.Parallel()

		testTime := time.Date(2023, 10, 15, 14, 30, 45, 0, time.UTC)

		file := dto.File{
			FileID:          "file123",
			OwnerID:         "user456",
			FileName:        "example.txt",
			Size:            1024,
			UpdateDateBegin: testTime,
			UpdateDateEnd:   testTime.Add(time.Hour),
			ExpirationDate:  testTime.Add(24 * time.Hour),
		}

		data, err := json.Marshal(file)
		require.NoError(t, err)

		// Verify JSON contains expected fields
		assert.Contains(t, string(data), `"file_id":"file123"`)
		assert.Contains(t, string(data), `"owner_id":"user456"`)
		assert.Contains(t, string(data), `"filename":"example.txt"`)
		assert.Contains(t, string(data), `"size":1024`)
	})

	t.Run("JSON unmarshaling", func(t *testing.T) {
		t.Parallel()

		jsonData := `{
			"file_id": "abc123",
			"owner_id": "owner789",
			"filename": "document.pdf",
			"size": 2048,
			"update_date_begin": "2023-10-15T14:30:45Z",
			"update_date_end": "2023-10-15T15:30:45Z",
			"expiration_date": "2023-10-16T14:30:45Z"
		}`

		var file dto.File
		err := json.Unmarshal([]byte(jsonData), &file)
		require.NoError(t, err)

		assert.Equal(t, "abc123", file.FileID)
		assert.Equal(t, "owner789", file.OwnerID)
		assert.Equal(t, "document.pdf", file.FileName)
		assert.Equal(t, int64(2048), file.Size)

		expectedTime := time.Date(2023, 10, 15, 14, 30, 45, 0, time.UTC)
		assert.Equal(t, expectedTime, file.UpdateDateBegin)
		assert.Equal(t, expectedTime.Add(time.Hour), file.UpdateDateEnd)
		assert.Equal(t, expectedTime.Add(24*time.Hour), file.ExpirationDate)
	})

	t.Run("round trip conversion", func(t *testing.T) {
		t.Parallel()

		original := dto.File{
			FileID:          "test-file-id",
			OwnerID:         "test-owner-id",
			FileName:        "test-file.txt",
			Size:            4096,
			UpdateDateBegin: time.Now().UTC(),
			UpdateDateEnd:   time.Now().Add(time.Minute).UTC(),
			ExpirationDate:  time.Now().Add(time.Hour).UTC(),
		}

		// Marshal to JSON
		data, err := json.Marshal(original)
		require.NoError(t, err)

		// Unmarshal back to struct
		var roundTrip dto.File
		err = json.Unmarshal(data, &roundTrip)
		require.NoError(t, err)

		// Compare (with time truncation due to JSON precision)
		assert.Equal(t, original.FileID, roundTrip.FileID)
		assert.Equal(t, original.OwnerID, roundTrip.OwnerID)
		assert.Equal(t, original.FileName, roundTrip.FileName)
		assert.Equal(t, original.Size, roundTrip.Size)
		
		// Times might lose precision in JSON, so we check within a second
		assert.WithinDuration(t, original.UpdateDateBegin, roundTrip.UpdateDateBegin, time.Second)
		assert.WithinDuration(t, original.UpdateDateEnd, roundTrip.UpdateDateEnd, time.Second)
		assert.WithinDuration(t, original.ExpirationDate, roundTrip.ExpirationDate, time.Second)
	})
}

func TestFileList(t *testing.T) {
	t.Parallel()

	t.Run("empty file list", func(t *testing.T) {
		t.Parallel()

		var fileList dto.FileList

		data, err := json.Marshal(fileList)
		require.NoError(t, err)

		assert.Equal(t, "null", string(data))

		// Test unmarshaling empty list
		jsonData := `[]`
		err = json.Unmarshal([]byte(jsonData), &fileList)
		require.NoError(t, err)
		assert.Empty(t, fileList)
	})

	t.Run("file list with multiple files", func(t *testing.T) {
		t.Parallel()

		testTime := time.Date(2023, 10, 15, 14, 30, 45, 0, time.UTC)

		fileList := dto.FileList{
			{
				FileID:          "file1",
				OwnerID:         "owner1",
				FileName:        "file1.txt",
				Size:            100,
				UpdateDateBegin: testTime,
				UpdateDateEnd:   testTime.Add(time.Minute),
				ExpirationDate:  testTime.Add(time.Hour),
			},
			{
				FileID:          "file2",
				OwnerID:         "owner2",
				FileName:        "file2.pdf",
				Size:            200,
				UpdateDateBegin: testTime.Add(time.Hour),
				UpdateDateEnd:   testTime.Add(time.Hour + time.Minute),
				ExpirationDate:  testTime.Add(2 * time.Hour),
			},
		}

		data, err := json.Marshal(fileList)
		require.NoError(t, err)

		var unmarshaled dto.FileList
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Len(t, unmarshaled, 2)
		assert.Equal(t, "file1", unmarshaled[0].FileID)
		assert.Equal(t, "file2", unmarshaled[1].FileID)
		assert.Equal(t, "file1.txt", unmarshaled[0].FileName)
		assert.Equal(t, "file2.pdf", unmarshaled[1].FileName)
	})

	t.Run("file list append operations", func(t *testing.T) {
		t.Parallel()

		var fileList dto.FileList

		// Test appending files
		file1 := dto.File{FileID: "1", FileName: "test1.txt"}
		file2 := dto.File{FileID: "2", FileName: "test2.txt"}

		fileList = append(fileList, file1)
		assert.Len(t, fileList, 1)

		fileList = append(fileList, file2)
		assert.Len(t, fileList, 2)

		assert.Equal(t, "1", fileList[0].FileID)
		assert.Equal(t, "2", fileList[1].FileID)
	})
}

func TestDTOEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("invalid JSON handling", func(t *testing.T) {
		t.Parallel()

		invalidJSON := `{"filename": "test", "size": "not_a_number"}`

		var info dto.InfoFile
		err := json.Unmarshal([]byte(invalidJSON), &info)
		assert.Error(t, err, "Should fail to unmarshal invalid JSON")
	})

	t.Run("missing JSON fields", func(t *testing.T) {
		t.Parallel()

		partialJSON := `{"filename": "test.txt"}`

		var info dto.InfoFile
		err := json.Unmarshal([]byte(partialJSON), &info)
		require.NoError(t, err)

		assert.Equal(t, "test.txt", info.Filename)
		assert.Equal(t, int64(0), info.Size) // Default value
		assert.Equal(t, 0, info.NbParts)     // Default value
	})

	t.Run("unicode filename handling", func(t *testing.T) {
		t.Parallel()

		info := dto.InfoFile{
			Filename: "测试文件.txt",
			Size:     1024,
			NbParts:  1,
		}

		data, err := json.Marshal(info)
		require.NoError(t, err)

		var unmarshaled dto.InfoFile
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, "测试文件.txt", unmarshaled.Filename)
	})
}