package main

import (
	"fmt"
	"os/exec"
	"log"
)

func main() {
	out, err := exec.Command("/bin/sh", "-c", "ps -A --no-headers | wc -l").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Number of running processes:    ", string(out))
}
