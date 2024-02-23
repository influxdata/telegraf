package snmp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"

	"github.com/gosnmp/gosnmp"
)

// Field holds the configuration for a Field to look up.
type Field struct {
	// Name will be the name of the field.
	Name string
	// OID is prefix for this field. The plugin will perform a walk through all
	// OIDs with this as their parent. For each value found, the plugin will strip
	// off the OID prefix, and use the remainder as the index. For multiple fields
	// to show up in the same row, they must share the same index.
	Oid string
	// OidIndexSuffix is the trailing sub-identifier on a table record OID that will be stripped off to get the record's index.
	OidIndexSuffix string
	// OidIndexLength specifies the length of the index in OID path segments. It can be used to remove sub-identifiers that vary in content or length.
	OidIndexLength int
	// IsTag controls whether this OID is output as a tag or a value.
	IsTag bool
	// Conversion controls any type conversion that is done on the value.
	//  "float"/"float(0)" will convert the value into a float.
	//  "float(X)" will convert the value into a float, and then move the decimal before Xth right-most digit.
	//  "int" will convert the value into an integer.
	//  "hwaddr" will convert a 6-byte string to a MAC address.
	//  "ipaddr" will convert the value to an IPv4 or IPv6 address.
	//  "enum"/"enum(1)" will convert the value according to its syntax. (Only supported with gosmi translator)
	Conversion string
	// Translate tells if the value of the field should be snmptranslated
	Translate bool
	// Secondary index table allows to merge data from two tables with different index
	//  that this filed will be used to join them. There can be only one secondary index table.
	SecondaryIndexTable bool
	// This field is using secondary index, and will be later merged with primary index
	//  using SecondaryIndexTable. SecondaryIndexTable and SecondaryIndexUse are exclusive.
	SecondaryIndexUse bool
	// Controls if entries from secondary table should be added or not if joining
	//  index is present or not. I set to true, means that join is outer, and
	//  index is prepended with "Secondary." for missing values to avoid overlapping
	//  indexes from both tables.
	// Can be set per field or globally with SecondaryIndexTable, global true overrides
	//  per field false.
	SecondaryOuterJoin bool

	initialized bool
	translator  Translator
}

// init() converts OID names to numbers, and sets the .Name attribute if unset.
func (f *Field) Init(tr Translator) error {
	if f.initialized {
		return nil
	}

	f.translator = tr

	// check if oid needs translation or name is not set
	if strings.ContainsAny(f.Oid, ":abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") || f.Name == "" {
		_, oidNum, oidText, conversion, err := f.translator.SnmpTranslate(f.Oid)
		if err != nil {
			return fmt.Errorf("translating: %w", err)
		}
		f.Oid = oidNum
		if f.Name == "" {
			f.Name = oidText
		}
		if f.Conversion == "" {
			f.Conversion = conversion
		}
		//TODO use textual convention conversion from the MIB
	}

	if f.SecondaryIndexTable && f.SecondaryIndexUse {
		return errors.New("SecondaryIndexTable and UseSecondaryIndex are exclusive")
	}

	if !f.SecondaryIndexTable && !f.SecondaryIndexUse && f.SecondaryOuterJoin {
		return errors.New("SecondaryOuterJoin set to true, but field is not being used in join")
	}

	f.initialized = true
	return nil
}

