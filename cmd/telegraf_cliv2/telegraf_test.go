package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubcommandConfig(t *testing.T) {
	tests := []struct {
		name            string
		commands        []string
		expectedHeaders []string
		removedHeaders  []string
		expectedPlugins []string
		removedPlugins  []string
	}{
		{
			name:     "no filters",
			commands: []string{"config"},
			expectedHeaders: []string{
				outputHeader,
				inputHeader,
				aggregatorHeader,
				processorHeader,
				serviceInputHeader,
			},
		},
		{
			name:     "filter sections for inputs",
			commands: []string{"config", "--section-filter", "inputs"},
			expectedHeaders: []string{
				inputHeader,
			},
			removedHeaders: []string{
				outputHeader,
				aggregatorHeader,
				processorHeader,
			},
		},
		{
			name:     "filter sections for inputs,outputs",
			commands: []string{"config", "--section-filter", "inputs:outputs"},
			expectedHeaders: []string{
				inputHeader,
				outputHeader,
			},
			removedHeaders: []string{
				aggregatorHeader,
				processorHeader,
			},
		},
		{
			name:     "filter input plugins",
			commands: []string{"config", "--input-filter", "cpu:file"},
			expectedPlugins: []string{
				"[[inputs.cpu]]",
				"[[inputs.file]]",
			},
			removedPlugins: []string{
				"[[inputs.disk]]",
			},
		},
		{
			name:     "filter output plugins",
			commands: []string{"config", "--output-filter", "influxdb:http"},
			expectedPlugins: []string{
				"[[outputs.influxdb]]",
				"[[outputs.http]]",
			},
			removedPlugins: []string{
				"[[outputs.file]]",
			},
		},
		{
			name:     "filter processor plugins",
			commands: []string{"config", "--processor-filter", "date:enum"},
			expectedPlugins: []string{
				"[[processors.date]]",
				"[[processors.enum]]",
			},
			removedPlugins: []string{
				"[[processors.parser]]",
			},
		},
		{
			name:     "filter aggregator plugins",
			commands: []string{"config", "--aggregator-filter", "basicstats:starlark"},
			expectedPlugins: []string{
				"[[aggregators.basicstats]]",
				"[[aggregators.starlark]]",
			},
			removedPlugins: []string{
				"[[aggregators.minmax]]",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			args := os.Args[0:1]
			args = append(args, test.commands...)
			err := runApp(args, buf)
			require.NoError(t, err)
			output := buf.String()
			for _, e := range test.expectedHeaders {
				require.Contains(t, output, e, "expected header not found")
			}
			for _, r := range test.removedHeaders {
				require.NotContains(t, output, r, "removed header found")
			}
			for _, e := range test.expectedPlugins {
				require.Contains(t, output, e, "expected plugin not found")
			}
			for _, r := range test.removedPlugins {
				require.NotContains(t, output, r, "removed plugin found")
			}
		})
	}
}
