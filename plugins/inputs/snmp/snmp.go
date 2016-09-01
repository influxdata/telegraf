package snmp

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/soniah/gosnmp"
)

const description = `Retrieves SNMP values from remote agents`
const sampleConfig = `
  agents = [ "127.0.0.1:161" ]
  timeout = "5s"
  version = 2

  # SNMPv1 & SNMPv2 parameters
  community = "public"

  # SNMPv2 & SNMPv3 parameters
  max_repetitions = 50

  # SNMPv3 parameters
  #sec_name = "myuser"
  #auth_protocol = "md5"         # Values: "MD5", "SHA", ""
  #auth_password = "password123"
  #sec_level = "authNoPriv"      # Values: "noAuthNoPriv", "authNoPriv", "authPriv"
  #context_name = ""
  #priv_protocol = ""            # Values: "DES", "AES", ""
  #priv_password = ""

  # measurement name
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
    # measurement name
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
    # auto populate table's fields using the MIB
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
				NestedErr: fmt.Errorf("%s", bytes.TrimRight(err.Stderr, "\n")),
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
	MaxRepetitions uint

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

	connectionCache map[string]snmpConnection
	initialized     bool
}

func (s *Snmp) init() error {
	if s.initialized {
		return nil
	}

	for i := range s.Tables {
		if err := s.Tables[i].init(); err != nil {
			return err
		}
	}

	for i := range s.Fields {
		if err := s.Fields[i].init(); err != nil {
			return err
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

	// Fields is the tags and values to look up.
	Fields []Field `toml:"field"`

	// OID for automatic field population.
	// If provided, init() will populate Fields with all the table columns of the
	// given OID.
	Oid string

	initialized bool
}

// init() populates Fields if a table OID is provided.
func (t *Table) init() error {
	if t.initialized {
		return nil
	}
	if t.Oid == "" {
		t.initialized = true
		return nil
	}

	mibPrefix := ""
	if err := snmpTranslate(&mibPrefix, &t.Oid, &t.Name); err != nil {
		return err
	}

	// first attempt to get the table's tags
	tagOids := map[string]struct{}{}
	// We have to guess that the "entry" oid is `t.Oid+".1"`. snmptable and snmptranslate don't seem to have a way to provide the info.
	if out, err := execCmd("snmptranslate", "-m", "all", "-Td", t.Oid+".1"); err == nil {
		lines := bytes.Split(out, []byte{'\n'})
		// get the MIB name if we didn't get it above
		if mibPrefix == "" {
			if i := bytes.Index(lines[0], []byte("::")); i != -1 {
				mibPrefix = string(lines[0][:i+2])
			}
		}

		for _, line := range lines {
			if !bytes.HasPrefix(line, []byte("  INDEX")) {
				continue
			}

			i := bytes.Index(line, []byte("{ "))
			if i == -1 { // parse error
				continue
			}
			line = line[i+2:]
			i = bytes.Index(line, []byte(" }"))
			if i == -1 { // parse error
				continue
			}
			line = line[:i]
			for _, col := range bytes.Split(line, []byte(", ")) {
				tagOids[mibPrefix+string(col)] = struct{}{}
			}
		}
	}

	// this won't actually try to run a query. The `-Ch` will just cause it to dump headers.
	out, err := execCmd("snmptable", "-m", "all", "-Ch", "-Cl", "-c", "public", "127.0.0.1", t.Oid)
	if err != nil {
		return Errorf(err, "getting table columns for %s", t.Oid)
	}
	cols := bytes.SplitN(out, []byte{'\n'}, 2)[0]
	if len(cols) == 0 {
		return fmt.Errorf("unable to get columns for table %s", t.Oid)
	}
	for _, col := range bytes.Split(cols, []byte{' '}) {
		if len(col) == 0 {
			continue
		}
		col := string(col)
		_, isTag := tagOids[mibPrefix+col]
		t.Fields = append(t.Fields, Field{Name: col, Oid: mibPrefix + col, IsTag: isTag})
	}

	// initialize all the nested fields
	for i := range t.Fields {
		if err := t.Fields[i].init(); err != nil {
			return err
		}
	}

	t.initialized = true
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
	// IsTag controls whether this OID is output as a tag or a value.
	IsTag bool
	// Conversion controls any type conversion that is done on the value.
	//  "float"/"float(0)" will convert the value into a float.
	//  "float(X)" will convert the value into a float, and then move the decimal before Xth right-most digit.
	//  "int" will conver the value into an integer.
	Conversion string

	initialized bool
}

// init() converts OID names to numbers, and sets the .Name attribute if unset.
func (f *Field) init() error {
	if f.initialized {
		return nil
	}

	if err := snmpTranslate(nil, &f.Oid, &f.Name); err != nil {
		return err
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
			Retries:        5,
			MaxRepetitions: 50,
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

	for _, agent := range s.Agents {
		gs, err := s.getConnection(agent)
		if err != nil {
			acc.AddError(Errorf(err, "agent %s", agent))
			continue
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
				acc.AddError(Errorf(err, "agent %s", agent))
			}
		}
	}

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
			return nil, fmt.Errorf("cannot have empty OID")
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
				return nil, Errorf(err, "performing get")
			} else if pkt != nil && len(pkt.Variables) > 0 && pkt.Variables[0].Type != gosnmp.NoSuchObject {
				ent := pkt.Variables[0]
				ifv[ent.Name[len(oid):]] = fieldConvert(f.Conversion, ent.Value)
			}
		} else {
			err := gs.Walk(oid, func(ent gosnmp.SnmpPDU) error {
				if len(ent.Name) <= len(oid) || ent.Name[:len(oid)+1] != oid+"." {
					return NestedError{} // break the walk
				}
				ifv[ent.Name[len(oid):]] = fieldConvert(f.Conversion, ent.Value)
				return nil
			})
			if err != nil {
				if _, ok := err.(NestedError); !ok {
					return nil, Errorf(err, "performing bulk walk")
				}
			}
		}

		for i, v := range ifv {
			rtr, ok := rows[i]
			if !ok {
				rtr = RTableRow{}
				rtr.Tags = map[string]string{}
				rtr.Fields = map[string]interface{}{}
				rows[i] = rtr
			}
			if f.IsTag {
				if vs, ok := v.(string); ok {
					rtr.Tags[f.Name] = vs
				} else {
					rtr.Tags[f.Name] = fmt.Sprintf("%v", v)
				}
			} else {
				rtr.Fields[f.Name] = v
			}
		}
	}

	rt := RTable{
		Name: t.Name,
		Time: time.Now(), //TODO record time at start
		Rows: make([]RTableRow, 0, len(rows)),
	}
	for _, r := range rows {
		if len(r.Tags) < tagCount {
			// don't add rows which are missing tags, as without tags you can't filter
			continue
		}
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
// result using `agent` as the cache key.
func (s *Snmp) getConnection(agent string) (snmpConnection, error) {
	if s.connectionCache == nil {
		s.connectionCache = map[string]snmpConnection{}
	}
	if gs, ok := s.connectionCache[agent]; ok {
		return gs, nil
	}

	gs := gosnmpWrapper{&gosnmp.GoSNMP{}}

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

	gs.MaxRepetitions = int(s.MaxRepetitions)

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

	s.connectionCache[agent] = gs
	return gs, nil
}

// fieldConvert converts from any type according to the conv specification
//  "float"/"float(0)" will convert the value into a float.
//  "float(X)" will convert the value into a float, and then move the decimal before Xth right-most digit.
//  "int" will convert the value into an integer.
//  "" will convert a byte slice into a string.
// Any other conv will return the input value unchanged.
func fieldConvert(conv string, v interface{}) interface{} {
	if conv == "" {
		if bs, ok := v.([]byte); ok {
			return string(bs)
		}
		return v
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
	}

	return v
}

// snmpTranslate resolves the given OID.
// The contents of the oid parameter will be replaced with the numeric oid value.
// If name is empty, the textual OID value is stored in it. If the textual OID cannot be translated, the numeric OID is stored instead.
// If mibPrefix is non-nil, the MIB in which the OID was found is stored, with a suffix of "::".
func snmpTranslate(mibPrefix *string, oid *string, name *string) error {
	if strings.ContainsAny(*oid, ":abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		out, err := execCmd("snmptranslate", "-m", "all", "-On", *oid)
		if err != nil {
			return Errorf(err, "translating %s", *oid)
		}
		*oid = string(bytes.TrimSuffix(out, []byte{'\n'}))
	}

	if *name == "" {
		out, err := execCmd("snmptranslate", "-m", "all", *oid)
		if err != nil {
			//TODO debug message
			*name = *oid
		} else {
			if i := bytes.Index(out, []byte("::")); i != -1 {
				if mibPrefix != nil {
					*mibPrefix = string(out[:i+2])
				}
				out = out[i+2:]
			}
			*name = string(bytes.TrimSuffix(out, []byte{'\n'}))
		}
	}

	return nil
}
