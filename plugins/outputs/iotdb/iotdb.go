//go:generate ../../../tools/readme_config_includer/generator
package iotdb

import (
	_ "embed"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/apache/iotdb-client-go/client"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/outputs"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

type IoTDB struct {
	Host            string          `toml:"host"`
	Port            string          `toml:"port"`
	User            string          `toml:"user"`
	Password        string          `toml:"password"`
	Timeout         config.Duration `toml:"timeout"`
	ConvertUint64To string          `toml:"uint64_conversion"`
	TimeStampUnit   string          `toml:"timestamp_precision"`
	TreatTagsAs     string          `toml:"convert_tags_to"`
	Log             telegraf.Logger `toml:"-"`

	session *client.Session
}

type recordsWithTags struct {
	// IoTDB Records basic data struct
	DeviceIDList     []string
	MeasurementsList [][]string
	ValuesList       [][]interface{}
	DataTypesList    [][]client.TSDataType
	TimestampList    []int64
	// extra tags
	TagsList [][]*telegraf.Tag
}

func (*IoTDB) SampleConfig() string {
	return sampleConfig
}

// Init is for setup, and validating config.
func (s *IoTDB) Init() error {
	if s.Timeout < 0 {
		return errors.New("negative timeout")
	}
	if !choice.Contains(s.ConvertUint64To, []string{"int64", "int64_clip", "text"}) {
		return fmt.Errorf("unknown 'uint64_conversion' method %q", s.ConvertUint64To)
	}
	if !choice.Contains(s.TimeStampUnit, []string{"second", "millisecond", "microsecond", "nanosecond"}) {
		return fmt.Errorf("unknown 'timestamp_precision' method %q", s.TimeStampUnit)
	}
	if !choice.Contains(s.TreatTagsAs, []string{"fields", "device_id"}) {
		return fmt.Errorf("unknown 'convert_tags_to' method %q", s.TreatTagsAs)
	}
	s.Log.Info("Initialization completed.")
	return nil
}

func (s *IoTDB) Connect() error {
	sessionConf := &client.Config{
		Host:     s.Host,
		Port:     s.Port,
		UserName: s.User,
		Password: s.Password,
	}
	var ss = client.NewSession(sessionConf)
	s.session = &ss
	timeoutInMs := int(time.Duration(s.Timeout).Milliseconds())
	if err := s.session.Open(false, timeoutInMs); err != nil {
		return fmt.Errorf("connecting to %s:%s failed: %w", s.Host, s.Port, err)
	}
	return nil
}

func (s *IoTDB) Close() error {
	_, err := s.session.Close()
	return err
}

// Write should write immediately to the output, and not buffer writes
// (Telegraf manages the buffer for you). Returning an error will fail this
// batch of writes and the entire batch will be retried automatically.
func (s *IoTDB) Write(metrics []telegraf.Metric) error {
	// Convert Metrics to Records with Tags
	rwt, err := s.convertMetricsToRecordsWithTags(metrics)
	if err != nil {
		return err
	}
	// Write to client.
	// If first writing fails, the client will automatically retry three times. If all fail, it returns an error.
	if err := s.writeRecordsWithTags(rwt); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}
	return nil
}

// Find out data type of the value and return it's id in TSDataType, and convert it if necessary.
func (s *IoTDB) getDataTypeAndValue(value interface{}) (client.TSDataType, interface{}) {
	switch v := value.(type) {
	case int32:
		return client.INT32, v
	case int64:
		return client.INT64, v
	case uint32:
		return client.INT64, int64(v)
	case uint64:
		switch s.ConvertUint64To {
		case "int64_clip":
			if v <= uint64(math.MaxInt64) {
				return client.INT64, int64(v)
			}
			return client.INT64, int64(math.MaxInt64)
		case "int64":
			return client.INT64, int64(v)
		case "text":
			return client.TEXT, strconv.FormatUint(v, 10)
		default:
			return client.UNKNOW, int64(0)
		}
	case float64:
		return client.DOUBLE, v
	case string:
		return client.TEXT, v
	case bool:
		return client.BOOLEAN, v
	default:
		return client.UNKNOW, int64(0)
	}
}

// convert Timestamp Unit according to config
func (s *IoTDB) convertTimestampOfMetric(m telegraf.Metric) (int64, error) {
	switch s.TimeStampUnit {
	case "second":
		return m.Time().Unix(), nil
	case "millisecond":
		return m.Time().UnixMilli(), nil
	case "microsecond":
		return m.Time().UnixMicro(), nil
	case "nanosecond":
		return m.Time().UnixNano(), nil
	default:
		return 0, fmt.Errorf("unknown timestamp_precision %q", s.TimeStampUnit)
	}
}

