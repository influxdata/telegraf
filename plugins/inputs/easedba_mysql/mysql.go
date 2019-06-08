package easedba_mysql

import (
	"database/sql"
	"fmt"
	"github.com/influxdata/telegraf/plugins/easedbautil"
	"github.com/influxdata/telegraf/plugins/inputs/easedba_mysql/global"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/easedba_mysql/v1"
)

type Mysql struct {
	Servers              []string `toml:"servers"`
	GatherDbSizes        bool     ` toml: "gather_db_sizes"`
	GatherReplication    bool     `toml:"gather_replication"`
	GatherSnapshot       bool     `toml:"gather_snapshot"`
	GatherInnodb         bool     `toml:"gather_innodb"`
	GatherGlobalStatuses bool     `toml:"gather_global_statuses"`
	GatherConnection     bool     `toml:"gather_connection_statuses"`
}

var sampleConfig = `
  ## specify servers via a url matching:
  ##  [username[:password]@][protocol[(address)]]/[?tls=[true|false|skip-verify|custom]]
  ##  see https://github.com/go-sql-driver/mysql#dsn-data-source-name
  ##  e.g.
  ##    servers = ["user:passwd@tcp(127.0.0.1:3306)/?tls=false"]
  ##    servers = ["user@tcp(127.0.0.1:3306)/?tls=false"]
  #
  ## If no servers are specified, then localhost is used as the host.
  servers = ["tcp(127.0.0.1:3306)/"]

  ## gather metrics of total table size, total index size, binlog size 
  gather_db_sizes							= true

  ## gather slave and master status
  gather_replication						= true

  ## gather the running sql ,transcation snapshots
  gather_snapshot							= true

`

var defaultTimeout = time.Second * time.Duration(5)

func (m *Mysql) SampleConfig() string {
	return sampleConfig
}

func (m *Mysql) Description() string {
	return "Read metrics from one or many mysql servers"
}

func (m *Mysql) Gather(acc telegraf.Accumulator) error {
	if len(m.Servers) == 0 {
		return fmt.Errorf("error: not found any mysql servers for monitoring.")
	}

	var wg sync.WaitGroup

	// Loop through each server and collect metrics
	for _, server := range m.Servers {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			acc.AddError(m.gatherServer(s, acc))
		}(server)
	}

	wg.Wait()
	return nil
}

const (
	globalStatusQuery = `SHOW GLOBAL STATUS`
	binaryLogsQuery   = `SHOW BINARY LOGS`
)

func (m *Mysql) gatherServer(server string, acc telegraf.Accumulator) error {
	server, err := dsnAddTimeout(server)
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", server)
	if err != nil {
		return err
	}

	defer db.Close()

	servtag := getDSNTag(server)

	status, ok := easedba_v1.GlobalStatus[servtag]
	if !ok {
		status = global.New(servtag)
		easedba_v1.GlobalStatus[servtag] = status
	}

	err = status.Fill(db)
	if err != nil {
		return err
	}

	//throughput index
	if m.GatherGlobalStatuses {
		err = m.gatherThroughput(db, server, acc, servtag)

		if err != nil {
			return err
		}
	}

	//add megaeasdba index
	if m.GatherConnection {
		err = m.gatherConnection(db, server, acc, servtag)
		if err != nil {
			return err
		}
	}

	if m.GatherInnodb {
		err = m.gatherInnodb(db, server, acc, servtag)
		if err != nil {
			return err
		}
	}

	if m.GatherDbSizes {
		err = m.gatherDbSizes(db, server, acc, servtag)
		if err != nil {
			return err
		}

	}

	if m.GatherReplication {
		err = m.gatherReplication(db, server, acc, servtag)
		if err != nil {
			return err
		}

	}

	if m.GatherSnapshot {
		err = m.gatherSnapshot(db, server, acc, servtag)
		if err != nil {
			return err
		}
	}

	return nil
}

// gatherThroughput can be used to get MySQL status metrics
// the mappings of actual names and names of each status to be exported
// to output is provided on mappings variable
func (m *Mysql) gatherThroughput(db *sql.DB, serv string, acc telegraf.Accumulator, servtag string) error {
	status, _ := easedba_v1.GlobalStatus[servtag]

	tags := map[string]string{"server": servtag}
	fields := make(map[string]interface{})

	for key := range status.CurrStatus {
		if converted, ok := easedba_v1.ThroughtMappings[key]; ok {
			delta, err := status.GetPropertyDelta(key)
			if err != nil {
				return fmt.Errorf("error getting %s throughput mertics:  %s", servtag, err)
			}
			fields[converted] = delta
		}
	}

	acc.AddFields(easedbautl.SchemaThroughput, fields, tags)

	return nil
}

// gatherconnection can be used to get MySQL status metrics
// the mappings of actual names and names of each status to be exported
// to output is provided on mappings variable
func (m *Mysql) gatherConnection(db *sql.DB, serv string, acc telegraf.Accumulator, servtag string) error {
	status, _ := easedba_v1.GlobalStatus[servtag]

	tags := map[string]string{"server": servtag}
	fields := make(map[string]interface{})
	for key := range status.CurrStatus {
		if converted, ok := easedba_v1.ConnectionMappings[key]; ok {
			val, err := status.GetPropertyDelta(key)
			if err != nil {
				return fmt.Errorf("error getting %s, connection metrics: %s", servtag, err)
			}
			fields[converted] = val
		}

	}

	acc.AddFields(easedbautl.SchemaConnection, fields, tags)
	return nil
}

// gathercinnodb can be used to get MySQL status metrics
// the mappings of actual names and names of each status to be exported
// to output is provided on mappings variable
func (m *Mysql) gatherInnodb(db *sql.DB, serv string, acc telegraf.Accumulator, servtag string) error {
	status, _ := easedba_v1.GlobalStatus[servtag]
	tags := map[string]string{"server": servtag}
	fields := make(map[string]interface{})

	for key := range status.CurrStatus {
		if converted, ok := easedba_v1.InnodbMappings[key]; ok {
			val, err := status.GetPropertyDelta(key)
			if err != nil {
				return fmt.Errorf("error getting %s innodb metrics: %s", servtag, err)
			}
			fields[converted] = val
		}
	}
	acc.AddFields(easedbautl.SchemaInnodb, fields, tags)

	return nil
}

func dsnAddTimeout(dsn string) (string, error) {
	conf, err := mysql.ParseDSN(dsn)
	if err != nil {
		return "", err
	}

	if conf.Timeout == 0 {
		conf.Timeout = time.Second * 5
	}

	return conf.FormatDSN(), nil
}

func getDSNTag(dsn string) string {
	conf, err := mysql.ParseDSN(dsn)
	if err != nil {
		return "127.0.0.1:3306"
	}
	return conf.Addr
}

func init() {
	inputs.Add("easedba_mysql", func() telegraf.Input {
		return &Mysql{}
	})
}
