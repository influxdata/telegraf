package huawei_routers_telemetry

import (
	"fmt"
	"io"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
  "unicode"

	"github.com/DamRCorba/huawei_telemetry_sensors"
	"github.com/DamRCorba/huawei_telemetry_sensors/sensors/huawei-telemetry"

	"github.com/golang/protobuf/proto"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type setReadBufferer interface {
	SetReadBuffer(bytes int) error
}

type HuaweiRoutersTelemetry struct {
	ServicePort     string        `toml:"service_port"`
	ReadBufferSize  internal.Size `toml:"read_buffer_size"`
	ContentEncoding string        `toml:"content_encoding"`
	Log telegraf.Logger `toml:"-"`
	connection net.PacketConn
	decoder internal.ContentDecoder

	wg              sync.WaitGroup

	acc telegraf.Accumulator
	io.Closer
}

/*
  Telemetry Decoder.

*/
func HuaweiTelemetryDecoder(body []byte, h *HuaweiRoutersTelemetry) (*metric.SeriesGrouper, error) {
	msg := &telemetry.Telemetry{}
	err := proto.Unmarshal(body[12:], msg)
	if err != nil {
		h.Log.Error("Unable to decode incoming packet:  %v", err)		
		return nil, err		
	}
	grouper := metric.NewSeriesGrouper()
	for _, gpbkv := range msg.GetDataGpb().GetRow() {
		dataTime := gpbkv.Timestamp
		if dataTime == 0 {
			dataTime = msg.MsgTimestamp
		}
		timestamp := time.Unix(0, int64(dataTime)*1000000)
		sensorMsg := huawei_sensorPath.GetMessageType(msg.GetSensorPath())
		err = proto.Unmarshal(gpbkv.Content, sensorMsg)
		if err != nil {
			h.Log.Error("Sensor Error:  %v", err)			
			return nil, err
		}
		fields, vals := huawei_sensorPath.SearchKey(gpbkv, msg.GetSensorPath())
		tags := make(map[string]string, len(fields)+3)
		tags["source"] = msg.GetNodeIdStr()
		tags["subscription"] = msg.GetSubscriptionIdStr()
		tags["path"] = msg.GetSensorPath()
		// Search for Tags
		for i := 0; i < len(fields); i++ {
			tags = huawei_sensorPath.AppendTags(fields[i], vals[i], tags, msg.GetSensorPath())
		}
		// Create Metrics
		for i := 0; i < len(fields); i++ {
			CreateMetrics(grouper, tags, timestamp, msg.GetSensorPath(), fields[i], vals[i])
		}
	}
	return grouper, nil
}

/*
  Listen UDP packets and call the telemetryDecoder.
*/
func (h *HuaweiRoutersTelemetry) listen() {
	buf := make([]byte, 64*1024) // 64kb - maximum size of IP packet
	for {
		n, _, err := h.connection.ReadFrom(buf)
		if err != nil {
			h.Log.Error("Unable to read buffer: %v", err)
			break
		}

		body, err := h.decoder.Decode(buf[:n])
		if err != nil {
			h.Log.Errorf("Unable to decode incoming packet: %v", err)
			continue
		}
		// Telemetry parsing over packet payload
		grouper, err := HuaweiTelemetryDecoder(body,h)
		if err != nil {
			h.Log.Errorf("Unable to decode telemetry information: %v", err)
			break
		}
		for _, metric := range grouper.Metrics() {
			h.acc.AddMetric(metric)
		}

		if err != nil {
			h.Log.Errorf("Unable to parse incoming packet: %v", err)
		}
	}
}


func (h *HuaweiRoutersTelemetry) Description() string {
	return "Input plugin for receiving Huawei Router Telemetry data via UDP"
}

func (h *HuaweiRoutersTelemetry) SampleConfig() string {
	return `
  ## UDP Service Port to capture Telemetry
  # service_port = "8080"

`
}

func (h *HuaweiRoutersTelemetry) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (h *HuaweiRoutersTelemetry) Start(acc telegraf.Accumulator) error {
	h.acc = acc

	var err error
	h.decoder, err = internal.NewContentDecoder(h.ContentEncoding)
	if err != nil {
		return err
	}

	pc, err := udpListen(":"+h.ServicePort)
	if err != nil {
		return err
	}

	if h.ReadBufferSize.Size > 0 {
		if srb, ok := pc.(setReadBufferer); ok {
			srb.SetReadBuffer(int(h.ReadBufferSize.Size))
		} else {
			h.Log.Warnf("Unable to set read buffer on a %s socket", "udp")
		}
	}

	h.Log.Infof("Listening Routers on port %s", pc.LocalAddr())
	h.connection = pc

	
	h.wg = sync.WaitGroup{}
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.listen()
	}()
	return nil
}

