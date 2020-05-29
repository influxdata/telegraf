package beanstalkd

import (
	"fmt"
	"io"
	"net/textproto"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"gopkg.in/yaml.v2"
)

const sampleConfig = `
  ## Server to collect data from
  server = "localhost:11300"

  ## List of tubes to gather stats about.
  ## If no tubes specified then data gathered for each tube on server reported by list-tubes command
  tubes = ["notifications"]
`

type Beanstalkd struct {
	Server string   `toml:"server"`
	Tubes  []string `toml:"tubes"`
}

func (b *Beanstalkd) Description() string {
	return "Collects Beanstalkd server and tubes stats"
}

func (b *Beanstalkd) SampleConfig() string {
	return sampleConfig
}

func (b *Beanstalkd) Gather(acc telegraf.Accumulator) error {
	connection, err := textproto.Dial("tcp", b.Server)
	if err != nil {
		return err
	}
	defer connection.Close()

	tubes := b.Tubes
	if len(tubes) == 0 {
		err = runQuery(connection, "list-tubes", &tubes)
		if err != nil {
			acc.AddError(err)
		}
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		err := b.gatherServerStats(connection, acc)
		if err != nil {
			acc.AddError(err)
		}
		wg.Done()
	}()

	for _, tube := range tubes {
		wg.Add(1)
		go func(tube string) {
			b.gatherTubeStats(connection, tube, acc)
			wg.Done()
		}(tube)
	}

	wg.Wait()

	return nil
}

func (b *Beanstalkd) gatherServerStats(connection *textproto.Conn, acc telegraf.Accumulator) error {
	stats := new(statsResponse)
	if err := runQuery(connection, "stats", stats); err != nil {
		return err
	}

	acc.AddFields("beanstalkd_overview",
		map[string]interface{}{
			"binlog_current_index":     stats.BinlogCurrentIndex,
			"binlog_max_size":          stats.BinlogMaxSize,
			"binlog_oldest_index":      stats.BinlogOldestIndex,
			"binlog_records_migrated":  stats.BinlogRecordsMigrated,
			"binlog_records_written":   stats.BinlogRecordsWritten,
			"cmd_bury":                 stats.CmdBury,
			"cmd_delete":               stats.CmdDelete,
			"cmd_ignore":               stats.CmdIgnore,
			"cmd_kick":                 stats.CmdKick,
			"cmd_list_tube_used":       stats.CmdListTubeUsed,
			"cmd_list_tubes":           stats.CmdListTubes,
			"cmd_list_tubes_watched":   stats.CmdListTubesWatched,
			"cmd_pause_tube":           stats.CmdPauseTube,
			"cmd_peek":                 stats.CmdPeek,
			"cmd_peek_buried":          stats.CmdPeekBuried,
			"cmd_peek_delayed":         stats.CmdPeekDelayed,
			"cmd_peek_ready":           stats.CmdPeekReady,
			"cmd_put":                  stats.CmdPut,
			"cmd_release":              stats.CmdRelease,
			"cmd_reserve":              stats.CmdReserve,
			"cmd_reserve_with_timeout": stats.CmdReserveWithTimeout,
			"cmd_stats":                stats.CmdStats,
			"cmd_stats_job":            stats.CmdStatsJob,
			"cmd_stats_tube":           stats.CmdStatsTube,
			"cmd_touch":                stats.CmdTouch,
			"cmd_use":                  stats.CmdUse,
			"cmd_watch":                stats.CmdWatch,
			"current_connections":      stats.CurrentConnections,
			"current_jobs_buried":      stats.CurrentJobsBuried,
			"current_jobs_delayed":     stats.CurrentJobsDelayed,
			"current_jobs_ready":       stats.CurrentJobsReady,
			"current_jobs_reserved":    stats.CurrentJobsReserved,
			"current_jobs_urgent":      stats.CurrentJobsUrgent,
			"current_producers":        stats.CurrentProducers,
			"current_tubes":            stats.CurrentTubes,
			"current_waiting":          stats.CurrentWaiting,
			"current_workers":          stats.CurrentWorkers,
			"job_timeouts":             stats.JobTimeouts,
			"max_job_size":             stats.MaxJobSize,
			"pid":                      stats.Pid,
			"rusage_stime":             stats.RusageStime,
			"rusage_utime":             stats.RusageUtime,
			"total_connections":        stats.TotalConnections,
			"total_jobs":               stats.TotalJobs,
			"uptime":                   stats.Uptime,
		},
		map[string]string{
			"hostname": stats.Hostname,
			"id":       stats.Id,
			"server":   b.Server,
			"version":  stats.Version,
		},
	)

	return nil
}

func (b *Beanstalkd) gatherTubeStats(connection *textproto.Conn, tube string, acc telegraf.Accumulator) error {
	stats := new(statsTubeResponse)
	if err := runQuery(connection, "stats-tube "+tube, stats); err != nil {
		return err
	}

	acc.AddFields("beanstalkd_tube",
		map[string]interface{}{
			"cmd_delete":            stats.CmdDelete,
			"cmd_pause_tube":        stats.CmdPauseTube,
			"current_jobs_buried":   stats.CurrentJobsBuried,
			"current_jobs_delayed":  stats.CurrentJobsDelayed,
			"current_jobs_ready":    stats.CurrentJobsReady,
			"current_jobs_reserved": stats.CurrentJobsReserved,
			"current_jobs_urgent":   stats.CurrentJobsUrgent,
			"current_using":         stats.CurrentUsing,
			"current_waiting":       stats.CurrentWaiting,
			"current_watching":      stats.CurrentWatching,
			"pause":                 stats.Pause,
			"pause_time_left":       stats.PauseTimeLeft,
			"total_jobs":            stats.TotalJobs,
		},
		map[string]string{
			"name":   stats.Name,
			"server": b.Server,
		},
	)

	return nil
}

