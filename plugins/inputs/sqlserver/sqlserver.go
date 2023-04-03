//go:generate ../../../tools/readme_config_includer/generator
package sqlserver

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Azure/go-autorest/autorest/adal"
	mssql "github.com/denisenkom/go-mssqldb"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// SQLServer struct
type SQLServer struct {
	Servers      []*config.Secret `toml:"servers"`
	QueryTimeout config.Duration  `toml:"query_timeout"`
	AuthMethod   string           `toml:"auth_method"`
	QueryVersion int              `toml:"query_version" deprecated:"1.16.0;use 'database_type' instead"`
	AzureDB      bool             `toml:"azuredb" deprecated:"1.16.0;use 'database_type' instead"`
	DatabaseType string           `toml:"database_type"`
	IncludeQuery []string         `toml:"include_query"`
	ExcludeQuery []string         `toml:"exclude_query"`
	HealthMetric bool             `toml:"health_metric"`
	Log          telegraf.Logger  `toml:"-"`

	pools       []*sql.DB
	queries     MapQuery
	adalToken   *adal.Token
	muCacheLock sync.RWMutex
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
	typeAzureSQLPool            = "AzureSQLPool"
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

type scanner interface {
	Scan(dest ...interface{}) error
}

func (s *SQLServer) initQueries() error {
	s.queries = make(MapQuery)
	queries := s.queries
	s.Log.Infof("Config: database_type: %s , query_version:%d , azuredb: %t", s.DatabaseType, s.QueryVersion, s.AzureDB)

	// To prevent query definition conflicts
	// Constant definitions for type "AzureSQLDB" start with sqlAzureDB
	// Constant definitions for type "AzureSQLManagedInstance" start with sqlAzureMI
	// Constant definitions for type "AzureSQLPool" start with sqlAzurePool
	// Constant definitions for type "SQLServer" start with sqlServer
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
	} else if s.DatabaseType == typeAzureSQLPool {
		queries["AzureSQLPoolResourceStats"] = Query{ScriptName: "AzureSQLPoolResourceStats", Script: sqlAzurePoolResourceStats, ResultByRow: false}
		queries["AzureSQLPoolResourceGovernance"] =
			Query{ScriptName: "AzureSQLPoolResourceGovernance", Script: sqlAzurePoolResourceGovernance, ResultByRow: false}
		queries["AzureSQLPoolDatabaseIO"] = Query{ScriptName: "AzureSQLPoolDatabaseIO", Script: sqlAzurePoolDatabaseIO, ResultByRow: false}
		queries["AzureSQLPoolOsWaitStats"] = Query{ScriptName: "AzureSQLPoolOsWaitStats", Script: sqlAzurePoolOsWaitStats, ResultByRow: false}
		queries["AzureSQLPoolMemoryClerks"] = Query{ScriptName: "AzureSQLPoolMemoryClerks", Script: sqlAzurePoolMemoryClerks, ResultByRow: false}
		queries["AzureSQLPoolPerformanceCounters"] =
			Query{ScriptName: "AzureSQLPoolPerformanceCounters", Script: sqlAzurePoolPerformanceCounters, ResultByRow: false}
		queries["AzureSQLPoolSchedulers"] = Query{ScriptName: "AzureSQLPoolSchedulers", Script: sqlAzurePoolSchedulers, ResultByRow: false}
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
		queries["SQLServerAvailabilityReplicaStates"] =
			Query{ScriptName: "SQLServerAvailabilityReplicaStates", Script: sqlServerAvailabilityReplicaStates, ResultByRow: false}
		queries["SQLServerDatabaseReplicaStates"] =
			Query{ScriptName: "SQLServerDatabaseReplicaStates", Script: sqlServerDatabaseReplicaStates, ResultByRow: false}
		queries["SQLServerRecentBackups"] = Query{ScriptName: "SQLServerRecentBackups", Script: sqlServerRecentBackups, ResultByRow: false}
	} else {
		// If this is an AzureDB instance, grab some extra metrics
		if s.AzureDB {
			queries["AzureDBResourceStats"] = Query{ScriptName: "AzureDBPerformanceCounters", Script: sqlAzureDBResourceStats, ResultByRow: false}
			queries["AzureDBResourceGovernance"] = Query{ScriptName: "AzureDBPerformanceCounters", Script: sqlAzureDBResourceGovernance, ResultByRow: false}
		}
		// Decide if we want to run version 1 or version 2 queries
		if s.QueryVersion == 2 {
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

	queryList := make([]string, 0, len(queries))
	for query := range queries {
		queryList = append(queryList, query)
	}
	s.Log.Infof("Config: Effective Queries: %#v\n", queryList)

	return nil
}

func (*SQLServer) SampleConfig() string {
	return sampleConfig
}

// Gather collect data from SQL Server
func (s *SQLServer) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var healthMetrics = make(map[string]*HealthMetric)

	for i, pool := range s.pools {
		dnsSecret, err := s.Servers[i].Get()
		if err != nil {
			acc.AddError(err)
			continue
		}
		dsn := string(dnsSecret)
		config.ReleaseSecret(dnsSecret)

		for _, query := range s.queries {
			wg.Add(1)
			go func(pool *sql.DB, query Query, dsn string) {
				defer wg.Done()
				queryError := s.gatherServer(pool, query, acc, dsn)

				if s.HealthMetric {
					mutex.Lock()
					s.gatherHealth(healthMetrics, dsn, queryError)
					mutex.Unlock()
				}

				acc.AddError(queryError)
			}(pool, query, dsn)
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
	if err := s.initQueries(); err != nil {
		acc.AddError(err)
		return err
	}

	// initialize mutual exclusion lock
	s.muCacheLock = sync.RWMutex{}

	for _, serv := range s.Servers {
		var pool *sql.DB

		switch strings.ToLower(s.AuthMethod) {
		case "connection_string":
			// Get the connection string potentially containing secrets
			dsn, err := serv.Get()
			if err != nil {
				acc.AddError(err)
				continue
			}

			// Use the DSN (connection string) directly. In this case,
			// empty username/password causes use of Windows
			// integrated authentication.
			pool, err = sql.Open("mssql", string(dsn))
			config.ReleaseSecret(dsn)
			if err != nil {
				acc.AddError(err)
				continue
			}
		case "aad":
			// AAD Auth with system-assigned managed identity (MSI)
			// AAD Auth is only supported for Azure SQL Database or Azure SQL Managed Instance
			if s.DatabaseType == "SQLServer" {
				err := errors.New("database connection failed : AAD auth is not supported for SQL VM i.e. DatabaseType=SQLServer")
				acc.AddError(err)
				continue
			}

			// get token from in-memory cache variable or from Azure Active Directory
			tokenProvider, err := s.getTokenProvider()
			if err != nil {
				acc.AddError(fmt.Errorf("error creating AAD token provider for system assigned Azure managed identity: %w", err))
				continue
			}

			// Get the connection string potentially containing secrets
			dsn, err := serv.Get()
			if err != nil {
				acc.AddError(err)
				continue
			}
			connector, err := mssql.NewAccessTokenConnector(string(dsn), tokenProvider)
			config.ReleaseSecret(dsn)
			if err != nil {
				acc.AddError(fmt.Errorf("error creating the SQL connector: %w", err))
				continue
			}

			pool = sql.OpenDB(connector)
		default:
			return fmt.Errorf("unknown auth method: %v", s.AuthMethod)
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

func (s *SQLServer) gatherServer(pool *sql.DB, query Query, acc telegraf.Accumulator, connectionString string) error {
	// execute query
	ctx := context.Background()
	// Use the query timeout if any
	if s.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(s.QueryTimeout))
		defer cancel()
	}
	rows, err := pool.QueryContext(ctx, query.Script)
	if err != nil {
		serverName, databaseName := getConnectionIdentifiers(connectionString)

		// Error msg based on the format in SSMS. SQLErrorClass() is another term for severity/level: http://msdn.microsoft.com/en-us/library/dd304156.aspx
		var sqlErr mssql.Error
		if errors.As(err, &sqlErr) {
			return fmt.Errorf("query %s failed for server: %s and database: %s with Msg %d, Level %d, State %d:, Line %d, Error: %w", query.ScriptName,
				serverName, databaseName, sqlErr.SQLErrorNumber(), sqlErr.SQLErrorClass(), sqlErr.SQLErrorState(), sqlErr.SQLErrorLineNo(), err)
		}

		return fmt.Errorf("query %s failed for server: %s and database: %s with Error: %w", query.ScriptName, serverName, databaseName, err)
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
	var fields = make(map[string]interface{})

	// store the column name with its *interface{}
	columnMap := make(map[string]*interface{})
	for _, column := range query.OrderedColumns {
		columnMap[column] = new(interface{})
	}

	columnVars := make([]interface{}, 0, len(columnMap))
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
				fields[header] = *val
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
		srv := config.NewSecret([]byte(defaultServer))
		s.Servers = append(s.Servers, &srv)
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
		return &SQLServer{
			AuthMethod: "connection_string",
		}
	})
}
