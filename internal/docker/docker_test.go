package docker_test

import (
	"testing"

	"github.com/influxdata/telegraf/internal/docker"
	"github.com/stretchr/testify/require"
)

func TestParseImage(t *testing.T) {
	tests := []struct {
		image         string
		parsedName    string
		parsedVersion string
	}{
		{
			image:         "postgres",
			parsedName:    "postgres",
			parsedVersion: "unknown",
		},
		{
			image:         "postgres:latest",
			parsedName:    "postgres",
			parsedVersion: "latest",
		},
		{
			image:         "coreos/etcd",
			parsedName:    "coreos/etcd",
			parsedVersion: "unknown",
		},
		{
			image:         "coreos/etcd:latest",
			parsedName:    "coreos/etcd",
			parsedVersion: "latest",
		},
		{
			image:         "quay.io/postgres",
			parsedName:    "quay.io/postgres",
			parsedVersion: "unknown",
		},
		{
			image:         "quay.io:4443/coreos/etcd",
			parsedName:    "quay.io:4443/coreos/etcd",
			parsedVersion: "unknown",
		},
		{
			image:         "quay.io:4443/coreos/etcd:latest",
			parsedName:    "quay.io:4443/coreos/etcd",
			parsedVersion: "latest",
		},
	}
	for _, tt := range tests {
		t.Run("parse name "+tt.image, func(t *testing.T) {
			imageName, imageVersion := docker.ParseImage(tt.image)
			require.Equal(t, tt.parsedName, imageName)
			require.Equal(t, tt.parsedVersion, imageVersion)
		})
	}
}
