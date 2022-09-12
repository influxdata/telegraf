package mongodb

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestConnectAndWriteIntegrationNoAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	servicePort := "27017"
	container := testutil.Container{
		Image:        "mongo",
		ExposedPorts: []string{servicePort},
		WaitingFor: wait.ForAll(
			wait.NewHTTPStrategy("/").WithPort(nat.Port(servicePort)),
			wait.ForLog("Waiting for connections"),
		),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	// Run test
	plugin := &MongoDB{
		Dsn: fmt.Sprintf("mongodb://%s:%s",
			container.Address,
			container.Ports[servicePort],
		),
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

func TestConnectAndWriteIntegrationSCRAMAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	initdb, err := filepath.Abs("testdata/auth_scram")
	require.NoError(t, err)

	servicePort := "27017"
	container := testutil.Container{
		Image:        "mongo",
		ExposedPorts: []string{servicePort},
		BindMounts: map[string]string{
			"/docker-entrypoint-initdb.d": initdb,
		},
		WaitingFor: wait.ForAll(
			wait.NewHTTPStrategy("/").WithPort(nat.Port(servicePort)),
			wait.ForLog("Waiting for connections"),
		),
	}
	err = container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	tests := []struct {
		name        string
		plugin      *MongoDB
		connErrFunc func(t *testing.T, err error)
	}{
		{
			name: "success with scram authentication",
			plugin: &MongoDB{
				Dsn: fmt.Sprintf("mongodb://%s:%s/admin",
					container.Address, container.Ports[servicePort]),
				AuthenticationType: "SCRAM",
				Username:           "root",
				Password:           "changeme",
				MetricDatabase:     "telegraf_test",
				MetricGranularity:  "seconds",
			},
			connErrFunc: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "fail with scram authentication bad password",
			plugin: &MongoDB{
				Dsn: fmt.Sprintf("mongodb://%s:%s/admin",
					container.Address, container.Ports[servicePort]),
				AuthenticationType:  "SCRAM",
				Username:            "root",
				Password:            "root",
				MetricDatabase:      "telegraf_test",
				MetricGranularity:   "seconds",
				ServerSelectTimeout: config.Duration(time.Duration(5) * time.Second),
			},
			connErrFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// validate config
			err := tt.plugin.Init()
			require.NoError(t, err)

			if err == nil {
				// connect
				err = tt.plugin.Connect()
				tt.connErrFunc(t, err)

				if err == nil {
					// insert mock metrics
					err = tt.plugin.Write(testutil.MockMetrics())
					require.NoError(t, err)

					// cleanup
					err = tt.plugin.Close()
					require.NoError(t, err)
				}
			}
		})
	}
}

