// +build go1.8,codegen

package api

import (
	"reflect"
	"strings"
	"testing"
)

func TestUniqueInputAndOutputs(t *testing.T) {
	shamelist["FooService"] = map[string]struct {
		input  bool
		output bool
	}{}
	v := shamelist["FooService"]["OpOutputNoRename"]
	v.output = true
	shamelist["FooService"]["OpOutputNoRename"] = v
	v = shamelist["FooService"]["InputNoRename"]
	v.input = true
	shamelist["FooService"]["OpInputNoRename"] = v
	v = shamelist["FooService"]["BothNoRename"]
	v.input = true
	v.output = true
	shamelist["FooService"]["OpBothNoRename"] = v

	cases := [][]struct {
		expectedInput  string
		expectedOutput string
		operation      string
		input          string
		inputRef       string
		output         string
		outputRef      string
	}{
		{
			{
				expectedInput:  "FooOperationInput",
				expectedOutput: "FooOperationOutput",
				operation:      "FooOperation",
				input:          "FooInputShape",
				inputRef:       "FooInputShapeRef",
				output:         "FooOutputShape",
				outputRef:      "FooOutputShapeRef",
			},
			{
				expectedInput:  "BarOperationInput",
				expectedOutput: "BarOperationOutput",
				operation:      "BarOperation",
				input:          "FooInputShape",
				inputRef:       "FooInputShapeRef",
				output:         "FooOutputShape",
				outputRef:      "FooOutputShapeRef",
			},
		},
		{
			{
				expectedInput:  "FooOperationInput",
				expectedOutput: "FooOperationOutput",
				operation:      "FooOperation",
				input:          "FooInputShape",
				inputRef:       "FooInputShapeRef",
				output:         "FooOutputShape",
				outputRef:      "FooOutputShapeRef",
			},
			{
				expectedInput:  "OpOutputNoRenameInput",
				expectedOutput: "OpOutputNoRenameOutputShape",
				operation:      "OpOutputNoRename",
				input:          "OpOutputNoRenameInputShape",
				inputRef:       "OpOutputNoRenameInputRef",
				output:         "OpOutputNoRenameOutputShape",
				outputRef:      "OpOutputNoRenameOutputRef",
			},
		},
		{
			{
				expectedInput:  "FooOperationInput",
				expectedOutput: "FooOperationOutput",
				operation:      "FooOperation",
				input:          "FooInputShape",
				inputRef:       "FooInputShapeRef",
				output:         "FooOutputShape",
				outputRef:      "FooOutputShapeRef",
			},
			{
				expectedInput:  "OpInputNoRenameInputShape",
				expectedOutput: "OpInputNoRenameOutput",
				operation:      "OpInputNoRename",
				input:          "OpInputNoRenameInputShape",
				inputRef:       "OpInputNoRenameInputRef",
				output:         "OpInputNoRenameOutputShape",
				outputRef:      "OpInputNoRenameOutputRef",
			},
		},
		{
			{
				expectedInput:  "FooOperationInput",
				expectedOutput: "FooOperationOutput",
				operation:      "FooOperation",
				input:          "FooInputShape",
				inputRef:       "FooInputShapeRef",
				output:         "FooOutputShape",
				outputRef:      "FooOutputShapeRef",
			},
			{
				expectedInput:  "OpInputNoRenameInputShape",
				expectedOutput: "OpInputNoRenameOutputShape",
				operation:      "OpBothNoRename",
				input:          "OpInputNoRenameInputShape",
				inputRef:       "OpInputNoRenameInputRef",
				output:         "OpInputNoRenameOutputShape",
				outputRef:      "OpInputNoRenameOutputRef",
			},
		},
	}

	for _, c := range cases {
		a := &API{
			name:       "FooService",
			Operations: map[string]*Operation{},
		}

		expected := map[string][]string{}
		a.Shapes = map[string]*Shape{}
		for _, op := range c {
			a.Operations[op.operation] = &Operation{
				ExportedName: op.operation,
			}
			a.Operations[op.operation].Name = op.operation
			a.Operations[op.operation].InputRef = ShapeRef{
				API:       a,
				ShapeName: op.inputRef,
				Shape: &Shape{
					API:       a,
					ShapeName: op.input,
				},
			}
			a.Operations[op.operation].OutputRef = ShapeRef{
				API:       a,
				ShapeName: op.outputRef,
				Shape: &Shape{
					API:       a,
					ShapeName: op.output,
				},
			}

			a.Shapes[op.input] = &Shape{
				ShapeName: op.input,
			}
			a.Shapes[op.output] = &Shape{
				ShapeName: op.output,
			}

			expected[op.operation] = append(expected[op.operation], op.expectedInput)
			expected[op.operation] = append(expected[op.operation], op.expectedOutput)
		}

		a.fixStutterNames()
		a.renameToplevelShapes()
		for k, v := range expected {
			if a.Operations[k].InputRef.Shape.ShapeName != v[0] {
				t.Errorf("Error %s case: Expected %q, but received %q", k, v[0], a.Operations[k].InputRef.Shape.ShapeName)
			}
			if a.Operations[k].OutputRef.Shape.ShapeName != v[1] {
				t.Errorf("Error %s case: Expected %q, but received %q", k, v[1], a.Operations[k].OutputRef.Shape.ShapeName)
			}
		}

	}
}

