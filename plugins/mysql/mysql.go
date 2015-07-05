package mysql

import (
	"database/sql"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/influxdb/telegraf/plugins"
)

type Mysql struct {
	Servers []string
}

var sampleConfig = `
# specify servers via a url matching:
#  [username[:password]@][protocol[(address)]]/[?tls=[true|false|skip-verify]]
#  e.g. root:root@http://10.0.0.18/?tls=false
#
# If no servers are specified, then localhost is used as the host.
servers = ["localhost"]`

func (m *Mysql) SampleConfig() string {
	return sampleConfig
}

func (m *Mysql) Description() string {
	return "Read metrics from one or many mysql servers"
}

var localhost = ""

func (m *Mysql) Gather(acc plugins.Accumulator) error {
	if len(m.Servers) == 0 {
		// if we can't get stats in this case, thats fine, don't report
		// an error.
		m.gatherServer(localhost, acc)
		return nil
	}

	for _, serv := range m.Servers {
		err := m.gatherServer(serv, acc)
		if err != nil {
			return err
		}
	}

	return nil
}

type mapping struct {
	onServer string
	inExport string
}

var mappings = []*mapping{
	{
		onServer: "Bytes_",
		inExport: "bytes_",
	},
	{
		onServer: "Com_",
		inExport: "commands_",
	},
	{
		onServer: "Handler_",
		inExport: "handler_",
	},
	{
		onServer: "Innodb_",
		inExport: "innodb_",
	},
	{
		onServer: "Threads_",
		inExport: "threads_",
	},
}

func (m *Mysql) gatherServer(serv string, acc plugins.Accumulator) error {
	if serv == "localhost" {
		serv = ""
	}

	db, err := sql.Open("mysql", serv)
	if err != nil {
		return err
	}

	defer db.Close()

	rows, err := db.Query(`SHOW /*!50002 GLOBAL */ STATUS`)
	if err != nil {
		return err
	}

	for rows.Next() {
		var name string
		var val interface{}

		err = rows.Scan(&name, &val)
		if err != nil {
			return err
		}

		var found bool

		for _, mapped := range mappings {
			if strings.HasPrefix(name, mapped.onServer) {
				i, _ := strconv.Atoi(string(val.([]byte)))
				acc.Add(mapped.inExport+name[len(mapped.onServer):], i, nil)
				found = true
			}
		}

		if found {
			continue
		}

		switch name {
		case "Queries":
			i, err := strconv.ParseInt(string(val.([]byte)), 10, 64)
			if err != nil {
				return err
			}

			acc.Add("queries", i, nil)
		case "Slow_queries":
			i, err := strconv.ParseInt(string(val.([]byte)), 10, 64)
			if err != nil {
				return err
			}

			acc.Add("slow_queries", i, nil)
		}
	}

	return nil
}

func init() {
	plugins.Add("mysql", func() plugins.Plugin {
		return &Mysql{}
	})
}
