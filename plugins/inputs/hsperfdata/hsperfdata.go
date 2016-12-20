package hsperfdata

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Hsperfdata struct {
	Directory      string
	Tags           []string
	Filter         string
	NormalizeTicks bool
}

// Perfdata structs, as defined by Hotspot (e.g. src/share/vm/runtime/vmStructs.[hc]pp)

type PrologueHeader struct {
	// 0xcafec0c0
	Magic uint32
	// big_endian == 0, little_endian == 1
	ByteOrder byte
	Major     byte
	Minor     byte
}

// endian-sensitive fields
type PrologueBody struct {
	Accessible   byte
	Used         int32
	Overflow     int32
	ModTimestamp int64
	EntryOffset  int32
	NumEntries   int32
}

type Entry struct {
	EntryLength  int32
	NameOffset   int32
	VectorLength int32
	DataType     byte
	Flags        byte
	DataUnits    byte
	DataVar      byte
	DataOffset   int32
}

// see: com.sun.hotspot.perfdata.Variability
const (
	V_Constant = iota + 1
	V_Monotonic
	V_Variable
)

// see: com.sun.hotspot.perfdata.Units
const (
	U_None = iota + 1
	U_Bytes
	U_Ticks
	U_Events
	U_String
	U_Hertz
)

const (
	HaveBeginTime = 1 << iota
	HaveFrequency
	HaveTicks
)

func (header *PrologueHeader) GetEndian() binary.ByteOrder {
	if header.ByteOrder == 0 {
		return binary.BigEndian
	} else {
		return binary.LittleEndian
	}
}

var sampleConfig = `
  ## Use an arbitary directory to gather perfdata. This can be useful if you
  ## want data belonging to a different user.
  # directory = "/tmp/hsperfdata_otheruser"
  #
  ## Use the value for these keys in the hsperfdata as tags, not fields. By
  ## default everything is a field.
  # tags = ["sun.rt.jvmVersion"]
  #
  ## Filter the keys in the hsperfdata that are turned into fields by a given
  ## regexp
  # filter = "^java\\."
  #
  ## Whether to convert fields measured in JVM "ticks" into nanoseconds using
  ## the "sun.os.hrt.frequency" key in the hsperfdata. If true, any values that
  ##  can't be converted will be dropped.
  # normalizeTicks = true
`

func (n *Hsperfdata) SampleConfig() string {
	return sampleConfig
}

func (n *Hsperfdata) GetFiles() (map[string]string, error) {
	dir := n.Directory
	if dir == "" {
		// pick a sensible default: /tmp/hsperfdata_<user>
		var user string
		if runtime.GOOS == "windows" {
			user = os.Getenv("USERNAME")
		} else {
			user = os.Getenv("USER")
		}
		if user == "" {
			return nil, fmt.Errorf("error: Environment variable USER not set")
		}
		dir = filepath.Join(os.TempDir(), "hsperfdata_"+user)
	}

	retval := make(map[string]string)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		// e.g. no such directory or no permissions - just don't record metrics
		return retval, nil
	}

	for _, f := range files {
		// the hsperfdata files are named after the pid
		if _, err := strconv.Atoi(f.Name()); err == nil {
			retval[filepath.Join(dir, f.Name())] = f.Name()
		}
	}

	return retval, nil
}

func (n *Hsperfdata) IsTag(name string) bool {
	for _, tag := range n.Tags {
		if name == tag {
			return true
		}
	}
	return false
}

