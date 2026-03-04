package auth

import vault "github.com/hashicorp/vault/api"

type VaultAuth interface {
	// Init validates the auth method options and sets any necessary defaults
	Init() error

	// Authenticate uses the provided configuration to authenticate to Vault
	Authenticate(*vault.Client) (*vault.Secret, error)
}
