package encoding

import (
	"image"
	"reflect"
	"testing"
	"time"
)

var encodeExpected = map[string]interface{}{
	"Level0":  int64(1),
	"Level1b": int64(2),
	"Level1c": int64(3),
	"Level1a": int64(5),
	"LEVEL1B": int64(6),
	"e": map[string]interface{}{
		"Level1a": int64(8),
		"Level1b": int64(9),
		"Level1c": int64(10),
		"Level1d": int64(11),
		"x":       int64(12),
	},
	"Loop1": int64(13),
	"Loop2": int64(14),
	"X":     int64(15),
	"Y":     int64(16),
	"Z":     int64(17),
}

func TestEncode(t *testing.T) {
	// Top is defined in decoder_test.go
	var in = Top{
		Level0: 1,
		Embed0: Embed0{
			Level1b: 2,
			Level1c: 3,
		},
		Embed0a: &Embed0a{
			Level1a: 5,
			Level1b: 6,
		},
		Embed0b: &Embed0b{
			Level1a: 8,
			Level1b: 9,
			Level1c: 10,
			Level1d: 11,
			Level1e: 12,
		},
		Loop: Loop{
			Loop1: 13,
			Loop2: 14,
		},
		Embed0p: Embed0p{
			Point: image.Point{X: 15, Y: 16},
		},
		Embed0q: Embed0q{
			Point: Point{Z: 17},
		},
	}

	got, err := Encode(&in)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, encodeExpected) {
		t.Errorf(" got: %v\nwant: %v\n", got, encodeExpected)
	}
}

type Optionals struct {
	Sr string `gorethink:"sr"`
	So string `gorethink:"so,omitempty"`
	Sw string `gorethink:"-"`

	Ir int `gorethink:"omitempty"` // actually named omitempty, not an option
	Io int `gorethink:"io,omitempty"`

	Tr time.Time `gorethink:"tr"`
	To time.Time `gorethink:"to,omitempty"`

	Slr []string `gorethink:"slr"`
	Slo []string `gorethink:"slo,omitempty"`

	Mr map[string]interface{} `gorethink:"mr"`
	Mo map[string]interface{} `gorethink:",omitempty"`
}

var optionalsExpected = map[string]interface{}{
	"sr":        "",
	"omitempty": int64(0),
	"tr":        map[string]interface{}{"$reql_type$": "TIME", "epoch_time": 0, "timezone": "+00:00"},
	"slr":       []interface{}{},
	"mr":        map[string]interface{}{},
}

func TestOmitEmpty(t *testing.T) {
	var o Optionals
	o.Sw = "something"
	o.Tr = time.Unix(0, 0)
	o.Mr = map[string]interface{}{}
	o.Mo = map[string]interface{}{}

	got, err := Encode(&o)
	if err != nil {
		t.Fatal(err)
	}
	if !jsonEqual(got, optionalsExpected) {
		t.Errorf("\ngot:  %#v\nwant: %#v\n", got, optionalsExpected)
	}
}

type IntType int

type MyStruct struct {
	IntType
}

func TestAnonymousNonstruct(t *testing.T) {
	var i IntType = 11
	a := MyStruct{i}
	var want = map[string]interface{}{"IntType": int64(11)}

	got, err := Encode(a)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestEncodePointer(t *testing.T) {
	v := Pointer{PPoint: &Point{Z: 1}, Point: Point{Z: 2}}
	var want = map[string]interface{}{
		"PPoint": map[string]interface{}{"Z": int64(1)},
		"Point":  map[string]interface{}{"Z": int64(2)},
	}

	got, err := Encode(v)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestEncodeNilPointer(t *testing.T) {
	v := Pointer{PPoint: nil, Point: Point{Z: 2}}
	var want = map[string]interface{}{
		"PPoint": nil,
		"Point":  map[string]interface{}{"Z": int64(2)},
	}

	got, err := Encode(v)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

type BugA struct {
	S string
}

type BugB struct {
	BugA
	S string
}

type BugC struct {
	S string
}

// Legal Go: We never use the repeated embedded field (S).
type BugX struct {
	A int
	BugA
	BugB
}

// Issue 5245.
func TestEmbeddedBug(t *testing.T) {
	v := BugB{
		BugA{"A"},
		"B",
	}
	got, err := Encode(v)
	if err != nil {
		t.Fatal("Encode:", err)
	}
	want := map[string]interface{}{"S": "B"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Encode: got %v want %v", got, want)
	}
	// Now check that the duplicate field, S, does not appear.
	x := BugX{
		A: 23,
	}
	got, err = Encode(x)
	if err != nil {
		t.Fatal("Encode:", err)
	}
	want = map[string]interface{}{"A": int64(23)}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Encode: got %v want %v", got, want)
	}
}

type BugD struct { // Same as BugA after tagging.
	XXX string `gorethink:"S"`
}

// BugD's tagged S field should dominate BugA's.
type BugY struct {
	BugA
	BugD
}

// Test that a field with a tag dominates untagged fields.
func TestTaggedFieldDominates(t *testing.T) {
	v := BugY{
		BugA{"BugA"},
		BugD{"BugD"},
	}
	got, err := Encode(v)
	if err != nil {
		t.Fatal("Encode:", err)
	}
	want := map[string]interface{}{"S": "BugD"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Encode: got %v want %v", got, want)
	}
}

// There are no tags here, so S should not appear.
type BugZ struct {
	BugA
	BugC
	BugY // Contains a tagged S field through BugD; should not dominate.
}

func TestDuplicatedFieldDisappears(t *testing.T) {
	v := BugZ{
		BugA{"BugA"},
		BugC{"BugC"},
		BugY{
			BugA{"nested BugA"},
			BugD{"nested BugD"},
		},
	}
	got, err := Encode(v)
	if err != nil {
		t.Fatal("Encode:", err)
	}
	want := map[string]interface{}{}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Encode: got %v want %v", got, want)
	}
}
