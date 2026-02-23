package mongodb

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/testutil"
)

func TestInitSuccess(t *testing.T) {
	tests := []struct {
		name        string
		granularity string
		database    string
	}{
		{
			name:        "missing metric database",
			granularity: "seconds",
		},
		{
			name:     "missing metric granularity",
			database: "telegraf_test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &MongoDB{
				Dsn:                "mongodb://localhost:27017",
				AuthenticationType: "NONE",
				MetricGranularity:  tt.granularity,
				MetricDatabase:     tt.database,
			}
			require.NoError(t, plugin.Init())
		})
	}
}

func TestInitFail(t *testing.T) {
	tests := []struct {
		name        string
		dsn         string
		auth        string
		username    config.Secret
		password    config.Secret
		granularity string
		expected    string
	}{
		{
			name:        "invalid metric granularity",
			dsn:         "mongodb://localhost:27017",
			auth:        "NONE",
			granularity: "somerandomgranularitythatdoesntwork",
			expected:    "invalid time series collection granularity",
		},
		{
			name:        "invalid connection string",
			dsn:         "asdf1234",
			auth:        "NONE",
			granularity: "seconds",
			expected:    "invalid connection string",
		},
		{
			name:        "invalid authentication type",
			dsn:         "mongodb://localhost:27017",
			auth:        "UNSUPPORTED",
			granularity: "seconds",
			expected:    "unsupported authentication type",
		},
		{
			name:        "plain missing username",
			dsn:         "mongodb://localhost:27017",
			auth:        "PLAIN",
			granularity: "seconds",
			expected:    "authentication for PLAIN must specify a username",
		},
		{
			name:        "plain missing password",
			dsn:         "mongodb://localhost:27017",
			auth:        "PLAIN",
			username:    config.NewSecret([]byte("somerandomusernamethatwontwork")),
			granularity: "seconds",
			expected:    "authentication for PLAIN must specify a password",
		},
		{
			name:        "scram missing username",
			dsn:         "mongodb://localhost:27017",
			auth:        "SCRAM",
			password:    config.NewSecret([]byte("somerandompasswordthatwontwork")),
			granularity: "seconds",
			expected:    "authentication for SCRAM must specify a username",
		},
		{
			name:        "scram missing password",
			dsn:         "mongodb://localhost:27017",
			auth:        "SCRAM",
			username:    config.NewSecret([]byte("somerandomusernamethatwontwork")),
			granularity: "seconds",
			expected:    "authentication for SCRAM must specify a password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &MongoDB{
				Dsn:                tt.dsn,
				AuthenticationType: tt.auth,
				Username:           tt.username,
				Password:           tt.password,
				MetricDatabase:     "telegraf_test",
				MetricGranularity:  tt.granularity,
			}
			require.ErrorContains(t, plugin.Init(), tt.expected)
		})
	}
}

func TestConnectAndWriteNoAuthIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	servicePort := "27017"
	container := testutil.Container{
		Image:        "mongo",
		ExposedPorts: []string{servicePort},
		WaitingFor:   wait.ForLog("Waiting for connections"),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Run test
	plugin := &MongoDB{
		Dsn:                "mongodb://" + container.Address + ":" + container.Ports[servicePort],
		AuthenticationType: "NONE",
		MetricDatabase:     "telegraf_test",
		MetricGranularity:  "seconds",
	}

	// validate config
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	require.NoError(t, plugin.Write(testutil.MockMetrics()))
	require.NoError(t, plugin.Close())
}

func TestConnectAndWriteSCRAMAuthIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	initdb, err := filepath.Abs("testdata/auth_scram/setup.js")
	require.NoError(t, err)

	servicePort := "27017"
	container := testutil.Container{
		Image:        "mongo",
		ExposedPorts: []string{servicePort},
		Files: map[string]string{
			"/docker-entrypoint-initdb.d/setup.js": initdb,
		},
		WaitingFor: wait.ForLog("Waiting for connections").WithOccurrence(2),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Setup plugin
	plugin := &MongoDB{
		Dsn: fmt.Sprintf("mongodb://%s:%s/admin",
			container.Address, container.Ports[servicePort]),
		AuthenticationType: "SCRAM",
		Username:           config.NewSecret([]byte("root")),
		Password:           config.NewSecret([]byte("changeme")),
		MetricDatabase:     "telegraf_test",
		MetricGranularity:  "seconds",
	}
	require.NoError(t, plugin.Init())

	// Connect and write metrics
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	require.NoError(t, plugin.Write(testutil.MockMetrics()))
}

func TestConnectAndWriteSCRAMAuthBadPasswordIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	initdb, err := filepath.Abs("testdata/auth_scram/setup.js")
	require.NoError(t, err)

	servicePort := "27017"
	container := testutil.Container{
		Image:        "mongo",
		ExposedPorts: []string{servicePort},
		Files: map[string]string{
			"/docker-entrypoint-initdb.d/setup.js": initdb,
		},
		WaitingFor: wait.ForLog("Waiting for connections").WithOccurrence(2),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Setup plugin
	plugin := &MongoDB{
		Dsn:                 "mongodb://" + container.Address + ":" + container.Ports[servicePort] + "/admin",
		AuthenticationType:  "SCRAM",
		Username:            config.NewSecret([]byte("root")),
		Password:            config.NewSecret([]byte("root")),
		MetricDatabase:      "telegraf_test",
		MetricGranularity:   "seconds",
		ServerSelectTimeout: config.Duration(time.Duration(5) * time.Second),
	}
	require.NoError(t, plugin.Init())

	// Check for failing connect
	require.ErrorContains(t, plugin.Connect(), "Authentication failed")
}

