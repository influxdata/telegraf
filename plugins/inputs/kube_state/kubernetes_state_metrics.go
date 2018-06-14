package kube_state

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubenetesState represents the config object for the plugin.
type KubenetesState struct {
	URL string

	// Bearer Token authorization file path
	BearerToken string `toml:"bearer_token"`

	// MaxConnections for worker pool tcp connections
	MaxConnections int `toml:"max_connections"`

	// HTTP Timeout specified as a string - 3s, 1m, 1h
	ResponseTimeout internal.Duration `toml:"response_timeout"`

	tls.ClientConfig

	client          *client
	rListHash       string
	filter          filter.Filter
	lastFilterBuilt time.Time
	// try to collect everything at first time
	firstTimeGather           bool
	ResourceListCheckInterval *internal.Duration `toml:"resouce_list_check_interval"`
	ResourceExclude           []string           `toml:"resource_exclude"`

	MaxConfigMapAge internal.Duration  `toml:"max_config_map_age"`
	MaxJobAge       *internal.Duration `toml:"max_job_age"`
}

var sampleConfig = `
  ## URL for the kubelet
  url = "http://1.1.1.1:10255"

  ## Use bearer token for authorization
  #  bearer_token = /path/to/bearer/token

  ## Set response_timeout (default 5 seconds)
  #  response_timeout = "5s"

  ## Optional TLS Config
  #  tls_ca = /path/to/cafile
  #  tls_cert = /path/to/certfile
  #  tls_key = /path/to/keyfile
  ## Use TLS but skip chain & host verification
  #  insecure_skip_verify = false

  ## Worker pool for kube_state_metric plugin only
  #  empty this field will use default value 30
  #  max_connections = 5

  ## Optional Max Config Map Age, this will ignore 
  ## old config map for the second time of gathering.
  ## Blank will use the default value 24 hours
  #  max_config_map_age = "24h" 

  ## Optional Max Job Age, this will ignore
  ## old job created for kube_job metrics for the second time of gathering.
  ## Blank will collect all jobs
  #  max_job_age = "24h"

  ## Optional Resources to exclude from gathering
  ## Leave them with blank with try to gather everything available.
  ## Values can be "cronjobs", "daemonsets", "deployments",
  ## "jobs", "limitranges", "nodes", "pods", "replicasets", "replicationcontrollers",
  ## "resourcequotas", "services", "statefulsets", "persistentvolumes",
  ## "persistentvolumeclaims", "namespaces", "horizontalpodautoscalers",
  ## "endpoints", "secrets", "configmaps"
  #  resource_exclude = [ "cronjobs", "daemonsets", "deployments" ]

  ## Optional Resouce List Check Interval, leave blank will use the default
  #  value of 1 hour. This is the interval to check available resource lists.
  #  resouce_list_check_interval = "1h"
`

// SampleConfig returns a sample config
func (ks *KubenetesState) SampleConfig() string {
	return sampleConfig
}

// Description returns the description of this plugin
func (ks *KubenetesState) Description() string {
	return "Read metrics from the kubernetes kubelet api"
}

// Gather collects kubernetes metrics from a given URL
func (ks *KubenetesState) Gather(acc telegraf.Accumulator) (err error) {
	var rList *metav1.APIResourceList
	var cutoff time.Time
	if ks.client == nil {
		if ks.client, rList, err = ks.initClient(); err != nil {
			return err
		}
		goto buildFilter
	}

	cutoff = time.Now().Add(-1 * ks.ResourceListCheckInterval.Duration)

	// Here we just test
	if ks.lastFilterBuilt.Unix() > 0 && ks.lastFilterBuilt.Before(cutoff) {
		goto doGather
	}

	rList, err = ks.client.getAPIResourceList(context.Background())
	if err != nil {
		return err
	}

buildFilter:
	ks.lastFilterBuilt = time.Now()
	if err = ks.buildFilter(rList); err != nil {
		return err
	}

doGather:
	var wg sync.WaitGroup
	for n, f := range availableCollectors {
		ctx := context.Background()
		if ks.filter.Match(n) {
			wg.Add(1)
			go func(n string, f func(ctx context.Context, acc telegraf.Accumulator, k *KubenetesState)) {
				defer wg.Done()
				println("!", n)
				f(ctx, acc, ks)
			}(n, f)
		}
	}
	wg.Wait()
	// always set ks.firstTimeGather to false
	ks.firstTimeGather = false

	return nil
}

