package socketstat

import (
	"bytes"
	"errors"
	"testing"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestSocketstat_Gather(t *testing.T) {
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i++
			ss := &Socketstat{
				SocketProto: tt.proto,
				lister: func(cmdName string, proto string, timeout config.Duration) (*bytes.Buffer, error) {
					return bytes.NewBufferString(tt.value), nil
				},
			}
			acc := new(testutil.Accumulator)

			err := ss.Init()
			require.NoError(t, err)

			err = acc.GatherError(ss.Gather)
			require.ErrorIs(t, err, tt.err)
			if len(tt.proto) == 0 {
				n := acc.NFields()
				require.Equalf(t, 0,  n, "%d: expected 0 values got %d", i, n)
				return
			}
			if len(tt.tags) == 0 {
				n := acc.NFields()
				if n != 0 {
					t.Errorf("%d: expected 0 values got %d", i, n)
				}
				return
			}
			n := 0
			for j, tags := range tt.tags {
				for k, fields := range tt.fields[j] {
					require.Greater(t, len(acc.Metrics), n)
					m := acc.Metrics[n]
					require.Equal(t, ss.measurement, m.Measurement, "%d %d %d: expected measurement '%#v' got '%#v'\n", i, j, k, ss.measurement, m.Measurement)
					require.Equal(t, tags, m.Tags, "%d %d %d: expected tags\n%#v got\n%#v\n", i, j, k, tags, m.Tags)
					require.Equal(t, fields, m.Fields, "%d %d %d: expected fields\n%#v got\n%#v\n", i, j, k, fields, m.Fields)
					n++
				}
			}
		})
	}
}

func TestSocketstat_Gather_listerError(t *testing.T) {
	errFoo := errors.New("error foobar")
	ss := &Socketstat{
		SocketProto: []string{"foobar"},
		lister: func(cmdName string, proto string, timeout config.Duration) (*bytes.Buffer, error) {
			return new(bytes.Buffer), errFoo
		},
	}
	acc := new(testutil.Accumulator)
	err := acc.GatherError(ss.Gather)
	require.ErrorIs(t, errFoo, err)
}
