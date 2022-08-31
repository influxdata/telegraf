package openldap

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/testutil"
)

const (
	servicePort       = "1389"
	servicePortSecure = "1636"
)

func TestOpenldapMockResult(t *testing.T) {
	var acc testutil.Accumulator

	mockSearchResult := ldap.SearchResult{
		Entries: []*ldap.Entry{
			{
				DN:         "cn=Total,cn=Connections,cn=Monitor",
				Attributes: []*ldap.EntryAttribute{{Name: "monitorCounter", Values: []string{"1"}}},
			},
		},
		Referrals: []string{},
		Controls:  []ldap.Control{},
	}

	o := &Openldap{
		Host: "localhost",
		Port: 389,
	}

	gatherSearchResult(&mockSearchResult, o, &acc)
	commonTests(t, o, &acc)
}

func TestOpenldapNoConnectionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Openldap{
		Host: "nosuchhost",
		Port: 389,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)         // test that we didn't return an error
	require.Zero(t, acc.NFields())  // test that we didn't return any fields
	require.NotEmpty(t, acc.Errors) // test that we set an error
}

func TestOpenldapGeneratesMetricsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image:        "bitnami/openldap",
		ExposedPorts: []string{servicePort},
		Env: map[string]string{
			"LDAP_ADMIN_USERNAME": "manager",
			"LDAP_ADMIN_PASSWORD": "secret",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("slapd starting"),
			wait.ForListeningPort(nat.Port(servicePort)),
		),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	port, err := strconv.Atoi(container.Ports[servicePort])
	require.NoError(t, err)

	o := &Openldap{
		Host:         container.Address,
		Port:         port,
		BindDn:       "CN=manager,DC=example,DC=org",
		BindPassword: "secret",
	}

	var acc testutil.Accumulator
	err = o.Gather(&acc)
	require.NoError(t, err)
	commonTests(t, o, &acc)
}

func TestOpenldapStartTLSIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pki := testutil.NewPKI("../../../testutil/pki")

	tlsPem, err := filepath.Abs(pki.ServerCertAndKeyPath())
	require.NoError(t, err)
	tlsCert, err := filepath.Abs(pki.ServerCertPath())
	require.NoError(t, err)
	tlsKey, err := filepath.Abs(pki.ServerKeyPath())
	require.NoError(t, err)

	container := testutil.Container{
		Image:        "bitnami/openldap",
		ExposedPorts: []string{servicePort},
		Env: map[string]string{
			"LDAP_ADMIN_USERNAME": "manager",
			"LDAP_ADMIN_PASSWORD": "secret",
			"LDAP_ENABLE_TLS":     "yes",
			"LDAP_TLS_CA_FILE":    "server.pem",
			"LDAP_TLS_CERT_FILE":  "server.crt",
			"LDAP_TLS_KEY_FILE":   "server.key",
		},
		BindMounts: map[string]string{
			"/server.pem": tlsPem,
			"/server.crt": tlsCert,
			"/server.key": tlsKey,
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("slapd starting"),
			wait.ForListeningPort(nat.Port(servicePort)),
		),
	}
	err = container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	port, err := strconv.Atoi(container.Ports[servicePort])
	require.NoError(t, err)

	cert, err := filepath.Abs(pki.ClientCertPath())
	require.NoError(t, err)

	o := &Openldap{
		Host:               container.Address,
		Port:               port,
		InsecureSkipVerify: true,
		BindDn:             "CN=manager,DC=example,DC=org",
		BindPassword:       "secret",
		TLSCA:              cert,
	}

	var acc testutil.Accumulator
	err = o.Gather(&acc)
	require.NoError(t, err)
	commonTests(t, o, &acc)
}

func TestOpenldapLDAPSIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pki := testutil.NewPKI("../../../testutil/pki")

	tlsPem, err := filepath.Abs(pki.ServerCertAndKeyPath())
	require.NoError(t, err)
	tlsCert, err := filepath.Abs(pki.ServerCertPath())
	require.NoError(t, err)
	tlsKey, err := filepath.Abs(pki.ServerKeyPath())
	require.NoError(t, err)

	container := testutil.Container{
		Image:        "bitnami/openldap",
		ExposedPorts: []string{servicePortSecure},
		Env: map[string]string{
			"LDAP_ADMIN_USERNAME": "manager",
			"LDAP_ADMIN_PASSWORD": "secret",
			"LDAP_ENABLE_TLS":     "yes",
			"LDAP_TLS_CA_FILE":    "server.pem",
			"LDAP_TLS_CERT_FILE":  "server.crt",
			"LDAP_TLS_KEY_FILE":   "server.key",
		},
		BindMounts: map[string]string{
			"/server.pem": tlsPem,
			"/server.crt": tlsCert,
			"/server.key": tlsKey,
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("slapd starting"),
			wait.ForListeningPort(nat.Port(servicePortSecure)),
		),
	}
	err = container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	port, err := strconv.Atoi(container.Ports[servicePortSecure])
	require.NoError(t, err)

	o := &Openldap{
		Host:               container.Address,
		Port:               port,
		InsecureSkipVerify: true,
		BindDn:             "CN=manager,DC=example,DC=org",
		BindPassword:       "secret",
	}

	var acc testutil.Accumulator
	err = o.Gather(&acc)
	require.NoError(t, err)
	commonTests(t, o, &acc)
}

func TestOpenldapInvalidSSLIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pki := testutil.NewPKI("../../../testutil/pki")
	tlsPem, err := filepath.Abs(pki.ServerCertAndKeyPath())
	require.NoError(t, err)
	tlsCert, err := filepath.Abs(pki.ServerCertPath())
	require.NoError(t, err)
	tlsKey, err := filepath.Abs(pki.ServerKeyPath())
	require.NoError(t, err)

	container := testutil.Container{
		Image:        "bitnami/openldap",
		ExposedPorts: []string{servicePortSecure},
		Env: map[string]string{
			"LDAP_ADMIN_USERNAME": "manager",
			"LDAP_ADMIN_PASSWORD": "secret",
			"LDAP_ENABLE_TLS":     "yes",
			"LDAP_TLS_CA_FILE":    "server.pem",
			"LDAP_TLS_CERT_FILE":  "server.crt",
			"LDAP_TLS_KEY_FILE":   "server.key",
		},
		BindMounts: map[string]string{
			"/server.pem": tlsPem,
			"/server.crt": tlsCert,
			"/server.key": tlsKey,
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("slapd starting"),
			wait.ForListeningPort(nat.Port(servicePortSecure)),
		),
	}
	err = container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	port, err := strconv.Atoi(container.Ports[servicePortSecure])
	require.NoError(t, err)

	o := &Openldap{
		Host:               container.Address,
		Port:               port,
		InsecureSkipVerify: true,
	}

	var acc testutil.Accumulator
	err = o.Gather(&acc)
	require.NoError(t, err)         // test that we didn't return an error
	require.Zero(t, acc.NFields())  // test that we didn't return any fields
	require.NotEmpty(t, acc.Errors) // test that we set an error
}

func TestOpenldapBindIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image:        "bitnami/openldap",
		ExposedPorts: []string{servicePort},
		Env: map[string]string{
			"LDAP_ADMIN_USERNAME": "manager",
			"LDAP_ADMIN_PASSWORD": "secret",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("slapd starting"),
			wait.ForListeningPort(nat.Port(servicePort)),
		),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	port, err := strconv.Atoi(container.Ports[servicePort])
	require.NoError(t, err)

	o := &Openldap{
		Host:               container.Address,
		Port:               port,
		InsecureSkipVerify: true,
		BindDn:             "CN=manager,DC=example,DC=org",
		BindPassword:       "secret",
	}

	var acc testutil.Accumulator
	err = o.Gather(&acc)
	require.NoError(t, err)
	commonTests(t, o, &acc)
}

func commonTests(t *testing.T, o *Openldap, acc *testutil.Accumulator) {
	// helpful local commands to run:
	// ldapwhoami -D "CN=manager,DC=example,DC=org" -H ldap://localhost:1389 -w secret
	// ldapsearch -D "CN=manager,DC=example,DC=org" -H "ldap://localhost:1389" -b cn=Monitor -w secret
	require.Empty(t, acc.Errors, "accumulator had no errors")
	require.True(t, acc.HasMeasurement("openldap"), "Has a measurement called 'openldap'")
	require.Equal(t, o.Host, acc.TagValue("openldap", "server"), "Has a tag value of server=o.Host")
	require.Equal(t, strconv.Itoa(o.Port), acc.TagValue("openldap", "port"), "Has a tag value of port=o.Port")
	require.True(t, acc.HasInt64Field("openldap", "total_connections"), "Has an integer field called total_connections")
}

func TestOpenldapReverseMetricsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image:        "bitnami/openldap",
		ExposedPorts: []string{servicePort},
		Env: map[string]string{
			"LDAP_ADMIN_USERNAME": "manager",
			"LDAP_ADMIN_PASSWORD": "secret",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("slapd starting"),
			wait.ForListeningPort(nat.Port(servicePort)),
		),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	port, err := strconv.Atoi(container.Ports[servicePort])
	require.NoError(t, err)

	o := &Openldap{
		Host:               container.Address,
		Port:               port,
		InsecureSkipVerify: true,
		BindDn:             "CN=manager,DC=example,DC=org",
		BindPassword:       "secret",
		ReverseMetricNames: true,
	}

	var acc testutil.Accumulator
	err = o.Gather(&acc)
	require.NoError(t, err)
	require.True(t, acc.HasInt64Field("openldap", "connections_total"), "Has an integer field called connections_total")
}
