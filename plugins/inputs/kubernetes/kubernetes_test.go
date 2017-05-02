package kubernetes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestKubernetesStats(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, response)
	}))
	defer ts.Close()

	k := &Kubernetes{
		URL: ts.URL,
	}

	var acc testutil.Accumulator
	err := acc.GatherError(k.Gather)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"cpu_usage_nanocores":        int64(56652446),
		"cpu_usage_core_nanoseconds": int64(101437561712262),
		"memory_usage_bytes":         int64(62529536),
		"memory_working_set_bytes":   int64(62349312),
		"memory_rss_bytes":           int64(47509504),
		"memory_page_faults":         int64(4769397409),
		"memory_major_page_faults":   int64(13),
		"rootfs_available_bytes":     int64(84379979776),
		"rootfs_capacity_bytes":      int64(105553100800),
		"logsfs_avaialble_bytes":     int64(84379979776),
		"logsfs_capacity_bytes":      int64(105553100800),
	}
	tags := map[string]string{
		"node_name":      "node1",
		"container_name": "kubelet",
	}
	acc.AssertContainsTaggedFields(t, "kubernetes_system_container", fields, tags)

	fields = map[string]interface{}{
		"cpu_usage_nanocores":              int64(576996212),
		"cpu_usage_core_nanoseconds":       int64(774129887054161),
		"memory_usage_bytes":               int64(12313182208),
		"memory_working_set_bytes":         int64(5081538560),
		"memory_rss_bytes":                 int64(35586048),
		"memory_page_faults":               int64(351742),
		"memory_major_page_faults":         int64(1236),
		"memory_available_bytes":           int64(10726387712),
		"network_rx_bytes":                 int64(213281337459),
		"network_rx_errors":                int64(0),
		"network_tx_bytes":                 int64(292869995684),
		"network_tx_errors":                int64(0),
		"fs_available_bytes":               int64(84379979776),
		"fs_capacity_bytes":                int64(105553100800),
		"fs_used_bytes":                    int64(16754286592),
		"runtime_image_fs_available_bytes": int64(84379979776),
		"runtime_image_fs_capacity_bytes":  int64(105553100800),
		"runtime_image_fs_used_bytes":      int64(5809371475),
	}
	tags = map[string]string{
		"node_name": "node1",
	}
	acc.AssertContainsTaggedFields(t, "kubernetes_node", fields, tags)

	fields = map[string]interface{}{
		"cpu_usage_nanocores":        int64(846503),
		"cpu_usage_core_nanoseconds": int64(56507553554),
		"memory_usage_bytes":         int64(30789632),
		"memory_working_set_bytes":   int64(30789632),
		"memory_rss_bytes":           int64(30695424),
		"memory_page_faults":         int64(10761),
		"memory_major_page_faults":   int64(0),
		"rootfs_available_bytes":     int64(84379979776),
		"rootfs_capacity_bytes":      int64(105553100800),
		"rootfs_used_bytes":          int64(57344),
		"logsfs_avaialble_bytes":     int64(84379979776),
		"logsfs_capacity_bytes":      int64(105553100800),
		"logsfs_used_bytes":          int64(24576),
	}
	tags = map[string]string{
		"node_name":      "node1",
		"container_name": "foocontainer",
		"namespace":      "foons",
		"pod_name":       "foopod",
	}
	acc.AssertContainsTaggedFields(t, "kubernetes_pod_container", fields, tags)

	fields = map[string]interface{}{
		"cpu_usage_nanocores":        int64(846503),
		"cpu_usage_core_nanoseconds": int64(56507553554),
		"memory_usage_bytes":         int64(0),
		"memory_working_set_bytes":   int64(0),
		"memory_rss_bytes":           int64(0),
		"memory_page_faults":         int64(0),
		"memory_major_page_faults":   int64(0),
		"rootfs_available_bytes":     int64(0),
		"rootfs_capacity_bytes":      int64(0),
		"rootfs_used_bytes":          int64(0),
		"logsfs_avaialble_bytes":     int64(0),
		"logsfs_capacity_bytes":      int64(0),
		"logsfs_used_bytes":          int64(0),
	}
	tags = map[string]string{
		"node_name":      "node1",
		"container_name": "stopped-container",
		"namespace":      "foons",
		"pod_name":       "stopped-pod",
	}
	acc.AssertContainsTaggedFields(t, "kubernetes_pod_container", fields, tags)

	fields = map[string]interface{}{
		"available_bytes": int64(7903948800),
		"capacity_bytes":  int64(7903961088),
		"used_bytes":      int64(12288),
	}
	tags = map[string]string{
		"node_name":   "node1",
		"volume_name": "volume1",
		"namespace":   "foons",
		"pod_name":    "foopod",
	}
	acc.AssertContainsTaggedFields(t, "kubernetes_pod_volume", fields, tags)

	fields = map[string]interface{}{
		"rx_bytes":  int64(70749124),
		"rx_errors": int64(0),
		"tx_bytes":  int64(47813506),
		"tx_errors": int64(0),
	}
	tags = map[string]string{
		"node_name": "node1",
		"namespace": "foons",
		"pod_name":  "foopod",
	}
	acc.AssertContainsTaggedFields(t, "kubernetes_pod_network", fields, tags)

}

