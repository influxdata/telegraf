package crypto

import (
	"log"
	"strings"
	"time"
)

const (
	networkTimeout = 5 // seconds
)

type algorithm int

// algorithm constants the order and id is the same as for nicehash
//go:generate stringer -type=algorithm
const (
	scrypt algorithm = iota
	sha256
	scryptnf
	x11
	x13
	keccak
	x15
	nist5
	neoscrypt
	lyra2re
	whirlpoolx
	qubit
	quark
	axiom
	lyra2rev2
	scryptjanenf16
	blake256r8
	blake256r14
	blake256r8vnl
	hodl
	ethash
	decred
	cryptonight
	lbry
	equihash
	pascal
	x11gost
	sia
	blake2s
	skunk
	cryptonightv7
	daggerhashimoto = ethash
	blake2b         = sia
	zcash           = equihash
)

//go:generate stringer -type=sourceType
type sourceType int

// sourceType enum
const (
	ACCOUNT sourceType = iota
	MINER
	GPU
	FAN
	CHAIN
	THREAD
	POOL
)

func unitMultilier(unit string) int64 {
	if len(unit) == 0 {
		return 0
	}
	// drop /s or /x part
	unit = strings.ToLower(strings.Split(unit, "/")[0])

	if strings.HasPrefix(unit, "k") { // kilo
		return 1000
	}
	if strings.HasPrefix(unit, "m") { // mega
		return 1000 * 1000
	}
	if strings.HasPrefix(unit, "g") { //giga
		return 1000 * 1000 * 1000
	}
	if strings.HasPrefix(unit, "t") { //tera
		return 1000 * 1000 * 1000 * 1000
	}
	if strings.HasPrefix(unit, "p") { //peta
		return 1000 * 1000 * 1000 * 1000 * 1000
	}
	if strings.HasPrefix(unit, "e") { //exa
		return 1000 * 1000 * 1000 * 1000 * 1000 * 1000
	}
	// if strings.HasPrefix(unit, "z") { //zeta
	// 	return 1000 * 1000 * 1000 * 1000 * 1000 * 1000 * 1000
	// }
	// if strings.HasPrefix(unit, "y") { //yotta
	// 	return 1000 * 1000 * 1000 * 1000 * 1000 * 1000 * 1000 * 1000
	// }
	return 1
}

// use eg:
// defer timeTrack(time.Now(), "function")
func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
