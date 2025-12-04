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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/go-autorest/autorest/adal" // legacy ADAL package for backward compatibility
	mssql "github.com/microsoft/go-mssqldb"
	_ "github.com/microsoft/go-mssqldb/namedpipe"    // required to support NP protocol
	_ "github.com/microsoft/go-mssqldb/sharedmemory" // required to support LPC protocol

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultServer = "Server=.;app name=telegraf;log=1;"

	typeAzureSQLDB                 = "AzureSQLDB"
	typeAzureSQLManagedInstance    = "AzureSQLManagedInstance"
	typeAzureSQLPool               = "AzureSQLPool"
	typeSQLServer                  = "SQLServer"
	typeAzureArcSQLManagedInstance = "AzureArcSQLManagedInstance"

	healthMetricName              = "sqlserver_telegraf_health"
	healthMetricInstanceTag       = "sql_instance"
	healthMetricDatabaseTag       = "database_name"
	healthMetricAttemptedQueries  = "attempted_queries"
	healthMetricSuccessfulQueries = "successful_queries"
	healthMetricDatabaseType      = "database_type"

	sqlAzureResourceID = "https://database.windows.net/"
)

type SQLServer struct {
	Servers            []*config.Secret `toml:"servers"`
	QueryTimeout       config.Duration  `toml:"query_timeout"`
	AuthMethod         string           `toml:"auth_method"`
	ClientID           string           `toml:"client_id"`
	DatabaseType       string           `toml:"database_type"`
	IncludeQuery       []string         `toml:"include_query"`
	ExcludeQuery       []string         `toml:"exclude_query"`
	HealthMetric       bool             `toml:"health_metric"`
	MaxOpenConnections int              `toml:"max_open_connections"`
	MaxIdleConnections int              `toml:"max_idle_connections"`
	Log                telegraf.Logger  `toml:"-"`

	pools   []*sql.DB
	queries mapQuery

	// Legacy token - kept for backward compatibility
	adalToken *adal.Token
	// New token using Azure Identity SDK
	azToken *azureToken
	// Config option to use legacy ADAL authentication instead of the newer Azure Identity SDK
	// When true, the deprecated ADAL library will be used
	// When false (default), the new Azure Identity SDK will be used
	UseAdalToken bool `toml:"use_deprecated_adal_authentication" deprecated:"1.40.0;migrate to MSAL authentication"`

	muCacheLock sync.RWMutex
}

type query struct {
	ScriptName     string
	Script         string
	ResultByRow    bool
	OrderedColumns []string
}

type mapQuery map[string]query

// healthMetric struct tracking the number of attempted vs. successful connections for each connection string
type healthMetric struct {
	attemptedQueries  int
	successfulQueries int
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func (*SQLServer) SampleConfig() string {
	return sampleConfig
}

func (s *SQLServer) Init() error {
	if len(s.Servers) == 0 {
		srv := config.NewSecret([]byte(defaultServer))
		s.Servers = append(s.Servers, &srv)
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
			pool, err = sql.Open("mssql", dsn.String())
			dsn.Destroy()
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
			connector, err := mssql.NewAccessTokenConnector(dsn.String(), tokenProvider)
			dsn.Destroy()
			if err != nil {
				acc.AddError(fmt.Errorf("error creating the SQL connector: %w", err))
				continue
			}

			pool = sql.OpenDB(connector)
		default:
			return fmt.Errorf("unknown auth method: %v", s.AuthMethod)
		}

		// Use max_open_connections if any
		if s.MaxOpenConnections > 0 {
			pool.SetMaxOpenConns(s.MaxOpenConnections)
		}

		// Use max_idle_connections if any
		if s.MaxIdleConnections > 0 {
			pool.SetMaxIdleConns(s.MaxIdleConnections)
		}

		s.pools = append(s.pools, pool)
	}

	return nil
}

