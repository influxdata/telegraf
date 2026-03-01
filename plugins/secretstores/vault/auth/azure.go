package auth

import (
	"context"
	"errors"
	"fmt"

	vault "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/azure"
)

type Azure struct {
	RoleName    string `toml:"role_name"`
	ResourceURL string `toml:"resource_url"`
}

// Validate checks if the provided configuration fields are valid
func (a *Azure) Validate() error {
	if a.RoleName == "" {
		return errors.New("azure role_name missing")
	}

	if a.ResourceURL == "" {
		a.ResourceURL = "https://management.azure.com/"
	}

	return nil
}

// Authenticate uses the provided configuration to authenticate to Vault
func (a *Azure) Authenticate(v *vault.Client) (*vault.Secret, error) {
	azureAuth, err := azure.NewAzureAuth(
		a.RoleName,
		azure.WithResource(a.ResourceURL),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Azure auth method: %w", err)
	}

	authInfo, err := v.Auth().Login(context.Background(), azureAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to Azure auth method: %w", err)
	}
	if authInfo == nil {
		return nil, errors.New("no auth info was returned after login")
	}

	return authInfo, nil
}
