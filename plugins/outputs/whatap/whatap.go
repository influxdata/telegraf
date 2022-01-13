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

	whatap_io "github.com/whatap/go-api/common/io"
	whatap_pack "github.com/whatap/go-api/common/lang/pack"
	whatap_hash "github.com/whatap/go-api/common/util/hash"
)

const (
	NetSrcAgentOneway = 10
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
	dest    int
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
	w.dest++
	if w.dest >= len(w.Servers) {
		w.dest = 0
	}

	u, err := url.Parse(w.Servers[w.dest])
	if err != nil {
		return fmt.Errorf("invalid address: %s", w.Servers[w.dest])
	}
	if u.Scheme != "tcp" {
		return fmt.Errorf("only tcp is supported: %s", w.Servers[w.dest])
	}
	client, err := net.DialTimeout(u.Scheme, u.Host, time.Duration(w.Timeout))
	if err != nil {
		return err
	}

	w.conn = client.(*net.TCPConn)
	w.Log.Info("Connected ", u.String())
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
			_ = w.Close()
			return err
		}
	}
	// Transform telegraf metrics to whatap protocol.
	for _, m := range metrics {
		p := whatap_pack.NewTagCountPack()
		p.Pcode = w.Pcode
		p.Oid = w.oid
		p.Category = fmt.Sprintf("%s%s", "telegraf_", m.Name())
		for k, v := range m.Fields() {
			p.Put(k, v)
		}
		for k, v := range m.Tags() {
			p.PutTag(k, v)
		}
		p.PutTag("oname", w.oname)

		// Convert time to microseconds.
		p.Time = m.Time().UnixNano() / int64(time.Millisecond)

		dout := whatap_io.NewDataOutputX()
		dout.WriteShort(p.GetPackType())
		p.Write(dout)

		if err := w.send(dout.ToByteArray()); err != nil {
			return err
		}
	}
	return nil
}
func (w *Whatap) send(b []byte) (err error) {
	// Transmits data in compliance with whatap protocol.
	dout := whatap_io.NewDataOutputX()
	dout.WriteByte(NetSrcAgentOneway)
	dout.WriteByte(0)
	dout.WriteLong(w.Pcode)
	dout.WriteLong(whatap_hash.Hash64Str(w.License))
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
			_ = w.Close()
			return err
		}
		nbyteleft -= nbytethistime
		pos += nbytethistime
	}
	return err
}
func init() {
	hostname, _ := os.Hostname()
	outputs.Add("whatap", func() telegraf.Output {
		return &Whatap{
			Timeout: config.Duration(60 * time.Second),
			oname:   hostname,
			oid:     whatap_hash.HashStr(hostname),
		}
	})
}
