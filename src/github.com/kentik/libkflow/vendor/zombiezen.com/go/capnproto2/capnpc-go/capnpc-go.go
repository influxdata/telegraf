/*
capnpc-go is the Cap'n proto code generator for Go.  It reads a
CodeGeneratorRequest from stdin and for a file foo.capnp it writes
foo.capnp.go.  This is usually invoked from `capnp compile -ogo`.

See https://capnproto.org/otherlang.html#how-to-write-compiler-plugins
for more details.
*/
package main

import (
	"bytes"
	"compress/zlib"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/std/capnp/schema"
)

// Non-stdlib import paths.
const (
	capnpImport   = "zombiezen.com/go/capnproto2"
	textImport    = capnpImport + "/encoding/text"
	schemasImport = capnpImport + "/schemas"
	serverImport  = capnpImport + "/server"
	contextImport = "golang.org/x/net/context"
)

// genoptions are parameters that control code generation.
// Usually passed on the command line.
type genoptions struct {
	promises      bool
	schemas       bool
	structStrings bool
}

type renderer interface {
	Render(name string, params interface{}) error
	Bytes() []byte
}

type templateRenderer struct {
	buf bytes.Buffer
	t   *template.Template
}

// Render calls ExecuteTemplate to render to its buffer.
func (tr *templateRenderer) Render(name string, params interface{}) error {
	return tr.t.ExecuteTemplate(&tr.buf, name, params)
}

// Bytes returns the accumulated bytes.
func (tr *templateRenderer) Bytes() []byte {
	return tr.buf.Bytes()
}

// generator builds up the generated code for a single file.
type generator struct {
	r       renderer
	fileID  uint64
	nodes   nodeMap
	imports imports
	data    staticData
	opts    genoptions
}

func newGenerator(fileID uint64, nodes nodeMap, opts genoptions) *generator {
	g := &generator{
		r:      &templateRenderer{t: templates},
		fileID: fileID,
		nodes:  nodes,
		opts:   opts,
	}
	g.imports.init()
	g.data.init(fileID)
	return g
}

// Basename returns the name of the schema file with the directory name removed.
func (g *generator) Basename() (string, error) {
	f, err := g.nodes.mustFind(g.fileID)
	if err != nil {
		return "", err
	}
	dn, err := f.DisplayName()
	if err != nil {
		return "", err
	}
	return filepath.Base(dn), nil
}

func (g *generator) Imports() *imports {
	return &g.imports
}

func (g *generator) Capnp() string {
	return g.imports.Capnp()
}

// generate produces unformatted Go source code from the nodes defined in it.
func (g *generator) generate() []byte {
	var out bytes.Buffer
	fmt.Fprintf(&out, "package %s\n\n", g.nodes[g.fileID].pkg)
	out.WriteString("// AUTO GENERATED - DO NOT EDIT\n\n")
	out.WriteString("import (\n")
	for _, imp := range g.imports.usedImports() {
		fmt.Fprintf(&out, "%v\n", imp)
	}
	out.WriteString(")\n")
	out.Write(g.r.Bytes())
	if len(g.data.buf) > 0 {
		writeByteLiteral(&out, g.data.name, g.data.buf)
	}
	return out.Bytes()
}

func writeByteLiteral(out *bytes.Buffer, name string, data []byte) {
	fmt.Fprintf(out, "var %s = []byte{", name)
	for i, b := range data {
		if i%8 == 0 {
			out.WriteByte('\n')
		}
		fmt.Fprintf(out, "%d,", b)
	}
	fmt.Fprintf(out, "\n}\n")
}

func (g *generator) defineSchemaVar() error {
	msg, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	req, _ := schema.NewRootCodeGeneratorRequest(seg)
	fnodes := g.nodes[g.fileID].nodes
	ids := make([]uint64, len(fnodes))
	for i, n := range fnodes {
		ids[i] = n.Id()
	}
	sort.Sort(uint64Slice(ids))
	// TODO(light): find largest object size and use that to allocate list
	nodes, _ := req.NewNodes(int32(len(g.nodes)))
	i := 0
	for _, id := range ids {
		n := g.nodes[id]
		if err := nodes.Set(i, n.Node); err != nil {
			return err
		}
		i++
	}
	var buf bytes.Buffer
	z, _ := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err := capnp.NewPackedEncoder(z).Encode(msg); err != nil {
		return err
	}
	if err := z.Close(); err != nil {
		return err
	}
	return renderSchemaVar(g.r, schemaVarParams{
		G:       g,
		FileID:  g.fileID,
		NodeIDs: ids,
		schema:  buf.Bytes(),
	})
}

