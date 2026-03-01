package auth

import vault "github.com/hashicorp/vault/api"

type VaultAuth interface {

	// Validate checks if the provided configuration fields are valid
	Validate() error

	// Authenticate uses the provided configuration to authenticate to Vault
	Authenticate(*vault.Client) (*vault.Secret, error)
}
