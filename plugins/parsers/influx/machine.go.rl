package influx

import (
	"errors"
)

var (
	ErrNameParse = errors.New("expected measurement name")
	ErrFieldParse = errors.New("expected field")
	ErrTagParse = errors.New("expected tag")
	ErrTimestampParse = errors.New("expected timestamp")
	ErrParse = errors.New("parse error")
)

%%{
machine LineProtocol;

action begin {
	m.pb = m.p
}

action yield {
	yield = true
	fnext align;
	fbreak;
}

action name_error {
	m.err = ErrNameParse
	fhold;
	fnext discard_line;
	fbreak;
}

action field_error {
	m.err = ErrFieldParse
	fhold;
	fnext discard_line;
	fbreak;
}

action tagset_error {
	m.err = ErrTagParse
	fhold;
	fnext discard_line;
	fbreak;
}

action timestamp_error {
	m.err = ErrTimestampParse
	fhold;
	fnext discard_line;
	fbreak;
}

action parse_error {
	m.err = ErrParse
	fhold;
	fnext discard_line;
	fbreak;
}

action hold_recover {
	fhold;
	fgoto main;
}

action discard {
	fgoto align;
}

action name {
	m.handler.SetMeasurement(m.text())
}

action tagkey {
	key = m.text()
}

action tagvalue {
	m.handler.AddTag(key, m.text())
}

action fieldkey {
	key = m.text()
}

action integer {
	m.handler.AddInt(key, m.text())
}

action float {
	m.handler.AddFloat(key, m.text())
}

action bool {
	m.handler.AddBool(key, m.text())
}

action string {
	m.handler.AddString(key, m.text())
}

action timestamp {
	m.handler.SetTimestamp(m.text())
}

ws =
	[\t\v\f ];

non_zero_digit =
	[1-9];

integer =
	'-'? ( digit | ( non_zero_digit digit* ) );

number =
	( integer ( '.' digit* )? ) | ( '.' digit* );

scientific =
	number 'e'i ["\-+"]? digit+;

timestamp =
	('-'? digit{1,19}) >begin %timestamp;

fieldkeychar =
	[^\t\n\f\r ,=\\] | ( '\\' [^\t\n\f\r] );

fieldkey =
	fieldkeychar+ >begin %fieldkey;

fieldfloat =
	(scientific | number) >begin %float;

fieldinteger =
	(integer 'i') >begin %integer;

false =
	"false" | "FALSE" | "False" | "F" | "f";

true =
	"true" | "TRUE" | "True" | "T" | "t";

fieldbool =
	(true | false) >begin %bool;

fieldstringchar =
	[^\\"] | '\\' [\\"];

fieldstring =
	fieldstringchar* >begin %string;

fieldstringquoted =
	'"' fieldstring '"';

fieldvalue = fieldinteger | fieldfloat | fieldstringquoted | fieldbool;

field =
	fieldkey '=' fieldvalue;

fieldset =
	field ( ',' field )*;

tagchar =
	[^\t\n\f\r ,=\\] | ( '\\' [^\t\n\f\r] );

tagkey =
	tagchar+ >begin %tagkey;

tagvalue =
	tagchar+ >begin %tagvalue;

tagset =
	(',' (tagkey '=' tagvalue) $err(tagset_error))*;

measurement_chars =
	[^\t\n\f\r ,\\] | ( '\\' [^\t\n\f\r] );

measurement_start =
	measurement_chars - '#';

measurement =
	(measurement_start measurement_chars*) >begin %name;

newline =
	[\r\n];

comment =
	'#' (any -- newline)* newline;

eol =
	ws* newline? >yield %eof(yield);

line =
	measurement
	tagset
	(ws+ fieldset) $err(field_error)
	(ws+ timestamp)? $err(timestamp_error)
	eol;

# The main machine parses a single line of line protocol.
main := line $err(parse_error);

# The discard_line machine discards the current line.  Useful for recovering
# on the next line when an error occurs.
discard_line :=
	(any - newline)* newline @discard;

# The align machine scans forward to the start of the next line.  This machine
# is used to skip over whitespace and comments, keeping this logic out of the
# main machine.
align :=
	(space* comment)* space* measurement_start @hold_recover %eof(yield);
}%%

%% write data;

type machine struct {
	data       []byte
	cs         int
	p, pe, eof int
	pb         int
	handler    Handler
	err        error
}

func NewMachine(handler Handler) *machine {
	m := &machine{
		handler: handler,
	}

	%% access m.;
	%% variable p m.p;
	%% variable pe m.pe;
	%% variable eof m.eof;
	%% variable data m.data;
	%% write init;

	return m
}

func (m *machine) SetData(data []byte) {
	m.data = data
	m.p = 0
	m.pb = 0
	m.pe = len(data)
	m.eof = len(data)
	m.err = nil

	%% write init;
	m.cs = LineProtocol_en_align
}

// ParseLine parses a line of input and returns true if more data can be
// parsed.
func (m *machine) ParseLine() bool {
	if m.data == nil || m.p >= m.pe {
		m.err = nil
		return false
	}

	m.err = nil
	var key []byte
	var yield bool

	%% write exec;

	// Even if there was an error, return true. On the next call to this
	// function we will attempt to scan to the next line of input and recover.
	if m.err != nil {
		return true
	}

	// Don't check the error state in the case that we just yielded, because
	// the yield indicates we just completed parsing a line.
	if !yield && m.cs == LineProtocol_error {
		m.err = ErrParse
		return true
	}

	return true
}

// Err returns the error that occurred on the last call to ParseLine.  If the
// result is nil, then the line was parsed successfully.
func (m *machine) Err() error {
	return m.err
}

// Position returns the current position into the input.
func (m *machine) Position() int {
	return m.p
}

func (m *machine) text() []byte {
	return m.data[m.pb:m.p]
}
