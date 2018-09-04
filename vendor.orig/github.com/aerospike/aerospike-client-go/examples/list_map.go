/*
 * Copyright 2012-2016 Aerospike, Inc.
 *
 * Portions may be licensed to Aerospike, Inc. under one or more contributor
 * license agreements.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 * use this file except in compliance with the License. You may obtain a copy of
 * the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations under
 * the License.
 */

package main

import (
	"bytes"
	"log"

	as "github.com/aerospike/aerospike-client-go"
	shared "github.com/aerospike/aerospike-client-go/examples/shared"
)

func main() {
	testListStrings(shared.Client)
	testListComplex(shared.Client)
	testMapStrings(shared.Client)
	testMapComplex(shared.Client)
	testListMapCombined(shared.Client)

	log.Println("Example finished successfully.")
}

/**
 * Write/Read []string directly instead of relying on java serializer.
 */
func testListStrings(client *as.Client) {
	log.Printf("Read/Write []string")
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "listkey1")
	client.Delete(shared.WritePolicy, key)

	list := []string{"string1", "string2", "string3"}

	bin := as.NewBin("listbin1", list)
	client.PutBins(shared.WritePolicy, key, bin)

	record, err := client.Get(shared.Policy, key, bin.Name)
	shared.PanicOnError(err)
	receivedList := record.Bins[bin.Name].([]interface{})

	validateSize(3, len(receivedList))
	validate("string1", receivedList[0])
	validate("string2", receivedList[1])
	validate("string3", receivedList[2])

	log.Printf("Read/Write []string successful.")
}

/**
 * Write/Read []interface{} directly instead of relying on java serializer.
 */
func testListComplex(client *as.Client) {
	log.Printf("Read/Write []interface{}")
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "listkey2")
	client.Delete(shared.WritePolicy, key)

	blob := []byte{3, 52, 125}
	list := []interface{}{"string1", 2, blob}

	bin := as.NewBin("listbin2", list)
	client.PutBins(shared.WritePolicy, key, bin)

	record, err := client.Get(shared.Policy, key, bin.Name)
	shared.PanicOnError(err)
	receivedList := record.Bins[bin.Name].([]interface{})

	validateSize(3, len(receivedList))
	validate("string1", receivedList[0])
	// Server convert numbers to long, so must expect long.
	validate(2, receivedList[1])
	validateBytes(blob, receivedList[2].([]byte))

	log.Printf("Read/Write []interface{} successful.")
}

/**
 * Write/Read map[string]string directly instead of relying on java serializer.
 */
func testMapStrings(client *as.Client) {
	log.Printf("Read/Write map[string]string")
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "mapkey1")
	client.Delete(shared.WritePolicy, key)

	amap := map[string]string{"key1": "string1",
		"key2": "string2",
		"key3": "string3",
	}
	bin := as.NewBin("mapbin1", amap)
	client.PutBins(shared.WritePolicy, key, bin)

	record, err := client.Get(shared.Policy, key, bin.Name)
	shared.PanicOnError(err)
	receivedMap := record.Bins[bin.Name].(map[interface{}]interface{})

	validateSize(3, len(receivedMap))
	validate("string1", receivedMap["key1"])
	validate("string2", receivedMap["key2"])
	validate("string3", receivedMap["key3"])

	log.Printf("Read/Write map[string]string successful")
}

/**
 * Write/Read map[interface{}]interface{} directly instead of relying on java serializer.
 */
func testMapComplex(client *as.Client) {
	log.Printf("Read/Write map[interface{}]interface{}")
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "mapkey2")
	client.Delete(shared.WritePolicy, key)

	blob := []byte{3, 52, 125}
	list := []int{
		100034,
		12384955,
		3,
		512,
	}

	amap := map[interface{}]interface{}{
		"key1": "string1",
		"key2": 2,
		"key3": blob,
		"key4": list,
	}

	bin := as.NewBin("mapbin2", amap)
	client.PutBins(shared.WritePolicy, key, bin)

	record, err := client.Get(shared.Policy, key, bin.Name)
	shared.PanicOnError(err)
	receivedMap := record.Bins[bin.Name].(map[interface{}]interface{})

	validateSize(4, len(receivedMap))
	validate("string1", receivedMap["key1"])
	// Server convert numbers to long, so must expect long.
	validate(2, receivedMap["key2"])
	validateBytes(blob, receivedMap["key3"].([]byte))

	receivedInner := receivedMap["key4"].([]interface{})
	validateSize(4, len(receivedInner))
	validate(100034, receivedInner[0])
	validate(12384955, receivedInner[1])
	validate(3, receivedInner[2])
	validate(512, receivedInner[3])

	log.Printf("Read/Write map[interface{}]interface{} successful")
}

/**
 * Write/Read Array/Map combination directly instead of relying on java serializer.
 */
func testListMapCombined(client *as.Client) {
	log.Printf("Read/Write Array/Map")
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "listmapkey")
	client.Delete(shared.WritePolicy, key)

	blob := []byte{3, 52, 125}
	inner := []interface{}{
		"string2",
		5,
	}

	innerMap := map[interface{}]interface{}{
		"a":    1,
		2:      "b",
		3:      blob,
		"list": inner,
	}

	list := []interface{}{
		"string1",
		8,
		inner,
		innerMap,
	}

	bin := as.NewBin("listmapbin", list)
	client.PutBins(shared.WritePolicy, key, bin)

	record, err := client.Get(shared.Policy, key, bin.Name)
	shared.PanicOnError(err)
	received := record.Bins[bin.Name].([]interface{})

	validateSize(4, len(received))
	validate("string1", received[0])
	// Server convert numbers to long, so must expect long.
	validate(8, received[1])

	receivedInner := received[2].([]interface{})
	validateSize(2, len(receivedInner))
	validate("string2", receivedInner[0])
	validate(5, receivedInner[1])

	receivedMap := received[3].(map[interface{}]interface{})
	validateSize(4, len(receivedMap))
	validate(1, receivedMap["a"])
	validate("b", receivedMap[2])
	validateBytes(blob, receivedMap[3].([]byte))

	receivedInner2 := receivedMap["list"].([]interface{})
	validateSize(2, len(receivedInner2))
	validate("string2", receivedInner2[0])
	validate(5, receivedInner2[1])

	log.Printf("Read/Write Array/Map successful")
}

func validateSize(expected, received int) {
	if received != expected {
		log.Fatalf(
			"Size mismatch: expected=%d received=%d", expected, received)
	}
}

func validate(expected, received interface{}) {
	if !(received == expected) {
		log.Fatalf(
			"Mismatch: expected=%v received=%v", expected, received)
	}
}

func validateBytes(expected []byte, received []byte) {
	if !bytes.Equal(expected, received) {
		log.Fatalf(
			"Mismatch: expected=%v received=%v", expected, received)
	}
}
