package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindHash(t *testing.T) {
	tests := []struct {
		testFile      string
		version       string
		expectedHases map[string]string
	}{
		{
			"testdata/godev_patch.html",
			"1.19.2",
			map[string]string{
				"go1.19.2.linux-amd64.tar.gz":  "5e8c5a74fe6470dd7e055a461acda8bb4050ead8c2df70f227e3ff7d8eb7eeb6",
				"go1.19.2.darwin-arm64.tar.gz": "35d819df25197c0be45f36ce849b994bba3b0559b76d4538b910d28f6395c00d",
				"go1.19.2.darwin-amd64.tar.gz": "16f8047d7b627699b3773680098fbaf7cc962b7db02b3e02726f78c4db26dfde",
			},
		},
		{
			"testdata/godev_minor.html",
			"1.19",
			map[string]string{
				"go1.19.linux-amd64.tar.gz":  "464b6b66591f6cf055bc5df90a9750bf5fbc9d038722bb84a9d56a2bea974be6",
				"go1.19.darwin-arm64.tar.gz": "859e0a54b7fcea89d9dd1ec52aab415ac8f169999e5fdfb0f0c15b577c4ead5e",
				"go1.19.darwin-amd64.tar.gz": "df6509885f65f0d7a4eaf3dfbe7dda327569787e8a0a31cbf99ae3a6e23e9ea8",
			},
		},
		{
			"testdata/godev_minor.html",
			"1.19.0",
			map[string]string{
				"go1.19.linux-amd64.tar.gz":  "464b6b66591f6cf055bc5df90a9750bf5fbc9d038722bb84a9d56a2bea974be6",
				"go1.19.darwin-arm64.tar.gz": "859e0a54b7fcea89d9dd1ec52aab415ac8f169999e5fdfb0f0c15b577c4ead5e",
				"go1.19.darwin-amd64.tar.gz": "df6509885f65f0d7a4eaf3dfbe7dda327569787e8a0a31cbf99ae3a6e23e9ea8",
			},
		},
	}

	for _, test := range tests {
		b, err := os.ReadFile(test.testFile)
		require.NoError(t, err)

		hashes, err := findHashes(bytes.NewReader(b), test.version)
		require.NoError(t, err)

		require.Equal(t, test.expectedHases, hashes)
	}
}
