//go:build windows
// +build windows

//go:generate go run ../../../tools/generate_plugindata/main.go
//go:generate go run ../../../tools/generate_plugindata/main.go --clean
// DON'T EDIT; This file is used as a template by tools/generate_plugindata
package win_perf_counters

func (m *Win_PerfCounters) SampleConfig() string {
	return `{{ .SampleConfig }}`
}
