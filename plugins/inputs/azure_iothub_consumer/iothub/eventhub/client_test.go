package eventhub

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestParseConnectionString(t *testing.T) {
	have, err := ParseConnectionString(
		"Endpoint=sb://namespace.windows.net/;" +
			"SharedAccessKeyName=policy-name;" +
			"SharedAccessKey=abcNg==;" +
			"EntityPath=hub-name",
	)
	if err != nil {
		t.Fatal(err)
	}

	want := &Credentials{
		Endpoint:            "namespace.windows.net",
		SharedAccessKeyName: "policy-name",
		SharedAccessKey:     "abcNg==",
		EntityPath:          "hub-name",
	}
	if *want != *have {
		t.Fatalf("ParseConnectionString = %#v, want %#v", have, want)
	}
}

func TestClient_Subscribe(t *testing.T) {
	cs := os.Getenv("TEST_EVENTHUB_CONNECTION_STRING")
	if cs == "" {
		t.Fatal("$TEST_EVENTHUB_CONNECTION_STRING is empty")
	}
	c, err := DialConnectionString(cs)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// subscribe to the eventhub for a second to validate that we don't get any errors
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := c.Subscribe(ctx, func(msg *Event) error {
		return msg.Accept()
	},
		WithSubscribeSince(time.Now()),
	); err != nil && err != context.DeadlineExceeded {
		t.Fatal(err)
	}
}
