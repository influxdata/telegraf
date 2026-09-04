package vault

import (
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

const (
	vaultImageV1 = "hashicorp/vault:1.20.4"
	vaultImageV2 = "hashicorp/vault:2.1"
)

func TestInitFail(t *testing.T) {
	tests := []struct {
		name     string
		token    config.Secret
		approle  *appRole
		expected string
	}{
		{
			name:     "no auth method",
			expected: "set either `token` or `approle`",
		},
		{
			name:  "both token and approle",
			token: config.NewSecret([]byte("some-token")),
			approle: &appRole{
				RoleID: "role",
				Secret: config.NewSecret([]byte("secret")),
			},
			expected: "only one authentication method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := Vault{
				ID:         "vault",
				Address:    "http://localhost:8200",
				MountPath:  "secret",
				SecretPath: "my/path",
				Token:      tt.token,
				AppRole:    tt.approle,
			}
			require.ErrorContains(t, plugin.Init(), tt.expected)
		})
	}
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name   string
		image  string
		engine string
	}{
		{
			name:   "v1.x with kv-v1",
			image:  vaultImageV1,
			engine: "kv-v1",
		},
		{
			name:   "v1.x with kv-v2",
			image:  vaultImageV1,
			engine: "kv-v2",
		},
		{
			name:   "v2.x with kv-v1",
			image:  vaultImageV2,
			engine: "kv-v1",
		},
		{
			name:   "v2.x with kv-v2",
			image:  vaultImageV2,
			engine: "kv-v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Commands for preparing the container
			commands := []string{
				// Enable authentication via approle
				"vault auth enable approle",
				"vault policy write my-policy /tmp/policy.hcl",
				"vault write auth/approle/role/my-role policies=my-policy",
				// Enable KV engine and add secrets
				"vault secrets enable -path=my-mount-path " + tt.engine,
				"vault kv put -mount=my-mount-path my-secret-path secret-some-name=secret-some-value",
			}

			// Setup the container
			policyPath, err := filepath.Abs("testdata/policy.hcl")
			require.NoError(t, err)
			container := &testutil.Container{
				Image:        tt.image,
				ExposedPorts: []string{"8200"},
				Env: map[string]string{
					"VAULT_ADDR":              "http://localhost:8200",
					"VAULT_DEV_ROOT_TOKEN_ID": "telegraf",
					"VAULT_TOKEN":             "telegraf",
				},
				Files: map[string]string{"/tmp/policy.hcl": policyPath},
				WaitingFor: wait.ForAll(
					wait.ForHTTP("/v1/sys/health").WithPort("8200"),
					wait.ForExec([]string{"/bin/sh", "-c", strings.Join(commands, " && ")}),
				),
			}
			require.NoError(t, container.Start(), "failed to start container")
			defer container.Terminate()

			addr := "http://" + container.Address + ":" + container.Ports["8200"]

			// Determine credentials from container
			buf, err := readInfo(container, []string{
				"vault", "read", "-field", "role_id", "auth/approle/role/my-role/role-id",
			})
			require.NoError(t, err)
			roleID := string(buf)
			buf, err = readInfo(container, []string{
				"vault", "write", "-field", "secret_id", "-force", "auth/approle/role/my-role/secret-id",
			})
			require.NoError(t, err)
			secretID := config.NewSecret(buf)
			defer secretID.Destroy()

			// Setup plugin
			plugin := &Vault{
				ID:         "test_integration_" + tt.engine,
				Address:    addr,
				MountPath:  "my-mount-path",
				SecretPath: "my-secret-path",
				Engine:     tt.engine,
				AppRole: &appRole{
					RoleID: roleID,
					Secret: secretID,
				},
			}
			require.NoError(t, plugin.Init())

			// Check if we can retrieve the secret
			secret, err := plugin.Get("secret-some-name")
			require.NoError(t, err)
			require.Equal(t, "secret-some-value", string(secret))
		})
	}
}

