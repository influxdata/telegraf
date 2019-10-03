package flux

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/telegraf/plugins/processors/flux/tables"
	// Cannot import builtin because we need to register telegraf.from().
	_ "github.com/influxdata/flux/stdlib"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/processors"
)

const (
	description  = "Runs a flux script to process inputs and produce outputs"
	sampleConfig = `
  ## Path to a Flux script.
  path = "path/to/script.flux"
  ## Enables debug mode if an output file is specified.
  debug = "path/to/out.csv"
`
	measurementColumnLabel = "_measurement"
)

type Flux struct {
	Path  string `toml:"path"`
	Debug string `toml:"debug"`

	Log telegraf.Logger `toml:"-"`

	debugFile *os.File
	program   flux.Program
}

func init() {
	processors.Add("flux", func() telegraf.Processor {
		return &Flux{}
	})
}

func (p *Flux) SampleConfig() string {
	return sampleConfig
}

func (p *Flux) Description() string {
	return description
}

func (p *Flux) init() error {
	if err := p.compile(); err != nil {
		return fmt.Errorf("script failed to compile: %v", err)
	}
	if len(p.Debug) > 0 && p.debugFile == nil {
		f, err := os.OpenFile(p.Debug, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("could not open debug file: %s", p.Debug)
		}
		p.debugFile = f
	}
	return nil
}

func (p *Flux) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	if len(metrics) == 0 {
		return nil
	}
	if err := p.init(); err != nil {
		p.Log.Errorf("Error on flux processor init: %v", err)
		return nil
	}

	tbls, err := getTables(metrics...)
	if err != nil {
		p.Log.Errorf("Could not extract tables from metrics: %v", err)
		return nil
	}
	if len(p.Debug) > 0 {
		encoded, err := getCSV(tbls)
		if err != nil {
			p.Log.Errorf("Error in encoding CSV: %v", err)
			return nil
		}
		if _, err := p.debugFile.WriteString(encoded); err != nil {
			p.Log.Errorf("Error while writing to debug file: %v", err)
		}
		return nil
	}
	max := time.Unix(0, 0)
	for _, m := range metrics {
		if m.Time().After(max) {
			max = m.Time()
		}
	}
	provider, err := tables.NewProvider(tbls)
	if err != nil {
		p.Log.Errorf("Internal error while creating table provider: %v", err)
		return nil
	}
	deps := Deps{
		Tables: provider,
		Max:    max,
	}
	ctx := deps.Inject(context.Background())
	q, err := p.program.Start(ctx, new(memory.Allocator))
	if err != nil {
		p.Log.Errorf("Error while running program: %v", err)
		return nil
	}
	defer q.Done()
	ms, err := getMetrics(q)
	if err != nil {
		p.Log.Errorf("Error while running program: %v", err)
		return nil
	}
	return ms
}

func (p *Flux) compile() error {
	if p.program != nil {
		return nil
	}
	bs, err := ioutil.ReadFile(p.Path)
	if err != nil {
		return err
	}
	script := string(bs)
	astPkg, err := flux.Parse(script)
	if err != nil {
		return err
	}
	p.program = lang.CompileAST(astPkg, time.Now())
	return nil
}

func getTables(metrics ...telegraf.Metric) ([]flux.Table, error) {
	tbls := make([]flux.Table, 0)
	builders := execute.NewGroupLookup()
	getBuilder := func(gk flux.GroupKey) (execute.TableBuilder, bool) {
		var tb execute.TableBuilder
		v, found := builders.Lookup(gk)
		if !found {
			tb = execute.NewColListTableBuilder(gk, new(memory.Allocator))
			builders.Set(gk, tb)
		} else {
			tb = v.(execute.TableBuilder)
		}
		return tb, !found
	}
	for _, m := range metrics {
		cols := make([]flux.ColMeta, len(m.TagList()))
		vals := make([]values.Value, len(m.TagList()))
		for i, tag := range m.TagList() {
			cols[i] = flux.ColMeta{
				Label: tag.Key,
				Type:  flux.TString,
			}
			vals[i] = values.NewString(tag.Value)
		}
		cols = append(cols, flux.ColMeta{
			Label: measurementColumnLabel,
			Type:  flux.TString,
		})
		vals = append(vals, values.NewString(m.Name()))
		gk := execute.NewGroupKey(cols, vals)
		tb, created := getBuilder(gk)
		if created {
			if err := execute.AddTableKeyCols(gk, tb); err != nil {
				return nil, err
			}
			for _, f := range m.FieldList() {
				t := flux.ColumnType(values.New(f.Value).Type())
				if _, err := tb.AddCol(flux.ColMeta{
					Label: f.Key,
					Type:  t,
				}); err != nil {
					return nil, err
				}
			}
			if _, err := tb.AddCol(flux.ColMeta{
				Label: execute.DefaultTimeColLabel,
				Type:  flux.TTime,
			}); err != nil {
				return nil, err
			}
		}
		timeIdx := execute.ColIdx(execute.DefaultTimeColLabel, tb.Cols())
		if timeIdx < 0 {
			return nil, fmt.Errorf("internal error: column %s should always exist", execute.DefaultTimeColLabel)
		}
		if err := tb.AppendTime(timeIdx, values.ConvertTime(m.Time())); err != nil {
			return nil, err
		}
		if err := execute.AppendKeyValues(gk, tb); err != nil {
			return nil, err
		}
		for _, f := range m.FieldList() {
			v := values.New(f.Value)
			j := execute.ColIdx(f.Key, tb.Cols())
			if j < 0 {
				return nil, fmt.Errorf("cannot find column %s", f.Key)
			}
			if err := tb.AppendValue(j, v); err != nil {
				return nil, err
			}
		}
	}

	var err error
	builders.Range(func(key flux.GroupKey, v interface{}) {
		builder := v.(execute.TableBuilder)
		tbl, e := builder.Table()
		if e != nil {
			err = e
			return
		}
		tbls = append(tbls, tbl)
	})
	return tbls, err
}

