//go:generate ../../../tools/readme_config_includer/generator
package zipkin

import (
	"context"
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/zipkin/trace"
)

//go:embed sample.conf
var sampleConfig string

var (
	// defaultNetwork is the network to listen on; use only in tests.
	defaultNetwork = "tcp"
)

const (
	// defaultPort is the default port zipkin listens on, which zipkin implementations expect.
	defaultPort = 9411

	// defaultRoute is the default route zipkin uses, and zipkin implementations expect.
	defaultRoute = "/api/v1/spans"

	// defaultShutdownTimeout is the max amount of time telegraf will wait for the plugin to shut down
	defaultShutdownTimeout = 5 * time.Second

	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 10 * time.Second
)

type Zipkin struct {
	Port         int             `toml:"port"`
	Path         string          `toml:"path"`
	ReadTimeout  config.Duration `toml:"read_timeout"`
	WriteTimeout config.Duration `toml:"write_timeout"`

	Log telegraf.Logger `toml:"-"`

	address   string
	handler   handler
	server    *http.Server
	waitGroup *sync.WaitGroup
}

// recorder represents a type which can record zipkin trace data as well as any accompanying errors, and process that data.
type recorder interface {
	record(trace.Trace) error
	error(error)
}

// handler represents a type which can register itself with a router for http routing, and a recorder for trace data collection.
type handler interface {
	register(router *mux.Router, recorder recorder) error
}

func (*Zipkin) SampleConfig() string {
	return sampleConfig
}

func (z *Zipkin) Start(acc telegraf.Accumulator) error {
	if z.ReadTimeout < config.Duration(time.Second) {
		z.ReadTimeout = config.Duration(defaultReadTimeout)
	}
	if z.WriteTimeout < config.Duration(time.Second) {
		z.WriteTimeout = config.Duration(defaultWriteTimeout)
	}

	z.handler = newSpanHandler(z.Path)

	var wg sync.WaitGroup
	z.waitGroup = &wg

	router := mux.NewRouter()
	converter := newLineProtocolConverter(acc)
	if err := z.handler.register(router, converter); err != nil {
		return err
	}

	z.server = &http.Server{
		Handler:      router,
		ReadTimeout:  time.Duration(z.ReadTimeout),
		WriteTimeout: time.Duration(z.WriteTimeout),
	}

	addr := ":" + strconv.Itoa(z.Port)
	ln, err := net.Listen(defaultNetwork, addr)
	if err != nil {
		return err
	}

	z.address = ln.Addr().String()
	z.Log.Infof("Started the zipkin listener on %s", z.address)

	wg.Add(1)
	go func() {
		defer wg.Done()

		z.listen(ln, acc)
	}()

	return nil
}

func (*Zipkin) Gather(telegraf.Accumulator) error { return nil }

func (z *Zipkin) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)

	defer z.waitGroup.Wait()
	defer cancel()

	z.server.Shutdown(ctx) //nolint:errcheck // Ignore the returned error as we cannot do anything about it anyway
}

// listen creates a http server on the zipkin instance it is called with, and
// serves http until it is stopped by Zipkin's (*Zipkin).Stop()  method.
func (z *Zipkin) listen(ln net.Listener, acc telegraf.Accumulator) {
	if err := z.server.Serve(ln); err != nil {
		// Because of the clean shutdown in `(*Zipkin).Stop()`
		// We're expecting a server closed error at some point
		// So we don't want to display it as an error.
		// This interferes with telegraf's internal data collection,
		// by making it appear as if a serious error occurred.
		if err != http.ErrServerClosed {
			acc.AddError(fmt.Errorf("error listening: %w", err))
		}
	}
}

func init() {
	inputs.Add("zipkin", func() telegraf.Input {
		return &Zipkin{
			Path: defaultRoute,
			Port: defaultPort,
		}
	})
}
