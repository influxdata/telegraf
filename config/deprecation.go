package config

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/fatih/color"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/processors"
)

// DeprecationInfo contains all important information to describe a deprecated entity
type DeprecationInfo struct {
	// Name of the plugin or plugin option
	Name string
	// LogLevel is the level of deprecation which currently corresponds to a log-level
	logLevel telegraf.LogLevel
	info     telegraf.DeprecationInfo
}

func (di *DeprecationInfo) determineEscalation() error {
	di.logLevel = telegraf.None
	if di.info.Since == "" {
		return nil
	}

	since, err := semver.NewVersion(di.info.Since)
	if err != nil {
		return fmt.Errorf("cannot parse 'since' version %q: %w", di.info.Since, err)
	}

	var removal *semver.Version
	if di.info.RemovalIn != "" {
		removal, err = semver.NewVersion(di.info.RemovalIn)
		if err != nil {
			return fmt.Errorf("cannot parse 'removal' version %q: %w", di.info.RemovalIn, err)
		}
	} else {
		removal = &semver.Version{Major: since.Major}
		removal.BumpMajor()
		di.info.RemovalIn = removal.String()
	}

	// Drop potential pre-release tags
	version := semver.Version{
		Major: telegrafVersion.Major,
		Minor: telegrafVersion.Minor,
		Patch: telegrafVersion.Patch,
	}
	if !version.LessThan(*removal) {
		di.logLevel = telegraf.Error
	} else if !version.LessThan(*since) {
		di.logLevel = telegraf.Warn
	}
	return nil
}

