package zookeeper

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Zookeeper is a zookeeper plugin
type Zookeeper struct {
	Servers []string
}

var sampleConfig = `
  ## An array of address to gather stats about. Specify an ip or hostname
  ## with port. ie localhost:2181, 10.0.0.1:2181, etc.

  ## If no servers are specified, then localhost is used as the host.
  ## If no port is specified, 2181 is used
  servers = [":2181"]
`

var defaultTimeout = time.Second * time.Duration(5)

// SampleConfig returns sample configuration message
func (z *Zookeeper) SampleConfig() string {
	return sampleConfig
}

// Description returns description of Zookeeper plugin
func (z *Zookeeper) Description() string {
	return `Reads 'mntr' stats from one or many zookeeper servers`
}

// Gather reads stats from all configured servers accumulates stats
func (z *Zookeeper) Gather(acc telegraf.Accumulator) error {
	if len(z.Servers) == 0 {
		return nil
	}

	for _, serverAddress := range z.Servers {
		acc.AddError(z.gatherServer(serverAddress, acc))
	}
	return nil
}

func (z *Zookeeper) gatherServer(address string, acc telegraf.Accumulator) error {
	var zookeeper_state string
	_, _, err := net.SplitHostPort(address)
	if err != nil {
		address = address + ":2181"
	}

	c, err := net.DialTimeout("tcp", address, defaultTimeout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	defer c.Close()

	// Extend connection
	c.SetDeadline(time.Now().Add(defaultTimeout))

	fmt.Fprintf(c, "%s\n", "mntr")
	rdr := bufio.NewReader(c)
	scanner := bufio.NewScanner(rdr)

	service := strings.Split(address, ":")
	if len(service) != 2 {
		return fmt.Errorf("Invalid service address: %s", address)
	}

	fields := make(map[string]interface{})
	for scanner.Scan() {
		line := scanner.Text()

		re := regexp.MustCompile(`^zk_(\w+)\s+([\w\.\-]+)`)
		parts := re.FindStringSubmatch(string(line))

		if len(parts) != 3 {
			return fmt.Errorf("unexpected line in mntr response: %q", line)
		}

		measurement := strings.TrimPrefix(parts[1], "zk_")
		if measurement == "server_state" {
			zookeeper_state = parts[2]
		} else {
			sValue := string(parts[2])

			iVal, err := strconv.ParseInt(sValue, 10, 64)
			if err == nil {
				fields[measurement] = iVal
			} else {
				fields[measurement] = sValue
			}
		}
	}
	tags := map[string]string{
		"server": service[0],
		"port":   service[1],
		"state":  zookeeper_state,
	}
	acc.AddFields("zookeeper", fields, tags)

	return nil
}

func init() {
	inputs.Add("zookeeper", func() telegraf.Input {
		return &Zookeeper{}
	})
}
