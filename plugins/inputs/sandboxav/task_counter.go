package sandboxav

import (
	"database/sql"
	"sync"
	"strings"
	"regexp"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	// go-mssqldb initialization
	//_ "github.com/zensqlmonitor/go-mssqldb"
	_ "github.com/denisenkom/go-mssqldb"
)

type Task struct {
	Servers []string `toml:"servers"`
	GroupBy []string `toml:"groupby"`
}

type TaskResult struct {
	Number uint32  `toml:"number"`
	GroupBy string  `toml:"group_by"`
}

var taskSampleConfig = `
    interval = "300s"
    ## check the number fo jobs for specify status.
    # servers = [
    #     "Server=192.168.1.10;Port=1433;User Id=<user>;Password=<pw>;Database=sandbox;Workstation ID=<colo>;",
    # ]
    # 
    ## Group by
    # groupby = ["utm_serial_number", "task_type", "session_type, submit_method"]
`

func (_ *Task) SampleConfig() string {
	return taskSampleConfig
}

func (_ *Task) Description() string {
	return "MAF: check task number with specify column."
}

func (t *Task) Gather(acc telegraf.Accumulator) error {
	if len(t.Servers) == 0 {
		t.Servers = []string{"Server=.;Port=1433;Database=master;app name=maf;log=1;Workstation ID=localhost"}
	}
	if len(t.GroupBy) == 0 {
		t.GroupBy = []string{"task_type"}
	}

	var wg sync.WaitGroup

	for _, server := range t.Servers {
		for _, groupby := range t.GroupBy {
			wg.Add(1)

			go func(server string, groupby string) {
				wg.Done()
				acc.AddError(t.gatherTask(server, groupby, acc))
			}(server, groupby)
		}
	}
	wg.Wait()
	return nil
}

func (_ *Task) gatherTask(server string, groupby string, acc telegraf.Accumulator) error {
	workstation := strings.Split(strings.Split(server, ";")[5], "=")[1]

	conn, err := sql.Open("mssql", server)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		return err
	}

	var sql string
	var columnName string
	var measurement string
	if match, _ := regexp.MatchString(".*,.*",  groupby); match {
		groupSlice := strings.Split(groupby, ",")
		newGroupSlice := make([]string, 0)
		for _, group := range groupSlice {
			newGroupSlice = append(newGroupSlice, group)
		}
		sql = fmt.Sprintf(`SELECT top 10 count(*) as number, %s + '_' + %s as method from tasks group by %s order by number desc`,
			newGroupSlice[0], newGroupSlice[1], groupby)
		columnName = "method"
		measurement = "task_methods"
	} else {
		sql = fmt.Sprintf(`SELECT top 10 count(*) as number, %s FROM tasks GROUP BY %s order by number desc`,
			groupby, groupby)
		columnName = groupby
		measurement = fmt.Sprintf("task_%s", columnName)
	}

	stmt, err := conn.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var rowsData []*TaskResult
	for rows.Next() {
		var row = new(TaskResult)
		rows.Scan(&row.Number, &row.GroupBy)
		rowsData = append(rowsData, row)
	}

	for _, oneRow := range rowsData {
		acc.AddFields(measurement,
			map[string]interface{}{
				"number": oneRow.Number,
			},
			map[string]string{
				columnName: oneRow.GroupBy,
				"server": workstation,
			},
		)
	}

	return nil
}

func init() {
	inputs.Add("task_counter", func() telegraf.Input {
		return &Task{}
	})
}
