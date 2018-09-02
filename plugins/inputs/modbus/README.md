# gomodbus Input Plugin

This input plugin is a Fault-tolerant, fail-fast implementation of Modbus protocol in Go.

### Supported functions
-------------------
Bit access:
*   Read Discrete Inputs
*   Read Coils
*   Write Single Coil
*   Write Multiple Coils

16-bit access:
*   Read Input Registers
*   Read Holding Registers
*   Write Single Register
*   Write Multiple Registers
*   Read/Write Multiple Registers
*   Mask Write Register
*   Read FIFO Queue

### Supported formats
-----------------
*   TCP
*   Serial (RTU, ASCII)

### Configuration:

```toml
[[inputs.bond]]
	## Set Modbus Config (Either TCP or RTU Client)
	## Modbust TCP Client
	## TCP Client = "localhost:502"
	Client = "localhost:502"

	## Modbus RTU Client
	## RTU Client = "/dev/ttyS0"
	## serial setup for RTUClient
	# serial = [11520,8,"N",1]

	## Call to device
	SlaveAddress = 1

	## Function Code to Device
	FunctionCode = 1

	## Device Memory Address
	Address = 1

	## Quantity of Values to read/write
	Quantity = 1

	## Array of values to write
	# Values = [0]

	# Timeout in seconds
	TimeOut = 5
```
Run:

```
telegraf --config telegraf.conf --input-filter modbus --test
```

### References
----------
-   [Modbus Specifications and Implementation Guides](http://www.modbus.org/specs.php)
