// Package dto provides data transfer objects for the ephemeralfiles API.
// These structures define the request and response formats used in API communications.
package dto

import "time"

// InfoFile contains information about a file to be uploaded.
type InfoFile struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	NbParts  int    `json:"nb_parts"`
}

// RequestAESKey contains the AES encryption key for E2E encrypted operations.
type RequestAESKey struct {
	AESKey string `json:"aeskey"`
}

// APIError is the error response from the API.
// The API can return error in two formats:
// 1. {"error": true, "msg": "error message"} (legacy format).
// 2. {"error": "error message", "message": "error message"} (new format).
type APIError struct {
	Err        bool   `json:"-"` // Not directly unmarshaled
	Message    string `json:"message"`
	LegacyMsg  string `json:"msg"`
	ErrorField any    `json:"error"` // Can be bool or string
}

// GetMessage returns the error message from either format.
func (e *APIError) GetMessage() string {
	// Priority: Message > LegacyMsg > ErrorField (if string)
	if e.Message != "" {
		return e.Message
	}
	if e.LegacyMsg != "" {
		return e.LegacyMsg
	}
	// Check if ErrorField is a string (new format)
	if errStr, ok := e.ErrorField.(string); ok && errStr != "" {
		return errStr
	}
	return "unknown error"
}

// File is the struct that represents a file in the API.
type File struct {
	FileID          string    `json:"file_id"`
	OwnerID         string    `json:"owner_id"`
	FileName        string    `json:"filename"`
	Size            int64     `json:"size"`
	UpdateDateBegin time.Time `json:"update_date_begin"`
	UpdateDateEnd   time.Time `json:"update_date_end"`
	ExpirationDate  time.Time `json:"expiration_date"`
}

// FileList is a list of files.
type FileList []File

// Organization represents an organization entity.
type Organization struct {
	ID                   string  `json:"id"`
	Name                 string  `json:"name"`
	DefaultRetentionDays int32   `json:"default_retention_days"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
	SubscriptionActive   bool    `json:"subscription_active"`
	UserRole             string  `json:"user_role,omitempty"`
	StorageLimitGB       float64 `json:"storage_limit_gb"`
	UsedStorageGB        float64 `json:"used_storage_gb"`
}

// OrganizationStorage represents storage information for an organization.
type OrganizationStorage struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	StorageLimitGB float64 `json:"storage_limit_gb"`
	UsedStorageGB  float64 `json:"used_storage_gb"`
	UsagePercent   float64 `json:"usage_percent"`
	IsFull         bool    `json:"is_full"`
}

// OrganizationFile represents a file in an organization.
type OrganizationFile struct {
	ID              string   `json:"id"`
	Filename        string   `json:"filename"`
	Size            int64    `json:"size"`
	OrganizationID  string   `json:"organization_id"`
	Tags            []string `json:"tags,omitempty"`
	UploadDateBegin string   `json:"upload_date_begin"`
	UploadDateEnd   string   `json:"upload_date_end,omitempty"`
	ExpirationDate  string   `json:"expiration_date"`
	OwnerID         string   `json:"owner_id"`
	OwnerEmail      string   `json:"owner_email,omitempty"`
}

// OrganizationStats represents organization statistics.
type OrganizationStats struct {
	OrganizationID string  `json:"organization_id"`
	FileCount      int64   `json:"file_count"`
	TotalSizeGB    float64 `json:"total_size_gb"`
	MemberCount    int64   `json:"member_count"`
	ActiveFiles    int64   `json:"active_files"`
	ExpiredFiles   int64   `json:"expired_files"`
}

// TagCount represents a tag with its usage count.
type TagCount struct {
	Tag   string `json:"tag"`
	Count int64  `json:"count"`
}
