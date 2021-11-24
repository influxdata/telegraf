package snmp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gosnmp/gosnmp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/snmp"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/sleepinggenius2/gosmi"
	"github.com/sleepinggenius2/gosmi/types"
)

const description = `Retrieves SNMP values from remote agents`
const sampleConfig = `
  ## Agent addresses to retrieve values from.
  ##   format:  agents = ["<scheme://><hostname>:<port>"]
  ##   scheme:  optional, either udp, udp4, udp6, tcp, tcp4, tcp6.
  ##            default is udp
  ##   port:    optional
  ##   example: agents = ["udp://127.0.0.1:161"]
  ##            agents = ["tcp://127.0.0.1:161"]
  ##            agents = ["udp4://v4only-snmp-agent"]
  agents = ["udp://127.0.0.1:161"]

  ## Timeout for each request.
  # timeout = "5s"

  ## SNMP version; can be 1, 2, or 3.
  # version = 2

  ## Path to mib files
  # path = ["/usr/share/snmp/mibs"]

  ## Agent host tag; the tag used to reference the source host
  # agent_host_tag = "agent_host"

  ## SNMP community string.
  # community = "public"

  ## Number of retries to attempt.
  # retries = 3

  ## The GETBULK max-repetitions parameter.
  # max_repetitions = 10

  ## SNMPv3 authentication and encryption options.
  ##
  ## Security Name.
  # sec_name = "myuser"
  ## Authentication protocol; one of "MD5", "SHA", "SHA224", "SHA256", "SHA384", "SHA512" or "".
  # auth_protocol = "MD5"
  ## Authentication password.
  # auth_password = "pass"
  ## Security Level; one of "noAuthNoPriv", "authNoPriv", or "authPriv".
  # sec_level = "authNoPriv"
  ## Context Name.
  # context_name = ""
  ## Privacy protocol used for encrypted messages; one of "DES", "AES" or "".
  # priv_protocol = ""
  ## Privacy password used for encrypted messages.
  # priv_password = ""
  
  ## Add fields and tables defining the variables you wish to collect.  This
  ## example collects the system uptime and interface variables.  Reference the
  ## full plugin documentation for configuration details.
`

// Snmp holds the configuration for the plugin.
type Snmp struct {
	// The SNMP agent to query. Format is [SCHEME://]ADDR[:PORT] (e.g.
	// udp://1.2.3.4:161).  If the scheme is not specified then "udp" is used.
	Agents []string `toml:"agents"`

	// The tag used to name the agent host
	AgentHostTag string `toml:"agent_host_tag"`

	snmp.ClientConfig

	Tables []Table `toml:"table"`

	// Name & Fields are the elements of a Table.
	// Telegraf chokes if we try to embed a Table. So instead we have to embed the
	// fields of a Table, and construct a Table during runtime.
	Name   string  // deprecated in 1.14; use name_override
	Fields []Field `toml:"field"`

	connectionCache []snmpConnection
	initialized     bool

	Log telegraf.Logger `toml:"-"`
}

func (s *Snmp) init() error {
	if s.initialized {
		return nil
	}

	err := s.getMibsPath()
	if err != nil {
		return err
	}

	s.connectionCache = make([]snmpConnection, len(s.Agents))

	for i := range s.Tables {
		if err := s.Tables[i].Init(); err != nil {
			return fmt.Errorf("initializing table %s: %w", s.Tables[i].Name, err)
		}
	}

	for i := range s.Fields {
		if err := s.Fields[i].init(); err != nil {
			return fmt.Errorf("initializing field %s: %w", s.Fields[i].Name, err)
		}
	}

	if len(s.AgentHostTag) == 0 {
		s.AgentHostTag = "agent_host"
	}

	s.initialized = true
	return nil
}

