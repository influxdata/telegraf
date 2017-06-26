package minecraft

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

// TestParsePlayerName tests different Minecraft RCON inputs for players
func TestParsePlayerName(t *testing.T) {
	// Test a valid input string to ensure player is extracted
	input := "1 tracked objective(s) for divislight:- jumps: 178 (jumps)"
	got, err := ParsePlayerName(input)
	want := "divislight"
	if err != nil {
		t.Fatalf("player returned error. Error: %s\n", err)
	}
	if got != want {
		t.Errorf("got %s\nwant %s\n", got, want)
	}

	// Test an invalid input string to ensure error is returned
	input = ""
	got, err = ParsePlayerName(input)
	want = ""
	if err == nil {
		t.Fatal("Expected error when player not present. No error found.")
	}
	if got != want {
		t.Errorf("got %s\n want %s\n", got, want)
	}

	// Test an invalid input string to ensure error is returned
	input = "1 tracked objective(s) for 😂:- jumps: 178 (jumps)"
	got, err = ParsePlayerName(input)
	want = "😂"
	if err != nil {
		t.Fatalf("player returned error. Error: %s\n", err)
	}
	if got != want {
		t.Errorf("got %s\n want %s\n", got, want)
	}
}

// TestParseScoreboard tests different Minecraft RCON inputs for scoreboard stats.
func TestParseScoreboard(t *testing.T) {
	// test a valid input string to ensure stats are parsed correctly.
	input := `1 tracked objective(s) for divislight:- jumps: 178 (jumps)- sword: 5 (sword)`
	got, err := ParseScoreboard(input)
	if err != nil {
		t.Fatal("Unexpected error")
	}

	want := []Score{
		{
			Name:  "jumps",
			Value: 178,
		},
		{
			Name:  "sword",
			Value: 5,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got: \n%#v\nWant: %#v", got, want)
	}

	// Tests a partial input string.
	input = `1 tracked objective(s) for divislight:- jumps: (jumps)- sword: 5 (sword)`
	got, err = ParseScoreboard(input)

	if err != nil {
		t.Fatal("Unexpected error")
	}

	want = []Score{
		{
			Name:  "sword",
			Value: 5,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got: \n%#v\nWant:\n%#v", got, want)
	}

	// Tests an empty string.
	input = ``
	_, err = ParseScoreboard(input)
	if err == nil {
		t.Fatal("Expected input error, but error was nil")
	}

	// Tests when a number isn't an integer.
	input = `1 tracked objective(s) for divislight:- jumps: 178.5 (jumps)- sword: 5 (sword)`
	got, err = ParseScoreboard(input)
	if err != nil {
		t.Fatal("Unexpected error")
	}

	want = []Score{
		{
			Name:  "sword",
			Value: 5,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got: \n%#v\nWant: %#v", got, want)
	}

	//Testing a real life data scenario with unicode characters
	input = `7 tracked objective(s) for mauxlaim:- total_kills: 39 (total_kills)- "howdy doody": 37 (dalevel)- howdy: 37 (lvl)- jumps: 1290 (jumps)- iron_pickaxe: 284 (iron_pickaxe)- cow_kills: 1 (cow_kills)- "asdf": 37 (😂)`
	got, err = ParseScoreboard(input)
	if err != nil {
		t.Fatal("Unexpected error")
	}

	want = []Score{
		{
			Name:  "total_kills",
			Value: 39,
		},
		{
			Name:  "dalevel",
			Value: 37,
		},
		{
			Name:  "lvl",
			Value: 37,
		},
		{
			Name:  "jumps",
			Value: 1290,
		},
		{
			Name:  "iron_pickaxe",
			Value: 284,
		},
		{
			Name:  "cow_kills",
			Value: 1,
		},
		{
			Name:  "😂",
			Value: 37,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got: \n%#v\nWant: %#v", got, want)
	}

}

type MockClient struct {
	Result []string
	Err    error
}

func (m *MockClient) Gather(d RCONClientProducer) ([]string, error) {
	return m.Result, m.Err
}

func TestGather(t *testing.T) {
	var acc testutil.Accumulator
	testConfig := Minecraft{
		Server: "biffsgang.net",
		Port:   "25575",
		client: &MockClient{
			Result: []string{
				`1 tracked objective(s) for divislight:- jumps: 178 (jumps)`,
				`7 tracked objective(s) for mauxlaim:- total_kills: 39 (total_kills)- "howdy doody": 37 (dalevel)- howdy: 37 (lvl)- jumps: 1290 (jumps)- iron_pickaxe: 284 (iron_pickaxe)- cow_kills: 1 (cow_kills)- "asdf": 37 (😂)`,
				`5 tracked objective(s) for torham:- total_kills: 29 (total_kills)- "howdy doody": 33 (dalevel)- howdy: 33 (lvl)- jumps: 263 (jumps)- "asdf": 33 (😂)`,
			},
			Err: nil,
		},
		clientSet: true,
	}

	err := testConfig.Gather(&acc)

	if err != nil {
		t.Fatalf("gather returned error. Error: %s\n", err)
	}

	if !testConfig.clientSet {
		t.Fatalf("clientSet should be true, client should be set")
	}

	tags := map[string]string{
		"player": "divislight",
		"server": "biffsgang.net:25575",
	}

	assertContainsTaggedStat(t, &acc, "minecraft", "jumps", 178, tags)
	tags["player"] = "mauxlaim"
	assertContainsTaggedStat(t, &acc, "minecraft", "cow_kills", 1, tags)
	tags["player"] = "torham"
	assertContainsTaggedStat(t, &acc, "minecraft", "total_kills", 29, tags)

}

func assertContainsTaggedStat(
	t *testing.T,
	acc *testutil.Accumulator,
	measurement string,
	field string,
	expectedValue int,
	tags map[string]string,
) {
	var actualValue int
	for _, pt := range acc.Metrics {
		if pt.Measurement == measurement && reflect.DeepEqual(pt.Tags, tags) {
			for fieldname, value := range pt.Fields {
				if fieldname == field {
					actualValue = value.(int)
					if value == expectedValue {
						return
					}
					t.Errorf("Expected value %d\n got value %d\n", expectedValue, value)
				}
			}
		}
	}
	msg := fmt.Sprintf(
		"Could not find measurement \"%s\" with requested tags within %s, Actual: %d",
		measurement, field, actualValue)
	t.Fatal(msg)

}
