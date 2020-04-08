package common

import (
	"reflect"
	"testing"
	"time"
)

func TestParseConnectionString(t *testing.T) {
	m, err := ParseConnectionString(
		"DeviceId=foo;SharedAccessKey=bar;SharedAccessKeyName=baz",
		"DeviceId",
		"SharedAccessKey",
	)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"DeviceId":            "foo",
		"SharedAccessKey":     "bar",
		"SharedAccessKeyName": "baz",
	}
	if !reflect.DeepEqual(want, m) {
		t.Fatalf("ParseConnectionString(s) = %v, want %v", m, want)
	}
}

func TestNewSharedAccessKey(t *testing.T) {
	sak := NewSharedAccessKey("test.azure-devices.net", "owner", "c2VjcmV0")
	if _, err := sak.Token(sak.HostName, time.Hour); err != nil {
		t.Fatal(err)
	}
}

func TestNewSharedAccessSignature(t *testing.T) {
	sas, err := NewSharedAccessSignature(
		"test.azure-devices.net",
		"owner",
		"c2VjcmV0",
		time.Date(2019, 1, 1, 1, 1, 1, 0, time.UTC).Add(time.Hour),
	)
	if err != nil {
		t.Fatal(err)
	}
	want := "SharedAccessSignature " +
		"sr=test.azure-devices.net" +
		"&sig=uux4PSAK1s9efNcBHin0pm7O5oENedTWkhqGp8JAyFY%3D" +
		"&se=1546308061" +
		"&skn=owner"
	if have := sas.String(); have != want {
		t.Fatalf("%#v.String() = %q, want %q", sas, have, want)
	}
}
