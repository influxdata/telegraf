package inputs_logparser

import (
	"fmt"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old inputs.logparser data structure
	var logparserPlugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &logparserPlugin); err != nil {
		return nil, "", err
	}

	// Initial tail plugin configuration
	tailPlugin := make(map[string]interface{})
	tailPlugin["data_format"] = "grok"

	// Comments: 'inputs.logparser' -> 'inputs.tail'

	// 'files' ([]string) -> 'files' ([]string)
	if rawFiles, found := logparserPlugin["files"]; found {
		files, err := migrations.AsStringSlice(rawFiles)
		if err != nil {
			return nil, "", err
		}
		tailPlugin["files"] = files
	}

	// 'from_beginning' (bool) -> 'initial_read_offset' (string)
	// 'from_beginning: true' -> 'initial_read_offset: "beginning"'
	// 'from_beginning: false' -> 'initial_read_offset: "saved-or-end"'
	if rawFromBeginning, found := logparserPlugin["from_beginning"]; found {
		fromBeginning, ok := rawFromBeginning.(bool)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'from_beginning'", rawFromBeginning)
		}
		if fromBeginning {
			tailPlugin["initial_read_offset"] = "beginning"
		} else {
			tailPlugin["initial_read_offset"] = "saved-or-end"
		}
	} else {
		// Default if 'from_beginning' is not set, i.e. false
		tailPlugin["initial_read_offset"] = "saved-or-end"
	}

	// 'watch_method' (string) -> 'watch_method' (string)
	if rawWatchMethod, found := logparserPlugin["watch_method"]; found {
		watchMethod, ok := rawWatchMethod.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'watch_method'", rawWatchMethod)
		}
		tailPlugin["watch_method"] = watchMethod
	}

	// Get the 'grok' configuration
	grok, err := getGrok(logparserPlugin)
	if err != nil {
		return nil, "", err
	}

	if grok != nil {
		// 'grok.patterns' ([]string) -> 'grok_patterns' ([]string)
		if rawPatterns, found := grok["patterns"]; found {
			patterns, err := migrations.AsStringSlice(rawPatterns)
			if err != nil {
				return nil, "", err
			}
			tailPlugin["grok_patterns"] = patterns
		}

		// 'grok.measurement' (string) -> 'name_override' (string)
		if rawMeasurement, found := grok["measurement"]; found {
			measurement, ok := rawMeasurement.(string)
			if !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'grok.measurement'", rawMeasurement)
			}
			tailPlugin["name_override"] = measurement
		}

		// 'grok.custom_pattern_files' ([]string) -> 'grok_custom_pattern_files' ([]string)
		if rawCustomPatternFiles, found := grok["custom_pattern_files"]; found {
			customPatternFiles, err := migrations.AsStringSlice(rawCustomPatternFiles)
			if err != nil {
				return nil, "", err
			}
			tailPlugin["grok_custom_pattern_files"] = customPatternFiles
		}

		// 'grok.custom_patterns' (string) -> 'grok_custom_patterns' (string)
		if rawCustomPatterns, found := grok["custom_patterns"]; found {
			customPatterns, ok := rawCustomPatterns.(string)
			if !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'grok.custom_patterns'", rawCustomPatterns)
			}
			tailPlugin["grok_custom_patterns"] = customPatterns
		}

		// 'grok.timezone' (string) -> 'grok_timezone' (string)
		if rawTimezone, found := grok["timezone"]; found {
			timezone, ok := rawTimezone.(string)
			if !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'grok.timezone'", rawTimezone)
			}
			tailPlugin["grok_timezone"] = timezone
		}

		// 'grok.unique_timestamp' (string) -> 'grok_unique_timestamp' (string)
		if rawUniqueTimestamp, found := grok["unique_timestamp"]; found {
			uniqueTimestamp, ok := rawUniqueTimestamp.(string)
			if !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'grok.unique_timestamp'", rawUniqueTimestamp)
			}
			tailPlugin["grok_unique_timestamp"] = uniqueTimestamp
		}
	}

	// Create the mew inputs.tail plugin configuration
	cfg := migrations.CreateTOMLStruct("inputs", "tail")
	cfg.Add("inputs", "tail", tailPlugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

func getGrok(logparserPlugin map[string]interface{}) (map[string]interface{}, error) {
	rawGrok := logparserPlugin["grok"]
	if rawGrok == nil {
		return nil, nil
	}

	grok, ok := rawGrok.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type %T for 'grok'", rawGrok)
	}

	return grok, nil
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginMigration("inputs.logparser", migrate)
}
