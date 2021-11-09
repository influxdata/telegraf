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

// deprecationInfo contains all important information to describe a deprecated entity
type deprecationInfo struct {
	// Name of the plugin or plugin option
	Name string
	// Level of deprecation
	Level Escalation
	// Since which version the plugin or plugin option is deprecated
	Since string
	// RemovalIn is an optional field denoting in which version the plugin or plugin option is actually removed
	RemovalIn string
	// Notice to the user about alternatives or further information
	Notice string
}

func (di *deprecationInfo) determineEscalation(major, minor int) {
	var removalMajor, removalMinor int

	sinceMajor, sinceMinor := ParseVersion(di.Since)
	if di.RemovalIn != "" {
		removalMajor, removalMinor = ParseVersion(di.RemovalIn)
	} else {
		removalMajor, removalMinor = sinceMajor+1, 0
		di.RemovalIn = fmt.Sprintf("%d.%d", removalMajor, removalMinor)
	}

	di.Level = None
	if major > removalMajor || (major == removalMajor && minor >= removalMinor) {
		di.Level = Error
	} else if major > sinceMajor || (major == sinceMajor && minor >= sinceMinor) {
		di.Level = Warn
	}
}

func (di *deprecationInfo) printPluginDeprecationNotice(prefix string) {
	if di.Notice != "" {
		log.Printf(
			"%s: Plugin %q deprecated since version %s and will be removed in %s: %s",
			prefix, di.Name, di.Since, di.RemovalIn, di.Notice,
		)
	} else {
		log.Printf(
			"%s: Plugin %q deprecated since version %s and will be removed in %s",
			prefix, di.Name, di.Since, di.RemovalIn,
		)
	}
}

func (di *deprecationInfo) printOptionDeprecationNotice(prefix, plugin string) {
	if di.Notice != "" {
		log.Printf(
			"%s: Option %q of plugin %q deprecated since version %s and will be removed in %s: %s",
			prefix, di.Name, plugin, di.Since, di.RemovalIn, di.Notice,
		)
	} else {
		log.Printf(
			"%s: Option %q of plugin %q deprecated since version %s and will be removed in %s",
			prefix, di.Name, plugin, di.Since, di.RemovalIn,
		)
	}
}

// pluginDeprecationInfo holds all information about a deprecated plugin or it's options
type pluginDeprecationInfo struct {
	deprecationInfo

	// Options deprecated for this plugin
	Options []deprecationInfo
}

func (c *Config) incrementPluginDeprecations(category string) {
	newcounts := []int64{1, 0}
	if counts, found := c.Deprecations[category]; found {
		newcounts = []int64{counts[0] + 1, counts[1]}
	}
	c.Deprecations[category] = newcounts
}

func (c *Config) incrementPluginOptionDeprecations(category string) {
	newcounts := []int64{0, 1}
	if counts, found := c.Deprecations[category]; found {
		newcounts = []int64{counts[0], counts[1] + 1}
	}
	c.Deprecations[category] = newcounts
}

func (c *Config) collectDeprecationInfo(category, name string, plugin interface{}, all bool) pluginDeprecationInfo {
	info := pluginDeprecationInfo{}
	info.Name = category + "." + name

	// First check if the whole plugin is deprecated
	if deprecatedPlugin, ok := plugin.(telegraf.PluginDeprecator); ok {
		info.Since, info.RemovalIn, info.Notice = deprecatedPlugin.DeprecationNotice()
		info.determineEscalation(c.VersionMajor, c.VersionMinor)
		if info.Level != None {
			c.incrementPluginDeprecations(category)
		}
	}

	// Check for deprecated options
	walkPluginStruct(reflect.ValueOf(plugin), func(field reflect.StructField, value reflect.Value) {
		// Try to report only those fields that are set
		if !all && value.IsZero() {
			return
		}

		tags := strings.SplitN(field.Tag.Get("deprecated"), ";", 3)
		if len(tags) < 1 || tags[0] == "" {
			return
		}
		optionInfo := deprecationInfo{
			Name:  field.Name,
			Since: tags[0],
		}

		if optionInfo.Level != None {
			c.incrementPluginOptionDeprecations(category)
		}

		if len(tags) > 1 {
			optionInfo.Notice = tags[len(tags)-1]
		}
		if len(tags) > 2 {
			optionInfo.RemovalIn = tags[1]
		}
		optionInfo.determineEscalation(c.VersionMajor, c.VersionMinor)

		// Get the toml field name
		option := field.Tag.Get("toml")
		if option != "" {
			optionInfo.Name = option
		}
		info.Options = append(info.Options, optionInfo)
	})

	return info
}

