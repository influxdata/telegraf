package format // import "collectd.org/format"

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"collectd.org/api"
)

// Graphite implements the Writer interface and writes ValueLists in Graphite
// format to W.
type Graphite struct {
	W                 io.Writer
	Prefix, Suffix    string
	EscapeChar        string
	SeparateInstances bool
	AlwaysAppendDS    bool // TODO(octo): Implement support.
	replacer          *strings.Replacer
}

func (g *Graphite) escape(in string) string {
	if g.replacer == nil {
		g.replacer = strings.NewReplacer(
			".", g.EscapeChar,
			"\t", g.EscapeChar,
			"\"", g.EscapeChar,
			"\\", g.EscapeChar,
			":", g.EscapeChar,
			"!", g.EscapeChar,
			"/", g.EscapeChar,
			"(", g.EscapeChar,
			")", g.EscapeChar,
			"\n", g.EscapeChar,
			"\r", g.EscapeChar)
	}
	return g.replacer.Replace(in)
}

func (g *Graphite) formatName(id api.Identifier, dsName string) string {
	var instanceSeparator = "-"
	if g.SeparateInstances {
		instanceSeparator = "."
	}

	host := g.escape(id.Host)
	plugin := g.escape(id.Plugin)
	if id.PluginInstance != "" {
		plugin += instanceSeparator + g.escape(id.PluginInstance)
	}

	typ := id.Type
	if id.TypeInstance != "" {
		typ += instanceSeparator + g.escape(id.TypeInstance)
	}

	name := g.Prefix + host + g.Suffix + "." + plugin + "." + typ
	if dsName != "" {
		name += "." + g.escape(dsName)
	}

	return name
}

func (g *Graphite) formatValue(v api.Value) (string, error) {
	switch v := v.(type) {
	case api.Gauge:
		return fmt.Sprintf("%.15g", v), nil
	case api.Derive, api.Counter:
		return fmt.Sprintf("%v", v), nil
	default:
		return "", fmt.Errorf("unexpected type %T", v)
	}
}

// Write formats the ValueList in the PUTVAL format and writes it to the
// assiciated io.Writer.
func (g *Graphite) Write(_ context.Context, vl *api.ValueList) error {
	for i, v := range vl.Values {
		dsName := ""
		if g.AlwaysAppendDS || len(vl.Values) != 1 {
			dsName = vl.DSName(i)
		}

		name := g.formatName(vl.Identifier, dsName)

		val, err := g.formatValue(v)
		if err != nil {
			return err
		}

		t := vl.Time
		if t.IsZero() {
			t = time.Now()
		}

		fmt.Fprintf(g.W, "%s %s %d\r\n", name, val, t.Unix())
	}

	return nil
}
