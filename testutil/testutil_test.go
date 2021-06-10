package testutil

import (
	"os"
	"testing"
)

func TestDockerHost(t *testing.T) {
	os.Unsetenv("DOCKER_HOST")

	host := GetLocalHost()

	if host != localhost {
		t.Fatalf("Host should be localhost when DOCKER_HOST is not set. Current value [%s]", host)
	}

	os.Setenv("DOCKER_HOST", "1.1.1.1")

	host = GetLocalHost()

	if host != "1.1.1.1" {
		t.Fatalf("Host should take DOCKER_HOST value when set. Current value is [%s] and DOCKER_HOST is [%s]", host, os.Getenv("DOCKER_HOST"))
	}

	os.Setenv("DOCKER_HOST", "tcp://1.1.1.1:8080")

	host = GetLocalHost()

	if host != "1.1.1.1" {
		t.Fatalf("Host should take DOCKER_HOST value when set. Current value is [%s] and DOCKER_HOST is [%s]", host, os.Getenv("DOCKER_HOST"))
	}
}
