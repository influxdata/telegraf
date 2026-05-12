package cisco_telemetry_mdt

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	mdtdialout "github.com/cisco-ie/nx-telemetry-proto/mdt_dialout"
	telemetry "github.com/cisco-ie/nx-telemetry-proto/telemetry_bis"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestTCPDialoutOverflow(t *testing.T) {
	// Setup plugin
	plugin := &CiscoTelemetryMDT{
		Transport:      "tcp",
		ServiceAddress: "127.0.0.1:0",
		MaxMsgSize:     1000,
		Log:            testutil.Logger{},
	}
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Send data exceeding the message length
	c := &client{
		transport: plugin.Transport,
		addr:      plugin.listener.Addr().String(),
	}
	require.NoError(t, c.connect(t.Context()))
	defer c.close()
	require.NoError(t, c.send(t.Context(), bytes.Repeat([]byte{0xd, 0xe, 0xa, 0xd}, 1024)))

	// Wait for the errors to be accumulated
	require.Eventually(t, func() bool {
		acc.Lock()
		defer acc.Unlock()
		return len(acc.Errors) > 0
	}, 3*time.Second, 100*time.Millisecond)

	acc.Lock()
	defer acc.Unlock()
	require.Len(t, acc.Errors, 1)
	require.ErrorContains(t, acc.Errors[0], "dialout packet too long")
}

func TestTCPDialoutMultiple(t *testing.T) {
	// Setup plugin and start
	plugin := &CiscoTelemetryMDT{
		Transport:      "tcp",
		ServiceAddress: "127.0.0.1:0",
		Aliases: map[string]string{
			"some":     "type:model/some/path",
			"parallel": "type:model/parallel/path",
			"other":    "type:model/other/path",
		},
		Log: testutil.Logger{},
	}
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	addr := plugin.listener.Addr().String()

	// Setup the root message
	msg := &telemetry.Telemetry{
		MsgTimestamp: 1543236572000,
		EncodingPath: "type:model/some/path",
		NodeId:       &telemetry.Telemetry_NodeIdStr{NodeIdStr: "hostname"},
		Subscription: &telemetry.Telemetry_SubscriptionIdStr{SubscriptionIdStr: "subscription"},
		DataGpbkv: []*telemetry.TelemetryField{
			{
				Fields: []*telemetry.TelemetryField{
					{
						Name: "keys",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "name",
								ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "str"},
							},
						},
					},
					{
						Name: "content",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "value",
								ValueByType: &telemetry.TelemetryField_Sint64Value{Sint64Value: -1},
							},
						},
					},
				},
			},
		},
	}

	// Mimic a first device
	c1 := &client{
		addr:      addr,
		transport: "tcp",
	}
	require.NoError(t, c1.connect(t.Context()))
	defer c1.close()
	data, err := proto.Marshal(msg)
	require.NoError(t, err)
	require.NoError(t, c1.send(t.Context(), data))

	// Mimic a second device
	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()

	// Send a valid message
	msg.EncodingPath = "type:model/parallel/path"
	data, err = proto.Marshal(msg)
	require.NoError(t, err)
	_, err = conn.Write(createTCPHeader(0, 0, 0, 0, uint32(len(data))))
	require.NoError(t, err)
	_, err = conn.Write(data)
	require.NoError(t, err)

	// Send a header with invalid flags, the server should close the connection
	_, err = conn.Write(createTCPHeader(0, 0, 0, 257, 0))
	require.NoError(t, err)
	_, err = conn.Read([]byte{0})
	require.ErrorIs(t, err, io.EOF)

	// Make the first device send some more data to ensure the previous error
	// did not influence the other connection
	msg.EncodingPath = "type:model/other/path"
	data, err = proto.Marshal(msg)
	require.NoError(t, err)
	require.NoError(t, c1.send(t.Context(), data))

	// Check that we get an error message for the invalid header send
	require.Eventually(t, func() bool {
		acc.Lock()
		defer acc.Unlock()
		return len(acc.Errors) > 0
	}, 3*time.Second, 100*time.Millisecond)

	acc.Lock()
	errs := make([]error, 0, len(acc.Errors))
	errs = append(errs, acc.Errors...)
	acc.Unlock()
	require.Len(t, errs, 1)
	require.ErrorContains(t, errs[0], "invalid dialout flags: 257")

	// Check the metrics
	expected := []telegraf.Metric{
		metric.New(
			"some",
			map[string]string{
				"path":         "type:model/some/path",
				"name":         "str",
				"source":       "hostname",
				"subscription": "subscription",
			},
			map[string]interface{}{"value": int64(-1)},
			time.Unix(0, 1543236572000000000),
		),
		metric.New(
			"parallel",
			map[string]string{
				"path":         "type:model/parallel/path",
				"name":         "str",
				"source":       "hostname",
				"subscription": "subscription",
			},
			map[string]interface{}{"value": int64(-1)},
			time.Unix(0, 1543236572000000000),
		),
		metric.New(
			"other",
			map[string]string{
				"path":         "type:model/other/path",
				"name":         "str",
				"source":       "hostname",
				"subscription": "subscription",
			},
			map[string]interface{}{"value": int64(-1)},
			time.Unix(0, 1543236572000000000),
		),
	}

	// Wait for the metrics to arrive
	require.Eventually(t, func() bool {
		return acc.NMetrics() >= uint64(len(expected))
	}, 3*time.Second, 100*time.Millisecond)

	// Check the metric nevertheless as we might get some metrics despite errors.
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
}

