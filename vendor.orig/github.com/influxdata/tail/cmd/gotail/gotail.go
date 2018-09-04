// Copyright (c) 2015 HPE Software Inc. All rights reserved.
// Copyright (c) 2013 ActiveState Software Inc. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/influxdata/tail"
)

func args2config() (tail.Config, int64) {
	config := tail.Config{Follow: true}
	n := int64(0)
	maxlinesize := int(0)
	flag.Int64Var(&n, "n", 0, "tail from the last Nth location")
	flag.IntVar(&maxlinesize, "max", 0, "max line size")
	flag.BoolVar(&config.Follow, "f", false, "wait for additional data to be appended to the file")
	flag.BoolVar(&config.ReOpen, "F", false, "follow, and track file rename/rotation")
	flag.BoolVar(&config.Poll, "p", false, "use polling, instead of inotify")
	flag.Parse()
	if config.ReOpen {
		config.Follow = true
	}
	config.MaxLineSize = maxlinesize
	return config, n
}

func main() {
	config, n := args2config()
	if flag.NFlag() < 1 {
		fmt.Println("need one or more files as arguments")
		os.Exit(1)
	}

	if n != 0 {
		config.Location = &tail.SeekInfo{-n, os.SEEK_END}
	}

	done := make(chan bool)
	for _, filename := range flag.Args() {
		go tailFile(filename, config, done)
	}

	for _, _ = range flag.Args() {
		<-done
	}
}

func tailFile(filename string, config tail.Config, done chan bool) {
	defer func() { done <- true }()
	t, err := tail.TailFile(filename, config)
	if err != nil {
		fmt.Println(err)
		return
	}
	for line := range t.Lines {
		fmt.Println(line.Text)
	}
	err = t.Wait()
	if err != nil {
		fmt.Println(err)
	}
}
