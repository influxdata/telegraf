package nats

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/plugins/common/netmonk"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestConnectAndWriteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Prepare httptest endpoint for agent verification
	r := mux.NewRouter()
	r.HandleFunc("/public/controller/server/server-12345/verify", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"client_id" : "agent-12345",
			"message_broker":{
				"type": "nats",
				"address": ["nats://localhost:4222"]
			},
			"auth":{
				"is_enabled":true,
				"username": "netmonkbroker",
				"password": "netmonkbrokersecret247"
			},
			"tls":{
				"is_enabled":false,
				"ca":"ca",
				"access":"access",
				"key":"key"
			}
		}`))
	}).Methods("POST")
	httpTestServer := httptest.NewServer(r)
	defer httpTestServer.Close()

	natsConf, err := filepath.Abs("config/nats-server.conf")
	require.NoError(t, err)
	authConf, err := filepath.Abs("config/auth.conf")
	require.NoError(t, err)

	container := testutil.Container{
		Image:        "nats:alpine3.17",
		ExposedPorts: []string{"4222:4222"},
		BindMounts: map[string]string{
			"/etc/nats/nats-server.conf": natsConf,
			"/etc/nats/auth.conf":        authConf,
		},
		WaitingFor: wait.ForLog("Server is ready"),
	}
	err = container.Start()
	require.NoError(t, err, "failed to start container")
	defer container.Terminate()

	server := []string{fmt.Sprintf("nats://%s:%s", container.Address, container.Ports["4222"])}
	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())
	n := &NATS{
		Servers:    server,
		Name:       "telegraf",
		Subject:    "telegraf",
		serializer: serializer,
		Agent: netmonk.Agent{
			NetmonkHost:      httpTestServer.URL,
			NetmonkServerID:  "server-12345",
			NetmonkServerKey: "12345",
		},
	}

	// Verify that we can connect to the NATS daemon
	err = n.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the NATS daemon
	err = n.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
