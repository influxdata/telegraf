package laravel_redis_queue

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/errchan"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type LaravelRedisQueue struct {
	Servers []string
	Queues  []string
}

var sampleConfig = `
  ## specify servers via a url matching:
  ##  [protocol://][:password]@address[:port]
  ##  e.g.
  ##    tcp://localhost:6379
  ##    tcp://:password@192.168.99.100
  ##    unix:///var/run/redis.sock
  ##
  ## If no servers are specified, then localhost is used as the host.
  ## If no port is specified, 6379 is used
  servers = ["tcp://localhost:6379"]

	## specify queues:
	##  [queue_name]
	##  e.g.
	##    queue1
	queues = ["queue1", "queue2"]
`

var defaultTimeout = 5 * time.Second

func (r *LaravelRedisQueue) SampleConfig() string {
	return sampleConfig
}

func (r *LaravelRedisQueue) Description() string {
	return "Read laravel queues count from one or many redis servers"
}

var ErrProtocolError = errors.New("redis protocol error")

const defaultPort = "6379"

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (r *LaravelRedisQueue) Gather(acc telegraf.Accumulator) error {
	if len(r.Servers) == 0 {
		url := &url.URL{
			Scheme: "tcp",
			Host:   ":6379",
		}
		r.gatherServer(url, acc)
		return nil
	}

	var wg sync.WaitGroup
	errChan := errchan.New(len(r.Servers))
	for _, serv := range r.Servers {
		if !strings.HasPrefix(serv, "tcp://") && !strings.HasPrefix(serv, "unix://") {
			serv = "tcp://" + serv
		}

		u, err := url.Parse(serv)
		if err != nil {
			return fmt.Errorf("Unable to parse to address '%s': %s", serv, err)
		} else if u.Scheme == "" {
			// fallback to simple string based address (i.e. "10.0.0.1:10000")
			u.Scheme = "tcp"
			u.Host = serv
			u.Path = ""
		}
		if u.Scheme == "tcp" {
			_, _, err := net.SplitHostPort(u.Host)
			if err != nil {
				u.Host = u.Host + ":" + defaultPort
			}
		}

		wg.Add(1)
		go func(serv string) {
			defer wg.Done()
			errChan.C <- r.gatherServer(u, acc)
		}(serv)
	}

	wg.Wait()
	return errChan.Error()
}

func (r *LaravelRedisQueue) gatherServer(addr *url.URL, acc telegraf.Accumulator) error {
	var address string

	if addr.Scheme == "unix" {
		address = addr.Path
	} else {
		address = addr.Host
	}
	c, err := net.DialTimeout(addr.Scheme, address, defaultTimeout)
	if err != nil {
		return fmt.Errorf("Unable to connect to redis server '%s': %s", address, err)
	}
	defer c.Close()

	// Extend connection
	c.SetDeadline(time.Now().Add(defaultTimeout))

	if addr.User != nil {
		pwd, set := addr.User.Password()
		if set && pwd != "" {
			c.Write([]byte(fmt.Sprintf("AUTH %s\r\n", pwd)))

			rdr := bufio.NewReader(c)

			line, err := rdr.ReadString('\n')
			if err != nil {
				return err
			}
			if line[0] != '+' {
				return fmt.Errorf("%s", strings.TrimSpace(line)[1:])
			}
		}
	}

	for _, queue := range r.Queues {
		// Static Pushed Queues
		c.Write([]byte("LLEN queues:" + queue + ":reserved\r\n"))
		c.Write([]byte("EOF\r\n"))
		rdr := bufio.NewReader(c)

		var tags map[string]string

		if addr.Scheme == "unix" {
			tags = map[string]string{"socket": addr.Path}
		} else {
			// Setup tags for all redis metrics
			host, port := "unknown", "unknown"
			// If there's an error, ignore and use 'unknown' tags
			host, port, _ = net.SplitHostPort(addr.Host)
			tags = map[string]string{"server": host, "port": port}
		}
		_ = gatherInfoOutput(rdr, acc, tags, "pushed_count_"+queue)

		// Static Delayed Queues
		c.Write([]byte("ZCARD queues:" + queue + ":delayed\r\n"))
		c.Write([]byte("EOF\r\n"))
		rdr = bufio.NewReader(c)
		_ = gatherInfoOutput(rdr, acc, tags, "delayed_count_"+queue)

		// Static Reserved Queues
		c.Write([]byte("ZCARD queues:" + queue + ":reserved\r\n"))
		c.Write([]byte("EOF\r\n"))
		rdr = bufio.NewReader(c)
		_ = gatherInfoOutput(rdr, acc, tags, "reserved_count_"+queue)
	}

	return nil
}

// gatherInfoOutput gathers
func gatherInfoOutput(rdr *bufio.Reader, acc telegraf.Accumulator, tags map[string]string, field_name string) error {
	scanner := bufio.NewScanner(rdr)
	fields := make(map[string]interface{})
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "ERR") {
			break
		}

		if len(line) == 0 {
			continue
		}

		fields[field_name] = string(line)
	}
	acc.AddFields("laravel_redis_queue", fields, tags)
	return nil
}

func init() {
	inputs.Add("laravel_redis_queue", func() telegraf.Input {
		return &LaravelRedisQueue{}
	})
}
