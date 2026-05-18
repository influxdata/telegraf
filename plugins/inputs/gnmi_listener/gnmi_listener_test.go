package gnmilistener

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	common_tls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/gnmi_listener/nokia"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestMutalTLSFail(t *testing.T) {
	// Setup plugin
	plugin := &GNMIListener{
		Address: "127.0.0.1:0",
		ServerConfig: common_tls.ServerConfig{
			TLSCert:           "../../../testutil/pki/servercert.pem",
			TLSKey:            "../../../testutil/pki/serverkey.pem",
			TLSAllowedCACerts: []string{"../../../testutil/pki/cacert.pem"},
		},
		Log: testutil.Logger{LogLevel: new(telegraf.Trace)},
	}
	require.NoError(t, plugin.Init())

	// Start the plugin
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Setup a client to mimic the device with incorrect credentials
	tmpDir := t.TempDir()
	require.NoError(t, generateCertClientAuth(tmpDir, time.Now().Add(-10*time.Minute), time.Now().Add(10*time.Minute)))
	clientFailTLS := common_tls.ClientConfig{
		TLSCA:   "../../../testutil/pki/cacert.pem",
		TLSCert: filepath.Join(tmpDir, "client.pem"),
		TLSKey:  filepath.Join(tmpDir, "client.key"),
	}
	dev, err := newDevice(plugin.addr, plugin.Protocol, &deviceConfig{ClientConfig: clientFailTLS})
	require.NoError(t, err)

	// Send the data
	_, err = dev.send(t.Context(), &gnmi.SubscribeResponse{
		Response: &gnmi.SubscribeResponse_Update{
			Update: &gnmi.Notification{
				Timestamp: 1673608605875353770,
				Update: []*gnmi.Update{
					{
						Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "test"}}},
						Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_IntVal{IntVal: 23}},
					},
				},
			},
		},
	})
	require.Error(t, err)

	plugin.Stop()
}

func TestCases(t *testing.T) {
	// Get all testcase directories
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Register the plugin
	inputs.Add("gnmi", func() telegraf.Input {
		return &GNMIListener{}
	})

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		t.Run(f.Name(), func(t *testing.T) {
			testcasePath := filepath.Join("testcases", f.Name())
			configFilename := filepath.Join(testcasePath, "telegraf.conf")
			inputFilename := filepath.Join(testcasePath, "responses.json")
			expectedFilename := filepath.Join(testcasePath, "expected.out")
			expectedErrorFilename := filepath.Join(testcasePath, "expected.err")
			clientConfigFilename := filepath.Join(testcasePath, "client.conf")

			// Load the input data
			buf, err := os.ReadFile(inputFilename)
			require.NoError(t, err)
			var entries []json.RawMessage
			require.NoError(t, json.Unmarshal(buf, &entries))
			responses := make([]*gnmi.SubscribeResponse, 0, len(entries))
			for _, entry := range entries {
				var r gnmi.SubscribeResponse
				require.NoError(t, protojson.Unmarshal(entry, &r))
				responses = append(responses, &r)
			}

			// Prepare the influx parser for expectations
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())

			// Read the expected output if any
			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				var err error
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}

			// Read the expected output if any
			var expectedErrors []string
			if _, err := os.Stat(expectedErrorFilename); err == nil {
				var err error
				expectedErrors, err = testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
				require.NotEmpty(t, expectedErrors)
			}

			// Load the configuration for the client simulating the device
			var clientCfg deviceConfig
			if _, err := os.Stat(clientConfigFilename); err == nil {
				buf, err := os.ReadFile(clientConfigFilename)
				require.NoError(t, err)
				require.NoError(t, toml.Unmarshal(buf, &clientCfg))
			}

			// Configure and setup the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			plugin := cfg.Inputs[0].Input.(*GNMIListener)
			plugin.Address = "127.0.0.1:0"
			plugin.Log = testutil.Logger{}

			// Start the plugin
			var acc testutil.Accumulator
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Setup a client to mimic the device
			dev, err := newDevice(plugin.addr, plugin.Protocol, &clientCfg)
			require.NoError(t, err)

			// Send the data
			for _, r := range responses {
				_, err := dev.send(t.Context(), r)
				require.NoError(t, err)
			}

			// Wait for the metrics to arrive
			require.Eventually(t,
				func() bool {
					return acc.NMetrics() >= uint64(len(expected))
				}, 15*time.Second, 100*time.Millisecond)
			plugin.Stop()

			// Check for errors
			require.Len(t, acc.Errors, len(expectedErrors))
			if len(acc.Errors) > 0 {
				actualErrorMsgs := make([]string, 0, len(acc.Errors))
				for _, err := range acc.Errors {
					actualErrorMsgs = append(actualErrorMsgs, err.Error())
				}
				require.ElementsMatch(t, actualErrorMsgs, expectedErrors)
			}

			// Check the metric nevertheless as we might get some metrics despite errors.
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
		})
	}
}

