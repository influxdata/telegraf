package snmp

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
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

  ## Each 'tag' is an "snmpget" request. Tags are inherited by snmp 'walk'
  ## and 'get' requests specified below. If a name for the tag is not provided,
  ## we will attempt to get the MIB name.
  [[inputs.snmp.tag]]
    name = "sys_name" # optional, tag name
    oid = ".1.3.6.1.2.1.1.5.0"
  [[inputs.snmp.tag]]
    name = "sys_location"
    oid = ".1.3.6.1.2.1.1.6.0"

  ## optional, name of the measurement that each 'get' field will be under.
  name = "snmp"
  ## Each 'get' is an "snmpget" request. If a name for the field is not provided,
  ## we will attempt to get the MIB name.
  [[inputs.snmp.get]]
    name = "snmp_in_packets" # optional, field name
    oid = ".1.3.6.1.2.1.11.1.0"
  [[inputs.snmp.get]]
    oid = ".1.3.6.1.2.1.11.2.0"

  ## An SNMP walk will do an "snmpwalk" from the given root OID.
  ## Each OID it encounters will be converted into a field on the measurement.
  ## We will attempt to get the MIB names for each field.
  [[inputs.snmp.walk]]
    ## optional, inherit_tags specifies which top-level tags to inherit.
    ## Globs are supported.
    inherit_tags = ["sys_*"]
    name = "snmp_metrics" # measurement name
    root_oid = ".1.3.6.1.2.1.11"

  [[inputs.snmp.walk]]
    ## optional, 'include' specifies MIB names to include in the walk.
    ## 'exclude' is also available, although they're mutually-exclusive.
    ## Globs are supported.
    include = ["if*"]
    name = "ifTable"
    root_oid = ".1.3.6.1.2.1.2.2"

  [[inputs.snmp.walk]]
    exclude = ["ifAlias"]
    name = "ifXTable"
    root_oid = ".1.3.6.1.2.1.31.1.1"
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

	// Name & Fields are the elements of a Table.
	// Telegraf chokes if we try to embed a Table. So instead we have to embed
	// the fields of a Table, and construct a Table during runtime.
	Name  string
	Gets  []Get  `toml:"get"`
	Walks []Walk `toml:"walk"`
	Tags  []Tag  `toml:"tag"`

	// oidToMib is a map of OIDs to MIBs.
	oidToMib map[string]string
	// translateBin is the location of the 'snmptranslate' binary.
	translateBin string

	connectionCache map[string]snmpConnection
}

// Get holds the configuration for a Get to look up.
type Get struct {
	// Name will be the name of the field.
	Name string
	// OID is prefix for this field.
	Oid string
}

// Walker holds the configuration for a Walker to look up.
type Walk struct {
	// Name will be the name of the measurement.
	Name string
	// OID is prefix for this field. The plugin will perform a walk through all
	// OIDs with this as their parent. For each value found, the plugin will strip
	// off the OID prefix, and use the remainder as the index. For multiple fields
	// to show up in the same row, they must share the same index.
	RootOid string
	// Include is a list of glob filters, only OID/MIBs that match these will
	// be included in the resulting output metrics.
	Include []string
	include filter.Filter
	// Exclude is the exact opposite of Include
	Exclude []string
	exclude filter.Filter
	// InheritTags is a list of tags to inherit from the top-level snmp.tag
	// configuration.
	InheritTags []string
	inheritTags filter.Filter
}

// Tag holds the config for a tag.
type Tag struct {
	// Name will be the name of the tag.
	Name string
	// OID is prefix for this field.
	Oid string
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
	for _, agent := range s.Agents {
		gs, err := s.getConnection(agent)
		if err != nil {
			acc.AddError(fmt.Errorf("Agent %s, err: %s", agent, err))
			continue
		}

		topTags := map[string]string{}
		// Gather all snmp tags
		for _, t := range s.Tags {
			tagval, err := get(gs, t.Oid)
			if err != nil {
				acc.AddError(fmt.Errorf("Agent %s, err: %s", agent, err))
				continue
			}
			if tagval == nil {
				continue
			}
			name := t.Name
			if name == "" {
				name, _ = s.getMibName(t.Oid)
			}
			if s, ok := tagval.(string); ok {
				topTags[name] = s
			}
		}

		// Gather all snmp gets
		fields := map[string]interface{}{}
		for _, g := range s.Gets {
			val, err := get(gs, g.Oid)
			if err != nil {
				acc.AddError(fmt.Errorf("Agent %s, err: %s", agent, err))
				continue
			}
			if val == nil {
				continue
			}
			name := g.Name
			if name == "" {
				name, _ = s.getMibName(g.Oid)
			}
			fields[name] = val
		}
		if len(fields) > 0 {
			acc.AddFields(s.Name, fields, topTags, time.Now())
		}

		// Gather all snmp walks
		for _, w := range s.Walks {
			w.compileWalkFilters(acc)
			allfields := map[string]map[string]interface{}{}
			now := time.Now()
			s.walk(gs, allfields, w.RootOid, w.include, w.exclude)
			for index, wfields := range allfields {
				tags := copyTags(topTags, w.inheritTags)
				tags["index"] = index
				acc.AddFields(w.Name, wfields, tags, now)
			}
		}
	}

	return nil
}

