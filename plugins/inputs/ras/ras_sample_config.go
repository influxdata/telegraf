//go:build linux && (386 || amd64 || arm || arm64)
// +build linux
// +build 386 amd64 arm arm64

//go:generate go run ../../../tools/generate_plugindata/main.go
//go:generate go run ../../../tools/generate_plugindata/main.go --clean
// DON'T EDIT; This file is used as a template by tools/generate_plugindata
package ras

func (r *Ras) SampleConfig() string {
	return `{{ .SampleConfig }}`
}