func TestIntegrationAppRoleSecretWrapped(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Commands for preparing the container
	commands := []string{
		// Enable authentication via approle
		"vault auth enable approle",
		"vault policy write my-policy /tmp/policy.hcl",
		"vault write auth/approle/role/my-role policies=my-policy",
		// Enable KV engine and add secrets
		"vault secrets enable -path=my-mount-path kv-v2",
		"vault kv put -mount=my-mount-path my-secret-path secret-some-name=secret-some-value",
	}

	// Setup the container
	policyPath, err := filepath.Abs("testdata/policy.hcl")
	require.NoError(t, err)
	container := &testutil.Container{
		Image:        vaultImageV1,
		ExposedPorts: []string{"8200"},
		Env: map[string]string{
			"VAULT_ADDR":              "http://localhost:8200",
			"VAULT_DEV_ROOT_TOKEN_ID": "telegraf",
			"VAULT_TOKEN":             "telegraf",
		},
		Files: map[string]string{"/tmp/policy.hcl": policyPath},
		WaitingFor: wait.ForAll(
			wait.ForHTTP("/v1/sys/health").WithPort("8200"),
			wait.ForExec([]string{"/bin/sh", "-c", strings.Join(commands, " && ")}),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	addr := "http://" + container.Address + ":" + container.Ports["8200"]

	// Determine credentials from container
	buf, err := readInfo(container, []string{
		"vault", "read", "-field", "role_id", "auth/approle/role/my-role/role-id",
	})
	require.NoError(t, err)
	roleID := string(buf)
	buf, err = readInfo(container, []string{
		"vault", "write", "-wrap-ttl=60s", "-force", "-field", "wrapping_token", "auth/approle/role/my-role/secret-id",
	})
	require.NoError(t, err)
	secretID := config.NewSecret(buf)
	defer secretID.Destroy()

	// Setup plugin
	plugin := &Vault{
		ID:         "test_integration_kv_v2",
		Address:    addr,
		MountPath:  "my-mount-path",
		SecretPath: "my-secret-path",
		AppRole: &appRole{
			RoleID:          roleID,
			Secret:          secretID,
			ResponseWrapped: true,
		},
	}
	require.NoError(t, plugin.Init())

	// Check if we can retrieve the secret
	secret, err := plugin.Get("secret-some-name")
	require.NoError(t, err)
	require.Equal(t, "secret-some-value", string(secret))
}

func TestIntegrationSetKeepsSiblings(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	for _, engine := range []string{"kv-v1", "kv-v2"} {
		t.Run(engine, func(t *testing.T) {
			// Commands for preparing the container
			commands := []string{
				// Enable KV engine and add secrets
				"vault secrets enable -path=my-mount-path " + engine,
				"vault kv put -mount=my-mount-path my-secret-path alpha=one beta=two",
			}

			// Setup the container
			container := &testutil.Container{
				Image:        vaultImageV1,
				ExposedPorts: []string{"8200"},
				Env: map[string]string{
					"VAULT_ADDR":              "http://localhost:8200",
					"VAULT_DEV_ROOT_TOKEN_ID": "telegraf",
					"VAULT_TOKEN":             "telegraf",
				},
				WaitingFor: wait.ForAll(
					wait.ForHTTP("/v1/sys/health").WithPort("8200"),
					wait.ForExec([]string{"/bin/sh", "-c", strings.Join(commands, " && ")}),
				),
			}
			require.NoError(t, container.Start(), "failed to start container")
			defer container.Terminate()

			addr := "http://" + container.Address + ":" + container.Ports["8200"]

			// Setup plugin
			plugin := &Vault{
				ID:         "test_" + engine,
				Address:    addr,
				MountPath:  "my-mount-path",
				SecretPath: "my-secret-path",
				Engine:     engine,
				Token:      config.NewSecret([]byte("telegraf")),
			}
			require.NoError(t, plugin.Init())

			// Create a new secret and check if it is there
			require.NoError(t, plugin.Set("gamma", "three"))

			keys, err := plugin.List()
			require.NoError(t, err)
			slices.Sort(keys)
			require.Equal(t, []string{"alpha", "beta", "gamma"}, keys)
		})
	}
}

func TestIntegrationRemove(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		engine string
		// Error reported by the engine for a path without any secret
		expected string
	}{
		{
			engine:   "kv-v1",
			expected: "secret not found",
		},
		{
			engine:   "kv-v2",
			expected: "no secret data found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.engine, func(t *testing.T) {
			// Commands for preparing the container
			commands := []string{
				// Enable KV engine and add secrets
				"vault secrets enable -path=my-mount-path " + tt.engine,
				"vault kv put -mount=my-mount-path my-secret-path alpha=one beta=two",
			}

			// Setup the container
			container := &testutil.Container{
				Image:        vaultImageV1,
				ExposedPorts: []string{"8200"},
				Env: map[string]string{
					"VAULT_ADDR":              "http://localhost:8200",
					"VAULT_DEV_ROOT_TOKEN_ID": "telegraf",
					"VAULT_TOKEN":             "telegraf",
				},
				WaitingFor: wait.ForAll(
					wait.ForHTTP("/v1/sys/health").WithPort("8200"),
					wait.ForExec([]string{"/bin/sh", "-c", strings.Join(commands, " && ")}),
				),
			}
			require.NoError(t, container.Start(), "failed to start container")
			defer container.Terminate()

			addr := "http://" + container.Address + ":" + container.Ports["8200"]

			// Setup plugin
			plugin := &Vault{
				ID:         "test_" + tt.engine,
				Address:    addr,
				MountPath:  "my-mount-path",
				SecretPath: "my-secret-path",
				Engine:     tt.engine,
				Token:      config.NewSecret([]byte("telegraf")),
			}
			require.NoError(t, plugin.Init())

			// Removing a secret must keep the sibling secrets at the same path
			require.NoError(t, plugin.Remove("beta"))

			keys, err := plugin.List()
			require.NoError(t, err)
			require.Equal(t, []string{"alpha"}, keys)

			value, err := plugin.Get("alpha")
			require.NoError(t, err)
			require.Equal(t, "one", string(value))

			// Removing a secret that does not exist must fail
			require.ErrorContains(t, plugin.Remove("beta"), `secret "beta" not found`)

			// Removing the last secret must remove the secret at the path completely
			require.NoError(t, plugin.Remove("alpha"))
			_, err = plugin.List()
			require.ErrorContains(t, err, tt.expected)
		})
	}
}

