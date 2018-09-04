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

const udf = `
local function putBin(r,name,value)
    if not aerospike:exists(r) then aerospike:create(r) end
    r[name] = value
    aerospike:update(r)
end

-- Set a particular bin
function writeBin(r,name,value)
    putBin(r,name,value)
end

-- Get a particular bin
function readBin(r,name)
    return r[name]
end

-- Return generation count of record
function getGeneration(r)
    return record.gen(r)
end

-- Update record only if gen hasn't changed
function writeIfGenerationNotChanged(r,name,value,gen)
    if record.gen(r) == gen then
        r[name] = value
        aerospike:update(r)
    end
end

-- Set a particular bin only if record does not already exist.
function writeUnique(r,name,value)
    if not aerospike:exists(r) then 
        aerospike:create(r) 
        r[name] = value
        aerospike:update(r)
    end
end

-- Validate value before writing.
function writeWithValidation(r,name,value)
    if (value >= 1 and value <= 10) then
        putBin(r,name,value)
    else
        error("1000:Invalid value") 
    end
end

-- Record contains two integer bins, name1 and name2.
-- For name1 even integers, add value to existing name1 bin.
-- For name1 integers with a multiple of 5, delete name2 bin.
-- For name1 integers with a multiple of 9, delete record. 
function processRecord(r,name1,name2,addValue)
    local v = r[name1]

    if (v % 9 == 0) then
        aerospike:remove(r)
        return
    end

    if (v % 5 == 0) then
        r[name2] = nil
        aerospike:update(r)
        return
    end

    if (v % 2 == 0) then
        r[name1] = v + addValue
        aerospike:update(r)
    end
end

-- Set expiration of record
-- function expire(r,ttl)
--    if record.ttl(r) == gen then
--        r[name] = value
--        aerospike:update(r)
--    end
-- end
`

func main() {
	register(shared.Client)
	writeUsingUdf(shared.Client)
	writeIfGenerationNotChanged(shared.Client)
	writeIfNotExists(shared.Client)
	writeWithValidation(shared.Client)
	writeListMapUsingUdf(shared.Client)
	writeBlobUsingUdf(shared.Client)

	log.Println("Example finished successfully.")
}

func register(client *as.Client) {
	task, err := client.RegisterUDF(shared.WritePolicy, []byte(udf), "record_example.lua", as.LUA)
	shared.PanicOnError(err)
	<-task.OnComplete()
}

func writeUsingUdf(client *as.Client) {
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "udfkey1")
	bin := as.NewBin("udfbin1", "string value")

	client.Execute(shared.WritePolicy, key, "record_example", "writeBin", as.NewValue(bin.Name), bin.Value)

	record, err := client.Get(shared.Policy, key, bin.Name)
	shared.PanicOnError(err)
	expected := bin.Value.String()
	received := record.Bins[bin.Name].(string)

	if received == expected {
		log.Printf("Data matched: namespace=%s set=%s key=%s bin=%s value=%s",
			key.Namespace(), key.SetName(), key.Value(), bin.Name, received)
	} else {
		log.Printf("Data mismatch: Expected %s. Received %s.", expected, received)
	}
}

func writeIfGenerationNotChanged(client *as.Client) {
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "udfkey2")
	bin := as.NewBin("udfbin2", "string value")

	// Seed record.
	client.PutBins(shared.WritePolicy, key, bin)

	// Get record generation.
	gen, err := client.Execute(shared.WritePolicy, key, "record_example", "getGeneration")
	shared.PanicOnError(err)

	// Write record if generation has not changed.
	client.Execute(shared.WritePolicy, key, "record_example", "writeIfGenerationNotChanged", as.NewValue(bin.Name), bin.Value, as.NewValue(gen))
	log.Printf("Record written.")
}

