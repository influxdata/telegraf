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
	EOF = errors.New("EOF")
)

%%{
machine LineProtocol;

action begin {
	m.pb = m.p
}

action name_error {
	err = ErrNameParse
	fhold;
	fnext discard_line;
	fbreak;
}

action field_error {
	err = ErrFieldParse
	fhold;
	fnext discard_line;
	fbreak;
}

action tagset_error {
	err = ErrTagParse
	fhold;
	fnext discard_line;
	fbreak;
}

action timestamp_error {
	err = ErrTimestampParse
	fhold;
	fnext discard_line;
	fbreak;
}

action parse_error {
	err = ErrParse
	fhold;
	fnext discard_line;
	fbreak;
}

action align_error {
	err = ErrParse
	fnext discard_line;
	fbreak;
}

action hold_recover {
	fhold;
	fgoto main;
}

action goto_align {
	fgoto align;
}

action found_metric {
	foundMetric = true
}

action name {
	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		fhold;
		fnext discard_line;
		fbreak;
	}
}

action tagkey {
	key = m.text()
}

action tagvalue {
	err = m.handler.AddTag(key, m.text())
	if err != nil {
		fhold;
		fnext discard_line;
		fbreak;
	}
}

action fieldkey {
	key = m.text()
}

action integer {
	err = m.handler.AddInt(key, m.text())
	if err != nil {
		fhold;
		fnext discard_line;
		fbreak;
	}
}

action unsigned {
	err = m.handler.AddUint(key, m.text())
	if err != nil {
		fhold;
		fnext discard_line;
		fbreak;
	}
}

action float {
	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		fhold;
		fnext discard_line;
		fbreak;
	}
}

action bool {
	err = m.handler.AddBool(key, m.text())
	if err != nil {
		fhold;
		fnext discard_line;
		fbreak;
	}
}

action string {
	err = m.handler.AddString(key, m.text())
	if err != nil {
		fhold;
		fnext discard_line;
		fbreak;
	}
}

action timestamp {
	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		fhold;
		fnext discard_line;
		fbreak;
	}
}

action incr_newline {
	m.lineno++
	m.sol = m.p
	m.sol++ // next char will be the first column in the line
}

action eol {
	fnext align;
	fbreak;
}

ws =
	[\t\v\f ];

newline =
	'\r'? '\n' %to(incr_newline);

non_zero_digit =
	[1-9];

integer =
	'-'? ( digit | ( non_zero_digit digit* ) );

unsigned =
	( digit | ( non_zero_digit digit* ) );

number =
	'-'? (digit+ ('.' digit*)? | '.' digit+);

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

fieldunsigned =
	(unsigned 'u') >begin %unsigned;

false =
	"false" | "FALSE" | "False" | "F" | "f";

true =
	"true" | "TRUE" | "True" | "T" | "t";

fieldbool =
	(true | false) >begin %bool;

fieldstringchar =
	[^\f\r\n\\"] | '\\' [\\"] | newline;

fieldstring =
	fieldstringchar* >begin %string;

fieldstringquoted =
	'"' fieldstring '"';

fieldvalue = fieldinteger | fieldunsigned | fieldfloat | fieldstringquoted | fieldbool;

field =
	fieldkey '=' fieldvalue;

fieldset =
	field ( ',' field )*;

tagchar =
	[^\t\n\f\r ,=\\] | ( '\\' [^\t\n\f\r\\] ) | '\\\\' %to{ fhold; };

tagkey =
	tagchar+ >begin %tagkey;

tagvalue =
	tagchar+ >begin %eof(tagvalue) %tagvalue;

tagset =
	((',' tagkey '=' tagvalue) $err(tagset_error))*;

measurement_chars =
	[^\t\n\f\r ,\\] | ( '\\' [^\t\n\f\r] );

measurement_start =
	measurement_chars - '#';

measurement =
	(measurement_start measurement_chars*) >begin %eof(name) %name;