// importForNode returns the import spec needed to reference n from
// rel's scope.  If none is needed (they are in the same package), then
// the zero importSpec is returned.
func importForNode(n, rel *node) (importSpec, error) {
	if n.pkg == "" {
		return importSpec{}, fmt.Errorf("internal error (bad schema?): missing package declaration for %s", n)
	}
	if n.imp == "" {
		return importSpec{}, fmt.Errorf("internal error (bad schema?): missing import declaration for %s", n)
	}
	if rel.imp == "" {
		return importSpec{}, fmt.Errorf("internal error (bad schema?): missing import declaration for %s", rel)
	}
	if n.imp == rel.imp {
		return importSpec{}, nil
	}
	return importSpec{path: n.imp, name: n.pkg}, nil
}

func (g *generator) RemoteNodeNew(n, rel *node) (string, error) {
	ref, err := makeNodeTypeRef(n, rel)
	if err != nil {
		return "", err
	}
	if ref.newfunc == "" {
		return "", fmt.Errorf("no new function for %s", ref.name)
	}
	if ref.imp.path == "" {
		return ref.newfunc, nil
	}
	qname := g.imports.add(ref.imp)
	return qname + "." + ref.newfunc, nil
}

func (g *generator) RemoteNodeName(n, rel *node) (string, error) {
	ref, err := makeNodeTypeRef(n, rel)
	if err != nil {
		return "", err
	}
	if ref.imp.path == "" {
		return ref.name, nil
	}
	qname := g.imports.add(ref.imp)
	return qname + "." + ref.name, nil
}

func (g *generator) RemoteTypeNew(t schema.Type, rel *node) (string, error) {
	ref, err := makeTypeRef(t, rel, g.nodes)
	if err != nil {
		return "", err
	}
	if ref.newfunc == "" {
		return "", fmt.Errorf("no new function for %s", ref.name)
	}
	if ref.imp.path == "" {
		return ref.newfunc, nil
	}
	qname := g.imports.add(ref.imp)
	return qname + "." + ref.newfunc, nil
}

func (g *generator) RemoteTypeName(t schema.Type, rel *node) (string, error) {
	ref, err := makeTypeRef(t, rel, g.nodes)
	if err != nil {
		return "", err
	}
	if ref.imp.path == "" {
		return ref.name, nil
	}
	qname := g.imports.add(ref.imp)
	return qname + "." + ref.name, nil
}

func (g *generator) defineEnum(n *node) error {
	es, _ := n.Enum().Enumerants()
	ev := make([]enumval, es.Len())
	for i := 0; i < es.Len(); i++ {
		e := es.At(i)
		ev[e.CodeOrder()] = makeEnumval(n, i, e)
	}
	nann, _ := n.Annotations()
	err := renderEnum(g.r, enumParams{
		G:           g,
		Node:        n,
		Annotations: parseAnnotations(nann),
		EnumValues:  ev,
	})
	if err != nil {
		return fmt.Errorf("enum %s: %v", n, err)
	}
	return nil
}

func isValueOfType(v schema.Value, t schema.Type) bool {
	// Ensure that the value is for the given type.  The schema ensures the union ordinals match.
	return !v.IsValid() || int(v.Which()) == int(t.Which())
}

