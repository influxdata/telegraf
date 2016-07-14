package aerospike

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/jameskeane/bcrypt"
)

const (
	MSG_HEADER_SIZE = 8
	MSG_TYPE_INFO   = 1 // Info is 1
	MSG_TYPE_AUTH   = 2 //
	MSG_VERSION     = 2

	// Field IDs
	USER       byte = 0
	CREDENTIAL byte = 3

	// Commands
	AUTHENTICATE byte = 0

	//constants from aerospike doc
	ERR_PASSWORD         = 62
	ERR_USER             = 60
	ERR_NOT_ENABLED      = 52
	ERR_SCHEME           = 53
	ERR_EXPIRED_PASSWORD = 63
	ERR_NOT_SUPPORTED    = 51
)

var errorCode2Msg map[int]string

var (
	STATISTICS_COMMAND = []byte("statistics\n")
	NAMESPACES_COMMAND = []byte("namespaces\n")
	LATENCY_COMMAND    = []byte("latency:back=60;\n") //get latency of previous minute
)

type aerospikeMessageHeader struct {
	Version uint8
	Type    uint8
	DataLen [6]byte
}

type aerospikeMessage struct {
	aerospikeMessageHeader
	Data []byte
}

// Taken from aerospike-client-go/types/message.go
func (msg *aerospikeMessage) Serialize() []byte {
	msg.DataLen = msgLenToBytes(int64(len(msg.Data)))
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, msg.aerospikeMessageHeader)
	binary.Write(buf, binary.BigEndian, msg.Data[:])
	return buf.Bytes()
}

type aerospikeInfoCommand struct {
	msg *aerospikeMessage
}

// Taken from aerospike-client-go/info.go
func (nfo *aerospikeInfoCommand) parseMultiResponse() (map[string]string, error) {
	responses := make(map[string]string)
	offset := int64(0)
	begin := int64(0)

	dataLen := int64(len(nfo.msg.Data))

	// Create reusable StringBuilder for performance.
	for offset < dataLen {
		b := nfo.msg.Data[offset]

		if b == '\t' {
			name := nfo.msg.Data[begin:offset]
			offset++
			begin = offset

			// Parse field value.
			for offset < dataLen {
				if nfo.msg.Data[offset] == '\n' {
					break
				}
				offset++
			}

			if offset > begin {
				value := nfo.msg.Data[begin:offset]
				responses[string(name)] = string(value)
			} else {
				responses[string(name)] = ""
			}
			offset++
			begin = offset
		} else if b == '\n' {
			if offset > begin {
				name := nfo.msg.Data[begin:offset]
				responses[string(name)] = ""
			}
			offset++
			begin = offset
		} else {
			offset++
		}
	}

	if offset > begin {
		name := nfo.msg.Data[begin:offset]
		responses[string(name)] = ""
	}
	return responses, nil
}

//wrapper for field
type Field struct {
	size   int32
	typeId byte
	data   []byte
}

//creates new field
func newField(typeId byte, data []byte) *Field {
	return &Field{
		size:   int32(len(data) + 1),
		typeId: typeId,
		data:   data,
	}
}

//serializes field and writes it to buf
func (f *Field) WriteToBuf(buf *bytes.Buffer) {
	binary.Write(buf, binary.BigEndian, f.size)
	binary.Write(buf, binary.BigEndian, f.typeId)
	buf.Write(f.data)

}

type endPointInfo struct {
	endpoint    string
	authEnabled bool
	username    string
	password    string
}

type Aerospike struct {
	Servers []string
}

var sampleConfig = `
  ## Aerospike servers to connect to (with port), 
  ## if auth is enabled provide "user:password@host:port format"
  ## This plugin will query all namespaces the aerospike
  ## server has configured and get stats for them.
  #	servers = ["user:password@localhost:3000"]
  servers = ["localhost:3000"]
 `

func (a *Aerospike) SampleConfig() string {
	return sampleConfig
}

func (a *Aerospike) Description() string {
	return "Read stats from an aerospike server"
}

