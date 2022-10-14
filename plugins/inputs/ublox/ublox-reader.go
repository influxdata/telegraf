package ublox

/*
#cgo LDFLAGS: -L./ublox-utils/build -lublox-utils
#include <stdlib.h>
#include <stdbool.h>
void *ublox_reader_new();
void ublox_reader_free(void *reader);
bool ublox_reader_init(void *reader, const char *device, char **err);
void ublox_reader_close(void *reader);
int ublox_reader_read(void *reader, bool *is_active, double *lat, double *lon, double *heading, double *pdop, bool wait_for_data, char **err);
*/
import "C"
import (
	"errors"
	"unsafe"
)

type GPSPos struct {
	Active  bool
	Lat     float64
	Lon     float64
	Heading float64
	Pdop    uint16
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
	var heading C.double
	var pdop C.double
	status := C.ublox_reader_read(unsafe.Pointer(reader.reader), &is_active, &lat, &lon, &heading, &pdop, C.bool(wait_for_data), &c_err)
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
	data.Heading = float64(heading)
	data.Pdop = uint16(pdop)

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
