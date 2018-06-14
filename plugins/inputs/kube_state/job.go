package kube_state

import (
	"context"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	v1batch "k8s.io/api/batch/v1"
)

var (
	jobMeasurement          = "kube_job"
	jobConditionMeasurement = "kube_job_condition"
)

func registerJobCollector(ctx context.Context, acc telegraf.Accumulator, ks *KubenetesState) {
	list, err := ks.client.getJobs(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, j := range list.Items {
		if err = ks.gatherJob(j, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubenetesState) gatherJob(j v1batch.Job, acc telegraf.Accumulator) error {
	var fields map[string]interface{}
	var tags map[string]string
	// only record completed job
	if j.Status.CompletionTime == nil {
		return nil
	} else if !ks.firstTimeGather && ks.MaxJobAge != nil && ks.MaxJobAge.Duration <
		time.Now().Sub(j.Status.CompletionTime.Time) {
		// ignore completed job more than ks.MaxJobAge
		goto gatherJobConditions
	}

	fields = map[string]interface{}{
		"status_succeeded": j.Status.Succeeded,
		"status_failed":    j.Status.Failed,
		"status_active":    j.Status.Active,
	}
	tags = map[string]string{
		"namespace": j.Namespace,
		"job_name":  j.Name,
	}
	for k, v := range j.Labels {
		tags["label_"+sanitizeLabelName(k)] = v
	}
	if j.Spec.Parallelism != nil {
		fields["spec_parallelism"] = *j.Spec.Parallelism
	}

	if j.Spec.Completions != nil {
		fields["spec_completions"] = *j.Spec.Completions
	}
	if !j.CreationTimestamp.IsZero() {
		fields["created"] = j.CreationTimestamp.Unix()
	}

	if j.Spec.ActiveDeadlineSeconds != nil {
		fields["spec_active_deadline_seconds"] = *j.Spec.ActiveDeadlineSeconds
	}

	if j.Status.StartTime != nil && !j.Status.StartTime.IsZero() {
		fields["status_start_time"] = j.Status.StartTime.Unix()
	}

	acc.AddFields(jobMeasurement, fields, tags, j.Status.CompletionTime.Time)
gatherJobConditions:
	for _, c := range j.Status.Conditions {
		ks.gatherJobCondition(c, j, acc)
	}
	return nil
}

func (ks *KubenetesState) gatherJobCondition(c v1batch.JobCondition, j v1batch.Job, acc telegraf.Accumulator) {
	fields := map[string]interface{}{
		"completed": 0,
		"failed":    0,
	}
	tags := map[string]string{
		"namespace": j.Namespace,
		"job_name":  j.Name,
		"condition": strings.ToLower(string(c.Status)),
	}
	switch c.Type {
	case v1batch.JobComplete:
		fields["completed"] = 1
	case v1batch.JobFailed:
		fields["failed"] = 1
	}
	acc.AddFields(jobConditionMeasurement, fields, tags, c.LastTransitionTime.Time)
}
