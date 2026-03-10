package auth

import (
	"context"
	"errors"
	"fmt"

	vault "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/userpass"

	"github.com/influxdata/telegraf/config"
)

type UserPass struct {
	Username string        `toml:"username"`
	Password config.Secret `toml:"password"`
}

// Init validates the auth method options and sets any necessary defaults
func (u *UserPass) Init() error {
	if u.Username == "" {
		return errors.New("userpass username missing")
	}
	if u.Password.Empty() {
		return errors.New("userpass password missing")
	}
	return nil
}

// Authenticate uses the provided configuration to authenticate to Vault
func (u *UserPass) Authenticate(v *vault.Client) (*vault.Secret, error) {
	secret, err := u.Password.Get()
	if err != nil {
		return nil, fmt.Errorf("getting secret failed: %w", err)
	}
	password := &userpass.Password{FromString: secret.String()}
	defer secret.Destroy()

	userPassAuth, err := userpass.NewUserpassAuth(u.Username, password)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Userpass auth method: %w", err)
	}

	authInfo, err := v.Auth().Login(context.Background(), userPassAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to Userpass auth method: %w", err)
	}
	if authInfo == nil {
		return nil, errors.New("no auth info was returned after login")
	}

	return authInfo, nil
}
