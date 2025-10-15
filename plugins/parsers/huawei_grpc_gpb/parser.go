package huawei_grpc_gpb

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/logger"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
	telemetry "github.com/influxdata/telegraf/plugins/parsers/huawei_grpc_gpb/telemetry_proto"
	"google.golang.org/protobuf/encoding/protojson"

	"log"
	"strconv"

	"google.golang.org/protobuf/proto"
)

const (
	KeySeperator        = "." // A nested delimiter for Tag or Field
	MsgTimeStampKeyName = "timestamp"
	JsonMsgKeyName      = "data_str"
	GpbMsgKeyName       = "data_gpb"
	RowKeyName          = "row"
	TimeFormat          = "2006-01-02 15:04:05" // time.RFC3339
	SensorPathKey       = "sensor_path"
)

type Parser struct {
	Log telegraf.Logger
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	telem_header := &telemetry.HuaweiTelemetry{}
	p.Log.Debugf("telemetry header : %s \n", telem_header)
	errParse := proto.Unmarshal(buf, telem_header)
	if errParse != nil {
		return nil, errParse
	}
	dataGpb := telem_header.GetDataGpb()
	if dataGpb != nil {
		// get protoPath
		protoPath := telem_header.ProtoPath
		// trans telemetry header into map[string]interface{}
		headerMap, errToMap := protoMsgToMap(telem_header)
		if errToMap != nil {
			return nil, errToMap
		}
		rows := dataGpb.GetRow()
		var rowsInMaps []map[string]interface{}
		var rowMsgs []proto.Message
		// Service layer decoding
		for _, row := range rows {
			contentMsg, errGetType := telemetry.GetTypeByProtoPath(protoPath, telemetry.DEFAULT_VERSION)
			if errGetType != nil {
				p.Log.Errorf("E! [grpc parser] get type according protoPath: %v", errGetType)
				return nil, errGetType
			}
			errDecode := proto.Unmarshal(row.Content, contentMsg)

			rowMap, errToMap := protoMsgToMap(contentMsg)
			if errToMap != nil {
				return nil, errToMap
			}
			rowMap[MsgTimeStampKeyName] = row.Timestamp
			rowsInMaps = append(rowsInMaps, rowMap)
			rowMsgs = append(rowMsgs, contentMsg)
			if errDecode != nil {
				return nil, errDecode
			}
		}
		p.debugLog(telem_header, rowMsgs)
		metrics, err := p.flattenProtoMsg(headerMap, rowsInMaps, "")
		return metrics, err
	}
	return nil, nil
}

func (p *Parser) debugLog(header *telemetry.HuaweiTelemetry, rows []proto.Message) {
	headerStr, err := json.MarshalIndent(header, "", " ")
	if err == nil {
		p.Log.Debugf("==================================== data start msg_timestamp: %v================================\n", header.MsgTimestamp)
		p.Log.Debugf("header is : \n%s", headerStr)
	} else {
		p.Log.Debugf("error when log header")
	}
	p.Log.Debugf("rows are : \n")
	for _, row := range rows {
		rowStr, err := json.MarshalIndent(row, "", " ")
		if err == nil {
			p.Log.Debugf("%s", rowStr)
		} else {
			p.Log.Debugf("error when log rows")
		}
	}
	p.Log.Debugf("==================================== data end ================================\n")
}

// Convert the Proto Message to a Map
func protoMsgToMap(protoMsg proto.Message) (map[string]interface{}, error) {
	// trans proto.Message into map[string]interface{}]
	protoToJson := protojson.MarshalOptions{
		UseEnumNumbers:  false,
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}
	pb, errToJson := protoToJson.Marshal(protoMsg)
	if errToJson != nil {
		return nil, fmt.Errorf("[grpc parser] proto message decode to json: %v", errToJson)
	}
	var msgMap map[string]interface{}
	errToMap := json.Unmarshal(pb, &msgMap)
	if errToMap != nil {
		return nil, fmt.Errorf("[grpc parser] proto message decodec to json: %v", errToMap)
	}
	return msgMap, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	panic("implement me")
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	panic("implement me")
}

func New() (*Parser, error) {
	return &Parser{
		Log: logger.New("parsers", "huawei_grpc_gpb", ""),
	}, nil
}

func init() {
	parsers.Add("huawei_grpc_gpb",
		func(defaultMetricName string) telegraf.Parser {
			parser, _ := New()
			return parser
		},
	)
}

type KVStruct struct {
	Fields map[string]interface{}
}

func (kv *KVStruct) FullFlattenStruct(fieldname string,
	v interface{},
	convertString bool,
	convertBool bool) error {
	if kv.Fields == nil {
		kv.Fields = make(map[string]interface{})
	}
	switch t := v.(type) {
	case map[string]interface{}:
		for k, v := range t {
			fieldKey := k
			if fieldname != "" {
				fieldKey = fieldname + KeySeperator + fieldKey
			}
			err := kv.FullFlattenStruct(fieldKey, v, convertString, convertBool)
			if err != nil {
				return err
			}
		}
	case []interface{}:
		for i, v := range t {
			fieldKey := strconv.Itoa(i)
			if fieldname != "" {
				fieldKey = fieldname + KeySeperator + fieldKey
			}
			err := kv.FullFlattenStruct(fieldKey, v, convertString, convertBool)
			if err != nil {
				return nil
			}
		}
	case float64:
		kv.Fields[fieldname] = v.(float64)
	case float32:
		kv.Fields[fieldname] = v.(float32)
	case uint64:
		kv.Fields[fieldname] = v.(uint64)
	case uint32:
		kv.Fields[fieldname] = v.(uint32)
	case int64:
		kv.Fields[fieldname] = v.(int64)
	case int32:
		kv.Fields[fieldname] = v.(int32)
	case string:
		if convertString {
			kv.Fields[fieldname] = v.(string)
		} else {
			return nil
		}
	case bool:
		if convertBool {
			kv.Fields[fieldname] = v.(bool)
		} else {
			return nil
		}
	case nil:
		return nil
	default:
		return fmt.Errorf("key Value Flattener : got unexpected type %T with value %v (%s)", t, t, fieldname)

	}
	return nil
}