func (s *SQLServer) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var healthMetrics = make(map[string]*healthMetric)

	for i, pool := range s.pools {
		dnsSecret, err := s.Servers[i].Get()
		if err != nil {
			acc.AddError(err)
			continue
		}
		dsn := dnsSecret.String()
		dnsSecret.Destroy()

		for _, q := range s.queries {
			wg.Add(1)
			go func(pool *sql.DB, q query, dsn string) {
				defer wg.Done()
				queryError := s.gatherServer(pool, q, acc, dsn)

				if s.HealthMetric {
					mutex.Lock()
					gatherHealth(healthMetrics, dsn, queryError)
					mutex.Unlock()
				}

				acc.AddError(queryError)
			}(pool, q, dsn)
		}
	}

	wg.Wait()

	if s.HealthMetric {
		s.accHealth(healthMetrics, acc)
	}

	return nil
}

// Stop cleanup server connection pools
func (s *SQLServer) Stop() {
	for _, pool := range s.pools {
		_ = pool.Close()
	}
}

func (s *SQLServer) initQueries() error {
	s.queries = make(mapQuery)
	queries := s.queries
	s.Log.Infof("Config: database_type: %s", s.DatabaseType)

	// If database_type is not set, default to SQLServer for backward compatibility
	if s.DatabaseType == "" {
		s.DatabaseType = typeSQLServer
		s.Log.Warnf("database_type not specified, defaulting to %s", typeSQLServer)
	}

	// To prevent query definition conflicts
	// Constant definitions for type "AzureSQLDB" start with sqlAzureDB
	// Constant definitions for type "AzureSQLManagedInstance" start with sqlAzureMI
	// Constant definitions for type "AzureSQLPool" start with sqlAzurePool
	// Constant definitions for type "AzureArcSQLManagedInstance" start with sqlAzureArcMI
	// Constant definitions for type "SQLServer" start with sqlServer
	if s.DatabaseType == typeAzureSQLDB {
		queries["AzureSQLDBResourceStats"] = query{ScriptName: "AzureSQLDBResourceStats", Script: sqlAzureDBResourceStats, ResultByRow: false}
		queries["AzureSQLDBResourceGovernance"] = query{ScriptName: "AzureSQLDBResourceGovernance", Script: sqlAzureDBResourceGovernance, ResultByRow: false}
		queries["AzureSQLDBWaitStats"] = query{ScriptName: "AzureSQLDBWaitStats", Script: sqlAzureDBWaitStats, ResultByRow: false}
		queries["AzureSQLDBDatabaseIO"] = query{ScriptName: "AzureSQLDBDatabaseIO", Script: sqlAzureDBDatabaseIO, ResultByRow: false}
		queries["AzureSQLDBServerProperties"] = query{ScriptName: "AzureSQLDBServerProperties", Script: sqlAzureDBProperties, ResultByRow: false}
		queries["AzureSQLDBOsWaitstats"] = query{ScriptName: "AzureSQLOsWaitstats", Script: sqlAzureDBOsWaitStats, ResultByRow: false}
		queries["AzureSQLDBMemoryClerks"] = query{ScriptName: "AzureSQLDBMemoryClerks", Script: sqlAzureDBMemoryClerks, ResultByRow: false}
		queries["AzureSQLDBPerformanceCounters"] = query{ScriptName: "AzureSQLDBPerformanceCounters", Script: sqlAzureDBPerformanceCounters, ResultByRow: false}
		queries["AzureSQLDBRequests"] = query{ScriptName: "AzureSQLDBRequests", Script: sqlAzureDBRequests, ResultByRow: false}
		queries["AzureSQLDBSchedulers"] = query{ScriptName: "AzureSQLDBSchedulers", Script: sqlAzureDBSchedulers, ResultByRow: false}
	} else if s.DatabaseType == typeAzureSQLManagedInstance {
		queries["AzureSQLMIResourceStats"] = query{ScriptName: "AzureSQLMIResourceStats", Script: sqlAzureMIResourceStats, ResultByRow: false}
		queries["AzureSQLMIResourceGovernance"] = query{ScriptName: "AzureSQLMIResourceGovernance", Script: sqlAzureMIResourceGovernance, ResultByRow: false}
		queries["AzureSQLMIDatabaseIO"] = query{ScriptName: "AzureSQLMIDatabaseIO", Script: sqlAzureMIDatabaseIO, ResultByRow: false}
		queries["AzureSQLMIServerProperties"] = query{ScriptName: "AzureSQLMIServerProperties", Script: sqlAzureMIProperties, ResultByRow: false}
		queries["AzureSQLMIOsWaitstats"] = query{ScriptName: "AzureSQLMIOsWaitstats", Script: sqlAzureMIOsWaitStats, ResultByRow: false}
		queries["AzureSQLMIMemoryClerks"] = query{ScriptName: "AzureSQLMIMemoryClerks", Script: sqlAzureMIMemoryClerks, ResultByRow: false}
		queries["AzureSQLMIPerformanceCounters"] = query{ScriptName: "AzureSQLMIPerformanceCounters", Script: sqlAzureMIPerformanceCounters, ResultByRow: false}
		queries["AzureSQLMIRequests"] = query{ScriptName: "AzureSQLMIRequests", Script: sqlAzureMIRequests, ResultByRow: false}
		queries["AzureSQLMISchedulers"] = query{ScriptName: "AzureSQLMISchedulers", Script: sqlAzureMISchedulers, ResultByRow: false}
	} else if s.DatabaseType == typeAzureSQLPool {
		queries["AzureSQLPoolResourceStats"] = query{ScriptName: "AzureSQLPoolResourceStats", Script: sqlAzurePoolResourceStats, ResultByRow: false}
		queries["AzureSQLPoolResourceGovernance"] =
			query{ScriptName: "AzureSQLPoolResourceGovernance", Script: sqlAzurePoolResourceGovernance, ResultByRow: false}
		queries["AzureSQLPoolDatabaseIO"] = query{ScriptName: "AzureSQLPoolDatabaseIO", Script: sqlAzurePoolDatabaseIO, ResultByRow: false}
		queries["AzureSQLPoolOsWaitStats"] = query{ScriptName: "AzureSQLPoolOsWaitStats", Script: sqlAzurePoolOsWaitStats, ResultByRow: false}
		queries["AzureSQLPoolMemoryClerks"] = query{ScriptName: "AzureSQLPoolMemoryClerks", Script: sqlAzurePoolMemoryClerks, ResultByRow: false}
		queries["AzureSQLPoolPerformanceCounters"] =
			query{ScriptName: "AzureSQLPoolPerformanceCounters", Script: sqlAzurePoolPerformanceCounters, ResultByRow: false}
		queries["AzureSQLPoolSchedulers"] = query{ScriptName: "AzureSQLPoolSchedulers", Script: sqlAzurePoolSchedulers, ResultByRow: false}
	} else if s.DatabaseType == typeAzureArcSQLManagedInstance {
		queries["AzureArcSQLMIDatabaseIO"] = query{ScriptName: "AzureArcSQLMIDatabaseIO", Script: sqlAzureArcMIDatabaseIO, ResultByRow: false}
		queries["AzureArcSQLMIServerProperties"] = query{ScriptName: "AzureArcSQLMIServerProperties", Script: sqlAzureArcMIProperties, ResultByRow: false}
		queries["AzureArcSQLMIOsWaitstats"] = query{ScriptName: "AzureArcSQLMIOsWaitstats", Script: sqlAzureArcMIOsWaitStats, ResultByRow: false}
		queries["AzureArcSQLMIMemoryClerks"] = query{ScriptName: "AzureArcSQLMIMemoryClerks", Script: sqlAzureArcMIMemoryClerks, ResultByRow: false}
		queries["AzureArcSQLMIPerformanceCounters"] =
			query{ScriptName: "AzureArcSQLMIPerformanceCounters", Script: sqlAzureArcMIPerformanceCounters, ResultByRow: false}
		queries["AzureArcSQLMIRequests"] = query{ScriptName: "AzureArcSQLMIRequests", Script: sqlAzureArcMIRequests, ResultByRow: false}
		queries["AzureArcSQLMISchedulers"] = query{ScriptName: "AzureArcSQLMISchedulers", Script: sqlAzureArcMISchedulers, ResultByRow: false}
	} else if s.DatabaseType == typeSQLServer { // These are still V2 queries and have not been refactored yet.
		queries["SQLServerPerformanceCounters"] = query{ScriptName: "SQLServerPerformanceCounters", Script: sqlServerPerformanceCounters, ResultByRow: false}
		queries["SQLServerWaitStatsCategorized"] = query{ScriptName: "SQLServerWaitStatsCategorized", Script: sqlServerWaitStatsCategorized, ResultByRow: false}
		queries["SQLServerDatabaseIO"] = query{ScriptName: "SQLServerDatabaseIO", Script: sqlServerDatabaseIO, ResultByRow: false}
		queries["SQLServerProperties"] = query{ScriptName: "SQLServerProperties", Script: sqlServerProperties, ResultByRow: false}
		queries["SQLServerMemoryClerks"] = query{ScriptName: "SQLServerMemoryClerks", Script: sqlServerMemoryClerks, ResultByRow: false}
		queries["SQLServerSchedulers"] = query{ScriptName: "SQLServerSchedulers", Script: sqlServerSchedulers, ResultByRow: false}
		queries["SQLServerRequests"] = query{ScriptName: "SQLServerRequests", Script: sqlServerRequests, ResultByRow: false}
		queries["SQLServerVolumeSpace"] = query{ScriptName: "SQLServerVolumeSpace", Script: sqlServerVolumeSpace, ResultByRow: false}
		queries["SQLServerCpu"] = query{ScriptName: "SQLServerCpu", Script: sqlServerRingBufferCPU, ResultByRow: false}
		queries["SQLServerAvailabilityReplicaStates"] =
			query{ScriptName: "SQLServerAvailabilityReplicaStates", Script: sqlServerAvailabilityReplicaStates, ResultByRow: false}
		queries["SQLServerDatabaseReplicaStates"] =
			query{ScriptName: "SQLServerDatabaseReplicaStates", Script: sqlServerDatabaseReplicaStates, ResultByRow: false}
		queries["SQLServerRecentBackups"] = query{ScriptName: "SQLServerRecentBackups", Script: sqlServerRecentBackups, ResultByRow: false}
		queries["SQLServerPersistentVersionStore"] =
			query{ScriptName: "SQLServerPersistentVersionStore", Script: sqlServerPersistentVersionStore, ResultByRow: false}
	} else {
		return fmt.Errorf("unsupported database_type: %s. Supported types are: %s, %s, %s, %s, %s",
			s.DatabaseType, typeAzureSQLDB, typeAzureSQLManagedInstance, typeAzureSQLPool, typeAzureArcSQLManagedInstance, typeSQLServer)
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

func (s *SQLServer) gatherServer(pool *sql.DB, query query, acc telegraf.Accumulator, connectionString string) error {
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

func (s *SQLServer) accRow(query query, acc telegraf.Accumulator, row scanner) error {
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
	tags := make(map[string]string, len(columnMap)+1)
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
func gatherHealth(healthMetrics map[string]*healthMetric, serv string, queryError error) {
	if healthMetrics[serv] == nil {
		healthMetrics[serv] = &healthMetric{}
	}

	healthMetrics[serv].attemptedQueries++
	if queryError == nil {
		healthMetrics[serv].successfulQueries++
	}
}

// accHealth accumulates the query health data contained within the healthMetrics map
func (s *SQLServer) accHealth(healthMetrics map[string]*healthMetric, acc telegraf.Accumulator) {
	for connectionString, connectionStats := range healthMetrics {
		sqlInstance, databaseName := getConnectionIdentifiers(connectionString)
		tags := map[string]string{healthMetricInstanceTag: sqlInstance, healthMetricDatabaseTag: databaseName}
		fields := map[string]interface{}{
			healthMetricAttemptedQueries:  connectionStats.attemptedQueries,
			healthMetricSuccessfulQueries: connectionStats.successfulQueries,
			healthMetricDatabaseType:      s.DatabaseType,
		}

		acc.AddFields(healthMetricName, fields, tags, time.Now())
	}
}

// ------------------------------------------------------------------------------
// Token Provider Implementation
// ------------------------------------------------------------------------------

// getTokenProvider returns a function that provides authentication tokens for SQL Server.
//
// DEPRECATION NOTICE:
// The ADAL authentication library is deprecated and will be removed in a future version.
// It is strongly recommended to migrate to the Azure Identity SDK.
// See the migration documentation at: https://learn.microsoft.com/en-us/azure/active-directory/develop/msal-migration
//
// This implementation supports both authentication methods:
// 1. Azure Identity SDK (default, recommended)
// 2. Legacy ADAL library (deprecated, maintained for backward compatibility)
//
// To control which authentication library is used, set the use_deprecated_adal_authentication config option:
// - use_deprecated_adal_authentication = true  : Use legacy ADAL authentication (deprecated)
// - use_deprecated_adal_authentication = false : Use Azure Identity SDK (recommended)
// - Not set                : Use Azure Identity SDK (recommended)
func (s *SQLServer) getTokenProvider() (func() (string, error), error) {
	// Check if use_deprecated_adal_authentication config option is set to determine which auth method to use
	// Default to using Azure Identity SDK if the config is not set
	useAzureIdentity := !s.UseAdalToken
	if useAzureIdentity {
		s.Log.Debugf("Using Azure Identity SDK for authentication (recommended)")
	} else {
		s.Log.Debugf("Using legacy ADAL for authentication (deprecated, will be removed in 1.40.0)")
	}

	var tokenString string

	if useAzureIdentity {
		// Use Azure Identity SDK
		s.muCacheLock.RLock()
		token, err := s.loadAzureToken()
		s.muCacheLock.RUnlock()

		// If the token is nil, expired, or there was an error loading it, refresh the token
		if err != nil || token == nil || token.IsExpired() {
			// Refresh token within a write-lock
			s.muCacheLock.Lock()
			defer s.muCacheLock.Unlock()

			// Load token again, in case it's been refreshed by another thread
			token, err = s.loadAzureToken()

			// Check loaded token's error/validity, then refresh/save token
			if err != nil || token == nil || token.IsExpired() {
				// Get new token
				newToken, err := s.refreshAzureToken()
				if err != nil {
					return nil, err
				}

				// Use the refreshed token
				tokenString = newToken.token
			} else {
				// Use locally cached token
				tokenString = token.token
			}
		} else {
			// Use locally cached token
			tokenString = token.token
		}
	} else {
		// Use legacy ADAL approach for backward compatibility
		s.muCacheLock.RLock()
		token, err := s.loadToken()
		s.muCacheLock.RUnlock()

		// If there's an error while loading token or found an expired token, refresh token and save it
		if err != nil || token.IsExpired() {
			// Refresh token within a write-lock
			s.muCacheLock.Lock()
			defer s.muCacheLock.Unlock()

			// Load token again, in case it's been refreshed by another thread
			token, err = s.loadToken()

			// Check loaded token's error/validity, then refresh/save token
			if err != nil || token.IsExpired() {
				// Get new token
				spt, err := s.refreshToken()
				if err != nil {
					return nil, err
				}

				// Use the refreshed token
				tokenString = spt.OAuthToken()
			} else {
				// Use locally cached token
				tokenString = token.OAuthToken()
			}
		} else {
			// Use locally cached token
			tokenString = token.OAuthToken()
		}
	}

	// Return acquired token
	//nolint:unparam // token provider function always returns nil error in this scenario
	return func() (string, error) {
		return tokenString, nil
	}, nil
}

// ------------------------------------------------------------------------------
// Legacy ADAL Token Methods - Kept for backward compatibility
// ------------------------------------------------------------------------------

// loadToken loads a token from in-memory cache using the legacy ADAL method.
//
// Deprecated: This method uses the deprecated ADAL library and will be removed in a future version.
// Use the Azure Identity SDK instead of setting use_deprecated_adal_authentication = false or omitting it.
// See migration documentation: https://learn.microsoft.com/en-us/azure/active-directory/develop/msal-migration
func (s *SQLServer) loadToken() (*adal.Token, error) {
	// This method currently does a simplistic task of reading from a variable (in-mem cache);
	// however, it's been structured here to allow extending the cache mechanism to a different approach in future

	if s.adalToken == nil {
		return nil, errors.New("token is nil or failed to load existing token")
	}

	return s.adalToken, nil
}

// refreshToken refreshes the token using the legacy ADAL method.
//
// Deprecated: This method uses the deprecated ADAL library and will be removed in a future version.
// Use the Azure Identity SDK instead of setting use_deprecated_adal_authentication = false or omitting it.
// See migration documentation: https://learn.microsoft.com/en-us/azure/active-directory/develop/msal-migration
func (s *SQLServer) refreshToken() (*adal.Token, error) {
	// get MSI endpoint to get a token
	msiEndpoint, err := adal.GetMSIVMEndpoint()
	if err != nil {
		return nil, fmt.Errorf("failed to get MSI endpoint: %w", err)
	}

	// get a new token for the resource id
	var spt *adal.ServicePrincipalToken
	if s.ClientID == "" {
		// Using system-assigned managed identity
		s.Log.Debugf("Using system-assigned managed identity with ADAL")
		spt, err = adal.NewServicePrincipalTokenFromMSI(msiEndpoint, sqlAzureResourceID)
		if err != nil {
			return nil, fmt.Errorf("failed to create service principal token from MSI: %w", err)
		}
	} else {
		// Using user-assigned managed identity
		s.Log.Debugf("Using user-assigned managed identity with ClientID: %s with ADAL", s.ClientID)
		spt, err = adal.NewServicePrincipalTokenFromMSIWithUserAssignedID(msiEndpoint, sqlAzureResourceID, s.ClientID)
		if err != nil {
			return nil, fmt.Errorf("failed to create service principal token from MSI with user-assigned ID: %w", err)
		}
	}

	// ensure the token is fresh
	if err := spt.EnsureFresh(); err != nil {
		return nil, fmt.Errorf("failed to ensure token freshness: %w", err)
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

// ------------------------------------------------------------------------------
// New Azure Identity SDK Token Methods
// ------------------------------------------------------------------------------

// loadAzureToken loads a token from in-memory cache using the Azure Identity SDK.
//
// This is the recommended authentication method for Azure SQL resources.
func (s *SQLServer) loadAzureToken() (*azureToken, error) {
	// This method reads from variable (in-mem cache) but can be extended
	// for different cache mechanisms in the future

	if s.azToken == nil {
		return nil, errors.New("token is nil or failed to load existing token")
	}

	return s.azToken, nil
}

// refreshAzureToken refreshes the token using the Azure Identity SDK.
//
// This is the recommended authentication method for Azure SQL resources.
func (s *SQLServer) refreshAzureToken() (*azureToken, error) {
	var options *azidentity.ManagedIdentityCredentialOptions

	if s.ClientID != "" {
		options = &azidentity.ManagedIdentityCredentialOptions{
			ID: azidentity.ResourceID(s.ClientID),
		}
	}
	cred, err := azidentity.NewManagedIdentityCredential(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create managed identity credential: %w", err)
	}

	// Get token from Azure AD
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	accessToken, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{sqlAzureResourceID + "/.default"},
	})
	if err != nil {
		credType := "system-assigned"
		if s.ClientID != "" {
			credType = fmt.Sprintf("user-assigned (ClientID: %s)", s.ClientID)
		}
		return nil, fmt.Errorf("failed to get token using %s managed identity: %w", credType, err)
	}

	// Save token to cache
	s.azToken = &azureToken{
		token:     accessToken.Token,
		expiresOn: accessToken.ExpiresOn,
	}

	return s.azToken, nil
}

func init() {
	inputs.Add("sqlserver", func() telegraf.Input {
		return &SQLServer{
			AuthMethod: "connection_string",
		}
	})
}
