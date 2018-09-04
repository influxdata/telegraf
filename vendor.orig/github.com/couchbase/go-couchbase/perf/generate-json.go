package main

import (
	"encoding/json"
	"github.com/Pallinder/go-randomdata"
)

type contacts struct {
	Name    string
	Email   string
	Age     int
	Address string
	City    string
	State   string
	Country string
}

// return a json marshalled document
func generateRandomDocument() ([]byte, error) {
	c := &contacts{}
	c.Name = randomdata.FullName(randomdata.RandomGender)
	c.Email = randomdata.Email()
	c.Age = randomdata.Number(20, 50)
	c.Address = randomdata.Address()
	c.City = randomdata.City()
	c.State = randomdata.State(randomdata.Large)
	c.Country = randomdata.Country(randomdata.FullCountry)

	return json.Marshal(c)
}