func TestGRPCDialoutError(t *testing.T) {
	// Setup plugin and start
	plugin := &CiscoTelemetryMDT{
		Transport:      "grpc",
		ServiceAddress: "127.0.0.1:0",
		Log:            testutil.Logger{},
	}
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	addr := plugin.listener.Addr().String()

	// Create a client and send an error message
	c := &client{
		addr:      addr,
		transport: plugin.Transport,
	}
	require.NoError(t, c.connect(t.Context()))
	defer c.close()
	require.NoError(t, c.sendGRPC(&mdtdialout.MdtDialoutArgs{Errors: "foobar"}))

	// Wait for the error message to appear
	require.Eventually(t, func() bool {
		acc.Lock()
		defer acc.Unlock()
		return len(acc.Errors) > 0
	}, 3*time.Second, 100*time.Millisecond)
	acc.Lock()
	defer acc.Unlock()
	require.Len(t, acc.Errors, 1)
	require.ErrorContains(t, acc.Errors[0], "error during GRPC dialout: foobar")
}

func TestGRPCDialoutMultiple(t *testing.T) {
	// Setup plugin and start
	plugin := &CiscoTelemetryMDT{
		Transport:      "grpc",
		ServiceAddress: "127.0.0.1:0",
		Aliases: map[string]string{
			"some":     "type:model/some/path",
			"parallel": "type:model/parallel/path",
			"other":    "type:model/other/path",
		},
		Log: testutil.Logger{},
	}
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	addr := plugin.listener.Addr().String()

	// Setup the root message
	msg := &telemetry.Telemetry{
		MsgTimestamp: 1543236572000,
		EncodingPath: "type:model/some/path",
		NodeId:       &telemetry.Telemetry_NodeIdStr{NodeIdStr: "hostname"},
		Subscription: &telemetry.Telemetry_SubscriptionIdStr{SubscriptionIdStr: "subscription"},
		DataGpbkv: []*telemetry.TelemetryField{
			{
				Fields: []*telemetry.TelemetryField{
					{
						Name: "keys",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "name",
								ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "str"},
							},
						},
					},
					{
						Name: "content",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "value",
								ValueByType: &telemetry.TelemetryField_Sint64Value{Sint64Value: -1},
							},
						},
					},
				},
			},
		},
	}

	// Mimic a first device
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := mdtdialout.NewGRPCMdtDialoutClient(conn)
	stream, err := client.MdtDialout(t.Context())
	require.NoError(t, err)

	// Send a valid message for the first device
	data, err := proto.Marshal(msg)
	require.NoError(t, err)
	require.NoError(t, stream.Send(&mdtdialout.MdtDialoutArgs{Data: data, ReqId: 456}))

	// Mimic a second device
	conn2, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client2 := mdtdialout.NewGRPCMdtDialoutClient(conn2)
	stream2, err := client2.MdtDialout(t.Context())
	require.NoError(t, err)

	// Send a valid message for the second device
	msg.EncodingPath = "type:model/parallel/path"
	data, err = proto.Marshal(msg)
	require.NoError(t, err)
	require.NoError(t, stream2.Send(&mdtdialout.MdtDialoutArgs{Data: data}))

	// Send an error message for the second device
	require.NoError(t, stream2.Send(&mdtdialout.MdtDialoutArgs{Errors: "testclose"}))
	_, err = stream2.Recv()
	require.ErrorIs(t, err, io.EOF)
	require.NoError(t, conn2.Close())

	// Send another valid message for the first device
	msg.EncodingPath = "type:model/other/path"
	data, err = proto.Marshal(msg)
	require.NoError(t, err)
	require.NoError(t, stream.Send(&mdtdialout.MdtDialoutArgs{Data: data}))

	// Send an error message for the first device
	require.NoError(t, stream.Send(&mdtdialout.MdtDialoutArgs{Errors: "testclose"}))
	_, err = stream.Recv()
	require.ErrorIs(t, err, io.EOF)

	// Wait for the errors to arrive
	require.Eventually(t, func() bool {
		acc.Lock()
		defer acc.Unlock()
		return len(acc.Errors) > 1
	}, 3*time.Second, 100*time.Millisecond)

	// Check the result
	acc.Lock()
	errs := make([]error, 0, len(acc.Errors))
	errs = append(errs, acc.Errors...)
	acc.Unlock()
	require.Len(t, errs, 2)
	require.ErrorContains(t, errs[0], "error during GRPC dialout: testclose")
	require.ErrorContains(t, errs[1], "error during GRPC dialout: testclose")

	// Check the metrics
	expected := []telegraf.Metric{
		metric.New(
			"some",
			map[string]string{
				"path":         "type:model/some/path",
				"name":         "str",
				"source":       "hostname",
				"subscription": "subscription",
			},
			map[string]interface{}{"value": int64(-1)},
			time.Unix(0, 1543236572000000000),
		),
		metric.New(
			"parallel",
			map[string]string{
				"path":         "type:model/parallel/path",
				"name":         "str",
				"source":       "hostname",
				"subscription": "subscription",
			},
			map[string]interface{}{"value": int64(-1)},
			time.Unix(0, 1543236572000000000),
		),
		metric.New(
			"other",
			map[string]string{
				"path":         "type:model/other/path",
				"name":         "str",
				"source":       "hostname",
				"subscription": "subscription",
			},
			map[string]interface{}{"value": int64(-1)},
			time.Unix(0, 1543236572000000000),
		),
	}
	// Wait for the metrics to arrive
	require.Eventually(t, func() bool {
		return acc.NMetrics() >= uint64(len(expected))
	}, 3*time.Second, 100*time.Millisecond)

	// Check the metric nevertheless as we might get some metrics despite errors.
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
}

