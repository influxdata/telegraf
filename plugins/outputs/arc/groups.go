package arc

import (
	"fmt"

	"github.com/tinylib/msgp/msgp"

	"github.com/influxdata/telegraf"
)

// timeColumn is the name of the column carrying the metric timestamps. It is
// reserved, so a tag or field with this name is treated as a collision.
const timeColumn = "time"

// cell is a single present value of a column. Exactly one of tag and field is
// set; both point into the metric's own tag and field lists, so no value is
// copied or boxed.
type cell struct {
	row   int
	tag   *telegraf.Tag
	field *telegraf.Field
}

// group collects the metrics of one measurement and encodes them in Arc's
// columnar format: {"m": <name>, "columns": {<column>: [<value per row>]}}.
// Rows without a value for a column carry nil in that column.
//
// Only the present values are kept in memory, so the cost is proportional to
// the number of tags and fields in the batch, not to rows times columns,
// which matters for sparse batches such as those of inputs.prometheus with
// metric_version = 2.
type group struct {
	name    string
	metrics []telegraf.Metric

	columns []string // column names in order of first appearance
	starts  []int    // cells[starts[i]:starts[i+1]] belong to columns[i]
	cells   []cell   // present values, grouped by column, rows ascending
}

// prepare builds the column layout of the group. It returns an error if a key
// is used both as a tag and a field of the same metric, or if a tag or field
// is named like the reserved time column. Such a group cannot be encoded, as
// the column would have more values than rows.
func (g *group) prepare() error {
	var total int
	for _, m := range g.metrics {
		total += len(m.TagList()) + len(m.FieldList())
	}

	// First pass: assign a column index to every cell and count the cells of
	// each column.
	index := make(map[string]int)
	cellColumn := make([]int, 0, total)
	var counts, lastRow []int
	assign := func(key string, row int) error {
		if key == timeColumn {
			return fmt.Errorf("measurement %q: key %q collides with the reserved time column", g.name, key)
		}
		idx, found := index[key]
		if !found {
			idx = len(g.columns)
			index[key] = idx
			g.columns = append(g.columns, key)
			counts = append(counts, 0)
			lastRow = append(lastRow, -1)
		} else if lastRow[idx] == row {
			return fmt.Errorf("measurement %q: key %q is used both as tag and field", g.name, key)
		}
		lastRow[idx] = row
		counts[idx]++
		cellColumn = append(cellColumn, idx)
		return nil
	}
	for row, m := range g.metrics {
		for _, f := range m.FieldList() {
			if err := assign(f.Key, row); err != nil {
				return err
			}
		}
		for _, t := range m.TagList() {
			if err := assign(t.Key, row); err != nil {
				return err
			}
		}
	}

	// Second pass: place the cells of each column next to each other, keeping
	// the row order.
	g.starts = make([]int, len(g.columns)+1)
	for i, c := range counts {
		g.starts[i+1] = g.starts[i] + c
	}
	next := make([]int, len(g.columns))
	copy(next, g.starts)
	g.cells = make([]cell, total)
	var pos int
	for row, m := range g.metrics {
		for _, f := range m.FieldList() {
			idx := cellColumn[pos]
			g.cells[next[idx]] = cell{row: row, field: f}
			next[idx]++
			pos++
		}
		for _, t := range m.TagList() {
			idx := cellColumn[pos]
			g.cells[next[idx]] = cell{row: row, tag: t}
			next[idx]++
			pos++
		}
	}

	return nil
}

// encode writes the group as a MessagePack map. prepare must have succeeded.
// nils is a run of MessagePack nil markers used to fill the absent cells of a
// column with as few writes as possible.
func (g *group) encode(w *msgp.Writer, nils []byte) error {
	rows := uint32(len(g.metrics))

	if err := w.WriteMapHeader(2); err != nil {
		return err
	}
	if err := w.WriteString("m"); err != nil {
		return err
	}
	if err := w.WriteString(g.name); err != nil {
		return err
	}
	if err := w.WriteString("columns"); err != nil {
		return err
	}
	if err := w.WriteMapHeader(uint32(len(g.columns)) + 1); err != nil {
		return err
	}

	for i, name := range g.columns {
		if err := w.WriteString(name); err != nil {
			return err
		}
		if err := w.WriteArrayHeader(rows); err != nil {
			return err
		}
		var row int
		for _, c := range g.cells[g.starts[i]:g.starts[i+1]] {
			if err := writeNils(w, nils, c.row-row); err != nil {
				return err
			}
			var err error
			if c.tag != nil {
				err = w.WriteString(c.tag.Value)
			} else {
				err = w.WriteIntf(c.field.Value)
			}
			if err != nil {
				return err
			}
			row = c.row + 1
		}
		if err := writeNils(w, nils, len(g.metrics)-row); err != nil {
			return err
		}
	}

	if err := w.WriteString(timeColumn); err != nil {
		return err
	}
	if err := w.WriteArrayHeader(rows); err != nil {
		return err
	}
	for _, m := range g.metrics {
		if err := w.WriteInt64(m.Time().UnixMilli()); err != nil {
			return err
		}
	}

	return nil
}

// writeNils writes n MessagePack nil markers.
func writeNils(w *msgp.Writer, nils []byte, n int) error {
	for n > 0 {
		k := min(n, len(nils))
		if _, err := w.Write(nils[:k]); err != nil {
			return err
		}
		n -= k
	}
	return nil
}
