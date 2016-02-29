package mysql

import (
	"database/sql"
	"net/url"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Mysql struct {
	Servers []string
}

var sampleConfig = `
  ## specify servers via a url matching:
  ##  [username[:password]@][protocol[(address)]]/[?tls=[true|false|skip-verify]]
  ##  see https://github.com/go-sql-driver/mysql#dsn-data-source-name
  ##  e.g.
  ##    root:passwd@tcp(127.0.0.1:3306)/?tls=false
  ##    root@tcp(127.0.0.1:3306)/?tls=false
  ##
  ## If no servers are specified, then localhost is used as the host.
  servers = ["tcp(127.0.0.1:3306)/"]
`

var defaultTimeout = time.Second * time.Duration(5)

func (m *Mysql) SampleConfig() string {
	return sampleConfig
}

func (m *Mysql) Description() string {
	return "Read metrics from one or many mysql servers"
}

var localhost = ""

func (m *Mysql) Gather(acc telegraf.Accumulator) error {
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
		onServer: "Aborted_",
		inExport: "aborted_",
	},
	{
		onServer: "Bytes_",
		inExport: "bytes_",
	},
	{
		onServer: "Com_",
		inExport: "commands_",
	},
	{
		onServer: "Created_",
		inExport: "created_",
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
		onServer: "Key_",
		inExport: "key_",
	},
	{
		onServer: "Open_",
		inExport: "open_",
	},
	{
		onServer: "Opened_",
		inExport: "opened_",
	},
	{
		onServer: "Qcache_",
		inExport: "qcache_",
	},
	{
		onServer: "Table_",
		inExport: "table_",
	},
	{
		onServer: "Tokudb_",
		inExport: "tokudb_",
	},
	{
		onServer: "Threads_",
		inExport: "threads_",
	},
}

func (m *Mysql) gatherServer(serv string, acc telegraf.Accumulator) error {
	// If user forgot the '/', add it
	if strings.HasSuffix(serv, ")") {
		serv = serv + "/"
	} else if serv == "localhost" {
		serv = ""
	}

	serv, err := dsnAddTimeout(serv)
	if err != nil {
		return err
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

	var servtag string
	servtag, err = parseDSN(serv)
	if err != nil {
		servtag = "localhost"
	}
	tags := map[string]string{"server": servtag}
	fields := make(map[string]interface{})
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
				fields[mapped.inExport+name[len(mapped.onServer):]] = i
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

			fields["queries"] = i
		case "Slow_queries":
			i, err := strconv.ParseInt(string(val.([]byte)), 10, 64)
			if err != nil {
				return err
			}

			fields["slow_queries"] = i
		}
	}
	acc.AddFields("mysql", fields, tags)

	conn_rows, err := db.Query("SELECT user, sum(1) FROM INFORMATION_SCHEMA.PROCESSLIST GROUP BY user")

	for conn_rows.Next() {
		var user string
		var connections int64

		err = conn_rows.Scan(&user, &connections)
		if err != nil {
			return err
		}

		tags := map[string]string{"server": servtag, "user": user}
		fields := make(map[string]interface{})

		if err != nil {
			return err
		}
		fields["connections"] = connections
		acc.AddFields("mysql_users", fields, tags)
	}

	return nil
}

func dsnAddTimeout(dsn string) (string, error) {

	// DSN "?timeout=5s" is not valid, but "/?timeout=5s" is valid ("" and "/"
	// are the same DSN)
	if dsn == "" {
		dsn = "/"
	}
	u, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}
	v := u.Query()

	// Only override timeout if not already defined
	if _, ok := v["timeout"]; ok == false {
		v.Add("timeout", defaultTimeout.String())
		u.RawQuery = v.Encode()
	}
	return u.String(), nil
}

func init() {
	inputs.Add("mysql", func() telegraf.Input {
		return &Mysql{}
	})
}
