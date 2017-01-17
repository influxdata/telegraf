package aerospike

import (
	"errors"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/errchan"
	"github.com/influxdata/telegraf/plugins/inputs"

	as "github.com/aerospike/aerospike-client-go"
)

const NO_DATA_ERROR = "error-no-data-yet-or-back-too-small"

type Aerospike struct {
	Servers []string
}

var sampleConfig = `
  ## Aerospike servers to connect to (with port)
  ## This plugin will query all namespaces the aerospike
  ## server has configured and get stats for them.
  servers = ["user:password@localhost:3000","localhost:3000"]
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
	errChan := errchan.New(len(a.Servers))
	wg.Add(len(a.Servers))
	for _, server := range a.Servers {
		go func(serv string) {
			defer wg.Done()
			errChan.C <- a.gatherServer(serv, acc)
		}(server)
	}

	wg.Wait()
	return errChan.Error()
}

func (a *Aerospike) gatherServer(rawurl string, acc telegraf.Accumulator) error {
	rawurl = "aero://" + rawurl
	urlInfo, err := url.Parse(rawurl)
	hostport := urlInfo.Host

	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return err
	}

	iport, err := strconv.Atoi(port)
	if err != nil {
		iport = 3000
	}

	var c *as.Client
	if urlInfo.User != nil {
		userInfo := urlInfo.User
		clientPolicy := as.NewClientPolicy()
		clientPolicy.User = userInfo.Username()
		clientPolicy.Password, _ = userInfo.Password()
		c, err = as.NewClientWithPolicy(clientPolicy, host, iport)

	} else {
		c, err = as.NewClient(host, iport)
	}

	if err != nil {
		log.Println("Connection failed with error ", err)
		return err
	}
	defer c.Close()

	nodes := c.GetNodes()
	for _, n := range nodes {

		err = a.gatherNodeStats(n, acc, hostport)
		if err != nil {
			log.Println("E!", err)
		}

		err = a.gatherNamespaceStats(n, acc, hostport)
		if err != nil {
			log.Println("E! ", err)
		}
		err = a.gatherLatency(n, acc, hostport)
		if err != nil {
			log.Println("E! ", err)
		}

		err = a.gatherThroughput(n, acc, hostport)
		if err != nil {
			log.Println("E! ", err)
		}

	}
	return nil
}

func (a *Aerospike) gatherNodeStats(n *as.Node, acc telegraf.Accumulator, hostport string) error {
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
	return nil
}

func (a *Aerospike) gatherThroughput(n *as.Node, acc telegraf.Accumulator, hostport string) error {

	throughput, err := as.RequestNodeInfo(n, "throughput:")
	if err != nil {
		return err
	}

	result, exists := throughput["throughput:"]
	if !exists {
		return errors.New("No key throughput: in response")
	}
	points := strings.Split(result, ";")
	i := 0
	for i < len(points)-1 {
		point := points[i]
		if strings.Compare(point, NO_DATA_ERROR) == 0 {
			i++
		} else {
			ind := strings.Index(point, ":")
			if ind == -1 {
				i++
				continue
			}
			namespaceMetric := point[0:ind]
			metric, namespace := splitNamespaceMetric(namespaceMetric)
			nextPoint := points[i+1]
			vals := strings.Split(nextPoint, ",")
			if len(vals) != 2 {
				i += 2
				continue
			}

			fields := make(map[string]interface{})
			tags := map[string]string{
				"aerospike_host": hostport,
				"namespace":      namespace,
			}

			metricName := "throughput_" + metric
			value, err := parseValue(vals[1])
			if err != nil {
				i += 2
				continue
			}
			fields[metricName] = value
			acc.AddFields("aerospike_throughput", fields, tags, time.Now())
			i = i + 2
		}
	}

	return nil
}

func (a *Aerospike) gatherLatency(n *as.Node, acc telegraf.Accumulator, hostport string) error {

	latency, err := as.RequestNodeInfo(n, "latency:")
	if err != nil {
		return err
	}

	result, exists := latency["latency:"]
	if !exists {
		return errors.New("No key latency: in response")
	}

	points := strings.Split(result, ";")
	i := 0
	for i < len(points) {
		point := points[i]
		if strings.Compare(point, NO_DATA_ERROR) == 0 {
			i++
		} else {
			ind := strings.Index(point, ":")
			if ind == -1 {
				i++
				continue
			}
			namespaceMetric := point[0:ind]
			metric, namespace := splitNamespaceMetric(namespaceMetric)
			spl1 := point[ind:]
			ind = strings.Index(spl1, ",")
			spl1 = spl1[ind+1:]
			unitTimes := strings.Split(spl1, ",")
			nextPoint := points[i+1]
			vals := strings.Split(nextPoint, ",")

			fields := make(map[string]interface{})
			tags := map[string]string{
				"aerospike_host": hostport,
				"namespace":      namespace,
			}
			for i := 1; i < len(unitTimes); i++ {
				metric := "latency_" + metric + "_" + unitTimes[i]
				value, err := parseValue(vals[i+1])
				if err != nil {

				} else {
					fields[metric] = value
				}

			}
			acc.AddFields("aerospike_latency", fields, tags, time.Now())
			i = i + 2
		}
	}

	return nil
}

func splitNamespaceMetric(str string) (string, string) {

	if strings.Contains(str, "{") {
		i1 := strings.Index(str, "{")
		i2 := strings.Index(str, "}")
		namespace := str[i1+1 : i2]
		metric := str[i2+2:]
		return metric, namespace
	}
	return str, "total"

}

func (a *Aerospike) gatherNamespaceStats(n *as.Node, acc telegraf.Accumulator, hostport string) error {

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
	}
	return v, nil

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