/*
  Creates and add metrics from json mapped data in telegraf metrics SeriesGrouper
  @params:
  - grouper (*metric.SeriesGrouper) - pointer of metric series to append data.
  - tags (map[string]string) json data mapped
  - timestamp (time.Time) -
  - path (string) - sensor path
  - subfield (string) - subkey data.
    vals (string) - subkey content

*/
func CreateMetrics(grouper *metric.SeriesGrouper, tags map[string]string, timestamp time.Time, path string, subfield string, vals string)  {
  if subfield == "ifAdminStatus" {
    name:= strings.Replace(subfield,"\"","",-1)
    if vals == "IfAdminStatus_UP" {
      grouper.Add(path, tags, timestamp, string(name), 1)
    } else {
      grouper.Add(path, tags, timestamp, string(name), 0)
    }
  }
  if subfield == "ifOperStatus" {
    name:= strings.Replace(subfield,"\"","",-1)
    if vals == "IfOperStatus_UP" {
      grouper.Add(path, tags, timestamp, string(name), 1)
    } else {
      grouper.Add(path, tags, timestamp, string(name), 0)
    }
  }
  if vals != "" && subfield != "ifName" && subfield != "position" && subfield != "pemIndex" && subfield != "address" && subfield != "i2c" && subfield != "channel" &&
  subfield != "queueType" && subfield != "ifAdminStatus" && subfield != "ifOperStatus" {
    name:= strings.Replace(subfield,"\"","",-1)
    endPointTypes:=huawei_sensorPath.GetTypeValue(path)
    grouper.Add(path, tags, timestamp, string(name), decodeVal(endPointTypes[name], vals))
  }
}

/*
  Append to the tags the telemetry values for position.
  @params:
  k - Key to evaluate
  v - Content of the Key
  tags - Global tags of the metric
  path - Telemetry path
  @returns
  original tag append the key if its a name Key.

*/
func AppendTags(k string, v string, tags map[string]string, path string) map[string]string {
  resolve := tags
  endPointTypes:=huawei_sensorPath.GetTypeValue(path)
  if endPointTypes[k] != nil {
    if reflect.TypeOf(decodeVal(endPointTypes[k], v)) == reflect.TypeOf("") {
      if k != "ifAdminStatus" {
          resolve[k] = v
      }
    }
  } else {
    if k == "ifName" || k == "position" || k == "pemIndex" || k == "i2c"{
      resolve[k] = v
    }

  }
  return resolve
}

/*
  Convert the telemetry Data to its type.
  @Params:
  tipo - telemetry path data type
  val - string value
  Returns the converted value
*/
func decodeVal(tipo interface{}, val string) interface{} {
  if tipo == nil {
    return val
  } else {
  value := reflect.New(tipo.(reflect.Type)).Elem().Interface()
  switch value.(type) {
  case uint32: resolve, _ := strconv.ParseUint(val,10,32); return resolve;
  case uint64: resolve,_ :=  strconv.ParseUint(val,10,64); return resolve;
  case int32: resolve,_ :=  strconv.ParseInt(val,10,32);   return resolve;
  case int64: resolve,_ :=  strconv.ParseInt(val,10,64);   return resolve;
  case float64: resolve, err :=  strconv.ParseFloat(val,64);
                if err != nil {
                  name:= strings.Replace(val,"\"","",-1)
                  resolve, _=  strconv.ParseFloat(name,64);
                }
                return resolve;
  case bool: resolve,_ :=  strconv.ParseBool(val); return resolve;
  }
  }
  resolve := val;
  return resolve;
}