func (a *Aerospike) Gather(acc telegraf.Accumulator) error {

	if len(a.Servers) == 0 {
		epInfo := &endPointInfo{
			endpoint:    "localhost:3000",
			authEnabled: false,
		}

		return epInfo.gatherServer(acc)
	}

	var wg sync.WaitGroup

	var outerr error

	for _, server := range a.Servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			epInfo, err := parseHostInfo(server)
			if err != nil {
				outerr = err

			} else {
				outerr = epInfo.gatherServer(acc)
			}
		}(server)
	}

	wg.Wait()
	return outerr
}

func (e *endPointInfo) gatherServer(acc telegraf.Accumulator) error {

	aerospikeInfo, err := e.getMap(STATISTICS_COMMAND)
	if err != nil {
		return fmt.Errorf("Aerospike info failed: %s", err)
	}
	fields := make(map[string]interface{})
	readAerospikeStats(aerospikeInfo, fields, e.endpoint, "")

	latencyInfo, err := e.get(LATENCY_COMMAND)
	if err != nil {
		return fmt.Errorf("Latency info failed %s", err.Error())
	}

	tags := map[string]string{
		"aerospike_host": e.endpoint,
		"namespace":      "_service",
	}

	readAerospikeLatency(latencyInfo, fields, e.endpoint)

	acc.AddFields("aerospike", fields, tags)

	namespaces, err := e.getList(NAMESPACES_COMMAND)
	if err != nil {
		return fmt.Errorf("Aerospike namespace list failed: %s", err)
	}

	for ix := range namespaces {
		nsInfo, err := e.getMap([]byte("namespace/" + namespaces[ix] + "\n"))
		if err != nil {
			return fmt.Errorf("Aerospike namespace '%s' query failed: %s", namespaces[ix], err)
		}
		fields := make(map[string]interface{})
		readAerospikeStats(nsInfo, fields, e.endpoint, namespaces[ix])

		tags["namespace"] = namespaces[ix]
		acc.AddFields("aerospike", fields, tags)

	}
	return nil
}

func parseHostInfo(server string) (*endPointInfo, error) {

	indexOfAt := strings.LastIndex(server, "@")

	if indexOfAt == -1 {
		return &endPointInfo{
			endpoint:    server,
			authEnabled: false,
			username:    "",
			password:    "",
		}, nil
	}

	endpoint := server[indexOfAt+1:]

	indexOfColon := strings.Index(server[:indexOfAt-1], ":")
	if indexOfColon == -1 {
		return nil, fmt.Errorf("Can't find username,password in '%s'", server)
	}
	username := server[0:indexOfColon]
	password := server[indexOfColon+1 : indexOfAt]

	return &endPointInfo{
		endpoint:    endpoint,
		authEnabled: true,
		username:    username,
		password:    password,
	}, nil
}

func (e *endPointInfo) getMap(key []byte) (map[string]string, error) {
	data, err := e.get(key)
	if err != nil {
		return nil, fmt.Errorf("Failed to get data: %s", err)
	}
	parsed, err := unmarshalMapInfo(data, string(key))
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal data: %s", err)
	}

	return parsed, nil
}

func (e *endPointInfo) getList(key []byte) ([]string, error) {
	data, err := e.get(key)
	if err != nil {
		return nil, fmt.Errorf("Failed to get data: %s", err)
	}
	parsed, err := unmarshalListInfo(data, string(key))
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal data: %s", err)
	}

	return parsed, nil
}

