package socketstat

import (
	"bytes"
	"errors"
	"reflect"
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
			if !reflect.DeepEqual(tt.err, err) {
				t.Errorf("%d: expected error '%#v' got '%#v'", i, tt.err, err)
			}
			if len(tt.proto) == 0 {
				n := acc.NFields()
				if n != 0 {
					t.Errorf("%d: expected 0 fields if no protocol specified got %d", i, n)
				}
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
					if len(acc.Metrics) < n+1 {
						t.Errorf("%d: expected at least %d values got %d", i, n+1, len(acc.Metrics))
						break
					}
					m := acc.Metrics[n]
					if !reflect.DeepEqual(m.Measurement, ss.measurement) {
						t.Errorf("%d %d %d: expected measurement '%#v' got '%#v'\n", i, j, k, ss.measurement, m.Measurement)
					}
					if !reflect.DeepEqual(m.Tags, tags) {
						t.Errorf("%d %d %d: expected tags\n%#v got\n%#v\n", i, j, k, tags, m.Tags)
					}
					if !reflect.DeepEqual(m.Fields, fields) {
						t.Errorf("%d %d %d: expected fields\n%#v got\n%#v\n", i, j, k, fields, m.Fields)
					}
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
	if !reflect.DeepEqual(err, errFoo) {
		t.Errorf("Expected error %#v got\n%#v\n", errFoo, err)
	}
}
