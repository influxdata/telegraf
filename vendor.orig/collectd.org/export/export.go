/*
Package export provides an interface to instrument Go code.

Instrumenting Go code with this package is very similar to the "expvar" package
in the vanilla Go distribution. In fact, the variables exported with this
package are also registered with the "expvar" package, so that you can also use
other existing metric collection frameworks with it. This package differs in
that it has an explicitly cumulative type, Derive.

The intended usage pattern of this package is as follows: First, global
variables are initialized with NewDerive(), NewGauge(), NewDeriveString() or
NewGaugeString(). The Run() function is called as a separate goroutine as part
of your program's initialization and, last but not least, the variables are
updated with their respective update functions, Add() for Derive and Set() for
Gauge.

  // Initialize global variable.
  var requestCounter = export.NewDeriveString("example.com/golang/total_requests")

  // Call Run() in its own goroutine.
  func main() {
          client, err := network.Dial(
                  net.JoinHostPort(network.DefaultIPv6Address, network.DefaultService),
                  network.ClientOptions{})
          if err !=  nil {
                  log.Fatal(err)
                  }
          go export.Run(client, export.Options{
                  Interval: 10 * time.Second,
          })
          // …
  }

  // Update variable.
  func requestHandler(w http.ResponseWriter, req *http.Request) {
          defer requestCounter.Add(1)
          // …
  }
*/
package export // import "collectd.org/export"

import (
	"context"
	"expvar"
	"log"
	"math"
	"strconv"
	"sync"
	"time"

	"collectd.org/api"
)

var (
	mutex sync.RWMutex
	vars  []Var
)

// Var is an abstract type for metrics exported by this package.
type Var interface {
	ValueList() *api.ValueList
}

// Publish adds v to the internal list of exported metrics.
func Publish(v Var) {
	mutex.Lock()
	defer mutex.Unlock()

	vars = append(vars, v)
}

// Options holds options for the Run() function.
type Options struct {
	Interval time.Duration
}

// Run periodically calls the ValueList function of each Var, sets the Time and
// Interval fields and passes it w.Write(). This function blocks indefinitely.
func Run(ctx context.Context, w api.Writer, opts Options) error {
	ticker := time.NewTicker(opts.Interval)

	for {
		select {
		case _ = <-ticker.C:
			mutex.RLock()
			for _, v := range vars {
				vl := v.ValueList()
				vl.Time = time.Now()
				vl.Interval = opts.Interval
				if err := w.Write(ctx, vl); err != nil {
					mutex.RUnlock()
					return err
				}
			}
			mutex.RUnlock()
		}
	}
}

// Derive represents a cumulative integer data type, for example "requests
// served since server start". It implements the Var and expvar.Var interfaces.
type Derive struct {
	mu    sync.RWMutex
	id    api.Identifier
	value api.Derive
}

// NewDerive initializes a new Derive, registers it with the "expvar" package
// and returns it. The initial value is zero.
func NewDerive(id api.Identifier) *Derive {
	d := &Derive{
		id:    id,
		value: 0,
	}

	Publish(d)
	expvar.Publish(id.String(), d)
	return d
}

// NewDeriveString parses s as an Identifier and returns a new Derive. If
// parsing s fails, it will panic. This simplifies initializing global
// variables.
func NewDeriveString(s string) *Derive {
	id, err := api.ParseIdentifier(s)
	if err != nil {
		log.Fatal(err)
	}

	return NewDerive(id)
}

// Add adds diff to d.
func (d *Derive) Add(diff int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.value += api.Derive(diff)
}

// String returns the string representation of d.
func (d *Derive) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return strconv.FormatInt(int64(d.value), 10)
}

// ValueList returns the ValueList representation of d. Both, Time and Interval
// are set to zero.
func (d *Derive) ValueList() *api.ValueList {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return &api.ValueList{
		Identifier: d.id,
		Values:     []api.Value{d.value},
	}
}

// Gauge represents an absolute floating point data type, for example "heap
// memory used". It implements the Var and expvar.Var interfaces.
type Gauge struct {
	mu    sync.RWMutex
	id    api.Identifier
	value api.Gauge
}

// NewGauge initializes a new Gauge, registers it with the "expvar" package and
// returns it. The initial value is NaN.
func NewGauge(id api.Identifier) *Gauge {
	g := &Gauge{
		id:    id,
		value: api.Gauge(math.NaN()),
	}

	Publish(g)
	expvar.Publish(id.String(), g)
	return g
}

// NewGaugeString parses s as an Identifier and returns a new Gauge. If parsing
// s fails, it will panic. This simplifies initializing global variables.
func NewGaugeString(s string) *Gauge {
	id, err := api.ParseIdentifier(s)
	if err != nil {
		log.Fatal(err)
	}

	return NewGauge(id)
}

// Set sets g to v.
func (g *Gauge) Set(v float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = api.Gauge(v)
}

// String returns the string representation of g.
func (g *Gauge) String() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return strconv.FormatFloat(float64(g.value), 'g', -1, 64)
}

// ValueList returns the ValueList representation of g. Both, Time and Interval
// are set to zero.
func (g *Gauge) ValueList() *api.ValueList {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return &api.ValueList{
		Identifier: g.id,
		Values:     []api.Value{g.value},
	}
}
