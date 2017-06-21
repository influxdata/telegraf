package main

import (
	"fmt"
	"log"
	"regexp"
)

func main() {
	r := RCON{
		Host:   "35.185.198.170",
		Port:   "25575",
		Passwd: "influx123",
	}

	results, err := r.Gather()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Full results", results)

	var newSlice []string
	if len(results) > 1 {
		newSlice = results[1:]
	} else {
		// TODO: newSlice
		return
	}

	var re = regexp.MustCompile(`for\s(.*):-`)

	for _, s := range newSlice {
		fmt.Printf("original string: %s\n", s)
		for i, match := range re.FindAllStringSubmatch(s, -1) {
			fmt.Println(match[1], "found at index", i, "\n")
		}
	}

	// NEXT STEPS
	// use regex to isolate stat name
	// use regex to isolate stats themselves
	// structure all data into JSON output format
	// use that data somehow

	// Result of running command stored in packet.Body
	/*
			var re = regexp.MustCompile(`for\s(.*):-`)
		    var str = `4 tracked objective(s) for mauxlaim:- total_kills: 2 (total_kills)- iron_pickaxe: 190 (iron_pickaxe)- jumps: 866 (jumps)- level: 35 (level),`

		    for i, match := range re.FindAllString(str, -1) {
		        fmt.Println(match, "found at index", i)
		    }
	*/
}
