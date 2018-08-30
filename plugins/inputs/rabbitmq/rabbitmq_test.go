package rabbitmq

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
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
    },
    "listeners": [
        {
            "name": "rabbit@node-a",
            "protocol": "amqp"
        },
        {
            "name": "rabbit@node-b",
            "protocol": "amqp"
        },
        {
            "name": "rabbit@node-a",
            "protocol": "clustering"
        },
        {
            "name": "rabbit@node-b",
            "protocol": "clustering"
        }
    ]
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
const sampleQueuesResponse = `
[
  {
    "memory": 21960,
    "messages": 0,
    "messages_details": {
      "rate": 0
    },
    "messages_ready": 0,
    "messages_ready_details": {
      "rate": 0
    },
    "messages_unacknowledged": 0,
    "messages_unacknowledged_details": {
      "rate": 0
    },
    "idle_since": "2015-11-01 8:22:15",
    "consumer_utilisation": "",
    "policy": "federator",
    "exclusive_consumer_tag": "",
    "consumers": 0,
    "recoverable_slaves": "",
    "state": "running",
    "messages_ram": 0,
    "messages_ready_ram": 0,
    "messages_unacknowledged_ram": 0,
    "messages_persistent": 0,
    "message_bytes": 0,
    "message_bytes_ready": 0,
    "message_bytes_unacknowledged": 0,
    "message_bytes_ram": 0,
    "message_bytes_persistent": 0,
    "disk_reads": 0,
    "disk_writes": 0,
    "backing_queue_status": {
      "q1": 0,
      "q2": 0,
      "delta": [
        "delta",
        "undefined",
        0,
        "undefined"
      ],
      "q3": 0,
      "q4": 0,
      "len": 0,
      "target_ram_count": "infinity",
      "next_seq_id": 0,
      "avg_ingress_rate": 0,
      "avg_egress_rate": 0,
      "avg_ack_ingress_rate": 0,
      "avg_ack_egress_rate": 0
    },
    "name": "collectd-queue",
    "vhost": "collectd",
    "durable": true,
    "auto_delete": false,
    "arguments": {},
    "node": "rabbit@testhost"
  },
  {
    "memory": 55528,
    "message_stats": {
    "ack": 223654927,
    "ack_details": {
      "rate": 0
    },
    "deliver": 224518745,
    "deliver_details": {
      "rate": 0
    },
    "deliver_get": 224518829,
    "deliver_get_details": {
      "rate": 0
    },
    "get": 19,
    "get_details": {
      "rate": 0
    },
    "get_no_ack": 65,
    "get_no_ack_details": {
      "rate": 0
    },
    "publish": 223883765,
    "publish_details": {
      "rate": 0
    },
    "redeliver": 863805,
    "redeliver_details": {
      "rate": 0
    }
    },
    "messages": 24,
    "messages_details": {
      "rate": 0
    },
    "messages_ready": 24,
    "messages_ready_details": {
      "rate": 0
    },
    "messages_unacknowledged": 0,
    "messages_unacknowledged_details": {
      "rate": 0
    },
    "idle_since": "2015-11-01 8:22:14",
    "consumer_utilisation": "",
    "policy": "",
    "exclusive_consumer_tag": "",
    "consumers": 0,
    "recoverable_slaves": "",
    "state": "running",
    "messages_ram": 24,
    "messages_ready_ram": 24,
    "messages_unacknowledged_ram": 0,
    "messages_persistent": 0,
    "message_bytes": 149220,
    "message_bytes_ready": 149220,
    "message_bytes_unacknowledged": 0,
    "message_bytes_ram": 149220,
    "message_bytes_persistent": 0,
    "disk_reads": 0,
    "disk_writes": 0,
    "backing_queue_status": {
      "q1": 0,
      "q2": 0,
      "delta": [
        "delta",
        "undefined",
        0,
        "undefined"
      ],
      "q3": 0,
      "q4": 24,
      "len": 24,
      "target_ram_count": "infinity",
      "next_seq_id": 223883765,
      "avg_ingress_rate": 0,
      "avg_egress_rate": 0,
      "avg_ack_ingress_rate": 0,
      "avg_ack_egress_rate": 0
    },
    "name": "telegraf",
    "vhost": "collectd",
    "durable": true,
    "auto_delete": false,
    "arguments": {},
    "node": "rabbit@testhost"
  },
  {
    "message_stats": {
      "ack": 1296077,
      "ack_details": {
        "rate": 0
      },
      "deliver": 1513176,
      "deliver_details": {
        "rate": 0.4
      },
      "deliver_get": 1513239,
      "deliver_get_details": {
        "rate": 0.4
      },
      "disk_writes": 7976,
      "disk_writes_details": {
        "rate": 0
      },
      "get": 40,
      "get_details": {
        "rate": 0
      },
      "get_no_ack": 23,
      "get_no_ack_details": {
        "rate": 0
      },
      "publish": 1325628,
      "publish_details": {
        "rate": 0.4
      },
      "redeliver": 216034,
      "redeliver_details": {
        "rate": 0
      }
    },
    "messages": 5,
    "messages_details": {
      "rate": 0.4
    },
    "messages_ready": 0,
    "messages_ready_details": {
      "rate": 0
    },
    "messages_unacknowledged": 5,
    "messages_unacknowledged_details": {
      "rate": 0.4
    },
    "policy": "federator",
    "exclusive_consumer_tag": "",
    "consumers": 1,
    "consumer_utilisation": 1,
    "memory": 122856,
    "recoverable_slaves": "",
    "state": "running",
    "messages_ram": 5,
    "messages_ready_ram": 0,
    "messages_unacknowledged_ram": 5,
    "messages_persistent": 0,
    "message_bytes": 150096,
    "message_bytes_ready": 0,
    "message_bytes_unacknowledged": 150096,
    "message_bytes_ram": 150096,
    "message_bytes_persistent": 0,
    "disk_reads": 0,
    "disk_writes": 7976,
    "backing_queue_status": {
      "q1": 0,
      "q2": 0,
      "delta": [
        "delta",
        "undefined",
        0,
        "undefined"
      ],
      "q3": 0,
      "q4": 0,
      "len": 0,
      "target_ram_count": "infinity",
      "next_seq_id": 1325628,
      "avg_ingress_rate": 0.19115840579934168,
      "avg_egress_rate": 0.19115840579934168,
      "avg_ack_ingress_rate": 0.19115840579934168,
      "avg_ack_egress_rate": 0.1492766485341716
    },
    "name": "telegraf",
    "vhost": "metrics",
    "durable": true,
    "auto_delete": false,
    "arguments": {},
    "node": "rabbit@testhost"
  }
]
`

