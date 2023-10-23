package questdb

import (
	"bufio"
	"encoding/json"
	"github.com/influxdata/telegraf/config"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func newSerializer(t *testing.T) influx.Serializer {
	serializer := influx.Serializer{UintSupport: false}
	err := serializer.Init()
	require.NoError(t, err)
	return serializer
}

func newQuestDB(addr string) *QuestDB {
	return &QuestDB{
		Address: addr,
	}
}

func newQuestDBAuth(addr string, user string, token string) *QuestDB {
	return &QuestDB{
		Address: addr,
		User:    user,
		Token:   config.NewSecret([]byte(token)),
	}
}

func TestQuestDB_tcp(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	sw := newQuestDB("tcp://" + listener.Addr().String())
	require.NoError(t, sw.Connect())

	lconn, err := listener.Accept()
	require.NoError(t, err)

	testQuestDBStream(t, sw, lconn)
}

func TestQuestDB_udp_not_supported(t *testing.T) {
	listener, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	sw := newQuestDB("udp://" + listener.LocalAddr().String())
	require.Error(t, sw.Connect())
}

func TestQuestDB_unix_unsupported(t *testing.T) {
	sw := newQuestDB("unix://whatever")
	require.Error(t, sw.Connect())
}

func TestQuestDB_unixgram_unsupported(t *testing.T) {
	sw := newQuestDB("unixgram://whatever")
	require.Error(t, sw.Connect())
}

func TestQuestDBNoAuthIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image:        "questdb/questdb",
		ExposedPorts: []string{"9009", "9000"},
		WaitingFor:   wait.ForLog("server-main enjoy").AsRegexp(),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start QuestDB container")
	defer container.Terminate()

	questdb := newQuestDB("tcp://" + container.Address + ":" + container.Ports["9009"])
	testWithQuestDB(t, questdb, &container)
}

func TestQuestDBAuthIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	auth, err := filepath.Abs("testdata/auth.txt")
	require.NoError(t, err)
	container := testutil.Container{
		Image:        "questdb/questdb",
		ExposedPorts: []string{"9009", "9000"},
		WaitingFor:   wait.ForLog("server-main enjoy").AsRegexp(),
		Env: map[string]string{
			"QDB_LINE_TCP_AUTH_DB_PATH": "conf/authDb.txt",
		},
		BindMounts: map[string]string{
			"/var/lib/questdb/conf/authDb.txt": auth,
		},
	}
	err = container.Start()
	require.NoError(t, err, "failed to start QuestDB container")
	defer container.Terminate()

	questdb := newQuestDBAuth("tcp://"+container.Address+":"+container.Ports["9009"], "testUser1", "UvuVb1USHGRRT08gEnwN2zGZrvM4MsLQ5brgF6SVkAw")
	testWithQuestDB(t, questdb, &container)
}

func testWithQuestDB(t *testing.T, questdb *QuestDB, container *testutil.Container) {
	err := questdb.Connect()
	require.NoError(t, err, "error connecting to QuestDB")
	defer questdb.Close()

	metrics := []telegraf.Metric{}
	metrics = append(metrics, testutil.TestMetric(1, "test"))
	metrics = append(metrics, testutil.TestMetric(2, "test"))

	err = questdb.Write(metrics)
	require.NoError(t, err, "error writing to QuestDB")

	response := queryTestTable(t, container)
	defer response.Body.Close()

	b, err := io.ReadAll(response.Body)
	require.NoError(t, err, "Error reading response body")

	var receivedObj map[string]interface{}
	err = json.Unmarshal(b, &receivedObj)
	require.NoError(t, err, "Error unmarshaling received JSON")

	expectedJSON := `{
		"query":"select * from test",
		"columns":[{"name":"tag1","type":"SYMBOL"},{"name":"value","type":"LONG"},{"name":"timestamp","type":"TIMESTAMP"}],
		"timestamp":2,
		"dataset":[["value1",1,"2009-11-10T23:00:00.000000Z"],["value1",2,"2009-11-10T23:00:00.000000Z"]],
		"count":2
	}`
	var expectedObj map[string]interface{}
	err = json.Unmarshal([]byte(expectedJSON), &expectedObj)
	require.NoError(t, err, "Error unmarshaling expected JSON")

	require.True(t, reflect.DeepEqual(receivedObj, expectedObj), "Received object does not match expected object")
}

