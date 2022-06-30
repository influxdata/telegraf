//go:build !windows
// +build !windows

package main

func run(inputFilters, outputFilters []string) {
	stop = make(chan struct{})
	reloadLoop(
		inputFilters,
		outputFilters,
	)
}