// Value formats a value from a schema (like a field default) as Go source.
func (g *generator) Value(rel *node, t schema.Type, v schema.Value) (string, error) {
	if !isValueOfType(v, t) {
		return "", fmt.Errorf("value type is %v, but found %v value", t.Which(), v.Which())
	}

	switch t.Which() {
	case schema.Type_Which_void:
		return "struct{}{}", nil

	case schema.Type_Which_interface:
		// The only statically representable interface value is null.
		return g.imports.Capnp() + ".Client(nil)", nil

	case schema.Type_Which_bool:
		if v.Bool() {
			return "true", nil
		} else {
			return "false", nil
		}

	case schema.Type_Which_uint8, schema.Type_Which_uint16, schema.Type_Which_uint32, schema.Type_Which_uint64:
		return fmt.Sprintf("uint%d(%d)", intbits(t.Which()), uintValue(v)), nil

	case schema.Type_Which_int8, schema.Type_Which_int16, schema.Type_Which_int32, schema.Type_Which_int64:
		return fmt.Sprintf("int%d(%d)", intbits(t.Which()), intValue(v)), nil

	case schema.Type_Which_float32:
		return fmt.Sprintf("%s.Float32frombits(0x%x)", g.imports.Math(), math.Float32bits(v.Float32())), nil

	case schema.Type_Which_float64:
		return fmt.Sprintf("%s.Float64frombits(0x%x)", g.imports.Math(), math.Float64bits(v.Float64())), nil

	case schema.Type_Which_text:
		text, _ := v.Text()
		return strconv.Quote(text), nil

	case schema.Type_Which_data:
		buf := make([]byte, 0, 1024)
		buf = append(buf, "[]byte{"...)
		data, _ := v.Data()
		for i, b := range data {
			if i > 0 {
				buf = append(buf, ',', ' ')
			}
			buf = strconv.AppendUint(buf, uint64(b), 10)
		}
		buf = append(buf, '}')
		return string(buf), nil

	case schema.Type_Which_enum:
		en := g.nodes[t.Enum().TypeId()]
		if en == nil || !en.IsValid() || en.Which() != schema.Node_Which_enum {
			return "", errors.New("expected enum type")
		}
		enums, _ := en.Enum().Enumerants()
		val := int(v.Enum())
		if val >= enums.Len() {
			rn, err := g.RemoteNodeName(en, rel)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%s(%d)", rn, val), nil
		}
		ev := makeEnumval(en, val, enums.At(val))
		imp, err := importForNode(en, rel)
		if err != nil {
			return "", err
		}
		if imp.path == "" {
			return ev.FullName(), nil
		}
		qname := g.imports.add(imp)
		return qname + "." + ev.FullName(), nil

	case schema.Type_Which_structType:
		data, _ := v.StructValuePtr()
		var buf bytes.Buffer
		tn, err := g.nodes.mustFind(t.StructType().TypeId())
		if err != nil {
			return "", err
		}
		sd, err := g.data.copyData(data)
		if err != nil {
			return "", err
		}
		err = templates.ExecuteTemplate(&buf, "structValue", structValueParams{
			G:     g,
			Node:  rel,
			Typ:   tn,
			Value: sd,
		})
		return buf.String(), err

	case schema.Type_Which_anyPointer:
		data, _ := v.AnyPointerPtr()
		var buf bytes.Buffer
		sd, err := g.data.copyData(data)
		if err != nil {
			return "", err
		}
		err = templates.ExecuteTemplate(&buf, "pointerValue", pointerValueParams{
			G:     g,
			Value: sd,
		})
		return buf.String(), err

	case schema.Type_Which_list:
		data, _ := v.ListPtr()
		var buf bytes.Buffer
		ftyp, err := g.RemoteTypeName(t, rel)
		if err != nil {
			return "", err
		}
		sd, err := g.data.copyData(data)
		if err != nil {
			return "", err
		}
		err = templates.ExecuteTemplate(&buf, "listValue", listValueParams{
			G:     g,
			Typ:   ftyp,
			Value: sd,
		})
		return buf.String(), err
	default:
		return "", fmt.Errorf("unhandled value type %v", t.Which())
	}
}

func (g *generator) defineAnnotation(n *node) error {
	err := renderAnnotation(g.r, annotationParams{
		G:    g,
		Node: n,
	})
	if err != nil {
		return fmt.Errorf("annotation %s: %v", n, err)
	}
	return nil
}

func isGoConstType(t schema.Type) bool {
	w := t.Which()
	return w == schema.Type_Which_bool ||
		w == schema.Type_Which_int8 ||
		w == schema.Type_Which_uint8 ||
		w == schema.Type_Which_int16 ||
		w == schema.Type_Which_uint16 ||
		w == schema.Type_Which_int32 ||
		w == schema.Type_Which_uint32 ||
		w == schema.Type_Which_int64 ||
		w == schema.Type_Which_uint64 ||
		w == schema.Type_Which_text ||
		w == schema.Type_Which_enum
}

func (g *generator) defineConstNodes(nodes []*node) error {
	constNodes := make([]*node, 0, len(nodes))
	for _, n := range nodes {
		if n.Which() != schema.Node_Which_const {
			continue
		}
		t, _ := n.Const().Type()
		if isGoConstType(t) {
			constNodes = append(constNodes, n)
		}
	}
	nc := len(constNodes)
	for _, n := range nodes {
		if n.Which() != schema.Node_Which_const {
			continue
		}
		t, _ := n.Const().Type()
		if !isGoConstType(t) {
			constNodes = append(constNodes, n)
		}
	}
	if len(constNodes) == 0 {
		// short path
		return nil
	}
	err := renderConstants(g.r, constantsParams{
		G:      g,
		Consts: constNodes[:nc],
		Vars:   constNodes[nc:],
	})
	if err != nil {
		return fmt.Errorf("file constants: %v", err)
	}
	return nil
}

