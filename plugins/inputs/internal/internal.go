//go:generate ../../../tools/readme_config_includer/generator
package internal

import (
	"crypto/tls"
	_ "embed"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	inter "github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/selfstat"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

type Self struct {
	CollectMemstats bool
	ServiceAddress  string `toml:"service_address"`
	bytesRecv       selfstat.Stat
	tlsint.ServerConfig
	acc            telegraf.Accumulator
	server         http.Server
	listener       net.Listener
	port           int
	startTime      time.Time
	timeFunc       influx.TimeFunc
	Log            telegraf.Logger `toml:"-"`
	requestsServed selfstat.Stat
	requestsRecv   selfstat.Stat
	mux            http.ServeMux
}

func NewSelf() telegraf.Input {
	return &Self{
		CollectMemstats: true,
	}
}

func (*Self) SampleConfig() string {
	return sampleConfig
}

func (s *Self) Gather(acc telegraf.Accumulator) error {
	if s.CollectMemstats {
		m := &runtime.MemStats{}
		runtime.ReadMemStats(m)
		fields := map[string]interface{}{
			"alloc_bytes":       m.Alloc,      // bytes allocated and not yet freed
			"total_alloc_bytes": m.TotalAlloc, // bytes allocated (even if freed)
			"sys_bytes":         m.Sys,        // bytes obtained from system (sum of XxxSys below)
			"pointer_lookups":   m.Lookups,    // number of pointer lookups
			"mallocs":           m.Mallocs,    // number of mallocs
			"frees":             m.Frees,      // number of frees
			// Main allocation heap statistics.
			"heap_alloc_bytes":    m.HeapAlloc,    // bytes allocated and not yet freed (same as Alloc above)
			"heap_sys_bytes":      m.HeapSys,      // bytes obtained from system
			"heap_idle_bytes":     m.HeapIdle,     // bytes in idle spans
			"heap_in_use_bytes":   m.HeapInuse,    // bytes in non-idle span
			"heap_released_bytes": m.HeapReleased, // bytes released to the OS
			"heap_objects":        m.HeapObjects,  // total number of allocated objects
			"num_gc":              m.NumGC,
		}
		acc.AddFields("internal_memstats", fields, map[string]string{})
	}

	telegrafVersion := inter.Version()
	goVersion := strings.TrimPrefix(runtime.Version(), "go")

	for _, m := range selfstat.Metrics() {
		if m.Name() == "internal_agent" {
			m.AddTag("go_version", goVersion)
		}
		m.AddTag("version", telegrafVersion)
		acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
	}

	return nil
}

func (m *Self) Init() error {
	tags := map[string]string{
		"address": m.ServiceAddress,
	}
	m.bytesRecv = selfstat.Register("internal", "bytes_received", tags)
	return nil
}

func (m *Self) Start(acc telegraf.Accumulator) error {
	m.acc = acc

	tlsConf, err := m.ServerConfig.TLSConfig()
	if err != nil {
		return err
	}

	m.server = http.Server{
		Addr:      m.ServiceAddress,
		Handler:   m,
		TLSConfig: tlsConf,
	}

	var listener net.Listener
	if tlsConf != nil {
		listener, err = tls.Listen("tcp", m.ServiceAddress, tlsConf)
		if err != nil {
			return err
		}
	} else {
		listener, err = net.Listen("tcp", m.ServiceAddress)
		if err != nil {
			return err
		}
	}
	m.listener = listener
	m.port = listener.Addr().(*net.TCPAddr).Port

	go func() {
		err = m.server.Serve(m.listener)
		if err != http.ErrServerClosed {
			m.Log.Infof("Error serving HTTP on %s", m.ServiceAddress)
		}
	}()

	m.startTime = m.timeFunc()

	m.Log.Infof("Started HTTP listener service on %s", m.ServiceAddress)

	return nil
}

func (m *Self) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	m.requestsRecv.Incr(1)
	m.mux.ServeHTTP(res, req)
	m.requestsServed.Incr(1)
}

func init() {
	inputs.Add("internal", NewSelf)
}
