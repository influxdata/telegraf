package main

// #include "../../kflow.h"
import "C"
import (
	"fmt"
	"net"
	"net/url"
	"os/signal"
	"reflect"
	"syscall"
	"time"
	"unsafe"

	"github.com/kentik/libkflow"
	"github.com/kentik/libkflow/api"
	"github.com/kentik/libkflow/flow"
)

var sender *libkflow.Sender
var errors chan error

//export kflowInit
func kflowInit(cfg *C.kflowConfig, customs **C.kflowCustom, n *C.uint32_t) C.int {
	errors = make(chan error, 100)

	if cfg == nil {
		return C.EKFLOWCONFIG
	}

	flowurl, err := url.Parse(C.GoString(cfg.URL))
	if err != nil {
		fail("invalid flow URL: %s", err)
		return C.EKFLOWCONFIG
	}

	apiurl, err := url.Parse(C.GoString(cfg.API.URL))
	if err != nil {
		fail("invalid API URL: %s", err)
		return C.EKFLOWCONFIG
	}

	metricsurl, err := url.Parse(C.GoString(cfg.metrics.URL))
	if err != nil {
		fail("invalid metrics URL: %s", err)
		return C.EKFLOWCONFIG
	}

	var (
		email   = C.GoString(cfg.API.email)
		token   = C.GoString(cfg.API.token)
		timeout = time.Duration(cfg.timeout) * time.Millisecond
		program = C.GoString(cfg.program)
		version = C.GoString(cfg.version)
		proxy   *url.URL
	)

	if program == "" || version == "" {
		return C.EKFLOWCONFIG
	}

	if cfg.proxy.URL != nil {
		proxy, err = url.Parse(C.GoString(cfg.proxy.URL))
		if err != nil {
			fail("invalid proxy URL: %s", err)
			return C.EKFLOWCONFIG
		}
	}

	config := libkflow.NewConfig(email, token, program, version)
	config.SetCapture(libkflow.Capture{
		Device:  C.GoString(cfg.capture.device),
		Snaplen: int32(cfg.capture.snaplen),
		Promisc: cfg.capture.promisc == 1,
	})
	config.SetProxy(proxy)
	config.SetTimeout(timeout)
	config.SetVerbose(int(cfg.verbose))
	config.OverrideURLs(apiurl, flowurl, metricsurl)

	switch {
	case cfg.device_id > 0:
		did := int(cfg.device_id)
		sender, err = libkflow.NewSenderWithDeviceID(did, errors, config)
	case cfg.device_if != nil:
		dif := C.GoString(cfg.device_if)
		sender, err = libkflow.NewSenderWithDeviceIF(dif, errors, config)
	case cfg.device_ip != nil:
		dip := net.ParseIP(C.GoString(cfg.device_ip))
		sender, err = libkflow.NewSenderWithDeviceIP(dip, errors, config)
	default:
		fail("no device identifier supplied")
		return C.EKFLOWCONFIG
	}

	if err != nil {
		switch err {
		case libkflow.ErrInvalidAuth:
			return C.EKFLOWAUTH
		case libkflow.ErrInvalidDevice:
			return C.EKFLOWNODEVICE
		default:
			fail("library setup error: %s", err)
			return C.EKFLOWCONFIG
		}
	}

	populateCustoms(sender.Device, customs, n)

	signal.Ignore(syscall.SIGPIPE)

	return 0
}

//export kflowSend
func kflowSend(cflow *C.kflow) C.int {
	if sender == nil {
		return C.EKFLOWNOINIT
	}

	ckflow := (*flow.Ckflow)(unsafe.Pointer(cflow))
	flow := flow.New(ckflow)
	sender.Send(&flow)

	return 0
}

//export kflowStop
func kflowStop(msec C.int) C.int {
	if sender == nil {
		return C.EKFLOWNOINIT
	}

	wait := time.Duration(msec) * time.Millisecond
	if !sender.Stop(wait) {
		return C.EKFLOWTIMEOUT
	}
	return 0
}

//export kflowError
func kflowError() *C.char {
	select {
	case err := <-errors:
		return C.CString(err.Error())
	default:
		return nil
	}
}

//export kflowVersion
func kflowVersion() *C.char {
	return C.CString(libkflow.Version)
}

func populateCustoms(device *api.Device, ptr **C.kflowCustom, cnt *C.uint32_t) {
	if ptr == nil || cnt == nil {
		return
	}

	n := len(device.Customs)
	*ptr = (*C.kflowCustom)(C.calloc(C.size_t(n), C.sizeof_kflowCustom))
	*cnt = C.uint32_t(n)

	customs := *(*[]C.kflowCustom)(unsafe.Pointer(&reflect.SliceHeader{
		Data: (uintptr)(unsafe.Pointer(*ptr)),
		Len:  int(n),
		Cap:  int(n),
	}))

	for i, c := range device.Customs {
		var vtype C.int
		switch c.Type {
		case "string":
			vtype = C.KFLOWCUSTOMSTR
		case "uint32":
			vtype = C.KFLOWCUSTOMU32
		case "float32":
			vtype = C.KFLOWCUSTOMF32
		}

		customs[i] = C.kflowCustom{
			id:    C.uint64_t(c.ID),
			name:  C.CString(c.Name),
			vtype: vtype,
		}
	}
}

func fail(format string, args ...interface{}) {
	errors <- fmt.Errorf(format, args...)
}

func main() {
}

const (
	EKFLOWCONFIG   = C.EKFLOWCONFIG
	EKFLOWNOINIT   = C.EKFLOWNOINIT
	EKFLOWNOMEM    = C.EKFLOWNOMEM
	EKFLOWTIMEOUT  = C.EKFLOWTIMEOUT
	EKFLOWSEND     = C.EKFLOWSEND
	EKFLOWNOCUSTOM = C.EKFLOWNOCUSTOM
	EKFLOWAUTH     = C.EKFLOWAUTH
	EKFLOWNODEVICE = C.EKFLOWNODEVICE
)
