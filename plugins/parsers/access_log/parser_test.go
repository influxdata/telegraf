package access_log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAccessLog(t *testing.T) {
	parser := AccessLogParser{
		MetricName: "access_log",
		DataType:   "string",
	}
	metrics, err := parser.Parse([]byte("0.806	0.553	192.168.165.89	2328	10.10.101.97	[2016-11-28T19:33:12+08:00]	api.feature-2274.test.angejia.com	PUT /common/users/actions/login HTTP/1.1	200	1997	-	okhttp/3.4.2	1.74	-	zaiTihyVyxUQA07TiPe4UMy8tUbD4MUnvZ7cUx94qq5K0d/WeyhnP9YyqYYeZyw4Q+v4m2GpRodnNU1hLPb8b+RCA1ihaX1hkF+lIoXvuVe2w+ezJVr1n/7/Hg797DQOSYAeNUEd27+QXXiLtboleq/P2Uf3K0uzcWP19Q3S1O0ObMs8JRBYwr36jlP1/KpzZkeyVX5bGi8nz24lXlZl3Ya3oaELZFKQ18vsbzslAxwYYKLeQ6KgHYxRLsOh8Vfo+ao7nBq2ubnLeYDwC411guDbGibDMgNBQ8pNX4LS4sp9lT9U3WgJJA7EIi5Ipt0NcPVnsuik09ZJx8aC2am1qBakvvBmqZVUvVhG/RP0d2x+C0lqkPkaSxJCrnOdHSEjnYpUA6muJXrm5VeMz12hRMqg/sv2UeV8urHU7h7zo1u+4CnB4f1ZHCD6QH3RRkc10Ob2KErv2gKnK+/q0+xvvQmhnY2ubNCMDXR36XX+7pqqjMU6IJvmEbnwDpcKxKRN1jPnGRLTIF/KLYmXdW2YUrbHuccAwVMLDksKVMzMVXuv+kw03/5g6na8L9dnt9ZeKxoqxhpdM16MPLZ9TjXFrthISmzUc2KMpf5JbYafGu632FKvsq4vhdlDcR6xH7G75bqNmzgVHOKe19Eqi+cNZ4rdMZYjHMMPY/Llnj4a4m5p/rqfsvpoW/WwIfKVsEpMGKZjqt/0u5VYS7PcPWOdBQwRtp0tzkihU6aXySrOTGJY/ykQ4Xjq+fj3P+N6FzucM6UaZnrYIB0UxEE1b97rK6b8fw20HuBcgD/3MGiZNYj7EnTph5lqrwgkczu92I8HKzCa0Zc5f1PRvmYNsyVugvBpZdDsCIM54ud5p6ybxaI=	app=a-broker;av=6.0.0;ccid=2;gcid=2;ch=C08;lng=121.393033;lat=31.170153;net=WIFI;p=android;pm=Android-SM-N9200;osv=6.0.1;dvid=35257507109530602:00:00:00:00:00;uid=46836-	-	HTTP/1.1	-"))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "access_log", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"hostname": "api.feature-2274.test.angejia.com",
		"method":   "PUT",
		"path":     "/common/users/actions/login",
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())
}
