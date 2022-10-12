package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindHash(t *testing.T) {
	b, err := os.ReadFile("testdata/godev.html")
	require.NoError(t, err)

	expectedVersion := "1.19.2"

	hashes, err := findHashes(bytes.NewReader(b), expectedVersion)
	require.NoError(t, err)

	expected := map[string]string{
		"go1.19.2.linux-amd64.tar.gz":  "5e8c5a74fe6470dd7e055a461acda8bb4050ead8c2df70f227e3ff7d8eb7eeb6",
		"go1.19.2.darwin-arm64.tar.gz": "35d819df25197c0be45f36ce849b994bba3b0559b76d4538b910d28f6395c00d",
		"go1.19.2.darwin-amd64.tar.gz": "16f8047d7b627699b3773680098fbaf7cc962b7db02b3e02726f78c4db26dfde",
	}

	require.Equal(t, expected, hashes)
}
