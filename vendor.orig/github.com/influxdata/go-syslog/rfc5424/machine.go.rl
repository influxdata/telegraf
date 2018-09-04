package rfc5424

import (
	"time"
	"fmt"
)

var (
	errPrival         = "expecting a priority value in the range 1-191 or equal to 0 [col %d]"
	errPri            = "expecting a priority value within angle brackets [col %d]"
	errVersion        = "expecting a version value in the range 1-999 [col %d]"
	errTimestamp      = "expecting a RFC3339MICRO timestamp or a nil value [col %d]"
	errHostname       = "expecting an hostname (from 1 to max 255 US-ASCII characters) or a nil value [col %d]"
	errAppname        = "expecting an app-name (from 1 to max 48 US-ASCII characters) or a nil value [col %d]"
	errProcid         = "expecting a procid (from 1 to max 128 US-ASCII characters) or a nil value [col %d]"
	errMsgid          = "expecting a msgid (from 1 to max 32 US-ASCII characters) or a nil value [col %d]"
	errStructuredData = "expecting a structured data section containing one or more elements (`[id( key=\"value\")*]+`) or a nil value [col %d]"
	errSdID           = "expecting a structured data element id (from 1 to max 32 US-ASCII characters; except `=`, ` `, `]`, and `\"` [col %d]"
	errSdIDDuplicated = "duplicate structured data element id [col %d]"
	errSdParam        = "expecting a structured data parameter (`key=\"value\"`, both part from 1 to max 32 US-ASCII characters; key cannot contain `=`, ` `, `]`, and `\"`, while value cannot contain `]`, backslash, and `\"` unless escaped) [col %d]"
	errMsg            = "expecting a free-form optional message in UTF-8 (starting with or without BOM) [col %d]"
	errEscape         = "expecting chars `]`, `\"`, and `\\` to be escaped within param value [col %d]"
	errParse          = "parsing error [col %d]"
)

const RFC3339MICRO = "2006-01-02T15:04:05.999999Z07:00"

