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
	ErrInvalidToken = errors.New("invalid token")
	Now             func() time.Time // for testing purposes
)

func init() {
	Now = time.Now
}

func CheckExpiration(token string) (expired bool, err error) {
	_, expirationTime, err := Whoami(token)
	if err != nil {
		return
	}
	expired = Now().After(expirationTime)
	return
}

func Whoami(token string) (email string, expirationTime time.Time, err error) {
	const defaultTokenFieldsNumber = 3
	var tokenData struct {
		Email          string `json:"email"`
		ExpirationTime int64  `json:"exp"`
	}

	splitTokens := strings.Split(token, ".")
	if len(splitTokens) != defaultTokenFieldsNumber {
		err = ErrInvalidToken
		return
	}
	dataPayload := splitTokens[1]
	decodedDataPayload, err := base64.RawStdEncoding.DecodeString(dataPayload)
	if err != nil {
		err = fmt.Errorf("error decoding data payload: %w", err)
		return
	}
	err = json.Unmarshal(decodedDataPayload, &tokenData)
	if err != nil {
		err = fmt.Errorf("error unmarshalling data payload: %w", err)
		return
	}
	email = tokenData.Email
	expirationTime = time.Unix(tokenData.ExpirationTime, 0)
	return
}
