//go:generate go run ../../../tools/generate_plugindata/main.go
//go:generate go run ../../../tools/generate_plugindata/main.go --clean
// DON'T EDIT; This file is used as a template by tools/generate_plugindata
package nfsclient

func (n *NFSClient) SampleConfig() string {
	return `{{ .SampleConfig }}`
}