/* Search for a string in a string array.
  @Params: a String Array
           x String to Search
  @Returns: Returns the index location de x in a or -1 if not Found
*/
func Find(a []string, x string) int {
    for i, n := range a {
        if x == n {
            return i
        }
    }
    return -1
}

/*
  Search de keys and vals of the data row in telemetry message.
  @params:
  - Message (*TelemetryRowGPB) - data buffer GPB of sensor data
  - sensorType (string) - sensor-path group.
  @returns:
  - keys (string) - Keys of the fields
  - vals (string) - Vals of the fields
*/
func SearchKey(Message *telemetry.TelemetryRowGPB, path string)  ([]string, []string){
  sensorType := strings.Split(path,":")[0]
  sensorMsg := huawei_sensorPath.GetMessageType(sensorType)
  err := proto.Unmarshal(Message.Content, sensorMsg)
  if (err != nil) {
    panic(err)
  }
  primero := reflect.ValueOf(sensorMsg).Interface()

  str := fmt.Sprintf("%v", primero)
  // format string to JsonString with some modifications.
  jsonString := strings.Replace(str,"<>", "0",-1)
  jsonString = strings.Replace(jsonString,"<", "{\"",-1)
  jsonString= strings.Replace(jsonString,">", "\"}",-1)
  jsonString= strings.Replace(jsonString," ", ",\"",-1)
  jsonString= strings.Replace(jsonString,":", "\":",-1)
  jsonString= strings.Replace(jsonString,",\"\"","",-1)
  jsonString= strings.Replace(jsonString,"},\"", "}",-1)
  jsonString= strings.Replace(jsonString,","," ",-1)
  jsonString= strings.Replace(jsonString,"{"," ",-1)
  jsonString= strings.Replace(jsonString,"}","",-1)
  jsonString="\""+jsonString
  if path == "huawei-ifm:ifm/interfaces/interface/ifDynamicInfo" { // Particular case.....
    jsonString= strings.Replace(jsonString,"IfOperStatus_UPifName\"","IfOperStatus_UP \"ifName\"",-1)
  }
  lastQuote := rune(0)
      f := func(c rune) bool {
          switch {
          case c == lastQuote:
              lastQuote = rune(0)
              return false
          case lastQuote != rune(0):
              return false
          case unicode.In(c, unicode.Quotation_Mark):
              lastQuote = c
              return false
          default:
              return unicode.IsSpace(c)

          }
      }

    // splitting string by space but considering quoted section
    items := strings.FieldsFunc(jsonString, f)

    // create and fill the map
    m := make(map[string]string)
    for _, item := range items {
        x := strings.Split(item, ":")
        m[x[0]] = x[1]
    }
    // get keys and vals of fields
    var keys []string
    var vals []string
    for k, v := range m {
        name:= strings.Replace(k,"\"","",-1) // remove quotes
        keys = append(keys, name)
        vals = append(vals, v)

    }
    // Adaptation to resolve Huawei bad struct Data.
    if path == "huawei-ifm:ifm/interfaces/interface" {
      if Find(keys, "ifAdminStatus") == -1 {
        keys = append(keys, "ifAdminStatus")
        vals = append(vals, "IfAdminStatus_DOWN")
      }
    }
    // Adaptation to resolve Huawei bad struct Data.
    if path == "huawei-ifm:ifm/interfaces/interface/ifDynamicInfo" {
      if Find(keys, "ifOperStatus") == -1 {
        keys = append(keys, "ifOperStatus")
        vals = append(vals, "IfOperStatus_DOWN")
      }
    }

  return keys, vals
}


func udpListen(address string) (net.PacketConn, error) {	
	var addr *net.UDPAddr
	var err error
	var ifi *net.Interface
	addr, err = net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}
	if addr.IP.IsMulticast() {
		return net.ListenMulticastUDP("udp", ifi, addr)
	}
	return net.ListenUDP("udp", addr)	
}

func (h *HuaweiRoutersTelemetry) Stop() {
	if h.connection != nil {
		h.connection.Close()		
	}
	h.wg.Wait()
}

func init() {
	inputs.Add("huawei_routers_telemetry", func() telegraf.Input { return &HuaweiRoutersTelemetry{} })
}
