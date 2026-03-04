package auth

import (
	"context"
	"errors"
	"fmt"

	vault "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"

	"github.com/influxdata/telegraf/config"
)

type AppRole struct {
	RoleID          string        `toml:"role_id"`
	ResponseWrapped bool          `toml:"response_wrapped"`
	Secret          config.Secret `toml:"secret"`
}

// Init validates the auth method options and sets any necessary defaults
func (a *AppRole) Init() error {
	if a.RoleID == "" {
		return errors.New("approle role_id missing")
	}
	if a.Secret.Empty() {
		return errors.New("approle secret missing")
	}
	return nil
}

// Authenticate uses the provided configuration to authenticate to Vault
func (a *AppRole) Authenticate(v *vault.Client) (*vault.Secret, error) {
	secret, err := a.Secret.Get()
	if err != nil {
		return nil, fmt.Errorf("getting secret failed: %w", err)
	}
	secretID := &approle.SecretID{FromString: secret.String()}
	defer secret.Destroy()

	var opts []approle.LoginOption
	if a.ResponseWrapped {
		opts = append(opts, approle.WithWrappingToken())
	}

	appRoleAuth, err := approle.NewAppRoleAuth(a.RoleID, secretID, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AppRole auth method: %w", err)
	}

	authInfo, err := v.Auth().Login(context.Background(), appRoleAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to AppRole auth method: %w", err)
	}
	if authInfo == nil {
		return nil, errors.New("no auth info was returned after login")
	}

	return authInfo, nil
}
