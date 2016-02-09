package snmp

import (
	"io/ioutil"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/soniah/gosnmp"
)

// Snmp is a snmp plugin
type Snmp struct {
	Host              []Host
	Get               []Data
	Bulk              []Data
	SnmptranslateFile string
}

type Host struct {
	Address   string
	Community string
	// SNMP version. Default 2
	Version int
	// SNMP timeout, in seconds. 0 means no timeout
	Timeout float64
	// SNMP retries
	Retries int
	// Data to collect (list of Data names)
	Collect []string
	// easy get oids
	GetOids []string
	// Oids
	getOids  []Data
	bulkOids []Data
}

type Data struct {
	Name string
	// OID (could be numbers or name)
	Oid string
	// Unit
	Unit string
	//  SNMP getbulk max repetition
	MaxRepetition uint8 `toml:"max_repetition"`
	// SNMP Instance (default 0)
	// (only used with  GET request and if
	//  OID is a name from snmptranslate file)
	Instance string
	// OID (only number) (used for computation)
	rawOid string
}

type Node struct {
	id       string
	name     string
	subnodes map[string]Node
}

var initNode = Node{
	id:       "1",
	name:     "",
	subnodes: make(map[string]Node),
}

var NameToOid = make(map[string]string)

var sampleConfig = `
  ### Use 'oids.txt' file to translate oids to names
  ### To generate 'oids.txt' you need to run:
  ###   snmptranslate -m all -Tz -On | sed -e 's/"//g' > /tmp/oids.txt
  ### Or if you have an other MIB folder with custom MIBs
  ###   snmptranslate -M /mycustommibfolder -Tz -On -m all | sed -e 's/"//g' > oids.txt
  snmptranslate_file = "/tmp/oids.txt"
  [[inputs.snmp.host]]
    address = "192.168.2.2:161"
    # SNMP community
    community = "public" # default public
    # SNMP version (1, 2 or 3)
    # Version 3 not supported yet
    version = 2 # default 2
    # SNMP response timeout
    timeout = 2.0 # default 2.0
    # SNMP request retries
    retries = 2 # default 2
    # Which get/bulk do you want to collect for this host
    collect = ["mybulk", "sysservices", "sysdescr"]
    # Simple list of OIDs to get, in addition to "collect"
    get_oids = []

  [[inputs.snmp.host]]
    address = "192.168.2.3:161"
    community = "public"
    version = 2
    timeout = 2.0
    retries = 2
    collect = ["mybulk"]
    get_oids = [
        "ifNumber",
        ".1.3.6.1.2.1.1.3.0",
    ]

  [[inputs.snmp.get]]
    name = "ifnumber"
    oid = "ifNumber"

  [[inputs.snmp.get]]
    name = "interface_speed"
    oid = "ifSpeed"
    instance = 0

  [[inputs.snmp.get]]
    name = "sysuptime"
    oid = ".1.3.6.1.2.1.1.3.0"
    unit = "second"

  [[inputs.snmp.bulk]]
    name = "mybulk"
    max_repetition = 127
    oid = ".1.3.6.1.2.1.1"

  [[inputs.snmp.bulk]]
    name = "ifoutoctets"
    max_repetition = 127
    oid = "ifOutOctets"
`

// SampleConfig returns sample configuration message
func (s *Snmp) SampleConfig() string {
	return sampleConfig
}

// Description returns description of Zookeeper plugin
func (s *Snmp) Description() string {
	return `Reads oids value from one or many snmp agents`
}

func fillnode(parentNode Node, oid_name string, ids []string) {
	// ids = ["1", "3", "6", ...]
	id, ids := ids[0], ids[1:]
	node, ok := parentNode.subnodes[id]
	if ok == false {
		node = Node{
			id:       id,
			name:     "",
			subnodes: make(map[string]Node),
		}
		if len(ids) == 0 {
			node.name = oid_name
		}
		parentNode.subnodes[id] = node
	}
	if len(ids) > 0 {
		fillnode(node, oid_name, ids)
	}
}

func findnodename(node Node, ids []string) (string, string) {
	// ids = ["1", "3", "6", ...]
	if len(ids) == 1 {
		return node.name, ids[0]
	}
	id, ids := ids[0], ids[1:]
	// Get node
	subnode, ok := node.subnodes[id]
	if ok {
		return findnodename(subnode, ids)
	}
	// We got a node
	// Get node name
	if node.name != "" && len(ids) == 0 && id == "0" {
		// node with instance 0
		return node.name, "0"
	} else if node.name != "" && len(ids) == 0 && id != "0" {
		// node with an instance
		return node.name, string(id)
	} else if node.name != "" && len(ids) > 0 {
		// node with subinstances
		return node.name, strings.Join(ids, ".")
	}
	// return an empty node name
	return node.name, ""
}

