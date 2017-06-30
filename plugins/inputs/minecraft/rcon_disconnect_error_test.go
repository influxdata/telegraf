package minecraft

import (
	"errors"
	"testing"
)

type MockRCONProducer struct {
	Err error
}

func (m *MockRCONProducer) newClient() (RCONClient, error) {
	return nil, m.Err
}

func TestRCONErrorHandling(t *testing.T) {
	m := &MockRCONProducer{
		Err: errors.New("Error: failed connection"),
	}
	c := &RCON{
		Server:   "craftstuff.com",
		Port:     "2222",
		Password: "pass",
		//Force fetching of new client
		client: nil,
	}

	_, err := c.Gather(m)
	if err == nil {
		t.Errorf("Error nil, unexpected result")
	}

	if c.client != nil {
		t.Fatal("c.client should be nil, unexpected result")
	}
}
