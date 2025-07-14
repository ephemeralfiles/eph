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
type APIError struct {
	Err     bool   `json:"error"`
	Message string `json:"msg"`
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