func (n *Hsperfdata) GatherOne(acc telegraf.Accumulator, file string, pid string) error {
	tags := map[string]string{"pid": pid}
	fields := make(map[string]interface{})

	// read a snapshot into memory
	data, err := ioutil.ReadFile(file)
	buffer := bytes.NewReader(data)

	header := PrologueHeader{}
	{
		err = binary.Read(buffer, binary.BigEndian, &header)
		if err != nil {
			return err
		}
		if header.Magic != 0xcafec0c0 {
			return fmt.Errorf("illegal magic %v", header.Magic)
		}
		if header.Major != 2 || header.Minor != 0 {
			return fmt.Errorf("unsupported version %v.%v", header.Major, header.Minor)
		}
	}

	body := PrologueBody{}
	{
		err = binary.Read(
			buffer,
			header.GetEndian(),
			&body)
		if body.Accessible != 1 {
			return fmt.Errorf("not accessible %v", body.Accessible)
		}
	}

	// "ticks" are the unit of measurement of time in the Hotspot JVM. We'll
	// work out when this sample was taking (in ticks) by taking the current
	// ticks and add the start time of the JVM.
	timePartsFound := 0
	jvmStart := time.Time{}
	ticks := int64(0)
	var frequency time.Duration // the length of a tick

	type tickfield struct {
		name  string
		value int64
	}
	unconvertedTickFields := []tickfield{}

	filter, err := regexp.Compile(n.Filter)
	if err != nil {
		return err
	}

	start_offset := body.EntryOffset
	entry := Entry{}
	for i := int32(1); i <= body.NumEntries; i++ {
		buffer.Seek(int64(start_offset), 0)
		err = binary.Read(buffer, header.GetEndian(), &entry)
		if err != nil {
			return err
		}

		name_start := int(start_offset) + int(entry.NameOffset)
		name_end := bytes.Index(data[name_start:], []byte{'\x00'})
		if name_end < 0 {
			return fmt.Errorf("invalid binary: %v", err)
		}
		name := string(data[name_start : int(name_start)+name_end])

		data_start := start_offset + entry.DataOffset

		var value interface{} = nil
		if entry.VectorLength == 0 {
			buffer.Seek(int64(data_start), 0)

			switch entry.DataType {
			case 'J':
				v := int64(0)
				err = binary.Read(buffer, header.GetEndian(), &v)

				if name == "sun.rt.createVmBeginTime" {
					// wall clock time in millis since the epoch. See
					// TraceVmCreationTime in management.hpp of Hotspot.
					jvmStart = time.Unix(0, v*int64(time.Millisecond))
					timePartsFound |= HaveBeginTime
				} else if name == "sun.os.hrt.ticks" {
					// The number of ticks since the Hotspot JVM started. See
					// HighResTimeSampler in statSampler.cpp, which delegates
					// to os::elapsed_counter.
					ticks = v
					timePartsFound |= HaveTicks
				} else if name == "sun.os.hrt.frequency" {
					// how big each "tick" is - but in Hz.
					frequency = time.Duration(time.Second.Nanoseconds() / v)
					timePartsFound |= HaveFrequency

					// now we have the frequency, we can convert all the tick
					// fields we were storing
					for _, unconverted := range unconvertedTickFields {
						n.append(
							acc,
							filter,
							tags,
							fields,
							unconverted.name,
							unconverted.value*frequency.Nanoseconds())
					}
				}

				if n.NormalizeTicks && entry.DataUnits == U_Ticks {
					if (timePartsFound & HaveFrequency) == HaveFrequency {
						value = v * frequency.Nanoseconds()
					} else {
						// don't have the frequency we need to convert, so put it
						// on the list for later
						unconvertedTickFields = append(unconvertedTickFields, tickfield{name, int64(v)})
					}
				} else {
					// just a normal value
					value = v
				}
			case 'I':
				v := int32(0)
				err = binary.Read(buffer, header.GetEndian(), &v)
				value = v
			case 'S':
				v := int16(0)
				err = binary.Read(buffer, header.GetEndian(), &v)
				value = v
			case 'B':
				v := byte(0)
				err = binary.Read(buffer, header.GetEndian(), &v)
				value = v
			case 'F':
				v := float32(0)
				err = binary.Read(buffer, header.GetEndian(), &v)
				value = v
			case 'D':
				v := float64(0)
				err = binary.Read(buffer, header.GetEndian(), &v)
				value = v
			}
			if err != nil {
				return err
			}
		} else if entry.DataType == 'B' && entry.DataUnits == U_String && entry.DataVar != V_Monotonic {
			// how string values are marked
			v := string(bytes.Trim(data[data_start:data_start+entry.VectorLength], "\x00"))

			// a special tag - the "name" of the running java process
			if name == "sun.rt.javaCommand" {
				procname := strings.SplitN(v, " ", 2)[0]
				if procname != "" {
					tags["procname"] = procname
				}
			}

			value = v
		}

		n.append(acc, filter, tags, fields, name, value)
		start_offset += entry.EntryLength
	}

	// Converting the number of ticks into a wall-clock time is machine-
	// specific.
	timePartsForConversion := HaveFrequency | HaveTicks | HaveBeginTime
	if (timePartsFound & timePartsForConversion) == timePartsForConversion {
		acc.AddFields("java", fields, tags, jvmStart.Add(time.Duration(ticks)*frequency))
	} else {
		// not enough info in the hsperfdata to reconstruct the time, so just
		// use the current time
		acc.AddFields("java", fields, tags)
	}

	return nil
}

func (n *Hsperfdata) append(
	acc telegraf.Accumulator,
	filter *regexp.Regexp,
	tags map[string]string,
	fields map[string]interface{},
	name string,
	value interface{}) {
	if value != nil {
		if n.IsTag(name) {
			// don't tag metrics with "nil", just skip the tag if it's not there
			if valuestr, ok := value.(string); ok {
				tags[name] = valuestr
			} else {
				tags[name] = fmt.Sprintf("%#v", value)
			}
		} else if filter.MatchString(name) {
			fields[name] = value
		}
	}
}

func (n *Hsperfdata) Gather(acc telegraf.Accumulator) error {
	files, err := n.GetFiles()
	if err != nil {
		// the directory doesn't exist - so there aren't any Java processes running
		return nil
	}

	var errS string
	for file, pid := range files {
		// if we can't read one pid file, keep going - as we might be able to
		// read others
		err = n.GatherOne(acc, file, pid)
		if err != nil {
			errS += err.Error() + " "
		}
	}

	if errS != "" {
		return fmt.Errorf(strings.Trim(errS, " "))
	}
	return nil
}

func (n *Hsperfdata) Description() string {
	return "Read performance data from running hotspot JVMs from shared memory"
}

func init() {
	inputs.Add("hsperfdata", func() telegraf.Input {
		return &Hsperfdata{}
	})
}
