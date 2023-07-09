package jwt

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestInit checks the initialization of the JWTGenerator.
func TestInit(t *testing.T) {
	j := &JWTGenerator{}

	err := j.Init()

	assert.Equal(t, err.Error(), "id missing")
}

// TestGet verifies the retrieval of JWT tokens for a specified user.
func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		token := Token{AccessToken: "access", RefreshToken: "refresh"}
		err := json.NewEncoder(rw).Encode(token)
		assert.NoError(t, err)
	}))
	defer server.Close()

	j := &JWTGenerator{
		Users:     []string{"testuser"},
		Urls:      []string{server.URL},
		Passwords: []string{"testpassword"},
		Tokens:    make(map[string]string),
	}

	tokenJSON, err := j.Get("testuser")
	assert.NoError(t, err)

	var token Token
	err = json.NewDecoder(bytes.NewReader(tokenJSON)).Decode(&token)
	assert.NoError(t, err)

	assert.Equal(t, "access", token.AccessToken)
	assert.Equal(t, "refresh", token.RefreshToken)
}

// TestSet confirms the assignment of a JWT token for a specific user.
func TestSet(t *testing.T) {
	j := &JWTGenerator{
		Tokens: make(map[string]string),
	}

	err := j.Set("testuser", "testtoken")
	assert.NoError(t, err)

	token, exists := j.Tokens["testuser"]
	assert.True(t, exists)
	assert.Equal(t, "testtoken", token)
}

// TestList checks the listing of users for whom the JWTGenerator can generate tokens.
func TestList(t *testing.T) {
	j := &JWTGenerator{
		Users: []string{"testuser1", "testuser2", "testuser3"},
	}

	users, err := j.List()
	assert.NoError(t, err)

	assert.Contains(t, users, "testuser1")
	assert.Contains(t, users, "testuser2")
	assert.Contains(t, users, "testuser3")
}

// TestGetResolver confirms the retrieval of a resolver function for a particular user.
func TestGetResolver(t *testing.T) {
	j := &JWTGenerator{
		Dynamic: true,
		Tokens:  make(map[string]string),
	}

	_, err := j.GetResolver("test")
	assert.NoError(t, err)
}
