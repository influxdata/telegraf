//go:generate go run ../../../scripts/generate_plugindata/main.go
//go:generate go run ../../../scripts/generate_plugindata/main.go --clean
package processes

func (p *Processes) SampleConfig() string {
	return `{{ .SampleConfig }}`
}
