package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCLIReference(t *testing.T) {
	var output bytes.Buffer
	require.NoError(t, runApp([]string{"telegraf", "--generate-cli-docs"}, &output, nil, nil, nil))
	require.Contains(t, output.String(), "### telegraf config migrate")
	require.Contains(t, output.String(), "### telegraf service install")
	require.Contains(t, output.String(), "`--service-name`")

	reference, err := os.ReadFile("../../docs/COMMANDS_AND_FLAGS.md")
	require.NoError(t, err)
	require.Equal(t, string(reference), output.String(), "run make docs-cli to update the reference")
}
