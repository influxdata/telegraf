package jwt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/secretstores"
	"io"
	"net/http"
	"net/url"
)

// JWTGenerator is a struct to handle JWT generation and storage for different users.
type JWTGenerator struct {
	ID        string            `toml:"id"`        // ID of the JWTGenerator
	Dynamic   bool              `toml:"dynamic"`   // Flag to indicate if dynamic resolution is enabled
	Users     []string          `toml:"users"`     // List of users for which JWT tokens can be generated
	Urls      []string          `toml:"urls"`      // List of URLs where JWT tokens can be generated
	Passwords []string          `toml:"passwords"` // List of passwords for the users
	Tokens    map[string]string // Store for JWT tokens indexed by user
}

// Token is a struct to handle JWT access and refresh tokens.
type Token struct {
	RefreshToken string `json:"refresh"`
	AccessToken  string `json:"access"`
}

// Credentials is a struct to handle username and password for JWT token generation.
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Init validates the JWTGenerator structure and returns an error if validation fails.
func (j *JWTGenerator) Init() error {
	if j.ID == "" {
		return errors.New("id missing")
	}

	return nil
}

// SampleConfig returns a string with sample configuration for the JWTGenerator.
func (j *JWTGenerator) SampleConfig() string {
	return `
    # Configuration for JWTGenerator
    # id = "jwt"
    # dynamic = false
    # users = ["user1", "user2", "user3"]
    # urls = ["url1", "url2", "url3"]
    # passwords = ["password1", "password2", "password3"]
    `
}

// Get retrieves a JWT token for the provided key (user).
func (j *JWTGenerator) Get(key string) ([]byte, error) {
	for i, user := range j.Users {
		if user == key {
			token, err := j.getJWTToken(user, j.Urls[i], j.Passwords[i])
			if err != nil {
				return nil, fmt.Errorf("error getting JWT token: %w", err)
			}
			tokenJSON, err := json.Marshal(token)
			if err != nil {
				return nil, fmt.Errorf("error marshalling JWT token to JSON: %w", err)
			}
			return tokenJSON, nil
		}
	}

	return nil, errors.New("user not found")
}

// Set sets a JWT token for a specific user.
func (j *JWTGenerator) Set(key string, value string) error {
	if key == "" || value == "" {
		return errors.New("key and value cannot be empty")
	}

	j.Tokens[key] = value

	return nil
}

// List returns a list of users that JWTGenerator can generate tokens for.
func (j *JWTGenerator) List() ([]string, error) {
	return j.Users, nil
}

// GetResolver returns a resolver function to retrieve JWT tokens for a specific user.
func (j *JWTGenerator) GetResolver(key string) (telegraf.ResolveFunc, error) {
	resolver := func() ([]byte, bool, error) {
		s, err := j.Get(key)
		return s, j.Dynamic, err
	}
	return resolver, nil
}

// getJWTToken generates and retrieves a JWT token for a specific user from a specific URL.
func (j *JWTGenerator) getJWTToken(username, urlStr, password string) (*Token, error) {
	// Validate URL
	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	data := map[string]string{"username": username, "password": password}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshalling JSON: %w", err)
	}

	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(jsonData)) // The provided URL should be secure and trusted.
	if err != nil {
		return nil, fmt.Errorf("error making POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error generating token: %s", string(bodyBytes))
	}

	var token Token
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &token, nil
}

// init registers the JWTGenerator with the SecretStore.
func init() {
	secretstores.Add("jwt", func(id string) telegraf.SecretStore {
		return &JWTGenerator{ID: id, Tokens: make(map[string]string)}
	})
}
