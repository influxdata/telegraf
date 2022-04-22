//go:generate go run ../../../tools/generate_plugindata/main.go
//go:generate go run ../../../tools/generate_plugindata/main.go --clean
// DON'T EDIT; This file is used as a template by tools/generate_plugindata
package kafka_consumer_legacy

func (k *Kafka) SampleConfig() string {
	return `{{ .SampleConfig }}`
}
