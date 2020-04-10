package iotservice

import (
	"reflect"
	"testing"
	"time"

	"github.com/amenzhinsky/iothub/common"
)

func TestToFromAMQPMessage(t *testing.T) {
	now := time.Now()
	want := &common.Message{
		MessageID:     "1",
		To:            "azure",
		ExpiryTime:    &now,
		CorrelationID: "id",
		UserID:        "admin",
		Properties:    map[string]string{"k": "v"},
		Payload:       []byte("hello"),
	}
	if have := FromAMQPMessage(toAMQPMessage(want)); !reflect.DeepEqual(have, want) {
		t.Fatalf("FromAMQPMessage(toAMQPMessage(want)) = %v, want = %v", have, want)
	}
}