func (g *generator) defineField(n *node, f field) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("field %s.%s: %v", n.shortDisplayName(), f.Name, err)
		}
	}()

	fann, _ := f.Annotations()
	ann := parseAnnotations(fann)
	t, _ := f.Slot().Type()
	def, _ := f.Slot().DefaultValue()
	if !isValueOfType(def, t) {
		return fmt.Errorf("default value type is %v, but found %v value", t.Which(), def.Which())
	}
	ftyp, err := g.RemoteTypeName(t, n)
	if err != nil {
		return err
	}
	params := structFieldParams{
		G:           g,
		Node:        n,
		Field:       f,
		Annotations: ann,
		FieldType:   ftyp,
	}
	switch t.Which() {
	case schema.Type_Which_void:
		return renderStructVoidField(g.r, structVoidFieldParams(params))
	case schema.Type_Which_bool:
		return renderStructBoolField(g.r, structBoolFieldParams{
			structFieldParams: params,
			Default:           def.Bool(),
		})

	case schema.Type_Which_uint8, schema.Type_Which_uint16, schema.Type_Which_uint32, schema.Type_Which_uint64:
		return renderStructUintField(g.r, structUintFieldParams{
			structFieldParams: params,
			Bits:              intbits(t.Which()),
			Default:           uintValue(def),
		})

	case schema.Type_Which_int8, schema.Type_Which_int16, schema.Type_Which_int32, schema.Type_Which_int64:
		return renderStructIntField(g.r, structIntFieldParams{
			structUintFieldParams: structUintFieldParams{
				structFieldParams: params,
				Bits:              intbits(t.Which()),
				Default:           uint64(intFieldDefaultMask(def)),
			},
		})

	case schema.Type_Which_enum:
		rn, err := g.RemoteTypeName(t, n)
		if err != nil {
			return err
		}
		return renderStructIntField(g.r, structIntFieldParams{
			structUintFieldParams: structUintFieldParams{
				structFieldParams: params,
				Bits:              16,
				Default:           uint64(def.Enum()),
			},
			EnumName: rn,
		})
	case schema.Type_Which_float32:
		return renderStructFloatField(g.r, structFloatFieldParams{
			structFieldParams: params,
			Bits:              32,
			Default:           uint64(math.Float32bits(def.Float32())),
		})

	case schema.Type_Which_float64:
		return renderStructFloatField(g.r, structFloatFieldParams{
			structFieldParams: params,
			Bits:              64,
			Default:           math.Float64bits(def.Float64()),
		})

	case schema.Type_Which_text:
		d, err := def.Text()
		if err != nil {
			return err
		}
		return renderStructTextField(g.r, structTextFieldParams{
			structFieldParams: params,
			Default:           d,
		})

	case schema.Type_Which_data:
		d, err := def.Data()
		if err != nil {
			return err
		}
		return renderStructDataField(g.r, structDataFieldParams{
			structFieldParams: params,
			Default:           d,
		})

	case schema.Type_Which_structType:
		var defref staticDataRef
		if sf, err := def.StructValuePtr(); err != nil {
			return err
		} else if sf.IsValid() {
			defref, err = g.data.copyData(sf)
			if err != nil {
				return err
			}
		}
		tn, err := g.nodes.mustFind(t.StructType().TypeId())
		if err != nil {
			return err
		}
		return renderStructStructField(g.r, structStructFieldParams{
			structFieldParams: params,
			TypeNode:          tn,
			Default:           defref,
		})

	case schema.Type_Which_anyPointer:
		var defref staticDataRef
		if p, err := def.AnyPointerPtr(); err != nil {
			return err
		} else if p.IsValid() {
			defref, err = g.data.copyData(p)
			if err != nil {
				return err
			}
		}
		return renderStructPointerField(g.r, structPointerFieldParams{
			structFieldParams: params,
			Default:           defref,
		})

	case schema.Type_Which_list:
		var defref staticDataRef
		if l, err := def.ListPtr(); err != nil {
			return err
		} else if l.IsValid() {
			defref, err = g.data.copyData(l)
			if err != nil {
				return err
			}
		}
		return renderStructListField(g.r, structListFieldParams{
			structFieldParams: params,
			Default:           defref,
		})

	case schema.Type_Which_interface:
		return renderStructInterfaceField(g.r, structInterfaceFieldParams(params))
	default:
		return fmt.Errorf("defining unhandled field type %v", t.Which())
	}
}

// typeRef is a Go reference to a Cap'n Proto type.
type typeRef struct {
	name    string
	newfunc string     // if absent, there is no New function for this type.
	imp     importSpec // optional
}