func (s *Snmp) Gather(acc telegraf.Accumulator) error {
	// Create oid tree
	if s.SnmptranslateFile != "" && len(initNode.subnodes) == 0 {
		data, err := ioutil.ReadFile(s.SnmptranslateFile)
		if err != nil {
			log.Printf("Reading SNMPtranslate file error: %s", err)
			return err
		} else {
			for _, line := range strings.Split(string(data), "\n") {
				oidsRegEx := regexp.MustCompile(`([^\t]*)\t*([^\t]*)`)
				oids := oidsRegEx.FindStringSubmatch(string(line))
				if oids[2] != "" {
					oid_name := oids[1]
					oid := oids[2]
					fillnode(initNode, oid_name, strings.Split(string(oid), "."))
					NameToOid[oid_name] = oid
				}
			}
		}
	}
	// Fetching data
	for _, host := range s.Host {
		// Set default args
		if len(host.Address) == 0 {
			host.Address = "127.0.0.1:161"
		}
		if host.Community == "" {
			host.Community = "public"
		}
		if host.Timeout <= 0 {
			host.Timeout = 2.0
		}
		if host.Retries <= 0 {
			host.Retries = 2
		}
		// Prepare host
		// Get Easy GET oids
		for _, oidstring := range host.GetOids {
			oid := Data{}
			if val, ok := NameToOid[oidstring]; ok {
				// TODO should we add the 0 instance ?
				oid.Name = oidstring
				oid.Oid = val
				oid.rawOid = "." + val + ".0"
			} else {
				oid.Name = oidstring
				oid.Oid = oidstring
				if string(oidstring[:1]) != "." {
					oid.rawOid = "." + oidstring
				} else {
					oid.rawOid = oidstring
				}
			}
			host.getOids = append(host.getOids, oid)
		}

		for _, oid_name := range host.Collect {
			// Get GET oids
			for _, oid := range s.Get {
				if oid.Name == oid_name {
					if val, ok := NameToOid[oid.Oid]; ok {
						// TODO should we add the 0 instance ?
						if oid.Instance != "" {
							oid.rawOid = "." + val + "." + oid.Instance
						} else {
							oid.rawOid = "." + val + ".0"
						}
					} else {
						oid.rawOid = oid.Oid
					}
					host.getOids = append(host.getOids, oid)
				}
			}
			// Get GETBULK oids
			for _, oid := range s.Bulk {
				if oid.Name == oid_name {
					if val, ok := NameToOid[oid.Oid]; ok {
						oid.rawOid = "." + val
					} else {
						oid.rawOid = oid.Oid
					}
					host.bulkOids = append(host.bulkOids, oid)
				}
			}
		}
		// Launch Get requests
		if err := host.SNMPGet(acc); err != nil {
			return err
		}
		if err := host.SNMPBulk(acc); err != nil {
			return err
		}
	}
	return nil
}

