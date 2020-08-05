package avro

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Schema struct {
	// Schema is the Avro schema string.
	Schema string `json:"schema"`
	// Subject where the schema is registered for.
	Subject string `json:"subject"`
	// Version of the returned schema.
	Version int `json:"version"`
	ID      int `json:"id,omitempty"`
}

func GetSchema(avroRegistry string) (string, int) {
	resp, err := http.Get(avroRegistry)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var schema Schema
	err = json.Unmarshal(body, &schema)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(schema.Schema, schema.ID)
	return schema.Schema, schema.ID
}
