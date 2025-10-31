package ldap

import (
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	common_tls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/testutil"
)

const (
	servicePortOpenLDAP       = "1389"
	servicePortOpenLDAPSecure = "1636"

	servicePort389DS       = "3389"
	servicePort389DSSecure = "3636"
)

func TestMockResult(t *testing.T) {
	// mock a query result
	mockSearchResult := &ldap.SearchResult{
		Entries: []*ldap.Entry{
			{
				DN:         "cn=Total,cn=Connections,cn=Monitor",
				Attributes: []*ldap.EntryAttribute{{Name: "monitorCounter", Values: []string{"1"}}},
			},
		},
	}

	// Setup the plugin
	plugin := &LDAP{}
	require.NoError(t, plugin.Init())

	// Setup the expectations
	expected := []telegraf.Metric{
		metric.New(
			"openldap",
			map[string]string{
				"server": "localhost",
				"port":   "389",
			},
			map[string]interface{}{
				"total_connections": int64(1),
			},
			time.Unix(0, 0),
		),
	}

	// Retrieve the converter
	requests := plugin.newOpenLDAPConfig()
	require.Len(t, requests, 1)
	converter := requests[0].convert
	require.NotNil(t, converter)

	// Test metric conversion
	actual := converter(mockSearchResult, time.Unix(0, 0))
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestInvalidTLSMode(t *testing.T) {
	plugin := &LDAP{
		Server: "foo://localhost",
	}
	require.ErrorContains(t, plugin.Init(), "invalid scheme")
}

func TestNoConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup the plugin
	plugin := &LDAP{Server: "ldap://nosuchhost"}
	require.NoError(t, plugin.Init())

	// Collect the metrics and compare
	var acc testutil.Accumulator
	require.ErrorContains(t, plugin.Gather(&acc), "connection failed")
	require.Empty(t, acc.GetTelegrafMetrics())
}

func TestOpenLDAPIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start the docker container
	container := testutil.Container{
		Image:        "bitnamilegacy/openldap",
		ExposedPorts: []string{servicePortOpenLDAP},
		Env: map[string]string{
			"LDAP_ADMIN_USERNAME": "manager",
			"LDAP_ADMIN_PASSWORD": "secret",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("slapd starting"),
			wait.ForListeningPort(nat.Port(servicePortOpenLDAP)),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Setup the plugin
	port := container.Ports[servicePortOpenLDAP]
	plugin := &LDAP{
		Server:       "ldap://" + container.Address + ":" + port,
		BindDn:       "CN=manager,DC=example,DC=org",
		BindPassword: config.NewSecret([]byte("secret")),
	}
	require.NoError(t, plugin.Init())

	// Setup the expectations
	expected := []telegraf.Metric{
		metric.New(
			"openldap",
			map[string]string{
				"server": container.Address,
				"port":   port,
			},
			map[string]interface{}{
				"abandon_operations_completed":     int64(0),
				"abandon_operations_initiated":     int64(0),
				"active_threads":                   int64(0),
				"add_operations_completed":         int64(0),
				"add_operations_initiated":         int64(0),
				"backload_threads":                 int64(0),
				"bind_operations_completed":        int64(0),
				"bind_operations_initiated":        int64(0),
				"bytes_statistics":                 int64(0),
				"compare_operations_completed":     int64(0),
				"compare_operations_initiated":     int64(0),
				"current_connections":              int64(0),
				"delete_operations_completed":      int64(0),
				"delete_operations_initiated":      int64(0),
				"entries_statistics":               int64(0),
				"extended_operations_completed":    int64(0),
				"extended_operations_initiated":    int64(0),
				"max_file_descriptors_connections": int64(0),
				"max_pending_threads":              int64(0),
				"max_threads":                      int64(0),
				"modify_operations_completed":      int64(0),
				"modify_operations_initiated":      int64(0),
				"modrdn_operations_completed":      int64(0),
				"modrdn_operations_initiated":      int64(0),
				"open_threads":                     int64(0),
				"operations_completed":             int64(0),
				"operations_initiated":             int64(0),
				"pdu_statistics":                   int64(0),
				"pending_threads":                  int64(0),
				"read_waiters":                     int64(0),
				"referrals_statistics":             int64(0),
				"search_operations_completed":      int64(0),
				"search_operations_initiated":      int64(0),
				"starting_threads":                 int64(0),
				"total_connections":                int64(0),
				"unbind_operations_completed":      int64(0),
				"unbind_operations_initiated":      int64(0),
				"uptime_time":                      int64(0),
				"write_waiters":                    int64(0),
			},
			time.Unix(0, 0),
		),
	}

	// Collect the metrics and compare
	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsStructureEqual(t, expected, actual, testutil.IgnoreTime())
}

func TestOpenLDAPReverseDNIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start the docker container
	container := testutil.Container{
		Image:        "bitnamilegacy/openldap",
		ExposedPorts: []string{servicePortOpenLDAP},
		Env: map[string]string{
			"LDAP_ADMIN_USERNAME": "manager",
			"LDAP_ADMIN_PASSWORD": "secret",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("slapd starting"),
			wait.ForListeningPort(nat.Port(servicePortOpenLDAP)),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Setup the plugin
	port := container.Ports[servicePortOpenLDAP]
	plugin := &LDAP{
		Server:            "ldap://" + container.Address + ":" + port,
		BindDn:            "CN=manager,DC=example,DC=org",
		BindPassword:      config.NewSecret([]byte("secret")),
		ReverseFieldNames: true,
	}
	require.NoError(t, plugin.Init())

	// Setup the expectations
	expected := []telegraf.Metric{
		metric.New(
			"openldap",
			map[string]string{
				"server": container.Address,
				"port":   port,
			},
			map[string]interface{}{
				"connections_max_file_descriptors": int64(0),
				"connections_total":                int64(0),
				"connections_current":              int64(0),
				"operations_bind_initiated":        int64(0),
				"operations_bind_completed":        int64(0),
				"operations_completed":             int64(0),
				"operations_initiated":             int64(0),
				"operations_unbind_initiated":      int64(0),
				"operations_unbind_completed":      int64(0),
				"operations_search_initiated":      int64(0),
				"operations_search_completed":      int64(0),
				"operations_compare_initiated":     int64(0),
				"operations_compare_completed":     int64(0),
				"operations_modify_initiated":      int64(0),
				"operations_modify_completed":      int64(0),
				"operations_modrdn_initiated":      int64(0),
				"operations_modrdn_completed":      int64(0),
				"operations_add_initiated":         int64(0),
				"operations_add_completed":         int64(0),
				"operations_delete_initiated":      int64(0),
				"operations_delete_completed":      int64(0),
				"operations_abandon_initiated":     int64(0),
				"operations_abandon_completed":     int64(0),
				"operations_extended_initiated":    int64(0),
				"operations_extended_completed":    int64(0),
				"statistics_bytes":                 int64(0),
				"statistics_pdu":                   int64(0),
				"statistics_entries":               int64(0),
				"statistics_referrals":             int64(0),
				"threads_max":                      int64(0),
				"threads_max_pending":              int64(0),
				"threads_open":                     int64(0),
				"threads_starting":                 int64(0),
				"threads_active":                   int64(0),
				"threads_pending":                  int64(0),
				"threads_backload":                 int64(0),
				"time_uptime":                      int64(0),
				"waiters_read":                     int64(0),
				"waiters_write":                    int64(0),
			},
			time.Unix(0, 0),
		),
	}

	// Collect the metrics and compare
	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsStructureEqual(t, expected, actual, testutil.IgnoreTime())
}

func TestOpenLDAPStartTLSIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup PKI for TLS testing
	pkiPaths, err := testutil.NewPKI("../../../testutil/pki").AbsolutePaths()
	require.NoError(t, err)

	// Start the docker container
	container := testutil.Container{
		Image:        "bitnamilegacy/openldap",
		ExposedPorts: []string{servicePortOpenLDAP},
		Env: map[string]string{
			"LDAP_ADMIN_USERNAME": "manager",
			"LDAP_ADMIN_PASSWORD": "secret",
			"LDAP_ENABLE_TLS":     "yes",
			"LDAP_TLS_CA_FILE":    "server.pem",
			"LDAP_TLS_CERT_FILE":  "server.crt",
			"LDAP_TLS_KEY_FILE":   "server.key",
		},
		Files: map[string]string{
			"/server.pem": pkiPaths.ServerPem,
			"/server.crt": pkiPaths.ServerCert,
			"/server.key": pkiPaths.ServerKey,
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("slapd starting"),
			wait.ForListeningPort(nat.Port(servicePortOpenLDAP)),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Setup the plugin
	port := container.Ports[servicePortOpenLDAP]
	plugin := &LDAP{
		Server:       "starttls://" + container.Address + ":" + port,
		BindDn:       "CN=manager,DC=example,DC=org",
		BindPassword: config.NewSecret([]byte("secret")),
		ClientConfig: common_tls.ClientConfig{
			TLSCA:              pkiPaths.ClientCert,
			InsecureSkipVerify: true,
		},
	}
	require.NoError(t, plugin.Init())

	// Setup the expectations
	expected := []telegraf.Metric{
		metric.New(
			"openldap",
			map[string]string{
				"server": container.Address,
				"port":   port,
			},
			map[string]interface{}{
				"abandon_operations_completed":     int64(0),
				"abandon_operations_initiated":     int64(0),
				"active_threads":                   int64(0),
				"add_operations_completed":         int64(0),
				"add_operations_initiated":         int64(0),
				"backload_threads":                 int64(0),
				"bind_operations_completed":        int64(0),
				"bind_operations_initiated":        int64(0),
				"bytes_statistics":                 int64(0),
				"compare_operations_completed":     int64(0),
				"compare_operations_initiated":     int64(0),
				"current_connections":              int64(0),
				"delete_operations_completed":      int64(0),
				"delete_operations_initiated":      int64(0),
				"entries_statistics":               int64(0),
				"extended_operations_completed":    int64(0),
				"extended_operations_initiated":    int64(0),
				"max_file_descriptors_connections": int64(0),
				"max_pending_threads":              int64(0),
				"max_threads":                      int64(0),
				"modify_operations_completed":      int64(0),
				"modify_operations_initiated":      int64(0),
				"modrdn_operations_completed":      int64(0),
				"modrdn_operations_initiated":      int64(0),
				"open_threads":                     int64(0),
				"operations_completed":             int64(0),
				"operations_initiated":             int64(0),
				"pdu_statistics":                   int64(0),
				"pending_threads":                  int64(0),
				"read_waiters":                     int64(0),
				"referrals_statistics":             int64(0),
				"search_operations_completed":      int64(0),
				"search_operations_initiated":      int64(0),
				"starting_threads":                 int64(0),
				"total_connections":                int64(0),
				"unbind_operations_completed":      int64(0),
				"unbind_operations_initiated":      int64(0),
				"uptime_time":                      int64(0),
				"write_waiters":                    int64(0),
			},
			time.Unix(0, 0),
		),
	}

	// Collect the metrics and compare
	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsStructureEqual(t, expected, actual, testutil.IgnoreTime())
}

func TestOpenLDAPLDAPSIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup PKI for TLS testing
	pkiPaths, err := testutil.NewPKI("../../../testutil/pki").AbsolutePaths()
	require.NoError(t, err)

	// Start the docker container
	container := testutil.Container{
		Image:        "bitnamilegacy/openldap",
		ExposedPorts: []string{servicePortOpenLDAPSecure},
		Env: map[string]string{
			"LDAP_ADMIN_USERNAME": "manager",
			"LDAP_ADMIN_PASSWORD": "secret",
			"LDAP_ENABLE_TLS":     "yes",
			"LDAP_TLS_CA_FILE":    "server.pem",
			"LDAP_TLS_CERT_FILE":  "server.crt",
			"LDAP_TLS_KEY_FILE":   "server.key",
		},
		Files: map[string]string{
			"/server.pem": pkiPaths.ServerPem,
			"/server.crt": pkiPaths.ServerCert,
			"/server.key": pkiPaths.ServerKey,
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("slapd starting"),
			wait.ForListeningPort(nat.Port(servicePortOpenLDAPSecure)),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Setup the plugin
	port := container.Ports[servicePortOpenLDAPSecure]
	plugin := &LDAP{
		Server:       "ldaps://" + container.Address + ":" + port,
		BindDn:       "CN=manager,DC=example,DC=org",
		BindPassword: config.NewSecret([]byte("secret")),
		ClientConfig: common_tls.ClientConfig{
			InsecureSkipVerify: true,
		},
	}
	require.NoError(t, plugin.Init())

	// Setup the expectations
	expected := []telegraf.Metric{
		metric.New(
			"openldap",
			map[string]string{
				"server": container.Address,
				"port":   port,
			},
			map[string]interface{}{
				"abandon_operations_completed":     int64(0),
				"abandon_operations_initiated":     int64(0),
				"active_threads":                   int64(0),
				"add_operations_completed":         int64(0),
				"add_operations_initiated":         int64(0),
				"backload_threads":                 int64(0),
				"bind_operations_completed":        int64(0),
				"bind_operations_initiated":        int64(0),
				"bytes_statistics":                 int64(0),
				"compare_operations_completed":     int64(0),
				"compare_operations_initiated":     int64(0),
				"current_connections":              int64(0),
				"delete_operations_completed":      int64(0),
				"delete_operations_initiated":      int64(0),
				"entries_statistics":               int64(0),
				"extended_operations_completed":    int64(0),
				"extended_operations_initiated":    int64(0),
				"max_file_descriptors_connections": int64(0),
				"max_pending_threads":              int64(0),
				"max_threads":                      int64(0),
				"modify_operations_completed":      int64(0),
				"modify_operations_initiated":      int64(0),
				"modrdn_operations_completed":      int64(0),
				"modrdn_operations_initiated":      int64(0),
				"open_threads":                     int64(0),
				"operations_completed":             int64(0),
				"operations_initiated":             int64(0),
				"pdu_statistics":                   int64(0),
				"pending_threads":                  int64(0),
				"read_waiters":                     int64(0),
				"referrals_statistics":             int64(0),
				"search_operations_completed":      int64(0),
				"search_operations_initiated":      int64(0),
				"starting_threads":                 int64(0),
				"total_connections":                int64(0),
				"unbind_operations_completed":      int64(0),
				"unbind_operations_initiated":      int64(0),
				"uptime_time":                      int64(0),
				"write_waiters":                    int64(0),
			},
			time.Unix(0, 0),
		),
	}

	// Collect the metrics and compare
	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsStructureEqual(t, expected, actual, testutil.IgnoreTime())
}

func Test389dsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start the docker container
	container := testutil.Container{
		Image:        "389ds/dirsrv",
		ExposedPorts: []string{servicePort389DS},
		Env: map[string]string{
			"DS_DM_PASSWORD": "secret",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("389-ds-container started"),
			wait.ForListeningPort(nat.Port(servicePort389DS)),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Setup the plugin
	port := container.Ports[servicePort389DS]
	plugin := &LDAP{
		Server:       "ldap://" + container.Address + ":" + port,
		Dialect:      "389ds",
		BindDn:       "cn=Directory Manager",
		BindPassword: config.NewSecret([]byte("secret")),
	}
	require.NoError(t, plugin.Init())

	// Setup the expectations
	expected := []telegraf.Metric{
		metric.New(
			"389ds",
			map[string]string{
				"server": container.Address,
				"port":   port,
			},
			map[string]interface{}{
				"add_operations":                     int64(0),
				"anonymous_binds":                    int64(0),
				"backends":                           int64(0),
				"bind_security_errors":               int64(0),
				"bytes_received":                     int64(0),
				"bytes_sent":                         int64(0),
				"cache_entries":                      int64(0),
				"cache_hits":                         int64(0),
				"chainings":                          int64(0),
				"compare_operations":                 int64(0),
				"connections":                        int64(0),
				"connections_in_max_threads":         int64(0),
				"connections_max_threads":            int64(0),
				"copy_entries":                       int64(0),
				"current_connections":                int64(0),
				"current_connections_at_max_threads": int64(0),
				"delete_operations":                  int64(0),
				"dtablesize":                         int64(0),
				"entries_returned":                   int64(0),
				"entries_sent":                       int64(0),
				"errors":                             int64(0),
				"in_operations":                      int64(0),
				"list_operations":                    int64(0),
				"maxthreads_per_conn_hits":           int64(0),
				"modify_operations":                  int64(0),
				"modrdn_operations":                  int64(0),
				"onelevel_search_operations":         int64(0),
				"operations_completed":               int64(0),
				"operations_initiated":               int64(0),
				"read_operations":                    int64(0),
				"read_waiters":                       int64(0),
				"referrals":                          int64(0),
				"referrals_returned":                 int64(0),
				"search_operations":                  int64(0),
				"security_errors":                    int64(0),
				"simpleauth_binds":                   int64(0),
				"strongauth_binds":                   int64(0),
				"threads":                            int64(0),
				"total_connections":                  int64(0),
				"unauth_binds":                       int64(0),
				"wholesubtree_search_operations":     int64(0),
			},
			time.Unix(0, 0),
		),
	}

	// Collect the metrics and compare
	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsStructureEqual(t, expected, actual, testutil.IgnoreTime())
}
