package dto

import "time"

type InfoFile struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	NbParts  int    `json:"nb_parts"`
}

type RequestAESKey struct {
	AESKey string `json:"aeskey"`
}

// APIError is the error response from the API
type APIError struct {
	Err     bool   `json:"error"`
	Message string `json:"msg"`
}

// File is the struct that represents a file in the API
type File struct {
	FileID          string    `json:"fileid"`
	OwnerID         string    `json:"ownerid"`
	FileName        string    `json:"filename"`
	Size            int64     `json:"size"`
	UpdateDateBegin time.Time `json:"update_date_begin"`
	UpdateDateEnd   time.Time `json:"update_date_end"`
	ExpirationDate  time.Time `json:"expiration_date"`
}

// FileList is a list of files
type FileList []File
