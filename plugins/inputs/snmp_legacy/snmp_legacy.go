package snmp_legacy

import (
	"io/ioutil"
	"log"
	"net"
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
	Table             []Table
	Subtable          []Subtable
	SnmptranslateFile string

	Log telegraf.Logger

	nameToOid   map[string]string
	initNode    Node
	subTableMap map[string]Subtable
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
	// Table
	Table []HostTable
	// Oids
	getOids  []Data
	bulkOids []Data
	tables   []HostTable
	// array of processed oids
	// to skip oid duplication
	processedOids []string

	OidInstanceMapping map[string]map[string]string
}

type Table struct {
	// name = "iftable"
	Name string
	// oid = ".1.3.6.1.2.1.31.1.1.1"
	Oid string
	//if empty get all instances
	//mapping_table = ".1.3.6.1.2.1.31.1.1.1.1"
	MappingTable string
	// if empty get all subtables
	// sub_tables could be not "real subtables"
	//sub_tables=[".1.3.6.1.2.1.2.2.1.13", "bytes_recv", "bytes_send"]
	SubTables []string
}

type HostTable struct {
	// name = "iftable"
	Name string
	// Includes only these instances
	// include_instances = ["eth0", "eth1"]
	IncludeInstances []string
	// Excludes only these instances
	// exclude_instances = ["eth20", "eth21"]
	ExcludeInstances []string
	// From Table struct
	oid          string
	mappingTable string
	subTables    []string
}