func (e *endPointInfo) get(key []byte) (map[string]string, error) {
	var err error
	var data map[string]string

	asInfo := &aerospikeInfoCommand{
		msg: &aerospikeMessage{
			aerospikeMessageHeader: aerospikeMessageHeader{
				Version: uint8(MSG_VERSION),
				Type:    uint8(MSG_TYPE_INFO),
				DataLen: msgLenToBytes(int64(len(key))),
			},
			Data: key,
		},
	}

	cmd := asInfo.msg.Serialize()
	addr, err := net.ResolveTCPAddr("tcp", e.endpoint)
	if err != nil {
		return data, fmt.Errorf("Lookup failed for '%s': %s", e.endpoint, err)
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return data, fmt.Errorf("Connection failed for '%s': %s", e.endpoint, err)
	}
	defer conn.Close()

	if e.authEnabled {
		err = e.authenticate(conn)
		if err != nil {
			return data, err
		}
	}

	_, err = conn.Write(cmd)
	if err != nil {
		return data, fmt.Errorf("Failed to send to '%s': %s", e.endpoint, err)
	}

	msgHeader := bytes.NewBuffer(make([]byte, MSG_HEADER_SIZE))
	_, err = readLenFromConn(conn, msgHeader.Bytes(), MSG_HEADER_SIZE)
	if err != nil {
		return data, fmt.Errorf("Failed to read header: %s", err)
	}
	err = binary.Read(msgHeader, binary.BigEndian, &asInfo.msg.aerospikeMessageHeader)
	if err != nil {
		return data, fmt.Errorf("Failed to unmarshal header: %s", err)
	}

	msgLen := msgLenFromBytes(asInfo.msg.aerospikeMessageHeader.DataLen)

	if int64(len(asInfo.msg.Data)) != msgLen {
		asInfo.msg.Data = make([]byte, msgLen)
	}

	_, err = readLenFromConn(conn, asInfo.msg.Data, len(asInfo.msg.Data))
	if err != nil {
		return data, fmt.Errorf("Failed to read from connection to '%s': %s", e.endpoint, err)
	}

	data, err = asInfo.parseMultiResponse()
	if err != nil {
		return data, fmt.Errorf("Failed to parse response from '%s': %s", e.endpoint, err)
	}

	return data, err
}

func (e *endPointInfo) authenticate(conn *net.TCPConn) error {

	buf := bytes.NewBuffer([]byte{})
	header := make([]byte, 16)

	for i := 0; i < 16; i++ {
		header[i] = 0
	}
	header[2] = AUTHENTICATE
	header[3] = byte(2) //field count

	binary.Write(buf, binary.BigEndian, header)
	usernameField := newField(USER, []byte(e.username))
	usernameField.WriteToBuf(buf)

	pw, err := bcrypt.Hash(e.password, "$2a$10$7EqJtq98hPqEX7fNZaFWoO")
	if err != nil {
		return err
	}

	passwordField := newField(CREDENTIAL, []byte(pw))
	passwordField.WriteToBuf(buf)

	data := buf.Bytes()

	asInfo := &aerospikeInfoCommand{
		msg: &aerospikeMessage{
			aerospikeMessageHeader: aerospikeMessageHeader{
				Version: uint8(MSG_VERSION),
				Type:    uint8(MSG_TYPE_AUTH),
				DataLen: msgLenToBytes(int64(len(data))),
			},
			Data: data,
		},
	}

	cmd := asInfo.msg.Serialize()
	_, err = conn.Write(cmd)
	if err != nil {
		return err
	}

	msgHeaderData := bytes.NewBuffer(make([]byte, MSG_HEADER_SIZE))
	var msgHeader aerospikeMessageHeader

	_, err = readLenFromConn(conn, msgHeaderData.Bytes(), MSG_HEADER_SIZE)
	if err != nil {
		return fmt.Errorf("Failed to read header: %s", err)
	}
	err = binary.Read(msgHeaderData, binary.BigEndian, &msgHeader)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal header: %s", err)
	}

	msgLen := msgLenFromBytes(msgHeader.DataLen)
	buffer := make([]byte, msgLen)
	_, err = readLenFromConn(conn, buffer, int(msgLen))
	if err != nil {
		return fmt.Errorf("Failed to read from connection to '%s': ", err)
	}

	errorCode := int(buffer[1])

	val, exist := errorCode2Msg[errorCode]
	if exist {
		return fmt.Errorf("Authentication failed: %s", val)
	} else if errorCode != 0 {
		return fmt.Errorf("Authentication request failed with errorcode %d", errorCode)
	}

	return nil
}

func readAerospikeStats(
	stats map[string]string,
	fields map[string]interface{},
	host string,
	namespace string,
) {
	tags := map[string]string{
		"aerospike_host": host,
		"namespace":      "_service",
	}

	if namespace != "" {
		tags["namespace"] = namespace
	}
	for key, value := range stats {
		// We are going to ignore all string based keys
		val, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			if strings.Contains(key, "-") {
				key = strings.Replace(key, "-", "_", -1)
			}
			fields[key] = val
		}
	}
}

