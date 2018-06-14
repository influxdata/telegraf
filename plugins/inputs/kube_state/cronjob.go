package kube_state

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/robfig/cron"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var cronJobMeasurement = "kube_cronjob"

func registerCronJobCollector(ctx context.Context, acc telegraf.Accumulator, ks *KubenetesState) {
	list, err := ks.client.getCronJobs(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, d := range list.Items {
		if err = ks.gatherCronJob(d, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubenetesState) gatherCronJob(j batchv1beta1.CronJob, acc telegraf.Accumulator) error {
	// ignore suspended cronjob
	if j.Spec.Suspend != nil && *j.Spec.Suspend {
		return nil
	}
	fields := map[string]interface{}{
		"status_active": len(j.Status.Active),
		"schedule":      j.Spec.Schedule,
	}
	tags := map[string]string{
		"namespace":          j.Namespace,
		"cronjob":            j.Name,
		"concurrency_policy": strings.ToLower(string(j.Spec.ConcurrencyPolicy)),
	}
	if j.Spec.StartingDeadlineSeconds != nil {
		fields["spec_starting_deadline_seconds"] = *j.Spec.StartingDeadlineSeconds
	}
	nextScheduledTime, err := getNextScheduledTime(j.Spec.Schedule, j.Status.LastScheduleTime, j.CreationTimestamp)
	if err != nil {
		return err
	}
	fields["next_schedule_time"] = nextScheduledTime.Second()
	for k, v := range j.Labels {
		tags["label_"+sanitizeLabelName(k)] = v
	}
	if !j.CreationTimestamp.IsZero() {
		fields["created"] = j.CreationTimestamp.Unix()
	}
	if j.Status.LastScheduleTime != nil {
		fields["status_last_schedule_time"] = j.Status.LastScheduleTime.Unix()
	}
	acc.AddFields(cronJobMeasurement, fields, tags)
	return nil
}

func getNextScheduledTime(schedule string, lastScheduleTime *metav1.Time, createdTime metav1.Time) (time.Time, error) {
	sched, err := cron.ParseStandard(schedule)
	if err != nil {
		return time.Time{}, fmt.Errorf("Failed to parse cron job schedule '%s': %s", schedule, err)
	}
	if !lastScheduleTime.IsZero() {
		return sched.Next((*lastScheduleTime).Time), nil
	}
	if !createdTime.IsZero() {
		return sched.Next(createdTime.Time), nil
	}
	return time.Time{}, fmt.Errorf("Created time and lastScheduleTime are both zero")
}
