package snmp

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gosnmp/gosnmp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
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
	//  "displayhint" will format the value according to the textual convention. (Only supported with gosmi translator)
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

// Init converts OID names to numbers, and sets the .Name attribute if unset.
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
	}

	if f.SecondaryIndexTable && f.SecondaryIndexUse {
		return errors.New("fields SecondaryIndexTable and UseSecondaryIndex are exclusive")
	}

	if !f.SecondaryIndexTable && !f.SecondaryIndexUse && f.SecondaryOuterJoin {
		return errors.New("field SecondaryOuterJoin set to true, but field is not being used in join")
	}

	switch f.Conversion {
	case "hwaddr", "enum(1)":
		config.PrintOptionValueDeprecationNotice("inputs.snmp", "field.conversion", f.Conversion, telegraf.DeprecationInfo{
			Since:  "1.33.0",
			Notice: "Use 'displayhint' instead",
		})
	}

	f.initialized = true
	return nil
}

// Convert converts from any type according to the conv specification
func (f *Field) Convert(ent gosnmp.SnmpPDU) (interface{}, error) {
	v := ent.Value

	// snmptranslate table field value here
	if f.Translate {
		if entOid, ok := v.(string); ok {
			_, _, oidText, _, err := f.translator.SnmpTranslate(entOid)
			if err == nil {
				// If no error translating, the original value should be replaced
				v = oidText
			}
		}
	}

	if f.Conversion == "" {
		// OctetStrings may contain hex data that needs its own conversion
		if ent.Type == gosnmp.OctetString && !utf8.Valid(v.([]byte)[:]) {
			return hex.EncodeToString(v.([]byte)), nil
		}
		if bs, ok := v.([]byte); ok {
			return string(bs), nil
		}
		return v, nil
	}

	var d int
	if _, err := fmt.Sscanf(f.Conversion, "float(%d)", &d); err == nil || f.Conversion == "float" {
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
			vf, err := strconv.ParseFloat(string(vt), 64)
			if err != nil {
				return nil, fmt.Errorf("failed to convert field to float with value %s: %w", vt, err)
			}
			v = vf / math.Pow10(d)
		case string:
			vf, err := strconv.ParseFloat(vt, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to convert field to float with value %s: %w", vt, err)
			}
			v = vf / math.Pow10(d)
		}
		return v, nil
	}

	if f.Conversion == "int" {
		var err error
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
			v, err = strconv.ParseInt(string(vt), 10, 64)
		case string:
			v, err = strconv.ParseInt(vt, 10, 64)
		}
		return v, err
	}

	// Deprecated: Use displayhint instead
	if f.Conversion == "hwaddr" {
		switch vt := v.(type) {
		case string:
			v = net.HardwareAddr(vt).String()
		case []byte:
			v = net.HardwareAddr(vt).String()
		default:
			return nil, fmt.Errorf("invalid type (%T) for hwaddr conversion", vt)
		}
		return v, nil
	}

	if f.Conversion == "hex" {
		switch vt := v.(type) {
		case string:
			switch ent.Type {
			case gosnmp.IPAddress:
				ip := net.ParseIP(vt)
				if ip4 := ip.To4(); ip4 != nil {
					v = hex.EncodeToString(ip4)
				} else {
					v = hex.EncodeToString(ip)
				}
			default:
				return nil, fmt.Errorf("unsupported Asn1BER (%#v) for hex conversion", ent.Type)
			}
		case []byte:
			v = hex.EncodeToString(vt)
		default:
			return nil, fmt.Errorf("unsupported type (%T) for hex conversion", vt)
		}
		return v, nil
	}

	split := strings.Split(f.Conversion, ":")
	if split[0] == "hextoint" && len(split) == 3 {
		endian := split[1]
		bit := split[2]

		bv, ok := v.([]byte)
		if !ok {
			return v, nil
		}

		var b []byte
		switch bit {
		case "uint64":
			b = make([]byte, 8)
		case "uint32":
			b = make([]byte, 4)
		case "uint16":
			b = make([]byte, 2)
		default:
			return nil, fmt.Errorf("invalid bit value (%s) for hex to int conversion", bit)
		}
		copy(b, bv)

		var byteOrder binary.ByteOrder
		switch endian {
		case "LittleEndian":
			byteOrder = binary.LittleEndian
		case "BigEndian":
			byteOrder = binary.BigEndian
		default:
			return nil, fmt.Errorf("invalid Endian value (%s) for hex to int conversion", endian)
		}

		switch bit {
		case "uint64":
			v = byteOrder.Uint64(b)
		case "uint32":
			v = byteOrder.Uint32(b)
		case "uint16":
			v = byteOrder.Uint16(b)
		}

		return v, nil
	}

	if f.Conversion == "ipaddr" {
		var ipbs []byte

		switch vt := v.(type) {
		case string:
			ipbs = []byte(vt)
		case []byte:
			ipbs = vt
		default:
			return nil, fmt.Errorf("invalid type (%T) for ipaddr conversion", vt)
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

	// Deprecated: Use displayhint instead
	if f.Conversion == "enum(1)" {
		return f.translator.SnmpFormatEnum(ent.Name, ent.Value, true)
	}

	if f.Conversion == "displayhint" {
		return f.translator.SnmpFormatDisplayHint(ent.Name, ent.Value)
	}

	return nil, fmt.Errorf("invalid conversion type %q", f.Conversion)
}
