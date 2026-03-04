package auth

import (
	"context"
	"errors"
	"fmt"

	vault "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/kubernetes"

	"github.com/influxdata/telegraf/config"
)

type Kubernetes struct {
	RoleName            string        `toml:"role_name"`
	ServiceAccountToken config.Secret `toml:"service_account_token"`
}

// Init validates the auth method options and sets any necessary defaults
func (k *Kubernetes) Init() error {
	if k.RoleName == "" {
		return errors.New("kubernetes role_name missing")
	}
	if k.ServiceAccountToken.Empty() {
		return errors.New("kubernetes service_account_token missing")
	}
	return nil
}

// Authenticate uses the provided configuration to authenticate to Vault
func (k *Kubernetes) Authenticate(client *vault.Client) (*vault.Secret, error) {
	secret, err := k.ServiceAccountToken.Get()
	if err != nil {
		return nil, fmt.Errorf("getting secret failed: %w", err)
	}
	opt := kubernetes.WithServiceAccountToken(secret.String())
	defer secret.Destroy()

	kubernetesAuth, err := kubernetes.NewKubernetesAuth(k.RoleName, opt)
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
