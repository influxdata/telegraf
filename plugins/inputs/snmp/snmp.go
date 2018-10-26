package snmp

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/soniah/gosnmp"
)

const description = `Retrieves SNMP values from remote agents`
const sampleConfig = `
  agents = [ "127.0.0.1:161" ]
  ## Timeout for each SNMP query.
  timeout = "5s"
  ## Number of retries to attempt within timeout.
  retries = 3
  ## SNMP version, values can be 1, 2, or 3
  version = 2

  ## SNMP community string.
  community = "public"

  ## The GETBULK max-repetitions parameter
  max_repetitions = 10

  ## SNMPv3 auth parameters
  #sec_name = "myuser"
  #auth_protocol = "md5"      # Values: "MD5", "SHA", ""
  #auth_password = "pass"
  #sec_level = "authNoPriv"   # Values: "noAuthNoPriv", "authNoPriv", "authPriv"
  #context_name = ""
  #priv_protocol = ""         # Values: "DES", "AES", ""
  #priv_password = ""

  ## measurement name
  name = "system"
  [[inputs.snmp.field]]
    name = "hostname"
    oid = ".1.0.0.1.1"
  [[inputs.snmp.field]]
    name = "uptime"
    oid = ".1.0.0.1.2"
  [[inputs.snmp.field]]
    name = "load"
    oid = ".1.0.0.1.3"
  [[inputs.snmp.field]]
    oid = "HOST-RESOURCES-MIB::hrMemorySize"

  [[inputs.snmp.table]]
    ## measurement name
    name = "remote_servers"
    inherit_tags = [ "hostname" ]
    [[inputs.snmp.table.field]]
      name = "server"
      oid = ".1.0.0.0.1.0"
      is_tag = true
    [[inputs.snmp.table.field]]
      name = "connections"
      oid = ".1.0.0.0.1.1"
    [[inputs.snmp.table.field]]
      name = "latency"
      oid = ".1.0.0.0.1.2"

  [[inputs.snmp.table]]
    ## auto populate table's fields using the MIB
    oid = "HOST-RESOURCES-MIB::hrNetworkTable"
`

// execCommand is so tests can mock out exec.Command usage.
var execCommand = exec.Command

// execCmd executes the specified command, returning the STDOUT content.
// If command exits with error status, the output is captured into the returned error.
func execCmd(arg0 string, args ...string) ([]byte, error) {
	out, err := execCommand(arg0, args...).Output()
	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			return nil, NestedError{
				Err:       err,
				NestedErr: fmt.Errorf("%s", bytes.TrimRight(err.Stderr, "\r\n")),
			}
		}
		return nil, err
	}
	return out, nil
}

// Snmp holds the configuration for the plugin.
type Snmp struct {
	// The SNMP agent to query. Format is ADDR[:PORT] (e.g. 1.2.3.4:161).
	Agents []string
	// Timeout to wait for a response.
	Timeout internal.Duration
	Retries int
	// Values: 1, 2, 3
	Version uint8

	// Parameters for Version 1 & 2
	Community string

	// Parameters for Version 2 & 3
	MaxRepetitions uint8

	// Parameters for Version 3
	ContextName string
	// Values: "noAuthNoPriv", "authNoPriv", "authPriv"
	SecLevel string
	SecName  string
	// Values: "MD5", "SHA", "". Default: ""
	AuthProtocol string
	AuthPassword string
	// Values: "DES", "AES", "". Default: ""
	PrivProtocol string
	PrivPassword string
	EngineID     string
	EngineBoots  uint32
	EngineTime   uint32

	Tables []Table `toml:"table"`

	// Name & Fields are the elements of a Table.
	// Telegraf chokes if we try to embed a Table. So instead we have to embed the
	// fields of a Table, and construct a Table during runtime.
	Name   string
	Fields []Field `toml:"field"`

	connectionCache []snmpConnection
	initialized     bool
}

func (s *Snmp) init() error {
	if s.initialized {
		return nil
	}

	s.connectionCache = make([]snmpConnection, len(s.Agents))

	for i := range s.Tables {
		if err := s.Tables[i].init(); err != nil {
			return Errorf(err, "initializing table %s", s.Tables[i].Name)
		}
	}

	for i := range s.Fields {
		if err := s.Fields[i].init(); err != nil {
			return Errorf(err, "initializing field %s", s.Fields[i].Name)
		}
	}

	s.initialized = true
	return nil
}