// PluginDeprecationInfo holds all information about a deprecated plugin or it's options
type PluginDeprecationInfo struct {
	DeprecationInfo

	// Options deprecated for this plugin
	Options []DeprecationInfo
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

func (c *Config) collectDeprecationInfo(category, name string, plugin interface{}, all bool) PluginDeprecationInfo {
	info := PluginDeprecationInfo{
		DeprecationInfo: DeprecationInfo{
			Name:     category + "." + name,
			logLevel: telegraf.None,
		},
	}

	// First check if the whole plugin is deprecated
	switch category {
	case "aggregators":
		if pi, deprecated := aggregators.Deprecations[name]; deprecated {
			info.DeprecationInfo.info = pi
		}
	case "inputs":
		if pi, deprecated := inputs.Deprecations[name]; deprecated {
			info.DeprecationInfo.info = pi
		}
	case "outputs":
		if pi, deprecated := outputs.Deprecations[name]; deprecated {
			info.DeprecationInfo.info = pi
		}
	case "processors":
		if pi, deprecated := processors.Deprecations[name]; deprecated {
			info.DeprecationInfo.info = pi
		}
	}
	if err := info.determineEscalation(); err != nil {
		panic(fmt.Errorf("plugin %q: %w", info.Name, err))
	}
	if info.logLevel != telegraf.None {
		c.incrementPluginDeprecations(category)
	}

	// Allow checking for names only.
	if plugin == nil {
		return info
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
		optionInfo := DeprecationInfo{Name: field.Name}
		optionInfo.info.Since = tags[0]

		if len(tags) > 1 {
			optionInfo.info.Notice = tags[len(tags)-1]
		}
		if len(tags) > 2 {
			optionInfo.info.RemovalIn = tags[1]
		}
		if err := optionInfo.determineEscalation(); err != nil {
			panic(fmt.Errorf("plugin %q option %q: %w", info.Name, field.Name, err))
		}

		if optionInfo.logLevel != telegraf.None {
			c.incrementPluginOptionDeprecations(category)
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

func (c *Config) printUserDeprecation(category, name string, plugin interface{}) error {
	info := c.collectDeprecationInfo(category, name, plugin, false)
	printPluginDeprecationNotice(info.logLevel, info.Name, info.info)

	if info.logLevel == telegraf.Error {
		return errors.New("plugin deprecated")
	}

	// Print deprecated options
	deprecatedOptions := make([]string, 0)
	for _, option := range info.Options {
		PrintOptionDeprecationNotice(info.Name, option.Name, option.info)
		if option.logLevel == telegraf.Error {
			deprecatedOptions = append(deprecatedOptions, option.Name)
		}
	}

	if len(deprecatedOptions) > 0 {
		return fmt.Errorf("plugin options %q deprecated", strings.Join(deprecatedOptions, ","))
	}

	return nil
}

func (c *Config) CollectDeprecationInfos(inFilter, outFilter, aggFilter, procFilter []string) map[string][]PluginDeprecationInfo {
	infos := make(map[string][]PluginDeprecationInfo)

	infos["inputs"] = make([]PluginDeprecationInfo, 0)
	for name, creator := range inputs.Inputs {
		if len(inFilter) > 0 && !sliceContains(name, inFilter) {
			continue
		}

		plugin := creator()
		info := c.collectDeprecationInfo("inputs", name, plugin, true)

		if info.logLevel != telegraf.None || len(info.Options) > 0 {
			infos["inputs"] = append(infos["inputs"], info)
		}
	}

	infos["outputs"] = make([]PluginDeprecationInfo, 0)
	for name, creator := range outputs.Outputs {
		if len(outFilter) > 0 && !sliceContains(name, outFilter) {
			continue
		}

		plugin := creator()
		info := c.collectDeprecationInfo("outputs", name, plugin, true)

		if info.logLevel != telegraf.None || len(info.Options) > 0 {
			infos["outputs"] = append(infos["outputs"], info)
		}
	}

	infos["processors"] = make([]PluginDeprecationInfo, 0)
	for name, creator := range processors.Processors {
		if len(procFilter) > 0 && !sliceContains(name, procFilter) {
			continue
		}

		plugin := creator()
		info := c.collectDeprecationInfo("processors", name, plugin, true)

		if info.logLevel != telegraf.None || len(info.Options) > 0 {
			infos["processors"] = append(infos["processors"], info)
		}
	}

	infos["aggregators"] = make([]PluginDeprecationInfo, 0)
	for name, creator := range aggregators.Aggregators {
		if len(aggFilter) > 0 && !sliceContains(name, aggFilter) {
			continue
		}

		plugin := creator()
		info := c.collectDeprecationInfo("aggregators", name, plugin, true)

		if info.logLevel != telegraf.None || len(info.Options) > 0 {
			infos["aggregators"] = append(infos["aggregators"], info)
		}
	}

	return infos
}

func (c *Config) PrintDeprecationList(plugins []PluginDeprecationInfo) {
	sort.Slice(plugins, func(i, j int) bool { return plugins[i].Name < plugins[j].Name })

	for _, plugin := range plugins {
		switch plugin.logLevel {
		case telegraf.Warn, telegraf.Error:
			fmt.Printf(
				"  %-40s %-5s since %-5s removal in %-5s %s\n",
				plugin.Name, plugin.logLevel, plugin.info.Since, plugin.info.RemovalIn, plugin.info.Notice,
			)
		}

		if len(plugin.Options) < 1 {
			continue
		}
		sort.Slice(plugin.Options, func(i, j int) bool { return plugin.Options[i].Name < plugin.Options[j].Name })
		for _, option := range plugin.Options {
			fmt.Printf(
				"  %-40s %-5s since %-5s removal in %-5s %s\n",
				plugin.Name+"/"+option.Name, option.logLevel, option.info.Since, option.info.RemovalIn, option.info.Notice,
			)
		}
	}
}

func printHistoricPluginDeprecationNotice(category, name string, info telegraf.DeprecationInfo) {
	prefix := "E! " + color.RedString("DeprecationError")
	log.Printf(
		"%s: Plugin %q deprecated since version %s and removed: %s",
		prefix, category+"."+name, info.Since, info.Notice,
	)
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
	// elements (structs, slices/arrays, maps) we need to descend into those elements as they might
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

func deprecationPrefix(level telegraf.LogLevel) string {
	switch level {
	case telegraf.Warn:
		return "W! " + color.YellowString("DeprecationWarning")
	case telegraf.Error:
		return "E! " + color.RedString("DeprecationError")
	}
	return ""
}

func printPluginDeprecationNotice(level telegraf.LogLevel, name string, info telegraf.DeprecationInfo) {
	switch level {
	case telegraf.Warn, telegraf.Error:
		prefix := deprecationPrefix(level)

		log.Printf(
			"%s: Plugin %q deprecated since version %s and will be removed in %s: %s",
			prefix, name, info.Since, info.RemovalIn, info.Notice,
		)
	}
}

func PrintOptionDeprecationNotice(plugin, option string, info telegraf.DeprecationInfo) {
	// Determine the log-level
	di := &DeprecationInfo{
		Name: plugin,
		info: info,
	}
	if err := di.determineEscalation(); err != nil {
		log.Printf("E! Determining log-level for option %s in plugin %s failed: %v", option, plugin, err)
		return
	}

	switch di.logLevel {
	case telegraf.Warn, telegraf.Error:
		prefix := deprecationPrefix(di.logLevel)
		log.Printf(
			"%s: Option %q of plugin %q deprecated since version %s and will be removed in %s: %s",
			prefix, option, plugin, info.Since, info.RemovalIn, info.Notice,
		)
	}
}

func PrintOptionValueDeprecationNotice(plugin, option string, value interface{}, info telegraf.DeprecationInfo) {
	// Determine the log-level
	di := &DeprecationInfo{
		Name: plugin,
		info: info,
	}
	if err := di.determineEscalation(); err != nil {
		log.Printf("E! Determining log-level for option %s in plugin %s failed: %v", option, plugin, err)
		return
	}

	switch di.logLevel {
	case telegraf.Warn, telegraf.Error:
		prefix := deprecationPrefix(di.logLevel)
		log.Printf(
			`%s: Value "%+v" for option %q of plugin %q deprecated since version %s and will be removed in %s: %s`,
			prefix, value, option, plugin, info.Since, info.RemovalIn, info.Notice,
		)
	}
}