eol_break =
	newline %to(eol)
	;

metric =
	measurement >err(name_error)
	tagset
	ws+ fieldset $err(field_error)
	(ws+ timestamp)? $err(timestamp_error)
	;

line_with_term =
	ws* metric ws* eol_break
	;

line_without_term =
	ws* metric ws*
	;

main :=
	(line_with_term*
	(line_with_term | line_without_term?)
    ) >found_metric
	;

# The discard_line machine discards the current line.  Useful for recovering
# on the next line when an error occurs.
discard_line :=
	(any -- newline)* newline @goto_align;

commentline =
	ws* '#' (any -- newline)* newline;

emptyline =
	ws* newline;

# The align machine scans forward to the start of the next line.  This machine
# is used to skip over whitespace and comments, keeping this logic out of the
# main machine.
#
# Skip valid lines that don't contain line protocol, any other data will move
# control to the main parser via the err action.
align :=
	(emptyline | commentline | ws+)* %err(hold_recover);

# Series is a machine for matching measurement+tagset
series :=
	(measurement >err(name_error) tagset eol_break?)
	>found_metric
	;
}%%

%% write data;

type Handler interface {
	SetMeasurement(name []byte) error
	AddTag(key []byte, value []byte) error
	AddInt(key []byte, value []byte) error
	AddUint(key []byte, value []byte) error
	AddFloat(key []byte, value []byte) error
	AddString(key []byte, value []byte) error
	AddBool(key []byte, value []byte) error
	SetTimestamp(tm []byte) error
}

type machine struct {
	data       []byte
	cs         int
	p, pe, eof int
	pb         int
	lineno     int
	sol        int
	handler    Handler
	initState  int
}

func NewMachine(handler Handler) *machine {
	m := &machine{
		handler: handler,
		initState: LineProtocol_en_align,
	}

	%% access m.;
	%% variable p m.p;
	%% variable cs m.cs;
	%% variable pe m.pe;
	%% variable eof m.eof;
	%% variable data m.data;
	%% write init;

	return m
}

func NewSeriesMachine(handler Handler) *machine {
	m := &machine{
		handler: handler,
		initState: LineProtocol_en_series,
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
	m.lineno = 1
	m.sol = 0
	m.pe = len(data)
	m.eof = len(data)

	%% write init;
	m.cs = m.initState
}

// Next parses the next metric line and returns nil if it was successfully
// processed.  If the line contains a syntax error an error is returned,
// otherwise if the end of file is reached before finding a metric line then
// EOF is returned.
func (m *machine) Next() error {
	if m.p == m.pe && m.pe == m.eof {
		return EOF
	}

	var err error
	var key []byte
	foundMetric := false

	%% write exec;

	if err != nil {
		return err
	}

	// This would indicate an error in the machine that was reported with a
	// more specific error.  We return a generic error but this should
	// possibly be a panic.
	if m.cs == %%{ write error; }%% {
		m.cs = LineProtocol_en_discard_line
		return ErrParse
	}

	// If we haven't found a metric line yet and we reached the EOF, report it
	// now.  This happens when the data ends with a comment or whitespace.
	//
	// Otherwise we have successfully parsed a metric line, so if we are at
	// the EOF we will report it the next call.
	if !foundMetric && m.p == m.pe && m.pe == m.eof {
		return EOF
	}

	return nil
}

// Position returns the current byte offset into the data.
func (m *machine) Position() int {
	return m.p
}

// LineOffset returns the byte offset of the current line.
func (m *machine) LineOffset() int {
	return m.sol
}

// LineNumber returns the current line number.  Lines are counted based on the
// regular expression `\r?\n`.
func (m *machine) LineNumber() int {
	return m.lineno
}

// Column returns the current column.
func (m *machine) Column() int {
	lineOffset := m.p - m.sol
	return lineOffset + 1
}

func (m *machine) text() []byte {
	return m.data[m.pb:m.p]
}
