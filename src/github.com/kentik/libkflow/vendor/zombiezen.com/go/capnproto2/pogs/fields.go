package pogs

import (
	"fmt"
	"reflect"
	"strings"

	"zombiezen.com/go/capnproto2/std/capnp/schema"
)

type fieldProps struct {
	schemaName string // empty if doesn't map to schema
	typ        fieldType
	fixedWhich string
	tagged     bool
}

type fieldType int

const (
	mappedField fieldType = iota
	whichField
	embedField
)

func parseField(f reflect.StructField, hasDiscrim bool) fieldProps {
	var p fieldProps
	tag := f.Tag.Get("capnp")
	p.tagged = tag != ""
	tname, opts := nextOpt(tag)
	switch tname {
	case "-":
		// omitted field
	case "":
		if f.Anonymous && isStructOrStructPtr(f.Type) {
			p.typ = embedField
			return p
		}
		if hasDiscrim && f.Name == "Which" {
			p.typ = whichField
			for len(opts) > 0 {
				var curr string
				curr, opts = nextOpt(opts)
				if strings.HasPrefix(curr, "which=") {
					p.fixedWhich = strings.TrimPrefix(curr, "which=")
					break
				}
			}
			return p
		}
		// TODO(light): check it's uppercase.
		x := f.Name[0] - 'A' + 'a'
		p.schemaName = string(x) + f.Name[1:]
	default:
		p.schemaName = tname
	}
	return p
}

func nextOpt(opts string) (head, tail string) {
	i := strings.Index(opts, ",")
	if i == -1 {
		return opts, ""
	}
	return opts[:i], opts[i+1:]
}

type fieldLoc struct {
	i    int
	path []int
}

func (loc fieldLoc) depth() int {
	if len(loc.path) > 0 {
		return len(loc.path)
	}
	return 1
}

func (loc fieldLoc) sub(i int) fieldLoc {
	n := len(loc.path)
	switch {
	case !loc.isValid():
		return fieldLoc{i: i}
	case n > 0:
		p := make([]int, n+1)
		copy(p, loc.path)
		p[n] = i
		return fieldLoc{path: p}
	default:
		return fieldLoc{path: []int{loc.i, i}}
	}
}

func (loc fieldLoc) isValid() bool {
	return loc.i >= 0
}

type structProps struct {
	fields     []fieldLoc
	whichLoc   fieldLoc // i == -1: none; i == -2: fixed
	fixedWhich uint16
}

func mapStruct(t reflect.Type, n schema.Node) (structProps, error) {
	fields, err := n.StructNode().Fields()
	if err != nil {
		return structProps{}, err
	}
	sp := structProps{
		fields:   make([]fieldLoc, fields.Len()),
		whichLoc: fieldLoc{i: -1},
	}
	for i := range sp.fields {
		sp.fields[i] = fieldLoc{i: -1}
	}
	sm := structMapper{
		sp:         &sp,
		t:          t,
		hasDiscrim: hasDiscriminant(n),
		fields:     fields,
	}
	if err := sm.visit(fieldLoc{i: -1}); err != nil {
		return structProps{}, err
	}
	for len(sm.embedQueue) > 0 {
		loc := sm.embedQueue[0]
		copy(sm.embedQueue, sm.embedQueue[1:])
		sm.embedQueue = sm.embedQueue[:len(sm.embedQueue)-1]
		if err := sm.visit(loc); err != nil {
			return structProps{}, err
		}
	}
	return sp, nil
}

type structMapper struct {
	sp         *structProps
	t          reflect.Type
	hasDiscrim bool
	fields     schema.Field_List
	embedQueue []fieldLoc
}

func (sm *structMapper) visit(base fieldLoc) error {
	t := sm.t
	if base.isValid() {
		t = typeFieldByLoc(t, base).Type
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" && !f.Anonymous {
			// unexported field
			continue
		}
		loc := base.sub(i)
		p := parseField(f, sm.hasDiscrim)
		if p.typ == embedField {
			sm.embedQueue = append(sm.embedQueue, loc)
			continue
		}
		if err := sm.visitField(loc, f, p); err != nil {
			return err
		}
	}
	return nil
}

