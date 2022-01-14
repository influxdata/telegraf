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
	NetSrcAgentOneway  = 10
	NetSrcAgentVersion = 0
	Scheme             = "tcp"
)

type Whatap struct {
	License string          `toml:"license"`
	Servers []string        `toml:"servers"`
	Pcode   int64           `toml:"project_code"`
	Timeout config.Duration `toml:"timeout"`
	Log     telegraf.Logger `toml:"-"`
	oname   string
	oid     int32
	conn    net.Conn
	dest    int
	hosts   []string
}

const sampleConfig = `
  ## You can create a project on the WhaTap site(https://www.whatap.io) 
  ## to get license, project code and server IP information.

  ## WhaTap license. Required
  license = "xxxx-xxxx-xxxx"

  ## WhaTap project code. Required
  project_code = 1111

  ## WhaTap server IP. Required
  ## Put multiple IPs. ["tcp://1.1.1.1:6600","tcp://2.2.2.2:6600"]
  servers = ["tcp://1.1.1.1:6600","tcp://2.2.2.2:6600"]

  ## Connection timeout.
  # timeout = "60s"
`

func (w *Whatap) Connect() error {
	// Change and connect multiple servers sequentially.
	host := w.nextHost()
	client, err := net.DialTimeout(Scheme, host, time.Duration(w.Timeout))
	if err != nil {
		return err
	}
	w.conn = client.(*net.TCPConn)
	w.Log.Info("Connected ", host)
	return nil
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

		if err := w.send(dout.ToByteArray()); err != nil {
			w.Log.Warnf("cannot send data: %v", err)
			_ = w.Close()
		}
	}
	return nil
}
func (w *Whatap) send(b []byte) (err error) {
	dout := wio.NewDataOutputX()
	// Header : add the header before sending
	dout.WriteByte(NetSrcAgentOneway)
	dout.WriteByte(NetSrcAgentVersion)
	dout.WriteLong(w.Pcode)
	dout.WriteLong(whash.Hash64Str(w.License))
	// Body : Set the converted whatap data(b []bytes) as the body
	dout.WriteIntBytes(b)
	sendbuf := dout.ToByteArray()

	nbyteleft := len(sendbuf)
	var pos int
	var nbytethistime int
	for 0 < nbyteleft {
		deadline := time.Now().Add(time.Duration(w.Timeout))
		if err := w.conn.SetWriteDeadline(deadline); err != nil {
			return fmt.Errorf("cannot set write deadline: %v", err)
		}

		nbytethistime, err = w.conn.Write(sendbuf[pos : pos+nbyteleft])
		if err != nil {
			return err
		}
		nbyteleft -= nbytethistime
		pos += nbytethistime
	}
	return nil
}
func (w *Whatap) nextHost() string {
	w.dest++
	if w.dest >= len(w.Servers) {
		w.dest = 0
	}
	return w.hosts[w.dest]
}
func (w *Whatap) Init() error {
	if len(w.Servers) == 0 {
		return fmt.Errorf("WhaTap server IP is Required")
	}
	w.hosts = make([]string, 0)
	for _, server := range w.Servers {
		u, err := url.Parse(server)
		if err != nil {
			return fmt.Errorf("invalid address: %s", server)
		}
		if u.Scheme != "tcp" {
			return fmt.Errorf("only tcp is supported: %s", server)
		}
		w.hosts = append(w.hosts, u.Host)
	}
	w.oname, _ = os.Hostname()
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
