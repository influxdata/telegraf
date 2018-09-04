package api

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewTypesDB(t *testing.T) {
	input := `
percent			value:GAUGE:0:100.1
total_bytes		value:DERIVE:0:U
signal_noise		value:GAUGE:U:0
mysql_qcache		hits:COUNTER:0:U, inserts:COUNTER:0:U, not_cached:COUNTER:0:U, lowmem_prunes:COUNTER:0:U, queries_in_cache:GAUGE:0:U
`

	db, err := NewTypesDB(strings.NewReader(input))
	if err != nil {
		t.Errorf("NewTypesDB() = %v, want %v", err, nil)
	}

	want := &DataSet{
		Name: "percent",
		Sources: []DataSource{
			DataSource{
				Name: "value",
				Type: reflect.TypeOf(Gauge(0)),
				Min:  0.0,
				Max:  100.1,
			},
		},
	}
	got, ok := db.DataSet("percent")
	if !ok || !reflect.DeepEqual(got, want) {
		t.Errorf("db[%q] = %v, %v, want %v, %v", "percent", got, ok, want, true)
	}

	got, ok = db.DataSet("total_bytes")
	if !ok {
		t.Fatal(`db.DataSet("total_bytes") missing`)
	}
	if !math.IsNaN(got.Sources[0].Max) {
		t.Errorf("got.Sources[0].Max = %g, want %g", got.Sources[0].Max, math.NaN())
	}

	got, ok = db.DataSet("signal_noise")
	if !ok {
		t.Fatal(`db.DataSet("signal_noise") missing`)
	}
	if !math.IsNaN(got.Sources[0].Min) {
		t.Errorf("got.Sources[0].Min = %g, want %g", got.Sources[0].Min, math.NaN())
	}

	got, ok = db.DataSet("mysql_qcache")
	if !ok {
		t.Fatal(`db.DataSet("mysql_qcache") missing`)
	}
	if len(got.Sources) != 5 {
		t.Errorf("len(got.Sources) = %d, want %d", len(got.Sources), 5)
	}
	for i, name := range []string{"hits", "inserts", "not_cached", "lowmem_prunes", "queries_in_cache"} {
		if got.Sources[i].Name != name {
			t.Errorf("got.Sources[%d].Name = %q, want %q", i, got.Sources[i].Name, name)
		}
	}
}

func TestTypesDB_ValueList(t *testing.T) {
	db, err := NewTypesDB(strings.NewReader(`
counter			value:COUNTER:U:U
gauge			value:GAUGE:U:U
derive			value:DERIVE:0:U
if_octets		rx:DERIVE:0:U, tx:DERIVE:0:U
	`))
	if err != nil {
		t.Errorf("NewTypesDB() = %v, want %v", err, nil)
	}

	id := Identifier{
		Host:   "example.com",
		Plugin: "golang",
		Type:   "gauge",
	}
	vl, err := db.ValueList(id, time.Unix(1469175855, 0), 10*time.Second, Gauge(42))
	if err != nil {
		t.Errorf("db.Values(%v, %v, %v, %v) = (%v, %v), want (..., %v)", id, time.Unix(1469175855, 0), 10*time.Second, Gauge(42.0), vl, err, nil)
	}

}

func TestDataSource_Value(t *testing.T) {
	cases := []struct {
		arg       interface{}
		typ       reflect.Type
		wantValue Value
		wantErr   bool
	}{
		// COUNTER
		{int(42), dsTypeCounter, Counter(42), false},
		{uint(42), dsTypeCounter, Counter(42), false},
		{int64(42), dsTypeCounter, Counter(42), false},
		{uint64(42), dsTypeCounter, Counter(42), false},
		{float32(42.5), dsTypeCounter, Counter(42), false},
		{float64(42.8), dsTypeCounter, Counter(42), false},
		{Counter(42), dsTypeCounter, Counter(42), false},
		{true, dsTypeCounter, nil, true},
		{"42", dsTypeCounter, nil, true},
		// DERIVE
		{int(42), dsTypeDerive, Derive(42), false},
		{uint(42), dsTypeDerive, Derive(42), false},
		{int64(42), dsTypeDerive, Derive(42), false},
		{uint64(42), dsTypeDerive, Derive(42), false},
		{float32(42.5), dsTypeDerive, Derive(42), false},
		{float64(42.8), dsTypeDerive, Derive(42), false},
		{Derive(42), dsTypeDerive, Derive(42), false},
		{true, dsTypeDerive, nil, true},
		{"42", dsTypeDerive, nil, true},
		// GAUGE
		{int(42), dsTypeGauge, Gauge(42), false},
		{uint(42), dsTypeGauge, Gauge(42), false},
		{int64(42), dsTypeGauge, Gauge(42), false},
		{uint64(42), dsTypeGauge, Gauge(42), false},
		{float32(42.5), dsTypeGauge, Gauge(42.5), false},
		{float64(42.8), dsTypeGauge, Gauge(42.8), false},
		{Gauge(42.9), dsTypeGauge, Gauge(42.9), false},
		{true, dsTypeGauge, nil, true},
		{"42", dsTypeGauge, nil, true},
	}

	for _, c := range cases {
		dsrc := DataSource{
			Name: "value",
			Type: c.typ,
			Min:  0.0,
			Max:  math.NaN(),
		}

		got, err := dsrc.Value(c.arg)
		if err != nil {
			if !c.wantErr {
				t.Errorf("dsrc.Type = %s; dsrc.Value(%v) = (%v, %v), want <err>", dsrc.Type.Name(), c.arg, got, err)
			}
			continue
		}

		if c.wantErr || !reflect.DeepEqual(got, c.wantValue) {
			var wantErr error
			if c.wantErr {
				wantErr = errors.New("<err>")
			}

			t.Errorf("dsrc.Type = %s; dsrc.Value(%v) = (%v, %v), want (%v, %v)", dsrc.Type.Name(), c.arg, got, err, c.wantValue, wantErr)
		}
	}
}
