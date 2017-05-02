package aerospike

import (
	"errors"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	as "github.com/aerospike/aerospike-client-go"
)

type Aerospike struct {
	Servers []string
}

var sampleConfig = `
  ## Aerospike servers to connect to (with port)
  ## This plugin will query all namespaces the aerospike
  ## server has configured and get stats for them.
  servers = ["localhost:3000"]
 `

func (a *Aerospike) SampleConfig() string {
	return sampleConfig
}

func (a *Aerospike) Description() string {
	return "Read stats from aerospike server(s)"
}

func (a *Aerospike) Gather(acc telegraf.Accumulator) error {
	if len(a.Servers) == 0 {
		return a.gatherServer("127.0.0.1:3000", acc)
	}

	var wg sync.WaitGroup
	wg.Add(len(a.Servers))
	for _, server := range a.Servers {
		go func(serv string) {
			defer wg.Done()
			acc.AddError(a.gatherServer(serv, acc))
		}(server)
	}

	wg.Wait()
	return nil
}

func (a *Aerospike) gatherServer(hostport string, acc telegraf.Accumulator) error {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return err
	}

	iport, err := strconv.Atoi(port)
	if err != nil {
		iport = 3000
	}

	c, err := as.NewClient(host, iport)
	if err != nil {
		return err
	}
	defer c.Close()

	nodes := c.GetNodes()
	for _, n := range nodes {
		tags := map[string]string{
			"aerospike_host": hostport,
		}
		fields := map[string]interface{}{
			"node_name": n.GetName(),
		}
		stats, err := as.RequestNodeStats(n)
		if err != nil {
			return err
		}
		for k, v := range stats {
			val, err := parseValue(v)
			if err == nil {
				fields[strings.Replace(k, "-", "_", -1)] = val
			} else {
				log.Printf("I! skipping aerospike field %v with int64 overflow", k)
			}
		}
		acc.AddFields("aerospike_node", fields, tags, time.Now())

		info, err := as.RequestNodeInfo(n, "namespaces")
		if err != nil {
			return err
		}
		namespaces := strings.Split(info["namespaces"], ";")

		for _, namespace := range namespaces {
			nTags := map[string]string{
				"aerospike_host": hostport,
			}
			nTags["namespace"] = namespace
			nFields := map[string]interface{}{
				"node_name": n.GetName(),
			}
			info, err := as.RequestNodeInfo(n, "namespace/"+namespace)
			if err != nil {
				continue
			}
			stats := strings.Split(info["namespace/"+namespace], ";")
			for _, stat := range stats {
				parts := strings.Split(stat, "=")
				if len(parts) < 2 {
					continue
				}
				val, err := parseValue(parts[1])
				if err == nil {
					nFields[strings.Replace(parts[0], "-", "_", -1)] = val
				} else {
					log.Printf("I! skipping aerospike field %v with int64 overflow", parts[0])
				}
			}
			acc.AddFields("aerospike_namespace", nFields, nTags, time.Now())
		}
	}
	return nil
}

func parseValue(v string) (interface{}, error) {
	if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
		return parsed, nil
	} else if _, err := strconv.ParseUint(v, 10, 64); err == nil {
		// int64 overflow, yet valid uint64
		return nil, errors.New("Number is too large")
	} else if parsed, err := strconv.ParseBool(v); err == nil {
		return parsed, nil
	} else {
		return v, nil
	}
}

func copyTags(m map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range m {
		out[k] = v
	}
	return out
}

func init() {
	inputs.Add("aerospike", func() telegraf.Input {
		return &Aerospike{}
	})
}
