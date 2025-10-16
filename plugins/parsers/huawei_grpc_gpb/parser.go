package huawei_grpc_gpb

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/logger"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
	telemetry "github.com/influxdata/telegraf/plugins/parsers/huawei_grpc_gpb/telemetry_proto"
)

const (
	// KeySeparator is a nested delimiter for Tag or Field
	KeySeparator = "."
	// MsgTimeStampKeyName is the key for timestamp
	MsgTimeStampKeyName = "timestamp"
	// JSONMsgKeyName is the key for JSON data
	JSONMsgKeyName = "data_str"
	// GPBMsgKeyName is the key for GPB data
	GPBMsgKeyName = "data_gpb"
	// RowKeyName is the key for row data
	RowKeyName = "row"
	// TimeFormat is the format for time (RFC3339)
	TimeFormat = "2006-01-02 15:04:05"
	// SensorPathKey is the key for sensor path
	SensorPathKey = "sensor_path"
)

type Parser struct {
	Log telegraf.Logger
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	telemHeader := &telemetry.HuaweiTelemetry{}
	p.Log.Debugf("telemetry header : %s \n", telemHeader)
	errParse := proto.Unmarshal(buf, telemHeader)
	if errParse != nil {
		return nil, errParse
	}
	dataGPB := telemHeader.GetDataGpb()
	if dataGPB != nil {
		// get protoPath
		protoPath := telemHeader.ProtoPath
		// trans telemetry header into map[string]interface{}
		headerMap, errToMap := protoMsgToMap(telemHeader)
		if errToMap != nil {
			return nil, errToMap
		}
		rows := dataGPB.GetRow()
		var rowsInMaps []map[string]interface{}
		var rowMsgs []proto.Message
		// Service layer decoding
		for _, row := range rows {
			contentMsg, errGetType := telemetry.GetTypeByProtoPath(protoPath, telemetry.DefaultVersion)
			if errGetType != nil {
				p.Log.Errorf("get type according to protoPath: %v", errGetType)
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
		p.debugLog(telemHeader, rowMsgs)
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
		p.Log.Debugf("error when logging header: %v", err)
	}
	p.Log.Debugf("rows are : \n")
	for _, row := range rows {
		rowStr, err := json.MarshalIndent(row, "", " ")
		if err == nil {
			p.Log.Debugf("%s", rowStr)
		} else {
			p.Log.Debugf("error when logging rows: %v", err)
		}
	}
	p.Log.Debugf("==================================== data end ================================\n")
}

// protoMsgToMap converts the Proto Message to a Map
func protoMsgToMap(protoMsg proto.Message) (map[string]interface{}, error) {
	// trans proto.Message into map[string]interface{}]
	protoToJSON := protojson.MarshalOptions{
		UseEnumNumbers:  false,
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}
	pb, errToJSON := protoToJSON.Marshal(protoMsg)
	if errToJSON != nil {
		return nil, fmt.Errorf("proto message decode to json: %w", errToJSON)
	}
	var msgMap map[string]interface{}
	errToMap := json.Unmarshal(pb, &msgMap)
	if errToMap != nil {
		return nil, fmt.Errorf("proto message decoded to json: %w", errToMap)
	}
	return msgMap, nil
}

func (*Parser) ParseLine(_ string) (telegraf.Metric, error) {
	return nil, errors.New("parseLineNotImplemented")
}

func (*Parser) SetDefaultTags(_ map[string]string) {
	// Not implemented
}

func New() (*Parser, error) {
	return &Parser{
		Log: logger.New("parsers", "huawei_grpc_gpb", ""),
	}, nil
}

func init() {
	parsers.Add("huawei_grpc_gpb",
		func(_ string) telegraf.Parser {
			parser, err := New()
			if err != nil {
				panic(err)
			}
			return parser
		},
	)
}

type KVStruct struct {
	Fields map[string]interface{}
}

// FullFlattenStruct flattens nested structures into a flat map
func (kv *KVStruct) FullFlattenStruct(fieldname string,
	v interface{},
	convertString, convertBool bool) error {
	if kv.Fields == nil {
		kv.Fields = make(map[string]interface{})
	}
	switch t := v.(type) {
	case map[string]interface{}:
		for k, v := range t {
			fieldKey := k
			if fieldname != "" {
				fieldKey = fieldname + KeySeparator + fieldKey
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
				fieldKey = fieldname + KeySeparator + fieldKey
			}
			if err := kv.FullFlattenStruct(fieldKey, v, convertString, convertBool); err != nil {
				return err
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
		if !convertString {
			return nil
		}
		kv.Fields[fieldname] = v.(string)
	case bool:
		if !convertBool {
			return nil
		}
		kv.Fields[fieldname] = v.(bool)
	case nil:
		return nil
	default:
		return fmt.Errorf("key value flattener: got unexpected type %T with value %v (%s)", t, t, fieldname)
	}
	return nil
}

// convertToNum checks if the data is a number and returns it
// Unused function commented out to pass linting
/*
func convertToNum(str string) (bool, int64) {
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return false, 0
	} else {
		return true, num
	}
}
*/

func (p *Parser) flattenProtoMsg(telemetryHeader map[string]interface{}, rowsDecodec []map[string]interface{},
	startFieldName string) ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 0, len(rowsDecodec))
	kvHeader := KVStruct{}
	errHeader := kvHeader.FullFlattenStruct("", telemetryHeader, true, true)
	if errHeader != nil {
		return nil, errHeader
	}

	// Remove noisy data_gpb content from header
	delete(kvHeader.Fields, GPBMsgKeyName)
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
		metricInstance := metric.New(telemetryHeader[SensorPathKey].(string), nil, fields, tm)
		metrics = append(metrics, metricInstance)
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
					return nil, time.Time{}, fmt.Errorf("when calculating time, key name is %s, time value is %v, error is %w", k, v, errCal)
				}
				if k == MsgTimeStampKeyName {
					timestamp = timeStruct
					p.Log.Debugf("the row timestamp is %s\n", timestamp.Format(TimeFormat))
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
	switch vTyped := v.(type) {
	case float64:
		vInFloat64 := vTyped
		if vInFloat64 < math.Pow10(11) {
			sec = int64(vInFloat64)
			nsec = 0
		}
		if vInFloat64 > math.Pow10(12) {
			sec = int64(vInFloat64 / 1000)
			nsec = (int64(vInFloat64) % 1000) * 1000 * 1000
		}
	case int64:
		vInInt64 := vTyped
		if float64(vInInt64) < math.Pow10(11) {
			sec = vInInt64
			nsec = 0
		}
		if float64(vInInt64) > math.Pow10(12) {
			sec = vInInt64 / 1000
			nsec = (vInInt64 % 1000) * 1000 * 1000
		}
	case uint64:
		vInUint64 := vTyped
		if float64(vInUint64) < math.Pow10(11) {
			sec = int64(vInUint64)
			nsec = 0
		}
		if float64(vInUint64) > math.Pow10(12) {
			sec = int64(vInUint64 / 1000)
			nsec = int64((vInUint64 % 1000) * 1000 * 1000)
		}
	case string:
		if strings.Contains(vTyped, ":") {
			return time.Time{}, vTyped, nil
		}
		timeInNum, errToNum := strconv.ParseUint(vTyped, 10, 64)
		if errToNum != nil {
			return time.Time{}, "", fmt.Errorf("failed to parse time: %w", errToNum)
		}
		if float64(timeInNum) < math.Pow10(11) {
			sec = int64(timeInNum)
			nsec = 0
		}
		if float64(timeInNum) > math.Pow10(12) {
			sec = int64(timeInNum / 1000)
			nsec = int64((timeInNum % 1000) * 1000 * 1000)
		}
	}

	if sec == 0 {
		return time.Time{}, "", errors.New("calculate error")
	}
	timeResult := time.Unix(sec, nsec)
	return timeResult, timeResult.Format(TimeFormat), nil
}
