[![Build Status](https://travis-ci.org/tbrandon/mbserver.svg?branch=master)](https://travis-ci.org/tbrandon/mbserver)
[![Coverage Status](http://codecov.io/github/tbrandon/mbserver/coverage.svg?branch=master)](http://codecov.io/github/tbrandon/mbserver?branch=master)
[![GoDoc](https://godoc.org/github.com/tbrandon/mbserver?status.svg)](https://godoc.org/github.com/tbrandon/mbserver)
[![Software License](https://img.shields.io/badge/License-MIT-green.svg)](https://github.com/tbrandon/mbserver/blob/master/LICENSE)

# Golang Modbus Server (Slave)

The Golang Modbus Server (Slave) responds to the following Modbus function requests:

Bit access:
- Read Discrete Inputs
- Read Coils
- Write Single Coil
- Write Multiple Coils

16-bit acess:
- Read Input Registers
- Read Multiple Holding Registers
- Write Single Holding Register
- Write Multiple Holding Registers

TCP and serial RTU access is supported.

The server internally allocates memory for 65536 coils, 65536 discrete inputs, 653356 holding registers and 65536 input registers.
On start, all values are initialzied to zero.  Modbus requests are processed in the order they are received and will not overlap/interfere with each other.

The golang [mbserver documentation](https://godoc.org/github.com/tbrandon/mbserver).

## Example Modbus TCP Server

Create a Modbus TCP Server (Slave):

```
package main

import (
	"log"
	"time"

	"github.com/tbrandon/mbserver"
)

func main() {
	serv := mbserver.NewServer()
	err := serv.ListenTCP("127.0.0.1:1502")
	if err != nil {
		log.Printf("%v\n", err)
	}
	defer serv.Close()

	// Wait forever
	for {
		time.Sleep(1 * time.Second)
	}
}
```
The server will continue to listen until killed (&lt;ctrl>-c).
Modbus typically uses port 502 (standard users require special permissions to listen on port 502). Change the port number as required.
Change the address to 0.0.0.0 to listen on all network interfaces.

An example of a client writing and reading holding regsiters:
```
package main

import (
	"fmt"

	"github.com/goburrow/modbus"
)

func main() {
	handler := modbus.NewTCPClientHandler("localhost:1502")
	// Connect manually so that multiple requests are handled in one session
	err := handler.Connect()
	defer handler.Close()
	client := modbus.NewClient(handler)

	_, err = client.WriteMultipleRegisters(0, 3, []byte{0, 3, 0, 4, 0, 5})
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	results, err := client.ReadHoldingRegisters(0, 3)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	fmt.Printf("results %v\n", results)
}

Outputs:
results [0 3 0 4 0 5]
```

## Example Listening on Multiple TCP Ports and Serial Devices

The Golang Modbus Server can listen on multiple TCP ports and serial devices.
In the following example, the Modbus server will be configured to listen on
127.0.0.1:1502, 0.0.0.0:3502, /dev/ttyUSB0 and /dev/ttyACM0

```
	serv := mbserver.NewServer()
	err := serv.ListenTCP("127.0.0.1:1502")
	if err != nil {
		log.Printf("%v\n", err)
	}

	err := serv.ListenTCP("0.0.0.0:3502")
	if err != nil {
		log.Printf("%v\n", err)
	}

	err := s.ListenRTU(&serial.Config{
		Address:  "/dev/ttyUSB0",
		BaudRate: 115200,
		DataBits: 8,
		StopBits: 1,
		Parity:   "N",
		Timeout:  10 * time.Second})
	if err != nil {
		t.Fatalf("failed to listen, got %v\n", err)
	}

	err := s.ListenRTU(&serial.Config{
		Address:  "/dev/ttyACM0",
		BaudRate: 9600,
		DataBits: 8,
		StopBits: 1,
		Parity:   "N",
		Timeout:  10 * time.Second,
		RS485: serial.RS485Config{
			Enabled: true,
			DelayRtsBeforeSend: 2 * time.Millisecond
			DelayRtsAfterSend: 3 * time.Millisecond
			RtsHighDuringSend: false,
			RtsHighAfterSend: false,
			RxDuringTx: false
			})
	if err != nil {
		t.Fatalf("failed to listen, got %v\n", err)
	}

	defer serv.Close()
```

Information on [serial port settings](https://godoc.org/github.com/goburrow/serial).

## Server Customization

 RegisterFunctionHandler allows the default server functionality to be overridden for a Modbus function code.
 ```
func (s *Server) RegisterFunctionHandler(funcCode uint8, function func(*Server, Framer) ([]byte, *Exception))
 ```

Example of overriding the default ReadDiscreteInputs funtion:

```
serv := NewServer()

// Override ReadDiscreteInputs function.
serv.RegisterFunctionHandler(2,
    func(s *Server, frame Framer) ([]byte, *Exception) {
        register, numRegs, endRegister := frame.registerAddressAndNumber()
        // Check the request is within the allocated memory
        if endRegister > 65535 {
            return []byte{}, &IllegalDataAddress
        }
        dataSize := numRegs / 8
        if (numRegs % 8) != 0 {
            dataSize++
        }
        data := make([]byte, 1+dataSize)
        data[0] = byte(dataSize)
        for i := range s.DiscreteInputs[register:endRegister] {
            // Return all 1s, regardless of the value in the DiscreteInputs array.
            shift := uint(i) % 8
            data[1+i/8] |= byte(1 << shift)
        }
        return data, &Success
    })

// Start the server.
err := serv.ListenTCP("localhost:4321")
if err != nil {
    log.Printf("%v\n", err)
    return
}
defer serv.Close()

// Wait for the server to start
time.Sleep(1 * time.Millisecond)

// Example of a client reading from the server started above.
// Connect a client.
handler := modbus.NewTCPClientHandler("localhost:4321")
err = handler.Connect()
if err != nil {
    log.Printf("%v\n", err)
    return
}
defer handler.Close()
client := modbus.NewClient(handler)

// Read discrete inputs.
results, err := client.ReadDiscreteInputs(0, 16)
if err != nil {
    log.Printf("%v\n", err)
}

fmt.Printf("results %v\n", results)
```
Output:
```
results [255 255]
```

## Benchmarks

Quanitify server read/write performance.  Benchmarks are for Modbus TCP operations.

Run benchmarks:
```
$ go test -bench=.
BenchmarkModbusWrite1968MultipleCoils-8            50000             30912 ns/op
BenchmarkModbusRead2000Coils-8                     50000             27875 ns/op
BenchmarkModbusRead2000DiscreteInputs-8            50000             27335 ns/op
BenchmarkModbusWrite123MultipleRegisters-8        100000             22655 ns/op
BenchmarkModbusRead125HoldingRegisters-8          100000             21117 ns/op
PASS
```
Operations per second are higher when requests are not forced to be  synchronously processed.
In the case of simultaneous client access, synchronous Modbus request processing prevents data corruption.

To understand performanc limitations, create a CPU profile graph for the WriteMultipleCoils benchmark:
```
go test -bench=.MultipleCoils -cpuprofile=cpu.out
go tool pprof modbus-server.test cpu.out
(pprof) web
```

## Race Conditions

There is a [known](https://github.com/golang/go/issues/10001) race condition in the code relating to calling Serial Read() and Close() functions in different go routines.

To check for race conditions, run:
```
go test --race
```
