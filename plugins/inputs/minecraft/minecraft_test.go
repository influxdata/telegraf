package minecraft

import "testing"

// TestParseUsername tests different Minecraft RCON inputs for usernames
func TestParseUsername(t *testing.T) {
	// Test a valid input string to ensure username is extracted
	input := "1 tracked objective(s) for divislight:- jumps: 178 (jumps)"
	got, err := ParseUsername(input)
	want := "divislight"
	if err != nil {
		t.Fatalf("username returned error. Error: %s\n", err)
	}
	if got != want {
		t.Errorf("got %s\nwant %s\n", got, want)
	}

	// Test an invalid input string to ensure error is returned
	input = ""
	got, err = ParseUsername(input)
	want = ""
	if err == nil {
		t.Fatal("Expected error when username not present. No error found.")
	}
	if got != want {
		t.Errorf("got %s\n want %s\n", got, want)
	}
}