var response = `
{
  "node": {
   "nodeName": "node1",
   "systemContainers": [
    {
     "name": "kubelet",
     "startTime": "2016-08-25T18:46:52Z",
     "cpu": {
      "time": "2016-09-27T16:57:31Z",
      "usageNanoCores": 56652446,
      "usageCoreNanoSeconds": 101437561712262
     },
     "memory": {
      "time": "2016-09-27T16:57:31Z",
      "usageBytes": 62529536,
      "workingSetBytes": 62349312,
      "rssBytes": 47509504,
      "pageFaults": 4769397409,
      "majorPageFaults": 13
     },
     "rootfs": {
      "availableBytes": 84379979776,
      "capacityBytes": 105553100800
     },
     "logs": {
      "availableBytes": 84379979776,
      "capacityBytes": 105553100800
     },
     "userDefinedMetrics": null
   },
   {
    "name": "bar",
    "startTime": "2016-08-25T18:46:52Z",
    "cpu": {
     "time": "2016-09-27T16:57:31Z",
     "usageNanoCores": 56652446,
     "usageCoreNanoSeconds": 101437561712262
    },
    "memory": {
     "time": "2016-09-27T16:57:31Z",
     "usageBytes": 62529536,
     "workingSetBytes": 62349312,
     "rssBytes": 47509504,
     "pageFaults": 4769397409,
     "majorPageFaults": 13
    },
    "rootfs": {
     "availableBytes": 84379979776,
     "capacityBytes": 105553100800
    },
    "logs": {
     "availableBytes": 84379979776,
     "capacityBytes": 105553100800
    },
    "userDefinedMetrics": null
   }
   ],
   "startTime": "2016-08-25T18:46:52Z",
   "cpu": {
    "time": "2016-09-27T16:57:41Z",
    "usageNanoCores": 576996212,
    "usageCoreNanoSeconds": 774129887054161
   },
   "memory": {
    "time": "2016-09-27T16:57:41Z",
    "availableBytes": 10726387712,
    "usageBytes": 12313182208,
    "workingSetBytes": 5081538560,
    "rssBytes": 35586048,
    "pageFaults": 351742,
    "majorPageFaults": 1236
   },
   "network": {
    "time": "2016-09-27T16:57:41Z",
    "rxBytes": 213281337459,
    "rxErrors": 0,
    "txBytes": 292869995684,
    "txErrors": 0
   },
   "fs": {
    "availableBytes": 84379979776,
    "capacityBytes": 105553100800,
    "usedBytes": 16754286592
   },
   "runtime": {
    "imageFs": {
     "availableBytes": 84379979776,
     "capacityBytes": 105553100800,
     "usedBytes": 5809371475
    }
   }
  },
  "pods": [
   {
    "podRef": {
     "name": "foopod",
     "namespace": "foons",
     "uid": "6d305b06-8419-11e6-825c-42010af000ae"
    },
    "startTime": "2016-09-26T18:45:42Z",
    "containers": [
     {
      "name": "foocontainer",
      "startTime": "2016-09-26T18:46:43Z",
      "cpu": {
       "time": "2016-09-27T16:57:32Z",
       "usageNanoCores": 846503,
       "usageCoreNanoSeconds": 56507553554
      },
      "memory": {
       "time": "2016-09-27T16:57:32Z",
       "usageBytes": 30789632,
       "workingSetBytes": 30789632,
       "rssBytes": 30695424,
       "pageFaults": 10761,
       "majorPageFaults": 0
      },
      "rootfs": {
       "availableBytes": 84379979776,
       "capacityBytes": 105553100800,
       "usedBytes": 57344
      },
      "logs": {
       "availableBytes": 84379979776,
       "capacityBytes": 105553100800,
       "usedBytes": 24576
      },
      "userDefinedMetrics": null
     }
    ],
    "network": {
     "time": "2016-09-27T16:57:34Z",
     "rxBytes": 70749124,
     "rxErrors": 0,
     "txBytes": 47813506,
     "txErrors": 0
    },
    "volume": [
     {
      "availableBytes": 7903948800,
      "capacityBytes": 7903961088,
      "usedBytes": 12288,
      "name": "volume1"
     },
     {
      "availableBytes": 7903956992,
      "capacityBytes": 7903961088,
      "usedBytes": 4096,
      "name": "volume2"
     },
     {
      "availableBytes": 7903948800,
      "capacityBytes": 7903961088,
      "usedBytes": 12288,
      "name": "volume3"
     },
     {
      "availableBytes": 7903952896,
      "capacityBytes": 7903961088,
      "usedBytes": 8192,
      "name": "volume4"
     }
    ]
   },
   {
    "podRef": {
     "name": "stopped-pod",
     "namespace": "foons",
     "uid": "da7c1865-d67d-4688-b679-c485ed44b2aa"
    },
    "startTime": null,
    "containers": [
     {
      "name": "stopped-container",
      "startTime": "2016-09-26T18:46:43Z",
      "cpu": {
       "time": "2016-09-27T16:57:32Z",
       "usageNanoCores": 846503,
       "usageCoreNanoSeconds": 56507553554
      }
     }
    ]
   }
  ]
 }`
