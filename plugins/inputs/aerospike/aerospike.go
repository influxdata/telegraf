package aerospike

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"net"
	"strconv"
	"strings"
	"sync"
)

const (
	MSG_HEADER_SIZE = 8
	MSG_TYPE        = 1 // Info is 1
	MSG_VERSION     = 2
)

var (
	STATISTICS_COMMAND = []byte("statistics\n")
	NAMESPACES_COMMAND = []byte("namespaces\n")
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

type Aerospike struct {
	Servers []string
}

var sampleConfig = `
  ## Aerospike servers to connect to (with port)
  ## This plugin will query all namespaces the aerospike
  ## server has configured and get stats for them.
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
		return a.gatherServer("127.0.0.1:3000", acc)
	}

	var wg sync.WaitGroup

	var outerr error

	for _, server := range a.Servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			outerr = a.gatherServer(server, acc)
		}(server)
	}

	wg.Wait()
	return outerr
}

func (a *Aerospike) gatherServer(host string, acc telegraf.Accumulator) error {
	aerospikeInfo, err := getMap(STATISTICS_COMMAND, host)
	if err != nil {
		return fmt.Errorf("Aerospike info failed: %s", err)
	}
	readAerospikeStats(aerospikeInfo, acc, host, "")
	namespaces, err := getList(NAMESPACES_COMMAND, host)
	if err != nil {
		return fmt.Errorf("Aerospike namespace list failed: %s", err)
	}
	for ix := range namespaces {
		nsInfo, err := getMap([]byte("namespace/"+namespaces[ix]+"\n"), host)
		if err != nil {
			return fmt.Errorf("Aerospike namespace '%s' query failed: %s", namespaces[ix], err)
		}
		readAerospikeStats(nsInfo, acc, host, namespaces[ix])
	}
	return nil
}

func getMap(key []byte, host string) (map[string]string, error) {
	data, err := get(key, host)
	if err != nil {
		return nil, fmt.Errorf("Failed to get data: %s", err)
	}
	parsed, err := unmarshalMapInfo(data, string(key))
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal data: %s", err)
	}

	return parsed, nil
}

func getList(key []byte, host string) ([]string, error) {
	data, err := get(key, host)
	if err != nil {
		return nil, fmt.Errorf("Failed to get data: %s", err)
	}
	parsed, err := unmarshalListInfo(data, string(key))
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal data: %s", err)
	}

	return parsed, nil
}

func get(key []byte, host string) (map[string]string, error) {
	var err error
	var data map[string]string

	asInfo := &aerospikeInfoCommand{
		msg: &aerospikeMessage{
			aerospikeMessageHeader: aerospikeMessageHeader{
				Version: uint8(MSG_VERSION),
				Type:    uint8(MSG_TYPE),
				DataLen: msgLenToBytes(int64(len(key))),
			},
			Data: key,
		},
	}

	cmd := asInfo.msg.Serialize()
	addr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		return data, fmt.Errorf("Lookup failed for '%s': %s", host, err)
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return data, fmt.Errorf("Connection failed for '%s': %s", host, err)
	}
	defer conn.Close()

	_, err = conn.Write(cmd)
	if err != nil {
		return data, fmt.Errorf("Failed to send to '%s': %s", host, err)
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
		return data, fmt.Errorf("Failed to read from connection to '%s': %s", host, err)
	}

	data, err = asInfo.parseMultiResponse()
	if err != nil {
		return data, fmt.Errorf("Failed to parse response from '%s': %s", host, err)
	}

	return data, err
}

func readAerospikeStats(
	stats map[string]string,
	acc telegraf.Accumulator,
	host string,
	namespace string,
) {
	fields := make(map[string]interface{})
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
	acc.AddFields("aerospike", fields, tags)
}

func unmarshalMapInfo(infoMap map[string]string, key string) (map[string]string, error) {
	key = strings.TrimSuffix(key, "\n")
	res := map[string]string{}

	v, exists := infoMap[key]
	if !exists {
		return res, fmt.Errorf("Key '%s' missing from info", key)
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

func init() {
	inputs.Add("aerospike", func() telegraf.Input {
		return &Aerospike{}
	})
}
