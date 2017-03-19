package ganglia

import (
	"time"
	"errors"
	"io"
	"bytes"
	"strconv"
	"math"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// This is an extremely simple parser for Ganglia 3.x packets. It completely discards all metadata packets.
// Further, for all string metrics, it attempts to parse them as a number. If it cannot parse it, it
// ignores that packet silently as well. This has mainly been tested with the Java Ganglia client, which
// likes to report all numeric metrics as formatted strings.
type GangliaParser struct {
	DefaultTags map[string]string
}

func (p *GangliaParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *GangliaParser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	return metrics[0], err
}

func (p *GangliaParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	r := bytes.NewReader(buf)
	id, err := readInt(r);                          if(err != nil) { return nil, err }
	if(id >= gmetric_min && id <= gmetric_max) {

		// read data
		hostname, err := readString(r);             if(err != nil) { return nil, err } // reported hostname
		name, err := readString(r);                 if(err != nil) { return nil, err } // metric name
		spoofed, err := readInt(r);                 if(err != nil) { return nil, err } // 1 if hostname is spoofed, 0 otherwise
		_, err = readString(r);                     if(err != nil) { return nil, err } // format string (ignored)
		value, err := readValue(r, id);             if(err != nil) { return nil, err } // metric value

		if(value == nil) {
			return nil, nil // non-error but invalid metric
		}

		// create and return metric object
		tags := make(map[string]string)
		tags["gangliaHost"] = hostname
		if(spoofed != 0) {
			tags["gangliaHostSpoofed"] = "true"
		} else {
			tags["gangliaHostSpoofed"] = "false"
		}
		for k, v := range p.DefaultTags {
			tags[k] = v
		}
		fields := make(map[string]interface{})
		fields["value"] = value
		t := time.Now().UTC()
		obj, err := metric.New(name, tags, fields, t);
		if(err != nil) {
			return nil, err
		}
		metrics := make([]telegraf.Metric, 1)
		metrics[0] = obj;
		return metrics, nil
	} else {
		return nil, nil
	}
}

func readValue(r *bytes.Reader, id int) (interface{}, error) {
	switch id {
	case gmetric_ushort:
		{
			tmp, err := readUshort(r);
			if(err != nil) { return nil, err }
			return int64(tmp), nil
		}
	case gmetric_short:
		{
			tmp, err := readShort(r);
			if(err != nil) { return nil, err }
			return int64(tmp), nil
		}
	case gmetric_int:
		{
			tmp, err := readInt(r)
			if(err != nil) { return nil, err }
			return int64(tmp), nil
		}
	case gmetric_uint:
		{
			tmp, err := readUint(r)
			if(err != nil) { return nil, err }
			return int64(tmp), nil
		}
	case gmetric_string:
		{
			// attempt to coerce value into a floating point number
			valueStr, err := readString(r);
			if(err != nil) { return nil, err }
			tmp, err := strconv.ParseFloat(valueStr, 64);
			if(err != nil || math.IsNaN(tmp) || math.IsInf(tmp, 0)) {
				// if we can't parse a string, just ignore it
				return nil, nil
			}
			return float64(tmp), nil
		}
	case gmetric_float:
		{
			tmp, err := readFloat(r)
			if(err != nil) { return nil, err }
			return float64(tmp), nil
		}
	case gmetric_double:
		{
			tmp, err := readDouble(r)
			if(err != nil) { return nil, err }
			return float64(tmp), nil
		}
	default:
		return nil, errors.New("Unsupported ganglia metric type")
	}
}

func readShort(r *bytes.Reader) (int16, error) {
	var buf [2]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	n := int16(buf[1]) | int16(buf[0])<<8
	return n, err
}

func readUshort(r *bytes.Reader) (uint16, error) {
	var buf [2]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	n := uint16(buf[1]) | uint16(buf[0])<<8
	return n, err
}

func readInt(r *bytes.Reader) (int, error) {
	var buf [4]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	n := int32(buf[3]) | int32(buf[2])<<8 | int32(buf[1])<<16 | int32(buf[0])<<24
	return int(n), err
}

func readUint(r *bytes.Reader) (uint32, error) {
	var buf [4]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	n := uint32(buf[3]) | uint32(buf[2])<<8 | uint32(buf[1])<<16 | uint32(buf[0])<<24
	return n, err
}

func readFloat(r *bytes.Reader) (float32, error) {
	var buf [4]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	n := uint32(buf[3]) | uint32(buf[2])<<8 | uint32(buf[1])<<16 | uint32(buf[0])<<24
	return math.Float32frombits(n), nil
}

func readDouble(r *bytes.Reader) (float64, error) {
	var buf [8]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	n := uint64(buf[7]) | uint64(buf[6])<<8 | uint64(buf[5])<<16 | uint64(buf[4])<<24 |
			uint64(buf[3])<<32 | uint64(buf[2])<<40 | uint64(buf[1])<<48 | uint64(buf[0])<<56
	return math.Float64frombits(n), nil
}


func readString(r *bytes.Reader) (string, error)  {
	size, err := readInt(r);
	if err != nil {
		return "", err
	}
	if(size == 0) {
		return "", nil
	}
	size += (4 - (size % 4)) % 4 // pad to a multiple of 4 bytes
	if(size > 1024) {
		return "", errors.New("Cannot read more than 1024 bytes in ganglia packet")
	}
	buf := make([]byte, size)
	_, err = io.ReadFull(r, buf); if err != nil { return "", err }
	return string(buf[0:size]), nil
}

const (
	// UNUSED: gmetadata_full = 128
	gmetric_ushort = 128 + 1
	gmetric_short = 128 + 2
	gmetric_int = 128 + 3
	gmetric_uint = 128 + 4
	gmetric_string = 128 + 5
	gmetric_float = 128 + 6
	gmetric_double = 128 + 7
	// UNUSED: gmetadata_request = 128 + 8

	gmetric_min = gmetric_ushort
	gmetric_max = gmetric_double
)
