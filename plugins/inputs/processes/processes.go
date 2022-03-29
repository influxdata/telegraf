//go:generate go run ../../../tools/generate_plugindata/main.go
//go:generate go run ../../../tools/generate_plugindata/main.go --clean
package processes

func (p *Processes) SampleConfig() string {
	return `{{ .SampleConfig }}`
}
