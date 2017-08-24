package elasticsearch

import (
	"context"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	urls := []string{"http://" + testutil.GetLocalHost() + ":9200"}

	e := &Elasticsearch{
		URLs:                urls,
		IndexName:           "test-%Y.%m.%d",
		Timeout:             internal.Duration{Duration: time.Second * 5},
		ManageTemplate:      true,
		TemplateName:        "telegraf",
		OverwriteTemplate:   false,
		HealthCheckInterval: internal.Duration{Duration: time.Second * 10},
	}

	// Verify that we can connect to Elasticsearch
	err := e.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to Elasticsearch
	err = e.Write(testutil.MockMetrics())
	require.NoError(t, err)

}

func TestTemplateManagementEmptyTemplate(t *testing.T) {
	urls := []string{"http://" + testutil.GetLocalHost() + ":9200"}

	ctx := context.Background()

	e := &Elasticsearch{
		URLs:              urls,
		IndexName:         "test-%Y.%m.%d",
		Timeout:           internal.Duration{Duration: time.Second * 5},
		ManageTemplate:    true,
		TemplateName:      "",
		OverwriteTemplate: true,
	}

	err := e.manageTemplate(ctx)
	require.Error(t, err)

}

func TestTemplateManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	urls := []string{"http://" + testutil.GetLocalHost() + ":9200"}

	e := &Elasticsearch{
		URLs:              urls,
		IndexName:         "test-%Y.%m.%d",
		Timeout:           internal.Duration{Duration: time.Second * 5},
		ManageTemplate:    true,
		TemplateName:      "telegraf",
		OverwriteTemplate: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.Timeout.Duration)
	defer cancel()

	err := e.Connect()
	require.NoError(t, err)

	err = e.manageTemplate(ctx)
	require.NoError(t, err)
}

func TestGetIndexName(t *testing.T) {
	e := &Elasticsearch{}

	var tests = []struct {
		EventTime time.Time
		IndexName string
		Expected  string
	}{
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			"indexname",
			"indexname",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			"indexname-%Y",
			"indexname-2014",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			"indexname-%Y-%m",
			"indexname-2014-12",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			"indexname-%Y-%m-%d",
			"indexname-2014-12-01",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			"indexname-%Y-%m-%d-%H",
			"indexname-2014-12-01-23",
		},
		{
			time.Date(2014, 12, 01, 23, 30, 00, 00, time.UTC),
			"indexname-%y-%m",
			"indexname-14-12",
		},
	}
	for _, test := range tests {
		indexName := e.GetIndexName(test.IndexName, test.EventTime)
		if indexName != test.Expected {
			t.Errorf("Expected indexname %s, got %s\n", indexName, test.Expected)
		}
	}
}
