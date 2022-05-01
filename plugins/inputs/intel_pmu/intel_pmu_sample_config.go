//go:build linux && amd64
// +build linux,amd64

//go:generate go run ../../../tools/generate_plugindata/main.go
//go:generate go run ../../../tools/generate_plugindata/main.go --clean
// DON'T EDIT; This file is used as a template by tools/generate_plugindata
package intel_pmu

func (i *IntelPMU) SampleConfig() string {
	return `{{ .SampleConfig }}`
}
