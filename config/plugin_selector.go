package config

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf/filter"
)

const (
	selectorSeparator = ";"
)

var (
	// keyRegex is the regex for valid keys in selectors
	keyRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	// valueRegex is the regex for valid values in selectors
	//
	// It is same as keyRegex, but allows wildcard characters like '*', '?'.
	valueRegex = regexp.MustCompile(`^[a-z0-9*?]([-a-z0-9*?]*[a-z0-9*?])?$`)
)

type pluginSelectors [][]string

func (ps *pluginSelectors) String() string {
	// Join each selector group with a comma, and then join the groups with a semicolon
	var result []string
	for _, group := range *ps {
		result = append(result, strings.Join(group, ","))
	}
	return strings.Join(result, ";")
}

var (
	// SelectorFlags holds the selectors for the running instance of telegraf
	// These are user provided flags that can be used to filter plugins based on labels
	SelectorFlags []string

	// parsedSelectors holds the parsed selectors from SelectorFlags
	// It is a slice of slices, where each inner slice contains key-value pairs
	// For example, if SelectorFlags contains ["env=prod,app=api", "region=dc-23,node=host-*-dc,kind=metrics"],
	// parsedSelectors will be:
	// {
	//   {"env=prod", "app=api"},
	//   {"region=dc-23", "node=host-*-dc", "kind=metrics"},
	// }
	parsedSelectors pluginSelectors
)

func ParseSelectors() error {
	selectors := make(pluginSelectors, 0, len(SelectorFlags))
	for _, selector := range SelectorFlags {
		keySet := make(map[string]struct{})
		groups := strings.Split(selector, selectorSeparator)
		if len(groups) == 0 {
			return fmt.Errorf("empty selector provided")
		}
		for _, group := range groups {
			kv := strings.SplitN(group, "=", 2)
			if len(kv) != 2 {
				return fmt.Errorf("invalid selector '%s', expected 'key=value'", group)
			}
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			if !keyRegex.MatchString(key) {
				return fmt.Errorf("invalid selector key %q", key)
			}
			if !valueRegex.MatchString(value) {
				return fmt.Errorf("invalid selector value %q", value)
			}
			if _, exists := keySet[key]; exists {
				return fmt.Errorf("duplicate key '%s' found in selectors", key)
			}
			keySet[key] = struct{}{}
		}
		selectors = append(selectors, groups)
	}
	parsedSelectors = selectors
	log.Printf("D! Telegraf configured with selectors: %s", parsedSelectors.String())
	return nil
}

// shouldPluginRun decides whether the plugin should run based on CLI selectors and plugin labels.
func shouldPluginRun(selectors pluginSelectors, pluginLabels map[string]string) bool {
	// If no selectors or no labels are provided, always run (backward compatibility)
	if len(selectors) == 0 || len(pluginLabels) == 0 {
		return true
	}

	// Loop over each selector string (AND logic inside each group)
	for _, selector := range selectors {
		if matchesSelectorGroup(selector, pluginLabels) {
			return true // OR: any matching group is enough
		}
	}

	return false
}

// matchesSelectorGroup checks if a single selector group matches the labels.
func matchesSelectorGroup(selector []string, pluginLabels map[string]string) bool {
	includePatterns := selector
	excludePatterns := []string{}

	includeExcludeFilter, err := filter.NewIncludeExcludeFilter(includePatterns, excludePatterns)
	if err != nil {
		log.Printf("E! Error creating IncludeExcludeFilter: %v", err)
		return false
	}

	for _, condition := range includePatterns {
		// len(kv) will always be 2 due to earlier validation
		kv := strings.SplitN(condition, "=", 2)
		key := strings.TrimSpace(kv[0])

		labelValue, ok := pluginLabels[key]
		if !ok {
			// Label key missing → fail this selector
			return false
		}
		if !includeExcludeFilter.Match(fmt.Sprintf("%s=%s", key, labelValue)) {
			return false
		}
	}

	// All conditions matched (AND)
	return true
}
