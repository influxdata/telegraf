package config

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
)

var headers = []string{"Name", "Source(s)"}

type pluginPrinter struct {
	name   string
	source string
}

type pluginNames []pluginPrinter

func (p *pluginNames) String() string {
	pluginMap := make(map[string][]string)
	for _, plugin := range *p {
		if _, ok := pluginMap[plugin.name]; !ok {
			pluginMap[plugin.name] = make([]string, 0)
		}
		pluginMap[plugin.name] = append(pluginMap[plugin.name], plugin.source)
	}

	data := make([][]any, 0, len(pluginMap))
	for name, sources := range pluginMap {
		var nameCountStr string
		if len(sources) > 1 {
			nameCountStr = fmt.Sprintf("%s (%dx)", name, len(sources))
		} else {
			nameCountStr = name
		}
		data = append(data, []any{nameCountStr, sources})
	}
	// sort the data based on counts(i.e number of sources)
	sort.Slice(data, func(i, j int) bool {
		return len(data[i][1].([]string)) > len(data[j][1].([]string))
	})

	if len(data) == 0 {
		return ""
	}

	return GetPluginTableContent(headers, data)
}

// GetPluginTableContent prints a bordered ASCII table.
//
// Inputs:
//
//	headers: slice of strings containing the headers of the table
//	data: slice of slices of strings containing the data to be printed
//
// Reference: https://github.com/olekukonko/tablewriter
func GetPluginTableContent(headers []string, data [][]any) string {
	// processedData will hold processed data with multi-line values(for multiple sources)
	processedData := make([][][]string, len(data))
	columnWidths := make([]int, len(headers))
	maxLinesPerRow := make([]int, len(data))

	var b bytes.Buffer

	writer := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer)

	// Helper function to convert data into multi-line strings
	convertToLines := func(value interface{}) []string {
		switch v := value.(type) {
		case []string:
			return v
		default:
			return []string{fmt.Sprintf("%v", v)}
		}
	}

	// Function to create a horizontal line
	createLine := func() string {
		line := "+"
		for _, width := range columnWidths {
			line += strings.Repeat("-", width+4) + "+"
		}
		return line
	}

	for i, row := range data {
		processedData[i] = make([][]string, len(row))
		for j, cell := range row {
			lines := convertToLines(cell)
			processedData[i][j] = lines

			// maximum width for each column
			for _, line := range lines {
				if len(line) > columnWidths[j] {
					columnWidths[j] = len(line)
				}
			}

			// maximum lines per row
			if len(lines) > maxLinesPerRow[i] {
				maxLinesPerRow[i] = len(lines)
			}
		}
	}

	// Print top line border
	// ex  + ---- +
	fmt.Fprintln(writer, createLine())

	// Print headers with borders
	// ex  | Name | Source(s) |
	fmt.Fprint(writer, "|")
	for i, header := range headers {
		fmt.Fprintf(writer, " %-*s |", columnWidths[i]+2, header)
	}
	fmt.Fprintln(writer)

	// Print separator line
	fmt.Fprintln(writer, createLine())

	// rows
	for rowIndex, row := range processedData {
		for lineIndex := 0; lineIndex < maxLinesPerRow[rowIndex]; lineIndex++ {
			fmt.Fprint(writer, "|")
			for colIndex, cell := range row {
				if lineIndex < len(cell) {
					fmt.Fprintf(writer, " %-*s |", columnWidths[colIndex]+2, cell[lineIndex])
				} else {
					fmt.Fprintf(writer, " %-*s |", columnWidths[colIndex]+2, "") // Empty space for missing lines
				}
			}
			fmt.Fprintln(writer)
		}
		// separator line
		fmt.Fprintln(writer, createLine())
	}
	writer.Flush()

	return b.String()
}
