//go:generate ../internal/cmd/mktemplates/mktemplates templates.go templates

package main

import (
	"strings"
	"text/template"
)

var templates = template.Must(template.New("").Funcs(template.FuncMap{
	"title": strings.Title,
}).Parse(
	"{{define \"_hasfield\"}}func (s {{.Node.Name}}) Has{{.Field.Name | title}}() bool {\n\tp, err := s.Struct.Ptr({{.Field.Slot.Offset}})\n\treturn p.IsValid() || err != nil \n}\n{{end}}{{define \"_interfaceMethod\"}}\t\t\tInterfaceID: {{.Interface.Id | printf \"%#x\"}},\n\t\t\tMethodID: {{.ID}},\n\t\t\tInterfaceName: {{.Interface.DisplayName | printf \"%q\"}},\n\t\t\tMethodName: {{.OriginalName | printf \"%q\"}},\n{{end}}{{define \"_settag\"}}{{if .Field.HasDiscriminant}}s.Struct.SetUint16({{.Node.DiscriminantOffset}}, {{.Field.DiscriminantValue}})\n{{end}}{{end}}{{define \"annotation\"}}const {{.Node.Name}} = uint64({{.Node.Id | printf \"%#x\"}})\n{{end}}{{define \"baseStructFuncs\"}}// {{.Node.Name}}_TypeID is the unique identifier for the type {{.Node.Name}}.\nconst {{.Node.Name}}_TypeID = {{.Node.Id | printf \"%#x\"}}\n\nfunc New{{.Node.Name}}(s *{{.G.Capnp}}.Segment) ({{.Node.Name}}, error) {\n\tst, err := {{$.G.Capnp}}.NewStruct(s, {{.G.ObjectSize .Node}})\n\treturn {{.Node.Name}}{st}, err\n}\n\nfunc NewRoot{{.Node.Name}}(s *{{.G.Capnp}}.Segment) ({{.Node.Name}}, error) {\n\tst, err := {{.G.Capnp}}.NewRootStruct(s, {{.G.ObjectSize .Node}})\n\treturn {{.Node.Name}}{st}, err\n}\n\nfunc ReadRoot{{.Node.Name}}(msg *{{.G.Capnp}}.Message) ({{.Node.Name}}, error) {\n\troot, err := msg.RootPtr()\n\treturn {{.Node.Name}}{root.Struct()}, err\n}\n{{if .StringMethod}}\nfunc (s {{.Node.Name}}) String() string {\n\tstr, _ := {{.G.Imports.Text}}.Marshal({{.Node.Id | printf \"%#x\"}}, s.Struct)\n\treturn str\n}\n{{end}}\n\n{{end}}{{define \"constants\"}}{{with .Consts}}// Constants defined in {{$.G.Basename}}.\nconst (\n{{range .}}\t{{.Name}} = {{$.G.Value . .Const.Type .Const.Value}}\n{{end}}\n)\n{{end}}\n{{with .Vars}}// Constants defined in {{$.G.Basename}}.\nvar (\n{{range .}}\t{{.Name}} = {{$.G.Value . .Const.Type .Const.Value}}\n{{end}}\n)\n{{end}}\n{{end}}{{define \"enum\"}}{{with .Annotations.Doc}}// {{.}}\n{{end}}type {{.Node.Name}} uint16\n\n{{with .EnumValues}}// Values of {{$.Node.Name}}.\nconst (\n{{range .}}{{.FullName}} {{$.Node.Name}} = {{.Val}}\n{{end}}\n)\n\n// String returns the enum's constant name.\nfunc (c {{$.Node.Name}}) String() string {\n\tswitch c {\n\t{{range .}}{{if .Tag}}case {{.FullName}}: return {{printf \"%q\" .Tag}}\n\t{{end}}{{end}}\n\tdefault: return \"\"\n\t}\n}\n\n// {{$.Node.Name}}FromString returns the enum value with a name,\n// or the zero value if there's no such value.\nfunc {{$.Node.Name}}FromString(c string) {{$.Node.Name}} {\n\tswitch c {\n\t{{range .}}{{if .Tag}}case {{printf \"%q\" .Tag}}: return {{.FullName}}\n\t{{end}}{{end}}\n\tdefault: return 0\n\t}\n}\n{{end}}\n\ntype {{.Node.Name}}_List struct { {{$.G.Capnp}}.List }\n\nfunc New{{.Node.Name}}_List(s *{{$.G.Capnp}}.Segment, sz int32) ({{.Node.Name}}_List, error) {\n\tl, err := {{.G.Capnp}}.NewUInt16List(s, sz)\n\treturn {{.Node.Name}}_List{l.List}, err\n}\n\nfunc (l {{.Node.Name}}_List) At(i int) {{.Node.Name}} {\n\tul := {{.G.Capnp}}.UInt16List{List: l.List}\n\treturn {{.Node.Name}}(ul.At(i))\n}\n\nfunc (l {{.Node.Name}}_List) Set(i int, v {{.Node.Name}}) {\n\tul := {{.G.Capnp}}.UInt16List{List: l.List}\n\tul.Set(i, uint16(v))\n}\n{{end}}{{define \"interfaceClient\"}}{{with .Annotations.Doc}}// {{.}}\n{{end}}type {{.Node.Name}} struct { Client {{.G.Capnp}}.Client }\n\n{{range .Methods}}func (c {{$.Node.Name}}) {{.Name | title}}(ctx {{$.G.Imports.Context}}.Context, params func({{$.G.RemoteNodeName .Params $.Node}}) error, opts ...{{$.G.Capnp}}.CallOption) {{$.G.RemoteNodeName .Results $.Node}}_Promise {\n\tif c.Client == nil {\n\t\treturn {{$.G.RemoteNodeName .Results $.Node}}_Promise{Pipeline: {{$.G.Capnp}}.NewPipeline({{$.G.Capnp}}.ErrorAnswer({{$.G.Capnp}}.ErrNullClient))}\n\t}\n\tcall := &{{$.G.Capnp}}.Call{\n\t\tCtx: ctx,\n\t\tMethod: {{$.G.Capnp}}.Method{\n\t\t\t{{template \"_interfaceMethod\" .}}\n\t\t},\n\t\tOptions: {{$.G.Capnp}}.NewCallOptions(opts),\n\t}\n\tif params != nil {\n\t\tcall.ParamsSize = {{$.G.ObjectSize .Params}}\n\t\tcall.ParamsFunc = func(s {{$.G.Capnp}}.Struct) error { return params({{$.G.RemoteNodeName .Params $.Node}}{Struct: s}) }\n\t}\n\treturn {{$.G.RemoteNodeName .Results $.Node}}_Promise{Pipeline: {{$.G.Capnp}}.NewPipeline(c.Client.Call(call))}\n}\n{{end}}\n{{end}}{{define \"interfaceServer\"}}type {{.Node.Name}}_Server interface {\n\t{{range .Methods}}\n\t{{.Name | title}}({{$.G.RemoteNodeName .Interface $.Node}}_{{.Name}}) error\n\t{{end}}\n}\n\nfunc {{.Node.Name}}_ServerToClient(s {{.Node.Name}}_Server) {{.Node.Name}} {\n\tc, _ := s.({{.G.Imports.Server}}.Closer)\n\treturn {{.Node.Name}}{Client: {{.G.Imports.Server}}.New({{.Node.Name}}_Methods(nil, s), c)}\n}\n\nfunc {{.Node.Name}}_Methods(methods []{{.G.Imports.Server}}.Method, s {{.Node.Name}}_Server) []{{.G.Imports.Server}}.Method {\n\tif cap(methods) == 0 {\n\t\tmethods = make([]{{.G.Imports.Server}}.Method, 0, {{len .Methods}})\n\t}\n\t{{range .Methods}}\n\tmethods = append(methods, {{$.G.Imports.Server}}.Method{\n\t\tMethod: {{$.G.Capnp}}.Method{\n\t\t\t{{template \"_interfaceMethod\" .}}\n\t\t},\n\t\tImpl: func(c {{$.G.Imports.Context}}.Context, opts {{$.G.Capnp}}.CallOptions, p, r {{$.G.Capnp}}.Struct) error {\n\t\t\tcall := {{$.G.RemoteNodeName .Interface $.Node}}_{{.Name}}{c, opts, {{$.G.RemoteNodeName .Params $.Node}}{Struct: p}, {{$.G.RemoteNodeName .Results $.Node}}{Struct: r} }\n\t\t\treturn s.{{.Name | title}}(call)\n\t\t},\n\t\tResultsSize: {{$.G.ObjectSize .Results}},\n\t})\n\t{{end}}\n\treturn methods\n}\n{{range .Methods}}{{if eq .Interface.Id $.Node.Id}}\n// {{$.Node.Name}}_{{.Name}} holds the arguments for a server call to {{$.Node.Name}}.{{.Name}}.\ntype {{$.Node.Name}}_{{.Name}} struct {\n\tCtx     {{$.G.Imports.Context}}.Context\n\tOptions {{$.G.Capnp}}.CallOptions\n\tParams  {{$.G.RemoteNodeName .Params $.Node}}\n\tResults {{$.G.RemoteNodeName .Results $.Node}}\n}\n{{end}}{{end}}\n{{end}}{{define \"listValue\"}}{{.Typ}}{List: {{.G.Capnp}}.MustUnmarshalRootPtr({{.Value}}).List()}{{end}}{{define \"pointerValue\"}}{{.G.Capnp}}.MustUnmarshalRootPtr({{.Value}}){{end}}{{define \"promise\"}}// {{.Node.Name}}_Promise is a wrapper for a {{.Node.Name}} promised by a client call.\ntype {{.Node.Name}}_Promise struct { *{{.G.Capnp}}.Pipeline }\n\nfunc (p {{.Node.Name}}_Promise) Struct() ({{.Node.Name}}, error) {\n\ts, err := p.Pipeline.Struct()\n\treturn {{.Node.Name}}{s}, err\n}\n\n{{end}}{{define \"promiseFieldAnyPointer\"}}func (p {{.Node.Name}}_Promise) {{.Field.Name | title}}() *{{.G.Capnp}}.Pipeline {\n\treturn p.Pipeline.GetPipeline({{.Field.Slot.Offset}})\n}\n\n{{end}}{{define \"promiseFieldInterface\"}}func (p {{.Node.Name}}_Promise) {{.Field.Name | title}}() {{.G.RemoteNodeName .Interface .Node}} {\n\treturn {{.G.RemoteNodeName .Interface .Node}}{Client: p.Pipeline.GetPipeline({{.Field.Slot.Offset}}).Client()}\n}\n\n{{end}}{{define \"promiseFieldStruct\"}}func (p {{.Node.Name}}_Promise) {{.Field.Name | title}}() {{.G.RemoteNodeName .Struct .Node}}_Promise {\n\treturn {{.G.RemoteNodeName .Struct .Node}}_Promise{Pipeline: p.Pipeline.{{if .Default.IsValid}}GetPipelineDefault({{.Field.Slot.Offset}}, {{.Default}}){{else}}GetPipeline({{.Field.Slot.Offset}}){{end}} }\n}\n\n{{end}}{{define \"promiseGroup\"}}func (p {{.Node.Name}}_Promise) {{.Field.Name | title}}() {{.Group.Name}}_Promise { return {{.Group.Name}}_Promise{p.Pipeline} }\n{{end}}{{define \"schemaVar\"}}const schema_{{.FileID | printf \"%x\"}} = {{.SchemaLiteral}}\n\nfunc init() {\n  {{.G.Imports.Schemas}}.Register(schema_{{.FileID | printf \"%x\"}},{{range .NodeIDs}}\n\t{{. | printf \"%#x\"}},{{end}})\n}\n{{end}}{{define \"structBoolField\"}}func (s {{.Node.Name}}) {{.Field.Name | title}}() bool {\n\treturn {{if .Default}}!{{end}}s.Struct.Bit({{.Field.Slot.Offset}})\n}\n\nfunc (s {{.Node.Name}}) Set{{.Field.Name | title}}(v bool) {\n\t{{template \"_settag\" .}}s.Struct.SetBit({{.Field.Slot.Offset}}, {{if .Default}}!{{end}}v)\n}\n\n{{end}}{{define \"structDataField\"}}func (s {{.Node.Name}}) {{.Field.Name | title}}() ({{.FieldType}}, error) {\n\tp, err := s.Struct.Ptr({{.Field.Slot.Offset}})\n\t{{with .Default}}return {{$.FieldType}}(p.DataDefault({{printf \"%#v\" .}})), err{{else}}return {{.FieldType}}(p.Data()), err{{end}}\n}\n\n{{template \"_hasfield\" .}}\n\nfunc (s {{.Node.Name}}) Set{{.Field.Name | title}}(v {{.FieldType}}) error {\n\t{{template \"_settag\" .}}d, err := {{.G.Capnp}}.NewData(s.Struct.Segment(), []byte(v))\n\tif err != nil {\n\t\treturn err\n\t}\n\treturn s.Struct.SetPtr({{.Field.Slot.Offset}}, d.List.ToPtr())\n}\n\n{{end}}{{define \"structEnums\"}}type {{.Node.Name}}_Which uint16\n\nconst (\n{{range .Fields}}\t{{$.Node.Name}}_Which_{{.Name}} {{$.Node.Name}}_Which = {{.DiscriminantValue}}\n{{end}}\n)\n\nfunc (w {{.Node.Name}}_Which) String() string {\n\tconst s = {{.EnumString.ValueString | printf \"%q\"}}\n\tswitch w {\n\t{{range $i, $f := .Fields}}case {{$.Node.Name}}_Which_{{.Name}}:\n\t\treturn s{{$.EnumString.SliceFor $i}}\n\t{{end}}\n\t}\n\treturn \"{{.Node.Name}}_Which(\" + {{.G.Imports.Strconv}}.FormatUint(uint64(w), 10) + \")\"\n}\n\n{{end}}{{define \"structFloatField\"}}func (s {{.Node.Name}}) {{.Field.Name | title}}() float{{.Bits}} {\n\treturn {{.G.Imports.Math}}.Float{{.Bits}}frombits(s.Struct.Uint{{.Bits}}({{.Offset}}){{with .Default}} ^ {{printf \"%#x\" .}}{{end}})\n}\n\nfunc (s {{.Node.Name}}) Set{{.Field.Name | title}}(v float{{.Bits}}) {\n\t{{template \"_settag\" .}}s.Struct.SetUint{{.Bits}}({{.Offset}}, {{.G.Imports.Math}}.Float{{.Bits}}bits(v){{with .Default}}^{{printf \"%#x\" .}}{{end}})\n}\n\n{{end}}{{define \"structFuncs\"}}{{if gt .Node.StructNode.DiscriminantCount 0}}\nfunc (s {{.Node.Name}}) Which() {{.Node.Name}}_Which {\n\treturn {{.Node.Name}}_Which(s.Struct.Uint16({{.Node.DiscriminantOffset}}))\n}\n{{end}}{{end}}{{define \"structGroup\"}}func (s {{.Node.Name}}) {{.Field.Name | title}}() {{.Group.Name}} { return {{.Group.Name}}(s) }\n{{if .Field.HasDiscriminant}}\nfunc (s {{.Node.Name}}) Set{{.Field.Name | title}}() { {{template \"_settag\" .}} }\n{{end}}\n{{end}}{{define \"structIntField\"}}func (s {{.Node.Name}}) {{.Field.Name | title}}() {{.ReturnType}} {\n\treturn {{.ReturnType}}(s.Struct.Uint{{.Bits}}({{.Offset}}){{with .Default}} ^ {{.}}{{end}})\n}\n\nfunc (s {{.Node.Name}}) Set{{.Field.Name | title}}(v {{.ReturnType}}) {\n\t{{template \"_settag\" .}}s.Struct.SetUint{{.Bits}}({{.Offset}}, uint{{.Bits}}(v){{with .Default}}^{{.}}{{end}})\n}\n\n{{end}}{{define \"structInterfaceField\"}}func (s {{.Node.Name}}) {{.Field.Name | title}}() {{.FieldType}} {\n\tp, _ := s.Struct.Ptr({{.Field.Slot.Offset}})\n\treturn {{.FieldType}}{Client: p.Interface().Client()}\n}\n\n{{template \"_hasfield\" .}}\n\nfunc (s {{.Node.Name}}) Set{{.Field.Name | title}}(v {{.FieldType}}) error {\n\t{{template \"_settag\" .}}if v.Client == nil {\n\t\treturn s.Struct.SetPtr({{.Field.Slot.Offset}}, capnp.Ptr{})\n\t}\n\tseg := s.Segment()\n\tin := {{.G.Capnp}}.NewInterface(seg, seg.Message().AddCap(v.Client))\n\treturn s.Struct.SetPtr({{.Field.Slot.Offset}}, in.ToPtr())\n}\n\n{{end}}{{define \"structList\"}}// {{.Node.Name}}_List is a list of {{.Node.Name}}.\ntype {{.Node.Name}}_List struct{ {{.G.Capnp}}.List }\n\n// New{{.Node.Name}} creates a new list of {{.Node.Name}}.\nfunc New{{.Node.Name}}_List(s *{{.G.Capnp}}.Segment, sz int32) ({{.Node.Name}}_List, error) {\n\tl, err := {{.G.Capnp}}.NewCompositeList(s, {{.G.ObjectSize .Node}}, sz)\n\treturn {{.Node.Name}}_List{l}, err\n}\n\nfunc (s {{.Node.Name}}_List) At(i int) {{.Node.Name}} { return {{.Node.Name}}{ s.List.Struct(i) } }\n\nfunc (s {{.Node.Name}}_List) Set(i int, v {{.Node.Name}}) error { return s.List.SetStruct(i, v.Struct) }\n{{end}}{{define \"structListField\"}}func (s {{.Node.Name}}) {{.Field.Name | title}}() ({{.FieldType}}, error) {\n\tp, err := s.Struct.Ptr({{.Field.Slot.Offset}})\n\t{{if .Default.IsValid}}if err != nil {\n\t\treturn {{.FieldType}}{}, err\n\t}\n\tl, err := p.ListDefault({{.Default}})\n\treturn {{.FieldType}}{List: l}, err{{else}}return {{.FieldType}}{List: p.List()}, err{{end}}\n}\n\n{{template \"_hasfield\" .}}\n\nfunc (s {{.Node.Name}}) Set{{.Field.Name | title}}(v {{.FieldType}}) error {\n\t{{template \"_settag\" .}}return s.Struct.SetPtr({{.Field.Slot.Offset}}, v.List.ToPtr())\n}\n\n// New{{.Field.Name | title}} sets the {{.Field.Name}} field to a newly\n// allocated {{.FieldType}}, preferring placement in s's segment.\nfunc (s {{.Node.Name}}) New{{.Field.Name | title}}(n int32) ({{.FieldType}}, error) {\n\t{{template \"_settag\" .}}l, err := {{.G.RemoteTypeNew .Field.Slot.Type .Node}}(s.Struct.Segment(), n)\n\tif err != nil {\n\t\treturn {{.FieldType}}{}, err\n\t}\n\terr = s.Struct.SetPtr({{.Field.Slot.Offset}}, l.List.ToPtr())\n\treturn l, err\n}\n\n{{end}}{{define \"structPointerField\"}}func (s {{.Node.Name}}) {{.Field.Name | title}}() ({{.G.Capnp}}.Pointer, error) {\n\t{{if .Default.IsValid}}p, err := s.Struct.Pointer({{.Field.Slot.Offset}})\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\treturn {{.G.Capnp}}.PointerDefault(p, {{.Default}}){{else}}return s.Struct.Pointer({{.Field.Slot.Offset}}){{end}}\n}\n\n{{template \"_hasfield\" .}}\n\nfunc (s {{.Node.Name}}) {{.Field.Name | title}}Ptr() ({{.G.Capnp}}.Ptr, error) {\n\t{{if .Default.IsValid}}p, err := s.Struct.Ptr({{.Field.Slot.Offset}})\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\treturn p.Default({{.Default}}){{else}}return s.Struct.Ptr({{.Field.Slot.Offset}}){{end}}\n}\n\nfunc (s {{.Node.Name}}) Set{{.Field.Name | title}}(v {{.G.Capnp}}.Pointer) error {\n\t{{template \"_settag\" .}}return s.Struct.SetPointer({{.Field.Slot.Offset}}, v)\n}\n\nfunc (s {{.Node.Name}}) Set{{.Field.Name | title}}Ptr(v {{.G.Capnp}}.Ptr) error {\n\t{{template \"_settag\" .}}return s.Struct.SetPtr({{.Field.Slot.Offset}}, v)\n}\n\n{{end}}{{define \"structStructField\"}}func (s {{.Node.Name}}) {{.Field.Name | title}}() ({{.FieldType}}, error) {\n\tp, err := s.Struct.Ptr({{.Field.Slot.Offset}})\n\t{{if .Default.IsValid}}if err != nil {\n\t\treturn {{.FieldType}}{}, err\n\t}\n\tss, err := p.StructDefault({{.Default}})\n\treturn {{.FieldType}}{Struct: ss}, err{{else}}return {{.FieldType}}{Struct: p.Struct()}, err{{end}}\n}\n\n{{template \"_hasfield\" .}}\n\nfunc (s {{.Node.Name}}) Set{{.Field.Name | title}}(v {{.FieldType}}) error {\n\t{{template \"_settag\" .}}return s.Struct.SetPtr({{.Field.Slot.Offset}}, v.Struct.ToPtr())\n}\n\n// New{{.Field.Name | title}} sets the {{.Field.Name}} field to a newly\n// allocated {{.FieldType}} struct, preferring placement in s's segment.\nfunc (s {{.Node.Name}}) New{{.Field.Name | title}}() ({{.FieldType}}, error) {\n\t{{template \"_settag\" .}}ss, err := {{.G.RemoteNodeNew .TypeNode .Node}}(s.Struct.Segment())\n\tif err != nil {\n\t\treturn {{.FieldType}}{}, err\n\t}\n\terr = s.Struct.SetPtr({{.Field.Slot.Offset}}, ss.Struct.ToPtr())\n\treturn ss, err\n}\n\n{{end}}{{define \"structTextField\"}}func (s {{.Node.Name}}) {{.Field.Name | title}}() (string, error) {\n\tp, err := s.Struct.Ptr({{.Field.Slot.Offset}})\n\t{{with .Default}}return p.TextDefault({{printf \"%q\" .}}), err{{else}}return p.Text(), err{{end}}\n}\n\n{{template \"_hasfield\" .}}\n\nfunc (s {{.Node.Name}}) {{.Field.Name | title}}Bytes() ([]byte, error) {\n\tp, err := s.Struct.Ptr({{.Field.Slot.Offset}})\n\t{{with .Default}}return p.TextBytesDefault({{printf \"%q\" .}}), err{{else}}return p.TextBytes(), err{{end}}\n}\n\nfunc (s {{.Node.Name}}) Set{{.Field.Name | title}}(v string) error {\n\t{{template \"_settag\" .}}t, err := {{.G.Capnp}}.NewText(s.Struct.Segment(), v)\n\tif err != nil {\n\t\treturn err\n\t}\n\treturn s.Struct.SetPtr({{.Field.Slot.Offset}}, t.List.ToPtr())\n}\n\n{{end}}{{define \"structTypes\"}}{{with .Annotations.Doc}}// {{.}}\n{{end}}type {{.Node.Name}} {{if .IsBase}}struct{ {{.G.Capnp}}.Struct }{{else}}{{.BaseNode.Name}}{{end}}\n{{end}}{{define \"structUintField\"}}func (s {{.Node.Name}}) {{.Field.Name | title}}() uint{{.Bits}} {\n\treturn s.Struct.Uint{{.Bits}}({{.Offset}}){{with .Default}} ^ {{.}}{{end}}\n}\n\nfunc (s {{.Node.Name}}) Set{{.Field.Name | title}}(v uint{{.Bits}}) {\n\t{{template \"_settag\" .}}s.Struct.SetUint{{.Bits}}({{.Offset}}, v{{with .Default}}^{{.}}{{end}})\n}\n\n{{end}}{{define \"structValue\"}}{{.G.RemoteNodeName .Typ .Node}}{Struct: {{.G.Capnp}}.MustUnmarshalRootPtr({{.Value}}).Struct()}{{end}}{{define \"structVoidField\"}}{{if .Field.HasDiscriminant}}func (s {{.Node.Name}}) Set{{.Field.Name | title}}() {\n\t{{template \"_settag\" .}}\n}\n\n{{end}}{{end}}"))

