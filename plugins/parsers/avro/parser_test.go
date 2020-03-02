package avro

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var DefaultTime = func() time.Time {
	return time.Unix(3600, 0)
}

func TestBasicAvroMessage(t *testing.T) {

	schema := `{"schema":"{\"type\":\"record\",\"name\":\"Value3\",\"namespace\":\"com.example.plant212\",\"fields\":[{\"name\":\"measurement\",\"type\":\"string\"},{\"name\":\"tag\",\"type\":\"string\"},{\"name\":\"field\",\"type\":\"long\"},{\"name\":\"timestamp\",\"type\":\"long\"}]}"}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(schema))
		fmt.Println("chiamato")
	}))
	defer ts.Close()

	p := Parser{
		SchemaRegistry:  ts.URL,
		Measurement:     "measurement",
		Tags:            []string{"tag"},
		Fields:          []string{"field"},
		Timestamp:       "timestamp",
		TimestampFormat: "unix",
		TimeFunc:        DefaultTime,
	}

	msg := []byte{0x00, 0x00, 0x00, 0x00, 0x17, 0x20, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x6d, 0x65, 0x61, 0x73, 0x75, 0x72, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x10, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x74, 0x61, 0x67, 0x26, 0xf0, 0xb6, 0x97, 0xd4, 0xb0, 0x5b}

	_, err := p.Parse(msg)

	require.NoError(t, err)
}