const sampleExchangesResponse = `
[
  {
    "arguments": { },
    "internal": false,
    "auto_delete": false,
    "durable": true,
    "type": "direct",
    "vhost": "\/",
    "name": ""
  },
  {
    "message_stats": {
      "publish_in_details": {
        "rate": 0
      },
      "publish_in": 2,
      "publish_out_details": {
        "rate": 0
      },
      "publish_out": 1
    },
    "arguments": { },
    "internal": false,
    "auto_delete": false,
    "durable": true,
    "type": "fanout",
    "vhost": "\/",
    "name": "telegraf"
  },
  {
    "arguments": { },
    "internal": false,
    "auto_delete": false,
    "durable": true,
    "type": "direct",
    "vhost": "\/",
    "name": "amq.direct"
  },
  {
    "arguments": { },
    "internal": false,
    "auto_delete": false,
    "durable": true,
    "type": "fanout",
    "vhost": "\/",
    "name": "amq.fanout"
  },
  {
    "arguments": { },
    "internal": false,
    "auto_delete": false,
    "durable": true,
    "type": "headers",
    "vhost": "\/",
    "name": "amq.headers"
  },
  {
    "arguments": { },
    "internal": false,
    "auto_delete": false,
    "durable": true,
    "type": "headers",
    "vhost": "\/",
    "name": "amq.match"
  },
  {
    "arguments": { },
    "internal": true,
    "auto_delete": false,
    "durable": true,
    "type": "topic",
    "vhost": "\/",
    "name": "amq.rabbitmq.log"
  },
  {
    "arguments": { },
    "internal": true,
    "auto_delete": false,
    "durable": true,
    "type": "topic",
    "vhost": "\/",
    "name": "amq.rabbitmq.trace"
  },
  {
    "arguments": { },
    "internal": false,
    "auto_delete": false,
    "durable": true,
    "type": "topic",
    "vhost": "\/",
    "name": "amq.topic"
  }
]
`

func TestRabbitMQGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string

		switch r.URL.Path {
		case "/api/overview":
			rsp = sampleOverviewResponse
		case "/api/nodes":
			rsp = sampleNodesResponse
		case "/api/queues":
			rsp = sampleQueuesResponse
		case "/api/exchanges":
			rsp = sampleExchangesResponse
		default:
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
	defer ts.Close()

	r := &RabbitMQ{
		URL: ts.URL,
	}

	var acc testutil.Accumulator

	err := acc.GatherError(r.Gather)
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
		"clustering_listeners",
		"amqp_listeners",
	}

	for _, metric := range intMetrics {
		assert.True(t, acc.HasInt64Field("rabbitmq_overview", metric))
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
		"running",
	}

	for _, metric := range nodeIntMetrics {
		assert.True(t, acc.HasInt64Field("rabbitmq_node", metric))
	}

	assert.True(t, acc.HasMeasurement("rabbitmq_queue"))

	exchangeIntMetrics := []string{
		"messages_publish_in",
		"messages_publish_out",
	}

	for _, metric := range exchangeIntMetrics {
		assert.True(t, acc.HasInt64Field("rabbitmq_exchange", metric))
	}
}
