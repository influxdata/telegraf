//go:generate go run ../../../tools/generate_plugindata/main.go
//go:generate go run ../../../tools/generate_plugindata/main.go --clean
// DON'T EDIT; This file is used as a template by tools/generate_plugindata
package tcp_listener

func (t *TCPListener) SampleConfig() string {
	return `{{ .SampleConfig }}`
}