func TestGRPCDialoutKeepalive(t *testing.T) {
	// Setup plugin and start
	plugin := &CiscoTelemetryMDT{
		Transport:      "grpc",
		ServiceAddress: "127.0.0.1:0",
		EnforcementPolicy: grpcEnforcementPolicy{
			PermitKeepaliveWithoutCalls: true,
			KeepaliveMinTime:            0,
		},
		Log: testutil.Logger{},
	}
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	addr := plugin.listener.Addr().String()

	// Setup the root message
	msg := &telemetry.Telemetry{
		MsgTimestamp: 1543236572000,
		EncodingPath: "type:model/some/path",
		NodeId:       &telemetry.Telemetry_NodeIdStr{NodeIdStr: "hostname"},
		Subscription: &telemetry.Telemetry_SubscriptionIdStr{SubscriptionIdStr: "subscription"},
		DataGpbkv: []*telemetry.TelemetryField{
			{
				Fields: []*telemetry.TelemetryField{
					{
						Name: "keys",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "name",
								ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "str"},
							},
						},
					},
					{
						Name: "content",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "value",
								ValueByType: &telemetry.TelemetryField_Sint64Value{Sint64Value: -1},
							},
						},
					},
				},
			},
		},
	}

	// Send a message
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := mdtdialout.NewGRPCMdtDialoutClient(conn)
	stream, err := client.MdtDialout(t.Context())
	require.NoError(t, err)

	data, err := proto.Marshal(msg)
	require.NoError(t, err)

	require.NoError(t, stream.Send(&mdtdialout.MdtDialoutArgs{Data: data, ReqId: 456}))
}

func TestCases(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Register the plugin
	inputs.Add("cisco_telemetry_mdt", func() telegraf.Input {
		return &CiscoTelemetryMDT{
			Transport:       "grpc",
			ServiceAddress:  "127.0.0.1:57000",
			SourceFieldName: "mdt_source",
		}
	})

	// Prepare the influx parser for expectations
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}
		testcasePath := filepath.Join("testcases", f.Name())
		configFilename := filepath.Join(testcasePath, "telegraf.conf")
		expectedFilename := filepath.Join(testcasePath, "expected.out")
		expectedErrorFilename := filepath.Join(testcasePath, "expected.err")

		t.Run(f.Name(), func(t *testing.T) {
			// Read the input packets
			packets, err := readInputData(testcasePath)
			require.NoError(t, err)

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

			// Configure and initialize the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			plugin := cfg.Inputs[0].Input.(*CiscoTelemetryMDT)
			plugin.ServiceAddress = "127.0.0.1:0"

			// Start the plugin
			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Send all data
			c := &client{
				transport: plugin.Transport,
				addr:      plugin.listener.Addr().String(),
			}
			require.NoError(t, c.connect(t.Context()))
			defer c.close()
			for i, payload := range packets {
				require.NoErrorf(t, c.send(t.Context(), payload), "sending packet %d", i)
			}

			// Wait for the errors to arrive
			require.Eventually(t, func() bool {
				acc.Lock()
				defer acc.Unlock()
				return len(acc.Errors) >= len(expectedErrors)
			}, 3*time.Second, 100*time.Millisecond)

			// Check the result
			var actualErrorMsgs []string
			acc.Lock()
			if len(acc.Errors) > 0 {
				for _, err := range acc.Errors {
					actualErrorMsgs = append(actualErrorMsgs, err.Error())
				}
			}
			acc.Unlock()
			require.ElementsMatch(t, actualErrorMsgs, expectedErrors)

			// Wait for the metrics to arrive
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected))
			}, 3*time.Second, 100*time.Millisecond)

			// Check the metric nevertheless as we might get some metrics despite errors.
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
		})
	}
}

