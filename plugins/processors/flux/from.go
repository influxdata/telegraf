package flux

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/telegraf/plugins/processors/flux/tables"
)

const fromKind = "fromTelegraf"

// FromOpSpec is the operation spec for a source that gets the tables as injected
// into dependencies and downstream them to downstream transformations.
type FromOpSpec struct{}

func init() {
	fromSignature := semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{},
		Required:   nil,
		Return:     flux.TableObjectType,
	}
	flux.RegisterPackageValue("telegraf", "from", flux.FunctionValue(fromKind, createFromOpSpec, fromSignature))
	flux.RegisterOpSpec(fromKind, newFromOp)
	plan.RegisterProcedureSpec(fromKind, newFromProcedure, fromKind)
	execute.RegisterSource(fromKind, createFromSource)

	flux.FinalizeBuiltIns()
}

func createFromOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	return newFromOp(), nil
}

func newFromOp() flux.OperationSpec {
	return new(FromOpSpec)
}

func (s *FromOpSpec) Kind() flux.OperationKind {
	return fromKind
}

type FromProcedureSpec struct {
	plan.DefaultCost
}

func newFromProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	_, ok := qs.(*FromOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &FromProcedureSpec{}, nil
}

func (s *FromProcedureSpec) Kind() plan.ProcedureKind {
	return fromKind
}

func (s *FromProcedureSpec) Copy() plan.ProcedureSpec {
	return new(FromProcedureSpec)
}

type key int

const dependenciesKey key = iota

type Deps struct {
	Tables tables.Provider
	Max    time.Time
}

func (d Deps) Inject(ctx context.Context) context.Context {
	return context.WithValue(ctx, dependenciesKey, d)
}

func getDeps(ctx context.Context) Deps {
	return ctx.Value(dependenciesKey).(Deps)
}

func createFromSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	_, ok := prSpec.(*FromProcedureSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", prSpec)
	}
	deps := getDeps(a.Context())
	return &Source{
		id:     dsid,
		tables: deps.Tables,
		max:    values.ConvertTime(deps.Max),
	}, nil
}

type Source struct {
	id execute.DatasetID
	ts []execute.Transformation

	tables tables.Provider
	max    execute.Time
}

func (s *Source) AddTransformation(t execute.Transformation) {
	s.ts = append(s.ts, t)
}

func (s *Source) Run(ctx context.Context) {
	var err error
OUTER:
	for _, t := range s.ts {
		for _, tbl := range s.tables.Get() {
			err = t.Process(s.id, tbl)
			if err != nil {
				break OUTER
			}
		}
		if err = t.UpdateWatermark(s.id, s.max); err != nil {
			break OUTER
		}
	}
	for _, t := range s.ts {
		t.Finish(s.id, err)
	}
}
