package qsc_qsys

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf/testutil"
	"net"
	"testing"
)

const statusGetFormat = `{
	"jsonrpc": "2.0",
    "result": {
        "Platform": "Core 110f",
        "State": "Active",
        "DesignName": "test",
        "DesignCode": "YfSbX5YGvGyb",
        "IsRedundant": false,
        "IsEmulator": false,
        "Status": {
            "Code": 5,
            "String": "Initializing - 3 OK, 1 Initializing"
        }
    },
    "id": %d
}`
const controlGetFormat = `{
"jsonrpc": "2.0",
"result": [
    {
        "Name": "CoreProcessorTemperature",
        "String": "64.0Â°C",
        "Value": 64.0,
        "Position": 0.456
    },
    {
        "Name": "CoreChassisTemperature",
        "String": "49.0Â°C",
        "Value": 49.0,
        "Position": 0.39599999
    },
    {
        "Name": "TSCStatus",
        "String": "OK",
        "Value": 0.0,
        "Position": 0.0,
        "Color": "green"
    },
    {
        "Name": "TSCBacklight",
        "String": "80.0%%",
        "Value": 80.0,
        "Position": 0.80000001,
        "Indeterminate": false
    }
],
"id": %d
}`

func TestGather(t *testing.T) {
	mockServer, listenErr := net.Listen("tcp", "127.0.0.1:0")
	if listenErr != nil {
		t.Fatal("Initializing test server failed")
	}
	defer mockServer.Close()

	go handleRequests(mockServer, t)
	testClient := QSC_QSYS{
		Server:        mockServer.Addr().String(),
		NamedControls: []string{"GainGain"},
	}

	var acc testutil.Accumulator
	err := testClient.Gather(&acc)

	if err != nil {
		t.Fatalf("Gather returned error. Error: %s\n", err)
	}

	fields := map[string]interface{}{
		"CoreProcessorTemperature": float64(64),
		"CoreChassisTemperature":   float64(49),
		"TSCStatus":                float64(0),
		"TSCBacklight":             float64(80),
		"state":                    "Active",
		"status":                   "Initializing - 3 OK, 1 Initializing",
	}

	acc.AssertContainsFields(t, "qsys", fields)
}

func handleRequests(sock net.Listener, t *testing.T) {
	conn, err := sock.Accept()
	if err != nil {
		t.Fatal("Error accepting test connection")
	}
	for {
		requestString, err := bufio.NewReader(conn).ReadBytes(0)
		if err != nil {
			return
		}
		// Strip the trailing nullchar
		unterminatedRequestString := requestString[:len(requestString)-1]

		var request JSONRPC
		jsonerr := json.Unmarshal(unterminatedRequestString, &request)
		if jsonerr != nil {
			fmt.Printf("Error unmarshaling JSON: %s", jsonerr)
		} else {
			switch request.Method {
			case "StatusGet":
				conn.Write([]byte(fmt.Sprintf(statusGetFormat, request.ID) + "\x00"))
			case "Control.Get":
				conn.Write([]byte(fmt.Sprintf(controlGetFormat, request.ID) + "\x00"))
			}
		}
	}
}
