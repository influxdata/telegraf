package dovecot

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Dovecot struct {
	Type    string
	Filters []string
	Servers []string
}

var defaultTimeout = time.Second * time.Duration(5)

var validQuery = map[string]bool{
	"user": true, "domain": true, "global": true, "ip": true,
}

// Reads stats from all configured servers.
func (d *Dovecot) Gather(acc telegraf.Accumulator) error {
	if !validQuery[d.Type] {
		return fmt.Errorf("error: %s is not a valid query type", d.Type)
	}

	if len(d.Servers) == 0 {
		d.Servers = append(d.Servers, "127.0.0.1:24242")
	}

	if len(d.Filters) <= 0 {
		d.Filters = append(d.Filters, "")
	}

	var wg sync.WaitGroup
	for _, server := range d.Servers {
		for _, filter := range d.Filters {
			wg.Add(1)
			go func(s string, f string) {
				defer wg.Done()
				acc.AddError(d.gatherServer(s, acc, d.Type, f))
			}(server, filter)
		}
	}

	wg.Wait()
	return nil
}

func (d *Dovecot) gatherServer(addr string, acc telegraf.Accumulator, qtype string, filter string) error {
	var proto string

	if strings.HasPrefix(addr, "/") {
		proto = "unix"
	} else {
		proto = "tcp"

		_, _, err := net.SplitHostPort(addr)
		if err != nil {
			return fmt.Errorf("%q on url %s", err.Error(), addr)
		}
	}

	c, err := net.DialTimeout(proto, addr, defaultTimeout)
	if err != nil {
		return fmt.Errorf("unable to connect to dovecot server '%s': %s", addr, err)
	}
	defer c.Close()

	// Extend connection
	if err := c.SetDeadline(time.Now().Add(defaultTimeout)); err != nil {
		return fmt.Errorf("setting deadline failed for dovecot server '%s': %s", addr, err)
	}

	msg := fmt.Sprintf("EXPORT\t%s", qtype)
	if len(filter) > 0 {
		msg += fmt.Sprintf("\t%s=%s", qtype, filter)
	}
	msg += "\n"

	if _, err := c.Write([]byte(msg)); err != nil {
		return fmt.Errorf("writing message %q failed for dovecot server '%s': %s", msg, addr, err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, c); err != nil {
		// We need to accept the timeout here as reading from the connection will only terminate on EOF
		// or on a timeout to happen. As EOF for TCP connections will only be sent on connection closing,
		// the only way to get the whole message is to wait for the timeout to happen.
		if nerr, ok := err.(net.Error); !ok || !nerr.Timeout() {
			return fmt.Errorf("copying message failed for dovecot server '%s': %s", addr, err)
		}
	}

	var host string
	if strings.HasPrefix(addr, "/") {
		host = addr
	} else {
		host, _, _ = net.SplitHostPort(addr)
	}

	return gatherStats(&buf, acc, host, qtype)
}

func gatherStats(buf *bytes.Buffer, acc telegraf.Accumulator, host string, qtype string) error {
	lines := strings.Split(buf.String(), "\n")
	head := strings.Split(lines[0], "\t")
	vals := lines[1:]

	for i := range vals {
		if vals[i] == "" {
			continue
		}
		val := strings.Split(vals[i], "\t")

		fields := make(map[string]interface{})
		tags := map[string]string{"server": host, "type": qtype}

		if qtype != "global" {
			tags[qtype] = val[0]
		}

		for n := range val {
			switch head[n] {
			case qtype:
				continue
			case "user_cpu", "sys_cpu", "clock_time":
				fields[head[n]] = secParser(val[n])
			case "reset_timestamp", "last_update":
				fields[head[n]] = timeParser(val[n])
			default:
				ival, _ := splitSec(val[n])
				fields[head[n]] = ival
			}
		}

		acc.AddFields("dovecot", fields, tags)
	}

	return nil
}

func splitSec(tm string) (sec int64, msec int64) {
	var err error
	ss := strings.Split(tm, ".")

	sec, err = strconv.ParseInt(ss[0], 10, 64)
	if err != nil {
		sec = 0
	}
	if len(ss) > 1 {
		msec, err = strconv.ParseInt(ss[1], 10, 64)
		if err != nil {
			msec = 0
		}
	} else {
		msec = 0
	}

	return sec, msec
}

func timeParser(tm string) time.Time {
	sec, msec := splitSec(tm)
	return time.Unix(sec, msec)
}

func secParser(tm string) float64 {
	sec, msec := splitSec(tm)
	return float64(sec) + (float64(msec) / 1000000.0)
}

func init() {
	inputs.Add("dovecot", func() telegraf.Input {
		return &Dovecot{}
	})
}
