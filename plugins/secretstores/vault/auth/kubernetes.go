package auth

import (
	"context"
	"errors"
	"fmt"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/kubernetes"

	"github.com/influxdata/telegraf/config"
)

type Kubernetes struct {
	RoleName string
	Secret   config.Secret
}

// Validate checks if the provided configuration fields are valid
func (k *Kubernetes) Validate() error {
	if k.RoleName == "" {
		return errors.New("kubernetes role_name missing")
	}
	if k.Secret.Empty() {
		return errors.New("kubernetes secret missing")
	}
	return nil
}

// Authenticate uses the provided configuration to authenticate to Vault
func (k *Kubernetes) Authenticate(client *vault.Client) (*vault.Secret, error) {
	secret, err := k.Secret.Get()
	if err != nil {
		return nil, fmt.Errorf("getting secret failed: %w", err)
	}
	opt := auth.WithServiceAccountToken(secret.String())
	defer secret.Destroy()

	kubernetesAuth, err := auth.NewKubernetesAuth(k.RoleName, opt)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Kubernetes auth method: %w", err)
	}

	authInfo, err := client.Auth().Login(context.Background(), kubernetesAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to Kubernetes auth method: %w", err)
	}
	if authInfo == nil {
		return nil, errors.New("no auth info was returned after login")
	}

	return authInfo, nil
}
