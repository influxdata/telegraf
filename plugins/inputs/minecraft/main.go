package main

import (
	"fmt"
	"log"
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
}