func TestConnectAndWriteX509AuthSuccessIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pki := testutil.NewPKI("../../../testutil/pki")

	// Bind mount files
	initdb, err := filepath.Abs("testdata/auth_x509/setup.js")
	require.NoError(t, err)
	cacert, err := filepath.Abs(pki.CACertPath())
	require.NoError(t, err)
	serverpem, err := filepath.Abs(pki.ServerCertAndKeyPath())
	require.NoError(t, err)

	servicePort := "27017"
	container := testutil.Container{
		Image:        "mongo",
		ExposedPorts: []string{servicePort},
		Files: map[string]string{
			"/docker-entrypoint-initdb.d/setup.js": initdb,
			"/cacert.pem":                          cacert,
			"/server.pem":                          serverpem,
		},
		Entrypoint: []string{
			"docker-entrypoint.sh",
			"--auth", "--setParameter", "authenticationMechanisms=MONGODB-X509",
			"--tlsMode", "preferTLS",
			"--tlsCAFile", "/cacert.pem",
			"--tlsCertificateKeyFile", "/server.pem",
		},
		WaitingFor: wait.ForLog("Waiting for connections").WithOccurrence(2),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	tests := []struct {
		name      string
		tlsConfig tls.ClientConfig
	}{
		{
			name: "default",
			tlsConfig: tls.ClientConfig{
				TLSCA:  pki.CACertPath(),
				TLSKey: pki.ClientCertAndKeyPath(),
			},
		},
		{
			name: "encrypted key file",
			tlsConfig: tls.ClientConfig{
				TLSCA:     pki.CACertPath(),
				TLSKey:    pki.ClientCertAndEncKeyPath(),
				TLSKeyPwd: "changeme",
			},
		},
		{
			name: "missing ca and insecure tls",
			tlsConfig: tls.ClientConfig{
				TLSKey:             pki.ClientCertAndKeyPath(),
				InsecureSkipVerify: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup plugin
			plugin := &MongoDB{
				Dsn:                 "mongodb://" + container.Address + ":" + container.Ports[servicePort],
				AuthenticationType:  "X509",
				MetricDatabase:      "telegraf_test",
				MetricGranularity:   "seconds",
				ServerSelectTimeout: config.Duration(time.Duration(5) * time.Second),
				TTL:                 config.Duration(time.Duration(5) * time.Minute),
				ClientConfig:        tt.tlsConfig,
			}
			require.NoError(t, plugin.Init())

			// Connect and write metrics
			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			require.NoError(t, plugin.Write(testutil.MockMetrics()))
		})
	}
}

func TestConnectAndWriteX509AuthFailIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pki := testutil.NewPKI("../../../testutil/pki")

	// Bind mount files
	initdb, err := filepath.Abs("testdata/auth_x509/setup.js")
	require.NoError(t, err)
	cacert, err := filepath.Abs(pki.CACertPath())
	require.NoError(t, err)
	serverpem, err := filepath.Abs(pki.ServerCertAndKeyPath())
	require.NoError(t, err)

	servicePort := "27017"
	container := testutil.Container{
		Image:        "mongo",
		ExposedPorts: []string{servicePort},
		Files: map[string]string{
			"/docker-entrypoint-initdb.d/setup.js": initdb,
			"/cacert.pem":                          cacert,
			"/server.pem":                          serverpem,
		},
		Entrypoint: []string{
			"docker-entrypoint.sh",
			"--auth", "--setParameter", "authenticationMechanisms=MONGODB-X509",
			"--tlsMode", "preferTLS",
			"--tlsCAFile", "/cacert.pem",
			"--tlsCertificateKeyFile", "/server.pem",
		},
		WaitingFor: wait.ForLog("Waiting for connections").WithOccurrence(2),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	tests := []struct {
		name      string
		tlsConfig tls.ClientConfig
		expected  string
	}{
		{
			name: "missing ca",
			tlsConfig: tls.ClientConfig{
				TLSKey: pki.ClientCertAndKeyPath(),
			},
			expected: "certificate signed by unknown authority",
		},
		{
			name: "invalid ca",
			tlsConfig: tls.ClientConfig{
				TLSCA:  pki.ClientCertAndKeyPath(),
				TLSKey: pki.ClientCertAndKeyPath(),
			},
			expected: "certificate signed by unknown authority",
		},
		{
			name: "invalid TLS key",
			tlsConfig: tls.ClientConfig{
				TLSCA:  pki.CACertPath(),
				TLSKey: pki.CACertPath(),
			},
			expected: "failed to find PRIVATE KEY",
		},
		{
			name: "wrong password for encrypted key file",
			tlsConfig: tls.ClientConfig{
				TLSCA:     pki.CACertPath(),
				TLSKey:    pki.ClientCertAndEncKeyPath(),
				TLSKeyPwd: "badpassword",
			},
			expected: "decryption password incorrect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup plugin
			plugin := &MongoDB{
				Dsn:                 "mongodb://" + container.Address + ":" + container.Ports[servicePort],
				AuthenticationType:  "X509",
				MetricDatabase:      "telegraf_test",
				MetricGranularity:   "seconds",
				ServerSelectTimeout: config.Duration(time.Duration(5) * time.Second),
				TTL:                 config.Duration(time.Duration(5) * time.Minute),
				ClientConfig:        tt.tlsConfig,
			}
			require.NoError(t, plugin.Init())

			// Check for failing connect
			require.ErrorContains(t, plugin.Connect(), tt.expected)
		})
	}
}
