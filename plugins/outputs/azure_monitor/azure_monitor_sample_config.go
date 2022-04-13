//go:generate go run ../../../tools/generate_plugindata/main.go
//go:generate go run ../../../tools/generate_plugindata/main.go --clean
// DON'T EDIT; This file is used as a template by tools/generate_plugindata
package azure_monitor

func (a *AzureMonitor) SampleConfig() string {
	return `{{ .SampleConfig }}`
}