func TestIntegrationSetCreatesNewPath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	for _, engine := range []string{"kv-v1", "kv-v2"} {
		t.Run(engine, func(t *testing.T) {
			// Commands for preparing the container
			commands := []string{
				// Enable KV engine and add secrets
				"vault secrets enable -path=my-mount-path " + engine,
			}

			// Setup the container
			container := &testutil.Container{
				Image:        vaultImageV1,
				ExposedPorts: []string{"8200"},
				Env: map[string]string{
					"VAULT_ADDR":              "http://localhost:8200",
					"VAULT_DEV_ROOT_TOKEN_ID": "telegraf",
					"VAULT_TOKEN":             "telegraf",
				},
				WaitingFor: wait.ForAll(
					wait.ForHTTP("/v1/sys/health").WithPort("8200"),
					wait.ForExec([]string{"/bin/sh", "-c", strings.Join(commands, " && ")}),
				),
			}
			require.NoError(t, container.Start(), "failed to start container")
			defer container.Terminate()

			addr := "http://" + container.Address + ":" + container.Ports["8200"]

			// Setup plugin
			plugin := &Vault{
				ID:         "test_" + engine,
				Address:    addr,
				MountPath:  "my-mount-path",
				SecretPath: "my-secret-path",
				Engine:     engine,
				Token:      config.NewSecret([]byte("telegraf")),
			}
			require.NoError(t, plugin.Init())

			// Add a new secret
			require.NoError(t, plugin.Set("gamma", "three"))

			value, err := plugin.Get("gamma")
			require.NoError(t, err)
			require.Equal(t, "three", string(value))
		})
	}
}

func readInfo(container *testutil.Container, command []string) ([]byte, error) {
	rc, reader, err := container.Exec(command, exec.Multiplexed())
	if err != nil {
		return nil, fmt.Errorf("executing command failed: %w", err)
	}
	buf, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("reading output failed: %w", err)
	}
	if rc != 0 {
		return nil, fmt.Errorf("command returned %d: %s", rc, string(buf))
	}
	return buf, nil
}
