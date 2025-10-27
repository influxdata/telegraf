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
	set := make(map[string]bool, len(m.FieldList())+len(m.TagList())+1)

	for _, f := range m.FieldList() {
		if _, found := g.columns[f.Key]; !found {
			g.columns[f.Key] = make([]interface{}, g.rows)
		}
		g.columns[f.Key] = append(g.columns[f.Key], f.Value)
		set[f.Key] = true
	}

	for _, f := range m.TagList() {
		if _, found := g.columns[f.Key]; !found {
			g.columns[f.Key] = make([]interface{}, g.rows)
		}
		g.columns[f.Key] = append(g.columns[f.Key], f.Value)
		set[f.Key] = true
	}

	g.columns["time"] = append(g.columns["time"], m.Time().UnixMilli())
	set["time"] = true

	for k := range g.columns {
		if !set[k] {
			g.columns[k] = append(g.columns[k], nil)
		}
	}

	g.rows++
}

func (g *group) produceMessage() (*arcColumnarData, error) {
	msg := &arcColumnarData{
		Measurement: g.name,
		Columns:     make(map[string]interface{}, len(g.columns)),
	}
	for k, v := range g.columns {
		if len(v) != g.rows {
			return nil, fmt.Errorf(
				"column %q has %d entries (expect %d); potential field and tag name collision",
				k, len(v), g.rows,
			)
		}
		msg.Columns[k] = v
	}

	return msg, nil
}