%%{
machine rfc5424;

include rfc5424 "rfc5424.rl";

# unsigned alphabet
alphtype uint8;

action mark {
	m.pb = m.p
}

action markmsg {
	m.msgat = m.p
}

action set_prival {
	output.priority = uint8(unsafeUTF8DecimalCodePointsToInt(m.text()))
	output.prioritySet = true
}

action set_version {
	output.version = uint16(unsafeUTF8DecimalCodePointsToInt(m.text()))
}

action set_timestamp {
	if t, e := time.Parse(RFC3339MICRO, string(m.text())); e != nil {
        m.err = fmt.Errorf("%s [col %d]", e, m.p)
		fhold;
    	fgoto fail;
    } else {
        output.timestamp = t
		output.timestampSet = true
    }
}

action set_hostname {
	output.hostname = string(m.text())
}

action set_appname {
	output.appname = string(m.text())
}

action set_procid {
	output.procID = string(m.text())
}

action set_msgid {
	output.msgID = string(m.text())
}

action ini_elements {
	output.structuredData = map[string]map[string]string{}
}

action set_id {
	if _, ok := output.structuredData[string(m.text())]; ok {
		// As per RFC5424 section 6.3.2 SD-ID MUST NOT exist more than once in a message
		m.err = fmt.Errorf(errSdIDDuplicated, m.p)
		fhold;
		fgoto fail;
	} else {
		id := string(m.text())
		output.structuredData[id] = map[string]string{}
		output.hasElements = true
		m.currentelem = id
	}
}

action ini_sdparam {
	m.backslashat = []int{}
}

action add_slash {
	m.backslashat = append(m.backslashat, m.p)
}

action set_paramname {
	m.currentparam = string(m.text())
}

action set_paramvalue {
	if output.hasElements {
		// (fixme) > what if SD-PARAM-NAME already exist for the current element (ie., current SD-ID)?

		// Store text
		text := m.text()
		
		// Strip backslashes only when there are ...
		if len(m.backslashat) > 0 {
			text = rmchars(text, m.backslashat, m.pb)
		}
		output.structuredData[m.currentelem][m.currentparam] = string(text)
	}
}

action set_msg {
	output.message = string(m.text())
}

action err_prival {
	m.err = fmt.Errorf(errPrival, m.p)
	fhold;
    fgoto fail;
}

action err_pri {
	m.err = fmt.Errorf(errPri, m.p)
	fhold;
    fgoto fail;
}

action err_version {
	m.err = fmt.Errorf(errVersion, m.p)
	fhold;
    fgoto fail;
}

action err_timestamp {
	m.err = fmt.Errorf(errTimestamp, m.p)
	fhold;
    fgoto fail;
}

action err_hostname {
	m.err = fmt.Errorf(errHostname, m.p)
	fhold;
    fgoto fail;
}

action err_appname {
	m.err = fmt.Errorf(errAppname, m.p)
	fhold;
    fgoto fail;
}

action err_procid {
	m.err = fmt.Errorf(errProcid, m.p)
	fhold;
    fgoto fail;
}

action err_msgid {
	m.err = fmt.Errorf(errMsgid, m.p)
	fhold;
    fgoto fail;
}

action err_structureddata {
	m.err = fmt.Errorf(errStructuredData, m.p)
	fhold;
    fgoto fail;
}

action err_sdid {
	delete(output.structuredData, m.currentelem)
	if len(output.structuredData) == 0 {
		output.hasElements = false
	}
	m.err = fmt.Errorf(errSdID, m.p)
	fhold;
    fgoto fail;
}

action err_sdparam {
	if len(output.structuredData) > 0 {
		delete(output.structuredData[m.currentelem], m.currentparam)
	}
	m.err = fmt.Errorf(errSdParam, m.p)
	fhold;
    fgoto fail;
}

action err_msg {
	// If error encountered within the message rule ...
	if m.msgat > 0 {
		// Save the text until valid (m.p is where the parser has stopped)
		output.message = string(m.data[m.msgat:m.p])
	}

	m.err = fmt.Errorf(errMsg, m.p)
	fhold;
    fgoto fail;
}

action err_escape {
	m.err = fmt.Errorf(errEscape, m.p)
	fhold;
    fgoto fail;
}

action err_parse {
	m.err = fmt.Errorf(errParse, m.p)
	fhold;
    fgoto fail;
}

nilvalue = '-';

nonzerodigit = '1'..'9';

# 1..191
privalrange = (('1' ('9' ('0'..'1'){,1} | '0'..'8' ('0'..'9'){,1}){,1}) | ('2'..'9' ('0'..'9'){,1}));

# 1..191 or 0
prival = (privalrange | '0') >mark %from(set_prival) $err(err_prival);

pri = ('<' prival '>') @err(err_pri);

version = (nonzerodigit digit{0,2} <err(err_version)) >mark %from(set_version) %eof(set_version) @err(err_version);

timestamp = (nilvalue | (fulldate >mark 'T' fulltime %set_timestamp %err(set_timestamp))) @err(err_timestamp);

hostname = hostnamerange >mark %set_hostname $err(err_hostname);

appname = appnamerange >mark %set_appname $err(err_appname);

procid = procidrange >mark %set_procid $err(err_procid);

msgid = msgidrange >mark %set_msgid $err(err_msgid);

header = (pri version sp timestamp sp hostname sp appname sp procid sp msgid) <>err(err_parse);

# \", \], \\
escapes = (bs >add_slash toescape) $err(err_escape);

# As per section 6.3.3 param value MUST NOT contain '"', '\' and ']', unless they are escaped.
# A backslash '\' followed by none of the this three characters is an invalid escape sequence.
# In this case, treat it as a regular backslash and the following character as a regular character (not altering the invalid sequence).
paramvalue = (utf8charwodelims* escapes*)+ >mark %set_paramvalue;

paramname = sdname >mark %set_paramname;

sdparam = (paramname '=' dq paramvalue dq) >ini_sdparam $err(err_sdparam);

# (note) > finegrained semantics of section 6.3.2 not represented here since not so useful for parsing goal
sdid = sdname >mark %set_id %err(set_id) $err(err_sdid);

sdelement = ('[' sdid (sp sdparam)* ']');

structureddata = nilvalue | sdelement+ >ini_elements $err(err_structureddata);

msg = (bom? utf8octets) >mark >markmsg %set_msg $err(err_msg);

fail := (any - [\n\r])* @err{ fgoto main; };

main := header sp structureddata (sp msg)? $err(err_parse);

}%%

%% write data noerror noprefix;

type machine struct {
	data         []byte
	cs           int
	p, pe, eof   int
	pb           int
	err          error
	currentelem  string
	currentparam string
	msgat        int
	backslashat  []int
}

// NewMachine creates a new FSM able to parse RFC5424 syslog messages.
func NewMachine() *machine {
	m := &machine{}

	%% access m.;
	%% variable p m.p;
	%% variable pe m.pe;
	%% variable eof m.eof;
	%% variable data m.data;

	return m
}

// Err returns the error that occurred on the last call to Parse.
//
// If the result is nil, then the line was parsed successfully.
func (m *machine) Err() error {
	return m.err
}

func (m *machine) text() []byte {
	return m.data[m.pb:m.p]
}

// Parse parses the input byte array as a RFC5424 syslog message.
//
// When a valid RFC5424 syslog message is given it outputs its structured representation.
// If the parsing detects an error it returns it with the position where the error occurred.
//
// It can also partially parse input messages returning a partially valid structured representation
// and the error that stopped the parsing.
func (m *machine) Parse(input []byte, bestEffort *bool) (*SyslogMessage, error) {
	m.data = input
	m.p = 0
	m.pb = 0
	m.msgat = 0
	m.backslashat = []int{}
	m.pe = len(input)
	m.eof = len(input)
	m.err = nil
	output := &syslogMessage{}

    %% write init;
    %% write exec;

	if m.cs < first_final || m.cs == en_fail {
		if bestEffort != nil && *bestEffort && output.valid() {
			// An error occurred but partial parsing is on and partial message is minimally valid
			return output.export(), m.err
		}
		return nil, m.err
	}

	return output.export(), nil
}