package minecraft

import "testing"

func TestParseUsername(t *testing.T) {
	input := "1 tracked objective(s) for divislight:- jumps: 178 (jumps)"
	got := ParseUsername(input)
	want := "divislight"

	if got != want {
		t.Errorf("got %s\nwant %s\n", got, want)
	}
}
