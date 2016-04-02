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
	Servers []string
	Domains []string
}

func (d *Dovecot) Description() string {
	return "Read statistics from one or many dovecot servers"
}

var sampleConfig = `
  ## specify dovecot servers via an address:port list
  ##  e.g.
  ##    localhost:24242
  ##
  ## If no servers are specified, then localhost is used as the host.
  servers = ["localhost:24242"]
  ## Only collect metrics for these domains, collect all if empty
  domains = []
`

var defaultTimeout = time.Second * time.Duration(5)

func (d *Dovecot) SampleConfig() string { return sampleConfig }

const defaultPort = "24242"

// Reads stats from all configured servers.
func (d *Dovecot) Gather(acc telegraf.Accumulator) error {

	if len(d.Servers) == 0 {
		d.Servers = append(d.Servers, "127.0.0.1:24242")
	}

	var wg sync.WaitGroup

	var outerr error

	var domains = make(map[string]bool)

	for _, dom := range d.Domains {
		domains[dom] = true
	}

	for _, serv := range d.Servers {
		wg.Add(1)
		go func(serv string) {
			defer wg.Done()
			outerr = d.gatherServer(serv, acc, domains)
		}(serv)
	}

	wg.Wait()

	return outerr
}

func (d *Dovecot) gatherServer(addr string, acc telegraf.Accumulator, doms map[string]bool) error {
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("Error: %s on url %s\n", err, addr)
	}

	c, err := net.DialTimeout("tcp", addr, defaultTimeout)
	if err != nil {
		return fmt.Errorf("Unable to connect to dovecot server '%s': %s", addr, err)
	}
	defer c.Close()

	// Extend connection
	c.SetDeadline(time.Now().Add(defaultTimeout))

	c.Write([]byte("EXPORT\tdomain\n\n"))
	var buf bytes.Buffer
	io.Copy(&buf, c)
	//	buf := bufio.NewReader(c)

	host, _, _ := net.SplitHostPort(addr)

	return gatherStats(&buf, acc, doms, host)
}

func gatherStats(buf *bytes.Buffer, acc telegraf.Accumulator, doms map[string]bool, host string) error {

	lines := strings.Split(buf.String(), "\n")
	head := strings.Split(lines[0], "\t")
	vals := lines[1:]

	for i := range vals {
		if vals[i] == "" {
			continue
		}
		val := strings.Split(vals[i], "\t")
		fields := make(map[string]interface{})
		if len(doms) > 0 && !doms[val[0]] {
			continue
		}
		tags := map[string]string{"server": host, "domain": val[0]}
		for n := range val {
			switch head[n] {
			case "domain":
				continue
				//				fields[head[n]] = val[n]
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
