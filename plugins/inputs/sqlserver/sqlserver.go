package sqlserver

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	"github.com/Azure/go-autorest/autorest/adal"
	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// SQLServer struct
type SQLServer struct {
	Servers      []string `toml:"servers"`
	QueryVersion int      `toml:"query_version"`
	AzureDB      bool     `toml:"azuredb"`
	DatabaseType string   `toml:"database_type"`
	IncludeQuery []string `toml:"include_query"`
	ExcludeQuery []string `toml:"exclude_query"`
	HealthMetric bool     `toml:"health_metric"`
	pools        []*sql.DB
	queries      MapQuery
	adalToken    *adal.Token
	muCacheLock  sync.RWMutex
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

// HealthMetric struct tracking the number of attempted vs successful connections for each connection string
type HealthMetric struct {
	AttemptedQueries  int
	SuccessfulQueries int
}

const defaultServer = "Server=.;app name=telegraf;log=1;"

const (
	typeAzureSQLDB              = "AzureSQLDB"
	typeAzureSQLManagedInstance = "AzureSQLManagedInstance"
	typeSQLServer               = "SQLServer"
)

const (
	healthMetricName              = "sqlserver_telegraf_health"
	healthMetricInstanceTag       = "sql_instance"
	healthMetricDatabaseTag       = "database_name"
	healthMetricAttemptedQueries  = "attempted_queries"
	healthMetricSuccessfulQueries = "successful_queries"
	healthMetricDatabaseType      = "database_type"
)

// resource id for Azure SQL Database
const sqlAzureResourceID = "https://database.windows.net/"

const sampleConfig = `
## Specify instances to monitor with a list of connection strings.
## All connection parameters are optional.
## By default, the host is localhost, listening on default port, TCP 1433.
##   for Windows, the user is the currently running AD user (SSO).
##   See https://github.com/denisenkom/go-mssqldb for detailed connection
##   parameters, in particular, tls connections can be created like so:
##   "encrypt=true;certificate=<cert>;hostNameInCertificate=<SqlServer host fqdn>"
servers = [
  "Server=192.168.1.10;Port=1433;User Id=<user>;Password=<pw>;app name=telegraf;log=1;",
]

## "database_type" enables a specific set of queries depending on the database type. If specified, it replaces azuredb = true/false and query_version = 2
## In the config file, the sql server plugin section should be repeated each with a set of servers for a specific database_type.
## Possible values for database_type are - "AzureSQLDB" or "AzureSQLManagedInstance" or "SQLServer"

## Queries enabled by default for database_type = "AzureSQLDB" are - 
## AzureSQLDBResourceStats, AzureSQLDBResourceGovernance, AzureSQLDBWaitStats, AzureSQLDBDatabaseIO, AzureSQLDBServerProperties, 
## AzureSQLDBOsWaitstats, AzureSQLDBMemoryClerks, AzureSQLDBPerformanceCounters, AzureSQLDBRequests, AzureSQLDBSchedulers

# database_type = "AzureSQLDB"

## A list of queries to include. If not specified, all the above listed queries are used.
# include_query = []

## A list of queries to explicitly ignore.
# exclude_query = []

## Queries enabled by default for database_type = "AzureSQLManagedInstance" are - 
## AzureSQLMIResourceStats, AzureSQLMIResourceGovernance, AzureSQLMIDatabaseIO, AzureSQLMIServerProperties, AzureSQLMIOsWaitstats, 
## AzureSQLMIMemoryClerks, AzureSQLMIPerformanceCounters, AzureSQLMIRequests, AzureSQLMISchedulers

# database_type = "AzureSQLManagedInstance"

# include_query = []

# exclude_query = []

## Queries enabled by default for database_type = "SQLServer" are - 
## SQLServerPerformanceCounters, SQLServerWaitStatsCategorized, SQLServerDatabaseIO, SQLServerProperties, SQLServerMemoryClerks, 
## SQLServerSchedulers, SQLServerRequests, SQLServerVolumeSpace, SQLServerCpu

database_type = "SQLServer"

include_query = []

## SQLServerAvailabilityReplicaStates and SQLServerDatabaseReplicaStates are optional queries and hence excluded here as default
exclude_query = ["SQLServerAvailabilityReplicaStates", "SQLServerDatabaseReplicaStates"]

## Following are old config settings, you may use them only if you are using the earlier flavor of queries, however it is recommended to use 
## the new mechanism of identifying the database_type there by use it's corresponding queries

## Optional parameter, setting this to 2 will use a new version
## of the collection queries that break compatibility with the original
## dashboards.
## Version 2 - is compatible from SQL Server 2012 and later versions and also for SQL Azure DB
# query_version = 2

## If you are using AzureDB, setting this to true will gather resource utilization metrics
# azuredb = false
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
	if s.DatabaseType == typeAzureSQLDB {
		queries["AzureSQLDBResourceStats"] = Query{ScriptName: "AzureSQLDBResourceStats", Script: sqlAzureDBResourceStats, ResultByRow: false}
		queries["AzureSQLDBResourceGovernance"] = Query{ScriptName: "AzureSQLDBResourceGovernance", Script: sqlAzureDBResourceGovernance, ResultByRow: false}
		queries["AzureSQLDBWaitStats"] = Query{ScriptName: "AzureSQLDBWaitStats", Script: sqlAzureDBWaitStats, ResultByRow: false}
		queries["AzureSQLDBDatabaseIO"] = Query{ScriptName: "AzureSQLDBDatabaseIO", Script: sqlAzureDBDatabaseIO, ResultByRow: false}
		queries["AzureSQLDBServerProperties"] = Query{ScriptName: "AzureSQLDBServerProperties", Script: sqlAzureDBProperties, ResultByRow: false}
		queries["AzureSQLDBOsWaitstats"] = Query{ScriptName: "AzureSQLOsWaitstats", Script: sqlAzureDBOsWaitStats, ResultByRow: false}
		queries["AzureSQLDBMemoryClerks"] = Query{ScriptName: "AzureSQLDBMemoryClerks", Script: sqlAzureDBMemoryClerks, ResultByRow: false}
		queries["AzureSQLDBPerformanceCounters"] = Query{ScriptName: "AzureSQLDBPerformanceCounters", Script: sqlAzureDBPerformanceCounters, ResultByRow: false}
		queries["AzureSQLDBRequests"] = Query{ScriptName: "AzureSQLDBRequests", Script: sqlAzureDBRequests, ResultByRow: false}
		queries["AzureSQLDBSchedulers"] = Query{ScriptName: "AzureSQLDBSchedulers", Script: sqlAzureDBSchedulers, ResultByRow: false}
	} else if s.DatabaseType == typeAzureSQLManagedInstance {
		queries["AzureSQLMIResourceStats"] = Query{ScriptName: "AzureSQLMIResourceStats", Script: sqlAzureMIResourceStats, ResultByRow: false}
		queries["AzureSQLMIResourceGovernance"] = Query{ScriptName: "AzureSQLMIResourceGovernance", Script: sqlAzureMIResourceGovernance, ResultByRow: false}
		queries["AzureSQLMIDatabaseIO"] = Query{ScriptName: "AzureSQLMIDatabaseIO", Script: sqlAzureMIDatabaseIO, ResultByRow: false}
		queries["AzureSQLMIServerProperties"] = Query{ScriptName: "AzureSQLMIServerProperties", Script: sqlAzureMIProperties, ResultByRow: false}
		queries["AzureSQLMIOsWaitstats"] = Query{ScriptName: "AzureSQLMIOsWaitstats", Script: sqlAzureMIOsWaitStats, ResultByRow: false}
		queries["AzureSQLMIMemoryClerks"] = Query{ScriptName: "AzureSQLMIMemoryClerks", Script: sqlAzureMIMemoryClerks, ResultByRow: false}
		queries["AzureSQLMIPerformanceCounters"] = Query{ScriptName: "AzureSQLMIPerformanceCounters", Script: sqlAzureMIPerformanceCounters, ResultByRow: false}
		queries["AzureSQLMIRequests"] = Query{ScriptName: "AzureSQLMIRequests", Script: sqlAzureMIRequests, ResultByRow: false}
		queries["AzureSQLMISchedulers"] = Query{ScriptName: "AzureSQLMISchedulers", Script: sqlAzureMISchedulers, ResultByRow: false}
	} else if s.DatabaseType == typeSQLServer { //These are still V2 queries and have not been refactored yet.
		queries["SQLServerPerformanceCounters"] = Query{ScriptName: "SQLServerPerformanceCounters", Script: sqlServerPerformanceCounters, ResultByRow: false}
		queries["SQLServerWaitStatsCategorized"] = Query{ScriptName: "SQLServerWaitStatsCategorized", Script: sqlServerWaitStatsCategorized, ResultByRow: false}
		queries["SQLServerDatabaseIO"] = Query{ScriptName: "SQLServerDatabaseIO", Script: sqlServerDatabaseIO, ResultByRow: false}
		queries["SQLServerProperties"] = Query{ScriptName: "SQLServerProperties", Script: sqlServerProperties, ResultByRow: false}
		queries["SQLServerMemoryClerks"] = Query{ScriptName: "SQLServerMemoryClerks", Script: sqlServerMemoryClerks, ResultByRow: false}
		queries["SQLServerSchedulers"] = Query{ScriptName: "SQLServerSchedulers", Script: sqlServerSchedulers, ResultByRow: false}
		queries["SQLServerRequests"] = Query{ScriptName: "SQLServerRequests", Script: sqlServerRequests, ResultByRow: false}
		queries["SQLServerVolumeSpace"] = Query{ScriptName: "SQLServerVolumeSpace", Script: sqlServerVolumeSpace, ResultByRow: false}
		queries["SQLServerCpu"] = Query{ScriptName: "SQLServerCpu", Script: sqlServerRingBufferCPU, ResultByRow: false}
		queries["SQLServerAvailabilityReplicaStates"] = Query{ScriptName: "SQLServerAvailabilityReplicaStates", Script: sqlServerAvailabilityReplicaStates, ResultByRow: false}
		queries["SQLServerDatabaseReplicaStates"] = Query{ScriptName: "SQLServerDatabaseReplicaStates", Script: sqlServerDatabaseReplicaStates, ResultByRow: false}
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
			queries["Cpu"] = Query{ScriptName: "Cpu", Script: sqlServerCPUV2, ResultByRow: false}
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

	var querylist []string
	for query := range queries {
		querylist = append(querylist, query)
	}
	log.Printf("I! [inputs.sqlserver] Config: Effective Queries: %#v\n", querylist)

	return nil
}

// Gather collect data from SQL Server
func (s *SQLServer) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var healthMetrics = make(map[string]*HealthMetric)

	for i, pool := range s.pools {
		for _, query := range s.queries {
			wg.Add(1)
			go func(pool *sql.DB, query Query, serverIndex int) {
				defer wg.Done()
				queryError := s.gatherServer(pool, query, acc)

				if s.HealthMetric {
					mutex.Lock()
					s.gatherHealth(healthMetrics, s.Servers[serverIndex], queryError)
					mutex.Unlock()
				}

				acc.AddError(queryError)
			}(pool, query, i)
		}
	}

	wg.Wait()

	if s.HealthMetric {
		s.accHealth(healthMetrics, acc)
	}

	return nil
}

// Start initialize a list of connection pools
func (s *SQLServer) Start(acc telegraf.Accumulator) error {
	if err := initQueries(s); err != nil {
		acc.AddError(err)
		return err
	}

	// initialize mutual exclusion lock
	s.muCacheLock = sync.RWMutex{}

	for _, serv := range s.Servers {
		var pool *sql.DB

		// setup connection based on authentication
		rx := regexp.MustCompile(`\b(?:(Password=((?:&(?:[a-z]+|#[0-9]+);|[^;]){0,})))\b`)

		// when password is provided in connection string, use SQL auth
		if rx.MatchString(serv) {
			var err error
			pool, err = sql.Open("mssql", serv)

			if err != nil {
				acc.AddError(err)
				continue
			}
		} else {
			// otherwise assume AAD Auth with system-assigned managed identity (MSI)

			// AAD Auth is only supported for Azure SQL Database or Azure SQL Managed Instance
			if s.DatabaseType == "SQLServer" {
				err := errors.New("database connection failed : AAD auth is not supported for SQL VM i.e. DatabaseType=SQLServer")
				acc.AddError(err)
				continue
			}

			// get token from in-memory cache variable or from Azure Active Directory
			tokenProvider, err := s.getTokenProvider()
			if err != nil {
				acc.AddError(fmt.Errorf("error creating AAD token provider for system assigned Azure managed identity : %s", err.Error()))
				continue
			}

			connector, err := mssql.NewAccessTokenConnector(serv, tokenProvider)
			if err != nil {
				acc.AddError(fmt.Errorf("error creating the SQL connector : %s", err.Error()))
				continue
			}

			pool = sql.OpenDB(connector)
		}

		s.pools = append(s.pools, pool)
	}

	return nil
}

// Stop cleanup server connection pools
func (s *SQLServer) Stop() {
	for _, pool := range s.pools {
		_ = pool.Close()
	}
}

func (s *SQLServer) gatherServer(pool *sql.DB, query Query, acc telegraf.Accumulator) error {
	// execute query
	rows, err := pool.Query(query.Script)
	if err != nil {
		return fmt.Errorf("script %s failed: %w", query.ScriptName, err)
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

	if s.DatabaseType != "" {
		tags["measurement_db_type"] = s.DatabaseType
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

// gatherHealth stores info about any query errors in the healthMetrics map
func (s *SQLServer) gatherHealth(healthMetrics map[string]*HealthMetric, serv string, queryError error) {
	if healthMetrics[serv] == nil {
		healthMetrics[serv] = &HealthMetric{}
	}

	healthMetrics[serv].AttemptedQueries++
	if queryError == nil {
		healthMetrics[serv].SuccessfulQueries++
	}
}

// accHealth accumulates the query health data contained within the healthMetrics map
func (s *SQLServer) accHealth(healthMetrics map[string]*HealthMetric, acc telegraf.Accumulator) {
	for connectionString, connectionStats := range healthMetrics {
		sqlInstance, databaseName := getConnectionIdentifiers(connectionString)
		tags := map[string]string{healthMetricInstanceTag: sqlInstance, healthMetricDatabaseTag: databaseName}
		fields := map[string]interface{}{
			healthMetricAttemptedQueries:  connectionStats.AttemptedQueries,
			healthMetricSuccessfulQueries: connectionStats.SuccessfulQueries,
			healthMetricDatabaseType:      s.getDatabaseTypeToLog(),
		}

		acc.AddFields(healthMetricName, fields, tags, time.Now())
	}
}

// getDatabaseTypeToLog returns the type of database monitored by this plugin instance
func (s *SQLServer) getDatabaseTypeToLog() string {
	if s.DatabaseType == typeAzureSQLDB || s.DatabaseType == typeAzureSQLManagedInstance || s.DatabaseType == typeSQLServer {
		return s.DatabaseType
	}

	logname := fmt.Sprintf("QueryVersion-%d", s.QueryVersion)
	if s.AzureDB {
		logname += "-AzureDB"
	}
	return logname
}

func (s *SQLServer) Init() error {
	if len(s.Servers) == 0 {
		log.Println("W! Warning: Server list is empty.")
	}

	return nil
}

// Get Token Provider by loading cached token or refreshed token
func (s *SQLServer) getTokenProvider() (func() (string, error), error) {
	var tokenString string

	// load token
	s.muCacheLock.RLock()
	token, err := s.loadToken()
	s.muCacheLock.RUnlock()

	// if there's error while loading token or found an expired token, refresh token and save it
	if err != nil || token.IsExpired() {
		// refresh token within a write-lock
		s.muCacheLock.Lock()
		defer s.muCacheLock.Unlock()

		// load token again, in case it's been refreshed by another thread
		token, err = s.loadToken()

		// check loaded token's error/validity, then refresh/save token
		if err != nil || token.IsExpired() {
			// get new token
			spt, err := s.refreshToken()
			if err != nil {
				return nil, err
			}

			// use the refreshed token
			tokenString = spt.OAuthToken()
		} else {
			// use locally cached token
			tokenString = token.OAuthToken()
		}
	} else {
		// use locally cached token
		tokenString = token.OAuthToken()
	}

	// return acquired token
	return func() (string, error) {
		return tokenString, nil
	}, nil
}

// Load token from in-mem cache
func (s *SQLServer) loadToken() (*adal.Token, error) {
	// This method currently does a simplistic task of reading a from variable (in-mem cache),
	// however it's been structured here to allow extending the cache mechanism to a different approach in future

	if s.adalToken == nil {
		return nil, fmt.Errorf("token is nil or failed to load existing token")
	}

	return s.adalToken, nil
}

// Refresh token for the resource, and save to in-mem cache
func (s *SQLServer) refreshToken() (*adal.Token, error) {
	// get MSI endpoint to get a token
	msiEndpoint, err := adal.GetMSIVMEndpoint()
	if err != nil {
		return nil, err
	}

	// get new token for the resource id
	spt, err := adal.NewServicePrincipalTokenFromMSI(msiEndpoint, sqlAzureResourceID)
	if err != nil {
		return nil, err
	}

	// ensure token is fresh
	if err := spt.EnsureFresh(); err != nil {
		return nil, err
	}

	// save token to local in-mem cache
	s.adalToken = &adal.Token{
		AccessToken:  spt.Token().AccessToken,
		RefreshToken: spt.Token().RefreshToken,
		ExpiresIn:    spt.Token().ExpiresIn,
		ExpiresOn:    spt.Token().ExpiresOn,
		NotBefore:    spt.Token().NotBefore,
		Resource:     spt.Token().Resource,
		Type:         spt.Token().Type,
	}

	return s.adalToken, nil
}

func init() {
	inputs.Add("sqlserver", func() telegraf.Input {
		return &SQLServer{Servers: []string{defaultServer}}
	})
}
