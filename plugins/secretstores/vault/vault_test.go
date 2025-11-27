package vault

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/vault"

	"github.com/influxdata/telegraf/config"
)

func createContainer(t *testing.T, initCommands []string) (*vault.VaultContainer, func()) {
	// Create container with a default AppRole setup
	initCommands = append([]string{
		"auth enable approle",
		"policy write my-policy /tmp/policy.hcl",
		"write auth/approle/role/my-role policies=my-policy",
	}, initCommands...)

	policyPath, err := filepath.Abs("testdata/policy.hcl")
	require.NoError(t, err)

	container, err := vault.Run(
		context.Background(),
		"hashicorp/vault:1.20.4",
		testcontainers.WithFiles(testcontainers.ContainerFile{
			HostFilePath:      policyPath,
			ContainerFilePath: "/tmp/policy.hcl",
			FileMode:          0644,
		}),
		vault.WithToken("SomeToken"),
		vault.WithInitCommand(initCommands...),
	)

	require.NoError(t, err)

	state, err := container.State(context.Background())
	require.NoError(t, err)
	require.True(t, state.Running)

	return container, func() {
		//nolint:errcheck // No need to check error on cleanup
		_ = container.Terminate(context.Background())
	}
}

func getRoleID(t *testing.T, container *vault.VaultContainer) string {
	t.Helper()

	exitCode, reader, err := container.Exec(context.Background(), []string{
		"vault", "read", "-format=raw", "auth/approle/role/my-role/role-id",
	})
	require.NoError(t, err)
	require.Zero(t, exitCode)

	output, err := io.ReadAll(reader)
	require.NoError(t, err)

	var resp roleIDResponse
	require.NoError(t, json.Unmarshal(sanitizeVaultResponse(t, output), &resp))
	return resp.Data.RoleID
}

type roleIDResponse struct {
	Data struct {
		RoleID string `json:"role_id"`
	} `json:"data"`
}

func getSecretID(t *testing.T, container *vault.VaultContainer) string {
	t.Helper()

	exitCode, reader, err := container.Exec(context.Background(), []string{
		"vault", "write", "-f", "-format=json", "auth/approle/role/my-role/secret-id",
	})
	require.NoError(t, err)
	require.Zero(t, exitCode)

	output, err := io.ReadAll(reader)
	require.NoError(t, err)

	var resp SecretIDResponse
	require.NoError(t, json.Unmarshal(sanitizeVaultResponse(t, output), &resp))
	return resp.Data.SecretID
}

type SecretIDResponse struct {
	Data struct {
		SecretID string `json:"secret_id"`
	} `json:"data"`
}

func getWrappedSecretID(t *testing.T, container *vault.VaultContainer) string {
	t.Helper()

	exitCode, reader, err := container.Exec(context.Background(), []string{
		"vault", "write", "-wrap-ttl=60s", "-f", "-format=json", "auth/approle/role/my-role/secret-id",
	})
	require.NoError(t, err)
	require.Zero(t, exitCode)

	output, err := io.ReadAll(reader)
	require.NoError(t, err)

	var resp WrappedSecretIDResponse
	require.NoError(t, json.Unmarshal(sanitizeVaultResponse(t, output), &resp))
	return resp.WrapInfo.Token
}

type WrappedSecretIDResponse struct {
	WrapInfo struct {
		Token string `json:"token"`
	} `json:"wrap_info"`
}

func sanitizeVaultResponse(t *testing.T, resp []byte) []byte {
	t.Helper()

	// Trim some junk characters in the response first.
	// "json"/"raw" format may still contain some non-JSON prefix/suffix
	startIndex := bytes.IndexByte(resp, byte('{'))
	endIndex := bytes.LastIndexByte(resp, byte('}'))
	return resp[startIndex : endIndex+1]
}

func TestIntegrationKVv1(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mountPath := "my-mount-path"
	secretPath := "my-secret-path"
	secretName := "secret-some-name"
	secretValue := "secret-some-value"

	container, closer := createContainer(t, []string{
		fmt.Sprintf("secrets enable -path=%s kv-v1", mountPath),
		fmt.Sprintf("kv put -mount=%s %s %s=%s", mountPath, secretPath, secretName, secretValue),
	})
	defer closer()

	addr, err := container.HttpHostAddress(context.Background())
	require.NoError(t, err)

	plugin := &Vault{
		ID:         "test_integration_kv_v1",
		Address:    addr,
		MountPath:  mountPath,
		SecretPath: secretPath,
		Engine:     "kv-v1",
		AppRole: &appRole{
			RoleID: getRoleID(t, container),
			Secret: config.NewSecret([]byte(getSecretID(t, container))),
		},
	}

	require.NoError(t, plugin.Init())

	secret, err := plugin.Get(secretName)
	require.NoError(t, err)
	require.Equal(t, []byte(secretValue), secret)
}

func TestIntegrationKVv2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mountPath := "my-mount-path"
	secretPath := "my-secret-path"
	secretName := "secret-some-name"
	secretValue := "secret-some-value"

	container, closer := createContainer(t, []string{
		fmt.Sprintf("secrets enable -path=%s kv-v2", mountPath),
		fmt.Sprintf("kv put -mount=%s %s %s=%s", mountPath, secretPath, secretName, secretValue),
	})
	defer closer()

	addr, err := container.HttpHostAddress(context.Background())
	require.NoError(t, err)

	plugin := &Vault{
		ID:         "test_integration_kv_v2",
		Address:    addr,
		MountPath:  mountPath,
		SecretPath: secretPath,
		AppRole: &appRole{
			RoleID: getRoleID(t, container),
			Secret: config.NewSecret([]byte(getSecretID(t, container))),
		},
	}

	require.NoError(t, plugin.Init())

	secret, err := plugin.Get(secretName)
	require.NoError(t, err)
	require.Equal(t, []byte(secretValue), secret)
}

func TestIntegrationAppRoleSecretWrapped(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mountPath := "my-mount-path"
	secretPath := "my-secret-path"
	secretName := "secret-some-name"
	secretValue := "secret-some-value"

	container, closer := createContainer(t, []string{
		fmt.Sprintf("secrets enable -path=%s kv-v2", mountPath),
		fmt.Sprintf("kv put -mount=%s %s %s=%s", mountPath, secretPath, secretName, secretValue),
	})
	defer closer()

	addr, err := container.HttpHostAddress(context.Background())
	require.NoError(t, err)

	plugin := &Vault{
		ID:         "test_integration_kv_v2",
		Address:    addr,
		MountPath:  mountPath,
		SecretPath: secretPath,
		AppRole: &appRole{
			RoleID:          getRoleID(t, container),
			Secret:          config.NewSecret([]byte(getWrappedSecretID(t, container))),
			ResponseWrapped: true,
		},
	}

	require.NoError(t, plugin.Init())

	secret, err := plugin.Get(secretName)
	require.NoError(t, err)
	require.Equal(t, []byte(secretValue), secret)
}
