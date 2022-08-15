//go:build !windows
// +build !windows

package main

func run(pprofErr <-chan error, inputFilters, outputFilters []string) error {
	stop = make(chan struct{})
	return reloadLoop(
		pprofErr,
		inputFilters,
		outputFilters,
	)
}
