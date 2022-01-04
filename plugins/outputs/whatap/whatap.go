package whatap

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	whatap_io "github.com/whatap/go-api/common/io"
	whatap_pack "github.com/whatap/go-api/common/lang/pack"
	whatap_hash "github.com/whatap/go-api/common/util/hash"
)

const (
	NETSRC_AGENT_ONEWAY = 10
)

type Whatap struct {
	License string        `toml:"license"`
	Servers []string      `toml:"servers"`
	Pcode   int64         `toml:"pcode"`
	Timeout time.Duration `toml:"timeout"`
	Oname   string
	Oid     int32
	Session TcpSession
}

type TcpSession struct {
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
	//log.Println("[outputs.whatap] Connect")
	w.Session.Dest += 1
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
	if client, err := net.DialTimeout(addr[0], addr[1], t); err != nil {
		return err
	} else {
		w.Session.Client = client.(*net.TCPConn)
		log.Println("I! [outputs.whatap] Connected tcp to ", addr)
	}

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
			w.Close()
			return err
		}
	}
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
		// Add Oname
		p.PutTag("oname", w.Oname)
		p.Time = m.Time().UnixNano() / int64(time.Millisecond)

		dout := whatap_io.NewDataOutputX()
		// pack type
		dout.WriteShort(p.GetPackType())
		p.Write(dout)

		if err := w.send(0, dout.ToByteArray()); err != nil {
			return err
		}
	}
	return nil
}
func (w *Whatap) send(code byte, b []byte) (err error) {
	dout := whatap_io.NewDataOutputX()
	dout.WriteByte(NETSRC_AGENT_ONEWAY)
	// ver
	dout.WriteByte(code)
	dout.WriteLong(w.Pcode)
	dout.WriteLong(whatap_hash.Hash64Str(w.License))
	dout.WriteIntBytes(b)
	// pack data
	sendbuf := dout.ToByteArray()

	nbyteleft := len(sendbuf)
	pos := 0
	for 0 < nbyteleft {
		nbytethistime := 0
		// Set Deadline
		err = w.Session.Client.SetWriteDeadline(time.Now().Add(
			w.Timeout * time.Millisecond))
		if err != nil {
			log.Println("W! [outputs.whatap] cannot set tcp write deadline:",
				err)
		}
		nbytethistime, err = w.Session.Client.Write(sendbuf[pos : pos+nbyteleft])
		if err != nil {
			w.Close()
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
			Session: TcpSession{},
			Oname:   hostname,
			Oid:     whatap_hash.HashStr(hostname),
		}

	})

}
