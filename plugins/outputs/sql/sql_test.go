package sql

import (
	"testing"
	// "time"
	// "github.com/influxdata/telegraf"
	// "github.com/influxdata/telegraf/metric"
	// "github.com/stretchr/testify/assert"
)

func TestSqlQuote(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

}

func TestSqlCreateStatement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

}

func TestSqlInsertStatement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}
