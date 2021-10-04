package config

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/fatih/color"

	"github.com/influxdata/telegraf"
)

type Escalation int

const (
	None Escalation = iota
	Warn
	Error
)

const (
	pluginWarnNotice = "Deprecated plugin will be removed soon, please switch to a supported plugin!"
	optionWarnNotice = "Deprecated options will be removed with the next major version, please adapt your config!"
)

func (c *Config) handleDeprecation(name string, plugin interface{}) error {
	// First check if the whole plugin is deprecated
	if deprecatedPlugin, ok := plugin.(telegraf.PluginDeprecator); ok {
		since, notice := deprecatedPlugin.DeprecationNotice()
		switch c.getDeprecationEscalation(since) {
		case Warn:
			prefix := "W! " + color.YellowString("DeprecationWarning")
			printPluginDeprecationNotice(prefix, name, since, notice)
			// We will not check for any deprecated options as the whole plugin is deprecated anyway.
			return nil
		case Error:
			prefix := "E! " + color.RedString("DeprecationError")
			printPluginDeprecationNotice(prefix, name, since, notice)
			// We are past the grace period
			return fmt.Errorf("plugin deprecated")
		}
	}

	// Check for deprecated options
	deprecatedOptions := make([]string, 0)
	walkPluginStruct(reflect.ValueOf(plugin), func(field reflect.StructField, value reflect.Value) {
		// Try to report only those fields that are set
		if value.IsZero() {
			return
		}

		tags := strings.SplitN(field.Tag.Get("deprecated"), ";", 2)
		if len(tags) < 1 || tags[0] == "" {
			return
		}
		since := tags[0]
		notice := ""
		if len(tags) > 1 {
			notice = tags[1]
		}

		// Get the toml field name
		option := field.Tag.Get("toml")
		if option == "" {
			option = field.Name
		}

		switch c.getDeprecationEscalation(since) {
		case Warn:
			prefix := "W! " + color.YellowString("DeprecationWarning")
			printOptionDeprecationNotice(prefix, name, option, since, notice)
		case Error:
			prefix := "E! " + color.RedString("DeprecationError")
			printOptionDeprecationNotice(prefix, name, option, since, notice)
			deprecatedOptions = append(deprecatedOptions, option)
		}
	})

	if len(deprecatedOptions) > 0 {
		return fmt.Errorf("plugin options %q deprecated", strings.Join(deprecatedOptions, ","))
	}

	return nil
}

func (c *Config) getDeprecationEscalation(since string) Escalation {
	sinceMajor, sinceMinor := parseVersion(since)
	if c.versionMajor > sinceMajor {
		return Error
	}
	if c.versionMajor == sinceMajor && c.versionMinor >= sinceMinor {
		return Warn
	}

	return None
}

func printPluginDeprecationNotice(prefix, name, since, notice string) {
	if notice != "" {
		log.Printf("%s: Plugin %q deprecated since version %s: %s", prefix, name, since, notice)
	} else {
		log.Printf("%s: Plugin %q deprecated since version %s", prefix, name, since)
	}
	log.Printf("Please note: %s", pluginWarnNotice)
}

func printOptionDeprecationNotice(prefix, name, option, since, notice string) {
	if notice != "" {
		log.Printf("%s: Option %q of plugin %q deprecated since version %s: %s", prefix, option, name, since, notice)
	} else {
		log.Printf("%s: Option %q of plugin %q deprecated since version %s", prefix, option, name, since)
	}
	log.Printf("Please note: %s", optionWarnNotice)
}

func walkPluginStruct(value reflect.Value, fn func(f reflect.StructField, fv reflect.Value)) {
	v := reflect.Indirect(value)
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if field.PkgPath != "" {
			continue
		}
		switch field.Type.Kind() {
		case reflect.Struct:
			walkPluginStruct(fieldValue, fn)

		case reflect.Array, reflect.Slice:
			for j := 0; j < fieldValue.Len(); j++ {
				fn(field, fieldValue.Index(j))
			}
		case reflect.Map:
			iter := fieldValue.MapRange()
			for iter.Next() {
				fn(field, iter.Value())
			}
		}
		fn(field, fieldValue)
	}
}

func parseVersion(version string) (major, minor int) {
	parts := strings.SplitN(version, ".", 3)
	if len(parts) < 2 {
		panic(fmt.Errorf("insufficient version fields in %q", version))
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		panic(fmt.Errorf("invalid version major in %q", version))
	}

	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		panic(fmt.Errorf("invalid version major in %q", version))
	}
	return major, minor
}
