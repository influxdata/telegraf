package zookeeper

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var zookeeperFormatRE = regexp.MustCompile(`^zk_(\w[\w\.\-]*)\s+([\w\.\-]+)`)

// Zookeeper is a zookeeper plugin
type Zookeeper struct {
	Servers []string
	Timeout config.Duration

	EnableTLS bool `toml:"enable_tls"`
	EnableSSL bool `toml:"enable_ssl" deprecated:"1.7.0;use 'enable_tls' instead"`
	tlsint.ClientConfig

	initialized bool
	tlsConfig   *tls.Config
}

var defaultTimeout = 5 * time.Second

func (z *Zookeeper) dial(ctx context.Context, addr string) (net.Conn, error) {
	var dialer net.Dialer
	if z.EnableTLS || z.EnableSSL {
		deadline, ok := ctx.Deadline()
		if ok {
			dialer.Deadline = deadline
		}
		return tls.DialWithDialer(&dialer, "tcp", addr, z.tlsConfig)
	}
	return dialer.DialContext(ctx, "tcp", addr)
}

// Gather reads stats from all configured servers accumulates stats
func (z *Zookeeper) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()

	if !z.initialized {
		tlsConfig, err := z.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
		z.tlsConfig = tlsConfig
		z.initialized = true
	}

	if z.Timeout < config.Duration(1*time.Second) {
		z.Timeout = config.Duration(defaultTimeout)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(z.Timeout))
	defer cancel()

	if len(z.Servers) == 0 {
		z.Servers = []string{":2181"}
	}

	for _, serverAddress := range z.Servers {
		acc.AddError(z.gatherServer(ctx, serverAddress, acc))
	}
	return nil
}

func (z *Zookeeper) gatherServer(ctx context.Context, address string, acc telegraf.Accumulator) error {
	var zookeeperState string
	_, _, err := net.SplitHostPort(address)
	if err != nil {
		address = address + ":2181"
	}

	c, err := z.dial(ctx, address)
	if err != nil {
		return err
	}
	defer c.Close()

	// Apply deadline to connection
	deadline, ok := ctx.Deadline()
	if ok {
		if err := c.SetDeadline(deadline); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(c, "%s\n", "mntr"); err != nil {
		return err
	}
	rdr := bufio.NewReader(c)
	scanner := bufio.NewScanner(rdr)

	service := strings.Split(address, ":")
	if len(service) != 2 {
		return fmt.Errorf("invalid service address: %s", address)
	}

	fields := make(map[string]interface{})
	for scanner.Scan() {
		line := scanner.Text()
		parts := zookeeperFormatRE.FindStringSubmatch(line)

		if len(parts) != 3 {
			return fmt.Errorf("unexpected line in mntr response: %q", line)
		}

		measurement := strings.TrimPrefix(parts[1], "zk_")
		if measurement == "server_state" {
			zookeeperState = parts[2]
		} else {
			sValue := parts[2]

			iVal, err := strconv.ParseInt(sValue, 10, 64)
			if err == nil {
				fields[measurement] = iVal
			} else {
				fields[measurement] = sValue
			}
		}
	}

	srv := "localhost"
	if service[0] != "" {
		srv = service[0]
	}

	tags := map[string]string{
		"server": srv,
		"port":   service[1],
		"state":  zookeeperState,
	}
	acc.AddFields("zookeeper", fields, tags)

	return nil
}

func init() {
	inputs.Add("zookeeper", func() telegraf.Input {
		return &Zookeeper{}
	})
}
