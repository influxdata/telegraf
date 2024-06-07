package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
	_ "github.com/influxdata/telegraf/migrations/all" // register all migrations
)

type section struct {
	name    string
	begin   int
	content *ast.Table
	raw     *bytes.Buffer
}

func splitToSections(root *ast.Table) ([]section, error) {
	var sections []section
	for name, elements := range root.Fields {
		switch name {
		case "inputs", "outputs", "processors", "aggregators":
			category, ok := elements.(*ast.Table)
			if !ok {
				return nil, fmt.Errorf("%q is not a table (%T)", name, category)
			}

			for plugin, elements := range category.Fields {
				tbls, ok := elements.([]*ast.Table)
				if !ok {
					return nil, fmt.Errorf("elements of \"%s.%s\" is not a list of tables (%T)", name, plugin, elements)
				}
				for _, tbl := range tbls {
					s := section{
						name:    name + "." + tbl.Name,
						begin:   tbl.Line,
						content: tbl,
						raw:     &bytes.Buffer{},
					}
					sections = append(sections, s)
				}
			}
		default:
			tbl, ok := elements.(*ast.Table)
			if !ok {
				return nil, fmt.Errorf("%q is not a table (%T)", name, elements)
			}
			s := section{
				name:    name,
				begin:   tbl.Line,
				content: tbl,
				raw:     &bytes.Buffer{},
			}
			sections = append(sections, s)
		}
	}

	// Sort the TOML elements by begin (line-number)
	sort.SliceStable(sections, func(i, j int) bool { return sections[i].begin < sections[j].begin })

	return sections, nil
}

func assignTextToSections(data []byte, sections []section) ([]section, error) {
	// Now assign the raw text to each section
	if sections[0].begin > 0 {
		sections = append([]section{{
			name:  "header",
			begin: 0,
			raw:   &bytes.Buffer{},
		}}, sections...)
	}

	var lineno int
	scanner := bufio.NewScanner(bytes.NewBuffer(data))
	for idx, next := range sections[1:] {
		var buf bytes.Buffer
		for lineno < next.begin-1 {
			if !scanner.Scan() {
				break
			}
			lineno++

			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "#") {
				buf.Write(scanner.Bytes())
				buf.WriteString("\n")
				continue
			} else if buf.Len() > 0 {
				if _, err := io.Copy(sections[idx].raw, &buf); err != nil {
					return nil, fmt.Errorf("copying buffer failed: %w", err)
				}
				buf.Reset()
			}

			sections[idx].raw.Write(scanner.Bytes())
			sections[idx].raw.WriteString("\n")
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("splitting by line failed: %w", err)
		}

		// If a comment is directly in front of the next section, without
		// newline, the comment is assigned to the next section.
		if buf.Len() > 0 {
			if _, err := io.Copy(sections[idx+1].raw, &buf); err != nil {
				return nil, fmt.Errorf("copying buffer failed: %w", err)
			}
			buf.Reset()
		}
	}
	// Write the remaining to the last section
	for scanner.Scan() {
		sections[len(sections)-1].raw.Write(scanner.Bytes())
		sections[len(sections)-1].raw.WriteString("\n")
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("splitting by line failed: %w", err)
	}
	return sections, nil
}

func ApplyMigrations(data []byte) ([]byte, uint64, error) {
	root, err := toml.Parse(data)
	if err != nil {
		return nil, 0, fmt.Errorf("parsing failed: %w", err)
	}

	// Split the configuration into sections containing the location
	// in the file.
	sections, err := splitToSections(root)
	if err != nil {
		return nil, 0, fmt.Errorf("splitting to sections failed: %w", err)
	}
	if len(sections) == 0 {
		return nil, 0, errors.New("no TOML configuration found")
	}

	// Assign the configuration text to the corresponding segments
	sections, err = assignTextToSections(data, sections)
	if err != nil {
		return nil, 0, fmt.Errorf("assigning text failed: %w", err)
	}

	// Do the actual plugin migration(s)
	var applied uint64
	for idx, s := range sections {
		migrate, found := migrations.PluginMigrations[s.name]
		if !found {
			continue
		}

		log.Printf("D!   migrating plugin %q in line %d...", s.name, s.begin)
		result, msg, err := migrate(s.content)
		if err != nil {
			return nil, 0, fmt.Errorf("migrating %q (line %d) failed: %w", s.name, s.begin, err)
		}
		if msg != "" {
			log.Printf("I! Plugin %q in line %d: %s", s.name, s.begin, msg)
		}
		s.raw = bytes.NewBuffer(result)
		tbl, err := toml.Parse(s.raw.Bytes())
		if err != nil {
			return nil, 0, fmt.Errorf("reparsing migrated %q (line %d) failed: %w", s.name, s.begin, err)
		}
		s.content = tbl
		sections[idx] = s
		applied++
	}

	// Do the actual plugin option migration(s)
	for idx, s := range sections {
		migrate, found := migrations.PluginOptionMigrations[s.name]
		if !found {
			continue
		}

		log.Printf("D!   migrating options of plugin %q in line %d...", s.name, s.begin)
		result, msg, err := migrate(s.content)
		if err != nil {
			if errors.Is(err, migrations.ErrNotApplicable) {
				continue
			}
			return nil, 0, fmt.Errorf("migrating options of %q (line %d) failed: %w", s.name, s.begin, err)
		}
		if msg != "" {
			log.Printf("I! Plugin %q in line %d: %s", s.name, s.begin, msg)
		}
		s.raw = bytes.NewBuffer(result)
		sections[idx] = s
		applied++
	}

	// Do general migrations applying to all plugins
	for idx, s := range sections {
		parts := strings.Split(s.name, ".")
		if len(parts) != 2 {
			continue
		}
		log.Printf("D!   applying general migrations to plugin %q in line %d...", s.name, s.begin)
		category, name := parts[0], parts[1]
		for _, migrate := range migrations.GeneralMigrations {
			result, msg, err := migrate(category, name, s.content)
			if err != nil {
				if errors.Is(err, migrations.ErrNotApplicable) {
					continue
				}
				return nil, 0, fmt.Errorf("migrating options of %q (line %d) failed: %w", s.name, s.begin, err)
			}
			if msg != "" {
				log.Printf("I! Plugin %q in line %d: %s", s.name, s.begin, msg)
			}
			s.raw = bytes.NewBuffer(result)
			applied++
		}
		sections[idx] = s
	}

	// Reconstruct the config file from the sections
	var buf bytes.Buffer
	for _, s := range sections {
		_, err = s.raw.WriteTo(&buf)
		if err != nil {
			return nil, applied, fmt.Errorf("joining output failed: %w", err)
		}
	}

	return buf.Bytes(), applied, nil
}