func TestCollidingFields(t *testing.T) {
	cases := map[string]struct {
		MemberRefs  map[string]*ShapeRef
		Expect      []string
		IsException bool
	}{
		"SimpleMembers": {
			MemberRefs: map[string]*ShapeRef{
				"Code":     &ShapeRef{},
				"Foo":      &ShapeRef{},
				"GoString": &ShapeRef{},
				"Message":  &ShapeRef{},
				"OrigErr":  &ShapeRef{},
				"SetFoo":   &ShapeRef{},
				"String":   &ShapeRef{},
				"Validate": &ShapeRef{},
			},
			Expect: []string{
				"Code",
				"Foo",
				"GoString_",
				"Message",
				"OrigErr",
				"SetFoo_",
				"String_",
				"Validate_",
			},
		},
		"ExceptionShape": {
			IsException: true,
			MemberRefs: map[string]*ShapeRef{
				"Code":    &ShapeRef{},
				"Message": &ShapeRef{},
				"OrigErr": &ShapeRef{},
				"Other":   &ShapeRef{},
				"String":  &ShapeRef{},
			},
			Expect: []string{
				"Code_",
				"Message_",
				"OrigErr_",
				"Other",
				"String_",
			},
		},
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			a := &API{
				Shapes: map[string]*Shape{
					"shapename": {
						ShapeName:  k,
						MemberRefs: c.MemberRefs,
						Exception:  c.IsException,
					},
				},
			}

			a.renameCollidingFields()

			for i, name := range a.Shapes["shapename"].MemberNames() {
				if e, a := c.Expect[i], name; e != a {
					t.Errorf("expect %v, got %v", e, a)
				}
			}
		})
	}
}

func TestSupressHTTP2EventStreams(t *testing.T) {
	const baseModel = `
{
  "version":"2.0",
  "metadata":{
    "apiVersion":"0000-00-00",
    "endpointPrefix":"rpcservice",
    "jsonVersion":"1.1",
    "protocol":"json",
    "protocolSettings":{"h2":"{h2Option}"},
    "serviceAbbreviation":"RPCService",
    "serviceFullName":"RPC Service",
    "serviceId":"RPCService",
    "signatureVersion":"v4",
    "targetPrefix":"RPCService_00000000",
    "uid":"RPCService-0000-00-00"
  },
  "operations":{
    "BarOp":{
      "name":"BarOp",
      "http":{
        "method":"POST",
        "requestUri":"/"
      },
      "input":{"shape": "BarOpRequest"},
      "output":{"shape":"BarOpResponse"}
    },
    "EventStreamOp":{
      "name":"EventStreamOp",
      "http":{
        "method":"POST",
        "requestUri":"/"
      },
      "input":{"shape": "EventStreamOpRequest"},
      "output":{"shape":"EventStreamOpResponse"}
    },
    "FooOp":{
      "name":"FooOp",
      "http":{
        "method":"POST",
        "requestUri":"/"
      },
      "input":{"shape": "FooOpRequest"},
      "output":{"shape":"FooOpResponse"}
    }
  },
  "shapes":{
    "BarOpRequest":{
      "type":"structure",
      "members":{}
    },
    "BarOpResponse":{
      "type":"structure",
      "members":{}
    },
    "EventStreamOpRequest":{
      "type":"structure",
      "members":{
      }
    },
    "EventStreamOpResponse":{
      "type":"structure",
      "members":{
        "EventStream":{"shape":"EventStream"}
      }
    },
    "FooOpRequest":{
      "type":"structure",
      "members":{}
    },
    "FooOpResponse":{
      "type":"structure",
      "members":{}
    },
    "EventStream":{
      "type":"structure",
      "members":{
        "Empty":{"shape":"EmptyEvent"}
	  },
      "eventstream":true
    },
    "EmptyEvent": {
      "type":"structure",
      "members":{},
      "event": true
    }
  }
}
`

	cases := map[string]struct {
		Model        string
		ExpectOps    []string
		ExpectShapes []string
	}{
		"control": {
			Model:     strings.Replace(baseModel, "{h2Option}", "", -1),
			ExpectOps: []string{"BarOp", "EventStreamOp", "FooOp"},
			ExpectShapes: []string{
				"BarOpInput", "BarOpOutput", "EmptyEvent",
				"EventStreamOpEventStream", "EventStreamOpInput",
				"EventStreamOpOutput", "FooOpInput", "FooOpOutput",
			},
		},
		"HTTP/2 with EventStreams": {
			Model:     strings.Replace(baseModel, "{h2Option}", "eventstream", 1),
			ExpectOps: []string{"BarOp", "FooOp"},
			ExpectShapes: []string{
				"BarOpInput", "BarOpOutput", "FooOpInput", "FooOpOutput",
			},
		},
		"HTTP/2 with optional": {
			Model:     strings.Replace(baseModel, "{h2Option}", "optional", 1),
			ExpectOps: []string{"BarOp", "EventStreamOp", "FooOp"},
			ExpectShapes: []string{
				"BarOpInput", "BarOpOutput", "EmptyEvent",
				"EventStreamOpEventStream", "EventStreamOpInput",
				"EventStreamOpOutput", "FooOpInput", "FooOpOutput",
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var a API
			a.AttachString(c.Model)
			a.APIGoCode()

			if e, a := c.ExpectOps, a.OperationNames(); !reflect.DeepEqual(e, a) {
				t.Errorf("expect %v ops, got %v", e, a)
			}

			if e, a := c.ExpectShapes, a.ShapeNames(); !reflect.DeepEqual(e, a) {
				t.Errorf("expect %v shapes, got %v", e, a)
			}
		})
	}
}
