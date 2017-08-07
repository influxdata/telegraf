package dcos

import (
	"testing"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestGather(t *testing.T) {
	dcos := &Dcos{"http://m1.dcos", "eyJhbGciOiJIUzI1NiIsImtpZCI6InNlY3JldCIsInR5cCI6IkpXVCJ9.eyJhdWQiOiIzeUY1VE9TemRsSTQ1UTF4c3B4emVvR0JlOWZOeG05bSIsImVtYWlsIjoidmxhc3RpbWlsLmhhamVrQGdtYWlsLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJleHAiOjEuNTAyMDMxNzY0ZSswOSwiaWF0IjoxLjUwMTU5OTc2NGUrMDksImlzcyI6Imh0dHBzOi8vZGNvcy5hdXRoMC5jb20vIiwic3ViIjoiZ29vZ2xlLW9hdXRoMnwxMTY1MTI1MjUyMTY4ODUyNjM4NjAiLCJ1aWQiOiJ2bGFzdGltaWwuaGFqZWtAZ21haWwuY29tIn0.4HDhPR990EAXUzgu_9iStEgHYxHpmNKsBDHsLEb1QW8"}
	var acc testutil.Accumulator
	require.NoError(t, dcos.Gather(&acc))
}
