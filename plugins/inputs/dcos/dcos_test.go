package dcos

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGather(t *testing.T) {
	dcos := &Dcos{"http://m1.dcos", "eyJhbGciOiJIUzI1NiIsImtpZCI6InNlY3JldCIsInR5cCI6IkpXVCJ9.eyJhdWQiOiIzeUY1VE9TemRsSTQ1UTF4c3B4emVvR0JlOWZOeG05bSIsImVtYWlsIjoidmxhc3RpbWlsLmhhamVrQGdtYWlsLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJleHAiOjEuNTAyMTc1NzFlKzA5LCJpYXQiOjEuNTAxNzQzNzFlKzA5LCJpc3MiOiJodHRwczovL2Rjb3MuYXV0aDAuY29tLyIsInN1YiI6Imdvb2dsZS1vYXV0aDJ8MTE2NTEyNTI1MjE2ODg1MjYzODYwIiwidWlkIjoidmxhc3RpbWlsLmhhamVrQGdtYWlsLmNvbSJ9.VhN_n5fi2_2eVLm5GgF8WqK2ORCTiepkBWxmgAqyDBs"}
	var acc testutil.Accumulator
	require.NoError(t, dcos.Gather(&acc))
}
