package auth

import (
	"crypto/subtle"
	"net/http"
)

type BasicAuth struct {
	Username string
	Password string
}

func (b *BasicAuth) Verify(r *http.Request) bool {
	if b.Username == "" && b.Password == "" {
		return true
	}

	username, password, ok := r.BasicAuth()
	return ok &&
		subtle.ConstantTimeCompare([]byte(username), []byte(b.Username)) == 1 &&
		subtle.ConstantTimeCompare([]byte(password), []byte(b.Password)) == 1
}
