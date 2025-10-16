package huawei_grpc_json

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/logger"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
)

const (
	// KeySeparator is a nested delimiter for Tag or Field
	KeySeparator = "."
	// MsgTimeStampKeyName is the key for timestamp
	MsgTimeStampKeyName = "timestamp"
	// JSONMsgKeyName is the key for JSON data
	JSONMsgKeyName = "data_str"
	// RowKeyName is the key for row data
	RowKeyName = "row"
	// TimeFormat is the format for time (RFC3339)
	TimeFormat = "2006-01-02 15:04:05"
	// SensorPathKey is the key for sensor path
	SensorPathKey = "sensor_path"
)

type Parser struct {
	// Unused fields commented out to pass linting
	// metricName   string
	// tagKeys      []string
	// stringFields filter.Filter
	// nameKey      string
	// query        string
	// timeKey      string
	// timeFormat   string
	// timezone     string
	// defaultTags  map[string]string
	// strict       bool
	Mark string
	Log  telegraf.Logger
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	// parse header firstly
	var msgMap map[string]interface{}
	errToMap := json.Unmarshal(buf, &msgMap)
	if errToMap != nil {
		return nil, fmt.Errorf("proto message decoded to map: %w", errToMap)
	}

	// parse row
	msgsInMaps := make([]map[string]interface{}, 0)
	rowsTemp := msgMap[JSONMsgKeyName].(map[string]interface{})[RowKeyName].([]interface{})
	for _, data := range rowsTemp {
		msgsInMaps = append(msgsInMaps, data.(map[string]interface{}))
	}
	// remove key : data_str
	delete(msgMap, JSONMsgKeyName)
	metrics, err := p.flattenProtoMsg(msgMap, msgsInMaps, "")
	return metrics, err
}

// debugLog logs the header and rows for debugging
// Unused function commented out to pass linting
/*
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
*/

func (*Parser) ParseLine(_ string) (telegraf.Metric, error) {
	return nil, errors.New("parseLineNotImplemented")
}

func (*Parser) SetDefaultTags(_ map[string]string) {
	// Not implemented
}

func New() (*Parser, error) {
	return &Parser{
		Log: logger.New("parsers", "huawei_grpc_json", ""),
	}, nil
}

func init() {
	parsers.Add("huawei_grpc_json",
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
		kv.Fields[fieldname] = t
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

func (p *Parser) flattenProtoMsg(telemetryHeader map[string]interface{}, rowsDecodec []map[string]interface{},
	startFieldName string) ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 0, len(rowsDecodec))
	kvHeader := KVStruct{}
	errHeader := kvHeader.FullFlattenStruct("", telemetryHeader, true, true)
	if errHeader != nil {
		return nil, errHeader
	}

	// one row one metric
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

// mergeMaps merges maps and extracts timestamp
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

// calTimeByStamp converts timestamp to time
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
