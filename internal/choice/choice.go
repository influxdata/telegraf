// Package choice provides basic functions for working with
// plugin options that must be one of several values.
package choice

import "fmt"

// Contains return true if the choice in the list of choices.
func Contains(choice string, choices []string) bool {
	for _, item := range choices {
		if item == choice {
			return true
		}
	}
	return false
}

// Check returns an error if a choice is not one of
// the available choices.
func Check(choice string, available []string) error {
	if !Contains(choice, available) {
		return fmt.Errorf("unknown choice %s", choice)
	}
	return nil
}

// CheckSlice returns an error if the choices is not a subset of
// available.
func CheckSlice(choices, available []string) error {
	for _, choice := range choices {
		err := Check(choice, available)
		if err != nil {
			return err
		}
	}
	return nil
}