// fieldConvert converts from any type according to the conv specification
func (f *Field) Convert(ent gosnmp.SnmpPDU) (v interface{}, err error) {
	if f.Conversion == "" {
		if bs, ok := ent.Value.([]byte); ok {
			return string(bs), nil
		}
		return ent.Value, nil
	}

	var d int
	if _, err := fmt.Sscanf(f.Conversion, "float(%d)", &d); err == nil || f.Conversion == "float" {
		v = ent.Value
		switch vt := v.(type) {
		case float32:
			v = float64(vt) / math.Pow10(d)
		case float64:
			v = vt / math.Pow10(d)
		case int:
			v = float64(vt) / math.Pow10(d)
		case int8:
			v = float64(vt) / math.Pow10(d)
		case int16:
			v = float64(vt) / math.Pow10(d)
		case int32:
			v = float64(vt) / math.Pow10(d)
		case int64:
			v = float64(vt) / math.Pow10(d)
		case uint:
			v = float64(vt) / math.Pow10(d)
		case uint8:
			v = float64(vt) / math.Pow10(d)
		case uint16:
			v = float64(vt) / math.Pow10(d)
		case uint32:
			v = float64(vt) / math.Pow10(d)
		case uint64:
			v = float64(vt) / math.Pow10(d)
		case []byte:
			vf, _ := strconv.ParseFloat(string(vt), 64)
			v = vf / math.Pow10(d)
		case string:
			vf, _ := strconv.ParseFloat(vt, 64)
			v = vf / math.Pow10(d)
		}
		return v, nil
	}

	if f.Conversion == "int" {
		v = ent.Value
		switch vt := v.(type) {
		case float32:
			v = int64(vt)
		case float64:
			v = int64(vt)
		case int:
			v = int64(vt)
		case int8:
			v = int64(vt)
		case int16:
			v = int64(vt)
		case int32:
			v = int64(vt)
		case int64:
			v = vt
		case uint:
			v = int64(vt)
		case uint8:
			v = int64(vt)
		case uint16:
			v = int64(vt)
		case uint32:
			v = int64(vt)
		case uint64:
			v = int64(vt)
		case []byte:
			v, _ = strconv.ParseInt(string(vt), 10, 64)
		case string:
			v, _ = strconv.ParseInt(vt, 10, 64)
		}
		return v, nil
	}

	if f.Conversion == "hwaddr" {
		switch vt := ent.Value.(type) {
		case string:
			v = net.HardwareAddr(vt).String()
		case []byte:
			v = net.HardwareAddr(vt).String()
		default:
			return nil, fmt.Errorf("invalid type (%T) for hwaddr conversion", v)
		}
		return v, nil
	}

	split := strings.Split(f.Conversion, ":")
	if split[0] == "hextoint" && len(split) == 3 {
		endian := split[1]
		bit := split[2]

		bv, ok := ent.Value.([]byte)
		if !ok {
			return ent.Value, nil
		}

		switch endian {
		case "LittleEndian":
			switch bit {
			case "uint64":
				v = binary.LittleEndian.Uint64(bv)
			case "uint32":
				v = binary.LittleEndian.Uint32(bv)
			case "uint16":
				v = binary.LittleEndian.Uint16(bv)
			default:
				return nil, fmt.Errorf("invalid bit value (%s) for hex to int conversion", bit)
			}
		case "BigEndian":
			switch bit {
			case "uint64":
				v = binary.BigEndian.Uint64(bv)
			case "uint32":
				v = binary.BigEndian.Uint32(bv)
			case "uint16":
				v = binary.BigEndian.Uint16(bv)
			default:
				return nil, fmt.Errorf("invalid bit value (%s) for hex to int conversion", bit)
			}
		default:
			return nil, fmt.Errorf("invalid Endian value (%s) for hex to int conversion", endian)
		}

		return v, nil
	}

	if f.Conversion == "ipaddr" {
		var ipbs []byte

		switch vt := ent.Value.(type) {
		case string:
			ipbs = []byte(vt)
		case []byte:
			ipbs = vt
		default:
			return nil, fmt.Errorf("invalid type (%T) for ipaddr conversion", v)
		}

		switch len(ipbs) {
		case 4, 16:
			v = net.IP(ipbs).String()
		default:
			return nil, fmt.Errorf("invalid length (%d) for ipaddr conversion", len(ipbs))
		}

		return v, nil
	}

	if f.Conversion == "enum" {
		return f.translator.SnmpFormatEnum(ent.Name, ent.Value, false)
	}

	if f.Conversion == "enum(1)" {
		return f.translator.SnmpFormatEnum(ent.Name, ent.Value, true)
	}

	return nil, fmt.Errorf("invalid conversion type %q", f.Conversion)
}
