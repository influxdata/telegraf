package appinsights

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Microsoft/ApplicationInsights-Go/appinsights/contracts"
)

type myStringer struct{}

func (s *myStringer) String() string {
	return "My stringer error"
}

type myError struct{}

func (s *myError) Error() string {
	return "My error error"
}

type myGoStringer struct{}

func (s *myGoStringer) Error() string {
	return "My go stringer error"
}

func TestExceptionTelemetry(t *testing.T) {
	// Test callstack capture -- these should all fit in 64 frames.
	for i := 9; i < 20; i++ {
		exd := testExceptionCallstack(t, i)
		checkDataContract(t, "ExceptionDetails.TypeName", exd.TypeName, "string")
		checkDataContract(t, "ExceptionDetails.Message", exd.Message, "Whoops")
		checkDataContract(t, "ExceptionDetails.HasFullStack", exd.HasFullStack, true)
	}

	// Test error types
	var err error
	err = &myError{}

	e1 := catchPanic(err)
	exd1 := e1.TelemetryData().(*contracts.ExceptionData).Exceptions[0]
	checkDataContract(t, "ExceptionDetails.Message", exd1.Message, "My error error")
	checkDataContract(t, "ExceptionDetails.TypeName", exd1.TypeName, "*appinsights.myError")

	e2 := catchPanic(&myStringer{})
	exd2 := e2.TelemetryData().(*contracts.ExceptionData).Exceptions[0]
	checkDataContract(t, "ExceptionDetails.Message", exd2.Message, "My stringer error")
	checkDataContract(t, "ExceptionDetails.TypeName", exd2.TypeName, "*appinsights.myStringer")

	e3 := catchPanic(&myGoStringer{})
	exd3 := e3.TelemetryData().(*contracts.ExceptionData).Exceptions[0]
	checkDataContract(t, "ExceptionDetails.Message", exd3.Message, "My go stringer error")
	checkDataContract(t, "ExceptionDetails.TypeName", exd3.TypeName, "*appinsights.myGoStringer")
}

func TestTrackPanic(t *testing.T) {
	mockClock()
	defer resetClock()
	client, transmitter := newTestChannelServer()
	defer transmitter.Close()

	catchTrackPanic(client, "~exception~")
	client.Channel().Close()

	req := transmitter.waitForRequest(t)
	if !strings.Contains(req.payload, "~exception~") {
		t.Error("Unexpected payload")
	}
}

func testExceptionCallstack(t *testing.T, n int) *contracts.ExceptionDetails {
	d := buildStack(n).TelemetryData().(*contracts.ExceptionData)
	checkDataContract(t, "len(Exceptions)", len(d.Exceptions), 1)
	ex := d.Exceptions[0]

	// Find the relevant range of frames
	frstart := -1
	frend := -1
	for i, f := range ex.ParsedStack {
		if strings.Contains(f.Method, "Collatz") {
			if frstart < 0 {
				frstart = i
			}
		} else {
			if frstart >= 0 && frend < 0 {
				frend = i
				break
			}
		}
	}

	expected := collatzFrames(n)

	if frend-frstart != len(expected) {
		t.Errorf("Wrong number of Collatz frames found.  Got %d, want %d.", frend-frstart, len(expected))
		return ex
	}

	j := len(expected) - 1
	for i := frstart; j >= 0 && i < len(ex.ParsedStack); i++ {
		checkDataContract(t, fmt.Sprintf("ParsedStack[%d].Method", i), ex.ParsedStack[i].Method, expected[j])
		if !strings.HasSuffix(ex.ParsedStack[i].Assembly, "/ApplicationInsights-Go/appinsights") {
			checkDataContract(t, fmt.Sprintf("ParsedStack[%d].Assembly", i), ex.ParsedStack[i].Assembly, "/ApplicationInsights-Go/appinsights")
		}
		if !strings.HasSuffix(ex.ParsedStack[i].FileName, "/exception_test.go") {
			checkDataContract(t, fmt.Sprintf("ParsedStack[%d].FileName", i), ex.ParsedStack[i].FileName, "exception_test.go")
		}

		j--
	}

	return ex
}

func collatzFrames(n int) []string {
	var result []string

	result = append(result, "panicTestCollatz")
	for n != 1 {
		if (n % 2) == 0 {
			result = append(result, "panicTestCollatzEven")
			n /= 2
		} else {
			result = append(result, "panicTestCollatzOdd")
			n = 1 + (3 * n)
		}

		result = append(result, "panicTestCollatz")
	}

	return result
}

func catchPanic(err interface{}) *ExceptionTelemetry {
	var result *ExceptionTelemetry

	func() {
		defer func() {
			if err := recover(); err != nil {
				result = NewExceptionTelemetry(err)
			}
		}()

		panic(err)
	}()

	return result
}

func buildStack(n int) *ExceptionTelemetry {
	var result *ExceptionTelemetry

	// Nest this so the panic doesn't supercede the return
	func() {
		defer func() {
			if err := recover(); err != nil {
				result = NewExceptionTelemetry(err)
			}
		}()

		panicTestCollatz(n)
	}()

	return result
}

// Test Collatz's conjecture for a given input; panic in the base case.
func panicTestCollatz(n int) int {
	if n == 1 {
		panic("Whoops")
	}

	if (n & 1) == 0 {
		return panicTestCollatzEven(n)
	} else {
		return panicTestCollatzOdd(n)
	}
}

func panicTestCollatzEven(n int) int {
	return panicTestCollatz(n / 2)
}

func panicTestCollatzOdd(n int) int {
	return panicTestCollatz((3 * n) + 1)
}

func catchTrackPanic(client TelemetryClient, err interface{}) {
	defer TrackPanic(client, false)
	panic(err)
}
