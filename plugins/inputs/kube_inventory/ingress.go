package kube_inventory

import (
	"context"
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

func (ki *KubernetesInventory) gatherIngress(i v1beta1EXT.Ingress, acc telegraf.Accumulator) error {
	if i.Metadata.CreationTimestamp.GetSeconds() == 0 && i.Metadata.CreationTimestamp.GetNanos() == 0 {
		return nil
	}

	fields := map[string]interface{}{
		"created":    time.Unix(i.Metadata.CreationTimestamp.GetSeconds(), int64(i.Metadata.CreationTimestamp.GetNanos())).UnixNano(),
		"generation": i.Metadata.GetGeneration(),
	}

	tags := map[string]string{
		"ingress_name": i.Metadata.GetName(),
		"namespace":    i.Metadata.GetNamespace(),
	}

	for _, ingress := range i.GetStatus().GetLoadBalancer().GetIngress() {
		tags["hostname"] = ingress.GetHostname()
		tags["ip"] = ingress.GetIp()

		for _, rule := range i.GetSpec().GetRules() {
			for _, path := range rule.GetIngressRuleValue().GetHttp().GetPaths() {
				fields["backend_service_port"] = path.GetBackend().GetServicePort().GetIntVal()
				fields["tls"] = i.GetSpec().GetTls() != nil

				tags["backend_service_name"] = path.GetBackend().GetServiceName()
				tags["path"] = path.GetPath()
				tags["host"] = rule.GetHost()

				acc.AddFields(ingressMeasurement, fields, tags)
			}
		}
	}

	return nil
}