// convert Metrics to Records with tags
func (s *IoTDB) convertMetricsToRecordsWithTags(metrics []telegraf.Metric) (*recordsWithTags, error) {
	var deviceidList []string
	var measurementsList [][]string
	var valuesList [][]interface{}
	var dataTypesList [][]client.TSDataType
	var timestampList []int64
	var tagsList [][]*telegraf.Tag

	for _, metric := range metrics {
		// write `metric` to the output sink here
		var tags []*telegraf.Tag
		tags = append(tags, metric.TagList()...)
		// deal with basic parameter
		var keys []string
		var values []interface{}
		var dataTypes []client.TSDataType
		for _, field := range metric.FieldList() {
			datatype, value := s.getDataTypeAndValue(field.Value)
			if datatype == client.UNKNOW {
				return nil, fmt.Errorf("datatype of %q is unknown, values: %v", field.Key, field.Value)
			}
			keys = append(keys, field.Key)
			values = append(values, value)
			dataTypes = append(dataTypes, datatype)
		}
		// Convert timestamp into specified unit
		ts, err := s.convertTimestampOfMetric(metric)
		if err != nil {
			return nil, err
		}
		timestampList = append(timestampList, ts)
		// append all metric data of this record to lists
		deviceidList = append(deviceidList, metric.Name())
		measurementsList = append(measurementsList, keys)
		valuesList = append(valuesList, values)
		dataTypesList = append(dataTypesList, dataTypes)
		tagsList = append(tagsList, tags)
	}
	rwt := &recordsWithTags{
		DeviceIDList:     deviceidList,
		MeasurementsList: measurementsList,
		ValuesList:       valuesList,
		DataTypesList:    dataTypesList,
		TimestampList:    timestampList,
		TagsList:         tagsList,
	}
	return rwt, nil
}

// modify recordsWithTags according to 'TreatTagsAs' Configuration
func (s *IoTDB) modifyRecordsWithTags(rwt *recordsWithTags) error {
	switch s.TreatTagsAs {
	case "fields":
		// method 1: treat Tag(Key:Value) as measurement
		for index, tags := range rwt.TagsList { // for each record
			for _, tag := range tags { // for each tag of this record, append it's Key:Value to measurements
				datatype, value := s.getDataTypeAndValue(tag.Value)
				if datatype == client.UNKNOW {
					return fmt.Errorf("datatype of %q is unknown, values: %v", tag.Key, value)
				}
				rwt.MeasurementsList[index] = append(rwt.MeasurementsList[index], tag.Key)
				rwt.ValuesList[index] = append(rwt.ValuesList[index], value)
				rwt.DataTypesList[index] = append(rwt.DataTypesList[index], datatype)
			}
		}
		return nil
	case "device_id":
		// method 2: treat Tag(Key:Value) as subtree of device id
		for index, tags := range rwt.TagsList { // for each record
			topic := []string{rwt.DeviceIDList[index]}
			for _, tag := range tags { // for each tag, append it's Value
				topic = append(topic, tag.Value)
			}
			rwt.DeviceIDList[index] = strings.Join(topic, ".")
		}
		return nil
	default:
		// something go wrong. This configuration should have been checked in func Init().
		return fmt.Errorf("unknown 'convert_tags_to' method: %q", s.TreatTagsAs)
	}
}

// Write records with tags to IoTDB server
func (s *IoTDB) writeRecordsWithTags(rwt *recordsWithTags) error {
	// deal with tags
	if err := s.modifyRecordsWithTags(rwt); err != nil {
		return err
	}
	// write to IoTDB server
	status, err := s.session.InsertRecords(rwt.DeviceIDList, rwt.MeasurementsList,
		rwt.DataTypesList, rwt.ValuesList, rwt.TimestampList)
	if status != nil {
		if verifyResult := client.VerifySuccess(status); verifyResult != nil {
			s.Log.Debug(verifyResult)
		}
	}
	return err
}

func init() {
	outputs.Add("iotdb", func() telegraf.Output { return newIoTDB() })
}

// create a new IoTDB struct with default values.
func newIoTDB() *IoTDB {
	return &IoTDB{
		Host:            "localhost",
		Port:            "6667",
		User:            "root",
		Password:        "root",
		Timeout:         config.Duration(time.Second * 5),
		ConvertUint64To: "int64_clip",
		TimeStampUnit:   "nanosecond",
		TreatTagsAs:     "device_id",
	}
}
