# rdmamap

This is golang package that provides map of rdma device with its character and network devices.

It uses sysfs and netlink interfaces provided by kernel to perform this mapping.

Local build and test

You can use go get command:
```
go get github.com/Mellanox/rdmamap
```

Example:

```
package main

import (
    "fmt"
    "github.com/Mellanox/rdmamap"
)

func main() {
	rdmaDevices := rdmamap.GetRdmaDeviceList()
	fmt.Println("Devices: ", rdmaDevices)
  
	for _, dev := range rdmaDevices {
		charDevices := rdmamap.GetRdmaCharDevices(dev)
		fmt.Printf("Rdma device: = %s", dev)
		fmt.Println(" Char devices: = ", charDevices)
	}
}

```
