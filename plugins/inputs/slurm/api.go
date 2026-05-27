package slurm

import "github.com/influxdata/telegraf"

type slurmAPI interface {
	gatherDiag(acc telegraf.Accumulator, source string) error
	gatherJobs(acc telegraf.Accumulator, source string) error
	gatherNodes(acc telegraf.Accumulator, source string) error
	gatherPartitions(acc telegraf.Accumulator, source string) error
	gatherReservations(acc telegraf.Accumulator, source string) error
}
