package config

import (
	"fmt"
	"log" //nolint:revive // log is ok here as the logging facility is not set-up yet
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/processors"
)

const (
	pluginWarnNotice = "Deprecated plugin will be removed soon, please switch to a supported plugin!"
	optionWarnNotice = "Deprecated options will be removed with the next major version, please adapt your config!"
)

// Escalation level for the plugin or option
type Escalation int

func (e Escalation) String() string {
	switch e {
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	}
	return "NONE"
}

const (
	// None means no deprecation
	None Escalation = iota
	// Warn means deprecated but still within the grace period
	Warn
	// Error means deprecated and beyond grace period
	Error
)

// DeprecationInfo contains all important information to describe a deprecated entity
type DeprecationInfo struct {
	// Name of the plugin or plugin option
	Name string
	// Level of deprecation
	Level Escalation
	// Since which version the plugin or plugin option is deprecated
	Since string
	// Notice to the user about alternatives or further information
	Notice string
}

// PluginDeprecationInfo holds all information about a deprecated plugin or it's options
type PluginDeprecationInfo struct {
	DeprecationInfo

	// Options deprecated for this plugin
	Options []DeprecationInfo
}

func (c *Config) collectDeprecationInfo(name string, plugin interface{}, all bool) PluginDeprecationInfo {
	info := PluginDeprecationInfo{}
	info.Name = name

	// First check if the whole plugin is deprecated
	if deprecatedPlugin, ok := plugin.(telegraf.PluginDeprecator); ok {
		info.Since, info.Notice = deprecatedPlugin.DeprecationNotice()
		info.Level = c.getDeprecationEscalation(info.Since)
	}

	// Check for deprecated options
	walkPluginStruct(reflect.ValueOf(plugin), func(field reflect.StructField, value reflect.Value) {
		// Try to report only those fields that are set
		if !all && value.IsZero() {
			return
		}

		tags := strings.SplitN(field.Tag.Get("deprecated"), ";", 2)
		if len(tags) < 1 || tags[0] == "" {
			return
		}
		optionInfo := DeprecationInfo{
			Name:  field.Name,
			Since: tags[0],
			Level: c.getDeprecationEscalation(tags[0]),
		}

		if len(tags) > 1 {
			optionInfo.Notice = tags[1]
		}
		// Get the toml field name
		option := field.Tag.Get("toml")
		if option != "" {
			optionInfo.Name = option
		}
		info.Options = append(info.Options, optionInfo)
	})

	return info
}

func (c *Config) printUserDeprecation(name string, plugin interface{}) error {
	info := c.collectDeprecationInfo(name, plugin, false)

	switch info.Level {
	case Warn:
		prefix := "W! " + color.YellowString("DeprecationWarning")
		printPluginDeprecationNotice(prefix, info.Name, info.Since, info.Notice)
		// We will not check for any deprecated options as the whole plugin is deprecated anyway.
		return nil
	case Error:
		prefix := "E! " + color.RedString("DeprecationError")
		printPluginDeprecationNotice(prefix, info.Name, info.Since, info.Notice)
		// We are past the grace period
		return fmt.Errorf("plugin deprecated")
	}

	// Print deprecated options
	deprecatedOptions := make([]string, 0)
	for _, option := range info.Options {
		switch option.Level {
		case Warn:
			prefix := "W! " + color.YellowString("DeprecationWarning")
			printOptionDeprecationNotice(prefix, info.Name, option.Name, option.Since, option.Notice)
		case Error:
			prefix := "E! " + color.RedString("DeprecationError")
			printOptionDeprecationNotice(prefix, info.Name, option.Name, option.Since, option.Notice)
			deprecatedOptions = append(deprecatedOptions, option.Name)
		}
	}

	if len(deprecatedOptions) > 0 {
		return fmt.Errorf("plugin options %q deprecated", strings.Join(deprecatedOptions, ","))
	}

	return nil
}

func (c *Config) CollectDeprecationInfos() map[string][]PluginDeprecationInfo {
	infos := make(map[string][]PluginDeprecationInfo)

	infos["inputs"] = make([]PluginDeprecationInfo, 0)
	for name, creator := range inputs.Inputs {
		plugin := creator()
		info := c.collectDeprecationInfo(name, plugin, true)

		if info.Level != None || len(info.Options) > 0 {
			infos["inputs"] = append(infos["inputs"], info)
		}
	}

	infos["outputs"] = make([]PluginDeprecationInfo, 0)
	for name, creator := range outputs.Outputs {
		plugin := creator()
		info := c.collectDeprecationInfo(name, plugin, true)

		if info.Level != None || len(info.Options) > 0 {
			infos["outputs"] = append(infos["outputs"], info)
		}
	}

	infos["processors"] = make([]PluginDeprecationInfo, 0)
	for name, creator := range processors.Processors {
		plugin := creator()
		info := c.collectDeprecationInfo(name, plugin, true)

		if info.Level != None || len(info.Options) > 0 {
			infos["processors"] = append(infos["processors"], info)
		}
	}

	infos["aggregators"] = make([]PluginDeprecationInfo, 0)
	for name, creator := range aggregators.Aggregators {
		plugin := creator()
		info := c.collectDeprecationInfo(name, plugin, true)

		if info.Level != None || len(info.Options) > 0 {
			infos["aggregators"] = append(infos["aggregators"], info)
		}
	}

	return infos
}

func (c *Config) PrintDeprecationList(infos []PluginDeprecationInfo) {
	sort.Slice(infos, func(i, j int) bool { return infos[i].Name < infos[j].Name })

	for _, info := range infos {
		switch info.Level {
		case Warn, Error:
			_, _ = fmt.Printf("  %-40s %-5s since %5s %s\n", info.Name, info.Level, info.Since, info.Notice)
		}

		if len(info.Options) < 1 {
			continue
		}
		sort.Slice(info.Options, func(i, j int) bool { return info.Options[i].Name < info.Options[j].Name })
		for _, option := range info.Options {
			_, _ = fmt.Printf("  %-40s %-5s since %5s %s\n", info.Name+"/"+option.Name, option.Level, option.Since, option.Notice)
		}
	}
}

func (c *Config) getDeprecationEscalation(since string) Escalation {
	sinceMajor, sinceMinor := ParseVersion(since)
	if c.VersionMajor > sinceMajor {
		return Error
	}
	if c.VersionMajor == sinceMajor && c.VersionMinor >= sinceMinor {
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


// walkPluginStruct iterates over the fields of a structure in depth-first search (to cover nested structures)
// and calls the given function for every visited field.
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

func ParseVersion(version string) (major, minor int) {
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