// check if the data is num and return as
func convertToNum(str string) (bool, int64) {
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return false, 0
	} else {
		return true, num
	}
}

func (p *Parser) flattenProtoMsg(telemetryHeader map[string]interface{}, rowsDecodec []map[string]interface{}, startFieldName string) ([]telegraf.Metric, error) {
	kvHeader := KVStruct{}
	errHeader := kvHeader.FullFlattenStruct("", telemetryHeader, true, true)
	if errHeader != nil {
		return nil, errHeader
	}

	//// debug start
	//p.Log.Debugf("D! -------------------------------------Header START-----------------------------------------\n")
	//for k, v := range kvHeader.Fields {
	//    p.Log.Debugf("D! k: %s, v: %v ", k, v)
	//}
	//p.Log.Debugf("D! ------------------------------------- Header END -----------------------------------------\n")
	//// debug end

	var metrics []telegraf.Metric
	// one row into one metric
	for _, rowDecodec := range rowsDecodec {
		kvWithRow := KVStruct{}
		errRows := kvWithRow.FullFlattenStruct(startFieldName, rowDecodec, true, true)
		if errRows != nil {
			return nil, errRows
		}
		fields, tm, errMerge := p.mergeMaps(kvHeader.Fields, kvWithRow.Fields)
		if errMerge != nil {
			return nil, errMerge
		}
		metric := metric.New(telemetryHeader[SensorPathKey].(string), nil, fields, tm)
		// if err != nil {
		// return nil, err
		// }
		// debug start
		//p.Log.Debugf("D! -------------------------------------Fields START time is %v-----------------------------------------\n", metric.Time())
		//for k, v := range metric.Fields() {
		//    p.Log.Debugf("k: %s, v: %v ", k, v)
		//}
		//p.Log.Debugf("D! ------------------------------------- Fields END -----------------------------------------\n")
		// debug end

		metrics = append(metrics, metric)
	}
	return metrics, nil
}

func (p *Parser) mergeMaps(maps ...map[string]interface{}) (map[string]interface{}, time.Time, error) {
	res := make(map[string]interface{})
	timestamp := time.Time{}
	for _, m := range maps {
		for k, v := range m {
			if strings.HasSuffix(k, "_time") || strings.HasSuffix(k, MsgTimeStampKeyName) {
				timeStruct, timeStr, errCal := calTimeByStamp(v)
				if errCal != nil {
					return nil, time.Time{}, fmt.Errorf("E! [grpc parser] when calculate time, key name is %s, time is %t, error is %v", k, v, errCal)
				}
				if k == MsgTimeStampKeyName {
					timestamp = timeStruct
					p.Log.Debugf("D! the row timestamp is %s\n", timestamp.Format(TimeFormat))
					continue
				}
				if timeStr != "" {
					res[k] = timeStr
					continue
				}
			}
			res[k] = v
		}
	}
	return res, timestamp, nil
}

// timestamp transfer into time
// ten bit timestamp with second, 13 bit timestamp with second
// time.Unix(s,ns)
func calTimeByStamp(v interface{}) (time.Time, string, error) {
	var sec int64
	var nsec int64
	switch v.(type) {
	case float64:
		vInFloat64 := v.(float64)
		if vInFloat64 < math.Pow10(11) {
			sec = int64(vInFloat64)
			nsec = 0
		}
		if vInFloat64 > math.Pow10(12) {
			sec = int64(vInFloat64 / 1000)
			nsec = (int64(vInFloat64) % 1000) * 1000 * 1000
		}
	case int64:
		vInInt64 := v.(int64)
		if float64(vInInt64) < math.Pow10(11) {
			sec = vInInt64
			nsec = 0
		}
		if float64(vInInt64) > math.Pow10(12) {
			sec = vInInt64 / 1000
			nsec = (vInInt64 % 1000) * 1000 * 1000
		}
	case uint64:
		vInUint64 := v.(uint64)
		if float64(vInUint64) < math.Pow10(11) {
			sec = int64(vInUint64)
			nsec = 0
		}
		if float64(vInUint64) > math.Pow10(12) {
			sec = int64(vInUint64 / 1000)
			nsec = int64((vInUint64 % 1000) * 1000 * 1000)
		}
	case string:
		if strings.Index(v.(string), ":") > -1 {
			return time.Time{}, v.(string), nil
		}
		timeInNum, errToNum := strconv.ParseUint(v.(string), 10, 64)
		if errToNum != nil {
			log.Printf("E! [grpc.parser.calTimeByStamp] v: %t , error : %v", v, errToNum)
		} else {
			if float64(timeInNum) < math.Pow10(11) {
				sec = int64(timeInNum)
				nsec = 0
			}
			if float64(timeInNum) > math.Pow10(12) {
				sec = int64(timeInNum / 1000)
				nsec = int64((timeInNum % 1000) * 1000 * 1000)
			}
		}
	}

	if sec == 0 {
		return time.Time{}, "", errors.New("calculate error")
	}
	time := time.Unix(sec, nsec)
	return time, time.Format(TimeFormat), nil
}
