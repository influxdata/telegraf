//go:build linux

package main

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/internal"
)

func TestGetConfigFiles(t *testing.T) {
	// Drop privileges as user root as root can access any file despite
	// restricted permissions
	if os.Geteuid() == 0 {
		if runtime.GOOS != "linux" {
			t.Skip("Dropping privileges is only supported on Linux")
		}
		u, err := user.Lookup("nobody")
		if u == nil || err != nil {
			t.Skip("Skipping as user 'nobody' is not known!")
		}
		uid, err := strconv.Atoi(u.Uid)
		require.NoError(t, err)
		require.NoError(t, syscall.Setfsuid(uid))
		defer func() {
			//nolint:errcheck // We cannot do anything if this call fails
			syscall.Setfsuid(0)
		}()
	}

	// Create a test case where we create a temporary configuration file
	// structure with one directory having insufficient permissions.
	root := t.TempDir()

	// Top-level configuration file containing a minimal agent section
	content := `
[agent]
  interval = "10s"
`
	require.NoError(t, os.WriteFile(filepath.Join(root, "telegraf.conf"), []byte(content), 0600))

	// Create a configuration directory containing two files
	require.NoError(t, os.Mkdir(filepath.Join(root, "telegraf.d"), 0700))
	content = `
[[inputs.cpu]]
`
	require.NoError(t, os.WriteFile(filepath.Join(root, "telegraf.d", "inputs.conf"), []byte(content), 0600))
	content = `
[[outputs.discard]]
`
	require.NoError(t, os.WriteFile(filepath.Join(root, "telegraf.d", "outputs.conf"), []byte(content), 0600))

	// Setup Telegraf
	agent := &Telegraf{
		GlobalFlags: GlobalFlags{
			config:    []string{filepath.Join(root, "telegraf.conf")},
			configDir: []string{filepath.Join(root, "telegraf.d")},
		},
	}

	// Fill the configuration files and check the result
	expected := []string{
		filepath.Join(root, "telegraf.conf"),
		filepath.Join(root, "telegraf.d", "inputs.conf"),
		filepath.Join(root, "telegraf.d", "outputs.conf"),
	}
	require.NoError(t, agent.getConfigFiles())
	require.ElementsMatch(t, expected, agent.configFiles)
}

func TestLoadConfigurationsPermissions(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("Dropping privileges is only supported on Linux")
	}

	tests := []struct {
		name        string
		permissions map[string]os.FileMode
		expected    []string
		expectedErr string
	}{
		{
			name: "agent -w----",
			permissions: map[string]os.FileMode{
				"telegraf.conf": 0200,
			},
			expectedErr: "telegraf.conf: permission denied",
		},
		{
			name: "inputs -w----",
			permissions: map[string]os.FileMode{
				filepath.Join("telegraf.d", "inputs.conf"): 0200,
			},
			expectedErr: "inputs.conf: permission denied",
		},
		{
			name: "outputs -w----",
			permissions: map[string]os.FileMode{
				filepath.Join("telegraf.d", "outputs.conf"): 0200,
			},
			expectedErr: "outputs.conf: permission denied",
		},
		{
			name: "telegraf.d  -w----",
			permissions: map[string]os.FileMode{
				"telegraf.d": 0200,
			},
			expectedErr: "telegraf.d: permission denied",
		},
		{
			name: "telegraf.d  -wx---",
			permissions: map[string]os.FileMode{
				"telegraf.d": 0300,
			},
			expectedErr: "telegraf.d: permission denied",
		},
		{
			name: "telegraf.d  rw----",
			permissions: map[string]os.FileMode{
				"telegraf.d": 0400,
			},
			expected: []string{
				"telegraf.conf",
			},
		},
		{
			name: "telegraf.d rwx---",
			permissions: map[string]os.FileMode{
				"telegraf.d": 0700,
			},
			expected: []string{
				"telegraf.conf",
				filepath.Join("telegraf.d", "inputs.conf"),
				filepath.Join("telegraf.d", "outputs.conf"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Drop privileges as user root as root can access any file despite
			// restricted permissions
			if os.Geteuid() == 0 {
				u, err := user.Lookup("nobody")
				if u == nil || err != nil {
					t.Skip("Skipping as user 'nobody' is not known!")
				}
				uid, err := strconv.Atoi(u.Uid)
				require.NoError(t, err)
				require.NoError(t, syscall.Setfsuid(uid))
				defer func() {
					//nolint:errcheck // We cannot do anything if this call fails
					syscall.Setfsuid(0)
				}()
			}

			// Create a test case where we create a temporary configuration file
			// structure with one directory having insufficient permissions.
			root := t.TempDir()

			// Top-level configuration file containing a minimal agent section
			content := `
[agent]
  interval = "10s"
`
			require.NoError(t, os.WriteFile(filepath.Join(root, "telegraf.conf"), []byte(content), 0600))

			// Create a configuration directory containing two files
			require.NoError(t, os.Mkdir(filepath.Join(root, "telegraf.d"), 0700))
			content = `
[[inputs.cpu]]
`
			require.NoError(t, os.WriteFile(filepath.Join(root, "telegraf.d", "inputs.conf"), []byte(content), 0600))
			content = `
[[outputs.discard]]
`
			require.NoError(t, os.WriteFile(filepath.Join(root, "telegraf.d", "outputs.conf"), []byte(content), 0600))

			// Apply the permissions
			for entry, perm := range tt.permissions {
				require.NoError(t, os.Chmod(filepath.Join(root, entry), perm))
			}
			// Make sure we can delete the directory
			defer func() {
				//nolint:gosec // It's okay to set permissions to rwx------ for the temporary dir
				require.NoError(t, os.Chmod(filepath.Join(root, "telegraf.d"), 0700))
			}()

			// Define a version to prevent Telegraf startup failure
			savedVersion := internal.Version
			internal.Version = "0.0.0"
			defer func() {
				internal.Version = savedVersion
			}()

			// Setup Telegraf
			agent := &Telegraf{
				GlobalFlags: GlobalFlags{
					config:    []string{filepath.Join(root, "telegraf.conf")},
					configDir: []string{filepath.Join(root, "telegraf.d")},
				},
			}

			// Fill the configuration files and check the resulting error
			_, err := agent.loadConfiguration()
			if tt.expectedErr != "" {
				require.ErrorContains(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				expectedFiles := make([]string, 0, len(tt.expected))
				for _, p := range tt.expected {
					expectedFiles = append(expectedFiles, filepath.Join(root, p))
				}
				require.ElementsMatch(t, expectedFiles, agent.configFiles)
			}
		})
	}
}