func (k *KubenetesState) buildFilter(rList *metav1.APIResourceList) error {
	hash, err := genHash(rList)
	if err != nil {
		return err
	}
	if k.rListHash == hash {
		return nil
	}
	k.rListHash = hash
	include := make([]string, len(rList.APIResources))
	for k, v := range rList.APIResources {
		include[k] = v.Name
	}
	k.filter, err = filter.NewIncludeExcludeFilter(include, k.ResourceExclude)
	return err
}

func genHash(rList *metav1.APIResourceList) (string, error) {
	buf := new(bytes.Buffer)
	for _, v := range rList.APIResources {
		if _, err := buf.WriteString(v.Name + "|"); err != nil {
			return "", err
		}
	}
	sum := md5.Sum(buf.Bytes())
	return string(sum[:]), nil
}

var availableCollectors = map[string]func(ctx context.Context, acc telegraf.Accumulator, ks *KubenetesState){
	"cronjobs":                 registerCronJobCollector,
	"daemonsets":               registerDaemonSetCollector,
	"deployments":              registerDeploymentCollector,
	"jobs":                     registerJobCollector,
	"limitranges":              registerLimitRangeCollector,
	"nodes":                    registerNodeCollector,
	"pods":                     registerPodCollector,
	"replicasets":              registerReplicaSetCollector,
	"replicationcontrollers":   registerReplicationControllerCollector,
	"resourcequotas":           registerResourceQuotaCollector,
	"services":                 registerServiceCollector,
	"statefulsets":             registerStatefulSetCollector,
	"persistentvolumes":        registerPersistentVolumeCollector,
	"persistentvolumeclaims":   registerPersistentVolumeClaimCollector,
	"namespaces":               registerNamespaceCollector,
	"horizontalpodautoscalers": registerHorizontalPodAutoScalerCollector,
	"endpoints":                registerEndpointCollector,
	"secrets":                  registerSecretCollector,
	"configmaps":               registerConfigMapCollector,
}

func (k *KubenetesState) initClient() (*client, *metav1.APIResourceList, error) {
	tlsCfg, err := k.ClientConfig.TLSConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("error parse kube state metrics config[%s]: %v", k.URL, err)
	}
	k.firstTimeGather = true

	// default 30 concurrent TCP connections
	if k.MaxConnections == 0 {
		k.MaxConnections = 30
	}

	// default check resourceList every hour
	if k.ResourceListCheckInterval == nil {
		k.ResourceListCheckInterval = &internal.Duration{
			Duration: time.Hour,
		}
	}

	// default to ignore configMaps created more than 1 hour
	if (k.MaxConfigMapAge == internal.Duration{Duration: 0}) {
		k.MaxConfigMapAge = internal.Duration{Duration: time.Hour * 24}
	}

	c := newClient(k.URL, k.ResponseTimeout.Duration, k.MaxConnections, k.BearerToken, tlsCfg)
	rList, err := c.getAPIResourceList(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("error connect to kubenetes api endpoint[%s]: %v", k.URL, err)
	}
	log.Printf("I! Kubenetes API group version is %s", rList.GroupVersion)
	return c, rList, nil
}

func init() {
	inputs.Add("kube_state", func() telegraf.Input {
		return &KubenetesState{}
	})
}

var invalidLabelCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func sanitizeLabelName(s string) string {
	return invalidLabelCharRE.ReplaceAllString(s, "_")
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
