package mqtt

import (
	"reflect"
	"testing"
)

func TestParseCloudToDeviceTopic(t *testing.T) {
	s := "devices/mydev/messages/devicebound/%24.to=%2Fdevices%2Fmydev%2Fmessages%2FdeviceBound&a[]=b&b=c"
	g, err := parseCloudToDeviceTopic(s)
	if err != nil {
		t.Fatal(err)
	}

	w := map[string]string{
		"$.to": "/devices/mydev/messages/deviceBound",
		"a[]":  "b",
		"b":    "c",
	}
	if !reflect.DeepEqual(g, w) {
		t.Errorf("parseCloudToDeviceTopic(%q) = %v, _, want %v", s, g, w)
	}
}

func TestParseDirectMethodTopic(t *testing.T) {
	s := "$iothub/methods/POST/add/?$rid=666"
	m, r, err := parseDirectMethodTopic(s)
	if err != nil {
		t.Fatal(err)
	}
	if m != "add" || r != 666 {
		t.Errorf("parseDirectMethodTopic(%q) = %q, %q, want %q, %q", s, m, r, "add", 666)
	}
}

func TestParseTwinPropsTopic(t *testing.T) {
	s := "$iothub/twin/res/200/?$rid=12&$version=4"
	c, r, v, err := parseTwinPropsTopic(s)
	if err != nil {
		t.Fatal(err)
	}
	if c != 200 || r != 12 || v != 4 {
		t.Errorf("ParseTwinPropsTopic(%q) = %d, %q, %d, _, want %d, %q, %d, _", s, c, r, v, 200, 12, 4)
	}
}
