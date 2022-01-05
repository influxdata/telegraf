package whatap

import (
	"fmt"
	"net"
	"os"
	"strings"
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
	Oname   string
	Oid     int32
	Session TCPSession
	Log     telegraf.Logger `toml:"-"`
}

type TCPSession struct {
	Client net.Conn
	Dest   int
}

var sampleConfig = `
  ## You can create a project on the WhaTap site(https://www.whatap.io) 
  ## to get license, project code and server IP information.

  ## WhaTap license. Required
  #license = "xxxx-xxxx-xxxx"

  ## WhaTap project code. Required
  #pcode = 1111

  ## WhaTap server IP. Required
  ## Put multiple IPs. ["tcp://1.1.1.1:6600","tcp://2.2.2.2:6600"]
  #servers = ["tcp://1.1.1.1:6600","tcp://2.2.2.2:6600"]

  ## Connection timeout.
  # timeout = "60s"
`

func (w *Whatap) Connect() error {
	w.Session.Dest++
	if w.Session.Dest >= len(w.Servers) {
		w.Session.Dest = 0
	}

	addr := strings.SplitN(w.Servers[w.Session.Dest], "://", 2)
	if len(addr) != 2 {
		return fmt.Errorf("invalid address: %s", w.Servers[w.Session.Dest])
	}
	if addr[0] != "tcp" {
		return fmt.Errorf("only tcp is supported: %s", w.Servers[w.Session.Dest])
	}

	t := w.Timeout * time.Millisecond
	client, err := net.DialTimeout(addr[0], addr[1], t)
	if err != nil {
		return err
	}

	w.Session.Client = client.(*net.TCPConn)
	w.Log.Info("Connected tcp to ", addr)
	return nil
}

func (w *Whatap) Close() error {
	if w.Session.Client == nil {
		return nil
	}
	err := w.Session.Client.Close()
	w.Session.Client = nil
	return err
}

func (w *Whatap) Description() string {
	return "Configuration for WhaTap"
}

func (w *Whatap) SampleConfig() string {
	return sampleConfig
}

func (w *Whatap) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	if w.Session.Client == nil {
		if err := w.Connect(); err != nil {
			_ = w.Close()
			return err
		}
	}
	// Transform telegraf metrics to whatap protocol.
	for _, m := range metrics {
		p := whatap_pack.NewTagCountPack()
		p.Pcode = w.Pcode
		p.Oid = w.Oid
		p.Category = fmt.Sprintf("%s%s", "telegraf_", m.Name())
		for k, v := range m.Fields() {
			p.Put(k, v)
		}
		for k, v := range m.Tags() {
			p.PutTag(k, v)
		}
		p.PutTag("oname", w.Oname)

		// Convert time to microseconds.
		p.Time = m.Time().UnixNano() / int64(time.Millisecond)

		dout := whatap_io.NewDataOutputX()
		dout.WriteShort(p.GetPackType())
		p.Write(dout)

		if err := w.send(0, dout.ToByteArray()); err != nil {
			return err
		}
	}
	return nil
}
func (w *Whatap) send(code byte, b []byte) (err error) {
	// Transmits data in compliance with whatap protocol.
	dout := whatap_io.NewDataOutputX()
	dout.WriteByte(NetSrcAgentOneway)
	dout.WriteByte(code)
	dout.WriteLong(w.Pcode)
	dout.WriteLong(whatap_hash.Hash64Str(w.License))
	dout.WriteIntBytes(b)
	sendbuf := dout.ToByteArray()

	nbyteleft := len(sendbuf)
	pos := 0
	for 0 < nbyteleft {
		nbytethistime := 0
		// Set Deadline
		err = w.Session.Client.SetWriteDeadline(time.Now().Add(
			w.Timeout * time.Millisecond))
		if err != nil {
			w.Log.Warn("cannot set tcp write deadline:", err)
		}
		nbytethistime, err = w.Session.Client.Write(sendbuf[pos : pos+nbyteleft])
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
			Timeout: 60 * time.Second,
			Session: TCPSession{},
			Oname:   hostname,
			Oid:     whatap_hash.HashStr(hostname),
		}
	})
}
