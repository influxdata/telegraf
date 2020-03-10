package main

import (
	"os"
	"strconv"
	"text/template"

	"github.com/gocarina/gocsv"
)

type Data struct {
	Name     string `csv:"Service Name"`
	RawPort  string `csv:"Port Number"`
	RealPort int64
	Protocol string `csv:"Transport Protocol"`
}

func main() {

	template, err := template.ParseFiles("../scripts/protodb/generated-embedded-data-go.template")
	if err != nil {
		panic(err)
	}

	csvFile, err := os.OpenFile("../scripts/protodb/service-names-port-numbers.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	rawData := []*Data{}
	if err := gocsv.UnmarshalFile(csvFile, &rawData); err != nil {
		panic(err)
	}

	data := []*Data{}
	for _, d := range rawData {
		// There are a bunch of port-port unassigned entires in the file and notes, this excludes them
		if i, e := strconv.ParseInt(d.RawPort, 10, 64); e == nil && len(d.Name) > 0 && len(d.Protocol) > 0 {
			data = append(data, &Data{Name: d.Name, RealPort: i, Protocol: d.Protocol})
		}
	}

	destinationFile, err := os.Create("../protodb/generated-embedded-data.go")
	if err != nil {
		panic(err)
	}

	if err := template.Execute(destinationFile, data); err != nil {
		panic(err)
	}
}
