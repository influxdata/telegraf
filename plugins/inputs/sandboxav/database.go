package sandboxav

import (
	"database/sql"
	"sync"
	"strings"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	// go-mssqldb initialization
	//_ "github.com/zensqlmonitor/go-mssqldb"
	_ "github.com/denisenkom/go-mssqldb"
	"log"
)

type Database struct {
	Gap uint `toml: "gap"`
	Timeout uint `toml: "timeout"`
	Servers []string `toml: "servers"`
}

type Job struct {
	sampleSha256 string
	analyzeFeature string
	status string
	analyzeResult string
	createTime uint64
	startTime uint64
	finishTime uint64
}
type Task struct {
	sampleSha256 string
	utmSerialNumber string
	sessionBegin uint64
	sessionType string
	method string
}

type Sample struct {
	sha256 string
	taskType string
	fileSize uint64
	status string
	createDate uint64
}

var dbSampleConfig = `
    # interval = "600s"
    ##  now - (gap + timeout) < create_time <= now - timeout
    ## gap should be the same with interval, but it's uint, not string.'
    # gap = 600
    ## sandboxav timeout, default it's 15m.' 
    # timeout = 900
    ## check the  sandboxav database status.
    # servers = [
    #     "Server=<DB-SERVER>;Port=<DB-PORT>;User Id=<DB-USER>;Password=<DB-PW>;Database=<DB-NAME>;Workstation ID=<colo>;",
    # ]
`

func (_ *Database) SampleConfig() string {
	return dbSampleConfig
}

func (_ *Database) Description() string {
	return "MAF: check sandboxav database status."
}

func (db *Database) Gather(acc telegraf.Accumulator) error {
	if len(db.Servers) == 0 {
		db.Servers = []string{"Server=.;Port=1433;Database=master;app name=maf;log=1;Workstation ID=localhost"}
	}

	var wg sync.WaitGroup

	for _, server := range db.Servers {
		wg.Add(1)
		go func(server string) {
			wg.Done()
			acc.AddError(db.queryDatabase(server))
		}(server)
	}
}

