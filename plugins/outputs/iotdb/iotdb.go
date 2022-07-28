//go:generate ../../../tools/readme_config_includer/generator
package iotdb

// iotdb.go

import (
	_ "embed"
	"errors"
	"fmt"
	"math"

	// Register IoTDB go client
	"github.com/apache/iotdb-client-go/client"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

type IoTDB struct {
	Host            string `toml:"host"`
	Port            string `toml:"port"`
	User            string `toml:"user"`
	Password        string `toml:"password"`
	Timeout         int    `toml:"timeout"`
	ConvertUint64To string `toml:"convertUint64To"`
	TimeStampUnit   string `toml:"timeStampUnit"`
	TreateTagsAs    string `toml:"treateTagsAs"`
	session         *client.Session

	Log telegraf.Logger `toml:"-"`
}

type RecordsWithTags struct {
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
	var errorMsg string
	if s.Timeout < 0 {
		errorMsg = fmt.Sprintf("IoTDB Config Warning: The value of 'timeout' is negative:%d. Now it's fixed to 0.", s.Timeout)
		s.Log.Warnf(errorMsg)
		s.Timeout = 0
	}
	if !(s.ConvertUint64To == "ToInt64" ||
		s.ConvertUint64To == "ForceToInt64" ||
		s.ConvertUint64To == "Text") {
		errorMsg = fmt.Sprintf("IoTDB Config Warning: The value of 'ConvertUint64To' is invalid: %s. Now it's fixed to 'ToInt64'.", s.ConvertUint64To)
		s.Log.Warnf(errorMsg)
		s.ConvertUint64To = "ToInt64"
	}
	if !(s.TimeStampUnit == "second" ||
		s.TimeStampUnit == "millisecond" ||
		s.TimeStampUnit == "microsecond" ||
		s.TimeStampUnit == "nanosecond") {
		errorMsg = fmt.Sprintf("IoTDB Config Warning: The value of 'TimeStampUnit' is invalid: %s. Now it's fixed to 'nanosecond'.", s.TimeStampUnit)
		s.Log.Warnf(errorMsg)
		s.TimeStampUnit = "nanosecond"
	}
	if !(s.TreateTagsAs == "Measurements" || s.TreateTagsAs == "DeviceID_subtree") {
		errorMsg = fmt.Sprintf("IoTDB Config Warning: The value of 'TreateTagsAs' is invalid: %s. Now it's fixed to 'Measurements'.", s.TreateTagsAs)
		s.Log.Warnf(errorMsg)
		s.TreateTagsAs = "Measurements"
	}
	s.Log.Info("IoTDB output plugin initialization completed.")
	return nil
}

func (s *IoTDB) Connect() error {
	// Make any connection required here
	config := &client.Config{
		Host:     s.Host,
		Port:     s.Port,
		UserName: s.User,
		Password: s.Password,
	}
	var ss = client.NewSession(config)
	s.session = &ss
	if err := s.session.Open(false, s.Timeout); err != nil {
		s.Log.Errorf("IoTDB Connect Error: Fail to connect host:'%s', port:'%s', err:%v", s.Host, s.Port, err)
		return err
	}

	return nil
}

func (s *IoTDB) Close() error {
	// Close any connections here.
	// Write will not be called once Close is called, so there is no need to synchronize.
	_, err := s.session.Close()
	if err != nil {
		s.Log.Errorf("IoTDB Close Error: %v", err)
	}
	return nil
}

// Write should write immediately to the output, and not buffer writes
// (Telegraf manages the buffer for you). Returning an error will fail this
// batch of writes and the entire batch will be retried automatically.
func (s *IoTDB) Write(metrics []telegraf.Metric) error {
	// Convert Metrics to Records with Tags
	recordsWithTags, err := s.ConvertMetricsToRecordsWithTags(metrics)
	if err != nil {
		s.Log.Errorf(err.Error())
		return err
	}
	// Wirte to client
	// status, err := s.session.InsertRecords(deviceidList, measurementsList, dataTypesList, valuesList, timestampList)
	err = s.WriteRecordsWithTags(recordsWithTags)
	if err != nil {
		s.Log.Errorf("IoTDB Write Error: %s", err.Error())
	}
	return err
}

// Find out data type of the value and return it's id in TSDataType, and convert it if nessary.
func (s *IoTDB) getDataTypeAndValue(value interface{}) (client.TSDataType, interface{}) {
	switch v := value.(type) {
	case int32:
		return client.INT32, v
	case int64:
		return client.INT64, v
	case uint32:
		return client.INT64, int64(v)
	case uint64:
		if s.ConvertUint64To == "ToInt64" {
			if v <= uint64(math.MaxInt64) {
				return client.INT64, int64(v)
			}
			return client.INT64, int64(math.MaxInt64)
		} else if s.ConvertUint64To == "ForceToInt64" {
			return client.INT64, int64(v)
		} else if s.ConvertUint64To == "Text" {
			return client.TEXT, fmt.Sprintf("%d", v)
		} else {
			s.Log.Errorf("unknown converstaion configuration of 'convertUint64To': %s", s.ConvertUint64To)
			return client.UNKNOW, int64(0)
		}
	case float64:
		return client.DOUBLE, v
	case string:
		return client.TEXT, v
	case bool:
		return client.BOOLEAN, v
	default:
		s.Log.Errorf("Unknown datatype: '%T' %v", value, value)
		return client.UNKNOW, int64(0)
	}
}

// convert Timestamp Unit accroding to config
func (s *IoTDB) ConvertTimestampOfMetric(m telegraf.Metric) (int64, error) {
	if s.TimeStampUnit == "second" {
		return m.Time().Unix(), nil
	} else if s.TimeStampUnit == "millisecond" {
		return m.Time().UnixMilli(), nil
	} else if s.TimeStampUnit == "microsecond" {
		return m.Time().UnixMicro(), nil
	} else if s.TimeStampUnit == "nanosecond" {
		return m.Time().UnixNano(), nil
	}
	errorMsg := fmt.Sprintf("IoTDB Configuration Error: unknown TimeStampUnit: %s", s.TimeStampUnit)
	s.Log.Errorf(errorMsg)
	return 0, errors.New(errorMsg)
}

// convert Metrics to Records with tags
func (s *IoTDB) ConvertMetricsToRecordsWithTags(metrics []telegraf.Metric) (*RecordsWithTags, error) {
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
		// deal with basic paramter
		var keys []string
		var values []interface{}
		var dataTypes []client.TSDataType
		for _, field := range metric.FieldList() {
			datatype, value := s.getDataTypeAndValue(field.Value)
			if datatype != client.UNKNOW {
				keys = append(keys, field.Key)
				values = append(values, value)
				dataTypes = append(dataTypes, datatype)
			}
		}
		// Convert timestamp into specified unit
		ts, tsError := s.ConvertTimestampOfMetric(metric)
		if tsError != nil {
			return nil, tsError
		}
		timestampList = append(timestampList, ts)
		// append all metric data of this record to lists
		deviceidList = append(deviceidList, metric.Name())
		measurementsList = append(measurementsList, keys)
		valuesList = append(valuesList, values)
		dataTypesList = append(dataTypesList, dataTypes)
		tagsList = append(tagsList, tags)
	}
	var recordsWithTags = &RecordsWithTags{
		DeviceIDList:     deviceidList,
		MeasurementsList: measurementsList,
		ValuesList:       valuesList,
		DataTypesList:    dataTypesList,
		TimestampList:    timestampList,
		TagsList:         tagsList,
	}
	return recordsWithTags, nil
}

