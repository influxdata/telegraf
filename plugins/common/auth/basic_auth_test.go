package auth

import (
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
)

func TestBasicAuth_VerifyWithCredentials(t *testing.T) {
	auth := BasicAuth{"username", "password"}

	r := httptest.NewRequest("GET", "/github", nil)
	r.SetBasicAuth(auth.Username, auth.Password)

	require.True(t, auth.Verify(r))
}

func TestBasicAuth_VerifyWithoutCredentials(t *testing.T) {
	auth := BasicAuth{}

	r := httptest.NewRequest("GET", "/github", nil)

	require.True(t, auth.Verify(r))
}

func TestBasicAuth_VerifyWithInvalidCredentials(t *testing.T) {
	auth := BasicAuth{"username", "password"}

	r := httptest.NewRequest("GET", "/github", nil)
	r.SetBasicAuth("wrong-username", "wrong-password")

	require.False(t, auth.Verify(r))
}
