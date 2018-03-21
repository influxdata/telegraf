package maf

import (
	"database/sql"
	"sync"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	// go-mssqldb initialization
	//_ "github.com/zensqlmonitor/go-mssqldb"
	_ "github.com/denisenkom/go-mssqldb"
)

type Task struct {
	Servers []string `toml:"servers"`
	Status []string `toml:"status"`
}

type Result struct {
	Number uint32  `toml:"number"`
	AnalyzeType string  `toml:"analyze_type"`
}

var jobSampleConfig = `
    interval = "300s"
    ## check the number fo jobs for specify status.
    # servers = [
    #     "Server=192.168.1.10;Port=1433;User Id=<user>;Password=<pw>;Database=sandbox;Workstation ID=<colo>;",
    # ]
    # 
    # status = ["success", "pending", "running", "failure"]
    # status = ["pending"]
    ## analyze_type
    ## [lastline, Office prefilter, PE signature prefilter, reversinglab, SMASH, SonicSandbox, static, virustotal, varay]
`

func (_ *Job) SampleConfig() string {
	return jobSampleConfig
}

func (_ *Job) Description() string {
	return "MAF: check job number with specify status."
}

func (j *Job) Gather(acc telegraf.Accumulator) error {
	if len(j.Servers) == 0 {
		j.Servers = []string{"Server=.;Port=1433;Database=master;app name=maf;log=1;Workstation ID=localhost"}
	}
	if len(j.Status) == 0 {
		j.Status = []string{"pending"}
	}

	var wg sync.WaitGroup

	for _, server := range j.Servers {
		for _, status := range j.Status {
			wg.Add(1)

			go func(server string, status string) {
				wg.Done()
				acc.AddError(j.gatherJobs(server, status, acc))
			}(server, status)
		}
	}
	wg.Wait()
	return nil
}

func (_ *Job) gatherJobs(server string, status string, acc telegraf.Accumulator) error {
	workstation := strings.Split(strings.Split(server, ";")[5], "=")[1]

	conn, err := sql.Open("mssql", server)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		return err
	}

	stmt, err := conn.Prepare(`SELECT count(*) as number, analyze_type FROM jobs WHERE status=? GROUP BY analyze_type`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(status)
	if err != nil {
		return err
	}
	defer rows.Close()

	var rowsData []*Result
	for rows.Next() {
		var row = new(Result)
		rows.Scan(&row.Number, &row.AnalyzeType)
		rowsData = append(rowsData, row)
	}

	for _, oneRow := range rowsData {
		acc.AddFields("maf_job_status",
			map[string]interface{}{
				"number": oneRow.Number,
			},
			map[string]string{
				"status": status,
				"analyze_type": oneRow.AnalyzeType,
				"server": workstation,
			},
		)
	}

	return nil
}

func init() {
	inputs.Add("maf_job_status", func() telegraf.Input {
		return &Job{}
	})
}