// Table holds the configuration for a SNMP table.
type Table struct {
	// Name will be the name of the measurement.
	Name string

	// Which tags to inherit from the top-level config.
	InheritTags []string

	// Adds each row's table index as a tag.
	IndexAsTag bool

	// Fields is the tags and values to look up.
	Fields []Field `toml:"field"`

	// OID for automatic field population.
	// If provided, init() will populate Fields with all the table columns of the
	// given OID.
	Oid string

	initialized bool
}

// init() builds & initializes the nested fields.
func (t *Table) init() error {
	if t.initialized {
		return nil
	}

	if err := t.initBuild(); err != nil {
		return err
	}

	// initialize all the nested fields
	for i := range t.Fields {
		if err := t.Fields[i].init(); err != nil {
			return Errorf(err, "initializing field %s", t.Fields[i].Name)
		}
	}

	t.initialized = true
	return nil
}

// initBuild initializes the table if it has an OID configured. If so, the
// net-snmp tools will be used to look up the OID and auto-populate the table's
// fields.
func (t *Table) initBuild() error {
	if t.Oid == "" {
		return nil
	}

	_, _, oidText, fields, err := snmpTable(t.Oid)
	if err != nil {
		return err
	}

	if t.Name == "" {
		t.Name = oidText
	}

	knownOIDs := map[string]bool{}
	for _, f := range t.Fields {
		knownOIDs[f.Oid] = true
	}
	for _, f := range fields {
		if !knownOIDs[f.Oid] {
			t.Fields = append(t.Fields, f)
		}
	}

	return nil
}

// Field holds the configuration for a Field to look up.
type Field struct {
	// Name will be the name of the field.
	Name string
	// OID is prefix for this field. The plugin will perform a walk through all
	// OIDs with this as their parent. For each value found, the plugin will strip
	// off the OID prefix, and use the remainder as the index. For multiple fields
	// to show up in the same row, they must share the same index.
	Oid string
	// OidIndexSuffix is the trailing sub-identifier on a table record OID that will be stripped off to get the record's index.
	OidIndexSuffix string
	// OidIndexLength specifies the length of the index in OID path segments. It can be used to remove sub-identifiers that vary in content or length.
	OidIndexLength int
	// IsTag controls whether this OID is output as a tag or a value.
	IsTag bool
	// Conversion controls any type conversion that is done on the value.
	//  "float"/"float(0)" will convert the value into a float.
	//  "float(X)" will convert the value into a float, and then move the decimal before Xth right-most digit.
	//  "int" will conver the value into an integer.
	//  "hwaddr" will convert a 6-byte string to a MAC address.
	//  "ipaddr" will convert the value to an IPv4 or IPv6 address.
	Conversion string

	initialized bool
}

// init() converts OID names to numbers, and sets the .Name attribute if unset.
func (f *Field) init() error {
	if f.initialized {
		return nil
	}

	_, oidNum, oidText, conversion, err := snmpTranslate(f.Oid)
	if err != nil {
		return Errorf(err, "translating")
	}
	f.Oid = oidNum
	if f.Name == "" {
		f.Name = oidText
	}
	if f.Conversion == "" {
		f.Conversion = conversion
	}

	//TODO use textual convention conversion from the MIB

	f.initialized = true
	return nil
}

// RTable is the resulting table built from a Table.
type RTable struct {
	// Name is the name of the field, copied from Table.Name.
	Name string
	// Time is the time the table was built.
	Time time.Time
	// Rows are the rows that were found, one row for each table OID index found.
	Rows []RTableRow
}

// RTableRow is the resulting row containing all the OID values which shared
// the same index.
type RTableRow struct {
	// Tags are all the Field values which had IsTag=true.
	Tags map[string]string
	// Fields are all the Field values which had IsTag=false.
	Fields map[string]interface{}
}

// NestedError wraps an error returned from deeper in the code.
type NestedError struct {
	// Err is the error from where the NestedError was constructed.
	Err error
	// NestedError is the error that was passed back from the called function.
	NestedErr error
}

// Error returns a concatenated string of all the nested errors.
func (ne NestedError) Error() string {
	return ne.Err.Error() + ": " + ne.NestedErr.Error()
}

// Errorf is a convenience function for constructing a NestedError.
func Errorf(err error, msg string, format ...interface{}) error {
	return NestedError{
		NestedErr: err,
		Err:       fmt.Errorf(msg, format...),
	}
}

