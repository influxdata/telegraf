package aerospike

import (
	"crypto/tls"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	as "github.com/aerospike/aerospike-client-go"

	"github.com/influxdata/telegraf"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
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

	DisableQueryNamespaces bool     `toml:"disable_query_namespaces"`
	Namespaces             []string `toml:"namespaces"`

	QuerySets bool     `toml:"query_sets"`
	Sets      []string `toml:"sets"`

	EnableTTLHistogram              bool `toml:"enable_ttl_histogram"`
	EnableObjectSizeLinearHistogram bool `toml:"enable_object_size_linear_histogram"`

	NumberHistogramBuckets int `toml:"num_histogram_buckets"`
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

  # Feature Options
  # Add namespace variable to limit the namespaces executed on
  # Leave blank to do all
  # disable_query_namespaces = true # default false
  # namespaces = ["namespace1", "namespace2"]

  # Enable set level telemetry
  # query_sets = true # default: false
  # Add namespace set combinations to limit sets executed on
  # Leave blank to do all sets
  # sets = ["namespace1/set1", "namespace1/set2", "namespace3"]

  # Histograms
  # enable_ttl_histogram = true # default: false
  # enable_object_size_linear_histogram = true # default: false

  # by default, aerospike produces a 100 bucket histogram
  # this is not great for most graphing tools, this will allow
  # the ability to squash this to a smaller number of buckets
  # To have a balanced histogram, the number of buckets chosen 
  # should divide evenly into 100.
  # num_histogram_buckets = 100 # default: 10
`

// On the random chance a hex value is all digits
// these are fields that can contain hex and should always be strings
var protectedHexFields = map[string]bool{
	"node_name":       true,
	"cluster_key":     true,
	"paxos_principal": true,
}

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

	if a.NumberHistogramBuckets == 0 {
		a.NumberHistogramBuckets = 10
	} else if a.NumberHistogramBuckets > 100 {
		a.NumberHistogramBuckets = 100
	} else if a.NumberHistogramBuckets < 1 {
		a.NumberHistogramBuckets = 10
	}

	if len(a.Servers) == 0 {
		return a.gatherServer(acc, "127.0.0.1:3000")
	}

	var wg sync.WaitGroup
	wg.Add(len(a.Servers))
	for _, server := range a.Servers {
		go func(serv string) {
			defer wg.Done()
			acc.AddError(a.gatherServer(acc, serv))
		}(server)
	}

	wg.Wait()
	return nil
}

func (a *Aerospike) gatherServer(acc telegraf.Accumulator, hostPort string) error {
	host, port, err := net.SplitHostPort(hostPort)
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
		stats, err := a.getNodeInfo(n)
		if err != nil {
			return err
		}
		a.parseNodeInfo(acc, stats, hostPort, n.GetName())

		namespaces, err := a.getNamespaces(n)
		if err != nil {
			return err
		}

		if !a.DisableQueryNamespaces {
			// Query Namespaces
			for _, namespace := range namespaces {
				stats, err = a.getNamespaceInfo(namespace, n)

				if err != nil {
					continue
				}
				a.parseNamespaceInfo(acc, stats, hostPort, namespace, n.GetName())

				if a.EnableTTLHistogram {
					err = a.getTTLHistogram(acc, hostPort, namespace, "", n)
					if err != nil {
						continue
					}
				}
				if a.EnableObjectSizeLinearHistogram {
					err = a.getObjectSizeLinearHistogram(acc, hostPort, namespace, "", n)
					if err != nil {
						continue
					}
				}
			}
		}

		if a.QuerySets {
			namespaceSets, err := a.getSets(n)
			if err == nil {
				for _, namespaceSet := range namespaceSets {
					namespace, set := splitNamespaceSet(namespaceSet)
					stats, err := a.getSetInfo(namespaceSet, n)

					if err != nil {
						continue
					}
					a.parseSetInfo(acc, stats, hostPort, namespaceSet, n.GetName())

					if a.EnableTTLHistogram {
						err = a.getTTLHistogram(acc, hostPort, namespace, set, n)
						if err != nil {
							continue
						}
					}

					if a.EnableObjectSizeLinearHistogram {
						err = a.getObjectSizeLinearHistogram(acc, hostPort, namespace, set, n)
						if err != nil {
							continue
						}
					}
				}
			}
		}
	}
	return nil
}

func (a *Aerospike) getNodeInfo(n *as.Node) (map[string]string, error) {
	stats, err := as.RequestNodeStats(n)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (a *Aerospike) parseNodeInfo(acc telegraf.Accumulator, stats map[string]string, hostPort string, nodeName string) {
	tags := map[string]string{
		"aerospike_host": hostPort,
		"node_name":      nodeName,
	}
	fields := make(map[string]interface{})

	for k, v := range stats {
		key := strings.Replace(k, "-", "_", -1)
		fields[key] = parseAerospikeValue(key, v)
	}
	acc.AddFields("aerospike_node", fields, tags, time.Now())
}

func (a *Aerospike) getNamespaces(n *as.Node) ([]string, error) {
	var namespaces []string
	if len(a.Namespaces) <= 0 {
		info, err := as.RequestNodeInfo(n, "namespaces")
		if err != nil {
			return namespaces, err
		}
		namespaces = strings.Split(info["namespaces"], ";")
	} else {
		namespaces = a.Namespaces
	}

	return namespaces, nil
}

func (a *Aerospike) getNamespaceInfo(namespace string, n *as.Node) (map[string]string, error) {
	stats, err := as.RequestNodeInfo(n, "namespace/"+namespace)
	if err != nil {
		return nil, err
	}

	return stats, err
}
func (a *Aerospike) parseNamespaceInfo(acc telegraf.Accumulator, stats map[string]string, hostPort string, namespace string, nodeName string) {
	nTags := map[string]string{
		"aerospike_host": hostPort,
		"node_name":      nodeName,
	}
	nTags["namespace"] = namespace
	nFields := make(map[string]interface{})

	stat := strings.Split(stats["namespace/"+namespace], ";")
	for _, pair := range stat {
		parts := strings.Split(pair, "=")
		if len(parts) < 2 {
			continue
		}
		key := strings.Replace(parts[0], "-", "_", -1)
		nFields[key] = parseAerospikeValue(key, parts[1])
	}
	acc.AddFields("aerospike_namespace", nFields, nTags, time.Now())
}

func (a *Aerospike) getSets(n *as.Node) ([]string, error) {
	var namespaceSets []string
	// Gather all sets
	if len(a.Sets) <= 0 {
		stats, err := as.RequestNodeInfo(n, "sets")
		if err != nil {
			return namespaceSets, err
		}

		stat := strings.Split(stats["sets"], ";")
		for _, setStats := range stat {
			// setInfo is "ns=test:set=foo:objects=1:tombstones=0"
			if len(setStats) > 0 {
				pairs := strings.Split(setStats, ":")
				var ns, set string
				for _, pair := range pairs {
					parts := strings.Split(pair, "=")
					if len(parts) == 2 {
						if parts[0] == "ns" {
							ns = parts[1]
						}
						if parts[0] == "set" {
							set = parts[1]
						}
					}
				}
				if len(ns) > 0 && len(set) > 0 {
					namespaceSets = append(namespaceSets, fmt.Sprintf("%s/%s", ns, set))
				}
			}
		}
	} else { // User has passed in sets
		namespaceSets = a.Sets
	}

	return namespaceSets, nil
}

func (a *Aerospike) getSetInfo(namespaceSet string, n *as.Node) (map[string]string, error) {
	stats, err := as.RequestNodeInfo(n, "sets/"+namespaceSet)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (a *Aerospike) parseSetInfo(acc telegraf.Accumulator, stats map[string]string, hostPort string, namespaceSet string, nodeName string) {
	stat := strings.Split(
		strings.TrimSuffix(
			stats[fmt.Sprintf("sets/%s", namespaceSet)], ";"), ":")
	nTags := map[string]string{
		"aerospike_host": hostPort,
		"node_name":      nodeName,
		"set":            namespaceSet,
	}
	nFields := make(map[string]interface{})
	for _, part := range stat {
		pieces := strings.Split(part, "=")
		if len(pieces) < 2 {
			continue
		}

		key := strings.Replace(pieces[0], "-", "_", -1)
		nFields[key] = parseAerospikeValue(key, pieces[1])
	}
	acc.AddFields("aerospike_set", nFields, nTags, time.Now())
}

func (a *Aerospike) getTTLHistogram(acc telegraf.Accumulator, hostPort string, namespace string, set string, n *as.Node) error {
	stats, err := a.getHistogram(namespace, set, "ttl", n)
	if err != nil {
		return err
	}

	nTags := createTags(hostPort, n.GetName(), namespace, set)
	a.parseHistogram(acc, stats, nTags, "ttl")

	return nil
}

func (a *Aerospike) getObjectSizeLinearHistogram(acc telegraf.Accumulator, hostPort string, namespace string, set string, n *as.Node) error {
	stats, err := a.getHistogram(namespace, set, "object-size-linear", n)
	if err != nil {
		return err
	}

	nTags := createTags(hostPort, n.GetName(), namespace, set)
	a.parseHistogram(acc, stats, nTags, "object-size-linear")

	return nil
}

func (a *Aerospike) getHistogram(namespace string, set string, histogramType string, n *as.Node) (map[string]string, error) {
	var queryArg string
	if len(set) > 0 {
		queryArg = fmt.Sprintf("histogram:type=%s;namespace=%v;set=%v", histogramType, namespace, set)
	} else {
		queryArg = fmt.Sprintf("histogram:type=%s;namespace=%v", histogramType, namespace)
	}

	stats, err := as.RequestNodeInfo(n, queryArg)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (a *Aerospike) parseHistogram(acc telegraf.Accumulator, stats map[string]string, nTags map[string]string, histogramType string) {
	nFields := make(map[string]interface{})

	for _, stat := range stats {
		for _, part := range strings.Split(stat, ":") {
			pieces := strings.Split(part, "=")
			if len(pieces) < 2 {
				continue
			}

			if pieces[0] == "buckets" {
				buckets := strings.Split(pieces[1], ",")

				// Normalize in case of less buckets than expected
				numRecordsPerBucket := 1
				if len(buckets) > a.NumberHistogramBuckets {
					numRecordsPerBucket = int(math.Ceil(float64(len(buckets)) / float64(a.NumberHistogramBuckets)))
				}

				bucketCount := 0
				bucketSum := int64(0) // cast to int64, as can have large object sums
				bucketName := 0
				for i, bucket := range buckets {
					// Sum records and increment bucket collection counter
					if bucketCount < numRecordsPerBucket {
						bucketSum = bucketSum + parseAerospikeValue("", bucket).(int64)
						bucketCount++
					}

					// Store records and reset counters
					// increment bucket name
					if bucketCount == numRecordsPerBucket {
						nFields[strconv.Itoa(bucketName)] = bucketSum

						bucketCount = 0
						bucketSum = 0
						bucketName++
					} else if i == (len(buckets) - 1) {
						// base/edge case where final bucket does not fully
						// fill number of records per bucket
						nFields[strconv.Itoa(bucketName)] = bucketSum
					}
				}
			}
		}
	}

	acc.AddFields(fmt.Sprintf("aerospike_histogram_%v", strings.Replace(histogramType, "-", "_", -1)), nFields, nTags, time.Now())
}

func splitNamespaceSet(namespaceSet string) (namespace string, set string) {
	split := strings.Split(namespaceSet, "/")
	return split[0], split[1]
}

func parseAerospikeValue(key string, v string) interface{} {
	if protectedHexFields[key] {
		return v
	} else if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
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

func createTags(hostPort string, nodeName string, namespace string, set string) map[string]string {
	nTags := map[string]string{
		"aerospike_host": hostPort,
		"node_name":      nodeName,
		"namespace":      namespace,
	}

	if len(set) > 0 {
		nTags["set"] = set
	}
	return nTags
}

func init() {
	inputs.Add("aerospike", func() telegraf.Input {
		return &Aerospike{}
	})
}
