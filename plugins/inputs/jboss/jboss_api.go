package jboss

// HostResponse expected GetHost response type
type ExecTypeResponse struct {
	Outcome string                 `json:"outcome"`
	Result  map[string]interface{} `json:"result"`
}

// HostResponse expected GetHost response type
type HostResponse struct {
	Outcome string   `json:"outcome"`
	Result  []string `json:"result"`
}

// DatasourceResponse expected GetDBStat response type
type DatasourceResponse struct {
	Outcome string          `json:"outcome"`
	Result  DatabaseMetrics `json:"result"`
}

// JMSResponse expected GetJMSTopicStat/GetJMSQueueStat response type
type JMSResponse struct {
	Outcome string                 `json:"outcome"`
	Result  map[string]interface{} `json:"result"`
}

// TransactionResponse transaction related metrics
type TransactionResponse struct {
	Outcome string                 `json:"outcome"`
	Result  map[string]interface{} `json:"result"`
}

// DatabaseMetrics database related metrics
type DatabaseMetrics struct {
	DataSource   map[string]DataSourceMetrics `json:"data-source"`
	XaDataSource map[string]DataSourceMetrics `json:"xa-data-source"`
}

//DataSourceMetrics Datasource related metrics
type DataSourceMetrics struct {
	JndiName   string       `json:"jndi-name"`
	Statistics DBStatistics `json:"statistics"`
}

// DBStatistics DB statistics per pool
type DBStatistics struct {
	Pool DBPoolStatistics `json:"pool"`
}

// DBPoolStatistics pool related statistics
type DBPoolStatistics struct {
	ActiveCount    interface{} `json:"ActiveCount"`
	AvailableCount interface{} `json:"AvailableCount"`
	InUseCount     interface{} `json:"InUseCount"`
}

// JVMResponse GetJVMStat expected response type
type JVMResponse struct {
	Outcome string     `json:"outcome"`
	Result  JVMMetrics `json:"result"`
}

// JVMMetrics JVM related metrics type
type JVMMetrics struct {
	Type map[string]interface{} `json:"type"`
}

// WebResponse getWebStatistics expected response type
type WebResponse struct {
	Outcome string                 `json:"outcome"`
	Result  map[string]interface{} `json:"result"`
}

// DeploymentResponse GetDeployments expected response type
type DeploymentResponse struct {
	Outcome string            `json:"outcome"`
	Result  DeploymentMetrics `json:"result"`
}

// DeploymentMetrics deployment related type
type DeploymentMetrics struct {
	Name          string                 `json:"name"`
	RuntimeName   string                 `json:"runtime-name"`
	Status        string                 `json:"status"`
	Subdeployment map[string]interface{} `json:"subdeployment"`
	Subsystem     map[string]interface{} `json:"subsystem"`
}

// WebMetrics  Web Modules related metrics
type WebMetrics struct {
	ActiveSessions    string                 `json:"active-sessions"`
	ContextRoot       string                 `json:"context-root"`
	ExpiredSessions   string                 `json:"expired-sessions"`
	MaxActiveSessions string                 `json:"max-active-sessions"`
	SessionsCreated   string                 `json:"sessions-created"`
	Servlet           map[string]interface{} `json:"servlet"`
}