func makeNodeTypeRef(n, rel *node) (typeRef, error) {
	imp, err := importForNode(n, rel)
	if err != nil {
		return typeRef{}, err
	}
	switch n.Which() {
	case schema.Node_Which_structNode:
		return typeRef{
			name:    n.Name,
			newfunc: "New" + n.Name,
			imp:     imp,
		}, nil
	case schema.Node_Which_enum, schema.Node_Which_interface:
		return typeRef{
			name: n.Name,
			imp:  imp,
		}, nil
	}
	return typeRef{}, fmt.Errorf("unable to reference type of node %v", n.Which())
}

var (
	staticTypeRefs = map[schema.Type_Which]typeRef{
		schema.Type_Which_void:       {},
		schema.Type_Which_bool:       {name: "bool"},
		schema.Type_Which_int8:       {name: "int8"},
		schema.Type_Which_int16:      {name: "int16"},
		schema.Type_Which_int32:      {name: "int32"},
		schema.Type_Which_int64:      {name: "int64"},
		schema.Type_Which_uint8:      {name: "uint8"},
		schema.Type_Which_uint16:     {name: "uint16"},
		schema.Type_Which_uint32:     {name: "uint32"},
		schema.Type_Which_uint64:     {name: "uint64"},
		schema.Type_Which_float32:    {name: "float32"},
		schema.Type_Which_float64:    {name: "float64"},
		schema.Type_Which_text:       {name: "string"},
		schema.Type_Which_data:       {name: "[]byte"},
		schema.Type_Which_anyPointer: {name: "Pointer", imp: capnpImportSpec},
	}
	staticListTypeRefs = map[schema.Type_Which]typeRef{
		// TODO(light): omitting newfunc since it doesn't have a similar type signature (no errors).
		schema.Type_Which_void: typeRef{name: "VoidList", imp: capnpImportSpec},

		schema.Type_Which_bool:    typeRef{name: "BitList", newfunc: "NewBitList", imp: capnpImportSpec},
		schema.Type_Which_int8:    typeRef{name: "Int8List", newfunc: "NewInt8List", imp: capnpImportSpec},
		schema.Type_Which_uint8:   typeRef{name: "UInt8List", newfunc: "NewUInt8List", imp: capnpImportSpec},
		schema.Type_Which_int16:   typeRef{name: "Int16List", newfunc: "NewInt16List", imp: capnpImportSpec},
		schema.Type_Which_uint16:  typeRef{name: "UInt16List", newfunc: "NewUInt16List", imp: capnpImportSpec},
		schema.Type_Which_int32:   typeRef{name: "Int32List", newfunc: "NewInt32List", imp: capnpImportSpec},
		schema.Type_Which_uint32:  typeRef{name: "UInt32List", newfunc: "NewUInt32List", imp: capnpImportSpec},
		schema.Type_Which_int64:   typeRef{name: "Int64List", newfunc: "NewInt64List", imp: capnpImportSpec},
		schema.Type_Which_uint64:  typeRef{name: "UInt64List", newfunc: "NewUInt64List", imp: capnpImportSpec},
		schema.Type_Which_float32: typeRef{name: "Float32List", newfunc: "NewFloat32List", imp: capnpImportSpec},
		schema.Type_Which_float64: typeRef{name: "Float64List", newfunc: "NewFloat64List", imp: capnpImportSpec},
		schema.Type_Which_text:    typeRef{name: "TextList", newfunc: "NewTextList", imp: capnpImportSpec},
		schema.Type_Which_data:    typeRef{name: "DataList", newfunc: "NewDataList", imp: capnpImportSpec},
	}
)

func makeTypeRef(t schema.Type, rel *node, nodes nodeMap) (typeRef, error) {
	nodeRef := func(id uint64) (typeRef, error) {
		ni, err := nodes.mustFind(id)
		if err != nil {
			return typeRef{}, err
		}
		return makeNodeTypeRef(ni, rel)
	}
	if ref, ok := staticTypeRefs[t.Which()]; ok {
		return ref, nil
	}
	switch t.Which() {
	case schema.Type_Which_enum:
		return nodeRef(t.Enum().TypeId())
	case schema.Type_Which_structType:
		return nodeRef(t.StructType().TypeId())
	case schema.Type_Which_interface:
		return nodeRef(t.Interface().TypeId())
	case schema.Type_Which_list:
		lt, _ := t.List().ElementType()
		if ref, ok := staticListTypeRefs[lt.Which()]; ok {
			return ref, nil
		}
		switch lt.Which() {
		case schema.Type_Which_enum:
			ref, err := nodeRef(lt.Enum().TypeId())
			if err != nil {
				return ref, err
			}
			ref.name = ref.name + "_List"
			ref.newfunc = "New" + ref.name
			return ref, nil
		case schema.Type_Which_structType:
			ref, err := nodeRef(lt.StructType().TypeId())
			if err != nil {
				return ref, err
			}
			ref.name = ref.name + "_List"
			ref.newfunc = "New" + ref.name
			return ref, nil
		case schema.Type_Which_anyPointer, schema.Type_Which_list, schema.Type_Which_interface:
			return typeRef{name: "PointerList", newfunc: "NewPointerList", imp: capnpImportSpec}, nil
		}
	}
	return typeRef{}, fmt.Errorf("unable to reference type %v", t.Which())
}

