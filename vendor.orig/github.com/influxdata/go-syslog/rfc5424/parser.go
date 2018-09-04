package rfc5424

import (
	"sync"
)

// Parser represent a RFC5424 FSM with mutex capabilities.
type Parser struct {
	sync.Mutex
	*machine
}

// NewParser creates a new parser and the underlying FSM.
func NewParser() *Parser {
	return &Parser{
		machine: NewMachine(),
	}
}

// Parse parses the input RFC5424 syslog message using its FSM.
//
// Best effort mode enables the partial parsing.
func (p *Parser) Parse(input []byte, bestEffort *bool) (*SyslogMessage, error) {
	p.Lock()
	defer p.Unlock()

	msg, err := p.machine.Parse(input, bestEffort)
	if err != nil {
		if bestEffort != nil && *bestEffort {
			return msg, err
		}
		return nil, err
	}

	return msg, nil
}