func TestConnectAndWriteIntegrationX509Auth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pki := testutil.NewPKI("../../../testutil/pki")

	// bind mount files
	initdb, err := filepath.Abs("testdata/auth_x509")
	require.NoError(t, err)
	cacert, err := filepath.Abs(pki.CACertPath())
	require.NoError(t, err)
	serverpem, err := filepath.Abs(pki.ServerCertAndKeyPath())
	require.NoError(t, err)

	servicePort := "27017"
	container := testutil.Container{
		Image:        "mongo",
		ExposedPorts: []string{servicePort},
		BindMounts: map[string]string{
			"/docker-entrypoint-initdb.d": initdb,
			"/cacert.pem":                 cacert,
			"/server.pem":                 serverpem,
		},
		Entrypoint: []string{
			"docker-entrypoint.sh",
			"--auth", "--setParameter", "authenticationMechanisms=MONGODB-X509",
			"--tlsMode", "preferTLS",
			"--tlsCAFile", "/cacert.pem",
			"--tlsCertificateKeyFile", "/server.pem",
		},
		WaitingFor: wait.ForAll(
			wait.NewHTTPStrategy("/").WithPort(nat.Port(servicePort)),
			wait.ForLog("Waiting for connections"),
		),
	}
	err = container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	tests := []struct {
		name        string
		plugin      *MongoDB
		connErrFunc func(t *testing.T, err error)
	}{
		{
			name: "success with x509 authentication",
			plugin: &MongoDB{
				Dsn: fmt.Sprintf("mongodb://%s:%s",
					container.Address, container.Ports[servicePort]),
				AuthenticationType:  "X509",
				MetricDatabase:      "telegraf_test",
				MetricGranularity:   "seconds",
				ServerSelectTimeout: config.Duration(time.Duration(5) * time.Second),
				TTL:                 config.Duration(time.Duration(5) * time.Minute),
				ClientConfig: tls.ClientConfig{
					TLSCA:              pki.CACertPath(),
					TLSKey:             pki.ClientCertAndKeyPath(),
					InsecureSkipVerify: false,
				},
			},
			connErrFunc: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "success with x509 authentication using encrypted key file",
			plugin: &MongoDB{
				Dsn: fmt.Sprintf("mongodb://%s:%s",
					container.Address, container.Ports[servicePort]),
				AuthenticationType:  "X509",
				MetricDatabase:      "telegraf_test",
				MetricGranularity:   "seconds",
				ServerSelectTimeout: config.Duration(time.Duration(5) * time.Second),
				TTL:                 config.Duration(time.Duration(5) * time.Minute),
				ClientConfig: tls.ClientConfig{
					TLSCA:              pki.CACertPath(),
					TLSKey:             pki.ClientCertAndEncKeyPath(),
					TLSKeyPwd:          "changeme",
					InsecureSkipVerify: false,
				},
			},
			connErrFunc: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "success with x509 authentication missing ca and using insceure tls",
			plugin: &MongoDB{
				Dsn: fmt.Sprintf("mongodb://%s:%s",
					container.Address, container.Ports[servicePort]),
				AuthenticationType:  "X509",
				MetricDatabase:      "telegraf_test",
				MetricGranularity:   "seconds",
				ServerSelectTimeout: config.Duration(time.Duration(5) * time.Second),
				TTL:                 config.Duration(time.Duration(5) * time.Minute),
				ClientConfig: tls.ClientConfig{
					TLSKey:             pki.ClientCertAndKeyPath(),
					InsecureSkipVerify: true,
				},
			},
			connErrFunc: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "fail with x509 authentication missing ca",
			plugin: &MongoDB{
				Dsn: fmt.Sprintf("mongodb://%s:%s",
					container.Address, container.Ports[servicePort]),
				AuthenticationType:  "X509",
				MetricDatabase:      "telegraf_test",
				MetricGranularity:   "seconds",
				ServerSelectTimeout: config.Duration(time.Duration(5) * time.Second),
				TTL:                 config.Duration(time.Duration(5) * time.Minute),
				ClientConfig: tls.ClientConfig{
					TLSKey:             pki.ClientCertAndKeyPath(),
					InsecureSkipVerify: false,
				},
			},
			connErrFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "fail with x509 authentication using encrypted key file",
			plugin: &MongoDB{
				Dsn: fmt.Sprintf("mongodb://%s:%s",
					container.Address, container.Ports[servicePort]),
				AuthenticationType:  "X509",
				MetricDatabase:      "telegraf_test",
				MetricGranularity:   "seconds",
				ServerSelectTimeout: config.Duration(time.Duration(5) * time.Second),
				TTL:                 config.Duration(time.Duration(5) * time.Minute),
				ClientConfig: tls.ClientConfig{
					TLSCA:              pki.CACertPath(),
					TLSKey:             pki.ClientCertAndEncKeyPath(),
					TLSKeyPwd:          "badpassword",
					InsecureSkipVerify: false,
				},
			},
			connErrFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "fail with x509 authentication using invalid ca",
			plugin: &MongoDB{
				Dsn: fmt.Sprintf("mongodb://%s:%s",
					container.Address, container.Ports[servicePort]),
				AuthenticationType:  "X509",
				MetricDatabase:      "telegraf_test",
				MetricGranularity:   "seconds",
				ServerSelectTimeout: config.Duration(time.Duration(5) * time.Second),
				TTL:                 config.Duration(time.Duration(5) * time.Minute),
				ClientConfig: tls.ClientConfig{
					TLSCA:              pki.ClientCertAndKeyPath(),
					TLSKey:             pki.ClientCertAndKeyPath(),
					InsecureSkipVerify: false,
				},
			},
			connErrFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "fail with x509 authentication using invalid key",
			plugin: &MongoDB{
				Dsn: fmt.Sprintf("mongodb://%s:%s",
					container.Address, container.Ports[servicePort]),
				AuthenticationType:  "X509",
				MetricDatabase:      "telegraf_test",
				MetricGranularity:   "seconds",
				ServerSelectTimeout: config.Duration(time.Duration(5) * time.Second),
				TTL:                 config.Duration(time.Duration(5) * time.Minute),
				ClientConfig: tls.ClientConfig{
					TLSCA:              pki.CACertPath(),
					TLSKey:             pki.CACertPath(),
					InsecureSkipVerify: false,
				},
			},
			connErrFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// validate config
			err := tt.plugin.Init()
			require.NoError(t, err)

			if err == nil {
				// connect
				err = tt.plugin.Connect()
				tt.connErrFunc(t, err)

				if err == nil {
					// insert mock metrics
					err = tt.plugin.Write(testutil.MockMetrics())
					require.NoError(t, err)

					// cleanup
					err = tt.plugin.Close()
					require.NoError(t, err)
				}
			}
		})
	}
}

func TestConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		plugin  *MongoDB
		errFunc func(t *testing.T, err error)
	}{
		{
			name: "fail with invalid connection string",
			plugin: &MongoDB{
				Dsn:                "asdf1234",
				AuthenticationType: "NONE",
				MetricDatabase:     "telegraf_test",
				MetricGranularity:  "seconds",
				TTL:                config.Duration(time.Duration(5) * time.Minute),
			},
		},
		{
			name: "fail with invalid metric granularity",
			plugin: &MongoDB{
				Dsn:                "mongodb://localhost:27017",
				AuthenticationType: "NONE",
				MetricDatabase:     "telegraf_test",
				MetricGranularity:  "somerandomgranularitythatdoesntwork",
			},
		},
		{
			name: "fail with scram authentication missing username field",
			plugin: &MongoDB{
				Dsn:                "mongodb://localhost:27017",
				AuthenticationType: "SCRAM",
				Password:           "somerandompasswordthatwontwork",
				MetricDatabase:     "telegraf_test",
				MetricGranularity:  "seconds",
			},
		},
		{
			name: "fail with scram authentication missing password field",
			plugin: &MongoDB{
				Dsn:                "mongodb://localhost:27017",
				AuthenticationType: "SCRAM",
				Username:           "somerandomusernamethatwontwork",
				MetricDatabase:     "telegraf_test",
				MetricGranularity:  "seconds",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// validate config
			err := tt.plugin.Init()
			require.Error(t, err)
		})
	}

	tests = []struct {
		name    string
		plugin  *MongoDB
		errFunc func(t *testing.T, err error)
	}{
		{
			name: "success init with missing metric database",
			plugin: &MongoDB{
				Dsn:                "mongodb://localhost:27017",
				AuthenticationType: "NONE",
				MetricGranularity:  "seconds",
			},
		},
		{
			name: "success init missing metric granularity",
			plugin: &MongoDB{
				Dsn:                "mongodb://localhost:27017",
				AuthenticationType: "NONE",
				MetricDatabase:     "telegraf_test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// validate config
			err := tt.plugin.Init()
			require.NoError(t, err)
		})
	}
}
