package kube_lite

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// todo: filter (include/exclude) old k8s measurements
// todo: add required permissions for `/stats/summary`
// todo: get new example of node output
// todo: ensure `/stats/summary` doesn't return pods in all-namespaces

// KubernetesState represents the config object for the plugin.
type KubernetesState struct {
	URL             string            `toml:"url"`
	BearerToken     string            `toml:"bearer_token"`
	BearerTokenFile string            `toml:"bearer_token_file"`
	Namespace       string            `toml:"namespace"`
	ResponseTimeout internal.Duration `toml:"response_timeout"` // Timeout specified as a string - 3s, 1m, 1h
	ResourceExclude []string          `toml:"resource_exclude"`
	ResourceInclude []string          `toml:"resource_include"`
	MaxConfigMapAge internal.Duration `toml:"max_config_map_age"`

	tls.ClientConfig

	// try to collect everything on first run
	firstTimeGather bool // apparently for configmaps
	client          *client
	roundTripper    http.RoundTripper
}

var sampleConfig = `
  ## URL for the kubelet
  url = "https://1.1.1.1"

  ## Namespace to use
  namespace = "default"

  ## Use bearer token for authorization (token has priority over file)
  # bearer_token = abc123
  ## or
  # bearer_token_file = /path/to/bearer/token

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
  # tls_ca = /path/to/cafile # for '/stats/summary' only
  # tls_cert = /path/to/certfile # for '/stats/summary' only
  # tls_key = /path/to/keyfile # for '/stats/summary' only
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

	// gather `/stats/summary`
	wg.Add(1)
	go func(k *KubernetesState) {
		defer wg.Done()
		acc.AddError(k.gatherSummary(k.URL, acc))
	}(ks)

	// todo: reimplement better include/exclude filter
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

	if ks.BearerToken == "" && ks.BearerTokenFile != "" {
		token, err := ioutil.ReadFile(ks.BearerTokenFile)
		if err != nil {
			return nil, err
		}
		ks.BearerToken = strings.TrimSpace(string(token))
	}

	return newClient(ks.URL, ks.Namespace, ks.BearerToken, ks.ResponseTimeout.Duration, tlsCfg)
}

func (ks *KubernetesState) gatherSummary(baseURL string, acc telegraf.Accumulator) error {
	url := fmt.Sprintf("%s/stats/summary", baseURL)
	var req, err = http.NewRequest("GET", url, nil)
	var resp *http.Response

	tlsCfg, err := ks.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	if ks.roundTripper == nil {
		if ks.ResponseTimeout.Duration < time.Second {
			ks.ResponseTimeout.Duration = time.Second * 5
		}
		ks.roundTripper = &http.Transport{
			TLSHandshakeTimeout:   5 * time.Second,
			TLSClientConfig:       tlsCfg,
			ResponseHeaderTimeout: ks.ResponseTimeout.Duration,
		}
	}

	if ks.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+ks.BearerToken)
	}

	req.Header.Add("Accept", "application/json")

	resp, err = ks.roundTripper.RoundTrip(req)
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", url, resp.Status)
	}

	summaryMetrics := &SummaryMetrics{}
	err = json.NewDecoder(resp.Body).Decode(summaryMetrics)
	if err != nil {
		return fmt.Errorf(`Error parsing response: %s`, err)
	}
	buildSystemContainerMetrics(summaryMetrics, acc)
	buildNodeMetrics(summaryMetrics, acc)
	buildPodMetrics(summaryMetrics, acc)

	return nil
}

func buildSystemContainerMetrics(summaryMetrics *SummaryMetrics, acc telegraf.Accumulator) {
	for _, container := range summaryMetrics.Node.SystemContainers {
		tags := map[string]string{
			"node_name":      summaryMetrics.Node.NodeName,
			"container_name": container.Name,
		}
		fields := map[string]interface{}{
			"cpu_usage_nanocores":        container.CPU.UsageNanoCores,
			"cpu_usage_core_nanoseconds": container.CPU.UsageCoreNanoSeconds,
			"memory_usage_bytes":         container.Memory.UsageBytes,
			"memory_working_set_bytes":   container.Memory.WorkingSetBytes,
			"memory_rss_bytes":           container.Memory.RSSBytes,
			"memory_page_faults":         container.Memory.PageFaults,
			"memory_major_page_faults":   container.Memory.MajorPageFaults,
			"rootfs_available_bytes":     container.RootFS.AvailableBytes,
			"rootfs_capacity_bytes":      container.RootFS.CapacityBytes,
			"logsfs_avaialble_bytes":     container.LogsFS.AvailableBytes,
			"logsfs_capacity_bytes":      container.LogsFS.CapacityBytes,
		}
		acc.AddFields("kubernetes_system_container", fields, tags)
	}
}

func buildNodeMetrics(summaryMetrics *SummaryMetrics, acc telegraf.Accumulator) {
	tags := map[string]string{
		"node_name": summaryMetrics.Node.NodeName,
	}
	fields := map[string]interface{}{
		"cpu_usage_nanocores":              summaryMetrics.Node.CPU.UsageNanoCores,
		"cpu_usage_core_nanoseconds":       summaryMetrics.Node.CPU.UsageCoreNanoSeconds,
		"memory_available_bytes":           summaryMetrics.Node.Memory.AvailableBytes,
		"memory_usage_bytes":               summaryMetrics.Node.Memory.UsageBytes,
		"memory_working_set_bytes":         summaryMetrics.Node.Memory.WorkingSetBytes,
		"memory_rss_bytes":                 summaryMetrics.Node.Memory.RSSBytes,
		"memory_page_faults":               summaryMetrics.Node.Memory.PageFaults,
		"memory_major_page_faults":         summaryMetrics.Node.Memory.MajorPageFaults,
		"network_rx_bytes":                 summaryMetrics.Node.Network.RXBytes,
		"network_rx_errors":                summaryMetrics.Node.Network.RXErrors,
		"network_tx_bytes":                 summaryMetrics.Node.Network.TXBytes,
		"network_tx_errors":                summaryMetrics.Node.Network.TXErrors,
		"fs_available_bytes":               summaryMetrics.Node.FileSystem.AvailableBytes,
		"fs_capacity_bytes":                summaryMetrics.Node.FileSystem.CapacityBytes,
		"fs_used_bytes":                    summaryMetrics.Node.FileSystem.UsedBytes,
		"runtime_image_fs_available_bytes": summaryMetrics.Node.Runtime.ImageFileSystem.AvailableBytes,
		"runtime_image_fs_capacity_bytes":  summaryMetrics.Node.Runtime.ImageFileSystem.CapacityBytes,
		"runtime_image_fs_used_bytes":      summaryMetrics.Node.Runtime.ImageFileSystem.UsedBytes,
	}
	acc.AddFields("kubernetes_node", fields, tags)
}

func buildPodMetrics(summaryMetrics *SummaryMetrics, acc telegraf.Accumulator) {
	for _, pod := range summaryMetrics.Pods {
		for _, container := range pod.Containers {
			tags := map[string]string{
				"node_name":      summaryMetrics.Node.NodeName,
				"namespace":      pod.PodRef.Namespace,
				"container_name": container.Name,
				"pod_name":       pod.PodRef.Name,
			}
			fields := map[string]interface{}{
				"cpu_usage_nanocores":        container.CPU.UsageNanoCores,
				"cpu_usage_core_nanoseconds": container.CPU.UsageCoreNanoSeconds,
				"memory_usage_bytes":         container.Memory.UsageBytes,
				"memory_working_set_bytes":   container.Memory.WorkingSetBytes,
				"memory_rss_bytes":           container.Memory.RSSBytes,
				"memory_page_faults":         container.Memory.PageFaults,
				"memory_major_page_faults":   container.Memory.MajorPageFaults,
				"rootfs_available_bytes":     container.RootFS.AvailableBytes,
				"rootfs_capacity_bytes":      container.RootFS.CapacityBytes,
				"rootfs_used_bytes":          container.RootFS.UsedBytes,
				"logsfs_avaialble_bytes":     container.LogsFS.AvailableBytes,
				"logsfs_capacity_bytes":      container.LogsFS.CapacityBytes,
				"logsfs_used_bytes":          container.LogsFS.UsedBytes,
			}
			acc.AddFields("kubernetes_pod_container", fields, tags)
		}

		for _, volume := range pod.Volumes {
			tags := map[string]string{
				"node_name":   summaryMetrics.Node.NodeName,
				"pod_name":    pod.PodRef.Name,
				"namespace":   pod.PodRef.Namespace,
				"volume_name": volume.Name,
			}
			fields := map[string]interface{}{
				"available_bytes": volume.AvailableBytes,
				"capacity_bytes":  volume.CapacityBytes,
				"used_bytes":      volume.UsedBytes,
			}
			acc.AddFields("kubernetes_pod_volume", fields, tags)
		}

		tags := map[string]string{
			"node_name": summaryMetrics.Node.NodeName,
			"pod_name":  pod.PodRef.Name,
			"namespace": pod.PodRef.Namespace,
		}
		fields := map[string]interface{}{
			"rx_bytes":  pod.Network.RXBytes,
			"rx_errors": pod.Network.RXErrors,
			"tx_bytes":  pod.Network.TXBytes,
			"tx_errors": pod.Network.TXErrors,
		}
		acc.AddFields("kubernetes_pod_network", fields, tags)
	}
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

func atoi(s string) int64 {
	i, err := strconv.Atoi(s)
	if err != nil {
		fmt.Println(err) // todo: remove
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
	podStatusMeasurement             = "kubernetes_pod_status"
	podContainerMeasurement          = "kubernetes_pod_container"
	statefulSetMeasurement           = "kubernetes_statefulset"
)

func init() {
	inputs.Add("kube_state", func() telegraf.Input {
		return &KubernetesState{
			ResponseTimeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}
