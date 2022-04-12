package kube_inventory

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	defaultServiceAccountPath = "/run/secrets/kubernetes.io/serviceaccount/token"
)

// KubernetesInventory represents the config object for the plugin.
type KubernetesInventory struct {
	URL               string          `toml:"url"`
	BearerToken       string          `toml:"bearer_token"`
	BearerTokenString string          `toml:"bearer_token_string"`
	Namespace         string          `toml:"namespace"`
	ResponseTimeout   config.Duration `toml:"response_timeout"` // Timeout specified as a string - 3s, 1m, 1h
	ResourceExclude   []string        `toml:"resource_exclude"`
	ResourceInclude   []string        `toml:"resource_include"`
	MaxConfigMapAge   config.Duration `toml:"max_config_map_age"`

	SelectorInclude []string `toml:"selector_include"`
	SelectorExclude []string `toml:"selector_exclude"`

	Log telegraf.Logger `toml:"-"`

	tls.ClientConfig
	client *client

	selectorFilter filter.Filter
}

func (ki *KubernetesInventory) Init() error {
	// If neither are provided, use the default service account.
	if ki.BearerToken == "" && ki.BearerTokenString == "" {
		ki.BearerToken = defaultServiceAccountPath
	}

	if ki.BearerToken != "" {
		token, err := os.ReadFile(ki.BearerToken)
		if err != nil {
			return err
		}
		ki.BearerTokenString = strings.TrimSpace(string(token))
	}

	var err error
	ki.client, err = newClient(ki.URL, ki.Namespace, ki.BearerTokenString, time.Duration(ki.ResponseTimeout), ki.ClientConfig)

	if err != nil {
		return err
	}

	return nil
}

// Gather collects kubernetes metrics from a given URL.
func (ki *KubernetesInventory) Gather(acc telegraf.Accumulator) (err error) {
	resourceFilter, err := filter.NewIncludeExcludeFilter(ki.ResourceInclude, ki.ResourceExclude)
	if err != nil {
		return err
	}

	ki.selectorFilter, err = filter.NewIncludeExcludeFilter(ki.SelectorInclude, ki.SelectorExclude)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	ctx := context.Background()

	for collector, f := range availableCollectors {
		if resourceFilter.Match(collector) {
			wg.Add(1)
			go func(f func(ctx context.Context, acc telegraf.Accumulator, k *KubernetesInventory)) {
				defer wg.Done()
				f(ctx, acc, ki)
			}(f)
		}
	}

	wg.Wait()

	return nil
}

var availableCollectors = map[string]func(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory){
	"daemonsets":             collectDaemonSets,
	"deployments":            collectDeployments,
	"endpoints":              collectEndpoints,
	"ingress":                collectIngress,
	"nodes":                  collectNodes,
	"pods":                   collectPods,
	"services":               collectServices,
	"statefulsets":           collectStatefulSets,
	"persistentvolumes":      collectPersistentVolumes,
	"persistentvolumeclaims": collectPersistentVolumeClaims,
}

func atoi(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

func (ki *KubernetesInventory) convertQuantity(s string, m float64) int64 {
	q, err := resource.ParseQuantity(s)
	if err != nil {
		ki.Log.Debugf("failed to parse quantity: %s", err.Error())
		return 0
	}
	f, err := strconv.ParseFloat(fmt.Sprint(q.AsDec()), 64)
	if err != nil {
		ki.Log.Debugf("failed to parse float: %s", err.Error())
		return 0
	}
	if m < 1 {
		m = 1
	}
	return int64(f * m)
}

func (ki *KubernetesInventory) createSelectorFilters() error {
	selectorFilter, err := filter.NewIncludeExcludeFilter(ki.SelectorInclude, ki.SelectorExclude)
	if err != nil {
		return err
	}
	ki.selectorFilter = selectorFilter
	return nil
}

var (
	daemonSetMeasurement             = "kubernetes_daemonset"
	deploymentMeasurement            = "kubernetes_deployment"
	endpointMeasurement              = "kubernetes_endpoint"
	ingressMeasurement               = "kubernetes_ingress"
	nodeMeasurement                  = "kubernetes_node"
	persistentVolumeMeasurement      = "kubernetes_persistentvolume"
	persistentVolumeClaimMeasurement = "kubernetes_persistentvolumeclaim"
	podContainerMeasurement          = "kubernetes_pod_container"
	serviceMeasurement               = "kubernetes_service"
	statefulSetMeasurement           = "kubernetes_statefulset"
)

func init() {
	inputs.Add("kube_inventory", func() telegraf.Input {
		return &KubernetesInventory{
			ResponseTimeout: config.Duration(time.Second * 5),
			Namespace:       "default",
			SelectorInclude: []string{},
			SelectorExclude: []string{"*"},
		}
	})
}
