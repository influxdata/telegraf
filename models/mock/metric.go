package models

import "github.com/influxdata/telegraf"

type MockMetric struct {
	telegraf.Metric
	AcceptF func()
	RejectF func()
	DropF   func()
}

func (m *MockMetric) Accept() {
	m.AcceptF()
}

func (m *MockMetric) Reject() {
	m.RejectF()
}

func (m *MockMetric) Drop() {
	m.DropF()
}

func (m *MockMetric) Unwrap() telegraf.Metric {
	return m.Metric
}
