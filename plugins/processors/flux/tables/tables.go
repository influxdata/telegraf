package tables

import (
	"sync"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
)

// Provider provides fresh tables.
type Provider interface {
	Get() []flux.Table
}

// provider implements Provider and offers fresh an ready-to-use tables at each Get() invocation.
// It is safe to call Get() concurrently.
// This is particularly useful, because many instances of from() can require the tables in parallel.
type provider struct {
	tbls  []flux.BufferedTable
	mutex sync.Mutex
}

func NewProvider(tbls []flux.Table) (Provider, error) {
	btbls, err := copyTables(tbls)
	if err != nil {
		return nil, err
	}
	return &provider{
		tbls: btbls,
	}, nil
}

func (p *provider) Get() []flux.Table {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return copyBufferedTables(p.tbls)
}

func copyTables(tables []flux.Table) ([]flux.BufferedTable, error) {
	tbls := make([]flux.BufferedTable, len(tables))
	for i, tbl := range tables {
		btbl, err := execute.CopyTable(tbl)
		if err != nil {
			return nil, err
		}
		tbls[i] = btbl
	}
	return tbls, nil
}

func copyBufferedTables(tables []flux.BufferedTable) []flux.Table {
	tbls := make([]flux.Table, len(tables))
	for i, tbl := range tables {
		tbls[i] = tbl.Copy()
	}
	return tbls
}
