package auth

import (
	"crypto/subtle"
	"net/http"
)

type BasicAuth struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
}

func (b *BasicAuth) Verify(r *http.Request) bool {
	if b.Username == "" && b.Password == "" {
		return true
	}

	username, password, ok := r.BasicAuth()

	usernameComparison := subtle.ConstantTimeCompare([]byte(username), []byte(b.Username)) == 1
	passwordComparison := subtle.ConstantTimeCompare([]byte(password), []byte(b.Password)) == 1
	return ok && usernameComparison && passwordComparison
}
