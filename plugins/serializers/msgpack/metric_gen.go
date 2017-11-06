package msgpack

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Metric) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "name":
			z.Name, err = dc.ReadString()
			if err != nil {
				return
			}
		case "time":
			z.Time, err = dc.ReadTime()
			if err != nil {
				return
			}
		case "tags":
			var zb0002 uint32
			zb0002, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Tags == nil && zb0002 > 0 {
				z.Tags = make(map[string]string, zb0002)
			} else if len(z.Tags) > 0 {
				for key, _ := range z.Tags {
					delete(z.Tags, key)
				}
			}
			for zb0002 > 0 {
				zb0002--
				var za0001 string
				var za0002 string
				za0001, err = dc.ReadString()
				if err != nil {
					return
				}
				za0002, err = dc.ReadString()
				if err != nil {
					return
				}
				z.Tags[za0001] = za0002
			}
		case "fields":
			var zb0003 uint32
			zb0003, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Fields == nil && zb0003 > 0 {
				z.Fields = make(map[string]interface{}, zb0003)
			} else if len(z.Fields) > 0 {
				for key, _ := range z.Fields {
					delete(z.Fields, key)
				}
			}
			for zb0003 > 0 {
				zb0003--
				var za0003 string
				var za0004 interface{}
				za0003, err = dc.ReadString()
				if err != nil {
					return
				}
				za0004, err = dc.ReadIntf()
				if err != nil {
					return
				}
				z.Fields[za0003] = za0004
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Metric) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "name"
	err = en.Append(0x84, 0xa4, 0x6e, 0x61, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Name)
	if err != nil {
		return
	}
	// write "time"
	err = en.Append(0xa4, 0x74, 0x69, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteTime(z.Time)
	if err != nil {
		return
	}
	// write "tags"
	err = en.Append(0xa4, 0x74, 0x61, 0x67, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.Tags)))
	if err != nil {
		return
	}
	for za0001, za0002 := range z.Tags {
		err = en.WriteString(za0001)
		if err != nil {
			return
		}
		err = en.WriteString(za0002)
		if err != nil {
			return
		}
	}
	// write "fields"
	err = en.Append(0xa6, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.Fields)))
	if err != nil {
		return
	}
	for za0003, za0004 := range z.Fields {
		err = en.WriteString(za0003)
		if err != nil {
			return
		}
		err = en.WriteIntf(za0004)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Metric) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "name"
	o = append(o, 0x84, 0xa4, 0x6e, 0x61, 0x6d, 0x65)
	o = msgp.AppendString(o, z.Name)
	// string "time"
	o = append(o, 0xa4, 0x74, 0x69, 0x6d, 0x65)
	o = msgp.AppendTime(o, z.Time)
	// string "tags"
	o = append(o, 0xa4, 0x74, 0x61, 0x67, 0x73)
	o = msgp.AppendMapHeader(o, uint32(len(z.Tags)))
	for za0001, za0002 := range z.Tags {
		o = msgp.AppendString(o, za0001)
		o = msgp.AppendString(o, za0002)
	}
	// string "fields"
	o = append(o, 0xa6, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73)
	o = msgp.AppendMapHeader(o, uint32(len(z.Fields)))
	for za0003, za0004 := range z.Fields {
		o = msgp.AppendString(o, za0003)
		o, err = msgp.AppendIntf(o, za0004)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Metric) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "name":
			z.Name, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "time":
			z.Time, bts, err = msgp.ReadTimeBytes(bts)
			if err != nil {
				return
			}
		case "tags":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Tags == nil && zb0002 > 0 {
				z.Tags = make(map[string]string, zb0002)
			} else if len(z.Tags) > 0 {
				for key, _ := range z.Tags {
					delete(z.Tags, key)
				}
			}
			for zb0002 > 0 {
				var za0001 string
				var za0002 string
				zb0002--
				za0001, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				za0002, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				z.Tags[za0001] = za0002
			}
		case "fields":
			var zb0003 uint32
			zb0003, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Fields == nil && zb0003 > 0 {
				z.Fields = make(map[string]interface{}, zb0003)
			} else if len(z.Fields) > 0 {
				for key, _ := range z.Fields {
					delete(z.Fields, key)
				}
			}
			for zb0003 > 0 {
				var za0003 string
				var za0004 interface{}
				zb0003--
				za0003, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				za0004, bts, err = msgp.ReadIntfBytes(bts)
				if err != nil {
					return
				}
				z.Fields[za0003] = za0004
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Metric) Msgsize() (s int) {
	s = 1 + 5 + msgp.StringPrefixSize + len(z.Name) + 5 + msgp.TimeSize + 5 + msgp.MapHeaderSize
	if z.Tags != nil {
		for za0001, za0002 := range z.Tags {
			_ = za0002
			s += msgp.StringPrefixSize + len(za0001) + msgp.StringPrefixSize + len(za0002)
		}
	}
	s += 7 + msgp.MapHeaderSize
	if z.Fields != nil {
		for za0003, za0004 := range z.Fields {
			_ = za0004
			s += msgp.StringPrefixSize + len(za0003) + msgp.GuessSize(za0004)
		}
	}
	return
}
