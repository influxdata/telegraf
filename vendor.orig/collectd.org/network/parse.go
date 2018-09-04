package network // import "collectd.org/network"

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"time"

	"collectd.org/api"
	"collectd.org/cdtime"
)

// ErrInvalid is returned when parsing the network data was aborted due to
// illegal data format.
var ErrInvalid = errors.New("invalid data")

// ParseOpts holds confiruation options for "Parse()".
type ParseOpts struct {
	// PasswordLookup is used lookup passwords to verify signed data and
	// decrypt encrypted data.
	PasswordLookup PasswordLookup
	// SecurityLevel determines the minimum security level expected by the
	// caller. If set to "Sign", only signed and encrypted data is returned
	// by Parse(), if set to "Encrypt", only encrypted data is returned.
	SecurityLevel SecurityLevel
	// TypesDB for looking up DS names and verify data source types.
	TypesDB *api.TypesDB
}

// Parse parses the binary network format and returns a slice of ValueLists. If
// a parse error is encountered, all ValueLists parsed to this point are
// returned as well as the error. Unknown "parts" are silently ignored.
func Parse(b []byte, opts ParseOpts) ([]*api.ValueList, error) {
	return parse(b, None, opts)
}

func readUint16(buf *bytes.Buffer) (uint16, error) {
	read := buf.Next(2)
	if len(read) != 2 {
		return 0, ErrInvalid
	}
	return binary.BigEndian.Uint16(read), nil
}

func parse(b []byte, sl SecurityLevel, opts ParseOpts) ([]*api.ValueList, error) {
	var valueLists []*api.ValueList

	var state api.ValueList
	buf := bytes.NewBuffer(b)

	for buf.Len() > 0 {
		partType, err := readUint16(buf)
		if err != nil {
			return nil, ErrInvalid
		}
		partLengthUnsigned, err := readUint16(buf)
		if err != nil {
			return nil, ErrInvalid
		}
		partLength := int(partLengthUnsigned)

		if partLength < 5 || partLength-4 > buf.Len() {
			return valueLists, fmt.Errorf("invalid length %d", partLength)
		}

		// First 4 bytes were already read
		partLength -= 4

		payload := buf.Next(partLength)
		if len(payload) != partLength {
			return valueLists, fmt.Errorf("invalid length: want %d, got %d", partLength, len(payload))
		}

		switch partType {
		case typeHost, typePlugin, typePluginInstance, typeType, typeTypeInstance:
			if err := parseIdentifier(partType, payload, &state); err != nil {
				return valueLists, err
			}

		case typeInterval, typeIntervalHR, typeTime, typeTimeHR:
			if err := parseTime(partType, payload, &state); err != nil {
				return valueLists, err
			}

		case typeValues:
			v, err := parseValues(payload)
			if err != nil {
				return valueLists, err
			}

			vl := state
			vl.Values = v

			if opts.TypesDB != nil {
				ds, ok := opts.TypesDB.DataSet(state.Type)
				if !ok {
					log.Printf("unable to find %q in TypesDB", state.Type)
					continue
				}

				// convert []api.Value to []interface{}
				ifValues := make([]interface{}, len(vl.Values))
				for i, v := range vl.Values {
					ifValues[i] = v
				}

				// cast all values to the correct data source type.
				// Returns an error if the number of values is incorrect.
				v, err := ds.Values(ifValues...)
				if err != nil {
					log.Printf("unable to convert values according to TypesDB: %v", err)
					continue
				}
				vl.Values = v
				vl.DSNames = ds.Names()
			}

			if opts.SecurityLevel <= sl {
				valueLists = append(valueLists, &vl)
			}

		case typeSignSHA256:
			vls, err := parseSignSHA256(payload, buf.Bytes(), opts)
			if err != nil {
				return valueLists, err
			}
			valueLists = append(valueLists, vls...)

		case typeEncryptAES256:
			vls, err := parseEncryptAES256(payload, opts)
			if err != nil {
				return valueLists, err
			}
			valueLists = append(valueLists, vls...)

		default:
			log.Printf("ignoring field of type %#x", partType)
		}
	}

	return valueLists, nil
}

func parseIdentifier(partType uint16, payload []byte, state *api.ValueList) error {
	str, err := parseString(payload)
	if err != nil {
		return err
	}

	switch partType {
	case typeHost:
		state.Identifier.Host = str
	case typePlugin:
		state.Identifier.Plugin = str
	case typePluginInstance:
		state.Identifier.PluginInstance = str
	case typeType:
		state.Identifier.Type = str
	case typeTypeInstance:
		state.Identifier.TypeInstance = str
	}

	return nil
}

func parseTime(partType uint16, payload []byte, state *api.ValueList) error {
	v, err := parseInt(payload)
	if err != nil {
		return err
	}

	switch partType {
	case typeInterval:
		state.Interval = time.Duration(v) * time.Second
	case typeIntervalHR:
		state.Interval = cdtime.Time(v).Duration()
	case typeTime:
		state.Time = time.Unix(int64(v), 0)
	case typeTimeHR:
		state.Time = cdtime.Time(v).Time()
	}

	return nil
}

func parseValues(b []byte) ([]api.Value, error) {
	buffer := bytes.NewBuffer(b)

	var n uint16
	if err := binary.Read(buffer, binary.BigEndian, &n); err != nil {
		return nil, err
	}

	if int(n*9) != buffer.Len() {
		return nil, ErrInvalid
	}

	types := make([]byte, n)
	values := make([]api.Value, n)

	if _, err := buffer.Read(types); err != nil {
		return nil, err
	}

	for i, typ := range types {
		switch typ {
		case dsTypeGauge:
			var v float64
			if err := binary.Read(buffer, binary.LittleEndian, &v); err != nil {
				return nil, err
			}
			values[i] = api.Gauge(v)

		case dsTypeDerive:
			var v int64
			if err := binary.Read(buffer, binary.BigEndian, &v); err != nil {
				return nil, err
			}
			values[i] = api.Derive(v)

		case dsTypeCounter:
			var v uint64
			if err := binary.Read(buffer, binary.BigEndian, &v); err != nil {
				return nil, err
			}
			values[i] = api.Counter(v)

		default:
			return nil, ErrInvalid
		}
	}

	return values, nil
}

func parseSignSHA256(pkg, payload []byte, opts ParseOpts) ([]*api.ValueList, error) {
	ok, err := verifySHA256(pkg, payload, opts.PasswordLookup)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New("SHA256 verification failure")
	}

	return parse(payload, Sign, opts)
}

func parseEncryptAES256(payload []byte, opts ParseOpts) ([]*api.ValueList, error) {
	plaintext, err := decryptAES256(payload, opts.PasswordLookup)
	if err != nil {
		return nil, errors.New("AES256 decryption failure")
	}

	return parse(plaintext, Encrypt, opts)
}

func parseInt(b []byte) (uint64, error) {
	if len(b) != 8 {
		return 0, ErrInvalid
	}

	var i uint64
	buf := bytes.NewBuffer(b)
	if err := binary.Read(buf, binary.BigEndian, &i); err != nil {
		return 0, err
	}

	return i, nil
}

func parseString(b []byte) (string, error) {
	if b[len(b)-1] != 0 {
		return "", ErrInvalid
	}

	buf := bytes.NewBuffer(b[:len(b)-1])
	return buf.String(), nil
}