// TODO find better names
type Subtable struct {
	//name = "bytes_send"
	Name string
	//oid = ".1.3.6.1.2.1.31.1.1.1.10"
	Oid string
	//unit = "octets"
	Unit string
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

var sampleConfig = `
  ## Use 'oids.txt' file to translate oids to names
  ## To generate 'oids.txt' you need to run:
  ##   snmptranslate -m all -Tz -On | sed -e 's/"//g' > /tmp/oids.txt
  ## Or if you have an other MIB folder with custom MIBs
  ##   snmptranslate -M /mycustommibfolder -Tz -On -m all | sed -e 's/"//g' > oids.txt
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
    instance = "0"

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

  [[inputs.snmp.host]]
    address = "192.168.2.13:161"
    #address = "127.0.0.1:161"
    community = "public"
    version = 2
    timeout = 2.0
    retries = 2
    #collect = ["mybulk", "sysservices", "sysdescr", "systype"]
    collect = ["sysuptime" ]
    [[inputs.snmp.host.table]]
      name = "iftable3"
      include_instances = ["enp5s0", "eth1"]

  # SNMP TABLEs
  # table without mapping neither subtables
  [[inputs.snmp.table]]
    name = "iftable1"
    oid = ".1.3.6.1.2.1.31.1.1.1"

  # table without mapping but with subtables
  [[inputs.snmp.table]]
    name = "iftable2"
    oid = ".1.3.6.1.2.1.31.1.1.1"
    sub_tables = [".1.3.6.1.2.1.2.2.1.13"]

  # table with mapping but without subtables
  [[inputs.snmp.table]]
    name = "iftable3"
    oid = ".1.3.6.1.2.1.31.1.1.1"
    # if empty. get all instances
    mapping_table = ".1.3.6.1.2.1.31.1.1.1.1"
    # if empty, get all subtables

  # table with both mapping and subtables
  [[inputs.snmp.table]]
    name = "iftable4"
    oid = ".1.3.6.1.2.1.31.1.1.1"
    # if empty get all instances
    mapping_table = ".1.3.6.1.2.1.31.1.1.1.1"
    # if empty get all subtables
    # sub_tables could be not "real subtables"
    sub_tables=[".1.3.6.1.2.1.2.2.1.13", "bytes_recv", "bytes_send"]
`

// SampleConfig returns sample configuration message
func (s *Snmp) SampleConfig() string {
	return sampleConfig
}

// Description returns description of Zookeeper plugin
func (s *Snmp) Description() string {
	return `DEPRECATED! PLEASE USE inputs.snmp INSTEAD.`
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
	// TODO put this in cache on first run
	// Create subtables mapping
	if len(s.subTableMap) == 0 {
		s.subTableMap = make(map[string]Subtable)
		for _, sb := range s.Subtable {
			s.subTableMap[sb.Name] = sb
		}
	}
	// TODO put this in cache on first run
	// Create oid tree
	if s.SnmptranslateFile != "" && len(s.initNode.subnodes) == 0 {
		s.nameToOid = make(map[string]string)
		s.initNode = Node{
			id:       "1",
			name:     "",
			subnodes: make(map[string]Node),
		}

		data, err := ioutil.ReadFile(s.SnmptranslateFile)
		if err != nil {
			s.Log.Errorf("Reading SNMPtranslate file error: %s", err.Error())
			return err
		} else {
			for _, line := range strings.Split(string(data), "\n") {
				oids := strings.Fields(string(line))
				if len(oids) == 2 && oids[1] != "" {
					oid_name := oids[0]
					oid := oids[1]
					fillnode(s.initNode, oid_name, strings.Split(string(oid), "."))
					s.nameToOid[oid_name] = oid
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
			if val, ok := s.nameToOid[oidstring]; ok {
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
					if val, ok := s.nameToOid[oid.Oid]; ok {
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
					if val, ok := s.nameToOid[oid.Oid]; ok {
						oid.rawOid = "." + val
					} else {
						oid.rawOid = oid.Oid
					}
					host.bulkOids = append(host.bulkOids, oid)
				}
			}
		}
		// Table
		for _, hostTable := range host.Table {
			for _, snmpTable := range s.Table {
				if hostTable.Name == snmpTable.Name {
					table := hostTable
					table.oid = snmpTable.Oid
					table.mappingTable = snmpTable.MappingTable
					table.subTables = snmpTable.SubTables
					host.tables = append(host.tables, table)
				}
			}
		}
		// Launch Mapping
		// TODO put this in cache on first run
		// TODO save mapping and computed oids
		// to do it only the first time
		// only if len(s.OidInstanceMapping) == 0
		if len(host.OidInstanceMapping) >= 0 {
			if err := host.SNMPMap(acc, s.nameToOid, s.subTableMap); err != nil {
				s.Log.Errorf("Mapping error for host %q: %s", host.Address, err.Error())
				continue
			}
		}
		// Launch Get requests
		if err := host.SNMPGet(acc, s.initNode); err != nil {
			s.Log.Errorf("Error for host %q: %s", host.Address, err.Error())
		}
		if err := host.SNMPBulk(acc, s.initNode); err != nil {
			s.Log.Errorf("Error for host %q: %s", host.Address, err.Error())
		}
	}
	return nil
}

func (h *Host) SNMPMap(
	acc telegraf.Accumulator,
	nameToOid map[string]string,
	subTableMap map[string]Subtable,
) error {
	if h.OidInstanceMapping == nil {
		h.OidInstanceMapping = make(map[string]map[string]string)
	}
	// Get snmp client
	snmpClient, err := h.GetSNMPClient()
	if err != nil {
		return err
	}
	// Deconnection
	defer snmpClient.Conn.Close()
	// Prepare OIDs
	for _, table := range h.tables {
		// We don't have mapping
		if table.mappingTable == "" {
			if len(table.subTables) == 0 {
				// If We don't have mapping table
				// neither subtables list
				// This is just a bulk request
				oid := Data{}
				oid.Oid = table.oid
				if val, ok := nameToOid[oid.Oid]; ok {
					oid.rawOid = "." + val
				} else {
					oid.rawOid = oid.Oid
				}
				h.bulkOids = append(h.bulkOids, oid)
			} else {
				// If We don't have mapping table
				// but we have subtables
				// This is a bunch of bulk requests
				// For each subtable ...
				for _, sb := range table.subTables {
					// ... we create a new Data (oid) object
					oid := Data{}
					// Looking for more information about this subtable
					ssb, exists := subTableMap[sb]
					if exists {
						// We found a subtable section in config files
						oid.Oid = ssb.Oid
						oid.rawOid = ssb.Oid
						oid.Unit = ssb.Unit
					} else {
						// We did NOT find a subtable section in config files
						oid.Oid = sb
						oid.rawOid = sb
					}
					// TODO check oid validity

					// Add the new oid to getOids list
					h.bulkOids = append(h.bulkOids, oid)
				}
			}
		} else {
			// We have a mapping table
			// We need to query this table
			// To get mapping between instance id
			// and instance name
			oid_asked := table.mappingTable
			oid_next := oid_asked
			need_more_requests := true
			// Set max repetition
			maxRepetition := uint8(32)
			// Launch requests
			for need_more_requests {
				// Launch request
				result, err3 := snmpClient.GetBulk([]string{oid_next}, 0, maxRepetition)
				if err3 != nil {
					return err3
				}

				lastOid := ""
				for _, variable := range result.Variables {
					lastOid = variable.Name
					if strings.HasPrefix(variable.Name, oid_asked) {
						switch variable.Type {
						// handle instance names
						case gosnmp.OctetString:
							// Check if instance is in includes instances
							getInstances := true
							if len(table.IncludeInstances) > 0 {
								getInstances = false
								for _, instance := range table.IncludeInstances {
									if instance == string(variable.Value.([]byte)) {
										getInstances = true
									}
								}
							}
							// Check if instance is in excludes instances
							if len(table.ExcludeInstances) > 0 {
								getInstances = true
								for _, instance := range table.ExcludeInstances {
									if instance == string(variable.Value.([]byte)) {
										getInstances = false
									}
								}
							}
							// We don't want this instance
							if !getInstances {
								continue
							}

							// remove oid table from the complete oid
							// in order to get the current instance id
							key := strings.Replace(variable.Name, oid_asked, "", 1)

							if len(table.subTables) == 0 {
								// We have a mapping table
								// but no subtables
								// This is just a bulk request

								// Building mapping table
								mapping := map[string]string{strings.Trim(key, "."): string(variable.Value.([]byte))}
								_, exists := h.OidInstanceMapping[table.oid]
								if exists {
									h.OidInstanceMapping[table.oid][strings.Trim(key, ".")] = string(variable.Value.([]byte))
								} else {
									h.OidInstanceMapping[table.oid] = mapping
								}

								// Add table oid in bulk oid list
								oid := Data{}
								oid.Oid = table.oid
								if val, ok := nameToOid[oid.Oid]; ok {
									oid.rawOid = "." + val
								} else {
									oid.rawOid = oid.Oid
								}
								h.bulkOids = append(h.bulkOids, oid)
							} else {
								// We have a mapping table
								// and some subtables
								// This is a bunch of get requests
								// This is the best case :)

								// For each subtable ...
								for _, sb := range table.subTables {
									// ... we create a new Data (oid) object
									oid := Data{}
									// Looking for more information about this subtable
									ssb, exists := subTableMap[sb]
									if exists {
										// We found a subtable section in config files
										oid.Oid = ssb.Oid + key
										oid.rawOid = ssb.Oid + key
										oid.Unit = ssb.Unit
										oid.Instance = string(variable.Value.([]byte))
									} else {
										// We did NOT find a subtable section in config files
										oid.Oid = sb + key
										oid.rawOid = sb + key
										oid.Instance = string(variable.Value.([]byte))
									}
									// TODO check oid validity

									// Add the new oid to getOids list
									h.getOids = append(h.getOids, oid)
								}
							}
						default:
						}
					} else {
						break
					}
				}
				// Determine if we need more requests
				if strings.HasPrefix(lastOid, oid_asked) {
					need_more_requests = true
					oid_next = lastOid
				} else {
					need_more_requests = false
				}
			}
		}
	}
	// Mapping finished

	// Create newoids based on mapping

	return nil
}

func (h *Host) SNMPGet(acc telegraf.Accumulator, initNode Node) error {
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
		_, err = h.HandleResponse(oidsList, result, acc, initNode)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Host) SNMPBulk(acc telegraf.Accumulator, initNode Node) error {
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
			last_oid, err := h.HandleResponse(oidsList, result, acc, initNode)
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
	if err != nil {
		return nil, err
	}
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

func (h *Host) HandleResponse(
	oids map[string]Data,
	result *gosnmp.SnmpPacket,
	acc telegraf.Accumulator,
	initNode Node,
) (string, error) {
	var lastOid string
	for _, variable := range result.Variables {
		lastOid = variable.Name
	nextresult:
		// Get only oid wanted
		for oid_key, oid := range oids {
			// Skip oids already processed
			for _, processedOid := range h.processedOids {
				if variable.Name == processedOid {
					break nextresult
				}
			}
			// If variable.Name is the same as oid_key
			// OR
			// the result is SNMP table which "." comes right after oid_key.
			// ex: oid_key: .1.3.6.1.2.1.2.2.1.16, variable.Name: .1.3.6.1.2.1.2.2.1.16.1
			if variable.Name == oid_key || strings.HasPrefix(variable.Name, oid_key+".") {
				switch variable.Type {
				// handle Metrics
				case gosnmp.Boolean, gosnmp.Integer, gosnmp.Counter32, gosnmp.Gauge32,
					gosnmp.TimeTicks, gosnmp.Counter64, gosnmp.Uinteger32, gosnmp.OctetString:
					// Prepare tags
					tags := make(map[string]string)
					if oid.Unit != "" {
						tags["unit"] = oid.Unit
					}
					// Get name and instance
					var oid_name string
					var instance string
					// Get oidname and instance from translate file
					oid_name, instance = findnodename(initNode,
						strings.Split(string(variable.Name[1:]), "."))
					// Set instance tag
					// From mapping table
					mapping, inMappingNoSubTable := h.OidInstanceMapping[oid_key]
					if inMappingNoSubTable {
						// filter if the instance in not in
						// OidInstanceMapping mapping map
						if instance_name, exists := mapping[instance]; exists {
							tags["instance"] = instance_name
						} else {
							continue
						}
					} else if oid.Instance != "" {
						// From config files
						tags["instance"] = oid.Instance
					} else if instance != "" {
						// Using last id of the current oid, ie:
						// with .1.3.6.1.2.1.31.1.1.1.10.3
						// instance is 3
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
					tags["snmp_host"], _, _ = net.SplitHostPort(h.Address)
					fields := make(map[string]interface{})
					fields[string(field_name)] = variable.Value

					h.processedOids = append(h.processedOids, variable.Name)
					acc.AddFields(field_name, fields, tags)
				case gosnmp.NoSuchObject, gosnmp.NoSuchInstance:
					// Oid not found
					log.Printf("E! [inputs.snmp_legacy] oid %q not found", oid_key)
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
	inputs.Add("snmp_legacy", func() telegraf.Input {
		return &Snmp{}
	})
}