func renderAnnotation(r renderer, p annotationParams) error {
	return r.Render("annotation", p)
}
func renderBaseStructFuncs(r renderer, p baseStructFuncsParams) error {
	return r.Render("baseStructFuncs", p)
}
func renderConstants(r renderer, p constantsParams) error {
	return r.Render("constants", p)
}
func renderEnum(r renderer, p enumParams) error {
	return r.Render("enum", p)
}
func renderInterfaceClient(r renderer, p interfaceClientParams) error {
	return r.Render("interfaceClient", p)
}
func renderInterfaceServer(r renderer, p interfaceServerParams) error {
	return r.Render("interfaceServer", p)
}
func renderListValue(r renderer, p listValueParams) error {
	return r.Render("listValue", p)
}
func renderPointerValue(r renderer, p pointerValueParams) error {
	return r.Render("pointerValue", p)
}
func renderPromise(r renderer, p promiseParams) error {
	return r.Render("promise", p)
}
func renderPromiseFieldAnyPointer(r renderer, p promiseFieldAnyPointerParams) error {
	return r.Render("promiseFieldAnyPointer", p)
}
func renderPromiseFieldInterface(r renderer, p promiseFieldInterfaceParams) error {
	return r.Render("promiseFieldInterface", p)
}
func renderPromiseFieldStruct(r renderer, p promiseFieldStructParams) error {
	return r.Render("promiseFieldStruct", p)
}
func renderPromiseGroup(r renderer, p promiseGroupParams) error {
	return r.Render("promiseGroup", p)
}
func renderSchemaVar(r renderer, p schemaVarParams) error {
	return r.Render("schemaVar", p)
}
func renderStructBoolField(r renderer, p structBoolFieldParams) error {
	return r.Render("structBoolField", p)
}
func renderStructDataField(r renderer, p structDataFieldParams) error {
	return r.Render("structDataField", p)
}
func renderStructEnums(r renderer, p structEnumsParams) error {
	return r.Render("structEnums", p)
}
func renderStructFloatField(r renderer, p structFloatFieldParams) error {
	return r.Render("structFloatField", p)
}
func renderStructFuncs(r renderer, p structFuncsParams) error {
	return r.Render("structFuncs", p)
}
func renderStructGroup(r renderer, p structGroupParams) error {
	return r.Render("structGroup", p)
}
func renderStructIntField(r renderer, p structIntFieldParams) error {
	return r.Render("structIntField", p)
}
func renderStructInterfaceField(r renderer, p structInterfaceFieldParams) error {
	return r.Render("structInterfaceField", p)
}
func renderStructList(r renderer, p structListParams) error {
	return r.Render("structList", p)
}
func renderStructListField(r renderer, p structListFieldParams) error {
	return r.Render("structListField", p)
}
func renderStructPointerField(r renderer, p structPointerFieldParams) error {
	return r.Render("structPointerField", p)
}
func renderStructStructField(r renderer, p structStructFieldParams) error {
	return r.Render("structStructField", p)
}
func renderStructTextField(r renderer, p structTextFieldParams) error {
	return r.Render("structTextField", p)
}
func renderStructTypes(r renderer, p structTypesParams) error {
	return r.Render("structTypes", p)
}
func renderStructUintField(r renderer, p structUintFieldParams) error {
	return r.Render("structUintField", p)
}
func renderStructValue(r renderer, p structValueParams) error {
	return r.Render("structValue", p)
}
func renderStructVoidField(r renderer, p structVoidFieldParams) error {
	return r.Render("structVoidField", p)
}
