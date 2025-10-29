package arc

import (
	"fmt"

	"github.com/influxdata/telegraf"
)

type group struct {
	name    string
	rows    int
	columns map[string][]interface{}
}

func (g *group) add(m telegraf.Metric) {
	// Track which columns are present in the current metric
	set := make(map[string]bool, len(m.FieldList())+len(m.TagList())+1)

	// Add all fields as columns
	for _, f := range m.FieldList() {
		if _, found := g.columns[f.Key]; !found {
			// Backfill column with nil for previous rows
			g.columns[f.Key] = make([]interface{}, g.rows)
		}
		g.columns[f.Key] = append(g.columns[f.Key], f.Value)
		set[f.Key] = true
	}

	// Add all tags as columns
	for _, f := range m.TagList() {
		if _, found := g.columns[f.Key]; !found {
			// Backfill column with nil for previous rows
			g.columns[f.Key] = make([]interface{}, g.rows)
		}
		g.columns[f.Key] = append(g.columns[f.Key], f.Value)
		set[f.Key] = true
	}

	// Always add timestamp column
	g.columns["time"] = append(g.columns["time"], m.Time().UnixMilli())
	set["time"] = true

	// For any existing columns not in this metric, append nil to maintain alignment
	for k := range g.columns {
		if !set[k] {
			g.columns[k] = append(g.columns[k], nil)
		}
	}

	g.rows++
}

func (g *group) produceMessage() (map[string]interface{}, error) {
	// Verify all columns have the same number of rows
	for k, v := range g.columns {
		if len(v) != g.rows {
			return nil, fmt.Errorf(
				"column %q has %d entries (expect %d); potential field and tag name collision",
				k, len(v), g.rows,
			)
		}
	}

	// Return the columnar data in Arc's expected format
	return map[string]interface{}{
		"m":       g.name,
		"columns": g.columns,
	}, nil
}
