package apcupsd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// timeFormatLong is the package time format of long timestamps
	// from a NIS.
	timeFormatLong = "2006-01-02 15:04:05 -0700"
)

var (
	// timeFormatShort is the package time format of date-only timestamps
	// from a NIS.
	timeFormatShort = []string{
		"2006-01-02",
		"01/02/06",
		// Also try European-style format.
		// TODO(mdlayher): find a way to always determine if the output is
		// month-first or day-first.
		"02/01/06",
	}

	// errInvalidKeyValuePair is returned when a message is not in the expected
	// "key : value" format.
	errInvalidKeyValuePair = errors.New("invalid key/value pair")

	// errInvalidDuration is returned when a value is not in the expected
	// duration format, e.g. "10 Seconds" or "2 minutes".
	errInvalidDuration = errors.New("invalid time duration")
)

// Status is the status of an APC Uninterruptible Power Supply (UPS), as returned
// by a NIS.
type Status struct {
	APC                         string
	Date                        time.Time
	Hostname                    string
	Version                     string
	UPSName                     string
	Cable                       string
	Driver                      string
	UPSMode                     string
	StartTime                   time.Time
	Model                       string
	Status                      string
	LineVoltage                 float64
	LoadPercent                 float64
	BatteryChargePercent        float64
	TimeLeft                    time.Duration
	MinimumBatteryChargePercent float64
	MinimumTimeLeft             time.Duration
	MaximumTime                 time.Duration
	Sense                       string
	LowTransferVoltage          float64
	HighTransferVoltage         float64
	AlarmDel                    time.Duration
	BatteryVoltage              float64
	LastTransfer                string
	NumberTransfers             int
	XOnBattery                  time.Time
	TimeOnBattery               time.Duration
	CumulativeTimeOnBattery     time.Duration
	XOffBattery                 time.Time
	LastSelftest                time.Time
	Selftest                    bool
	StatusFlags                 string
	SerialNumber                string
	BatteryDate                 time.Time
	NominalInputVoltage         float64
	NominalBatteryVoltage       float64
	NominalPower                int
	Firmware                    string
	EndAPC                      time.Time
	InternalTemp                float64
	OutputVoltage               float64
	LineFrequency               float64
}

// parseKV parses an input key/value string in "key : value" format, and
// sets the appropriate struct field from the input data.
func (s *Status) parseKV(kv string) error {
	sp := strings.SplitN(kv, ":", 2)
	if len(sp) != 2 {
		return errInvalidKeyValuePair
	}

	k := strings.TrimSpace(sp[0])
	v := strings.TrimSpace(sp[1])

	// Attempt to match various common data types

	if match := s.parseKVString(k, v); match {
		return nil
	}

	if match, err := s.parseKVFloat(k, v); match {
		return err
	}

	if match, err := s.parseKVTime(k, v); match {
		return err
	}

	if match, err := s.parseKVDuration(k, v); match {
		return err
	}

	// Attempt to match uncommon data types

	var err error
	switch k {
	case keyNumXfers:
		s.NumberTransfers, err = strconv.Atoi(v)
	case keyNomPower:
		f := strings.SplitN(v, " ", 2)
		s.NominalPower, err = strconv.Atoi(f[0])
	case keySelftest:
		s.Selftest = v == "YES"
	}

	return err
}

// List of keys sent by a NIS, used to map values to Status fields.
const (
	keyAPC           = "APC"
	keyDate          = "DATE"
	keyHostname      = "HOSTNAME"
	keyVersion       = "VERSION"
	keyUPSName       = "UPSNAME"
	keyCable         = "CABLE"
	keyDriver        = "DRIVER"
	keyUPSMode       = "UPSMODE"
	keyStartTime     = "STARTTIME"
	keyModel         = "MODEL"
	keyStatus        = "STATUS"
	keyLineV         = "LINEV"
	keyLoadPct       = "LOADPCT"
	keyBCharge       = "BCHARGE"
	keyTimeLeft      = "TIMELEFT"
	keyMBattChg      = "MBATTCHG"
	keyMinTimeL      = "MINTIMEL"
	keyMaxTime       = "MAXTIME"
	keySense         = "SENSE"
	keyLoTrans       = "LOTRANS"
	keyHiTrans       = "HITRANS"
	keyAlarmDel      = "ALARMDEL"
	keyBattV         = "BATTV"
	keyLastXfer      = "LASTXFER"
	keyNumXfers      = "NUMXFERS"
	keyXOnBat        = "XONBATT"
	keyTOnBatt       = "TONBATT"
	keyCumOnBatt     = "CUMONBATT"
	keyXOffBat       = "XOFFBATT"
	keyLastStest     = "LASTSTEST"
	keySelftest      = "SELFTEST"
	keyStatFlag      = "STATFLAG"
	keySerialNo      = "SERIALNO"
	keyBattDate      = "BATTDATE"
	keyNomInV        = "NOMINV"
	keyNomBattV      = "NOMBATTV"
	keyNomPower      = "NOMPOWER"
	keyFirmware      = "FIRMWARE"
	keyEndAPC        = "END APC"
	keyITemp         = "ITEMP"
	keyOutV          = "OUTPUTV"
	keyLineFrequency = "LINEFREQ"
)

