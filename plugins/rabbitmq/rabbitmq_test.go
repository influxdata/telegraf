package rabbitmq

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/koksan83/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleOverviewResponse = `
{
	"message_stats": {
        "ack": 5246,
        "ack_details": {
            "rate": 0.0
        },
        "deliver": 5246,
        "deliver_details": {
            "rate": 0.0
        },
        "deliver_get": 5246,
        "deliver_get_details": {
            "rate": 0.0
        },
        "publish": 5258,
        "publish_details": {
            "rate": 0.0
        }
    },
    "object_totals": {
        "channels": 44,
        "connections": 44,
        "consumers": 65,
        "exchanges": 43,
        "queues": 62
    },
    "queue_totals": {
        "messages": 0,
        "messages_details": {
            "rate": 0.0
        },
        "messages_ready": 0,
        "messages_ready_details": {
            "rate": 0.0
        },
        "messages_unacknowledged": 0,
        "messages_unacknowledged_details": {
            "rate": 0.0
        }
    }
}
`

const sampleNodesResponse = `
[
    {
        "db_dir": "/var/lib/rabbitmq/mnesia/rabbit@vagrant-ubuntu-trusty-64",
        "disk_free": 37768282112,
        "disk_free_alarm": false,
        "disk_free_details": {
            "rate": 0.0
        },
        "disk_free_limit": 50000000,
        "enabled_plugins": [
            "rabbitmq_management"
        ],
        "fd_total": 1024,
        "fd_used": 63,
        "fd_used_details": {
            "rate": 0.0
        },
        "io_read_avg_time": 0,
        "io_read_avg_time_details": {
            "rate": 0.0
        },
        "io_read_bytes": 1,
        "io_read_bytes_details": {
            "rate": 0.0
        },
        "io_read_count": 1,
        "io_read_count_details": {
            "rate": 0.0
        },
        "io_sync_avg_time": 0,
        "io_sync_avg_time_details": {
            "rate": 0.0
        },
        "io_write_avg_time": 0,
        "io_write_avg_time_details": {
            "rate": 0.0
        },
        "log_file": "/var/log/rabbitmq/rabbit@vagrant-ubuntu-trusty-64.log",
        "mem_alarm": false,
        "mem_limit": 2503771750,
        "mem_used": 159707080,
        "mem_used_details": {
            "rate": 15185.6
        },
        "mnesia_disk_tx_count": 16,
        "mnesia_disk_tx_count_details": {
            "rate": 0.0
        },
        "mnesia_ram_tx_count": 296,
        "mnesia_ram_tx_count_details": {
            "rate": 0.0
        },
        "name": "rabbit@vagrant-ubuntu-trusty-64",
        "net_ticktime": 60,
        "os_pid": "14244",
        "partitions": [],
        "proc_total": 1048576,
        "proc_used": 783,
        "proc_used_details": {
            "rate": 0.0
        },
        "processors": 1,
        "rates_mode": "basic",
        "run_queue": 0,
        "running": true,
        "sasl_log_file": "/var/log/rabbitmq/rabbit@vagrant-ubuntu-trusty-64-sasl.log",
        "sockets_total": 829,
        "sockets_used": 45,
        "sockets_used_details": {
            "rate": 0.0
        },
        "type": "disc",
        "uptime": 7464827
    }
]
`

func TestRabbitMQGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string

		if r.URL.Path == "/api/overview" {
			rsp = sampleOverviewResponse
		} else if r.URL.Path == "/api/nodes" {
			rsp = sampleNodesResponse
		} else {
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
	defer ts.Close()

	r := &RabbitMQ{
		Servers: []*Server{
			{
				URL: ts.URL,
			},
		},
	}

	var acc testutil.Accumulator

	err := r.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{
		"messages",
		"messages_ready",
		"messages_unacked",

		"messages_acked",
		"messages_delivered",
		"messages_published",

		"channels",
		"connections",
		"consumers",
		"exchanges",
		"queues",
	}

	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntValue(metric))
	}

	nodeIntMetrics := []string{
		"disk_free",
		"disk_free_limit",
		"fd_total",
		"fd_used",
		"mem_limit",
		"mem_used",
		"proc_total",
		"proc_used",
		"run_queue",
		"sockets_total",
		"sockets_used",
	}

	for _, metric := range nodeIntMetrics {
		assert.True(t, acc.HasIntValue(metric))
	}
}
