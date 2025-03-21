package config

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

var headers = []string{"Name", "Source(s)"}

type pluginPrinter struct {
	name   string
	source string
}

type pluginNames []pluginPrinter

func getPluginSourcesTable(pluginNames []pluginPrinter) string {
	if !PrintPluginConfigSource {
		return ""
	}

	if len(pluginNames) == 0 {
		return ""
	}

	data := make([][]any, 0, len(pluginNames))
	rows := make(map[string][]string)
	for _, plugin := range pluginNames {
		if _, ok := rows[plugin.name]; !ok {
			rows[plugin.name] = make([]string, 0)
		}
		rows[plugin.name] = append(rows[plugin.name], plugin.source)
	}

	for name, sources := range rows {
		var nameCountStr string
		if len(sources) > 1 {
			nameCountStr = fmt.Sprintf("%s (%dx)", name, len(sources))
		} else {
			nameCountStr = name
		}
		data = append(data, []any{nameCountStr, sources})
	}
	sort.Slice(data, func(i, j int) bool {
		return len(data[i][1].([]string)) > len(data[j][1].([]string))
	})
	return getTableString(headers, data)
}

func getTableString(headers []string, data [][]any) string {
	buff := new(bytes.Buffer)

	t := table.NewWriter()
	t.SetOutputMirror(buff)
	t.AppendHeader(convertToRow(headers))

	// Append rows
	for _, row := range data {
		processedRow := make([]interface{}, len(row))
		for i, col := range row {
			switch v := col.(type) {
			case []string: // Convert slices to multi-line strings
				var source map[string]int
				for _, s := range v {
					if source == nil {
						source = make(map[string]int)
					}
					source[s]++
				}
				// sort the sources according to the count
				sources := make([]string, 0, len(source))
				for s := range source {
					sources = append(sources, s)
				}
				sort.Slice(sources, func(i, j int) bool {
					return source[sources[i]] > source[sources[j]]
				})
				for i, s := range sources {
					if source[s] > 1 {
						sources[i] = fmt.Sprintf("%s (%dx)", s, source[s])
					}
				}
				processedRow[i] = strings.Join(sources, "\n")
			default:
				processedRow[i] = v
			}
		}
		t.AppendRow(processedRow)
	}

	t.Style().Options.SeparateRows = true
	return t.Render()
}

// Helper function to convert headers to table.Row
func convertToRow(data []string) table.Row {
	row := make(table.Row, len(data))
	for i, val := range data {
		row[i] = val
	}
	return row
}