func writeIfNotExists(client *as.Client) {
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "udfkey3")
	binName := "udfbin3"

	// Delete record if it already exists.
	client.Delete(shared.WritePolicy, key)

	// Write record only if not already exists. This should succeed.
	client.Execute(shared.WritePolicy, key, "record_example", "writeUnique", as.NewValue(binName), as.NewValue("first"))

	// Verify record written.
	record, err := client.Get(shared.Policy, key, binName)
	shared.PanicOnError(err)
	expected := "first"
	received := record.Bins[binName].(string)

	if received == expected {
		log.Printf("Record written: namespace=%s set=%s key=%s bin=%s value=%s",
			key.Namespace(), key.SetName(), key.Value(), binName, received)
	} else {
		log.Printf("Data mismatch: Expected %s. Received %s.", expected, received)
	}

	// Write record second time. This should fail.
	log.Printf("Attempt second write.")
	client.Execute(shared.WritePolicy, key, "record_example", "writeUnique", as.NewValue(binName), as.NewValue("second"))

	// Verify record not written.
	record, err = client.Get(shared.Policy, key, binName)
	shared.PanicOnError(err)
	received = record.Bins[binName].(string)

	if received == expected {
		log.Printf("Success. Record remained unchanged: namespace=%s set=%s key=%s bin=%s value=%s",
			key.Namespace(), key.SetName(), key.Value(), binName, received)
	} else {
		log.Printf("Data mismatch: Expected %s. Received %s.", expected, received)
	}
}

func writeWithValidation(client *as.Client) {
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "udfkey4")
	binName := "udfbin4"

	// Lua function writeWithValidation accepts number between 1 and 10.
	// Write record with valid value.
	log.Printf("Write with valid value.")
	client.Execute(shared.WritePolicy, key, "record_example", "writeWithValidation", as.NewValue(binName), as.NewValue(4))

	// Write record with invalid value.
	log.Printf("Write with invalid value.")

	_, err := client.Execute(shared.WritePolicy, key, "record_example", "writeWithValidation", as.NewValue(binName), as.NewValue(11))
	if err == nil {
		log.Printf("UDF should not have succeeded!")
	} else {
		log.Printf("Success. UDF resulted in exception as expected.")
	}
}

func writeListMapUsingUdf(client *as.Client) {
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "udfkey5")

	inner := []interface{}{"string2", int64(8)}
	innerMap := map[interface{}]interface{}{"a": int64(1), int64(2): "b", "list": inner}
	list := []interface{}{"string1", int64(4), inner, innerMap}

	binName := "udfbin5"

	client.Execute(shared.WritePolicy, key, "record_example", "writeBin", as.NewValue(binName), as.NewValue(list))

	received, err := client.Execute(shared.WritePolicy, key, "record_example", "readBin", as.NewValue(binName))
	shared.PanicOnError(err)

	if testEq(received.([]interface{}), list) {
		log.Printf("UDF data matched: namespace=%s set=%s key=%s bin=%s value=%s",
			key.Namespace(), key.SetName(), key.Value(), binName, received)
	} else {
		log.Println("UDF data mismatch")
		log.Println("Expected ", list)
		log.Println("Received ", received)
	}
}

func writeBlobUsingUdf(client *as.Client) {
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "udfkey6")
	binName := "udfbin6"

	// Create packed blob using standard java tools.
	dos := bytes.Buffer{}
	// dos.Write(9845)
	dos.WriteString("Hello world.")
	blob := dos.Bytes()

	client.Execute(shared.WritePolicy, key, "record_example", "writeBin", as.NewValue(binName), as.NewValue(blob))
	received, err := client.Execute(shared.WritePolicy, key, "record_example", "readBin", as.NewValue(binName))
	shared.PanicOnError(err)

	if bytes.Equal(blob, received.([]byte)) {
		log.Printf("Blob data matched: namespace=%s set=%s key=%s bin=%v value=%v",
			key.Namespace(), key.SetName(), key.Value(), binName, received)
	} else {
		log.Fatalf(
			"Mismatch: expected=%v received=%v", blob, received)
	}
}

func testEq(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