// modifiy RecordsWithTags according to 'TreateTagsAs' Configuration
func (s *IoTDB) ModifiyRecordsWithTags(rwt *RecordsWithTags) error {
	if s.TreateTagsAs == "Measurements" {
		// method 1: treate Tag(Key:Value) as measurement
		for index, tags := range rwt.TagsList { // for each record
			for _, tag := range tags { // for each tag of this record, append it's Key:Value to measurements
				datatype, value := s.getDataTypeAndValue(tag.Value)
				if datatype != client.UNKNOW {
					rwt.MeasurementsList[index] = append(rwt.MeasurementsList[index], tag.Key)
					rwt.ValuesList[index] = append(rwt.ValuesList[index], value)
					rwt.DataTypesList[index] = append(rwt.DataTypesList[index], datatype)
				}
			}
		}
		return nil
	} else if s.TreateTagsAs == "DeviceID_subtree" {
		// method 2: treate Tag(Key:Value) as subtree of device id
		for index, tags := range rwt.TagsList { // for each record
			subfix := ""
			for _, tag := range tags { // for each tag, append it's Value
				subfix = subfix + "." + tag.Value
			}
			rwt.DeviceIDList[index] = rwt.DeviceIDList[index] + subfix
		}
		return nil
	}
	// something go wrong. This configuration should have been checked in func Init().
	errorMsg := fmt.Sprintf("IoTDB Configuration Error: unknown TreateTagsAs: %s", s.TreateTagsAs)
	s.Log.Errorf(errorMsg)
	return errors.New(errorMsg)
}

// Write records with tags to IoTDB server
func (s *IoTDB) WriteRecordsWithTags(rwt *RecordsWithTags) error {
	// deal with tags
	modifyError := s.ModifiyRecordsWithTags(rwt)
	if modifyError != nil {
		return modifyError
	}
	// write to IoTDB server
	status, err := s.session.InsertRecords(rwt.DeviceIDList, rwt.MeasurementsList,
		rwt.DataTypesList, rwt.ValuesList, rwt.TimestampList)
	if status != nil {
		if verifyResult := client.VerifySuccess(status); verifyResult != nil {
			s.Log.Info(verifyResult)
		}
	}
	return err
}

func init() {
	outputs.Add("iotdb", func() telegraf.Output { return newIoTDB() })
}

func newIoTDB() *IoTDB {
	return &IoTDB{}
}
