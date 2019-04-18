package kafkabeat

const kafkabeatInfo = `
{
  "beat": "kafkabeat",
  "hostname": "node-6",
  "name": "node-6-test",
  "uuid": "9c1c8697-acb4-4df0-987d-28197814f785",
  "version": "6.6.2"
}
`

const kafkabeatStats = `
{
  "beat": {
    "cpu": {
      "system": {
        "ticks": 1393890,
        "time": {
          "ms": 1393890
        }
      },
      "total": {
        "ticks": 52339260,
        "time": {
          "ms": 52339264
        },
        "value": 52339260
      },
      "user": {
        "ticks": 50945370,
        "time": {
          "ms": 50945374
        }
      }
    },
    "info": {
      "ephemeral_id": "7bc722be-e7df-4c07-ae4c-f7f82b23ffea",
      "uptime": {
        "ms": 65057537
      }
    },
    "memstats": {
      "gc_next": 559016128,
      "memory_alloc": 280509808,
      "memory_total": 4596157344344,
      "rss": 368422912
    }
  },
  "libbeat": {
    "config": {
      "module": {
        "running": 0,
        "starts": 0,
        "stops": 0
      },
      "reloads": 0
    },
    "output": {
      "events": {
        "acked": 186307311,
        "active": 0,
        "batches": 1753223,
        "dropped": 0,
        "duplicates": 0,
        "failed": 0,
        "total": 186307311
      },
      "read": {
        "bytes": 1248297178,
        "errors": 0
      },
      "type": "elasticsearch",
      "write": {
        "bytes": 60016355484,
        "errors": 0
      }
    },
    "pipeline": {
      "clients": 1,
      "events": {
        "active": 0,
        "dropped": 0,
        "failed": 0,
        "filtered": 0,
        "published": 186307311,
        "retry": 106,
        "total": 186307311
      },
      "queue": {
        "acked": 186307311
      }
    }
  },
  "system": {
    "cpu": {
      "cores": 32
    },
    "load": {
      "1": 10.76,
      "15": 7.19,
      "5": 10.7,
      "norm": {
        "1": 0.3363,
        "15": 0.2247,
        "5": 0.3344
      }
    }
  }
}
`
