package kube_state

import (
	"context"

	"github.com/influxdata/telegraf"
	"k8s.io/api/core/v1"
)

var (
	endpointsMeasurement = "kube_endpoint"
)

func registerEndpointCollector(ctx context.Context, acc telegraf.Accumulator, ks *KubenetesState) {
	list, err := ks.client.getEndpoints(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, e := range list.Items {
		if err = ks.gatherEndpoints(e, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubenetesState) gatherEndpoints(e v1.Endpoints, acc telegraf.Accumulator) error {
	fields := map[string]interface{}{}
	tags := map[string]string{
		"namespace": e.Namespace,
		"endpoint":  e.Name,
	}
	for k, v := range e.Labels {
		tags["label_"+sanitizeLabelName(k)] = v
	}
	if !e.CreationTimestamp.IsZero() {
		fields["created"] = e.CreationTimestamp.Unix()
	}
	var available int
	for _, s := range e.Subsets {
		available += len(s.Addresses) * len(s.Ports)
	}
	fields["address_available"] = available

	var notReady int
	for _, s := range e.Subsets {
		notReady += len(s.NotReadyAddresses) * len(s.Ports)
	}
	fields["address_not_ready"] = notReady

	acc.AddFields(endpointsMeasurement, fields, tags)
	return nil
}
