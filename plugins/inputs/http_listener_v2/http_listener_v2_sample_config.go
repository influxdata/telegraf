//go:generate go run ../../../tools/generate_plugindata/main.go
//go:generate go run ../../../tools/generate_plugindata/main.go --clean
// DON'T EDIT; This file is used as a template by tools/generate_plugindata
package http_listener_v2

func (h *HTTPListenerV2) SampleConfig() string {
	return `{{ .SampleConfig }}`
}
