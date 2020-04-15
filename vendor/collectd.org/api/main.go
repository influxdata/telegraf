// Package api defines data types representing core collectd data types.
package api // import "collectd.org/api"

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// Value represents either a Gauge or a Derive. It is Go's equivalent to the C
// union value_t. If a function accepts a Value, you may pass in either a Gauge
// or a Derive. Passing in any other type may or may not panic.
type Value interface {
	Type() string
}

// Gauge represents a gauge metric value, such as a temperature.
// This is Go's equivalent to the C type "gauge_t".
type Gauge float64

// Type returns "gauge".
func (v Gauge) Type() string { return "gauge" }

// Derive represents a counter metric value, such as bytes sent over the
// network. When the counter wraps around (overflows) or is reset, this is
// interpreted as a (huge) negative rate, which is discarded.
// This is Go's equivalent to the C type "derive_t".
type Derive int64

// Type returns "derive".
func (v Derive) Type() string { return "derive" }

// Counter represents a counter metric value, such as bytes sent over the
// network. When a counter value is smaller than the previous value, a wrap
// around (overflow) is assumed. This causes huge spikes in case a counter is
// reset. Only use Counter for very specific cases. If in doubt, use Derive
// instead.
// This is Go's equivalent to the C type "counter_t".
type Counter uint64

// Type returns "counter".
func (v Counter) Type() string { return "counter" }

// Identifier identifies one metric.
type Identifier struct {
	Host                   string
	Plugin, PluginInstance string
	Type, TypeInstance     string
}

// ParseIdentifier parses the identifier encoded in s and returns it.
func ParseIdentifier(s string) (Identifier, error) {
	fields := strings.Split(s, "/")
	if len(fields) != 3 {
		return Identifier{}, fmt.Errorf("not a valid identifier: %q", s)
	}

	id := Identifier{
		Host:   fields[0],
		Plugin: fields[1],
		Type:   fields[2],
	}

	if i := strings.Index(id.Plugin, "-"); i != -1 {
		id.PluginInstance = id.Plugin[i+1:]
		id.Plugin = id.Plugin[:i]
	}

	if i := strings.Index(id.Type, "-"); i != -1 {
		id.TypeInstance = id.Type[i+1:]
		id.Type = id.Type[:i]
	}

	return id, nil
}

// ValueList represents one (set of) data point(s) of one metric. It is Go's
// equivalent of the C type value_list_t.
type ValueList struct {
	Identifier
	Time     time.Time
	Interval time.Duration
	Values   []Value
	DSNames  []string
}

// DSName returns the name of the data source at the given index. If vl.DSNames
// is nil, returns "value" if there is a single value and a string
// representation of index otherwise.
func (vl *ValueList) DSName(index int) string {
	if vl.DSNames != nil {
		return vl.DSNames[index]
	} else if len(vl.Values) != 1 {
		return strconv.FormatInt(int64(index), 10)
	}

	return "value"
}

// Writer are objects accepting a ValueList for writing, for example to the
// network.
type Writer interface {
	Write(context.Context, *ValueList) error
}

// String returns a string representation of the Identifier.
func (id Identifier) String() string {
	str := id.Host + "/" + id.Plugin
	if id.PluginInstance != "" {
		str += "-" + id.PluginInstance
	}
	str += "/" + id.Type
	if id.TypeInstance != "" {
		str += "-" + id.TypeInstance
	}
	return str
}

// Dispatcher implements a multiplexer for Writer, i.e. each ValueList
// written to it is copied and written to each registered Writer.
type Dispatcher struct {
	writers []Writer
}

// Add adds a Writer to the Dispatcher.
func (d *Dispatcher) Add(w Writer) {
	d.writers = append(d.writers, w)
}

// Len returns the number of Writers belonging to the Dispatcher.
func (d *Dispatcher) Len() int {
	return len(d.writers)
}

// Write starts a new Goroutine for each Writer which creates a copy of the
// ValueList and then calls the Writer with the copy. It returns nil
// immediately.
func (d *Dispatcher) Write(ctx context.Context, vl *ValueList) error {
	for _, w := range d.writers {
		go func(w Writer) {
			vlCopy := vl
			vlCopy.Values = make([]Value, len(vl.Values))
			copy(vlCopy.Values, vl.Values)

			if err := w.Write(ctx, vlCopy); err != nil {
				log.Printf("%T.Write(): %v", w, err)
			}
		}(w)
	}
	return nil
}
