package aerospike

import (
	"crypto/tls"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"

	as "github.com/aerospike/aerospike-client-go"
)

type Aerospike struct {
	Servers []string `toml:"servers"`

	Username string `toml:"username"`
	Password string `toml:"password"`

	EnableTLS bool `toml:"enable_tls"`
	EnableSSL bool `toml:"enable_ssl"` // deprecated in 1.7; use enable_tls
	tlsint.ClientConfig

	initialized bool
	tlsConfig   *tls.Config
}

var sampleConfig = `
  ## Aerospike servers to connect to (with port)
  ## This plugin will query all namespaces the aerospike
  ## server has configured and get stats for them.
  servers = ["localhost:3000"]

  # username = "telegraf"
  # password = "pa$$word"

  ## Optional TLS Config
  # enable_tls = false
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## If false, skip chain & host verification
  # insecure_skip_verify = true
 `

func (a *Aerospike) SampleConfig() string {
	return sampleConfig
}

func (a *Aerospike) Description() string {
	return "Read stats from aerospike server(s)"
}

func (a *Aerospike) Gather(acc telegraf.Accumulator) error {
	if !a.initialized {
		tlsConfig, err := a.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
		if tlsConfig == nil && (a.EnableTLS || a.EnableSSL) {
			tlsConfig = &tls.Config{}
		}
		a.tlsConfig = tlsConfig
		a.initialized = true
	}

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

	policy := as.NewClientPolicy()
	policy.User = a.Username
	policy.Password = a.Password
	policy.TlsConfig = a.tlsConfig
	c, err := as.NewClientWithPolicy(policy, host, iport)
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
			val := parseValue(v)
			fields[strings.Replace(k, "-", "_", -1)] = val
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
				val := parseValue(parts[1])
				nFields[strings.Replace(parts[0], "-", "_", -1)] = val
			}
			acc.AddFields("aerospike_namespace", nFields, nTags, time.Now())
		}
	}
	return nil
}

func parseValue(v string) interface{} {
	if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
		return parsed
	} else if parsed, err := strconv.ParseUint(v, 10, 64); err == nil {
		return parsed
	} else if parsed, err := strconv.ParseBool(v); err == nil {
		return parsed
	} else {
		// leave as string
		return v
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