// walk does a walk and populates the given 'fields' map with whatever it finds.
// as it goes, it attempts to lookup the MIB name of each OID it encounters.
func (s *Snmp) walk(
	gs snmpConnection,
	fields map[string]map[string]interface{},
	rootOid string,
	include filter.Filter,
	exclude filter.Filter,
) {
	gs.Walk(rootOid, func(ent gosnmp.SnmpPDU) error {
		name, index := s.getMibName(ent.Name)
		if include != nil {
			if !include.Match(name) {
				return nil
			}
		} else if exclude != nil {
			if exclude.Match(name) {
				return nil
			}
		}
		if _, ok := fields[index]; ok {
			fields[index][name] = normalize(ent.Value)
		} else {
			fields[index] = map[string]interface{}{
				name: normalize(ent.Value),
			}
		}
		return nil
	})
}

// get simply gets the given OID and converts it to the given type.
func get(gs snmpConnection, oid string) (interface{}, error) {
	pkt, err := gs.Get([]string{oid})
	if err != nil {
		return nil, fmt.Errorf("Error performing get: %s", err)
	}
	if pkt != nil && len(pkt.Variables) > 0 && pkt.Variables[0].Type != gosnmp.NoSuchObject {
		ent := pkt.Variables[0]
		return normalize(ent.Value), nil
	}
	return nil, nil
}

func (s *Snmp) getMibName(oid string) (string, string) {
	name, ok := s.oidToMib[oid]
	if !ok {
		// lookup the mib using snmptranslate
		name = lookupOidName(s.translateBin, oid)
		s.oidToMib[oid] = name
	}
	return separateIndex(name)
}

func (w *Walk) compileWalkFilters(acc telegraf.Accumulator) {
	// if it's the first gather, compile any inherit_tags filter:
	if len(w.InheritTags) > 0 && w.inheritTags == nil {
		var err error
		w.inheritTags, err = filter.CompileFilter(w.InheritTags)
		if err != nil {
			acc.AddError(err)
		}
	}
	// if it's the first gather, compile any include filter:
	if len(w.Include) > 0 && w.include == nil {
		var err error
		w.include, err = filter.CompileFilter(w.Include)
		if err != nil {
			acc.AddError(err)
		}
	}
	// if it's the first gather, compile any exclude filter:
	if len(w.Exclude) > 0 && w.exclude == nil {
		var err error
		w.exclude, err = filter.CompileFilter(w.Exclude)
		if err != nil {
			acc.AddError(err)
		}
	}
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
			return nil, fmt.Errorf("reconnecting %s", err)
		}
		host = agent
		portStr = "161"
	}
	gs.Target = host

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("reconnecting %s", err)
	}
	gs.Port = uint16(port)

	if s.Timeout != "" {
		if gs.Timeout, err = time.ParseDuration(s.Timeout); err != nil {
			return nil, fmt.Errorf("reconnecting %s", err)
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
	}

	if err := gs.Connect(); err != nil {
		return nil, fmt.Errorf("setting up connection %s", err)
	}

	s.connectionCache[agent] = gs
	return gs, nil
}

// normalize normalizes the given interface for metric storage.
func normalize(v interface{}) interface{} {
	switch vt := v.(type) {
	case []byte:
		v = string(vt)
	}

	return v
}

func copyTags(in map[string]string, inheritTags filter.Filter) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		if inheritTags != nil {
			if !inheritTags.Match(k) {
				continue
			}
		}
		out[k] = v
	}
	return out
}

// lookupOidName looks up the MIB name of the given OID using the provided
// snmptranslate binary. If a name is not found, then we just return the OID.
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

// separateIndex takes an input string (either a MIB or an OID) and separates
// out the index from it, ie:
//   ifName.1     -> (ifName, 1)
//   snmpInPkts.0 -> (snmpInPkts, 0)
//   .1.3.6.4.2.0 -> (.1.3.6.4.2, 0)
func separateIndex(in string) (string, string) {
	items := strings.Split(in, ".")
	if len(items) == 1 {
		return in, "0"
	}
	index := items[len(items)-1]
	return strings.Join(items[0:len(items)-1], "."), index
}

func init() {
	bin, _ := exec.LookPath("snmptranslate")
	inputs.Add("snmp", func() telegraf.Input {
		return &Snmp{
			Name:           "snmp",
			Retries:        5,
			MaxRepetitions: 50,
			translateBin:   bin,
			oidToMib:       make(map[string]string),
		}
	})
}
