package rdmamap

import (
	"bytes"
	"fmt"
	"github.com/vishvananda/netlink"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	RdmaClassName     = "infiniband"
	RdmaClassDir      = "/sys/class/infiniband"
	RdmaIbUcmDir      = "/sys/class/infiniband_cm"
	RdmaUcmFilePrefix = "ucm"

	RdmaUmadDir        = "/sys/class/infiniband_mad"
	RdmaIssmFilePrefix = "issm"
	RdmaUmadFilxPrefix = "umad"

	RdmaUverbsDir        = "/sys/class/infiniband_verbs"
	RdmaUverbsFilxPrefix = "uverbs"

	RdmaGidAttrDir     = "gid_attrs"
	RdmaGidAttrNdevDir = "ndevs"
	RdmaPortsdir       = "ports"

	RdmaNodeGuidFile = "node_guid"
	RdmaUcmDevice    = "/dev/infiniband/rdma_cm"
	RdmaDeviceDir    = "/dev/infiniband"

	RdmaCountersDir   = "counters"
	RdmaHwCountersDir = "hw_counters"

	PciDevDir = "/sys/bus/pci/devices"
)

// Returns a list of rdma device names
func GetRdmaDeviceList() []string {
	var rdmaDevices []string
	fd, err := os.Open(RdmaClassDir)
	if err != nil {
		return nil
	}
	fileInfos, err := fd.Readdir(-1)
	defer fd.Close()

	for i := range fileInfos {
		if fileInfos[i].IsDir() {
			continue
		}
		rdmaDevices = append(rdmaDevices, fileInfos[i].Name())
	}
	return rdmaDevices
}

func isDirForRdmaDevice(rdmaDeviceName string, dirName string) bool {
	fileName := filepath.Join(dirName, "ibdev")

	fd, err := os.OpenFile(fileName, os.O_RDONLY, 0444)
	if err != nil {
		return false
	}
	defer fd.Close()

	fd.Seek(0, os.SEEK_SET)
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), rdmaDeviceName)
}

func getCharDevice(rdmaDeviceName string, classDir string,
	charDevPrefix string) (string, error) {
	fd, err := os.Open(classDir)
	if err != nil {
		return "", err
	}
	fileInfos, err := fd.Readdir(-1)
	defer fd.Close()

	for i := range fileInfos {
		if fileInfos[i].Name() == "." || fileInfos[i].Name() == ".." {
			continue
		}
		if strings.Contains(fileInfos[i].Name(), charDevPrefix) == false {
			continue
		}
		dirName := filepath.Join(classDir, fileInfos[i].Name())
		if isDirForRdmaDevice(rdmaDeviceName, dirName) == false {
			continue
		}
		deviceFile := filepath.Join("/dev/infiniband", fileInfos[i].Name())
		return deviceFile, nil
	}
	return "", fmt.Errorf("No ucm device found")

}

func getUcmDevice(rdmaDeviceName string) (string, error) {
	return getCharDevice(rdmaDeviceName,
		RdmaIbUcmDir,
		RdmaUcmFilePrefix)
}

func getIssmDevice(rdmaDeviceName string) (string, error) {

	return getCharDevice(rdmaDeviceName,
		RdmaUmadDir,
		RdmaIssmFilePrefix)
}

func getUmadDevice(rdmaDeviceName string) (string, error) {

	return getCharDevice(rdmaDeviceName,
		RdmaUmadDir,
		RdmaUmadFilxPrefix)
}

func getUverbDevice(rdmaDeviceName string) (string, error) {

	return getCharDevice(rdmaDeviceName,
		RdmaUverbsDir,
		RdmaUverbsFilxPrefix)
}

func getRdmaUcmDevice() (string, error) {
	info, err := os.Stat(RdmaUcmDevice)
	if err != nil {
		return "", err
	}
	if info.Name() == "rdma_cm" {
		return RdmaUcmDevice, nil
	} else {
		return "", fmt.Errorf("Invalid file name rdma_cm")
	}
}

// Returns a list of character device absolute path for a requested
// rdmaDeviceName.
// Returns nil if no character devices are found.
func GetRdmaCharDevices(rdmaDeviceName string) []string {

	var rdmaCharDevices []string

	ucm, err := getUcmDevice(rdmaDeviceName)
	if err == nil {
		rdmaCharDevices = append(rdmaCharDevices, ucm)
	}
	issm, err := getIssmDevice(rdmaDeviceName)
	if err == nil {
		rdmaCharDevices = append(rdmaCharDevices, issm)
	}
	umad, err := getUmadDevice(rdmaDeviceName)
	if err == nil {
		rdmaCharDevices = append(rdmaCharDevices, umad)
	}
	uverb, err := getUverbDevice(rdmaDeviceName)
	if err == nil {
		rdmaCharDevices = append(rdmaCharDevices, uverb)
	}
	rdma_cm, err := getRdmaUcmDevice()
	if err == nil {
		rdmaCharDevices = append(rdmaCharDevices, rdma_cm)
	}

	return rdmaCharDevices
}

