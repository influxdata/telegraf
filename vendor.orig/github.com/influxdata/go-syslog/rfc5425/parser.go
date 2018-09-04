package rfc5425

import (
	"fmt"
	"io"

	"github.com/influxdata/go-syslog/rfc5424"
)

// Parser is capable to parse byte buffer on the basis of RFC5425.
//
// Use NewParser function to instantiate one.
type Parser struct {
	msglen     int64
	s          Scanner
	p          rfc5424.Parser
	last       Token
	stepback   bool // Wheter to retrieve the last token or not
	bestEffort bool // Best effort mode flag
}

// ParserOpt represents the type of options setters.
type ParserOpt func(p *Parser) *Parser

// NewParser returns a pointer to a new instance of Parser.
func NewParser(r io.Reader, opts ...ParserOpt) *Parser {
	p := &Parser{
		s: *NewScanner(r),
		p: *rfc5424.NewParser(),
	}

	for _, opt := range opts {
		p = opt(p)
	}

	return p
}

// WithBestEffort sets the best effort mode on.
//
// When active the parser tries to recover as much of the syslog messages as possible.
func WithBestEffort() ParserOpt {
	return func(p *Parser) *Parser {
		p.bestEffort = true
		return p
	}
}

// Result represent the resulting syslog message and (eventually) errors occured during parsing.
type Result struct {
	Message      *rfc5424.SyslogMessage
	MessageError error
	Error        error
}

// ResultHandler is a function the user can use to specify what to do with every Result instance.
type ResultHandler func(result *Result)

// Parse parses the incoming bytes accumulating the results.
func (p *Parser) Parse() []Result {
	results := []Result{}

	acc := func(result *Result) {
		results = append(results, *result)
	}

	p.ParseExecuting(acc)

	return results
}

// ParseExecuting parses the incoming bytes executing the handler function for each Result.
//
// It stops parsing when an error regarding RFC 5425 is found.
func (p *Parser) ParseExecuting(handler ResultHandler) {
	for {
		var tok Token

		// First token MUST be a MSGLEN
		if tok = p.scan(); tok.typ != MSGLEN {
			handler(&Result{
				Error: fmt.Errorf("found %s, expecting a %s", tok, MSGLEN),
			})
			break
		}

		// Next we MUST see a WS
		if tok = p.scan(); tok.typ != WS {
			handler(&Result{
				Error: fmt.Errorf("found %s, expecting a %s", tok, WS),
			})
			break
		}

		// Next we MUST see a SYSLOGMSG with length equal to MSGLEN
		if tok = p.scan(); tok.typ != SYSLOGMSG {
			e := fmt.Errorf(`found %s after "%s", expecting a %s containing %d octets`, tok, tok.lit, SYSLOGMSG, p.s.msglen)
			// Overflow case
			if len(tok.lit) < int(p.s.msglen) && p.bestEffort {
				// Though MSGLEN was not respected, we try to parse the existing SYSLOGMSG as a RFC5424 syslog message
				result := p.parse(tok.lit)
				result.Error = e
				handler(result)
				break
			}

			handler(&Result{
				Error: e,
			})
			break
		}

		// Parse the SYSLOGMSG literal pretending it is a RFC5424 syslog message
		result := p.parse(tok.lit)
		if p.bestEffort || result.MessageError == nil {
			handler(result)
		}
		if !p.bestEffort && result.MessageError != nil {
			handler(&Result{MessageError: result.MessageError})
			break
		}

		// Next we MUST see an EOF otherwise the parsing we'll start again
		if tok = p.scan(); tok.typ == EOF {
			break
		} else {
			p.unscan()
		}
	}
}

func (p *Parser) parse(input []byte) *Result {
	sys, err := p.p.Parse(input, &p.bestEffort)

	return &Result{
		Message:      sys,
		MessageError: err,
	}
}

// scan returns the next token from the underlying scanner;
// if a token has been unscanned then read that instead.
func (p *Parser) scan() Token {
	// If we have a token on the buffer, then return it.
	if p.stepback {
		p.stepback = false
		return p.last
	}

	// Otherwise read the next token from the scanner.
	tok := p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.last = tok

	return tok
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() {
	p.stepback = true
}