func init() {
	inputs.Add("snmp", func() telegraf.Input {
		return &Snmp{
			Name:           "snmp",
			Retries:        3,
			MaxRepetitions: 10,
			Timeout:        internal.Duration{Duration: 5 * time.Second},
			Version:        2,
			Community:      "public",
		}
	})
}

// SampleConfig returns the default configuration of the input.
func (s *Snmp) SampleConfig() string {
	return sampleConfig
}

// Description returns a one-sentence description on the input.
func (s *Snmp) Description() string {
	return description
}

// Gather retrieves all the configured fields and tables.
// Any error encountered does not halt the process. The errors are accumulated
// and returned at the end.
func (s *Snmp) Gather(acc telegraf.Accumulator) error {
	if err := s.init(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	for i, agent := range s.Agents {
		wg.Add(1)
		go func(i int, agent string) {
			defer wg.Done()
			gs, err := s.getConnection(i)
			if err != nil {
				acc.AddError(Errorf(err, "agent %s", agent))
				return
			}

			// First is the top-level fields. We treat the fields as table prefixes with an empty index.
			t := Table{
				Name:   s.Name,
				Fields: s.Fields,
			}
			topTags := map[string]string{}
			if err := s.gatherTable(acc, gs, t, topTags, false); err != nil {
				acc.AddError(Errorf(err, "agent %s", agent))
			}

			// Now is the real tables.
			for _, t := range s.Tables {
				if err := s.gatherTable(acc, gs, t, topTags, true); err != nil {
					acc.AddError(Errorf(err, "agent %s: gathering table %s", agent, t.Name))
				}
			}
		}(i, agent)
	}
	wg.Wait()

	return nil
}

func (s *Snmp) gatherTable(acc telegraf.Accumulator, gs snmpConnection, t Table, topTags map[string]string, walk bool) error {
	rt, err := t.Build(gs, walk)
	if err != nil {
		return err
	}

	for _, tr := range rt.Rows {
		if !walk {
			// top-level table. Add tags to topTags.
			for k, v := range tr.Tags {
				topTags[k] = v
			}
		} else {
			// real table. Inherit any specified tags.
			for _, k := range t.InheritTags {
				if v, ok := topTags[k]; ok {
					tr.Tags[k] = v
				}
			}
		}
		if _, ok := tr.Tags["agent_host"]; !ok {
			tr.Tags["agent_host"] = gs.Host()
		}
		acc.AddFields(rt.Name, tr.Fields, tr.Tags, rt.Time)
	}

	return nil
}

// Build retrieves all the fields specified in the table and constructs the RTable.
func (t Table) Build(gs snmpConnection, walk bool) (*RTable, error) {
	rows := map[string]RTableRow{}

	tagCount := 0
	for _, f := range t.Fields {
		if f.IsTag {
			tagCount++
		}

		if len(f.Oid) == 0 {
			return nil, fmt.Errorf("cannot have empty OID on field %s", f.Name)
		}
		var oid string
		if f.Oid[0] == '.' {
			oid = f.Oid
		} else {
			// make sure OID has "." because the BulkWalkAll results do, and the prefix needs to match
			oid = "." + f.Oid
		}

		// ifv contains a mapping of table OID index to field value
		ifv := map[string]interface{}{}

		if !walk {
			// This is used when fetching non-table fields. Fields configured a the top
			// scope of the plugin.
			// We fetch the fields directly, and add them to ifv as if the index were an
			// empty string. This results in all the non-table fields sharing the same
			// index, and being added on the same row.
			if pkt, err := gs.Get([]string{oid}); err != nil {
				return nil, Errorf(err, "performing get on field %s", f.Name)
			} else if pkt != nil && len(pkt.Variables) > 0 && pkt.Variables[0].Type != gosnmp.NoSuchObject && pkt.Variables[0].Type != gosnmp.NoSuchInstance {
				ent := pkt.Variables[0]
				fv, err := fieldConvert(f.Conversion, ent.Value)
				if err != nil {
					return nil, Errorf(err, "converting %q (OID %s) for field %s", ent.Value, ent.Name, f.Name)
				}
				ifv[""] = fv
			}
		} else {
			err := gs.Walk(oid, func(ent gosnmp.SnmpPDU) error {
				if len(ent.Name) <= len(oid) || ent.Name[:len(oid)+1] != oid+"." {
					return NestedError{} // break the walk
				}

				idx := ent.Name[len(oid):]
				if f.OidIndexSuffix != "" {
					if !strings.HasSuffix(idx, f.OidIndexSuffix) {
						// this entry doesn't match our OidIndexSuffix. skip it
						return nil
					}
					idx = idx[:len(idx)-len(f.OidIndexSuffix)]
				}
				if f.OidIndexLength != 0 {
					i := f.OidIndexLength + 1 // leading separator
					idx = strings.Map(func(r rune) rune {
						if r == '.' {
							i -= 1
						}
						if i < 1 {
							return -1
						}
						return r
					}, idx)
				}

				fv, err := fieldConvert(f.Conversion, ent.Value)
				if err != nil {
					return Errorf(err, "converting %q (OID %s) for field %s", ent.Value, ent.Name, f.Name)
				}
				ifv[idx] = fv
				return nil
			})
			if err != nil {
				if _, ok := err.(NestedError); !ok {
					return nil, Errorf(err, "performing bulk walk for field %s", f.Name)
				}
			}
		}

		for idx, v := range ifv {
			rtr, ok := rows[idx]
			if !ok {
				rtr = RTableRow{}
				rtr.Tags = map[string]string{}
				rtr.Fields = map[string]interface{}{}
				rows[idx] = rtr
			}
			if t.IndexAsTag && idx != "" {
				if idx[0] == '.' {
					idx = idx[1:]
				}
				rtr.Tags["index"] = idx
			}
			// don't add an empty string
			if vs, ok := v.(string); !ok || vs != "" {
				if f.IsTag {
					if ok {
						rtr.Tags[f.Name] = vs
					} else {
						rtr.Tags[f.Name] = fmt.Sprintf("%v", v)
					}
				} else {
					rtr.Fields[f.Name] = v
				}
			}
		}
	}

	rt := RTable{
		Name: t.Name,
		Time: time.Now(), //TODO record time at start
		Rows: make([]RTableRow, 0, len(rows)),
	}
	for _, r := range rows {
		rt.Rows = append(rt.Rows, r)
	}
	return &rt, nil
}

// snmpConnection is an interface which wraps a *gosnmp.GoSNMP object.
// We interact through an interface so we can mock it out in tests.
type snmpConnection interface {
	Host() string
	//BulkWalkAll(string) ([]gosnmp.SnmpPDU, error)
	Walk(string, gosnmp.WalkFunc) error
	Get(oids []string) (*gosnmp.SnmpPacket, error)
}

// gosnmpWrapper wraps a *gosnmp.GoSNMP object so we can use it as a snmpConnection.
type gosnmpWrapper struct {
	*gosnmp.GoSNMP
}

// Host returns the value of GoSNMP.Target.
func (gsw gosnmpWrapper) Host() string {
	return gsw.Target
}

// Walk wraps GoSNMP.Walk() or GoSNMP.BulkWalk(), depending on whether the
// connection is using SNMPv1 or newer.
// Also, if any error is encountered, it will just once reconnect and try again.
func (gsw gosnmpWrapper) Walk(oid string, fn gosnmp.WalkFunc) error {
	var err error
	// On error, retry once.
	// Unfortunately we can't distinguish between an error returned by gosnmp, and one returned by the walk function.
	for i := 0; i < 2; i++ {
		if gsw.Version == gosnmp.Version1 {
			err = gsw.GoSNMP.Walk(oid, fn)
		} else {
			err = gsw.GoSNMP.BulkWalk(oid, fn)
		}
		if err == nil {
			return nil
		}
		if err := gsw.GoSNMP.Connect(); err != nil {
			return Errorf(err, "reconnecting")
		}
	}
	return err
}

// Get wraps GoSNMP.GET().
// If any error is encountered, it will just once reconnect and try again.
func (gsw gosnmpWrapper) Get(oids []string) (*gosnmp.SnmpPacket, error) {
	var err error
	var pkt *gosnmp.SnmpPacket
	for i := 0; i < 2; i++ {
		pkt, err = gsw.GoSNMP.Get(oids)
		if err == nil {
			return pkt, nil
		}
		if err := gsw.GoSNMP.Connect(); err != nil {
			return nil, Errorf(err, "reconnecting")
		}
	}
	return nil, err
}

// getConnection creates a snmpConnection (*gosnmp.GoSNMP) object and caches the
// result using `agentIndex` as the cache key.  This is done to allow multiple
// connections to a single address.  It is an error to use a connection in
// more than one goroutine.
func (s *Snmp) getConnection(idx int) (snmpConnection, error) {
	if gs := s.connectionCache[idx]; gs != nil {
		return gs, nil
	}

	agent := s.Agents[idx]

	gs := gosnmpWrapper{&gosnmp.GoSNMP{}}
	s.connectionCache[idx] = gs

	host, portStr, err := net.SplitHostPort(agent)
	if err != nil {
		if err, ok := err.(*net.AddrError); !ok || err.Err != "missing port in address" {
			return nil, Errorf(err, "parsing host")
		}
		host = agent
		portStr = "161"
	}
	gs.Target = host

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, Errorf(err, "parsing port")
	}
	gs.Port = uint16(port)

	gs.Timeout = s.Timeout.Duration

	gs.Retries = s.Retries

	switch s.Version {
	case 3:
		gs.Version = gosnmp.Version3
	case 2, 0:
		gs.Version = gosnmp.Version2c
	case 1:
		gs.Version = gosnmp.Version1
	default:
		return nil, fmt.Errorf("invalid version")
	}

	if s.Version < 3 {
		if s.Community == "" {
			gs.Community = "public"
		} else {
			gs.Community = s.Community
		}
	}

	gs.MaxRepetitions = s.MaxRepetitions

	if s.Version == 3 {
		gs.ContextName = s.ContextName

		sp := &gosnmp.UsmSecurityParameters{}
		gs.SecurityParameters = sp
		gs.SecurityModel = gosnmp.UserSecurityModel

		switch strings.ToLower(s.SecLevel) {
		case "noauthnopriv", "":
			gs.MsgFlags = gosnmp.NoAuthNoPriv
		case "authnopriv":
			gs.MsgFlags = gosnmp.AuthNoPriv
		case "authpriv":
			gs.MsgFlags = gosnmp.AuthPriv
		default:
			return nil, fmt.Errorf("invalid secLevel")
		}

		sp.UserName = s.SecName

		switch strings.ToLower(s.AuthProtocol) {
		case "md5":
			sp.AuthenticationProtocol = gosnmp.MD5
		case "sha":
			sp.AuthenticationProtocol = gosnmp.SHA
		case "":
			sp.AuthenticationProtocol = gosnmp.NoAuth
		default:
			return nil, fmt.Errorf("invalid authProtocol")
		}

		sp.AuthenticationPassphrase = s.AuthPassword

		switch strings.ToLower(s.PrivProtocol) {
		case "des":
			sp.PrivacyProtocol = gosnmp.DES
		case "aes":
			sp.PrivacyProtocol = gosnmp.AES
		case "":
			sp.PrivacyProtocol = gosnmp.NoPriv
		default:
			return nil, fmt.Errorf("invalid privProtocol")
		}

		sp.PrivacyPassphrase = s.PrivPassword

		sp.AuthoritativeEngineID = s.EngineID

		sp.AuthoritativeEngineBoots = s.EngineBoots

		sp.AuthoritativeEngineTime = s.EngineTime
	}

	if err := gs.Connect(); err != nil {
		return nil, Errorf(err, "setting up connection")
	}

	return gs, nil
}

