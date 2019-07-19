package kube_inventory

import (
	"testing"
	"time"

	"github.com/ericchiang/k8s/util/intstr"
	"github.com/influxdata/telegraf/internal/tls"
)

type mockHandler struct {
	responseMap map[string]interface{}
}

func toStrPtr(s string) *string {
	return &s
}

func toInt32Ptr(i int32) *int32 {
	return &i
}

func toInt64Ptr(i int64) *int64 {
	return &i
}

func toBoolPtr(b bool) *bool {
	return &b
}

func toIntStrPtrS(s string) *intstr.IntOrString {
	return &intstr.IntOrString{StrVal: &s}
}

func toIntStrPtrI(i int32) *intstr.IntOrString {
	return &intstr.IntOrString{IntVal: &i}
}
func TestNewClient(t *testing.T) {
	_, err := newClient("https://127.0.0.1:443/", "default", "abc123", time.Second, tls.ClientConfig{})
	if err != nil {
		t.Errorf("Failed to create new client - %s", err.Error())
	}
}