// intFieldDefaultMask returns the XOR mask used when getting or setting
// signed integer struct fields.
func intFieldDefaultMask(v schema.Value) uint64 {
	mask := uint64(1)<<intbits(schema.Type_Which(v.Which())) - 1
	return uint64(intValue(v)) & mask
}

// intValue returns the signed integer value of a schema value or zero
// if the value is invalid (the null pointer). Panics if the value is
// not a signed integer.
func intValue(v schema.Value) int64 {
	if !v.IsValid() {
		return 0
	}
	switch v.Which() {
	case schema.Value_Which_int8:
		return int64(v.Int8())
	case schema.Value_Which_int16:
		return int64(v.Int16())
	case schema.Value_Which_int32:
		return int64(v.Int32())
	case schema.Value_Which_int64:
		return v.Int64()
	}
	panic("unreachable")
}

// uintValue returns the unsigned integer value of a schema value or
// zero if the value is invalid (the null pointer). Panics if the value
// is not an unsigned integer.
func uintValue(v schema.Value) uint64 {
	if !v.IsValid() {
		return 0
	}
	switch v.Which() {
	case schema.Value_Which_uint8:
		return uint64(v.Uint8())
	case schema.Value_Which_uint16:
		return uint64(v.Uint16())
	case schema.Value_Which_uint32:
		return uint64(v.Uint32())
	case schema.Value_Which_uint64:
		return v.Uint64()
	}
	panic("unreachable")
}

// intbits returns the number of bits that an integer type requires.
func intbits(t schema.Type_Which) uint {
	switch t {
	case schema.Type_Which_uint8, schema.Type_Which_int8:
		return 8
	case schema.Type_Which_uint16, schema.Type_Which_int16:
		return 16
	case schema.Type_Which_uint32, schema.Type_Which_int32:
		return 32
	case schema.Type_Which_uint64, schema.Type_Which_int64:
		return 64
	default:
		panic("unreachable")
	}
}