func (c *Config) printUserDeprecation(category, name string, plugin interface{}) error {
	info := c.collectDeprecationInfo(category, name, plugin, false)

	switch info.Level {
	case Warn:
		prefix := "W! " + color.YellowString("DeprecationWarning")
		info.printPluginDeprecationNotice(prefix)
		// We will not check for any deprecated options as the whole plugin is deprecated anyway.
		return nil
	case Error:
		prefix := "E! " + color.RedString("DeprecationError")
		info.printPluginDeprecationNotice(prefix)
		// We are past the grace period
		return fmt.Errorf("plugin deprecated")
	}

	// Print deprecated options
	deprecatedOptions := make([]string, 0)
	for _, option := range info.Options {
		switch option.Level {
		case Warn:
			prefix := "W! " + color.YellowString("DeprecationWarning")
			option.printOptionDeprecationNotice(prefix, info.Name)
		case Error:
			prefix := "E! " + color.RedString("DeprecationError")
			option.printOptionDeprecationNotice(prefix, info.Name)
			deprecatedOptions = append(deprecatedOptions, option.Name)
		}
	}

	if len(deprecatedOptions) > 0 {
		return fmt.Errorf("plugin options %q deprecated", strings.Join(deprecatedOptions, ","))
	}

	return nil
}

func (c *Config) CollectDeprecationInfos(inFilter, outFilter, aggFilter, procFilter []string) map[string][]pluginDeprecationInfo {
	infos := make(map[string][]pluginDeprecationInfo)

	infos["inputs"] = make([]pluginDeprecationInfo, 0)
	for name, creator := range inputs.Inputs {
		if len(inFilter) > 0 && !sliceContains(name, inFilter) {
			continue
		}

		plugin := creator()
		info := c.collectDeprecationInfo("inputs", name, plugin, true)

		if info.Level != None || len(info.Options) > 0 {
			infos["inputs"] = append(infos["inputs"], info)
		}
	}

	infos["outputs"] = make([]pluginDeprecationInfo, 0)
	for name, creator := range outputs.Outputs {
		if len(outFilter) > 0 && !sliceContains(name, outFilter) {
			continue
		}

		plugin := creator()
		info := c.collectDeprecationInfo("outputs", name, plugin, true)

		if info.Level != None || len(info.Options) > 0 {
			infos["outputs"] = append(infos["outputs"], info)
		}
	}

	infos["processors"] = make([]pluginDeprecationInfo, 0)
	for name, creator := range processors.Processors {
		if len(procFilter) > 0 && !sliceContains(name, procFilter) {
			continue
		}

		plugin := creator()
		info := c.collectDeprecationInfo("processors", name, plugin, true)

		if info.Level != None || len(info.Options) > 0 {
			infos["processors"] = append(infos["processors"], info)
		}
	}

	infos["aggregators"] = make([]pluginDeprecationInfo, 0)
	for name, creator := range aggregators.Aggregators {
		if len(aggFilter) > 0 && !sliceContains(name, aggFilter) {
			continue
		}

		plugin := creator()
		info := c.collectDeprecationInfo("aggregators", name, plugin, true)

		if info.Level != None || len(info.Options) > 0 {
			infos["aggregators"] = append(infos["aggregators"], info)
		}
	}

	return infos
}

func (c *Config) PrintDeprecationList(infos []pluginDeprecationInfo) {
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

// walkPluginStruct iterates over the fields of a structure in depth-first search (to cover nested structures)
// and calls the given function for every visited field.
func walkPluginStruct(value reflect.Value, fn func(f reflect.StructField, fv reflect.Value)) {
	v := reflect.Indirect(value)
	t := v.Type()

	// Only works on structs
	if t.Kind() != reflect.Struct {
		return
	}

	// Walk over the struct fields and call the given function. If we encounter more complex embedded
	// elements (stucts, slices/arrays, maps) we need to descend into those elements as they might
	// contain structures nested in the current structure.
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
				element := fieldValue.Index(j)
				// The array might contain structs
				walkPluginStruct(element, fn)
				fn(field, element)
			}
		case reflect.Map:
			iter := fieldValue.MapRange()
			for iter.Next() {
				element := iter.Value()
				// The map might contain structs
				walkPluginStruct(element, fn)
				fn(field, element)
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