// Internal functionality

type deviceConfig struct {
	common_tls.ClientConfig
}

type device interface {
	send(context.Context, *gnmi.SubscribeResponse) (*gnmi.SubscribeRequest, error)
}

func newDevice(addr, protocol string, cfg *deviceConfig) (device, error) {
	if protocol == "nokia" {
		tlscfg, err := cfg.ClientConfig.TLSConfig()
		if err != nil {
			return nil, fmt.Errorf("creating client TLS failed: %w", err)
		}
		return &nokiaDevice{
			addr:   addr,
			tlscfg: tlscfg,
		}, nil
	}
	return nil, fmt.Errorf("unknown protocol %q", protocol)
}

type nokiaDevice struct {
	addr   string
	tlscfg *tls.Config
}

func (d *nokiaDevice) send(ctx context.Context, msg *gnmi.SubscribeResponse) (*gnmi.SubscribeRequest, error) {
	var creds credentials.TransportCredentials

	// Setup the connection credentials
	if d.tlscfg == nil {
		creds = insecure.NewCredentials()
	} else {
		creds = credentials.NewTLS(d.tlscfg)
	}

	// Connect to the server
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}
	conn, err := grpc.NewClient(d.addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("dialing server %q failed: %w", d.addr, err)
	}
	conn.Connect()

	// Create a nokia dial-out client
	client := nokia.NewDialoutTelemetryClient(conn)
	sctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	stream, err := client.Publish(sctx, grpc.WaitForReady(false))
	if err != nil {
		return nil, fmt.Errorf("creating Nokia dial-out stream failed: %w", err)
	}
	defer stream.CloseSend()

	if err := stream.Send(msg); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
		return nil, err
	}

	// Wait for the response
	resp, err := stream.Recv()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
		return nil, err
	}
	time.Sleep(time.Second)
	return resp, nil
}

func generateCertClientAuth(tmpDir string, start, end time.Time) error {
	// Create the CA certificate
	caPriv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("creating CA key failed: %w", err)
	}

	ca := &x509.Certificate{
		SerialNumber: big.NewInt(342350),
		Subject: pkix.Name{
			Organization: []string{"Testing Inc."},
			Country:      []string{"US"},
			CommonName:   "Root CA",
		},
		NotBefore:             start,
		NotAfter:              end,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPriv.PublicKey, caPriv)
	if err != nil {
		return fmt.Errorf("creating CA certificate failed: %w", err)
	}
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caBytes})

	// Write CA cert
	if err := os.WriteFile(filepath.Join(tmpDir, "ca.pem"), caPEM, 0600); err != nil {
		return fmt.Errorf("writing CA certificate failed: %w", err)
	}

	// Create a client certificate and sign it
	clientPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("creating client key failed: %w", err)
	}
	clientPrivDer, err := x509.MarshalPKCS8PrivateKey(clientPriv)
	if err != nil {
		return fmt.Errorf("marshalling client key to PKCS8 failed: %w", err)
	}
	clientPrivPem := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: clientPrivDer})

	// Writing client private key
	// Write client cert
	if err := os.WriteFile(filepath.Join(tmpDir, "client.key"), clientPrivPem, 0600); err != nil {
		return fmt.Errorf("writing client private key failed: %w", err)
	}

	client := &x509.Certificate{
		SerialNumber: big.NewInt(342352),
		Subject: pkix.Name{
			Organization: []string{"Testing Inc."},
			Country:      []string{"US"},
			CommonName:   "Client Malcom",
		},
		NotBefore:   start,
		NotAfter:    end,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}
	clientBytes, err := x509.CreateCertificate(rand.Reader, client, ca, &clientPriv.PublicKey, caPriv)
	if err != nil {
		return fmt.Errorf("creating client certificate failed: %w", err)
	}
	clientPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientBytes})

	// Write client cert
	if err := os.WriteFile(filepath.Join(tmpDir, "client.pem"), clientPEM, 0600); err != nil {
		return fmt.Errorf("writing client certificate failed: %w", err)
	}

	return nil
}
