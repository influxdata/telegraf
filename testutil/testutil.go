package testutil

import "os"

var localhost = "localhost"

func GetLocalHost() string {
	if dockerHostVar := os.Getenv("DOCKER_HOST"); dockerHostVar != "" {
		return dockerHostVar
	}
	return localhost
}
