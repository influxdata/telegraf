package rfc5425

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenTypeString(t *testing.T) {
	const NOTEXISTING = 1000
	assert.Equal(t, fmt.Sprintf("TokenType(%d)", NOTEXISTING), TokenType(NOTEXISTING).String())
	assert.Equal(t, "ILLEGAL", ILLEGAL.String())
	assert.Equal(t, "WS", TokenType(WS).String())
}

func TestTokenString(t *testing.T) {
	tok := Token{typ: SYSLOGMSG, lit: []byte("<1>1 - - - - - -")}
	assert.Equal(t, "SYSLOGMSG(<1>1 - - - - - -)", tok.String())
}
