package protobuf

import (
	"bytes"
	"encoding/binary"
	fmt "fmt"

	proto "github.com/golang/protobuf/proto"
	"github.com/influxdata/telegraf"
)

type ProtobufSerializer struct {
	PrependLength bool
}

func convertFields(orig map[string]interface{}) (map[string]*FieldValue, error) {
	r := make(map[string]*FieldValue)

	for k, v := range orig {
		switch v := v.(type) {
		case string:
			r[k] = &FieldValue{&FieldValue_StringValue{v}}
		case int64:
			r[k] = &FieldValue{&FieldValue_IntValue{v}}
		case float64:
			r[k] = &FieldValue{&FieldValue_FloatValue{v}}
		case bool:
			r[k] = &FieldValue{&FieldValue_BoolValue{v}}
		default:
			return nil, fmt.Errorf("Unsupported field value data type: %T", v)
		}
	}

	return r, nil
}

func (s *ProtobufSerializer) Serialize(m telegraf.Metric) ([]byte, error) {
	name := m.Name()
	ts := m.Time()
	tags := m.Tags()
	fields, err := convertFields(m.Fields())
	if err != nil {
		return nil, err
	}

	protoMessage, err := proto.Marshal(&Metric{
		Name:      name,
		Timestamp: ts.UnixNano(),
		Tags:      tags,
		Fields:    fields,
	})

	if err != nil {
		return nil, err
	}

	if s.PrependLength {
		messageLength := len(protoMessage)

		buf := &bytes.Buffer{}
		binary.Write(buf, binary.LittleEndian, uint32(messageLength))
		buf.Write(protoMessage)

		protoMessage = buf.Bytes()
	}

	return protoMessage, nil
}