// fieldConvert converts from any type according to the conv specification
//  "float"/"float(0)" will convert the value into a float.
//  "float(X)" will convert the value into a float, and then move the decimal before Xth right-most digit.
//  "int" will convert the value into an integer.
//  "hwaddr" will convert the value into a MAC address.
//  "ipaddr" will convert the value into into an IP address.
//  "" will convert a byte slice into a string.
func fieldConvert(conv string, v interface{}) (interface{}, error) {
	if conv == "" {
		if bs, ok := v.([]byte); ok {
			return string(bs), nil
		}
		return v, nil
	}

	var d int
	if _, err := fmt.Sscanf(conv, "float(%d)", &d); err == nil || conv == "float" {
		switch vt := v.(type) {
		case float32:
			v = float64(vt) / math.Pow10(d)
		case float64:
			v = float64(vt) / math.Pow10(d)
		case int:
			v = float64(vt) / math.Pow10(d)
		case int8:
			v = float64(vt) / math.Pow10(d)
		case int16:
			v = float64(vt) / math.Pow10(d)
		case int32:
			v = float64(vt) / math.Pow10(d)
		case int64:
			v = float64(vt) / math.Pow10(d)
		case uint:
			v = float64(vt) / math.Pow10(d)
		case uint8:
			v = float64(vt) / math.Pow10(d)
		case uint16:
			v = float64(vt) / math.Pow10(d)
		case uint32:
			v = float64(vt) / math.Pow10(d)
		case uint64:
			v = float64(vt) / math.Pow10(d)
		case []byte:
			vf, _ := strconv.ParseFloat(string(vt), 64)
			v = vf / math.Pow10(d)
		case string:
			vf, _ := strconv.ParseFloat(vt, 64)
			v = vf / math.Pow10(d)
		}
		return v, nil
	}

	if conv == "int" {
		switch vt := v.(type) {
		case float32:
			v = int64(vt)
		case float64:
			v = int64(vt)
		case int:
			v = int64(vt)
		case int8:
			v = int64(vt)
		case int16:
			v = int64(vt)
		case int32:
			v = int64(vt)
		case int64:
			v = int64(vt)
		case uint:
			v = int64(vt)
		case uint8:
			v = int64(vt)
		case uint16:
			v = int64(vt)
		case uint32:
			v = int64(vt)
		case uint64:
			v = int64(vt)
		case []byte:
			v, _ = strconv.Atoi(string(vt))
		case string:
			v, _ = strconv.Atoi(vt)
		}
		return v, nil
	}

	if conv == "hwaddr" {
		switch vt := v.(type) {
		case string:
			v = net.HardwareAddr(vt).String()
		case []byte:
			v = net.HardwareAddr(vt).String()
		default:
			return nil, fmt.Errorf("invalid type (%T) for hwaddr conversion", v)
		}
		return v, nil
	}

	if conv == "ipaddr" {
		var ipbs []byte

		switch vt := v.(type) {
		case string:
			ipbs = []byte(vt)
		case []byte:
			ipbs = vt
		default:
			return nil, fmt.Errorf("invalid type (%T) for ipaddr conversion", v)
		}

		switch len(ipbs) {
		case 4, 16:
			v = net.IP(ipbs).String()
		default:
			return nil, fmt.Errorf("invalid length (%d) for ipaddr conversion", len(ipbs))
		}

		return v, nil
	}

	return nil, fmt.Errorf("invalid conversion type '%s'", conv)
}

