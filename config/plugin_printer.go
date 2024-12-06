package config

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
)

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

	return GetPluginTableContent([]string{"Name", "Source(s)"}, data)
}

// GetPluginTableContent prints a bordered ASCII table.
func GetPluginTableContent(headers []string, data [][]any) string {
	// Initialize the buffer
	var b bytes.Buffer

	// Initialize the tab writer
	writer := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer)
	// Helper function to convert data into multi-line strings
	convertToLines := func(value interface{}) []string {
		switch v := value.(type) {
		case []string: // If the value is a slice of strings, return as-is
			return v
		default: // Handle other types as single-line strings
			return []string{fmt.Sprintf("%v", v)}
		}
	}

	// Prepare processed data with multi-line values
	processedData := make([][][]string, len(data))
	columnWidths := make([]int, len(headers))
	maxLinesPerRow := make([]int, len(data))

	for i, row := range data {
		processedData[i] = make([][]string, len(row))
		for j, cell := range row {
			lines := convertToLines(cell)
			processedData[i][j] = lines

			// Calculate the maximum width for each column
			for _, line := range lines {
				if len(line) > columnWidths[j] {
					columnWidths[j] = len(line)
				}
			}

			// Update maximum lines per row
			if len(lines) > maxLinesPerRow[i] {
				maxLinesPerRow[i] = len(lines)
			}
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

	// Print top border
	fmt.Fprintln(writer, createLine())

	// Print headers with borders
	fmt.Fprint(writer, "|")
	for i, header := range headers {
		fmt.Fprintf(writer, " %-*s |", columnWidths[i]+2, header)
	}
	fmt.Fprintln(writer)

	// Print separator line
	fmt.Fprintln(writer, createLine())

	// Print rows with borders
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
		// Print separator line after each row
		fmt.Fprintln(writer, createLine())
	}

	// Flush the writer
	writer.Flush()

	return b.String()
}
