package ublox

/*
#cgo LDFLAGS: -L./ublox-utils/build/lib -lublox-utils
#include "ublox-reader.h"
*/
import "C"
import (
	"errors"
	"time"
	"unsafe"
)

const (
	// XXX match original FusionMode values
	InitializationMode  uint8 = 0
	FusionMode          uint8 = 1
	SuspendedFusionMode uint8 = 2
	DisabledFusionMode  uint8 = 3

	None uint8 = 255
)

const (
	// XXX match original FixType values
	NoFix                      uint8 = 0
	DeadReckoningOnly          uint8 = 1
	Fix2d                      uint8 = 2
	Fix3d                      uint8 = 3
	GNSS_deadReckoningCombined uint8 = 4
	TimeOnlyFix                uint8 = 5
)

type GPSPos struct {
	Active        bool
	Lat           float64
	Lon           float64
	HorizontalAcc float64

	Heading         float64
	HeadingOfMotion float64
	HeadingAcc      float64
	HeadingIsValid  bool

	Speed    float64
	SpeedAcc float64

	Pdop    uint16
	Hdop    uint16
	SatNum  uint8
	FixType uint8

	FusionMode uint8
	Sensors    []byte

	SWVersion string
	HWVersion string
	FWVersion string

	Ts time.Time
}

type UbloxReader struct {
	device   string
	reader   unsafe.Pointer
	needInit bool
}

// note after usage you should call Free method
func NewUbloxReader(ublox_device string) *UbloxReader {
	var reader UbloxReader
	reader.device = ublox_device
	reader.reader = C.ublox_reader_new()
	reader.needInit = true
	return &reader
}

// param wait_for_data set to true if you want to wait for new data
// return position info in case of success or nil and error in case of error,
// or nil and nil if no data (wait_for_data is false or Close called)
// note safe to use after error
func (reader *UbloxReader) Pop(wait_for_data bool) (*GPSPos, error) {
	var c_err *C.char

	// init if needed
	if reader.needInit {
		cdevice := C.CString(reader.device)
		defer C.free(unsafe.Pointer(cdevice))
		if C.ublox_reader_init(unsafe.Pointer(reader.reader), cdevice, &c_err) == false {
			err := C.GoString(c_err)
			C.free(unsafe.Pointer(c_err))
			return nil, errors.New(err)
		}
	}

	reader.needInit = false

	// read data
	var is_active C.bool
	var lat C.double
	var lon C.double
	var horizontal_acc C.double

	var heading C.double
	var headingOfMot C.double
	var headingAcc C.double
	var headingIsValid C.bool

	var speed C.double
	var speedAcc C.double

	var pdop C.uint
	var satNum C.uint
	var fixType C.uint

	var fusion_mode C.uint

	var sensorsLen C.uint
	sensors := C.CString(string(make([]byte, 4*16)))
	defer C.free(unsafe.Pointer(sensors))

	var sec C.longlong
	var nsec C.longlong

	swVersion := C.CString(string(make([]byte, 30)))
	hwVersion := C.CString(string(make([]byte, 30)))
	fwVersion := C.CString(string(make([]byte, 30)))
	defer C.free(unsafe.Pointer(swVersion))
	defer C.free(unsafe.Pointer(hwVersion))
	defer C.free(unsafe.Pointer(fwVersion))

	var hdop C.uint

	fusion_mode = C.uint(None)

	status := C.ublox_reader_read(unsafe.Pointer(reader.reader), &is_active, &lat, &lon, &horizontal_acc, &heading, &headingOfMot, &headingAcc, &headingIsValid, &speed, &speedAcc, &pdop, &satNum, &fixType, &fusion_mode, sensors, &sensorsLen, swVersion, hwVersion, fwVersion, &hdop, &sec, &nsec, C.bool(wait_for_data), &c_err)
	if status == -1 {
		err := C.GoString(c_err)
		C.free(unsafe.Pointer(c_err))
		reader.needInit = true
		return nil, errors.New(err)
	} else if status == 0 {
		return nil, nil
	}

	var data GPSPos
	data.Active = bool(is_active)
	data.Lat = float64(lat)
	data.Lon = float64(lon)
	data.HorizontalAcc = float64(horizontal_acc)

	data.Heading = float64(heading)
	data.HeadingOfMotion = float64(headingOfMot)
	data.HeadingAcc = float64(headingAcc)
	data.HeadingIsValid = bool(headingIsValid)

	data.Speed = float64(speed)
	data.SpeedAcc = float64(speedAcc)

	data.Pdop = uint16(pdop)
	data.Hdop = uint16(hdop)
	data.SatNum = uint8(satNum)
	data.FixType = uint8(fixType)

	data.FusionMode = uint8(fusion_mode)
	data.Sensors = C.GoBytes(unsafe.Pointer(sensors), C.int(sensorsLen)*4)

	data.SWVersion = C.GoString(swVersion)
	data.HWVersion = C.GoString(hwVersion)
	data.FWVersion = C.GoString(fwVersion)

	data.Ts = time.Unix(int64(sec), int64(nsec))

	return &data, nil
}

// interrupts Pop call
// note don't use reader after this call
// threadsafe
func (reader *UbloxReader) Close() {
	C.ublox_reader_close(unsafe.Pointer(reader.reader))
}

func (reader *UbloxReader) Free() {
	C.ublox_reader_free(unsafe.Pointer(reader.reader))
}

func (reader *UbloxReader) UpdateVersionInfo() error {
	var c_err *C.char

	if C.ublox_reader_update_version_info(unsafe.Pointer(reader.reader), &c_err) == -1 {
		err := C.GoString(c_err)
		C.free(unsafe.Pointer(c_err))
		return errors.New(err)
	}

	return nil
}