func (s *Snmp) getMibsPath() error {
	gosmi.Init()
	var folders []string
	for _, mibPath := range s.Path {
		gosmi.AppendPath(mibPath)
		folders = append(folders, mibPath)
		err := filepath.Walk(mibPath, func(path string, info os.FileInfo, err error) error {
			// symlinks are files so we need to double check if any of them are folders
			// Will check file vs directory later on
			if info.Mode()&os.ModeSymlink != 0 {
				link, err := os.Readlink(path)
				if err != nil {
					s.Log.Warnf("Bad symbolic link %v", link)
				}
				folders = append(folders, link)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("Filepath could not be walked %v", err)
		}
		for _, folder := range folders {
			err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
				// checks if file or directory
				if info.IsDir() {
					gosmi.AppendPath(path)
				} else if info.Mode()&os.ModeSymlink == 0 {
					_, err := gosmi.LoadModule(info.Name())
					if err != nil {
						s.Log.Warnf("Module could not be loaded %v", err)
					}
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("Filepath could not be walked %v", err)
			}
		}
		folders = []string{}
	}
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

// Init() builds & initializes the nested fields.
func (t *Table) Init() error {
	//makes sure oid or name is set in config file
	//otherwise snmp will produce metrics with an empty name
	if t.Oid == "" && t.Name == "" {
		return fmt.Errorf("SNMP table in config file is not named. One or both of the oid and name settings must be set")
	}

	if t.initialized {
		return nil
	}

	if err := t.initBuild(); err != nil {
		return err
	}

	secondaryIndexTablePresent := false
	// initialize all the nested fields
	for i := range t.Fields {
		if err := t.Fields[i].init(); err != nil {
			return fmt.Errorf("initializing field %s: %w", t.Fields[i].Name, err)
		}
		if t.Fields[i].SecondaryIndexTable {
			if secondaryIndexTablePresent {
				return fmt.Errorf("only one field can be SecondaryIndexTable")
			}
			secondaryIndexTablePresent = true
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
	// Translate tells if the value of the field should be snmptranslated
	Translate bool
	// Secondary index table allows to merge data from two tables with different index
	//  that this filed will be used to join them. There can be only one secondary index table.
	SecondaryIndexTable bool
	// This field is using secondary index, and will be later merged with primary index
	//  using SecondaryIndexTable. SecondaryIndexTable and SecondaryIndexUse are exclusive.
	SecondaryIndexUse bool
	// Controls if entries from secondary table should be added or not if joining
	//  index is present or not. I set to true, means that join is outer, and
	//  index is prepended with "Secondary." for missing values to avoid overlaping
	//  indexes from both tables.
	// Can be set per field or globally with SecondaryIndexTable, global true overrides
	//  per field false.
	SecondaryOuterJoin bool

	initialized bool
}

// init() converts OID names to numbers, and sets the .Name attribute if unset.
func (f *Field) init() error {
	if f.initialized {
		return nil
	}

	// check if oid needs translation or name is not set
	if strings.ContainsAny(f.Oid, ":abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") || f.Name == "" {
		_, oidNum, oidText, conversion, err := SnmpTranslate(f.Oid)
		if err != nil {
			return fmt.Errorf("translating: %w", err)
		}
		f.Oid = oidNum
		if f.Name == "" {
			f.Name = oidText
		}
		if f.Conversion == "" {
			f.Conversion = conversion
		}
		//TODO use textual convention conversion from the MIB
	}

	if f.SecondaryIndexTable && f.SecondaryIndexUse {
		return fmt.Errorf("SecondaryIndexTable and UseSecondaryIndex are exclusive")
	}

	if !f.SecondaryIndexTable && !f.SecondaryIndexUse && f.SecondaryOuterJoin {
		return fmt.Errorf("SecondaryOuterJoin set to true, but field is not being used in join")
	}

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

type walkError struct {
	msg string
	err error
}

func (e *walkError) Error() string {
	return e.msg
}

func (e *walkError) Unwrap() error {
	return e.err
}

func init() {
	inputs.Add("snmp", func() telegraf.Input {
		return &Snmp{
			Name: "snmp",
			ClientConfig: snmp.ClientConfig{
				Retries:        3,
				MaxRepetitions: 10,
				Timeout:        config.Duration(5 * time.Second),
				Version:        2,
				Path:           []string{"/usr/share/snmp/mibs"},
				Community:      "public",
			},
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
				acc.AddError(fmt.Errorf("agent %s: %w", agent, err))
				return
			}

			// First is the top-level fields. We treat the fields as table prefixes with an empty index.
			t := Table{
				Name:   s.Name,
				Fields: s.Fields,
			}
			topTags := map[string]string{}
			if err := s.gatherTable(acc, gs, t, topTags, false); err != nil {
				acc.AddError(fmt.Errorf("agent %s: %w", agent, err))
			}

			// Now is the real tables.
			for _, t := range s.Tables {
				if err := s.gatherTable(acc, gs, t, topTags, true); err != nil {
					acc.AddError(fmt.Errorf("agent %s: gathering table %s: %w", agent, t.Name, err))
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
		if _, ok := tr.Tags[s.AgentHostTag]; !ok {
			tr.Tags[s.AgentHostTag] = gs.Host()
		}
		acc.AddFields(rt.Name, tr.Fields, tr.Tags, rt.Time)
	}

	return nil
}

// Build retrieves all the fields specified in the table and constructs the RTable.
func (t Table) Build(gs snmpConnection, walk bool) (*RTable, error) {
	rows := map[string]RTableRow{}

	//translation table for secondary index (when preforming join on two tables)
	secIdxTab := make(map[string]string)
	secGlobalOuterJoin := false
	for i, f := range t.Fields {
		if f.SecondaryIndexTable {
			secGlobalOuterJoin = f.SecondaryOuterJoin
			if i != 0 {
				t.Fields[0], t.Fields[i] = t.Fields[i], t.Fields[0]
			}
			break
		}
	}

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
				if errors.Is(err, gosnmp.ErrUnknownSecurityLevel) {
					return nil, fmt.Errorf("unknown security level (sec_level)")
				} else if errors.Is(err, gosnmp.ErrUnknownUsername) {
					return nil, fmt.Errorf("unknown username (sec_name)")
				} else if errors.Is(err, gosnmp.ErrWrongDigest) {
					return nil, fmt.Errorf("wrong digest (auth_protocol, auth_password)")
				} else if errors.Is(err, gosnmp.ErrDecryption) {
					return nil, fmt.Errorf("decryption error (priv_protocol, priv_password)")
				} else {
					return nil, fmt.Errorf("performing get on field %s: %w", f.Name, err)
				}
			} else if pkt != nil && len(pkt.Variables) > 0 && pkt.Variables[0].Type != gosnmp.NoSuchObject && pkt.Variables[0].Type != gosnmp.NoSuchInstance {
				ent := pkt.Variables[0]
				fv, err := fieldConvert(f.Conversion, ent.Value)
				if err != nil {
					return nil, fmt.Errorf("converting %q (OID %s) for field %s: %w", ent.Value, ent.Name, f.Name, err)
				}
				ifv[""] = fv
			}
		} else {
			err := gs.Walk(oid, func(ent gosnmp.SnmpPDU) error {
				if len(ent.Name) <= len(oid) || ent.Name[:len(oid)+1] != oid+"." {
					return &walkError{} // break the walk
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
							i--
						}
						if i < 1 {
							return -1
						}
						return r
					}, idx)
				}

				// snmptranslate table field value here
				if f.Translate {
					if entOid, ok := ent.Value.(string); ok {
						_, _, oidText, _, err := SnmpTranslate(entOid)
						if err == nil {
							// If no error translating, the original value for ent.Value should be replaced
							ent.Value = oidText
						}
					}
				}

				fv, err := fieldConvert(f.Conversion, ent.Value)
				if err != nil {
					return &walkError{
						msg: fmt.Sprintf("converting %q (OID %s) for field %s", ent.Value, ent.Name, f.Name),
						err: err,
					}
				}
				ifv[idx] = fv
				return nil
			})
			if err != nil {
				// Our callback always wraps errors in a walkError.
				// If this error isn't a walkError, we know it's not
				// from the callback
				if _, ok := err.(*walkError); !ok {
					return nil, fmt.Errorf("performing bulk walk for field %s: %w", f.Name, err)
				}
			}
		}

		for idx, v := range ifv {
			if f.SecondaryIndexUse {
				if newidx, ok := secIdxTab[idx]; ok {
					idx = newidx
				} else {
					if !secGlobalOuterJoin && !f.SecondaryOuterJoin {
						continue
					}
					idx = ".Secondary" + idx
				}
			}
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
				if f.SecondaryIndexTable {
					//indexes are stored here with prepending "." so we need to add them if needed
					var vss string
					if ok {
						vss = "." + vs
					} else {
						vss = fmt.Sprintf(".%v", v)
					}
					if idx[0] == '.' {
						secIdxTab[vss] = idx
					} else {
						secIdxTab[vss] = "." + idx
					}
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

// getConnection creates a snmpConnection (*gosnmp.GoSNMP) object and caches the
// result using `agentIndex` as the cache key.  This is done to allow multiple
// connections to a single address.  It is an error to use a connection in
// more than one goroutine.
func (s *Snmp) getConnection(idx int) (snmpConnection, error) {
	if gs := s.connectionCache[idx]; gs != nil {
		return gs, nil
	}

	agent := s.Agents[idx]

	var err error
	var gs snmp.GosnmpWrapper
	gs, err = snmp.NewWrapper(s.ClientConfig)
	if err != nil {
		return nil, err
	}

	err = gs.SetAgent(agent)
	if err != nil {
		return nil, err
	}

	s.connectionCache[idx] = gs

	if err := gs.Connect(); err != nil {
		return nil, fmt.Errorf("setting up connection: %w", err)
	}

	return gs, nil
}

// fieldConvert converts from any type according to the conv specification
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
			v = vt / math.Pow10(d)
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
			v = vt
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
			v, _ = strconv.ParseInt(string(vt), 10, 64)
		case string:
			v, _ = strconv.ParseInt(vt, 10, 64)
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

	split := strings.Split(conv, ":")
	if split[0] == "hextoint" && len(split) == 3 {
		endian := split[1]
		bit := split[2]

		bv, ok := v.([]byte)
		if !ok {
			return v, nil
		}

		switch endian {
		case "LittleEndian":
			switch bit {
			case "uint64":
				v = binary.LittleEndian.Uint64(bv)
			case "uint32":
				v = binary.LittleEndian.Uint32(bv)
			case "uint16":
				v = binary.LittleEndian.Uint16(bv)
			default:
				return nil, fmt.Errorf("invalid bit value (%s) for hex to int conversion", bit)
			}
		case "BigEndian":
			switch bit {
			case "uint64":
				v = binary.BigEndian.Uint64(bv)
			case "uint32":
				v = binary.BigEndian.Uint32(bv)
			case "uint16":
				v = binary.BigEndian.Uint16(bv)
			default:
				return nil, fmt.Errorf("invalid bit value (%s) for hex to int conversion", bit)
			}
		default:
			return nil, fmt.Errorf("invalid Endian value (%s) for hex to int conversion", endian)
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
	mibName, oidNum, oidText, _, err = SnmpTranslate(oid)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("translating: %w", err)
	}

	mibPrefix := mibName + "::"

	// first attempt to get the table's tags
	tagOids := map[string]struct{}{}
	// mimcks grabbing INDEX {} that is returned from snmptranslate -Td MibName
	node, err := gosmi.GetNodeByOID(types.OidMustFromString(oidNum))

	if err != nil {
		return "", "", "", nil, fmt.Errorf("getting submask: %w", err)
	}

	for _, index := range node.GetIndex() {
		tagOids[mibPrefix+index.Name] = struct{}{}
	}

	// grabs all columns from the table
	// mimmicks grabbing everything returned from snmptable -Ch -Cl -c public 127.0.0.1 oidFullName
	col := node.GetRow().AsTable().ColumnOrder

	for _, c := range col {
		_, isTag := tagOids[mibPrefix+c]
		fields = append(fields, Field{Name: c, Oid: mibPrefix + c, IsTag: isTag})
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
func SnmpTranslate(oid string) (mibName string, oidNum string, oidText string, conversion string, err error) {
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
		// snmpTranslation.Lock to release. But I don't know that the extra complexity
		// is worth it. Especially when it would slam the system pretty hard if lots
		// of lookups are being performed.

		stc.mibName, stc.oidNum, stc.oidText, stc.conversion, stc.err = snmpTranslateCall(oid)
		snmpTranslateCaches[oid] = stc
	}

	snmpTranslateCachesLock.Unlock()

	return stc.mibName, stc.oidNum, stc.oidText, stc.conversion, stc.err
}

func snmpTranslateCall(oid string) (mibName string, oidNum string, oidText string, conversion string, err error) {
	var out gosmi.SmiNode
	var end string
	if strings.ContainsAny(oid, "::") {
		// split given oid
		// for example RFC1213-MIB::sysUpTime.0
		s := strings.Split(oid, "::")
		// node becomes sysUpTime.0
		node := s[1]
		if strings.ContainsAny(node, ".") {
			s = strings.Split(node, ".")
			// node becomes sysUpTime
			node = s[0]
			end = "." + s[1]
		}

		out, err = gosmi.GetNode(node)
		if err != nil {
			return oid, oid, oid, oid, err
		}

		oidNum = "." + out.RenderNumeric() + end
	} else if strings.ContainsAny(oid, "abcdefghijklnmopqrstuvwxyz") {
		//handle mixed oid ex. .iso.2.3
		s := strings.Split(oid, ".")
		for i := range s {
			if strings.ContainsAny(s[i], "abcdefghijklmnopqrstuvwxyz") {
				out, err = gosmi.GetNode(s[i])
				if err != nil {
					return oid, oid, oid, oid, err
				}
				s[i] = out.RenderNumeric()
			}
		}
		oidNum = strings.Join(s, ".")
		out, _ = gosmi.GetNodeByOID(types.OidMustFromString(oidNum))
	} else {
		out, err = gosmi.GetNodeByOID(types.OidMustFromString(oid))
		oidNum = oid
		// ensure modules are loaded or node will be empty (might not error)
		// do not return the err as the oid is numeric and telegraf can continue
		//nolint:nilerr
		if err != nil || out.Name == "iso" {
			return oid, oid, oid, oid, nil
		}
	}

	tc := out.GetSubtree()

	for i := range tc {
		// case where the mib doesn't have a conversion so Type struct will be nil
		// prevents seg fault
		if tc[i].Type == nil {
			break
		}
		switch tc[i].Type.Name {
		case "MacAddress", "PhysAddress":
			conversion = "hwaddr"
		case "InetAddressIPv4", "InetAddressIPv6", "InetAddress", "IPSIpAddress":
			conversion = "ipaddr"
		}
	}

	oidText = out.RenderQualified()
	i := strings.Index(oidText, "::")
	if i == -1 {
		return "", oid, oid, oid, fmt.Errorf("not found")
	}
	mibName = oidText[:i]
	oidText = oidText[i+2:] + end

	return mibName, oidNum, oidText, conversion, nil
}