type snmpTableCache struct {
	mibName string
	oidNum  string
	oidText string
	fields  []Field
	err     error
}

var snmpTableCaches map[string]snmpTableCache
var snmpTableCachesLock sync.Mutex

// snmpTable resolves the given OID as a table, providing information about the
// table and fields within.
func snmpTable(oid string) (mibName string, oidNum string, oidText string, fields []Field, err error) {
	snmpTableCachesLock.Lock()
	if snmpTableCaches == nil {
		snmpTableCaches = map[string]snmpTableCache{}
	}

	var stc snmpTableCache
	var ok bool
	if stc, ok = snmpTableCaches[oid]; !ok {
		stc.mibName, stc.oidNum, stc.oidText, stc.fields, stc.err = snmpTableCall(oid)
		snmpTableCaches[oid] = stc
	}

	snmpTableCachesLock.Unlock()
	return stc.mibName, stc.oidNum, stc.oidText, stc.fields, stc.err
}

func snmpTableCall(oid string) (mibName string, oidNum string, oidText string, fields []Field, err error) {
	mibName, oidNum, oidText, _, err = snmpTranslate(oid)
	if err != nil {
		return "", "", "", nil, Errorf(err, "translating")
	}

	mibPrefix := mibName + "::"
	oidFullName := mibPrefix + oidText

	// first attempt to get the table's tags
	tagOids := map[string]struct{}{}
	// We have to guess that the "entry" oid is `oid+".1"`. snmptable and snmptranslate don't seem to have a way to provide the info.
	if out, err := execCmd("snmptranslate", "-Td", oidFullName+".1"); err == nil {
		scanner := bufio.NewScanner(bytes.NewBuffer(out))
		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "  INDEX") {
				continue
			}

			i := strings.Index(line, "{ ")
			if i == -1 { // parse error
				continue
			}
			line = line[i+2:]
			i = strings.Index(line, " }")
			if i == -1 { // parse error
				continue
			}
			line = line[:i]
			for _, col := range strings.Split(line, ", ") {
				tagOids[mibPrefix+col] = struct{}{}
			}
		}
	}

	// this won't actually try to run a query. The `-Ch` will just cause it to dump headers.
	out, err := execCmd("snmptable", "-Ch", "-Cl", "-c", "public", "127.0.0.1", oidFullName)
	if err != nil {
		return "", "", "", nil, Errorf(err, "getting table columns")
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(out))
	scanner.Scan()
	cols := scanner.Text()
	if len(cols) == 0 {
		return "", "", "", nil, fmt.Errorf("could not find any columns in table")
	}
	for _, col := range strings.Split(cols, " ") {
		if len(col) == 0 {
			continue
		}
		_, isTag := tagOids[mibPrefix+col]
		fields = append(fields, Field{Name: col, Oid: mibPrefix + col, IsTag: isTag})
	}

	return mibName, oidNum, oidText, fields, err
}

