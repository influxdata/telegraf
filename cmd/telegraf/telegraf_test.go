package main

import "testing"

func TestIsolatedPlugin(t *testing.T) {
	name := "cpu"
	configPath := "testconfig.conf"
	isolatedPlugin(name, configPath, 1)
}