func queryTestTable(t *testing.T, container *testutil.Container) *http.Response {
	client := http.Client{}

	for i := 0; i < 30; i++ {
		response, err := client.Get("http://" + container.Address + ":" + container.Ports["9000"] + "/exec?query=select%20*%20from%20test")
		if err != nil {
			t.Logf("QuestDB HTTP API error: %s", err)
			time.Sleep(1 * time.Second)
			continue
		}
		if response.StatusCode != 200 {
			t.Logf("QuestDB HTTP API error: %s", response.Status)
			time.Sleep(1 * time.Second)
			continue
		}
		require.NotNil(t, response)
		require.Equal(t, 200, response.StatusCode)
		return response
	}
	require.Fail(t, "QuestDB HTTP API did not become available")
	return nil
}

func testQuestDBStream(t *testing.T, sw *QuestDB, lconn net.Conn) {
	serializer := newSerializer(t)

	metrics := []telegraf.Metric{}
	metrics = append(metrics, testutil.TestMetric(1, "test"))
	mbs1out, _ := serializer.Serialize(metrics[0])
	mbs1out, _ = sw.encoder.Encode(mbs1out)
	metrics = append(metrics, testutil.TestMetric(2, "test"))
	mbs2out, _ := serializer.Serialize(metrics[1])
	mbs2out, _ = sw.encoder.Encode(mbs2out)

	err := sw.Write(metrics)
	require.NoError(t, err)

	scnr := bufio.NewScanner(lconn)
	require.True(t, scnr.Scan())
	mstr1in := scnr.Text() + "\n"
	require.True(t, scnr.Scan())
	mstr2in := scnr.Text() + "\n"

	require.Equal(t, string(mbs1out), mstr1in)
	require.Equal(t, string(mbs2out), mstr2in)
}

func TestQuestDB_Write_err(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	sw := newQuestDB("tcp://" + listener.Addr().String())
	require.NoError(t, sw.Connect())
	require.NoError(t, sw.Conn.(*net.TCPConn).SetReadBuffer(256))

	lconn, err := listener.Accept()
	require.NoError(t, err)
	err = lconn.(*net.TCPConn).SetWriteBuffer(256)
	require.NoError(t, err)

	metrics := []telegraf.Metric{testutil.TestMetric(1, "testerr")}

	// close the socket to generate an error
	err = lconn.Close()
	require.NoError(t, err)

	err = sw.Conn.Close()
	require.NoError(t, err)

	err = sw.Write(metrics)
	require.Error(t, err)
	require.Nil(t, sw.Conn)
}

func TestQuestDB_Write_reconnect(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	sw := newQuestDB("tcp://" + listener.Addr().String())
	require.NoError(t, sw.Connect())
	require.NoError(t, sw.Conn.(*net.TCPConn).SetReadBuffer(256))

	lconn, err := listener.Accept()
	require.NoError(t, err)
	err = lconn.(*net.TCPConn).SetWriteBuffer(256)
	require.NoError(t, err)

	err = lconn.Close()
	require.NoError(t, err)
	sw.Conn = nil

	wg := sync.WaitGroup{}
	wg.Add(1)
	var lerr error
	go func() {
		lconn, lerr = listener.Accept()
		wg.Done()
	}()

	metrics := []telegraf.Metric{testutil.TestMetric(1, "testerr")}
	err = sw.Write(metrics)
	require.NoError(t, err)

	wg.Wait()
	require.NoError(t, lerr)

	serializer := newSerializer(t)
	mbsout, _ := serializer.Serialize(metrics[0])
	buf := make([]byte, 256)
	n, err := lconn.Read(buf)
	require.NoError(t, err)
	require.Equal(t, string(mbsout), string(buf[:n]))
}
