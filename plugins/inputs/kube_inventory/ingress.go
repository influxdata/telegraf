package kube_inventory

import (
	"context"
	"time"

	"github.com/ericchiang/k8s/apis/core/v1"
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

func (ki *KubernetesInventory) gatherIngress(i v1beta1EXT.Ingress, acc telegraf.Accumulator) error {
	if i.Metadata.CreationTimestamp.GetSeconds() == 0 && i.Metadata.CreationTimestamp.GetNanos() == 0 {
		return nil
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

// todo: do we want to add cardinality and collect values from `i.GetStatus().GetLoadBalancer().GetIngress()`
func (ki *KubernetesInventory) gatherIngressWithIps(i v1beta1EXT.Ingress, acc telegraf.Accumulator) error {
	if i.Metadata.CreationTimestamp.GetSeconds() == 0 && i.Metadata.CreationTimestamp.GetNanos() == 0 {
		return nil
	}

	fields := map[string]interface{}{
		"created":    time.Unix(i.Metadata.CreationTimestamp.GetSeconds(), int64(i.Metadata.CreationTimestamp.GetNanos())).UnixNano(),
		"generation": i.Metadata.GetGeneration(),
	}

	tags := map[string]string{
		"name":      i.Metadata.GetName(),
		"namespace": i.Metadata.GetNamespace(),
	}

	for _, ingress := range i.GetStatus().GetLoadBalancer().GetIngress() {
		tags["ip"] = getHostOrIP(ingress)

		for _, rule := range i.GetSpec().GetRules() {
			for _, path := range rule.GetIngressRuleValue().GetHttp().GetPaths() {
				fields["backend_service_port"] = path.GetBackend().GetServicePort().GetIntVal()

				tags["backend_service_name"] = path.GetBackend().GetServiceName()
				tags["path"] = path.GetPath()

				acc.AddFields(ingressMeasurement, fields, tags)
			}
		}
	}

	return nil
}

func getHostOrIP(ingress *v1.LoadBalancerIngress) string {
	if name := ingress.GetHostname(); name != "" {
		return name
	}

	return ingress.GetIp()
}
