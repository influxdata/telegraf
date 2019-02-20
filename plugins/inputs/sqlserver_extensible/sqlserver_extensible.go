package sqlserver_extensible

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
	"log"
	"regexp"
	"sync"

	// go-mssqldb initialization
	_ "github.com/zensqlmonitor/go-mssqldb"
)

// SQLServer struct
type SQLServer struct {
	Servers []string
	Query   []struct {
		Server         string
		Version        string
		Statement      string
		Scriptfile     string
		Measurement    string
		Tags           []string
		Fields         []string
		Fieldname      []string
		OrderedColumns []string
	}
}

// Query struct
type Query struct {
	Server         string
	Version        string
	Statement      string
	Scriptfile     string
	Measurement    string
	Tags           []string
	Fields         []string
	Fieldname      []string
	OrderedColumns []string
}

var defaultServer = "Server=.;app name=telegraf;log=1;"

var sampleConfig = `
  ## Specify instances to monitor with a list of connection strings.
  ## All connection parameters are optional.
  ## By default, the host is localhost, listening on default port, TCP 1433.
  ##   for Windows, the user is the currently running AD user (SSO).
  ##   See https://github.com/denisenkom/go-mssqldb for detailed connection
  ##   parameters.
  servers = [
   "Server=192.168.1.30;Port=3333;User Id=telegraf;Password=telegraf;app name=telegraf;log=1;",
  ]
  ## Structure :
  ## [[inputs.sqlserver_extensible.query]]
  ## version string (10, 10.5, 11, 12, 13) - minimum version able to run statement
  ## statement string
  ## scriptfile string - use a script from a a file if statement is empty 
  ## measurement string  
  ## tags array of string, column(s) must be present in your query
  ## fields array of string, column(s) must be present in your query
  ## fieldname array of string to replace fieldname, column(s) must be present in your query

  [[inputs.sqlserver_extensible.query]]
  version="12"
  statement="select replace(rtrim(counter_name),' ','_') as counter_name, replace(rtrim(instance_name),' ','_') as instance_name, cntr_value from sys.dm_os_performance_counters where (counter_name in ('SQL Compilations/sec','SQL Re-Compilations/sec','User Connections','Batch Requests/sec','Logouts/sec','Logins/sec','Processes blocked','Latch Waits/sec','Full Scans/sec','Index Searches/sec','Page Splits/sec','Page Lookups/sec','Page Reads/sec','Page Writes/sec','Readahead Pages/sec','Lazy Writes/sec','Checkpoint Pages/sec','Database Cache Memory (KB)','Log Pool Memory (KB)','Optimizer Memory (KB)','SQL Cache Memory (KB)','Connection Memory (KB)','Lock Memory (KB)', 'Memory broker clerk size','Page life expectancy')) or (instance_name in ('_Total','Column store object pool') and counter_name in ('Transactions/sec','Write Transactions/sec','Log Flushes/sec','Log Flush Wait Time','Lock Timeouts/sec','Number of Deadlocks/sec','Lock Waits/sec','Latch Waits/sec','Memory broker clerk size','Log Bytes Flushed/sec','Bytes Sent to Replica/sec','Log Send Queue','Bytes Sent to Transport/sec','Sends to Replica/sec','Bytes Sent to Transport/sec','Sends to Transport/sec','Bytes Received from Replica/sec','Receives from Replica/sec','Flow Control Time (ms/sec)','Flow Control/sec','Resent Messages/sec','Redone Bytes/sec') or (object_name = 'SQLServer:Database Replica' and counter_name in ('Log Bytes Received/sec','Log Apply Pending Queue','Redone Bytes/sec','Recovery Queue','Log Apply Ready Queue') and instance_name = '_Total')) or (object_name = 'SQLServer:Database Replica' and counter_name in ('Transaction Delay'));"
  measurement="sql_server_perf_counters"
  tags=["server_name"]
  fields=["cntr_value"]
  fieldname=["counter_name", "instance_name"]

  [[inputs.sqlserver_extensible.query]]
  version="12"
  scriptfile = "/path/to/scripts_sql/memory_clerk.sql" 
  tags=["counter_name", "server_name"]
  fields=["Buffer pool", "Cache (objects)", "Cache (sql plans)", "Other"]
`

// SampleConfig return the sample configuration
func (s *SQLServer) SampleConfig() string {
	return sampleConfig
}

