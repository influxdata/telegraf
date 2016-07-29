package snmp

import (
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
  agents = ["127.0.0.1:161"]
  version = 2 # Values: 1, 2, or 3

  ## SNMPv1 & SNMPv2 parameters
  community = "public"

  ## SNMPv2 & SNMPv3 parameters
  max_repetitions = 50

  ## SNMPv3 parameters
  #sec_name = "myuser"
  #auth_protocol = "md5"         # Values: "MD5", "SHA", ""
  #auth_password = "password123"
  #sec_level = "authNoPriv"      # Values: "noAuthNoPriv", "authNoPriv", "authPriv"
  #context_name = ""
  #priv_protocol = ""            # Values: "DES", "AES", ""
  #priv_password = ""

  ## measurement name
  name = "system"
  ## SNMP fields are gotten by using an "snmpget" request. If a name is not
  ## specified, we attempt to use snmptranslate on the OID to get the MIB name.
  [[inputs.snmp.field]]
    name = "hostname"
    oid = ".1.2.3.0.1.1"
  [[inputs.snmp.field]]
    name = "uptime"
    oid = ".1.2.3.0.1.200"
  [[inputs.snmp.field]]
    oid = ".1.2.3.0.1.201"

  [[inputs.snmp.table]]
    ## measurement name
    name = "remote_servers"
    inherit_tags = ["hostname"]
    ## SNMP table fields must be specified individually. If the table field has
    ## multiple rows, they will all be gotten.
    [[inputs.snmp.table.field]]
      name = "server"
      oid = ".1.2.3.0.0.0"
      is_tag = true
    [[inputs.snmp.table.field]]
      name = "connections"
      oid = ".1.2.3.0.0.1"
    [[inputs.snmp.table.field]]
      name = "latency"
      oid = ".1.2.3.0.0.2"
`

// Snmp holds the configuration for the plugin.
type Snmp struct {
	// The SNMP agent to query. Format is ADDR[:PORT] (e.g. 1.2.3.4:161).
	Agents []string
	// Timeout to wait for a response. Value is anything accepted by time.ParseDuration().
	Timeout string
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

	inited bool
}

// Table holds the configuration for a SNMP table.
type Table struct {
	// Name will be the name of the measurement.
	Name string

	// Which tags to inherit from the top-level config.
	InheritTags []string

	// Fields is the tags and values to look up.
	Fields []Field `toml:"field"`
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

// Errors is a list of errors accumulated during an interval.
type Errors []error

func (errs Errors) Error() string {
	s := ""
	for _, err := range errs {
		if s == "" {
			s = err.Error()
		} else {
			s = s + ". " + err.Error()
		}
	}
	return s
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
	if !s.inited {
		s.initOidNames()
	}
	s.inited = true
	var errs Errors
	for _, agent := range s.Agents {
		gs, err := s.getConnection(agent)
		if err != nil {
			errs = append(errs, Errorf(err, "agent %s", agent))
			continue
		}

		// First is the top-level fields. We treat the fields as table prefixes with an empty index.
		t := Table{
			Name:   s.Name,
			Fields: s.Fields,
		}
		topTags := map[string]string{}
		if err := s.gatherTable(acc, gs, t, topTags, false); err != nil {
			errs = append(errs, Errorf(err, "agent %s", agent))
		}

		// Now is the real tables.
		for _, t := range s.Tables {
			if err := s.gatherTable(acc, gs, t, topTags, true); err != nil {
				errs = append(errs, Errorf(err, "agent %s", agent))
			}
		}
	}

	if errs == nil {
		return nil
	}
	return errs
}

// initOidNames loops through each [[inputs.snmp.field]] defined.
// If the field doesn't have a 'name' defined, it will attempt to use
// snmptranslate to get a name for the OID. If snmptranslate doesn't return a
// name, or snmptranslate is not available, then use the OID as the name.
func (s *Snmp) initOidNames() {
	bin, _ := exec.LookPath("snmptranslate")

	// Lookup names for each OID defined as a "field"
	for i, field := range s.Fields {
		if field.Name != "" {
			continue
		}
		s.Fields[i].Name = lookupOidName(bin, field.Oid)
	}

	// Lookup names for each OID defined as a "table.field"
	for i, table := range s.Tables {
		for j, field := range table.Fields {
			if field.Name != "" {
				continue
			}
			s.Tables[i].Fields[j].Name = lookupOidName(bin, field.Oid)
		}
	}
}

func lookupOidName(bin, oid string) string {
	name := oid
	if bin != "" {
		out, err := internal.CombinedOutputTimeout(
			exec.Command(bin, "-Os", oid),
			time.Millisecond*250)
		if err == nil && len(out) > 0 {
			name = strings.TrimSpace(string(out))
		}
	}
	return name
}

func (s *Snmp) gatherTable(
	acc telegraf.Accumulator,
	gs snmpConnection,
	t Table,
	topTags map[string]string,
	walk bool,
) error {
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

	if s.Timeout != "" {
		if gs.Timeout, err = time.ParseDuration(s.Timeout); err != nil {
			return nil, Errorf(err, "parsing timeout")
		}
	} else {
		gs.Timeout = time.Second * 1
	}

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

func init() {
	inputs.Add("snmp", func() telegraf.Input {
		return &Snmp{
			Retries:        5,
			MaxRepetitions: 50,
		}
	})
}
