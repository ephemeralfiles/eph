package ephcli_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jwtPart1 struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type jwtPart2 struct {
	Email string `json:"email"`
	Exp   int64  `json:"exp"`
}

func genToken() string {
	ExpirationDate := time.Now().Add(time.Hour).Unix()
	part1 := jwtPart1{Alg: "HS256", Typ: "JWT"}
	part2 := jwtPart2{Email: "test@test.com", Exp: ExpirationDate}
	tokenPart1, _ := json.Marshal(part1) //nolint:errchkjson
	tokenPart2, _ := json.Marshal(part2) //nolint:errchkjson
	tokenPart1Base64 := base64.StdEncoding.EncodeToString(tokenPart1)
	tokenPart2Base64 := base64.StdEncoding.EncodeToString(tokenPart2)
	token := fmt.Sprintf("%s.%s.", tokenPart1Base64, tokenPart2Base64)

	return token
}

func TestGetBoxInfos(t *testing.T) {
	t.Parallel()
	t.Run("standard case, no error", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"capacity_mb": 100, "used_mb": 50, "remaining_mb": 50}`)
		}))
		defer ts.Close()

		client := ts.Client()

		e := ephcli.NewClient(genToken())
		e.SetHTTPClient(client)
		e.SetEndpoint(ts.URL)

		b, err := e.GetBoxInfos()
		require.NoError(t, err)
		assert.NotNil(t, b)
		assert.Equal(t, int64(100), b.CapacityMb)
		assert.Equal(t, int64(50), b.UsedMb)
		assert.Equal(t, int64(50), b.RemainingMb)
	})

	t.Run("case with no data", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{}`)
		}))
		defer ts.Close()

		client := ts.Client()

		e := ephcli.NewClient(genToken())
		e.SetHTTPClient(client)
		e.SetEndpoint(ts.URL)

		b, err := e.GetBoxInfos()
		require.NoError(t, err)
		assert.NotNil(t, b)
		assert.Equal(t, int64(0), b.CapacityMb)
		assert.Equal(t, int64(0), b.UsedMb)
		assert.Equal(t, int64(0), b.RemainingMb)
	})

	t.Run("case with error", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "{}")
		}))
		defer ts.Close()

		client := ts.Client()

		e := ephcli.NewClient(genToken())
		e.SetHTTPClient(client)
		e.SetEndpoint(ts.URL)

		_, err := e.GetBoxInfos()
		assert.Error(t, err)
	})

	t.Run("case with wrong data returned", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "lskdfkjbsd")
		}))
		defer ts.Close()

		client := ts.Client()

		e := ephcli.NewClient(genToken())
		e.SetHTTPClient(client)
		e.SetEndpoint(ts.URL)

		_, err := e.GetBoxInfos()
		require.Error(t, err)
	})

	t.Run("case with no web server to answer", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		client := ts.Client()
		ts.Close() // stop web server

		e := ephcli.NewClient(genToken())
		e.SetHTTPClient(client)
		e.SetEndpoint(ts.URL)

		_, err := e.GetBoxInfos()
		require.Error(t, err)
	})
}
