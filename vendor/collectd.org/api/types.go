// Package api defines data types representing core collectd data types.
package api // import "collectd.org/api"

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrNoDataset is returned when the data set cannot be found.
	ErrNoDataset = errors.New("no such dataset")
)

var (
	dsTypeCounter = reflect.TypeOf(Counter(0))
	dsTypeDerive  = reflect.TypeOf(Derive(0))
	dsTypeGauge   = reflect.TypeOf(Gauge(0))
)

// TypesDB holds the type definitions of one or more types.db(5) files.
type TypesDB struct {
	rows map[string]*DataSet
}

// NewTypesDB parses collectd's types.db file and returns an initialized
// TypesDB object that can be queried.
func NewTypesDB(r io.Reader) (*TypesDB, error) {
	db := &TypesDB{
		rows: make(map[string]*DataSet),
	}

	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		ds, err := parseSet(s.Text())
		if err != nil {
			continue
		}

		db.rows[ds.Name] = ds
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	return db, nil
}

// Merge adds all entries in other to db, possibly overwriting entries in db
// with the same name.
func (db *TypesDB) Merge(other *TypesDB) {
	for k, v := range other.rows {
		db.rows[k] = v
	}
}

// DataSet returns the DataSet "typ".
// This is similar to collectd's plugin_get_ds() function.
func (db *TypesDB) DataSet(typ string) (*DataSet, bool) {
	s, ok := db.rows[typ]
	return s, ok
}

// ValueList initializes and returns a new ValueList. The number of values
// arguments must match the number of DataSources in the vl.Type DataSet and
// are converted to []Value using DataSet.Values().
func (db *TypesDB) ValueList(id Identifier, t time.Time, interval time.Duration, values ...interface{}) (*ValueList, error) {
	ds, ok := db.DataSet(id.Type)
	if !ok {
		return nil, ErrNoDataset
	}

	v, err := ds.Values(values...)
	if err != nil {
		return nil, err
	}

	return &ValueList{
		Identifier: id,
		Time:       t,
		Interval:   interval,
		DSNames:    ds.Names(),
		Values:     v,
	}, nil
}

// DataSet defines the metrics of a given "Type", i.e. the name, kind (Derive,
// Gauge, …) and minimum and maximum values.
type DataSet struct {
	Name    string
	Sources []DataSource
}

func parseSet(line string) (*DataSet, error) {
	ds := &DataSet{}

	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)

	for s.Scan() {
		if ds.Name == "" {
			ds.Name = s.Text()
			continue
		}

		dsrc, err := parseSource(s.Text())
		if err != nil {
			return nil, err
		}

		ds.Sources = append(ds.Sources, *dsrc)
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	return ds, nil
}

// Names returns a slice of the data source names. This can be used to populate ValueList.DSNames.
func (ds *DataSet) Names() []string {
	var ret []string
	for _, dsrc := range ds.Sources {
		ret = append(ret, dsrc.Name)
	}

	return ret
}

// Values converts the arguments to the Value interface type and returns them
// as a slice. It expects the same number of arguments as it has Sources and
// will return an error if there is a mismatch. Each argument is converted to a
// Counter, Derive or Gauge according to the corresponding DataSource.Type.
func (ds *DataSet) Values(args ...interface{}) ([]Value, error) {
	if len(args) != len(ds.Sources) {
		return nil, fmt.Errorf("len(args) = %d, want %d", len(args), len(ds.Sources))
	}

	var ret []Value
	for i, arg := range args {
		v, err := ds.Sources[i].Value(arg)
		if err != nil {
			return nil, err
		}

		ret = append(ret, v)
	}

	return ret, nil
}

// Check does sanity checking of vl and returns an error if it finds a problem.
// Sanity checking includes checking the concrete types in the Values slice
// against the DataSet's Sources.
func (ds *DataSet) Check(vl *ValueList) error {
	if ds.Name != vl.Type {
		return fmt.Errorf("vl.Type = %q, want %q", vl.Type, ds.Name)
	}

	if len(ds.Sources) != len(vl.Values) {
		return fmt.Errorf("len(vl.Values) = %d, want %d", len(vl.Values), len(ds.Sources))
	}

	if len(ds.Sources) != len(vl.DSNames) {
		return fmt.Errorf("len(vl.DSNames) = %d, want %d", len(vl.DSNames), len(ds.Sources))
	}

	for i, dsrc := range ds.Sources {
		if dsrc.Name != vl.DSNames[i] {
			return fmt.Errorf("vl.DSNames[%d] = %q, want %q", i, vl.DSNames[i], dsrc.Name)
		}

		if reflect.TypeOf(vl.Values[i]) != dsrc.Type {
			return fmt.Errorf("vl.Values[%d] is a %T, want %s", i, vl.Values[i], dsrc.Type.Name())
		}
	}

	return nil
}

// DataSource defines one metric within a "Type" / DataSet. Type is one of
// Counter, Derive and Gauge. Min and Max apply to the rates of Counter and
// Derive types, not the raw incremental value.
type DataSource struct {
	Name     string
	Type     reflect.Type
	Min, Max float64
}

func parseSource(line string) (*DataSource, error) {
	line = strings.TrimSuffix(line, ",")

	f := strings.Split(line, ":")
	if len(f) != 4 {
		return nil, fmt.Errorf("unexpected field count: %d", len(f))
	}

	dsrc := &DataSource{
		Name: f[0],
		Min:  math.NaN(),
		Max:  math.NaN(),
	}

	switch f[1] {
	case "COUNTER":
		dsrc.Type = dsTypeCounter
	case "DERIVE":
		dsrc.Type = dsTypeDerive
	case "GAUGE":
		dsrc.Type = dsTypeGauge
	default:
		return nil, fmt.Errorf("invalid data source type %q", f[1])
	}

	if f[2] != "U" {
		v, err := strconv.ParseFloat(f[2], 64)
		if err != nil {
			return nil, err
		}
		dsrc.Min = v
	}

	if f[3] != "U" {
		v, err := strconv.ParseFloat(f[3], 64)
		if err != nil {
			return nil, err
		}
		dsrc.Max = v
	}

	return dsrc, nil
}

// Value converts arg to a Counter, Derive or Gauge and returns it as the Value
// interface type. Returns an error if arg cannot be converted.
func (dsrc DataSource) Value(arg interface{}) (Value, error) {
	if !reflect.TypeOf(arg).ConvertibleTo(dsrc.Type) {
		return nil, fmt.Errorf("cannot convert %T to %s", arg, dsrc.Type.Name())
	}

	v := reflect.ValueOf(arg).Convert(dsrc.Type)
	switch dsrc.Type {
	case dsTypeCounter:
		return v.Interface().(Counter), nil
	case dsTypeDerive:
		return v.Interface().(Derive), nil
	case dsTypeGauge:
		return v.Interface().(Gauge), nil
	}

	return nil, fmt.Errorf("unexpected data sourc type %s", dsrc.Type.Name())
}
