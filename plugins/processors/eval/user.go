package eval

import (
	"fmt"

	"github.com/influxdata/telegraf"
)

type User struct {
	metric *telegraf.Metric
}

func (u *User) SetMetric(m *telegraf.Metric) {
	u.metric = m
}

func (u *User) Main() {
	fmt.Println("metric", u.metric)
}
