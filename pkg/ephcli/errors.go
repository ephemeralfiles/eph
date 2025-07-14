package ephcli

import "errors"

// Error constants for the ephcli package.
var (
	// HTTP and API errors.
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
	ErrSendingRequest       = errors.New("error sending request")
	ErrCreatingRequest      = errors.New("error creating request")
	ErrDecodingResponse     = errors.New("error decoding response")
	ErrReadingResponse      = errors.New("error reading response")

	// Encryption and security errors.
	ErrParsePEMBlock          = errors.New("failed to parse PEM block")
	ErrParsePublicKey         = errors.New("failed to parse public key")
	ErrNotRSAPublicKey        = errors.New("not an RSA public key")
	ErrEncryptionFailed       = errors.New("encryption failed")
	ErrCiphertextTooShort     = errors.New("ciphertext too short")
	ErrGeneratingAESKey       = errors.New("error generating AES key")
	ErrCreatingCipherBlock    = errors.New("error creating new cipher block")
	ErrGeneratingRandomIV     = errors.New("error generating random IV")

	// File and I/O errors.
	ErrOpeningFile         = errors.New("error opening file")
	ErrGettingFileInfo     = errors.New("error getting file info")
	ErrReadingChunk        = errors.New("error reading chunk")
	ErrEncryptingChunk     = errors.New("error encrypting chunk")
	ErrCreatingFormFile    = errors.New("error creating form file")
	ErrWritingChunkToForm  = errors.New("error writing chunk to form")
	ErrClosingWriter       = errors.New("error closing writer")
	ErrSeekingInFile       = errors.New("error seeking in file")
	ErrWritingChunkToFile  = errors.New("error writing chunk to file")
	ErrDecryptingChunk     = errors.New("error decrypting chunk")

	// Payload and marshalling errors.
	ErrMarshallingPayload = errors.New("error marshalling payload")

	// GitHub release errors.
	ErrGettingLatestRelease = errors.New("error getting latest release")

	// Header and response errors.
	ErrDecodePEMBlock               = errors.New("failed to decode PEM block containing public key")
	ErrMissingHeaderFileID          = errors.New("missing X-File-Id header")
	ErrMissingHeaderPublicKey       = errors.New("missing X-File-Public-Key header")
	ErrMissingHeaderTransactionID   = errors.New("missing X-Transaction-Id header")
	ErrMissingHeaderUploadID        = errors.New("missing X-Upload-Id header")
)