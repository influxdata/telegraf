//go:build windows

package app

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWindowsFlagsAreSet(t *testing.T) {
	expectedString := "test"

	commands := []string{
		"--service", expectedString,
		"--service-name", expectedString,
		"--service-display-name", expectedString,
		"--service-restart-delay", expectedString,
		"--service-auto-restart",
		"--console",
	}

	buf := new(bytes.Buffer)
	args := os.Args[0:1]
	args = append(args, commands...)
	m := NewMockTelegraf()
	runner := NewRunner(
		WithArgs(args),
		WithOutputWriter(buf),
		WithPProfServer(NewMockServer()),
		WithTelegrafConfig(NewMockConfig(buf)),
		WithTelegrafApp(m),
	)
	err := runner.RunApp()
	require.NoError(t, err)

	require.Equal(t, expectedString, m.service)
	require.Equal(t, expectedString, m.serviceName)
	require.Equal(t, expectedString, m.serviceDisplayName)
	require.Equal(t, expectedString, m.serviceRestartDelay)
	require.Equal(t, true, m.serviceAutoRestart)
	require.Equal(t, true, m.console)
}