func (db *Database) queryDatabase(server string) {
	workstation := strings.Split(strings.Split(server, ";")[5], "=")[1]

	conn, err := sql.Open("mssql", server)
	if err != nil {
		return err
	}

	if err := conn.Ping(); err != nil {
		return err
	}

	defer conn.Close()

    // Get job status.
	stmtJob, err := conn.Prepare(`declare @now numeric;
set @now = DATEDIFF(s, '1970-01-01 00:00:00', GETUTCDATE());
select sample_sha256, analyze_feature, status, analyze_result, create_time, start_time, finish_time 
from jobs WITH (NOLOCK)
where @now - (? + ?) < create_time and create_time <= @now - ?`)
	if err != nil {
		return err
	}

	rowsJob, err := stmtJob.Query(db.Gap, db.Timeout, db.Timeout)
	if err != nil {
		return err
	}

	var jobs []*Job
	for rowsJob.Next() {
		var job = new(Job)
		rowsJob.Scan(&job.sampleSha256, &job.analyzeFeature, &job.status, &job.analyzeResult, &job.createTime, &job.startTime, &job.finishTime)
		jobs = append(jobs, job)
	}

	rowsJob.Close()
	stmtJob.Close()

	// Get task status.
	stmtTask, err := conn.Prepare(`declare @now numeric;
set @now = DATEDIFF(s, '1970-01-01 00:00:00', GETUTCDATE());
select sample_sha256, utm_serial_number, session_begin, session_type + '_' + submit_method method
from tasks WITH (NOLOCK)
where @now - (? + ?) < session_begin and session_begin <= @now - ?`)
	if err != nil {
		return err
	}

	rowsTask, err := stmtTask.Query(db.Gap, db.Timeout, db.Timeout)
	if err != nil {
		return err
	}

	var tasks []*Task
	for rowsTask.Next() {
		var task = new(Task)
		rowsTask.Scan(&task.sampleSha256, &task.utmSerialNumber, &task.sessionBegin, &task.method)
		tasks = append(tasks, task)
	}

	rowsTask.Close()
	stmtTask.Close()

	// Get sample status.
	stmtSample, err := conn.Prepare(`declare @now numeric;
set @now = DATEDIFF(s, '1970-01-01 00:00:00', GETUTCDATE());
select sha256, task_type, file_size, status, create_date
from samples WITH (NOLOCK)
where @now - (? + ?) < create_date and create_date <= ?`)
	if err != nil {
		return err
	}

	rowsSample, err := stmtSample.Query(db.Gap, db.Timeout, db.Timeout)
	if err != nil {
		return err
	}

	var samples []*Sample
	for rowsSample.Next() {
		var sample = new(Sample)
		rowsSample.Scan(&sample.sha256, &sample.taskType, &sample.fileSize, &sample.status, &sample.createDate)
		samples = append(samples, sample)
	}

	rowsSample.Close()
	stmtSample.Close()

	// Analyze result.
	var allSamples = make(map[string]*Sample, 0)
	var snCounter = make(map[string]uint64, 0)
    var methodCounter = make(map[string]uint64, 0)
    var typeCounter = make(map[string]uint64, 0)

    type Af map[string]uint64
	"count": 1, "sum_pending": 0, "sum_running": 0, "failure": 0, "unknown": 0, "good": 0, "bad": 0,
    jobCounter := make(map[string]map[string]uint64)

    var jobCounter = make(map[string]Af, 0)
	coloCounter := map[string]uint64{
		"submit":     0,
		"unique":     0,
		"submitsize": 0,
		"uniquesize": 0,
		"good":       0,
		"bad":        0,
		"failure":    0,
	}

	// analyze samples
	for _, oneRow := range samples {
		allSamples[oneRow.sha256] = oneRow
		if len(oneRow.taskType) > 0 {
			if _, ok := typeCounter[oneRow.taskType]; ok {
				typeCounter[oneRow.taskType] += 1
			} else {
				typeCounter[oneRow.taskType] = 1
			}
		}
		counter["unique"] += 1
		counter["uniquesize"] += oneRow.fileSize
		switch oneRow.status {
		case "good":
			counter["good"] += 1
		case "bad":
			counter["bad"] += 1
		case "failure":
			counter["failure"] += 1
		}
	}

	// analyze task
	for _, oneRow := range tasks {
		if len(oneRow.utmSerialNumber) > 0 {
			if _, ok := snCounter[oneRow.utmSerialNumber]; ok {
				snCounter[oneRow.utmSerialNumber] += 1
			} else {
				snCounter[oneRow.utmSerialNumber] = 1
			}
		}
		if len(oneRow.method) > 0 {
			if _, ok := methodCounter[oneRow.method]; ok {
				methodCounter[oneRow.method] += 1
			} else {
				methodCounter[oneRow.method] = 1
			}
		}
		counter["submit"] += 1
		if _, ok := allSamples[oneRow.sampleSha256]; ok {
			counter["submitsize"] += allSamples[oneRow.sampleSha256].fileSize
		}
	}

	// analyze job
	for _, oneRow := range jobs {
		if len(oneRow.analyzeFeature) > 0 {
			var af Af
			if _, ok := jobCounter[oneRow.analyzeFeature]; ok {
				af = jobCounter[oneRow.analyzeFeature]
			} else {
				af = map[string]uint64{
					"count": 1, "sum_pending": 0, "sum_running": 0, "failure": 0, "unknown": 0, "good": 0, "bad": 0,
				}
				jobCounter[oneRow.analyzeFeature] = af
			}
			af["count"] += 1
			af["sum_pending"] += oneRow.startTime - oneRow.createTime
			af["sum_running"] += oneRow.finishTime - oneRow.startTime
			if oneRow.status == "failure" {
				af["failure"] += 1
			} else {
				if oneRow.analyzeResult == "unknown" {
					af["unknown"] += 1
				} else {
					if result, _ := strconv.Atoi(oneRow.analyzeResult); result >= int(50) {
						af["bad"] += 1
					} else {
						af["good"] += 1
					}
				}
			}
		}
	}

	// dump to influxdb
	for sn, number := range snCounter {
		acc.AddFields("tasks_sn",
			map[string]interface{}{
				"value": number,
			},
			map[string]string{
				"sn": sn,
				"server": workstation,
			},
		)
	}

	for method, number := range methodCounter {
		acc.AddFields("tasks_method",
			map[string]interface{}{
				"value": number,
			},
			map[string]string{
				"method": method,
				"server": workstation,
			},
		)
	}

	for mytype, number := range typeCounter {
		acc.AddFields("samples_type",
			map[string]interface{}{
				"value": number,
			},
			map[string]string{
				"type": mytype,
				"server": workstation,
			},
		)
	}

	var counter := map[string]uint64{}
	for sn, number := range counter {
		acc.AddFields("colos",
			map[string]interface{}{
				"value": number,
			},
			map[string]string{
				"sn": sn,
				"server": workstation,
			},
		)
	}
	obj = { "measurement": "colos",
		"tags": {"colo": k},
		"time": self.timestr,
		"fields": self.dict[k]}
	objs.append(obj)

	"count": 1, "sum_pending": 0, "sum_running": 0, "failure": 0, "unknown": 0, "good": 0, "bad": 0,
	for analyzeFeature, af := range jobCounter {
		acc.AddFields("jobs_type",
			map[string]interface{}{
				"value": af.count,
				"failure": af.failure,
				"unknown": af.unknown,
				"good": af.good,
				"bad": af.bad,
				"pending_time": af.sum_pending/af.count,
				"running_time": af.sum_running/af.count,
			},
			map[string]string{
				"type": analyzeFeature,
				"server": workstation,
			},
		)
	}


	return nil
}

func init() {
	inputs.Add("sandboxav", func() telegraf.Input {
		return &Database{}
	})
}