// Gets a list of ports for a specified device
func GetPorts(rdmaDeviceName string) []string {
	var ports []string

	portsDir := filepath.Join(RdmaClassDir, rdmaDeviceName, RdmaPortsdir)
	fd, err := os.Open(portsDir)
	if err != nil {
		return nil
	}
	fileInfos, err := fd.Readdir(-1)
	defer fd.Close()

	for i := range fileInfos {
		if fileInfos[i].Name() == "." || fileInfos[i].Name() == ".." {
			continue
		}
		ports = append(ports, fileInfos[i].Name())
	}
	return ports
}

func getNetdeviceIds(rdmaDeviceName string, port string) []string {
	var indices []string

	dir := filepath.Join(RdmaClassDir, rdmaDeviceName, RdmaPortsdir, port,
		RdmaGidAttrDir, RdmaGidAttrNdevDir)

	fd, err := os.Open(dir)
	if err != nil {
		return nil
	}
	fileInfos, err := fd.Readdir(-1)
	defer fd.Close()

	for i := range fileInfos {
		if fileInfos[i].Name() == "." || fileInfos[i].Name() == ".." {
			continue
		}
		indices = append(indices, fileInfos[i].Name())
	}
	return indices
}

func isNetdevForRdma(rdmaDeviceName string, port string,
	index string, netdevName string) bool {

	fileName := filepath.Join(RdmaClassDir, rdmaDeviceName, RdmaPortsdir, port,
		RdmaGidAttrDir, RdmaGidAttrNdevDir, index)

	fd, err := os.OpenFile(fileName, os.O_RDONLY, 0444)
	if err != nil {
		return false
	}
	defer fd.Close()

	fd.Seek(0, os.SEEK_SET)
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return false
	}
	if strings.TrimSuffix(string(data), "\n") == netdevName {
		return true
	} else {
		return false
	}
}

func getRdmaDeviceForEth(netdevName string) (string, error) {
	// Iterate over the list of rdma devices,
	// read the gid table attribute netdev
	// if the netdev matches, found the matching rdma device

	devices := GetRdmaDeviceList()
	for _, dev := range devices {
		ports := GetPorts(dev)
		for _, port := range ports {
			indices := getNetdeviceIds(dev, port)
			for _, index := range indices {
				found := isNetdevForRdma(dev, port, index, netdevName)
				if found == true {
					return dev, nil
				}
			}
		}
	}
	return "", fmt.Errorf("rdma device not found for netdev %v", netdevName)
}

func getNodeGuid(rdmaDeviceName string) ([]byte, error) {
	var nodeGuid []byte

	fileName := filepath.Join(RdmaClassDir, rdmaDeviceName, RdmaNodeGuidFile)

	fd, err := os.OpenFile(fileName, os.O_RDONLY, 0444)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	fd.Seek(0, os.SEEK_SET)
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	data = data[:len(data)-1]
	var j int
	for _, b := range data {
		if b == ':' {
			continue
		}
		c, err := strconv.ParseUint(string(b), 16, 8)
		if err != nil {
			return nil, err
		}
		if (j % 2) == 0 {
			nodeGuid = append(nodeGuid, byte(c)<<4)
		} else {
			nodeGuid[j/2] |= byte(c)
		}
		j++
	}
	return nodeGuid, nil
}

func getRdmaDeviceForIb(netdevName string, linkAttr *netlink.LinkAttrs) (string, error) {
	// Match the node_guid EUI bytes with the IpoIB netdevice hw address EUI

	lleui64 := linkAttr.HardwareAddr[12:]

	devices := GetRdmaDeviceList()
	for _, dev := range devices {
		nodeGuid, err := getNodeGuid(dev)
		if err != nil {
			return "", err
		}
		if bytes.Compare(lleui64, nodeGuid) == 0 {
			return dev, nil
		}
	}
	return "", nil
}

//Get RDMA device for the netdevice
func GetRdmaDeviceForNetdevice(netdevName string) (string, error) {

	handle, err := netlink.LinkByName(netdevName)
	if err != nil {
		return "", err
	}
	netAttr := handle.Attrs()
	if netAttr.EncapType == "ether" {
		return getRdmaDeviceForEth(netdevName)
	} else if netAttr.EncapType == "infiniband" {
		return getRdmaDeviceForIb(netdevName, netAttr)
	} else {
		return "", fmt.Errorf("Unknown device type")
	}
}

//Returns true if rdma device exist for netdevice, else false
func IsRDmaDeviceForNetdevice(netdevName string) bool {
	rdma, _ := GetRdmaDeviceForNetdevice(netdevName)
	if rdma == "" {
		return false
	} else {
		return true
	}
}

//Get list of RDMA devices for a pci device.
//When switchdev mode is used, there may be more than one rdma device.
//Example pcidevName: 0000:05:00:00,
//when found, returns list of devices one or more devices names such as
//mlx5_0, mlx5_10
func GetRdmaDevicesForPcidev(pcidevName string) []string {

	var rdmadevs []string

	dirName := filepath.Join(PciDevDir, pcidevName, RdmaClassName)

	entries, err := ioutil.ReadDir(dirName)
	if err != nil {
		return rdmadevs
	}

	for _, entry := range entries {
		if entry.IsDir() == false {
			continue
		}
		rdmadevs = append(rdmadevs, entry.Name())
	}
	return rdmadevs
}
