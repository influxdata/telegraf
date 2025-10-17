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
		_ = container.Terminate(context.Background())
	}
}

func getRoleID(t *testing.T, container *vault.VaultContainer) string {
	t.Helper()

	exitCode, reader, err := container.Exec(context.Background(), []string{"vault", "read", "-format=raw", "auth/approle/role/my-role/role-id"})
	require.NoError(t, err)
	require.Zero(t, exitCode)

	output, err := io.ReadAll(reader)
	require.NoError(t, err)

	// Trim some junk characters in the response first. "raw" format may still contain some non-JSON prefix
	startIndex := bytes.IndexByte(output, byte('{'))

	var resp RoleIDResponse
	require.NoError(t, json.Unmarshal(output[startIndex:], &resp))
	return resp.Data.RoleID
}

type RoleIDResponse struct {
	Data struct {
		RoleID string `json:"role_id"`
	} `json:"data"`
}

func getSecretID(t *testing.T, container *vault.VaultContainer) string {
	t.Helper()

	exitCode, reader, err := container.Exec(context.Background(), []string{"vault", "write", "-f", "-format=json", "auth/approle/role/my-role/secret-id"})
	require.NoError(t, err)
	require.Zero(t, exitCode)

	output, err := io.ReadAll(reader)
	require.NoError(t, err)

	// Trim some junk characters in the response first. "json" format may still contain some non-JSON prefix
	startIndex := bytes.IndexByte(output, byte('{'))

	var resp SecretIDResponse
	require.NoError(t, json.Unmarshal(output[startIndex:], &resp))
	return resp.Data.SecretID
}

type SecretIDResponse struct {
	Data struct {
		SecretID string `json:"secret_id"`
	} `json:"data"`
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
		UseKVv1:    true,
		AppRole: &appRole{
			RoleID:   getRoleID(t, container),
			SecretID: getSecretID(t, container),
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
			RoleID:   getRoleID(t, container),
			SecretID: getSecretID(t, container),
		},
	}

	require.NoError(t, plugin.Init())

	secret, err := plugin.Get(secretName)
	require.NoError(t, err)
	require.Equal(t, []byte(secretValue), secret)
}
