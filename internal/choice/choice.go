// Package choice provides basic functions for working with
// plugin options that must be one of several values.
package choice

import (
	"fmt"
	"slices"
)

// Check returns an error if a choice is not one of
// the available choices.
func Check(choice string, available []string) error {
	if !slices.Contains(available, choice) {
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