func (sm *structMapper) visitField(loc fieldLoc, f reflect.StructField, p fieldProps) error {
	switch p.typ {
	case mappedField:
		if p.schemaName == "" {
			return nil
		}
		fi := fieldIndex(sm.fields, p.schemaName)
		if fi < 0 {
			return fmt.Errorf("%v has unknown field %s, maps to %s", sm.t, f.Name, p.schemaName)
		}
		switch oldloc := sm.sp.fields[fi]; {
		case oldloc.i == -2:
			// Prior tag collision, do nothing.
		case oldloc.i == -3 && p.tagged && loc.depth() == len(oldloc.path):
			// A tagged field wins over untagged fields.
			sm.sp.fields[fi] = loc
		case oldloc.isValid() && oldloc.depth() < loc.depth():
			// This is deeper, don't override.
		case oldloc.isValid() && oldloc.depth() == loc.depth():
			oldp := parseField(typeFieldByLoc(sm.t, oldloc), sm.hasDiscrim)
			if oldp.tagged && p.tagged {
				// Tag collision
				sm.sp.fields[fi] = fieldLoc{i: -2}
			} else if p.tagged {
				sm.sp.fields[fi] = loc
			} else if !oldp.tagged {
				// Multiple untagged fields.  Keep path, because we need it for depth.
				sm.sp.fields[fi].i = -3
			}
		default:
			sm.sp.fields[fi] = loc
		}
	case whichField:
		if sm.sp.whichLoc.i != -1 {
			return fmt.Errorf("%v embeds multiple Which fields", sm.t)
		}
		switch {
		case p.fixedWhich != "":
			fi := fieldIndex(sm.fields, p.fixedWhich)
			if fi < 0 {
				return fmt.Errorf("%v.Which is tagged with unknown field %s", sm.t, p.fixedWhich)
			}
			dv := sm.fields.At(fi).DiscriminantValue()
			if dv == schema.Field_noDiscriminant {
				return fmt.Errorf("%v.Which is tagged with non-union field %s", sm.t, p.fixedWhich)
			}
			sm.sp.whichLoc = fieldLoc{i: -2}
			sm.sp.fixedWhich = dv
		case f.Type.Kind() != reflect.Uint16:
			return fmt.Errorf("%v.Which is type %v, not uint16", sm.t, f.Type)
		default:
			sm.sp.whichLoc = loc
		}
	}
	return nil
}

// fieldBySchemaName returns the field for the given name.
// Returns an invalid value if the field was not found or it is
// contained inside a nil anonymous struct pointer.
func (sp structProps) fieldByOrdinal(val reflect.Value, i int) reflect.Value {
	return fieldByLoc(val, sp.fields[i], false)
}

// makeFieldBySchemaName returns the field for the given name, creating
// its parent anonymous structs if necessary.  Returns an invalid value
// if the field was not found.
func (sp structProps) makeFieldByOrdinal(val reflect.Value, i int) reflect.Value {
	return fieldByLoc(val, sp.fields[i], true)
}

// which returns the value of the discriminator field.
func (sp structProps) which(val reflect.Value) (discrim uint16, ok bool) {
	if sp.whichLoc.i == -2 {
		return sp.fixedWhich, true
	}
	f := fieldByLoc(val, sp.whichLoc, false)
	if !f.IsValid() {
		return 0, false
	}
	return uint16(f.Uint()), true
}

// setWhich sets the value of the discriminator field, creating its
// parent anonymous structs if necessary.  Returns whether the struct
// had a field to set.
func (sp structProps) setWhich(val reflect.Value, discrim uint16) error {
	if sp.whichLoc.i == -2 {
		if discrim != sp.fixedWhich {
			return fmt.Errorf("extract union field @%d into %v; expected @%d", discrim, val.Type(), sp.fixedWhich)
		}
		return nil
	}
	f := fieldByLoc(val, sp.whichLoc, true)
	if !f.IsValid() {
		return noWhichError{val.Type()}
	}
	f.SetUint(uint64(discrim))
	return nil
}

type noWhichError struct {
	t reflect.Type
}

func (e noWhichError) Error() string {
	return fmt.Sprintf("%v has no field Which", e.t)
}

func isNoWhichError(e error) bool {
	_, ok := e.(noWhichError)
	return ok
}

func fieldByLoc(val reflect.Value, loc fieldLoc, mkparents bool) reflect.Value {
	if !loc.isValid() {
		return reflect.Value{}
	}
	if len(loc.path) > 0 {
		for i, x := range loc.path {
			if i > 0 {
				if val.Kind() == reflect.Ptr {
					if val.IsNil() {
						if !mkparents {
							return reflect.Value{}
						}
						val.Set(reflect.New(val.Type().Elem()))
					}
					val = val.Elem()
				}
			}
			val = val.Field(x)
		}
		return val
	}
	return val.Field(loc.i)
}

func typeFieldByLoc(t reflect.Type, loc fieldLoc) reflect.StructField {
	if len(loc.path) > 0 {
		return t.FieldByIndex(loc.path)
	}
	return t.Field(loc.i)
}

func hasDiscriminant(n schema.Node) bool {
	return n.Which() == schema.Node_Which_structNode && n.StructNode().DiscriminantCount() > 0
}

func shortDisplayName(n schema.Node) []byte {
	dn, _ := n.DisplayNameBytes()
	return dn[n.DisplayNamePrefixLength():]
}

func fieldIndex(fields schema.Field_List, name string) int {
	for i := 0; i < fields.Len(); i++ {
		b, _ := fields.At(i).NameBytes()
		if bytesStrEqual(b, name) {
			return i
		}
	}
	return -1
}

func bytesStrEqual(b []byte, s string) bool {
	if len(b) != len(s) {
		return false
	}
	for i := range b {
		if b[i] != s[i] {
			return false
		}
	}
	return true
}

func isStructOrStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Struct || t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}
