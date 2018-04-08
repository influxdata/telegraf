package sandboxav

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

type Sample struct {
	Servers []string `toml:"servers"`
	Status []string `toml:"status"`
}

type SampleResult struct {
	Number uint32  `toml:"number"`
	TaskType string  `toml:"task_type"`
}

var sampleSampleConfig = `
    # interval = "300s"
    ## check the number fo samples for specify status.
    # servers = [
    #     "Server=192.168.1.10;Port=1433;User Id=<user>;Password=<pw>;Database=sandbox;Workstation ID=<colo>;",
    # ]
    ## Default all status.
    # status = ["good", "bad", "pending", "running", "failure"]
    ## Task type, except "NULL"
    # task_type = ["apk", "archive", "jar", "msi_x32", "office", "pdf", "pe_x32", "pe_x64", "script", "unknown"]
`
func (_ *Sample) SampleConfig() string {
	return sampleSampleConfig
}

func (_ *Sample) Description() string {
	return "MAF: check samples number with specify status."
}

func (s *Sample) Gather(acc telegraf.Accumulator) error {
	if len(s.Servers) == 0 {
		s.Servers = []string{"Server=.;Port=1433;Database=master;app name=maf;log=1;Workstation Id=localhost"}
	}
	if len(s.Status) == 0 {
		s.Status = []string{"pending"}
	}

	var wg sync.WaitGroup

	for _, server := range s.Servers {
		for _, status := range s.Status {
			wg.Add(1)

			go func(server string, status string) {
				wg.Done()
				acc.AddError(s.gatherSamples(server, status, acc))
			}(server, status)
		}
	}
	wg.Wait()
	return nil
}

func (_ *Sample) gatherSamples(server string, status string, acc telegraf.Accumulator) error {
	workstation := strings.Split(strings.Split(server, ";")[5], "=")[1]

	conn, err := sql.Open("mssql", server)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		return err
	}

	stmt, err := conn.Prepare(`select count(*) as number, task_type from samples where status=? group by task_type`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(status)
	if err != nil {
		return err
	}
	defer rows.Close()

	var rowsData []*SampleResult
	for rows.Next() {
		var row = new(SampleResult)
		rows.Scan(&row.Number, &row.TaskType)
		rowsData = append(rowsData, row)
	}

	for _, oneRow := range rowsData {
		acc.AddFields("sample_counter",
			map[string]interface{}{
				"number": oneRow.Number,
			},
			map[string]string{
				"status": status,
				"task_type": oneRow.TaskType,
				"server": workstation,
			},
		)
	}
	return nil
}

func init() {
	inputs.Add("sample_counter", func() telegraf.Input {
		return &Sample{}
	})
}