func readAerospikeLatency(
	stats map[string]string,
	fields map[string]interface{},
	host string,
) {

	key := strings.TrimSuffix(string(LATENCY_COMMAND), "\n")
	data := stats[key]

	splitted := strings.Split(data, ";")

	for i := 0; i < len(splitted)-1; i = i + 2 {
		ind := strings.Index(splitted[i], ":")
		if ind == -1 {
			continue
		}

		metrName := splitted[i][0:ind]
		spl1 := splitted[i][ind:]
		ind = strings.Index(spl1, ",")
		spl1 = spl1[ind+1:]
		unitTimes := strings.Split(spl1, ",")

		vals := strings.Split(splitted[i+1], ",")

		//fmt.Println("Got ", metrName, " dime ", unitTimes, " vals ", vals)

		for i := 1; i < len(unitTimes); i++ {
			metric := metrName + "_" + unitTimes[i] //strings.Replace(unitTimes[i], ">", "_gt_", 1)
			value, err := strconv.ParseFloat(vals[i], 64)
			if err != nil {
				fmt.Println("Failed to parse float when parsing latency ")
				continue
			}
			fields[metric] = value
		}

	}
	//fmt.Println("aerospike", "Tags", tags, "Fields", fields)

}
func unmarshalMapInfo(infoMap map[string]string, key string) (map[string]string, error) {
	key = strings.TrimSuffix(key, "\n")
	res := map[string]string{}

	v, exists := infoMap[key]
	if !exists {
		errString := ""
		for k, v := range infoMap {
			if strings.HasPrefix(k, "ERROR:") || strings.HasPrefix(k, "Error:") {
				errString = k + ": " + v
			}
		}
		if errString == "" {
			return res, fmt.Errorf("Key '%s' missing from info", key)
		}
		return res, fmt.Errorf("Key '%s' missing from info: Probable Error: '%s' ", key, errString)

	}

	values := strings.Split(v, ";")
	for i := range values {
		kv := strings.Split(values[i], "=")
		if len(kv) > 1 {
			res[kv[0]] = kv[1]
		}
	}

	return res, nil
}

func unmarshalListInfo(infoMap map[string]string, key string) ([]string, error) {
	key = strings.TrimSuffix(key, "\n")

	v, exists := infoMap[key]
	if !exists {
		return []string{}, fmt.Errorf("Key '%s' missing from info", key)
	}

	values := strings.Split(v, ";")
	return values, nil
}

func readLenFromConn(c net.Conn, buffer []byte, length int) (total int, err error) {
	var r int
	for total < length {
		r, err = c.Read(buffer[total:length])
		total += r
		if err != nil {
			break
		}
	}
	return
}

// Taken from aerospike-client-go/types/message.go
func msgLenToBytes(DataLen int64) [6]byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(DataLen))
	res := [6]byte{}
	copy(res[:], b[2:])
	return res
}

// Taken from aerospike-client-go/types/message.go
func msgLenFromBytes(buf [6]byte) int64 {
	nbytes := append([]byte{0, 0}, buf[:]...)
	DataLen := binary.BigEndian.Uint64(nbytes)
	return int64(DataLen)
}

func populateErrorCode2Msg() {
	errorCode2Msg = make(map[int]string)

	errorCode2Msg[ERR_USER] = "No user supplied or unknown user."
	errorCode2Msg[ERR_PASSWORD] = "Password does not exists or not recognized."
	errorCode2Msg[ERR_NOT_ENABLED] = "Security functionality not enabled by connected server."
	errorCode2Msg[ERR_SCHEME] = "Security scheme not supported."
	errorCode2Msg[ERR_EXPIRED_PASSWORD] = "Expired password."
	errorCode2Msg[ERR_NOT_SUPPORTED] = "Security functionality not supported by connected server."
}

func init() {

	populateErrorCode2Msg()

	inputs.Add("aerospike", func() telegraf.Input {
		return &Aerospike{}
	})
}