func runQuery(connection *textproto.Conn, cmd string, result interface{}) error {
	requestId, err := connection.Cmd(cmd)
	if err != nil {
		return err
	}

	connection.StartResponse(requestId)
	defer connection.EndResponse(requestId)

	status, err := connection.ReadLine()
	if err != nil {
		return err
	}

	size := 0
	if _, err = fmt.Sscanf(status, "OK %d", &size); err != nil {
		return err
	}

	body := make([]byte, size+2)
	if _, err = io.ReadFull(connection.R, body); err != nil {
		return err
	}

	return yaml.Unmarshal(body, result)
}

func init() {
	inputs.Add("beanstalkd", func() telegraf.Input {
		return &Beanstalkd{}
	})
}

type statsResponse struct {
	BinlogCurrentIndex    int     `yaml:"binlog-current-index"`
	BinlogMaxSize         int     `yaml:"binlog-max-size"`
	BinlogOldestIndex     int     `yaml:"binlog-oldest-index"`
	BinlogRecordsMigrated int     `yaml:"binlog-records-migrated"`
	BinlogRecordsWritten  int     `yaml:"binlog-records-written"`
	CmdBury               int     `yaml:"cmd-bury"`
	CmdDelete             int     `yaml:"cmd-delete"`
	CmdIgnore             int     `yaml:"cmd-ignore"`
	CmdKick               int     `yaml:"cmd-kick"`
	CmdListTubeUsed       int     `yaml:"cmd-list-tube-used"`
	CmdListTubes          int     `yaml:"cmd-list-tubes"`
	CmdListTubesWatched   int     `yaml:"cmd-list-tubes-watched"`
	CmdPauseTube          int     `yaml:"cmd-pause-tube"`
	CmdPeek               int     `yaml:"cmd-peek"`
	CmdPeekBuried         int     `yaml:"cmd-peek-buried"`
	CmdPeekDelayed        int     `yaml:"cmd-peek-delayed"`
	CmdPeekReady          int     `yaml:"cmd-peek-ready"`
	CmdPut                int     `yaml:"cmd-put"`
	CmdRelease            int     `yaml:"cmd-release"`
	CmdReserve            int     `yaml:"cmd-reserve"`
	CmdReserveWithTimeout int     `yaml:"cmd-reserve-with-timeout"`
	CmdStats              int     `yaml:"cmd-stats"`
	CmdStatsJob           int     `yaml:"cmd-stats-job"`
	CmdStatsTube          int     `yaml:"cmd-stats-tube"`
	CmdTouch              int     `yaml:"cmd-touch"`
	CmdUse                int     `yaml:"cmd-use"`
	CmdWatch              int     `yaml:"cmd-watch"`
	CurrentConnections    int     `yaml:"current-connections"`
	CurrentJobsBuried     int     `yaml:"current-jobs-buried"`
	CurrentJobsDelayed    int     `yaml:"current-jobs-delayed"`
	CurrentJobsReady      int     `yaml:"current-jobs-ready"`
	CurrentJobsReserved   int     `yaml:"current-jobs-reserved"`
	CurrentJobsUrgent     int     `yaml:"current-jobs-urgent"`
	CurrentProducers      int     `yaml:"current-producers"`
	CurrentTubes          int     `yaml:"current-tubes"`
	CurrentWaiting        int     `yaml:"current-waiting"`
	CurrentWorkers        int     `yaml:"current-workers"`
	Hostname              string  `yaml:"hostname"`
	Id                    string  `yaml:"id"`
	JobTimeouts           int     `yaml:"job-timeouts"`
	MaxJobSize            int     `yaml:"max-job-size"`
	Pid                   int     `yaml:"pid"`
	RusageStime           float64 `yaml:"rusage-stime"`
	RusageUtime           float64 `yaml:"rusage-utime"`
	TotalConnections      int     `yaml:"total-connections"`
	TotalJobs             int     `yaml:"total-jobs"`
	Uptime                int     `yaml:"uptime"`
	Version               string  `yaml:"version"`
}

type statsTubeResponse struct {
	CmdDelete           int    `yaml:"cmd-delete"`
	CmdPauseTube        int    `yaml:"cmd-pause-tube"`
	CurrentJobsBuried   int    `yaml:"current-jobs-buried"`
	CurrentJobsDelayed  int    `yaml:"current-jobs-delayed"`
	CurrentJobsReady    int    `yaml:"current-jobs-ready"`
	CurrentJobsReserved int    `yaml:"current-jobs-reserved"`
	CurrentJobsUrgent   int    `yaml:"current-jobs-urgent"`
	CurrentUsing        int    `yaml:"current-using"`
	CurrentWaiting      int    `yaml:"current-waiting"`
	CurrentWatching     int    `yaml:"current-watching"`
	Name                string `yaml:"name"`
	Pause               int    `yaml:"pause"`
	PauseTimeLeft       int    `yaml:"pause-time-left"`
	TotalJobs           int    `yaml:"total-jobs"`
}