// Description return plugin description
func (s *SQLServer) Description() string {
	return "Read metrics from Microsoft SQL Server"
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func feedStatement(version string, statement string) string {
	var flags string = "SET NOCOUNT ON;"
	var versionCheck string = "IF CAST(LEFT(CAST(SERVERPROPERTY('productversion') as varchar), 4) as numeric(4,2)) < " + version + "RETURN;"

	// check flags
	rexp1 := regexp.MustCompile(`(?i)SET NOCOUNT ON`)
	if rexp1.MatchString(statement) {
		flags = ""
	}
	return flags + versionCheck + statement
}

// Gather collect data from SQL Server
func (s *SQLServer) Gather(acc telegraf.Accumulator) error {

	if len(s.Servers) == 0 {
		s.Servers = append(s.Servers, defaultServer)
	}

	var query Query
	var wg sync.WaitGroup

	// foreach server
	for _, serv := range s.Servers {
		query.Server = serv

		// foreach query
		for q := range s.Query {
			var itemQ Query = s.Query[q]

			// checks
			if itemQ.Version == "" {
				return errors.New("SQL Server product version is mandatory.")
			}
			if itemQ.Statement == "" && itemQ.Scriptfile == "" {
				return errors.New("SQL statement is mandatory.")
			}
			if len(itemQ.Fields) > 1 && len(itemQ.Fieldname) > 0 {
				return errors.New("Replacement of field name allowed only if one field returned.")
			}
			// stmt
			if itemQ.Statement != "" {
				query.Statement = feedStatement(itemQ.Version, itemQ.Statement)
			} else {
				scriptFromFile, err := ioutil.ReadFile(itemQ.Scriptfile)
				query.Statement = feedStatement(itemQ.Version, string(scriptFromFile))
				if err != nil {
					var msg string = fmt.Sprintf("Error while reading script file %s: %s", itemQ.Scriptfile, err)
					return errors.New(msg)
				}
			}
			log.Printf("D! sqlserver_extensible: statement %s \n", query.Statement)

			// tags
			query.Tags = itemQ.Tags

			// measurement
			if itemQ.Measurement != "" {
				query.Measurement = itemQ.Measurement
			} else {
				query.Measurement = ""
			}
			// fields
			query.Fields = itemQ.Fields

			// fieldname
			query.Fieldname = itemQ.Fieldname

			// go routines
			wg.Add(1)
			go func(serv string, query Query) {
				defer wg.Done()
				acc.AddError(s.gatherServer(serv, query, acc))
			}(serv, query)
		}
	}

	wg.Wait()
	return nil
}

// query @@SERVERNAME if not specified
func getServerName(server string) (interface{}, error) {
	var servername string

	// deferred opening
	conn, err := sql.Open("mssql", server)
	if err != nil {
		return "", err
	}
	// verify that a connection can be made
	err = conn.Ping()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// execute query
	rows, err := conn.Query("SET NOCOUNT ON;SELECT REPLACE(@@SERVERNAME, '\\', ':') as server_name;")
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&servername); err != nil {
			return "", err
		}
	}
	return servername, rows.Err()
}

func (s *SQLServer) gatherServer(server string, query Query, acc telegraf.Accumulator) error {
	conn, err := sql.Open("mssql", server)
	if err != nil {
		return err
	}
	// verify that a connection can be made
	err = conn.Ping()
	if err != nil {
		return err
	}
	defer conn.Close()

	// execute query
	rows, err := conn.Query(query.Statement)
	if err != nil {
		return err
	}
	defer rows.Close()

	// grab the columns information
	query.OrderedColumns, err = rows.Columns()
	if err != nil {
		return err
	}
	for rows.Next() {
		err = s.accRow(query, rows, acc)
		if err != nil {
			return err
		}
	}
	return rows.Err()
}

func (s *SQLServer) accRow(query Query, row scanner, acc telegraf.Accumulator) error {

	var columnVars []interface{}
	var measurementb bytes.Buffer
	var measurement string
	var fields = make(map[string]interface{})
	tags := map[string]string{}

	// store the column name with its *interface{}
	columnMap := make(map[string]*interface{})
	for _, column := range query.OrderedColumns {
		columnMap[column] = new(interface{})
	}
	for i := 0; i < len(columnMap); i++ {
		columnVars = append(columnVars, columnMap[query.OrderedColumns[i]])
	}

	// rows scan
	err := row.Scan(columnVars...)
	if err != nil {
		return err
	}

	// check for server_name in the result
	_, ok := columnMap["server_name"]
	if !ok {
		servername, err := getServerName(query.Server)
		if err != nil {
			return err
		}
		columnMap["server_name"] = &servername
	}

	// measurement
	if query.Measurement == "" {
		measurementb.WriteString((*columnMap["measurement"]).(string))
		measurement = measurementb.String()
		delete(columnMap, "measurement")
	} else {
		measurement = query.Measurement
	}

	// columnMap
COLUMN:

	for col, val := range columnMap {
		log.Printf("D! sqlserver_extensible: column: %s = %T: %s\n", col, *val, *val)

		// tags
		for _, tag := range query.Tags {
			if col != tag {
				continue
			}
			switch v := (*val).(type) {
			case string:
				tags[col] = v
			case []byte:
				tags[col] = string(v)
			case int64, int32, int:
				tags[col] = fmt.Sprintf("%d", v)
			default:
				log.Println("Failed to add additional tag", col)
			}
			continue COLUMN
		}

		// fields
		for _, field := range query.Fields {
			if col != field {
				continue
			}
			// default fieldname
			var fieldheader string = col

			// custom fieldname
			if len(query.Fieldname) > 0 {
				fieldheader = ""
				for _, fname := range query.Fieldname {
					_, ok := columnMap[fname]
					if ok {
						fheader := (*columnMap[fname]).(string)
						if len(fheader) > 0 {
							fieldheader += fheader + " | "
						}
					}
				}
				sz := len(fieldheader)
				if sz > 0 && fieldheader[sz-2] == '|' {
					fieldheader = fieldheader[:sz-3]
				}
			}
			if v, ok := (*val).([]byte); ok {
				fields[fieldheader] = string(v)
			} else {
				fields[fieldheader] = *val
			}
			continue COLUMN
		}
	}

	acc.AddFields(measurement, fields, tags)
	return nil
}

func init() {
	inputs.Add("sqlserver_extensible", func() telegraf.Input {
		return &SQLServer{}
	})
}
