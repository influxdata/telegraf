package kube_lite

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// KubernetesState represents the config object for the plugin.
type KubernetesState struct {
	URL             string            `toml:url`
	BearerToken     string            `toml:"bearer_token"`
	Namespace       string            `toml:"namespace"`
	ResponseTimeout internal.Duration `toml:"response_timeout"` // Timeout specified as a string - 3s, 1m, 1h
	ResourceExclude []string          `toml:"resource_exclude"`
	ResourceInclude []string          `toml:"resource_include"`
	MaxConfigMapAge internal.Duration `toml:"max_config_map_age"`

	tls.ClientConfig

	// try to collect everything on first run
	firstTimeGather bool // apparently for configmaps
	client          *client
}

var sampleConfig = `
  ## URL for the kubelet
  url = "https://1.1.1.1"

  ## Namespace to use
  namespace = "default"

  ## Use bearer token for authorization
  #  bearer_token = "abc123"

  ## Set response_timeout (default 5 seconds)
  #  response_timeout = "5s"

  ## Optional Resources to exclude from gathering
  ## Leave them with blank with try to gather everything available.
  ## Values can be - "configmaps", "daemonsets", deployments", "nodes",
  ## "persistentvolumes", "persistentvolumeclaims", "pods", "statefulsets"
  #  resource_exclude = [ "deployments", "nodes", "statefulsets" ]

  ## Optional Resources to include when gathering
  ## Overrides resource_exclude if both set.
  #  resource_include = [ "deployments", "nodes", "statefulsets" ]

  ## Optional max age for config map
  #  max_config_map_age = "1h"

  ## Optional TLS Config
  ## Use TLS but skip chain & host verification
  #  insecure_skip_verify = false
`

// SampleConfig returns a sample config
func (ks *KubernetesState) SampleConfig() string {
	return sampleConfig
}

// Description returns the description of this plugin
func (ks *KubernetesState) Description() string {
	return "Read metrics from the kubernetes kubelet api"
}

// Gather collects kubernetes metrics from a given URL.
// todo: convert to service?
func (ks *KubernetesState) Gather(acc telegraf.Accumulator) (err error) {
	if ks.client == nil {
		if ks.client, err = ks.initClient(); err != nil {
			return err
		}
	}

	var wg sync.WaitGroup
	// todo: reimplement better include/exclude filter
	if len(ks.ResourceInclude) == 0 {
		for n, f := range availableCollectors {
			ctx := context.Background()
			wg.Add(1)
			go func(n string, f func(ctx context.Context, acc telegraf.Accumulator, k *KubernetesState)) {
				defer wg.Done()
				f(ctx, acc, ks)
			}(n, f)
		}
	} else {
		for _, n := range ks.ResourceInclude {
			ctx := context.Background()
			wg.Add(1)
			go func(n string, f func(ctx context.Context, acc telegraf.Accumulator, k *KubernetesState)) {
				defer wg.Done()
				f(ctx, acc, ks)
			}(n, availableCollectors[n])
		}
	}

	wg.Wait()
	// always set ks.firstTimeGather to false
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
	"pods":         collectPods,
	"statefulsets": collectStatefulSets,
}

func (ks *KubernetesState) initClient() (*client, error) {
	tlsCfg, err := ks.ClientConfig.TLSConfig()
	if err != nil {
		return nil, fmt.Errorf("error parse kube state metrics config[%s]: %v", ks.URL, err)
	}
	ks.firstTimeGather = true

	if len(ks.ResourceInclude) == 0 {
		for i := range ks.ResourceExclude {
			// todo: likely to break reloading config file
			delete(availableCollectors, ks.ResourceExclude[i])
		}
	}

	return newClient(ks.URL, ks.Namespace, ks.BearerToken, ks.ResponseTimeout.Duration, tlsCfg)
}

func init() {
	inputs.Add("kube_state", func() telegraf.Input {
		return &KubernetesState{
			ResponseTimeout: internal.Duration{Duration: time.Second * 5},
		}
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
