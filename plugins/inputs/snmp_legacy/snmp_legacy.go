package snmp_legacy

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
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
	internalGetOids []Data
	bulkOids        []Data
	tables          []HostTable
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
	MaxRepetition uint32 `toml:"max_repetition"`
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

func fillnode(parentNode Node, oidName string, ids []string) {
	// ids = ["1", "3", "6", ...]
	id, ids := ids[0], ids[1:]
	node, ok := parentNode.subnodes[id]
	if !ok {
		node = Node{
			id:       id,
			name:     "",
			subnodes: make(map[string]Node),
		}
		if len(ids) == 0 {
			node.name = oidName
		}
		parentNode.subnodes[id] = node
	}
	if len(ids) > 0 {
		fillnode(node, oidName, ids)
	}
}

func findNodeName(node Node, ids []string) (oidName string, instance string) {
	// ids = ["1", "3", "6", ...]
	if len(ids) == 1 {
		return node.name, ids[0]
	}
	id, ids := ids[0], ids[1:]
	// Get node
	subnode, ok := node.subnodes[id]
	if ok {
		return findNodeName(subnode, ids)
	}
	// We got a node
	// Get node name
	if node.name != "" && len(ids) == 0 && id == "0" {
		// node with instance 0
		return node.name, "0"
	} else if node.name != "" && len(ids) == 0 && id != "0" {
		// node with an instance
		return node.name, id
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

		data, err := os.ReadFile(s.SnmptranslateFile)
		if err != nil {
			s.Log.Errorf("Reading SNMPtranslate file error: %s", err.Error())
			return err
		}

		for _, line := range strings.Split(string(data), "\n") {
			oids := strings.Fields(line)
			if len(oids) == 2 && oids[1] != "" {
				oidName := oids[0]
				oid := oids[1]
				fillnode(s.initNode, oidName, strings.Split(oid, "."))
				s.nameToOid[oidName] = oid
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
				if oidstring[:1] != "." {
					oid.rawOid = "." + oidstring
				} else {
					oid.rawOid = oidstring
				}
			}
			host.internalGetOids = append(host.internalGetOids, oid)
		}

		for _, oidName := range host.Collect {
			// Get GET oids
			for _, oid := range s.Get {
				if oid.Name == oidName {
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
					host.internalGetOids = append(host.internalGetOids, oid)
				}
			}
			// Get GETBULK oids
			for _, oid := range s.Bulk {
				if oid.Name == oidName {
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
			if err := host.SNMPMap(s.nameToOid, s.subTableMap); err != nil {
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

					// Add the new oid to bulkOids list
					h.bulkOids = append(h.bulkOids, oid)
				}
			}
		} else {
			// We have a mapping table
			// We need to query this table
			// To get mapping between instance id
			// and instance name
			oidAsked := table.mappingTable
			oidNext := oidAsked
			needMoreRequests := true
			// Set max repetition
			maxRepetition := uint32(32)
			// Launch requests
			for needMoreRequests {
				// Launch request
				result, err3 := snmpClient.GetBulk([]string{oidNext}, 0, maxRepetition)
				if err3 != nil {
					return err3
				}

				lastOid := ""
				for _, variable := range result.Variables {
					lastOid = variable.Name
					if strings.HasPrefix(variable.Name, oidAsked) {
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
							key := strings.Replace(variable.Name, oidAsked, "", 1)

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

									// Add the new oid to internalGetOids list
									h.internalGetOids = append(h.internalGetOids, oid)
								}
							}
						default:
						}
					} else {
						break
					}
				}
				// Determine if we need more requests
				if strings.HasPrefix(lastOid, oidAsked) {
					needMoreRequests = true
					oidNext = lastOid
				} else {
					needMoreRequests = false
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
	for _, oid := range h.internalGetOids {
		oidsList[oid.rawOid] = oid
	}
	oidsNameList := make([]string, 0, len(oidsList))
	for _, oid := range oidsList {
		oidsNameList = append(oidsNameList, oid.rawOid)
	}

	// gosnmp.MAX_OIDS == 60
	// TODO use gosnmp.MAX_OIDS instead of hard coded value
	maxOids := 60
	// limit 60 (MAX_OIDS) oids by requests
	for i := 0; i < len(oidsList); i = i + maxOids {
		// Launch request
		maxIndex := i + maxOids
		if i+maxOids > len(oidsList) {
			maxIndex = len(oidsList)
		}
		result, err3 := snmpClient.Get(oidsNameList[i:maxIndex]) // Get() accepts up to g.MAX_OIDS
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
		oidAsked := oid
		needMoreRequests := true
		// Set max repetition
		maxRepetition := oidsList[oid].MaxRepetition
		if maxRepetition <= 0 {
			maxRepetition = 32
		}
		// Launch requests
		for needMoreRequests {
			// Launch request
			result, err3 := snmpClient.GetBulk([]string{oid}, 0, maxRepetition)
			if err3 != nil {
				return err3
			}
			// Handle response
			lastOid, err := h.HandleResponse(oidsList, result, acc, initNode)
			if err != nil {
				return err
			}
			// Determine if we need more requests
			if strings.HasPrefix(lastOid, oidAsked) {
				needMoreRequests = true
				oid = lastOid
			} else {
				needMoreRequests = false
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
	host, portStr, err := net.SplitHostPort(h.Address)
	if err != nil {
		portStr = "161"
	}
	// convert port_str to port in uint16
	port64, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, err
	}
	port := uint16(port64)
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
		for oidKey, oid := range oids {
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
			if variable.Name == oidKey || strings.HasPrefix(variable.Name, oidKey+".") {
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
					var oidName string
					var instance string
					// Get oidname and instance from translate file
					oidName, instance = findNodeName(initNode,
						strings.Split(variable.Name[1:], "."))
					// Set instance tag
					// From mapping table
					mapping, inMappingNoSubTable := h.OidInstanceMapping[oidKey]
					if inMappingNoSubTable {
						// filter if the instance in not in
						// OidInstanceMapping mapping map
						if instanceName, exists := mapping[instance]; exists {
							tags["instance"] = instanceName
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
					var fieldName string
					if oidName != "" {
						// Set fieldname as oid name from translate file
						fieldName = oidName
					} else {
						// Set fieldname as oid name from inputs.snmp.get section
						// Because the result oid is equal to inputs.snmp.get section
						fieldName = oid.Name
					}
					tags["snmp_host"], _, _ = net.SplitHostPort(h.Address)
					fields := make(map[string]interface{})
					fields[fieldName] = variable.Value

					h.processedOids = append(h.processedOids, variable.Name)
					acc.AddFields(fieldName, fields, tags)
				case gosnmp.NoSuchObject, gosnmp.NoSuchInstance:
					// Oid not found
					log.Printf("E! [inputs.snmp_legacy] oid %q not found", oidKey)
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
