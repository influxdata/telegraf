package main

import (
	"bytes"
	"fmt"
)

type annotationParams struct {
	G    *generator
	Node *node
}

type constantsParams struct {
	G      *generator
	Consts []*node
	Vars   []*node
}

type enumParams struct {
	G           *generator
	Node        *node
	Annotations *annotations
	EnumValues  []enumval
}

type structTypesParams struct {
	G           *generator
	Node        *node
	Annotations *annotations
	BaseNode    *node
}

func (p structTypesParams) IsBase() bool {
	return p.Node == p.BaseNode
}

type baseStructFuncsParams struct {
	G            *generator
	Node         *node
	StringMethod bool
}

type structFuncsParams struct {
	G    *generator
	Node *node
}

type structGroupParams struct {
	G     *generator
	Node  *node
	Group *node
	Field field
}

type structFieldParams struct {
	G           *generator
	Node        *node
	Field       field
	Annotations *annotations
	FieldType   string
}

type (
	structFloatFieldParams     structUintFieldParams
	structInterfaceFieldParams structFieldParams
	structVoidFieldParams      structFieldParams
	structListFieldParams      structObjectFieldParams
	structPointerFieldParams   structObjectFieldParams
	structStructFieldParams    structObjectFieldParams
)

type structBoolFieldParams struct {
	structFieldParams
	Default bool
}

type structUintFieldParams struct {
	structFieldParams
	Bits    uint
	Default uint64
}

func (p structUintFieldParams) Offset() uint32 {
	return p.Field.Slot().Offset() * uint32(p.Bits/8)
}

func (p structFloatFieldParams) Offset() uint32 {
	return structUintFieldParams(p).Offset()
}

type structIntFieldParams struct {
	structUintFieldParams
	EnumName string
}

func (p structIntFieldParams) ReturnType() string {
	if p.EnumName != "" {
		return p.EnumName
	}
	return fmt.Sprintf("int%d", p.Bits)
}

type structTextFieldParams struct {
	structFieldParams
	Default string
}

type structDataFieldParams struct {
	structFieldParams
	Default []byte
}

type structObjectFieldParams struct {
	structFieldParams
	TypeNode *node
	Default  staticDataRef
}

type structListParams struct {
	G    *generator
	Node *node
}

type structEnumsParams struct {
	G          *generator
	Node       *node
	Fields     []field
	EnumString enumString
}

type promiseParams struct {
	G      *generator
	Node   *node
	Fields []field
}

type promiseGroupParams struct {
	G     *generator
	Node  *node
	Field field
	Group *node
}

type promiseFieldStructParams struct {
	G       *generator
	Node    *node
	Field   field
	Struct  *node
	Default staticDataRef
}

type promiseFieldAnyPointerParams struct {
	G     *generator
	Node  *node
	Field field
}

type promiseFieldInterfaceParams struct {
	G         *generator
	Node      *node
	Field     field
	Interface *node
}

type interfaceClientParams struct {
	G           *generator
	Node        *node
	Annotations *annotations
	Methods     []interfaceMethod
}

type interfaceServerParams struct {
	G           *generator
	Node        *node
	Annotations *annotations
	Methods     []interfaceMethod
}

type structValueParams struct {
	G     *generator
	Node  *node
	Typ   *node
	Value staticDataRef
}

type pointerValueParams struct {
	G     *generator
	Value staticDataRef
}

type listValueParams struct {
	G     *generator
	Typ   string
	Value staticDataRef
}

type schemaVarParams struct {
	G       *generator
	FileID  uint64
	NodeIDs []uint64
	schema  []byte
}

func (p schemaVarParams) SchemaLiteral() string {
	const width = 16
	var out bytes.Buffer
	out.WriteByte('"')
	for i, b := range p.schema {
		if i > 0 && i%width == 0 {
			out.WriteString("\" +\n\t\"")
		}
		switch {
		case b < ' ' || b > '~':
			// unprintable
			out.WriteString("\\x")
			out.WriteByte(hexdigit(b >> 4))
			out.WriteByte(hexdigit(b & 0xf))
		case b == '"':
			out.WriteString("\\\"")
		case b == '\\':
			out.WriteString("\\\\")
		default:
			out.WriteByte(b)
		}
	}
	out.WriteByte('"')
	return out.String()
}

func hexdigit(b byte) byte {
	if b < 10 {
		return b + '0'
	}
	return (b - 10) + 'a'
}
