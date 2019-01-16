package kube_state

import (
	"context"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// KubernetesState represents the config object for the plugin.
type KubernetesState struct {
	URL               string            `toml:"url"`
	BearerToken       string            `toml:"bearer_token"`
	BearerTokenString string            `toml:"bearer_token_string"`
	Namespace         string            `toml:"namespace"`
	ResponseTimeout   internal.Duration `toml:"response_timeout"` // Timeout specified as a string - 3s, 1m, 1h
	ResourceExclude   []string          `toml:"resource_exclude"`
	ResourceInclude   []string          `toml:"resource_include"`
	MaxConfigMapAge   internal.Duration `toml:"max_config_map_age"`

	tls.ClientConfig

	// try to collect everything on first run
	firstTimeGather bool // apparently for configmaps
	client          *client
}

var sampleConfig = `
  ## URL for the kubelet
  url = "https://127.0.0.1"

  ## Namespace to use
  # namespace = "default"

  ## Use bearer token for authorization. ('bearer_token' takes priority)
  # bearer_token = "/path/to/bearer/token"
  ## OR
  # bearer_token_string = "abc_123"

  ## Set response_timeout (default 5 seconds)
  # response_timeout = "5s"

  ## Optional Resources to exclude from gathering
  ## Leave them with blank with try to gather everything available.
  ## Values can be - "configmaps", "daemonsets", deployments", "nodes",
  ## "persistentvolumes", "persistentvolumeclaims", "pods", "statefulsets"
  # resource_exclude = [ "deployments", "nodes", "statefulsets" ]

  ## Optional Resources to include when gathering
  ## Overrides resource_exclude if both set.
  # resource_include = [ "deployments", "nodes", "statefulsets" ]

  ## Optional max age for config map
  # max_config_map_age = "1h"

  ## Optional TLS Config
  # tls_ca = "/path/to/cafile"
  # tls_cert = "/path/to/certfile"
  # tls_key = "/path/to/keyfile"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

// SampleConfig returns a sample config
func (ks *KubernetesState) SampleConfig() string {
	return sampleConfig
}

// Description returns the description of this plugin
func (ks *KubernetesState) Description() string {
	return "Read metrics from the Kubernetes api"
}

// Gather collects kubernetes metrics from a given URL.
func (ks *KubernetesState) Gather(acc telegraf.Accumulator) (err error) {
	if ks.client == nil {
		if ks.client, err = ks.initClient(); err != nil {
			return err
		}
	}

	var wg sync.WaitGroup

	if len(ks.ResourceInclude) == 0 {
		for _, f := range availableCollectors {
			ctx := context.Background()
			wg.Add(1)
			go func(f func(ctx context.Context, acc telegraf.Accumulator, k *KubernetesState)) {
				defer wg.Done()
				f(ctx, acc, ks)
			}(f)
		}
	} else {
		for _, n := range ks.ResourceInclude {
			ctx := context.Background()
			wg.Add(1)
			go func(f func(ctx context.Context, acc telegraf.Accumulator, k *KubernetesState)) {
				defer wg.Done()
				f(ctx, acc, ks)
			}(availableCollectors[n])
		}
	}

	wg.Wait()
	ks.firstTimeGather = false

	return nil
}

var availableCollectors = map[string]func(ctx context.Context, acc telegraf.Accumulator, ks *KubernetesState){
	"configmaps":             collectConfigMaps,
	"daemonsets":             collectDaemonSets,
	"deployments":            collectDeployments,
	"nodes":                  collectNodes,
	"persistentvolumes":      collectPersistentVolumes,
	"persistentvolumeclaims": collectPersistentVolumeClaims,
	"pods":                   collectPods,
	"statefulsets":           collectStatefulSets,
}

func (ks *KubernetesState) initClient() (*client, error) {
	ks.firstTimeGather = true

	if len(ks.ResourceInclude) == 0 {
		for i := range ks.ResourceExclude {
			delete(availableCollectors, ks.ResourceExclude[i])
		}
	}

	if ks.BearerToken != "" {
		token, err := ioutil.ReadFile(ks.BearerToken)
		if err != nil {
			return nil, err
		}
		ks.BearerTokenString = strings.TrimSpace(string(token))
	}

	return newClient(ks.URL, ks.Namespace, ks.BearerTokenString, ks.ResponseTimeout.Duration, ks.ClientConfig)
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func atoi(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return int64(i)
}

var (
	configMapMeasurement             = "kubernetes_configmap"
	daemonSetMeasurement             = "kubernetes_daemonset"
	deploymentMeasurement            = "kubernetes_deployment"
	nodeMeasurement                  = "kubernetes_node"
	persistentVolumeMeasurement      = "kubernetes_persistentvolume"
	persistentVolumeClaimMeasurement = "kubernetes_persistentvolumeclaim"
	podStatusMeasurement             = "kubernetes_pod"
	podContainerMeasurement          = "kubernetes_pod_container"
	statefulSetMeasurement           = "kubernetes_statefulset"
)

func init() {
	inputs.Add("kube_state", func() telegraf.Input {
		return &KubernetesState{
			ResponseTimeout: internal.Duration{Duration: time.Second * 5},
			Namespace:       "default",
		}
	})
}
