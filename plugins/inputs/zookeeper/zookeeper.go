//go:generate ../../../tools/config_includer/generator
//go:generate ../../../tools/readme_config_includer/generator
package zookeeper

import (
	"bufio"
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	common_tls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

var zookeeperFormatRE = regexp.MustCompile(`^zk_(\w[\w\.\-]*)\s+([\w\.\-]+)`)

type Zookeeper struct {
	Servers     []string        `toml:"servers"`
	Timeout     config.Duration `toml:"timeout"`
	ParseFloats string          `toml:"parse_floats"`

	EnableTLS bool `toml:"enable_tls" deprecated:"1.37.0;1.40.0;use 'tls_enable' instead"`
	common_tls.ClientConfig

	tlsConfig *tls.Config
}

func (*Zookeeper) SampleConfig() string {
	return sampleConfig
}

func (z *Zookeeper) Init() error {
	z.ClientConfig.Enable = &z.EnableTLS
	tlsConfig, err := z.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}
	z.tlsConfig = tlsConfig

	if z.Timeout < config.Duration(1*time.Second) {
		z.Timeout = config.Duration(5 * time.Second)
	}

	if len(z.Servers) == 0 {
		z.Servers = []string{":2181"}
	}

	return nil
}

func (z *Zookeeper) Gather(acc telegraf.Accumulator) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(z.Timeout))
	defer cancel()

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
			continue
		}

		sValue := parts[2]

		// First attempt to parse as an int
		iVal, err := strconv.ParseInt(sValue, 10, 64)
		if err == nil {
			fields[measurement] = iVal
			continue
		}

		// If set, attempt to parse as a float
		if z.ParseFloats == "float" {
			fVal, err := strconv.ParseFloat(sValue, 64)
			if err == nil {
				fields[measurement] = fVal
				continue
			}
		}

		// Finally, save as a string
		fields[measurement] = sValue
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

func (z *Zookeeper) dial(ctx context.Context, addr string) (net.Conn, error) {
	var dialer net.Dialer
	if z.tlsConfig != nil {
		deadline, ok := ctx.Deadline()
		if ok {
			dialer.Deadline = deadline
		}
		return tls.DialWithDialer(&dialer, "tcp", addr, z.tlsConfig)
	}
	return dialer.DialContext(ctx, "tcp", addr)
}

func init() {
	inputs.Add("zookeeper", func() telegraf.Input {
		return &Zookeeper{}
	})
}