func (h *Host) SNMPGet(acc telegraf.Accumulator) error {
	// Get snmp client
	snmpClient, err := h.GetSNMPClient()
	if err != nil {
		return err
	}
	// Deconnection
	defer snmpClient.Conn.Close()
	// Prepare OIDs
	oidsList := make(map[string]Data)
	for _, oid := range h.getOids {
		oidsList[oid.rawOid] = oid
	}
	oidsNameList := make([]string, 0, len(oidsList))
	for _, oid := range oidsList {
		oidsNameList = append(oidsNameList, oid.rawOid)
	}

	// gosnmp.MAX_OIDS == 60
	// TODO use gosnmp.MAX_OIDS instead of hard coded value
	max_oids := 60
	// limit 60 (MAX_OIDS) oids by requests
	for i := 0; i < len(oidsList); i = i + max_oids {
		// Launch request
		max_index := i + max_oids
		if i+max_oids > len(oidsList) {
			max_index = len(oidsList)
		}
		result, err3 := snmpClient.Get(oidsNameList[i:max_index]) // Get() accepts up to g.MAX_OIDS
		if err3 != nil {
			return err3
		}
		// Handle response
		_, err = h.HandleResponse(oidsList, result, acc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Host) SNMPBulk(acc telegraf.Accumulator) error {
	// Get snmp client
	snmpClient, err := h.GetSNMPClient()
	if err != nil {
		return err
	}
	// Deconnection
	defer snmpClient.Conn.Close()
	// Prepare OIDs
	oidsList := make(map[string]Data)
	for _, oid := range h.bulkOids {
		oidsList[oid.rawOid] = oid
	}
	oidsNameList := make([]string, 0, len(oidsList))
	for _, oid := range oidsList {
		oidsNameList = append(oidsNameList, oid.rawOid)
	}
	// TODO Trying to make requests with more than one OID
	// to reduce the number of requests
	for _, oid := range oidsNameList {
		oid_asked := oid
		need_more_requests := true
		// Set max repetition
		maxRepetition := oidsList[oid].MaxRepetition
		if maxRepetition <= 0 {
			maxRepetition = 32
		}
		// Launch requests
		for need_more_requests {
			// Launch request
			result, err3 := snmpClient.GetBulk([]string{oid}, 0, maxRepetition)
			if err3 != nil {
				return err3
			}
			// Handle response
			last_oid, err := h.HandleResponse(oidsList, result, acc)
			if err != nil {
				return err
			}
			// Determine if we need more requests
			if strings.HasPrefix(last_oid, oid_asked) {
				need_more_requests = true
				oid = last_oid
			} else {
				need_more_requests = false
			}
		}
	}
	return nil
}

func (h *Host) GetSNMPClient() (*gosnmp.GoSNMP, error) {
	// Prepare Version
	var version gosnmp.SnmpVersion
	if h.Version == 1 {
		version = gosnmp.Version1
	} else if h.Version == 3 {
		version = gosnmp.Version3
	} else {
		version = gosnmp.Version2c
	}
	// Prepare host and port
	host, port_str, err := net.SplitHostPort(h.Address)
	if err != nil {
		port_str = string("161")
	}
	// convert port_str to port in uint16
	port_64, err := strconv.ParseUint(port_str, 10, 16)
	port := uint16(port_64)
	// Get SNMP client
	snmpClient := &gosnmp.GoSNMP{
		Target:    host,
		Port:      port,
		Community: h.Community,
		Version:   version,
		Timeout:   time.Duration(h.Timeout) * time.Second,
		Retries:   h.Retries,
	}
	// Connection
	err2 := snmpClient.Connect()
	if err2 != nil {
		return nil, err2
	}
	// Return snmpClient
	return snmpClient, nil
}

func (h *Host) HandleResponse(oids map[string]Data, result *gosnmp.SnmpPacket, acc telegraf.Accumulator) (string, error) {
	var lastOid string
	for _, variable := range result.Variables {
		lastOid = variable.Name
		// Remove unwanted oid
		for oid_key, oid := range oids {
			if strings.HasPrefix(variable.Name, oid_key) {
				switch variable.Type {
				// handle Metrics
				case gosnmp.Boolean, gosnmp.Integer, gosnmp.Counter32, gosnmp.Gauge32,
					gosnmp.TimeTicks, gosnmp.Counter64, gosnmp.Uinteger32:
					// Prepare tags
					tags := make(map[string]string)
					if oid.Unit != "" {
						tags["unit"] = oid.Unit
					}
					// Get name and instance
					var oid_name string
					var instance string
					// Get oidname and instannce from translate file
					oid_name, instance = findnodename(initNode,
						strings.Split(string(variable.Name[1:]), "."))

					if instance != "" {
						tags["instance"] = instance
					}

					// Set name
					var field_name string
					if oid_name != "" {
						// Set fieldname as oid name from translate file
						field_name = oid_name
					} else {
						// Set fieldname as oid name from inputs.snmp.get section
						// Because the result oid is equal to inputs.snmp.get section
						field_name = oid.Name
					}
					tags["host"], _, _ = net.SplitHostPort(h.Address)
					fields := make(map[string]interface{})
					fields[string(field_name)] = variable.Value

					acc.AddFields(field_name, fields, tags)
				case gosnmp.NoSuchObject, gosnmp.NoSuchInstance:
					// Oid not found
					log.Printf("[snmp input] Oid not found: %s", oid_key)
				default:
					// delete other data
				}
				break
			}
		}
	}
	return lastOid, nil
}

func init() {
	inputs.Add("snmp", func() telegraf.Input {
		return &Snmp{}
	})
}
