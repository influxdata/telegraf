package sqlserver

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/denisenkom/go-mssqldb" // go-mssqldb initialization
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// SQLServer struct
type SQLServer struct {
	Servers       []string `toml:"servers"`
	QueryVersion  int      `toml:"query_version"`
	AzureDB       bool     `toml:"azuredb"`
	DatabaseType  string   `toml:"database_type"`
	IncludeQuery  []string `toml:"include_query"`
	ExcludeQuery  []string `toml:"exclude_query"`
	queries       MapQuery
	isInitialized bool
}

// Query struct
type Query struct {
	ScriptName     string
	Script         string
	ResultByRow    bool
	OrderedColumns []string
}

// MapQuery type
type MapQuery map[string]Query

const defaultServer = "Server=.;app name=telegraf;log=1;"

const sampleConfig = `
## Specify instances to monitor with a list of connection strings.
## All connection parameters are optional.
## By default, the host is localhost, listening on default port, TCP 1433.
##   for Windows, the user is the currently running AD user (SSO).
##   See https://github.com/denisenkom/go-mssqldb for detailed connection
##   parameters, in particular, tls connections can be created like so:
##   "encrypt=true;certificate=<cert>;hostNameInCertificate=<SqlServer host fqdn>"
# servers = [
#  "Server=192.168.1.10;Port=1433;User Id=<user>;Password=<pw>;app name=telegraf;log=1;",
# ]

## This enables a specific set of queries depending on the database type. If specified, it replaces azuredb = true/false and query_version = 2
## In the config file, the sql server plugin section should be repeated  each with a set of servers for a specific database_type.
## Possible values for database_type are  
## "AzureSQLDB" 
## "SQLServer"
## "AzureSQLManagedInstance"
# database_type = "AzureSQLDB"


## Optional parameter, setting this to 2 will use a new version
## of the collection queries that break compatibility with the original
## dashboards.
## Version 2 - is compatible from SQL Server 2012 and later versions and also for SQL Azure DB
query_version = 2

## If you are using AzureDB, setting this to true will gather resource utilization metrics
# azuredb = false

## Possible queries
## Version 2:
## - PerformanceCounters
## - WaitStatsCategorized
## - DatabaseIO
## - ServerProperties
## - MemoryClerk
## - Schedulers
## - SqlRequests
## - VolumeSpace
## - Cpu

## Version 1:
## - PerformanceCounters
## - WaitStatsCategorized
## - CPUHistory
## - DatabaseIO
## - DatabaseSize
## - DatabaseStats
## - DatabaseProperties
## - MemoryClerk
## - VolumeSpace
## - PerformanceMetrics


## Queries enabled by default for specific Database Type
## database_type =  AzureSQLDB
	## AzureDBWaitStats, AzureDBResourceStats, AzureDBResourceGovernance, sqlAzureDBDatabaseIO

## A list of queries to include. If not specified, all the above listed queries are used.
# include_query = []

## A list of queries to explicitly ignore.
exclude_query = [ 'Schedulers' , 'SqlRequests']
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

func initQueries(s *SQLServer) error {
	s.queries = make(MapQuery)
	queries := s.queries
	log.Printf("I! [inputs.sqlserver] Config: database_type: %s , query_version:%d , azuredb: %t", s.DatabaseType, s.QueryVersion, s.AzureDB)

	// New config option database_type
	// To prevent query definition conflicts
	// Constant defintiions for type "AzureSQLDB" start with sqlAzureDB
	// Constant defintiions for type "AzureSQLManagedInstance" start with sqlAzureMI
	// Constant defintiions for type "SQLServer" start with sqlServer
	if s.DatabaseType == "AzureSQLDB" {
		queries["AzureSQLDBResourceStats"] = Query{ScriptName: "AzureSQLDBResourceStats", Script: sqlAzureDBResourceStats, ResultByRow: false}
		queries["AzureSQLDBResourceGovernance"] = Query{ScriptName: "AzureSQLDBResourceGovernance", Script: sqlAzureDBResourceGovernance, ResultByRow: false}
		queries["AzureSQLDBWaitStats"] = Query{ScriptName: "AzureSQLDBWaitStats", Script: sqlAzureDBWaitStats, ResultByRow: false}
		queries["AzureSQLDBDatabaseIO"] = Query{ScriptName: "AzureSQLDBDatabaseIO", Script: sqlAzureDBDatabaseIO, ResultByRow: false}
		queries["AzureSQLDBServerProperties"] = Query{ScriptName: "AzureSQLDBServerProperties", Script: sqlAzureDBProperties, ResultByRow: false}
		queries["AzureSQLDBOsWaitstats"] = Query{ScriptName: "AzureSQLOsWaitstats", Script: sqlAzureDBOsWaitStats, ResultByRow: false}
		queries["AzureSQLDBMemoryClerks"] = Query{ScriptName: "AzureSQLDBMemoryClerks", Script: sqlAzureDBMemoryClerks, ResultByRow: false}
		queries["AzureSQLDBPerformanceCounters"] = Query{ScriptName: "AzureSQLDBPerformanceCounters", Script: sqlAzureDBPerformanceCounters, ResultByRow: false}
		queries["AzureSQLDBRequests"] = Query{ScriptName: "AzureSQLDBRequests", Script: sqlAzureDBRequests, ResultByRow: false}
		queries["AzureSQLDBSchedulers"] = Query{ScriptName: "AzureSQLDBSchedulers", Script: sqlServerSchedulers, ResultByRow: false}
	} else if s.DatabaseType == "AzureSQLManagedInstance" {
		queries["AzureSQLMIResourceStats"] = Query{ScriptName: "AzureSQLMIResourceStats", Script: sqlAzureMIResourceStats, ResultByRow: false}
		queries["AzureSQLMIResourceGovernance"] = Query{ScriptName: "AzureSQLMIResourceGovernance", Script: sqlAzureMIResourceGovernance, ResultByRow: false}
		queries["AzureSQLMIDatabaseIO"] = Query{ScriptName: "AzureSQLMIDatabaseIO", Script: sqlAzureMIDatabaseIO, ResultByRow: false}
		queries["AzureSQLMIServerProperties"] = Query{ScriptName: "AzureSQLMIServerProperties", Script: sqlAzureMIProperties, ResultByRow: false}
		queries["AzureSQLMIOsWaitstats"] = Query{ScriptName: "AzureSQLMIOsWaitstats", Script: sqlAzureMIOsWaitStats, ResultByRow: false}
		queries["AzureSQLMIMemoryClerks"] = Query{ScriptName: "AzureSQLMIMemoryClerks", Script: sqlAzureMIMemoryClerks, ResultByRow: false}
		queries["AzureSQLMIPerformanceCounters"] = Query{ScriptName: "AzureSQLMIPerformanceCounters", Script: sqlAzureMIPerformanceCounters, ResultByRow: false}
		queries["AzureSQLMIRequests"] = Query{ScriptName: "AzureSQLMIRequests", Script: sqlAzureMIRequests, ResultByRow: false}
		queries["AzureSQLMISchedulers"] = Query{ScriptName: "AzureSQLMISchedulers", Script: sqlServerSchedulers, ResultByRow: false}
	} else if s.DatabaseType == "SQLServer" { //These are still V2 queries and have not been refactored yet.
		queries["SQLServerPerformanceCounters"] = Query{ScriptName: "SQLServerPerformanceCounters", Script: sqlServerPerformanceCounters, ResultByRow: false}
		queries["SQLServerWaitStatsCategorized"] = Query{ScriptName: "SQLServerWaitStatsCategorized", Script: sqlServerWaitStatsCategorized, ResultByRow: false}
		queries["SQLServerDatabaseIO"] = Query{ScriptName: "SQLServerDatabaseIO", Script: sqlServerDatabaseIO, ResultByRow: false}
		queries["SQLServerProperties"] = Query{ScriptName: "SQLServerProperties", Script: sqlServerProperties, ResultByRow: false}
		queries["SQLServerMemoryClerks"] = Query{ScriptName: "SQLServerMemoryClerks", Script: sqlServerMemoryClerks, ResultByRow: false}
		queries["SQLServerSchedulers"] = Query{ScriptName: "SQLServerSchedulers", Script: sqlServerSchedulers, ResultByRow: false}
		queries["SQLServerRequests"] = Query{ScriptName: "SQLServerRequests", Script: sqlServerRequests, ResultByRow: false}
		queries["SQLServerVolumeSpace"] = Query{ScriptName: "SQLServerVolumeSpace", Script: sqlServerVolumeSpace, ResultByRow: false}
		queries["SQLServerCpu"] = Query{ScriptName: "SQLServerCpu", Script: sqlServerRingBufferCpu, ResultByRow: false}
	} else {
		// If this is an AzureDB instance, grab some extra metrics
		if s.AzureDB {
			queries["AzureDBResourceStats"] = Query{ScriptName: "AzureDBPerformanceCounters", Script: sqlAzureDBResourceStats, ResultByRow: false}
			queries["AzureDBResourceGovernance"] = Query{ScriptName: "AzureDBPerformanceCounters", Script: sqlAzureDBResourceGovernance, ResultByRow: false}
		}
		// Decide if we want to run version 1 or version 2 queries
		if s.QueryVersion == 2 {
			log.Println("W! DEPRECATION NOTICE: query_version=2 is being deprecated in favor of database_type.")
			queries["PerformanceCounters"] = Query{ScriptName: "PerformanceCounters", Script: sqlPerformanceCountersV2, ResultByRow: true}
			queries["WaitStatsCategorized"] = Query{ScriptName: "WaitStatsCategorized", Script: sqlWaitStatsCategorizedV2, ResultByRow: false}
			queries["DatabaseIO"] = Query{ScriptName: "DatabaseIO", Script: sqlDatabaseIOV2, ResultByRow: false}
			queries["ServerProperties"] = Query{ScriptName: "ServerProperties", Script: sqlServerPropertiesV2, ResultByRow: false}
			queries["MemoryClerk"] = Query{ScriptName: "MemoryClerk", Script: sqlMemoryClerkV2, ResultByRow: false}
			queries["Schedulers"] = Query{ScriptName: "Schedulers", Script: sqlServerSchedulersV2, ResultByRow: false}
			queries["SqlRequests"] = Query{ScriptName: "SqlRequests", Script: sqlServerRequestsV2, ResultByRow: false}
			queries["VolumeSpace"] = Query{ScriptName: "VolumeSpace", Script: sqlServerVolumeSpaceV2, ResultByRow: false}
			queries["Cpu"] = Query{ScriptName: "Cpu", Script: sqlServerCpuV2, ResultByRow: false}
		} else {
			log.Println("W! DEPRECATED: query_version=1 has been deprecated in favor of database_type.")
			queries["PerformanceCounters"] = Query{ScriptName: "PerformanceCounters", Script: sqlPerformanceCounters, ResultByRow: true}
			queries["WaitStatsCategorized"] = Query{ScriptName: "WaitStatsCategorized", Script: sqlWaitStatsCategorized, ResultByRow: false}
			queries["CPUHistory"] = Query{ScriptName: "CPUHistory", Script: sqlCPUHistory, ResultByRow: false}
			queries["DatabaseIO"] = Query{ScriptName: "DatabaseIO", Script: sqlDatabaseIO, ResultByRow: false}
			queries["DatabaseSize"] = Query{ScriptName: "DatabaseSize", Script: sqlDatabaseSize, ResultByRow: false}
			queries["DatabaseStats"] = Query{ScriptName: "DatabaseStats", Script: sqlDatabaseStats, ResultByRow: false}
			queries["DatabaseProperties"] = Query{ScriptName: "DatabaseProperties", Script: sqlDatabaseProperties, ResultByRow: false}
			queries["MemoryClerk"] = Query{ScriptName: "MemoryClerk", Script: sqlMemoryClerk, ResultByRow: false}
			queries["VolumeSpace"] = Query{ScriptName: "VolumeSpace", Script: sqlVolumeSpace, ResultByRow: false}
			queries["PerformanceMetrics"] = Query{ScriptName: "PerformanceMetrics", Script: sqlPerformanceMetrics, ResultByRow: false}
		}
	}

	filterQueries, err := filter.NewIncludeExcludeFilter(s.IncludeQuery, s.ExcludeQuery)
	if err != nil {
		return err
	}

	for query := range queries {
		if !filterQueries.Match(query) {
			delete(queries, query)
		}
	}

	// Set a flag so we know that queries have already been initialized
	s.isInitialized = true
	var querylist []string
	for query := range queries {
		querylist = append(querylist, query)
	}
	log.Printf("I! [inputs.sqlserver] Config: Effective Queries: %#v\n", querylist)

	return nil
}

// Gather collect data from SQL Server
func (s *SQLServer) Gather(acc telegraf.Accumulator) error {
	if !s.isInitialized {
		if err := initQueries(s); err != nil {
			acc.AddError(err)
			return err
		}
	}

	if len(s.Servers) == 0 {
		s.Servers = append(s.Servers, defaultServer)
	}

	var wg sync.WaitGroup

	for _, serv := range s.Servers {
		for _, query := range s.queries {
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

func (s *SQLServer) gatherServer(server string, query Query, acc telegraf.Accumulator) error {
	// deferred opening
	conn, err := sql.Open("mssql", server)
	if err != nil {
		return err
	}
	defer conn.Close()

	// execute query
	rows, err := conn.Query(query.Script)
	if err != nil {
		return fmt.Errorf("Script %s failed: %w", query.ScriptName, err)
		//return   err
	}
	defer rows.Close()

	// grab the column information from the result
	query.OrderedColumns, err = rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		err = s.accRow(query, acc, rows)
		if err != nil {
			return err
		}
	}
	return rows.Err()
}

func (s *SQLServer) accRow(query Query, acc telegraf.Accumulator, row scanner) error {
	var columnVars []interface{}
	var fields = make(map[string]interface{})

	// store the column name with its *interface{}
	columnMap := make(map[string]*interface{})
	for _, column := range query.OrderedColumns {
		columnMap[column] = new(interface{})
	}
	// populate the array of interface{} with the pointers in the right order
	for i := 0; i < len(columnMap); i++ {
		columnVars = append(columnVars, columnMap[query.OrderedColumns[i]])
	}
	// deconstruct array of variables and send to Scan
	err := row.Scan(columnVars...)
	if err != nil {
		return err
	}

	// measurement: identified by the header
	// tags: all other fields of type string
	tags := map[string]string{}
	var measurement string
	for header, val := range columnMap {
		if str, ok := (*val).(string); ok {
			if header == "measurement" {
				measurement = str
			} else {
				tags[header] = str
			}
		}
	}

	if query.ResultByRow {
		// add measurement to Accumulator
		acc.AddFields(measurement,
			map[string]interface{}{"value": *columnMap["value"]},
			tags, time.Now())
	} else {
		// values
		for header, val := range columnMap {
			if _, ok := (*val).(string); !ok {
				fields[header] = (*val)
			}
		}
		// add fields to Accumulator
		acc.AddFields(measurement, fields, tags, time.Now())
	}
	return nil
}

func init() {
	inputs.Add("sqlserver", func() telegraf.Input {
		return &SQLServer{}
	})
}
