package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/config"
)

type pluginState map[string]bool
type selection map[string]pluginState

func ImportConfigurations(files, dirs []string) (*selection, int, error) {
	sel := selection(make(map[string]pluginState))

	// Initialize the categories
	for _, category := range categories {
		sel[category] = make(map[string]bool)
	}

	// Gather all configuration files
	var filenames []string
	filenames = append(filenames, files...)

	for _, dir := range dirs {
		// Walk the directory and get the packages
		elements, err := os.ReadDir(dir)
		if err != nil {
			return nil, 0, fmt.Errorf("reading directory %q failed: %w", dir, err)
		}

		for _, element := range elements {
			if element.IsDir() || filepath.Ext(element.Name()) != ".conf" {
				continue
			}

			filenames = append(filenames, filepath.Join(dir, element.Name()))
		}
	}
	if len(filenames) == 0 {
		return &sel, 0, errors.New("no configuration files given or found")
	}

	// Do the actual import
	err := sel.importFiles(filenames)
	return &sel, len(filenames), err
}

func (s *selection) Filter(p packageCollection) *packageCollection {
	enabled := packageCollection{
		packages: map[string][]packageInfo{},
	}

	for category, pkgs := range p.packages {
		var categoryEnabledPackages []packageInfo
		settings := (*s)[category]
		for _, pkg := range pkgs {
			if _, found := settings[pkg.Plugin]; found {
				categoryEnabledPackages = append(categoryEnabledPackages, pkg)
			}
		}
		enabled.packages[category] = categoryEnabledPackages
	}

	// Make sure we update the list of default parsers used by
	// the remaining packages
	enabled.FillDefaultParsers()

	// If the user did not configure any parser, we want to include
	// the default parsers if any to preserve a functional set of
	// plugins.
	if len(enabled.packages["parsers"]) == 0 && len(enabled.defaultParsers) > 0 {
		var parsers []packageInfo
		for _, pkg := range p.packages["parsers"] {
			for _, name := range enabled.defaultParsers {
				if pkg.Plugin == name {
					parsers = append(parsers, pkg)
					break
				}
			}
		}
		enabled.packages["parsers"] = parsers
	}

	return &enabled
}

func (s *selection) importFiles(configurations []string) error {
	for _, cfg := range configurations {
		buf, err := config.LoadConfigFile(cfg)
		if err != nil {
			return fmt.Errorf("reading %q failed: %v", cfg, err)
		}

		if err := s.extractPluginsFromConfig(buf); err != nil {
			return fmt.Errorf("extracting plugins from %q failed: %v", cfg, err)
		}
	}

	return nil
}

func (s *selection) extractPluginsFromConfig(buf []byte) error {
	table, err := toml.Parse(trimBOM(buf))
	if err != nil {
		return fmt.Errorf("parsing TOML failed: %w", err)
	}

	for category, subtbl := range table.Fields {
		categoryTbl, ok := subtbl.(*ast.Table)
		if !ok {
			continue
		}

		if _, found := (*s)[category]; !found {
			continue
		}

		for name, data := range categoryTbl.Fields {
			(*s)[category][name] = true

			// We need to check the data_format field to get all required parsers
			switch category {
			case "inputs", "processors":
				pluginTables, ok := data.([]*ast.Table)
				if !ok {
					continue
				}
				for _, subsubtbl := range pluginTables {
					for field, fieldData := range subsubtbl.Fields {
						if field != "data_format" {
							continue
						}
						kv := fieldData.(*ast.KeyValue)
						name := kv.Value.(*ast.String)
						(*s)["parsers"][name.Value] = true
					}
				}
			}
		}
	}

	return nil
}

func trimBOM(f []byte) []byte {
	return bytes.TrimPrefix(f, []byte("\xef\xbb\xbf"))
}