// parseKVString parses a simple string into the appropriate Status field.
// It returns true if a field was matched, and false if not.
func (s *Status) parseKVString(k string, v string) bool {
	switch k {
	case keyAPC:
		s.APC = v
	case keyHostname:
		s.Hostname = v
	case keyVersion:
		s.Version = v
	case keyUPSName:
		s.UPSName = v
	case keyCable:
		s.Cable = v
	case keyDriver:
		s.Driver = v
	case keyUPSMode:
		s.UPSMode = v
	case keyModel:
		s.Model = v
	case keyStatus:
		s.Status = v
	case keySense:
		s.Sense = v
	case keyLastXfer:
		s.LastTransfer = v
	case keyStatFlag:
		s.StatusFlags = v
	case keySerialNo:
		s.SerialNumber = v
	case keyFirmware:
		s.Firmware = v
	default:
		return false
	}

	return true
}

// parseKVFloat parses a float64 value into the appropriate Status field.
// It returns true if a field was matched, and false if not.
func (s *Status) parseKVFloat(k string, v string) (bool, error) {
	f := strings.SplitN(v, " ", 2)

	// Save repetition for function calls
	parse := func() (float64, error) {
		return strconv.ParseFloat(f[0], 64)
	}

	var err error
	switch k {
	case keyLineV:
		s.LineVoltage, err = parse()
	case keyLoadPct:
		s.LoadPercent, err = parse()
	case keyBCharge:
		s.BatteryChargePercent, err = parse()
	case keyMBattChg:
		s.MinimumBatteryChargePercent, err = parse()
	case keyLoTrans:
		s.LowTransferVoltage, err = parse()
	case keyHiTrans:
		s.HighTransferVoltage, err = parse()
	case keyBattV:
		s.BatteryVoltage, err = parse()
	case keyNomInV:
		s.NominalInputVoltage, err = parse()
	case keyNomBattV:
		s.NominalBatteryVoltage, err = parse()
	case keyITemp:
		s.InternalTemp, err = parse()
	case keyOutV:
		s.OutputVoltage, err = parse()
	case keyLineFrequency:
		s.LineFrequency, err = parse()
	default:
		return false, nil
	}

	return true, err
}

// parseKVTime parses a time.Time value into the appropriate Status field.
// It returns true if a field was matched, and false if not.
func (s *Status) parseKVTime(k string, v string) (bool, error) {
	var err error
	switch k {
	case keyDate:
		s.Date, err = parseOptionalTime(v, timeFormatLong)
	case keyStartTime:
		s.StartTime, err = parseOptionalTime(v, timeFormatLong)
	case keyXOnBat:
		s.XOnBattery, err = parseOptionalTime(v, timeFormatLong)
	case keyXOffBat:
		s.XOffBattery, err = parseOptionalTime(v, timeFormatLong)
	case keyLastStest:
		s.LastSelftest, err = parseOptionalTime(v, timeFormatLong)
	case keyBattDate:
		s.BatteryDate, err = parseOptionalTime(v, timeFormatShort...)
	case keyEndAPC:
		s.EndAPC, err = parseOptionalTime(v, timeFormatLong)
	default:
		return false, nil
	}

	return true, err
}

// parseKVDuration parses a time.Duration into the appropriate Status field.
// It returns true if a field was matched, and false if not.
func (s *Status) parseKVDuration(k string, v string) (bool, error) {
	// Save repetition for function calls
	parse := func() (time.Duration, error) {
		return parseDuration(v)
	}

	var err error
	switch k {
	case keyTimeLeft:
		s.TimeLeft, err = parse()
	case keyMinTimeL:
		s.MinimumTimeLeft, err = parse()
	case keyMaxTime:
		s.MaximumTime, err = parse()
	case keyAlarmDel:
		// No alarm delay configured.
		if v == "No alarm" {
			break
		}

		s.AlarmDel, err = parse()
	case keyTOnBatt:
		s.TimeOnBattery, err = parse()
	case keyCumOnBatt:
		s.CumulativeTimeOnBattery, err = parse()
	default:
		return false, nil
	}

	return true, err
}

// parseDuration parses a duration value returned from a NIS as a
// time.Duration.
func parseDuration(d string) (time.Duration, error) {
	ss := strings.SplitN(d, " ", 2)
	if len(ss) != 2 {
		return 0, errInvalidDuration
	}

	num := ss[0]
	unit := ss[1]

	// Normalize units into ones that time.ParseDuration expects
	switch strings.ToLower(unit) {
	case "minutes":
		unit = "m"
	case "seconds":
		unit = "s"
	}

	return time.ParseDuration(fmt.Sprintf("%s%s", num, unit))
}

// parseOptionalTime parses a time string.
//
// In addition to the specified layouts, it also accepts the special value "N/A"
// (which apcupsd reports for some values and conditions); this value is mapped
// to time.Time{}. The caller can check for this with time.IsZero().
func parseOptionalTime(value string, layouts ...string) (time.Time, error) {
	if value == "N/A" {
		return time.Time{}, nil
	}

	for _, la := range layouts {
		if time, err := time.Parse(la, value); err == nil {
			return time, nil
		}
	}

	return time.Time{}, fmt.Errorf("can't parse time: %q", value)
}
