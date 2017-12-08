/*
Package pogs provides functions to convert Cap'n Proto messages to and
from Go structs.  pogs operates similarly to encoding/json: define a
struct that is optionally marked up with tags, then Insert and Extract
will copy the fields to and from the corresponding Cap'n Proto struct.

Inserting

To copy data into a Cap'n Proto struct, we use the Insert function.
Consider the following schema:

	struct Message {
		name @0 :Text;
		body @1 :Text;
		time @2 :Int64;
	}

and the Go struct:

	type Message struct {
		Name string
		Body string
		Time int64
	}

We can copy the Go struct into a Cap'n Proto struct like this:

	_, arena, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	root, _ := myschema.NewRootMessage(arena)
	m := &Message{"Alice", "Hello", 1294706395881547000}
	err := pogs.Insert(myschema.Message_TypeID, root.Struct, m)

Note that if any field names in our Go struct don't match to a field in
the Cap'n Proto struct, Insert returns an error.  We'll see how to fix
that in a moment.

Extracting

Copying data back out from a Cap'n Proto struct is quite similar: we
pass a pointer to our Go struct to Extract.

	m := new(Message)
	err := pogs.Extract(m, myschema.Message_TypeID, root.Struct)

Types

The mapping between Cap'n Proto types and underlying Go types is as
follows:

	Bool                          -> bool
	Int8, Int16, Int32, Int64     -> int8, int16, int32, int64
	UInt8, UInt16, UInt32, UInt64 -> uint8, uint16, uint32, uint64
	Float32, Float64              -> float32, float64
	Text                          -> either []byte or string
	Data                          -> []byte
	List                          -> slice
	enum                          -> uint16
	struct                        -> a struct or pointer to struct

Note that the unsized int and uint type can't be used: int and float
types must match in size.  For Data and Text fields using []byte, the
filled-in byte slice will point to original segment.

Renaming and Omitting Fields

By default, the Go field name is the same as the Cap'n Proto schema
field name with the first letter capitalized.  If we want to change this
mapping, we use the capnp field tag.

	type MessageRenamed struct {
		Subject    string `capnp:"name"`
		Body       string
		SentMillis int64  `capnp:"time"`
	}

Using a "-" will cause the field to be ignored by the Insert and
Extract functions.

	type ExtraFieldsMessage struct {
		ID   uint64 `capnp:"-"`
		Name string
		Body string
		Time int64
	}

Unions

Since Go does not have support for variant types, Go structs that want
to use fields inside a Cap'n Proto union must have an explicit
discriminant field called Which.  The Extract function will populate the
Which field and the Insert function will read the Which field to
determine which field to set.  Given this schema:

	struct Shape {
		area @0 :Float64;

		union {
			circle @1 :Float64;
			square @2 :Float64;
		}
	}

the Go struct should look like this:

	type Shape struct {
		Area float64

		Which  myschema.Shape_Which  // or any other uint16 type
		Circle float64
		Square float64
	}

Attempting to use fields in a union without a uint16 Which field will
result in an error.  There is one exception: we can declare our Which
field to be fixed to one particular union value by using a field tag.

	type Square struct {
		Which struct{} `capnp:",which=square"`
		Area  float64
		Width float64  `capnp:"square"`
	}

This can be useful if we want to use a different Go type depending on
which field in the union is set.

	shape, err := myschema.ReadRootShape(msg)
	if err != nil {
		return nil, err
	}
	switch shape.Which() {
	case myschema.Shape_Which_square:
		sq := new(Square)
		err = pogs.Extract(sq, myschema.Square_TypeID, shape.Struct)
		return sq, err
	case myschema.Shape_Which_circle:
		// ...
	}

Embedding

Anonymous struct fields are usually extracted or inserted as if their
inner exported fields were fields in the outer struct, subject to the
rules in the next paragraph.  An anonymous struct field with a name
given in its capnp tag is treated as having that name, rather than being
anonymous.  An anonymous struct field with a capnp tag of "-" will be
ignored.

The visibility rules for struct fields are amended for pogs in the same
way they are amended in encoding/json: if there are multiple fields at
the same level, and that level is the least nested, the following extra
rules apply:

1) Of those fields, if any are capnp-tagged, only tagged fields are
considered, even if there are multiple untagged fields that would
otherwise conflict.
2) If there is exactly one field (tagged or not according to the first
rule), that is selected.
3) Otherwise, there are multiple fields, and all are ignored; no error
occurs.
*/
package pogs // import "zombiezen.com/go/capnproto2/pogs"
