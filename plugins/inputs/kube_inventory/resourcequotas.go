package kube_inventory

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/influxdata/telegraf"
)

func collectResourceQuotas(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getResourceQuotas(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, i := range list.Items {
		ki.gatherResourceQuota(i, acc)
	}
}

func (ki *KubernetesInventory) gatherResourceQuota(r corev1.ResourceQuota, acc telegraf.Accumulator) {
	fields := map[string]interface{}{}
	tags := map[string]string{
		"resource":  r.Name,
		"namespace": r.Namespace,
	}

	for resourceName, val := range r.Status.Hard {
		switch resourceName {
		case "cpu", "limits.cpu", "requests.cpu":
			key := "hard_cpu"
			if strings.Contains(string(resourceName), "limits") {
				key = key + "_limits"
			} else if strings.Contains(string(resourceName), "requests") {
				key = key + "_requests"
			}
			fields[key] = ki.convertQuantity(val.String(), 1)
		case "memory", "limits.memory", "requests.memory":
			key := "hard_memory"
			if strings.Contains(string(resourceName), "limits") {
				key = key + "_limits"
			} else if strings.Contains(string(resourceName), "requests") {
				key = key + "_requests"
			}
			fields[key] = ki.convertQuantity(val.String(), 1)
		case "pods":
			fields["hard_pods"] = atoi(val.String())
		}
	}

	for resourceName, val := range r.Status.Used {
		switch resourceName {
		case "cpu", "requests.cpu", "limits.cpu":
			key := "used_cpu"
			if strings.Contains(string(resourceName), "limits") {
				key = key + "_limits"
			} else if strings.Contains(string(resourceName), "requests") {
				key = key + "_requests"
			}
			fields[key] = ki.convertQuantity(val.String(), 1)
		case "memory", "requests.memory", "limits.memory":
			key := "used_memory"
			if strings.Contains(string(resourceName), "limits") {
				key = key + "_limits"
			} else if strings.Contains(string(resourceName), "requests") {
				key = key + "_requests"
			}
			fields[key] = ki.convertQuantity(val.String(), 1)
		case "pods":
			fields["used_pods"] = atoi(val.String())
		}
	}

	acc.AddFields(resourcequotaMeasurement, fields, tags)
}
