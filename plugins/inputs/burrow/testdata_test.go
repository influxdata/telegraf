package burrow

import (
	"time"
)

type Expected struct {
	Tags   map[string]string
	Fields map[string]interface{}
	Time   time.Time
}

const clusterListResponse = `
{
  "error": false,
  "message": "cluster list returned",
  "clusters": [
    "clustername1",
    "clustername2"
  ],
  "request": {
    "uri": "/v2/kafka",
    "host": "responding.host.example.com",
    "cluster": "",
    "group": "",
    "topic": ""
  }
}
`

const topicListResponse = `
{
  "error": false,
  "message": "broker topic list returned",
  "topics": [
    "topicA",
    "topicB"
  ],
  "request": {
    "uri": "/v2/kafka/clustername/topic",
    "host": "responding.host.example.com",
    "cluster": "clustername",
    "group": "",
    "topic": ""
  }
}
`

const consumerListResponse = `
{
  "error": false,
  "message": "consumer list returned",
  "consumers": [
    "group1",
    "group2"
  ],
  "request": {
    "uri": "/v2/kafka/clustername/consumer",
    "host": "responding.host.example.com",
    "cluster": "clustername",
    "group": "",
    "topic": ""
  }
}
`

const topicDetailResponse = `
{
  "error": false,
  "message": "broker topic offsets returned",
  "offsets": [
    2290903,
    2898892,
    3902933,
    2328823
  ],
  "request": {
    "uri": "/v2/kafka/clustername/topic/topicname",
    "host": "responding.host.example.com",
    "cluster": "clustername",
    "group": "",
    "topic": "topicname"
  }
}
`

var topicDetailExpected = []*Expected{
	&Expected{
		Tags:   map[string]string{"cluster": "clustername1", "topic": "topicB", "partition": "0"},
		Fields: map[string]interface{}{"offset": int64(2290903)},
	},
	&Expected{
		Tags:   map[string]string{"cluster": "clustername1", "topic": "topicB", "partition": "1"},
		Fields: map[string]interface{}{"offset": int64(2898892)},
	},
	&Expected{
		Tags:   map[string]string{"cluster": "clustername1", "topic": "topicB", "partition": "2"},
		Fields: map[string]interface{}{"offset": int64(3902933)},
	},
	&Expected{
		Tags:   map[string]string{"cluster": "clustername1", "topic": "topicB", "partition": "3"},
		Fields: map[string]interface{}{"offset": int64(2328823)},
	},
}

const consumerStatusResponse = `
{
  "error": false,
  "message": "consumer group status returned",
  "status": {
    "cluster": "clustername1",
    "group": "groupname1",
    "status": "WARN",
    "complete": true,
    "maxlag": {
      "topic": "topicB",
      "partition": 0,
      "status": "WARN",
      "start": {
        "offset": 823889,
        "timestamp": 1432423256000,
        "lag": 20
      },
      "end": {
        "offset": 824743,
        "timestamp": 1432423796000,
        "lag": 25
      }
    },
    "partitions": [
      {
        "topic": "topicB",
        "partition": 0,
        "status": "WARN",
        "start": {
          "offset": 823889,
          "timestamp": 1432423256000,
          "lag": 20
        },
        "end": {
          "offset": 824743,
          "timestamp": 1432423796000,
          "lag": 25
        }
      }
    ]
  },
  "request": {
    "uri": "/v2/kafka/clustername/consumer/groupname/status",
    "host": "responding.host.example.com",
    "cluster": "clustername",
    "group": "groupname",
    "topic": ""
  }
}
`

var consumerStatusExpected = []*Expected{
	&Expected{
		Tags: map[string]string{
			"cluster":   "clustername1",
			"group":     "group1",
			"topic":     "topicB",
			"partition": "0",
			"status":    "WARN",
		},
		Fields: map[string]interface{}{"lag": int64(20), "offset": int64(823889)},
		Time:   time.Unix(1432423256, 0),
	},
	&Expected{
		Tags: map[string]string{
			"cluster":   "clustername1",
			"group":     "group1",
			"topic":     "topicB",
			"partition": "0",
			"status":    "WARN",
		},
		Fields: map[string]interface{}{"lag": int64(25), "offset": int64(824743)},
		Time:   time.Unix(1432423796, 0),
	},
}
