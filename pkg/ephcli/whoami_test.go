package ephcli_test

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWhoami(t *testing.T) {
	// Cannot run in parallel because subtests modify global ephcli.Now variable
	t.Run("valid token", func(t *testing.T) {
		expirationTime := strconv.FormatInt(time.Date(2021, 3, 22, 0, 0, 0, 0, time.UTC).Unix(), 10)
		ephcli.Now = func() time.Time {
			now := time.Date(2021, 3, 21, 0, 0, 0, 0, time.UTC) // 1 day before expiration
			return now
		}
		lastPartOfToken := base64.RawStdEncoding.EncodeToString(
			[]byte(`{"email":"test@test.com","exp":` + expirationTime + `}`))
		token := fmt.Sprintf("header.%s.signature", lastPartOfToken)

		result, err := ephcli.CheckExpiration(token)
		require.NoError(t, err)
		assert.False(t, result)
	})
	t.Run("expired token", func(t *testing.T) {
		expirationTime := strconv.FormatInt(time.Date(2021, 3, 22, 0, 0, 0, 0, time.UTC).Unix(), 10)
		ephcli.Now = func() time.Time {
			now := time.Date(2021, 3, 23, 0, 0, 0, 0, time.UTC) // 1 day after expiration
			return now
		}
		lastPartOfToken := base64.RawStdEncoding.EncodeToString(
			[]byte(`{"email":"test@test.com","exp":` + expirationTime + `}`))
		token := fmt.Sprintf("header.%s.signature", lastPartOfToken)
		result, err := ephcli.CheckExpiration(token)
		require.NoError(t, err)
		assert.True(t, result)
	})
	t.Run("invalid payload", func(t *testing.T) {
		t.Parallel()
		lastPartOfToken := base64.RawStdEncoding.EncodeToString([]byte(`{"`))
		token := fmt.Sprintf("header.%s.signature", lastPartOfToken)
		_, err := ephcli.CheckExpiration(token)
		require.Error(t, err)
	})
	t.Run("invalid token", func(t *testing.T) {
		t.Parallel()
		_, err := ephcli.CheckExpiration("wrongTokenFormat")
		require.Error(t, err)
	})
}