type snmpTranslateCache struct {
	mibName    string
	oidNum     string
	oidText    string
	conversion string
	err        error
}

var snmpTranslateCachesLock sync.Mutex
var snmpTranslateCaches map[string]snmpTranslateCache

// snmpTranslate resolves the given OID.
func snmpTranslate(oid string) (mibName string, oidNum string, oidText string, conversion string, err error) {
	snmpTranslateCachesLock.Lock()
	if snmpTranslateCaches == nil {
		snmpTranslateCaches = map[string]snmpTranslateCache{}
	}

	var stc snmpTranslateCache
	var ok bool
	if stc, ok = snmpTranslateCaches[oid]; !ok {
		// This will result in only one call to snmptranslate running at a time.
		// We could speed it up by putting a lock in snmpTranslateCache and then
		// returning it immediately, and multiple callers would then release the
		// snmpTranslateCachesLock and instead wait on the individual
		// snmpTranlsation.Lock to release. But I don't know that the extra complexity
		// is worth it. Especially when it would slam the system pretty hard if lots
		// of lookups are being perfomed.

		stc.mibName, stc.oidNum, stc.oidText, stc.conversion, stc.err = snmpTranslateCall(oid)
		snmpTranslateCaches[oid] = stc
	}

	snmpTranslateCachesLock.Unlock()

	return stc.mibName, stc.oidNum, stc.oidText, stc.conversion, stc.err
}

