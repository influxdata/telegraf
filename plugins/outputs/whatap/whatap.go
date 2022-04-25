package whatap

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"

	wio "github.com/whatap/go-api/common/io"
	wpack "github.com/whatap/go-api/common/lang/pack"
	whash "github.com/whatap/go-api/common/util/hash"
)

const (
	netSrcAgentOneway  = 10
	netSrcAgentVersion = 0
	netScheme          = "tcp"
)

type Whatap struct {
	License string          `toml:"license"`
	Servers []string        `toml:"servers"`
	Pcode   int64           `toml:"pcode"`
	Timeout config.Duration `toml:"timeout"`
	Log     telegraf.Logger `toml:"-"`
	oname   string
	oid     int32
	conn    net.Conn
	hosts   []string
}

const sampleConfig = `
  ## You can create a project on the WhaTap site(https://www.whatap.io) 
  ## to get license, project code and server IP information.

  ## WhaTap license. Required
  license = "xxxx-xxxx-xxxx"

  ## WhaTap project code. Required
  pcode = 1111

  ## WhaTap server IP. Required
  ## Put multiple IPs. ["tcp://1.1.1.1:6600","tcp://2.2.2.2:6600"]
  servers = ["tcp://1.1.1.1:6600","tcp://2.2.2.2:6600"]

  ## Connection timeout.
  # timeout = "60s"
`

func (w *Whatap) Connect() error {
	// Change and connect multiple servers sequentially.
	for _, host := range w.hosts {
		client, err := net.DialTimeout(netScheme, host, time.Duration(w.Timeout))
		if err != nil {
			w.Log.Errorf("connecting to %q failed: %v", host, err)
			continue
		}
		w.conn = client.(*net.TCPConn)
		w.Log.Info("Connected ", host)
		return nil
	}
	return fmt.Errorf("could not connect to any server")
}

func (w *Whatap) Close() error {
	if w.conn == nil {
		return nil
	}
	err := w.conn.Close()
	w.conn = nil
	return err
}

func (w *Whatap) Description() string {
	return "Plugin to send metrics to a WhaTap server"
}

func (w *Whatap) SampleConfig() string {
	return sampleConfig
}

func (w *Whatap) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	if w.conn == nil {
		if err := w.Connect(); err != nil {
			return err
		}
	}
	// Transform telegraf metrics to whatap data.
	for _, m := range metrics {
		p := wpack.NewTagCountPack()
		p.Pcode = w.Pcode
		p.Oid = w.oid
		p.Category = "telegraf_" + m.Name()
		for k, v := range m.Fields() {
			p.Put(k, v)
		}
		for k, v := range m.Tags() {
			p.PutTag(k, v)
		}
		p.PutTag("oname", w.oname)

		// Convert time to microseconds.
		p.Time = m.Time().UnixNano() / int64(time.Millisecond)

		dout := wio.NewDataOutputX()
		dout.WriteShort(p.GetPackType())
		p.Write(dout)
		dout.WriteHeader(netSrcAgentOneway, netSrcAgentVersion, w.Pcode,
			whash.Hash64Str(w.License))

		if err := w.send(dout.ToByteArray()); err != nil {
			w.Log.Warnf("cannot send data: %v", err)
			_ = w.Close()
			return err
		}
	}
	return nil
}
func (w *Whatap) send(sendbuf []byte) (err error) {
	for pos := 0; pos < len(sendbuf); {
		deadline := time.Now().Add(time.Duration(w.Timeout))
		if err := w.conn.SetWriteDeadline(deadline); err != nil {
			return fmt.Errorf("cannot set write deadline: %v", err)
		}
		nbytethistime, err := w.conn.Write(sendbuf[pos:])
		if err != nil {
			return err
		}
		pos += nbytethistime
	}
	return nil
}
func (w *Whatap) Init() error {
	w.hosts = make([]string, 0)
	for _, server := range w.Servers {
		u, err := url.Parse(server)
		if err != nil {
			w.Log.Errorf("invalid address: %s", server)
			continue
		}
		if u.Scheme != "tcp" {
			w.Log.Errorf("only tcp is supported: %s", server)
			continue
		}
		w.hosts = append(w.hosts, u.Host)
	}
	if len(w.hosts) == 0 {
		return fmt.Errorf("no WhaTap server IP configured")
	}

	hn, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %v", err)
	}
	w.oname = hn
	w.oid = whash.HashStr(w.oname)
	return nil
}
func init() {
	outputs.Add("whatap", func() telegraf.Output {
		return &Whatap{
			Timeout: config.Duration(60 * time.Second),
		}
	})
}