func getMetrics(q flux.Query) ([]telegraf.Metric, error) {
	ms := make([]telegraf.Metric, 0)
	for r := range q.Results() {
		if err := r.Tables().Do(func(table flux.Table) error {
			measurementIdx := execute.ColIdx(measurementColumnLabel, table.Cols())
			if measurementIdx < 0 {
				return fmt.Errorf("column %s not found, results should always include that", measurementColumnLabel)
			}
			if table.Cols()[measurementIdx].Type != flux.TString {
				return fmt.Errorf("column %s must be of type string", measurementColumnLabel)
			}
			timeIdx := execute.ColIdx(execute.DefaultTimeColLabel, table.Cols())
			if timeIdx >= 0 && table.Cols()[timeIdx].Type != flux.TTime {
				timeIdx = -1
			}
			return table.Do(func(reader flux.ColReader) error {
				for i := 0; i < reader.Len(); i++ {
					name := reader.Strings(measurementIdx).ValueString(i)
					tags := make(map[string]string, len(table.Key().Cols()))
					fields := make(map[string]interface{}, len(reader.Cols()))
					for i := 0; i < len(table.Key().Cols()); i++ {
						if l := table.Key().Cols()[i].Label; l != measurementColumnLabel {
							tags[table.Key().Cols()[i].Label] = table.Key().ValueString(i)
						}
					}
					for j := 0; j < len(reader.Cols()); j++ {
						if table.Key().HasCol(reader.Cols()[j].Label) || j == timeIdx {
							continue
						}
						v := execute.ValueForRow(reader, i, j)
						fields[reader.Cols()[j].Label] = extractValue(v)
					}
					ts := time.Now()
					if timeIdx >= 0 {
						ts = execute.ValueForRow(reader, i, timeIdx).Time().Time()
					}
					m, err := metric.New(name, tags, fields, ts)
					if err != nil {
						return err
					}
					ms = append(ms, m)
				}
				return nil
			})
		}); err != nil {
			return nil, err
		}
	}
	return ms, q.Err()
}

func extractValue(v values.Value) interface{} {
	if v.IsNull() {
		return nil
	}
	switch t := v.Type(); t {
	case semantic.Int:
		return v.Int()
	case semantic.UInt:
		return v.UInt()
	case semantic.Float:
		return v.Float()
	case semantic.Duration:
		return v.Duration()
	case semantic.Time:
		return v.Time()
	case semantic.String:
		return v.Str()
	case semantic.Bool:
		return v.Bool()
	default:
		panic(fmt.Errorf("unspported type %s", t.Nature().String()))
	}
}

func getCSV(tables []flux.Table) (string, error) {
	enc := csv.NewResultEncoder(csv.DefaultEncoderConfig())
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	_, err := enc.Encode(w, &result{tables: tables})
	if err != nil {
		return "", err
	}
	if err := w.Flush(); err != nil {
		return "", nil
	}
	return b.String(), nil
}

type result struct {
	tables []flux.Table
}

func (r *result) Do(f func(flux.Table) error) error {
	for _, tbl := range r.tables {
		if err := f(tbl); err != nil {
			return err
		}
	}
	return nil
}

func (r *result) Name() string {
	return "_result"
}

func (r *result) Tables() flux.TableIterator {
	return r
}
