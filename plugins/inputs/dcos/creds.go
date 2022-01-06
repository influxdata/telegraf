package dcos

import (
	"context"
	"crypto/rsa"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	// How long before expiration to renew token
	relogDuration = 5 * time.Minute
)

type Credentials interface {
	Token(ctx context.Context, client Client) (string, error)
	IsExpired() bool
}

type ServiceAccount struct {
	AccountID  string
	PrivateKey *rsa.PrivateKey

	auth *AuthToken
}

type TokenCreds struct {
	Path string
}

type NullCreds struct {
}

func (c *ServiceAccount) Token(ctx context.Context, client Client) (string, error) {
	auth, err := client.Login(ctx, c)
	if err != nil {
		return "", err
	}
	c.auth = auth
	return auth.Text, nil
}

func (c *ServiceAccount) IsExpired() bool {
	return c.auth.Text != "" || c.auth.Expire.Add(relogDuration).After(time.Now())
}

func (c *TokenCreds) Token(_ context.Context, _ Client) (string, error) {
	octets, err := os.ReadFile(c.Path)
	if err != nil {
		return "", fmt.Errorf("error reading token file %q: %s", c.Path, err)
	}
	if !utf8.Valid(octets) {
		return "", fmt.Errorf("token file does not contain utf-8 encoded text: %s", c.Path)
	}
	token := strings.TrimSpace(string(octets))
	return token, nil
}

func (c *TokenCreds) IsExpired() bool {
	return true
}

func (c *NullCreds) Token(_ context.Context, _ Client) (string, error) {
	return "", nil
}

func (c *NullCreds) IsExpired() bool {
	return true
}