func snmpTranslateCall(oid string) (mibName string, oidNum string, oidText string, conversion string, err error) {
	var out []byte
	if strings.ContainsAny(oid, ":abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		out, err = execCmd("snmptranslate", "-Td", "-Ob", oid)
	} else {
		out, err = execCmd("snmptranslate", "-Td", "-Ob", "-m", "all", oid)
		if err, ok := err.(*exec.Error); ok && err.Err == exec.ErrNotFound {
			// Silently discard error if snmptranslate not found and we have a numeric OID.
			// Meaning we can get by without the lookup.
			return "", oid, oid, "", nil
		}
	}
	if err != nil {
		return "", "", "", "", err
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(out))
	ok := scanner.Scan()
	if !ok && scanner.Err() != nil {
		return "", "", "", "", Errorf(scanner.Err(), "getting OID text")
	}

	oidText = scanner.Text()

	i := strings.Index(oidText, "::")
	if i == -1 {
		// was not found in MIB.
		if bytes.Contains(out, []byte("[TRUNCATED]")) {
			return "", oid, oid, "", nil
		}
		// not truncated, but not fully found. We still need to parse out numeric OID, so keep going
		oidText = oid
	} else {
		mibName = oidText[:i]
		oidText = oidText[i+2:]
	}

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "  -- TEXTUAL CONVENTION ") {
			tc := strings.TrimPrefix(line, "  -- TEXTUAL CONVENTION ")
			switch tc {
			case "MacAddress", "PhysAddress":
				conversion = "hwaddr"
			case "InetAddressIPv4", "InetAddressIPv6", "InetAddress", "IPSIpAddress":
				conversion = "ipaddr"
			}
		} else if strings.HasPrefix(line, "::= { ") {
			objs := strings.TrimPrefix(line, "::= { ")
			objs = strings.TrimSuffix(objs, " }")

			for _, obj := range strings.Split(objs, " ") {
				if len(obj) == 0 {
					continue
				}
				if i := strings.Index(obj, "("); i != -1 {
					obj = obj[i+1:]
					oidNum += "." + obj[:strings.Index(obj, ")")]
				} else {
					oidNum += "." + obj
				}
			}
			break
		}
	}

	return mibName, oidNum, oidText, conversion, nil
}
