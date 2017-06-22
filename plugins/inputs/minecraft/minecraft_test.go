package minecraft

import (
	"fmt"
	"reflect"
	"testing"
)

// TestParsePlayerName tests different Minecraft RCON inputs for playerNames
func TestParsePlayerName(t *testing.T) {
	// Test a valid input string to ensure playerName is extracted
	input := "1 tracked objective(s) for divislight:- jumps: 178 (jumps)"
	got, err := ParsePlayerName(input)
	want := "divislight"
	if err != nil {
		t.Fatalf("playerName returned error. Error: %s\n", err)
	}
	if got != want {
		t.Errorf("got %s\nwant %s\n", got, want)
	}

	// Test an invalid input string to ensure error is returned
	input = ""
	got, err = ParsePlayerName(input)
	want = ""
	if err == nil {
		t.Fatal("Expected error when playerName not present. No error found.")
	}
	if got != want {
		t.Errorf("got %s\n want %s\n", got, want)
	}

	// Test an invalid input string to ensure error is returned
	input = "1 tracked objective(s) for 😂:- jumps: 178 (jumps)"
	got, err = ParsePlayerName(input)
	want = "😂"
	if err != nil {
		t.Fatalf("playerName returned error. Error: %s\n", err)
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

	fmt.Println(got)

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

	fmt.Println(got)

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
	got, err = ParseScoreboard(input)

	fmt.Println(got)

	if err == nil {
		t.Fatal("Expected input error, but error was nil")
	}

	// Tests when a number isn't an integer.
	input = `1 tracked objective(s) for divislight:- jumps: 178.5 (jumps)- sword: 5 (sword)`
	got, err = ParseScoreboard(input)

	fmt.Println(got)

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

	fmt.Println(got)

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
