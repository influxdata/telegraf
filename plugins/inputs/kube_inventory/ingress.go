package kube_inventory

import (
	"context"
	"errors"
	"time"

	v1beta1EXT "github.com/ericchiang/k8s/apis/extensions/v1beta1"

	"github.com/influxdata/telegraf"
)

func collectIngress(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getIngress(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, i := range list.Items {
		if err = ki.gatherIngress(*i, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

// todo: do we want to add cardinality and collect values from `i.GetStatus().GetLoadBalancer().GetIngress()`
func (ki *KubernetesInventory) gatherIngress(i v1beta1EXT.Ingress, acc telegraf.Accumulator) error {
	if i.Metadata.CreationTimestamp.GetSeconds() == 0 && i.Metadata.CreationTimestamp.GetNanos() == 0 {
		return nil
	}

	if i.Status.LoadBalancer == nil {
		return errors.New("invalid nil loadbalancer")
	}

	fields := map[string]interface{}{
		"created":    time.Unix(i.Metadata.CreationTimestamp.GetSeconds(), int64(i.Metadata.CreationTimestamp.GetNanos())).UnixNano(),
		"generation": i.Metadata.GetGeneration(),
	}

	tags := map[string]string{
		"name":      i.Metadata.GetName(),
		"namespace": i.Metadata.GetNamespace(),
	}

	for _, rule := range i.GetSpec().GetRules() {
		for _, path := range rule.GetIngressRuleValue().GetHttp().GetPaths() {
			fields["backend_service_port"] = path.GetBackend().GetServicePort().GetIntVal()

			tags["backend_service_name"] = path.GetBackend().GetServiceName()
			tags["path"] = path.GetPath()

			acc.AddFields(ingressMeasurement, fields, tags)
		}
	}

	return nil
}
