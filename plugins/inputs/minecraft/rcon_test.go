package minecraft

import (
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/minecraft/internal/rcon"
)

type MockRCONClient struct {
	Result *rcon.Packet
	Err    error
}

func (m *MockRCONClient) Authorize(password string) (*rcon.Packet, error) {
	return m.Result, m.Err
}
func (m *MockRCONClient) Execute(command string) (*rcon.Packet, error) {
	return m.Result, m.Err
}

// TestRCONGather test the RCON gather function
func TestRCONGather(t *testing.T) {
	mock := &MockRCONClient{
		Result: &rcon.Packet{
			Body: `Showing 1 tracked objective(s) for divislight:- jumps: 178 (jumps)Showing 7 tracked objective(s) for mauxlaim:- total_kills: 39 (total_kills)- "howdy doody": 37 (dalevel)- howdy: 37 (lvl)- jumps: 1290 (jumps)- iron_pickaxe: 284 (iron_pickaxe)- cow_kills: 1 (cow_kills)- "asdf": 37 (😂)Showing 5 tracked objective(s) for torham:- total_kills: 29 (total_kills)- "howdy doody": 33 (dalevel)- howdy: 33 (lvl)- jumps: 263 (jumps)- "asdf": 33 (😂)`,
		},
		Err: nil,
	}

	want := []string{
		` 1 tracked objective(s) for divislight:- jumps: 178 (jumps)`,
		` 7 tracked objective(s) for mauxlaim:- total_kills: 39 (total_kills)- "howdy doody": 37 (dalevel)- howdy: 37 (lvl)- jumps: 1290 (jumps)- iron_pickaxe: 284 (iron_pickaxe)- cow_kills: 1 (cow_kills)- "asdf": 37 (😂)`,
		` 5 tracked objective(s) for torham:- total_kills: 29 (total_kills)- "howdy doody": 33 (dalevel)- howdy: 33 (lvl)- jumps: 263 (jumps)- "asdf": 33 (😂)`,
	}

	client := &RCON{
		Server:   "craftstuff.com",
		Port:     "2222",
		Password: "pass",
		client:   mock,
	}

	d := defaultClientProducer{}
	got, err := client.Gather(d)
	if err != nil {
		t.Fatalf("Gather returned an error. Error %s\n", err)
	}
	for i, s := range got {
		if want[i] != s {
			t.Fatalf("Got %s at index %d, want %s at index %d", s, i, want[i], i)
		}
	}

	client.client = &MockRCONClient{
		Result: &rcon.Packet{
			Body: "",
		},
		Err: nil,
	}

	got, err = client.Gather(defaultClientProducer{})
	if err != nil {
		t.Fatalf("Gather returned an error. Error %s\n", err)
	}
	if len(got) != 0 {
		t.Fatalf("Expected empty slice of length %d, got slice of length %d", 0, len(got))
	}
}