// Internal

func readInputData(path string) ([][]byte, error) {
	// Read all payloads
	files, err := filepath.Glob(filepath.Join(path, "packet*.json"))
	if err != nil {
		return nil, fmt.Errorf("globbing failed: %w", err)
	}
	sort.Strings(files)
	data := make([][]byte, 0, len(files))
	for _, fn := range files {
		buf, err := os.ReadFile(fn)
		if err != nil {
			return nil, fmt.Errorf("reading %q failed: %w", fn, err)
		}

		var msg telemetry.Telemetry
		if err := protojson.Unmarshal(buf, &msg); err != nil {
			return nil, fmt.Errorf("decoding packet %q failed: %w", fn, err)
		}
		d, err := proto.Marshal(&msg)
		if err != nil {
			return nil, fmt.Errorf("reencoding packet %q failed: %w", fn, err)
		}
		data = append(data, d)
	}

	return data, nil
}

type client struct {
	addr      string
	transport string
	tlscfg    *tls.Config

	conn   net.Conn
	stream mdtdialout.GRPCMdtDialout_MdtDialoutClient
	reqid  atomic.Int64
}

func (c *client) connect(ctx context.Context) error {
	switch c.transport {
	case "grpc":
		var creds credentials.TransportCredentials

		// Setup the connection credentials
		if c.tlscfg == nil {
			creds = insecure.NewCredentials()
		} else {
			creds = credentials.NewTLS(c.tlscfg)
		}

		// Connect to the server
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(creds),
		}
		conn, err := grpc.NewClient(c.addr, opts...)
		if err != nil {
			return fmt.Errorf("dialing server %q failed: %w", c.addr, err)
		}
		conn.Connect()

		// Create a dial-out client
		client := mdtdialout.NewGRPCMdtDialoutClient(conn)
		stream, err := client.MdtDialout(ctx)
		if err != nil {
			return fmt.Errorf("creating dial-out stream failed: %w", err)
		}
		c.stream = stream
	case "tcp":
		conn, err := net.Dial("tcp", c.addr)
		if err != nil {
			return fmt.Errorf("connecting to %q failed: %w", c.addr, err)
		}
		c.conn = conn
	default:
		return fmt.Errorf("unknown transport %q", c.transport)
	}

	return nil
}

func (c *client) close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	if c.stream != nil {
		c.stream.CloseSend()
	}
}

func (c *client) send(ctx context.Context, payload []byte) error {
	switch c.transport {
	case "grpc":
		data := &mdtdialout.MdtDialoutArgs{
			ReqId: c.reqid.Add(1),
			Data:  payload,
		}
		if err := c.sendGRPC(data); err != nil {
			return fmt.Errorf("sending via GRPC failed: %w", err)
		}
	case "tcp":
		// TCP Dialout telemetry framing header
		var buf bytes.Buffer
		if _, err := buf.Write(createTCPHeader(0, 0, 0, 0, uint32(len(payload)))); err != nil {
			return fmt.Errorf("writing header failed: %w", err)
		}
		// Payload
		if _, err := buf.Write(payload); err != nil {
			return fmt.Errorf("writing payload failed: %w", err)
		}

		if err := c.sendTCP(ctx, buf.Bytes()); err != nil {
			return fmt.Errorf("sending via TCP failed: %w", err)
		}
	default:
		return fmt.Errorf("unknown transport %q", c.transport)
	}

	return nil
}

func (c *client) sendGRPC(msg *mdtdialout.MdtDialoutArgs) error {
	if err := c.stream.Send(msg); err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}

func createTCPHeader(mtype, encap, version, flags uint16, length uint32) []byte {
	buf := make([]byte, 0, 12)
	buf = binary.BigEndian.AppendUint16(buf, mtype)
	buf = binary.BigEndian.AppendUint16(buf, encap)
	buf = binary.BigEndian.AppendUint16(buf, version)
	buf = binary.BigEndian.AppendUint16(buf, flags)
	buf = binary.BigEndian.AppendUint32(buf, length)

	return buf
}

func (c *client) sendTCP(ctx context.Context, data []byte) error {
	errch := make(chan error)
	go func() {
		_, err := c.conn.Write(data)
		errch <- err
	}()

	select {
	case err := <-errch:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