func (g *generator) defineStruct(n *node) error {
	if err := g.defineStructTypes(n, n); err != nil {
		return err
	}
	if err := g.defineStructEnums(n); err != nil {
		return err
	}
	if err := g.defineBaseStructFuncs(n); err != nil {
		return err
	}
	if err := g.defineStructFuncs(n); err != nil {
		return err
	}
	if err := g.defineStructList(n); err != nil {
		return err
	}
	if g.opts.promises {
		if err := g.defineStructPromise(n); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) defineStructTypes(n, baseNode *node) error {
	nann, _ := n.Annotations()
	ann := parseAnnotations(nann)
	err := renderStructTypes(g.r, structTypesParams{
		G:           g,
		Node:        n,
		Annotations: ann,
		BaseNode:    baseNode,
	})
	if err != nil {
		dn, _ := n.DisplayName()
		return fmt.Errorf("struct type for %s: %v", dn, err)
	}

	for _, f := range n.codeOrderFields() {
		if f.Which() == schema.Field_Which_group {
			grp, err := g.nodes.mustFind(f.Group().TypeId())
			if err != nil {
				return err
			}
			if err := g.defineStructTypes(grp, baseNode); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *generator) defineStructEnums(n *node) error {
	fields := n.codeOrderFields()
	members := make([]field, 0, len(fields))
	es := make(enumString, 0, len(fields))
	for _, f := range fields {
		if f.DiscriminantValue() != schema.Field_noDiscriminant {
			members = append(members, f)
			es = append(es, f.Name)
		}
	}
	if n.StructNode().DiscriminantCount() > 0 {
		err := renderStructEnums(g.r, structEnumsParams{
			G:          g,
			Node:       n,
			Fields:     members,
			EnumString: es,
		})
		if err != nil {
			return fmt.Errorf("struct enums for %s: %v", n, err)
		}
	}
	for _, f := range fields {
		if f.Which() == schema.Field_Which_group {
			grp, err := g.nodes.mustFind(f.Group().TypeId())
			if err != nil {
				return err
			}
			if err := g.defineStructEnums(grp); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *generator) defineStructFuncs(n *node) error {
	err := renderStructFuncs(g.r, structFuncsParams{
		G:    g,
		Node: n,
	})
	if err != nil {
		return fmt.Errorf("struct funcs for %s: %v", n, err)
	}

	for _, f := range n.codeOrderFields() {
		switch f.Which() {
		case schema.Field_Which_slot:
			if err := g.defineField(n, f); err != nil {
				return err
			}
		case schema.Field_Which_group:
			grp, err := g.nodes.mustFind(f.Group().TypeId())
			if err != nil {
				return err
			}
			err = renderStructGroup(g.r, structGroupParams{
				G:     g,
				Node:  n,
				Group: grp,
				Field: f,
			})
			if err != nil {
				return fmt.Errorf("struct group for %s: %v", grp, err)
			}
			if err := g.defineStructFuncs(grp); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *generator) ObjectSize(n *node) (string, error) {
	if n.Which() != schema.Node_Which_structNode {
		return "", fmt.Errorf("object size called for %v node", n.Which())
	}
	return fmt.Sprintf("%s.ObjectSize{DataSize: %d, PointerCount: %d}",
		g.imports.Capnp(),
		int(n.StructNode().DataWordCount())*8,
		n.StructNode().PointerCount()), nil
}

func (g *generator) defineBaseStructFuncs(n *node) error {
	err := renderBaseStructFuncs(g.r, baseStructFuncsParams{
		G:            g,
		Node:         n,
		StringMethod: g.opts.structStrings,
	})
	if err != nil {
		return fmt.Errorf("base struct functions for %s: %v", n, err)
	}
	return nil
}

func (g *generator) defineStructList(n *node) error {
	err := renderStructList(g.r, structListParams{
		G:    g,
		Node: n,
	})
	if err != nil {
		return fmt.Errorf("new struct function for %s: %v", n, err)
	}
	return nil
}

func (g *generator) defineStructPromise(n *node) error {
	err := renderPromise(g.r, promiseParams{
		G:      g,
		Node:   n,
		Fields: n.codeOrderFields(),
	})
	if err != nil {
		return fmt.Errorf("promise for struct %s: %v", n, err)
	}

	for _, f := range n.codeOrderFields() {
		switch f.Which() {
		case schema.Field_Which_slot:
			t, _ := f.Slot().Type()
			if tw := t.Which(); tw != schema.Type_Which_structType && tw != schema.Type_Which_interface && tw != schema.Type_Which_anyPointer {
				continue
			}
			if err := g.definePromiseField(n, f); err != nil {
				return fmt.Errorf("promise field %s.%s: %v", n.shortDisplayName(), f.Name, err)
			}
		case schema.Field_Which_group:
			grp, err := g.nodes.mustFind(f.Group().TypeId())
			if err != nil {
				return fmt.Errorf("promise group %s.%s: %v", n.shortDisplayName(), f.Name, err)
			}
			err = renderPromiseGroup(g.r, promiseGroupParams{
				G:     g,
				Node:  n,
				Field: f,
				Group: grp,
			})
			if err != nil {
				return fmt.Errorf("promise for group %s: %v", grp, err)
			}
			if err := g.defineStructPromise(grp); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *generator) definePromiseField(n *node, f field) error {
	slot := f.Slot()
	switch t, _ := slot.Type(); t.Which() {
	case schema.Type_Which_structType:
		ni, err := g.nodes.mustFind(t.StructType().TypeId())
		if err != nil {
			return err
		}
		params := promiseFieldStructParams{
			G:      g,
			Node:   n,
			Field:  f,
			Struct: ni,
		}
		if def, _ := slot.DefaultValue(); def.IsValid() && def.Which() == schema.Value_Which_structValue {
			if sf, _ := def.StructValuePtr(); sf.IsValid() {
				params.Default, err = g.data.copyData(sf)
				if err != nil {
					return err
				}
			}
		}
		return renderPromiseFieldStruct(g.r, params)
	case schema.Type_Which_anyPointer:
		return renderPromiseFieldAnyPointer(g.r, promiseFieldAnyPointerParams{
			G:     g,
			Node:  n,
			Field: f,
		})
	case schema.Type_Which_interface:
		ni, err := g.nodes.mustFind(t.Interface().TypeId())
		if err != nil {
			return err
		}
		return renderPromiseFieldInterface(g.r, promiseFieldInterfaceParams{
			G:         g,
			Node:      n,
			Field:     f,
			Interface: ni,
		})
	default:
		panic("unreachable")
	}
}

func (g *generator) defineInterface(n *node) error {
	m, err := methodSet(nil, n, g.nodes)
	if err != nil {
		return fmt.Errorf("building method set of interface %s: %v", n, err)
	}
	nann, _ := n.Annotations()
	err = renderInterfaceClient(g.r, interfaceClientParams{
		G:           g,
		Node:        n,
		Annotations: parseAnnotations(nann),
		Methods:     m,
	})
	if err != nil {
		return fmt.Errorf("interface client %s: %v", n, err)
	}
	err = renderInterfaceServer(g.r, interfaceServerParams{
		G:           g,
		Node:        n,
		Annotations: parseAnnotations(nann),
		Methods:     m,
	})
	if err != nil {
		return fmt.Errorf("interface server %s: %v", n, err)
	}
	return nil
}

type enumString []string

func (es enumString) ValueString() string {
	return strings.Join([]string(es), "")
}

func (es enumString) SliceFor(i int) string {
	n := 0
	for _, v := range es[:i] {
		n += len(v)
	}
	return fmt.Sprintf("[%d:%d]", n, n+len(es[i]))
}

func (g *generator) defineFile() error {
	f := g.nodes[g.fileID]
	if f == nil {
		return fmt.Errorf("no node in schema matches %#x", g.fileID)
	}
	if f.pkg == "" {
		return errors.New("missing package annotation")
	}

	for _, n := range f.nodes {
		if n.Which() == schema.Node_Which_annotation {
			if err := g.defineAnnotation(n); err != nil {
				return err
			}
		}
	}
	if err := g.defineConstNodes(f.nodes); err != nil {
		return err
	}
	for _, n := range f.nodes {
		var err error
		switch n.Which() {
		case schema.Node_Which_enum:
			err = g.defineEnum(n)
		case schema.Node_Which_structNode:
			if !n.StructNode().IsGroup() {
				err = g.defineStruct(n)
			}
		case schema.Node_Which_interface:
			err = g.defineInterface(n)
		}
		if err != nil {
			return err
		}
	}
	if g.opts.schemas {
		if err := g.defineSchemaVar(); err != nil {
			return err
		}
	}
	return nil
}

func generateFile(reqf schema.CodeGeneratorRequest_RequestedFile, nodes nodeMap, opts genoptions) error {
	if opts.structStrings && !opts.schemas {
		return errors.New("cannot generate struct String() methods without embedding schemas")
	}
	id := reqf.Id()
	fname, _ := reqf.Filename()
	g := newGenerator(id, nodes, opts)
	if err := g.defineFile(); err != nil {
		return err
	}

	if dirPath, _ := filepath.Split(fname); dirPath != "" {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	unformatted := g.generate()
	formatted, fmtErr := format.Source(unformatted)
	if fmtErr != nil {
		formatted = unformatted
	}

	file, err := os.Create(fname + ".go")
	if err != nil {
		return err
	}
	_, werr := file.Write(formatted)
	cerr := file.Close()
	if fmtErr != nil {
		return fmtErr
	}
	if werr != nil {
		return err
	}
	if cerr != nil {
		return err
	}
	return nil
}

func main() {
	var opts genoptions
	flag.BoolVar(&opts.promises, "promises", true, "generate code for promises")
	flag.BoolVar(&opts.schemas, "schemas", true, "embed schema information in generated code")
	flag.BoolVar(&opts.structStrings, "structstrings", true, "generate String() methods for structs (-schemas must be true)")
	flag.Parse()

	msg, err := capnp.NewDecoder(os.Stdin).Decode()
	if err != nil {
		fmt.Fprintln(os.Stderr, "capnpc-go: reading input:", err)
		os.Exit(1)
	}
	req, err := schema.ReadRootCodeGeneratorRequest(msg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "capnpc-go: reading input:", err)
		os.Exit(1)
	}
	nodes, err := buildNodeMap(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "capnpc-go:", err)
		os.Exit(1)
	}
	success := true
	reqFiles, _ := req.RequestedFiles()
	for i := 0; i < reqFiles.Len(); i++ {
		reqf := reqFiles.At(i)
		err := generateFile(reqf, nodes, opts)
		if err != nil {
			fname, _ := reqf.Filename()
			fmt.Fprintf(os.Stderr, "capnpc-go: generating %s: %v\n", fname, err)
			success = false
		}
	}
	if !success {
		os.Exit(1)
	}
}

type uint64Slice []uint64

func (p uint64Slice) Len() int           { return len(p) }
func (p uint64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p uint64Slice) Less(i, j int) bool { return p[i] < p[j] }
