package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/config"
)

type instance struct {
	category   string
	name       string
	enabled    bool
	dataformat string
}

type selection struct {
	plugins map[string][]instance
}

func ImportConfigurations(files, dirs []string) (*selection, int, error) {
	sel := &selection{
		plugins: make(map[string][]instance),
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
		return sel, 0, errors.New("no configuration files given or found")
	}

	// Do the actual import
	err := sel.importFiles(filenames)
	return sel, len(filenames), err
}

func (s *selection) Filter(p packageCollection) (*packageCollection, error) {
	enabled := packageCollection{
		packages: map[string][]packageInfo{},
	}

	implicitlyConfigured := make(map[string]bool)
	for category, pkgs := range p.packages {
		for _, pkg := range pkgs {
			key := category + "." + pkg.Plugin
			instances, found := s.plugins[key]
			if !found {
				continue
			}

			// The package was configured so add it to the enabled list
			enabled.packages[category] = append(enabled.packages[category], pkg)

			// Check if the instances configured a data-format and decide if it
			// is a parser or serializer depending on the plugin type.
			// If no data-format was configured, check the default settings in
			// case this plugin supports a data-format setting but the user
			// didn't set it.
			for _, instance := range instances {
				parser := pkg.DefaultParser
				serializer := pkg.DefaultSerializer
				if instance.dataformat != "" {
					switch category {
					case "inputs":
						parser = instance.dataformat
					case "processors":
						parser = instance.dataformat
						// The execd processor requires both a parser and serializer
						if pkg.Plugin == "execd" {
							serializer = instance.dataformat
						}
					case "outputs":
						serializer = instance.dataformat
					}
				}
				if parser != "" {
					implicitlyConfigured["parsers."+parser] = true
				}
				if serializer != "" {
					implicitlyConfigured["serializers."+serializer] = true
				}
			}
		}
	}

	// Iterate over all plugins AGAIN to add the implicitly configured packages
	// such as parsers and serializers
	for category, pkgs := range p.packages {
		for _, pkg := range pkgs {
			key := category + "." + pkg.Plugin

			// Skip the plugins that were explicitly configured as we already
			// added them above.
			if _, found := s.plugins[key]; found {
				continue
			}

			// Add the package if it was implicitly configured e.g. by a
			// 'data_format' setting or by a default value for the data-format
			if _, implicit := implicitlyConfigured[key]; implicit {
				enabled.packages[category] = append(enabled.packages[category], pkg)
			}
		}
	}

	// Check if all packages in the config were covered
	available := make(map[string]bool)
	for category, pkgs := range p.packages {
		for _, pkg := range pkgs {
			available[category+"."+pkg.Plugin] = true
		}
	}

	var unknown []string
	for pkg := range s.plugins {
		if !available[pkg] {
			unknown = append(unknown, pkg)
		}
	}
	for pkg := range implicitlyConfigured {
		if !available[pkg] {
			unknown = append(unknown, pkg)
		}
	}
	if len(unknown) > 0 {
		return nil, fmt.Errorf("configured but unknown packages %q", strings.Join(unknown, ","))
	}

	return &enabled, nil
}

func (s *selection) importFiles(configurations []string) error {
	for _, cfg := range configurations {
		buf, _, err := config.LoadConfigFile(cfg)
		if err != nil {
			return fmt.Errorf("reading %q failed: %w", cfg, err)
		}

		if err := s.extractPluginsFromConfig(buf); err != nil {
			return fmt.Errorf("extracting plugins from %q failed: %w", cfg, err)
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
		// Check if we should handle the category, i.e. it contains plugins
		// to configure.
		var valid bool
		for _, c := range categories {
			if c == category {
				valid = true
				break
			}
		}
		if !valid {
			continue
		}

		categoryTbl, ok := subtbl.(*ast.Table)
		if !ok {
			continue
		}

		for name, data := range categoryTbl.Fields {
			key := category + "." + name
			cfg := instance{
				category: category,
				name:     name,
				enabled:  true,
			}

			// We need to check the data_format field to get all required
			// parsers and serializers
			pluginTables, ok := data.([]*ast.Table)
			if ok {
				for _, subsubtbl := range pluginTables {
					var dataformat string
					for field, fieldData := range subsubtbl.Fields {
						if field != "data_format" {
							continue
						}
						kv := fieldData.(*ast.KeyValue)
						option := kv.Value.(*ast.String)
						dataformat = option.Value
					}
					cfg.dataformat = dataformat
				}
			}
			s.plugins[key] = append(s.plugins[key], cfg)
		}
	}

	return nil
}

func trimBOM(f []byte) []byte {
	return bytes.TrimPrefix(f, []byte("\xef\xbb\xbf"))
}
