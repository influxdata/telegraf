package snmp

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
)

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
	translator  Translator
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

// Init() builds & initializes the nested fields.
func (t *Table) Init(tr Translator) error {
	//makes sure oid or name is set in config file
	//otherwise snmp will produce metrics with an empty name
	if t.Oid == "" && t.Name == "" {
		return errors.New("SNMP table in config file is not named. One or both of the oid and name settings must be set")
	}

	if t.initialized {
		return nil
	}

	t.translator = tr
	if err := t.initBuild(); err != nil {
		return err
	}

	secondaryIndexTablePresent := false
	// initialize all the nested fields
	for i := range t.Fields {
		if err := t.Fields[i].Init(t.translator); err != nil {
			return fmt.Errorf("initializing field %s: %w", t.Fields[i].Name, err)
		}
		if t.Fields[i].SecondaryIndexTable {
			if secondaryIndexTablePresent {
				return errors.New("only one field can be SecondaryIndexTable")
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

	_, _, oidText, fields, err := t.translator.SnmpTable(t.Oid)
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

// Build retrieves all the fields specified in the table and constructs the RTable.
func (t Table) Build(gs Connection, walk bool) (*RTable, error) {
	rows := map[string]RTableRow{}

	//translation table for secondary index (when performing join on two tables)
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
					return nil, errors.New("unknown security level (sec_level)")
				} else if errors.Is(err, gosnmp.ErrUnknownUsername) {
					return nil, errors.New("unknown username (sec_name)")
				} else if errors.Is(err, gosnmp.ErrWrongDigest) {
					return nil, errors.New("wrong digest (auth_protocol, auth_password)")
				} else if errors.Is(err, gosnmp.ErrDecryption) {
					return nil, errors.New("decryption error (priv_protocol, priv_password)")
				}
				return nil, fmt.Errorf("performing get on field %s: %w", f.Name, err)
			} else if pkt != nil && len(pkt.Variables) > 0 && pkt.Variables[0].Type != gosnmp.NoSuchObject && pkt.Variables[0].Type != gosnmp.NoSuchInstance {
				ent := pkt.Variables[0]
				fv, err := f.Convert(ent)
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
						_, _, oidText, _, err := t.translator.SnmpTranslate(entOid)
						if err == nil {
							// If no error translating, the original value for ent.Value should be replaced
							ent.Value = oidText
						}
					}
				}

				fv, err := f.Convert(ent)
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
				var walkErr *walkError
				if !errors.As(err, &walkErr) {
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
