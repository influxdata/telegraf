package models

import (
	"reflect"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
)

// logName returns the log-friendly name/type.
func logName(pluginType, name, alias string) string {
	if alias == "" {
		return pluginType + "." + name
	}
	return pluginType + "." + name + "::" + alias
}

func SetLoggerOnPlugin(i interface{}, logger telegraf.Logger) {
	valI := reflect.ValueOf(i)

	if valI.Type().Kind() != reflect.Ptr {
		valI = reflect.New(reflect.TypeOf(i))
	}

	field := valI.Elem().FieldByName("Log")
	if !field.IsValid() {
		return
	}

	if field.Type().String() != "telegraf.Logger" || !field.CanSet() {
		logger.Debugf(
			"Plugin %q defines a 'Log' field on its struct of an unexpected type %q. Expected telegraf.Logger",
			valI.Type().Name(), field.Type().String(),
		)
		return
	}

	field.Set(reflect.ValueOf(logger))
}

func SetStatisticsOnPlugin(plugin interface{}, logger telegraf.Logger, tags map[string]string) {
	// Find the statistics collector
	instance := reflect.Indirect(reflect.ValueOf(plugin))
	field := instance.FieldByName("Statistics")
	if !field.IsValid() {
		return
	}

	// Validate the type and make sure we can actually set the struct field
	if field.Type().String() != "*selfstat.Collector" || !field.CanSet() {
		logger.Debugf(
			"Plugin %q defines a 'Statistics' field on its struct of an unexpected type %q. Expected *selfstat.Collector",
			instance.Type().Name(), field.Type().String(),
		)
		return
	}

	// Create a new collector and set it
	collector := selfstat.NewCollector(tags)
	field.Set(reflect.ValueOf(collector))
}
