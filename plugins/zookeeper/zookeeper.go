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

	"github.com/influxdb/telegraf/plugins"
)

// Zookeeper is a zookeeper plugin
type Zookeeper struct {
	Servers []string
}

var sampleConfig = `
	# An array of address to gather stats about. Specify an ip or hostname
	# with port. ie localhost:2181, 10.0.0.1:2181, etc.

	# If no servers are specified, then localhost is used as the host.
	# If no port is specified, 2181 is used
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
func (z *Zookeeper) Gather(acc plugins.Accumulator) error {
	if len(z.Servers) == 0 {
		return nil
	}

	for _, serverAddress := range z.Servers {
		if err := z.gatherServer(serverAddress, acc); err != nil {
			return err
		}
	}
	return nil
}

func (z *Zookeeper) gatherServer(address string, acc plugins.Accumulator) error {
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

	fmt.Fprintf(c, "%s\n", "mntr")

	rdr := bufio.NewReader(c)

	scanner := bufio.NewScanner(rdr)

	for scanner.Scan() {
		line := scanner.Text()

		re := regexp.MustCompile(`^zk_(\w+)\s+([\w\.\-]+)`)
		parts := re.FindStringSubmatch(string(line))

		service := strings.Split(address, ":")

		if len(parts) != 3 || len(service) != 2 {
			return fmt.Errorf("unexpected line in mntr response: %q", line)
		}

		tags := map[string]string{"server": service[0], "port": service[1]}

		measurement := strings.TrimPrefix(parts[1], "zk_")
		sValue := string(parts[2])

		iVal, err := strconv.ParseInt(sValue, 10, 64)
		if err == nil {
			acc.Add(measurement, iVal, tags)
		} else {
			acc.Add(measurement, sValue, tags)
		}
	}

	return nil
}

func init() {
	plugins.Add("zookeeper", func() plugins.Plugin {
		return &Zookeeper{}
	})
}
