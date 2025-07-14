package ephcli

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrInvalidToken is returned when a JWT token is invalid or malformed.
	ErrInvalidToken = errors.New("invalid token")
	// Now is a function that returns the current time, exposed for testing purposes.
	Now             func() time.Time
)

func init() {
	Now = time.Now
}

// CheckExpiration checks if a JWT token has expired.
func CheckExpiration(token string) (bool, error) {
	_, expirationTime, err := Whoami(token)
	if err != nil {
		return false, err
	}
	return Now().After(expirationTime), nil
}

// Whoami extracts user information from a JWT token, returning email and expiration time.
func Whoami(token string) (string, time.Time, error) {
	const defaultTokenFieldsNumber = 3
	var tokenData struct {
		Email          string `json:"email"`
		ExpirationTime int64  `json:"exp"`
	}

	splitTokens := strings.Split(token, ".")
	if len(splitTokens) != defaultTokenFieldsNumber {
		return "", time.Time{}, ErrInvalidToken
	}
	dataPayload := splitTokens[1]
	decodedDataPayload, err := base64.RawStdEncoding.DecodeString(dataPayload)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error decoding data payload: %w", err)
	}
	err = json.Unmarshal(decodedDataPayload, &tokenData)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error unmarshalling data payload: %w", err)
	}
	return tokenData.Email, time.Unix(tokenData.ExpirationTime, 0), nil
}
