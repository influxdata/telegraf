package aerospike

import (
	"errors"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	as "github.com/aerospike/aerospike-client-go"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Aerospike struct {
	Servers []string
}

const (
	_FIELD = 0
	_VALUE = 1
)

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
			"node_name":      n.GetName(),
		}
		fields := make(map[string]interface{})
		stats, err := as.RequestNodeStats(n)
		if err != nil {
			return err
		}
		for k, v := range stats {
			val, err := parseValue(v)
			if err == nil {
				fields[strings.Replace(k, "-", "_", -1)] = val
			} else {
				log.Printf("I! skipping aerospike field %v with int64 overflow: %q", k, v)
			}
		}
		acc.AddFields("aerospike_node", fields, tags, time.Now())
		//Finding the latency metrics
		infoLatency, err := as.RequestNodeInfo(n, "latency:")
		if err != nil {
			return err
		}

		latency, err := parseNodeLatency(infoLatency)
		if err != nil {
			return err
		}

		info, err := as.RequestNodeInfo(n, "namespaces")
		if err != nil {
			return err
		}
		namespaces := strings.Split(info["namespaces"], ";")
		for _, namespace := range namespaces {
			nTags := map[string]string{
				"aerospike_host": hostport,
				"node_name":      n.GetName(),
			}
			nTags["namespace"] = namespace
			nFields := make(map[string]interface{})
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
					log.Printf("I! skipping aerospike field %v with int64 overflow: %q", parts[0], parts[1])
				}
			}
			if latencyMap, ok := latency[namespace]; ok {
				for k, v := range latencyMap {
					val, err := parseValue(v)
					if err == nil {
						nFields[strings.Replace(k, ">", "", -1)] = val
					} else {
						log.Printf("I! skipping aerospike field %v with int64 overflow: %q", k, v)
					}
				}
			}
			acc.AddFields("aerospike_namespace", nFields, nTags, time.Now())
		}
	}
	return nil
}

func parseNodeLatency(latencyMap map[string]string) (map[string]map[string]string, error) {
	res := make(map[string]map[string]string)
	v, exists := latencyMap["latency:"]
	if !exists {
		return res, nil
	}
	values := strings.Split(v, ";")
	flag_type := _FIELD
	var tmp_list []string
	for _, value := range values {
		if value == "error-no-data-yet-or-back-too-small" {
			continue
		}
		kv := strings.Split(value, ",")
		if flag_type == _FIELD {
			tmp_list = kv
			flag_type = _VALUE
		} else {
			tmp_map := make(map[string]string)
			namespace := ""
			operation := ""
			for in, field := range tmp_list {
				if in == 0 {
					matched_string := regexp.MustCompile("{.+}").FindStringSubmatch(field)
					namespace = strings.Trim(matched_string[0], "{}")
					matched_string = regexp.MustCompile("-[a-zA-Z]+:").FindStringSubmatch(field)
					operation = strings.Trim(matched_string[0], "-:")
					continue
				}
				tmp_map[operation+"_"+field] = kv[in]
			}
			for k, v := range tmp_map {
				if res[namespace] == nil {
					res[namespace] = make(map[string]string)
				}
				res[namespace][k] = v
			}
			flag_type = _FIELD
		}
	}
	return res, nil
}

func parseValue(v string) (interface{}, error) {
	if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
		return parsed, nil
	} else if _, err := strconv.ParseUint(v, 10, 64); err == nil {
		// int64 overflow, yet valid uint64
		return nil, errors.New("Number is too large")
	} else if parsed, err := strconv.ParseBool(v); err == nil {
		return parsed, nil
	} else if parsed, err := strconv.ParseFloat(v, 64); err == nil {
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
