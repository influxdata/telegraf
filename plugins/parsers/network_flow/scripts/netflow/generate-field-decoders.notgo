package main

import (
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/gocarina/gocsv"
)

type Data struct {
	ElementID string `csv:"ElementID"`
	ID        int
	Name      string `csv:"Name"`
	Type      string `csv:"Abstract Data Type"`
	AsTag     bool
	NotAsTag  bool
}

var elemementsAsTag = map[int64]bool{
	4:   true,
	5:   true,
	6:   true,
	7:   true,
	8:   true,
	9:   true,
	10:  true,
	11:  true,
	12:  true,
	13:  true,
	14:  true,
	16:  true,
	17:  true,
	18:  true,
	27:  true,
	28:  true,
	48:  true,
	61:  true,
	70:  true,
	89:  true,
	234: true,
	235: true,
}

func main() {

	template, err := template.ParseFiles("../scripts/netflow/generated-field-decoders-go.template")
	if err != nil {
		panic(err)
	}

	csvFile, err := os.OpenFile("../scripts/netflow/ipfix-information-elements.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
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
		i, e := strconv.ParseInt(d.ElementID, 10, 32)
		if e == nil && len(strings.TrimSpace(d.Name)) != 0 {
			d.ID = int(i)
			d.AsTag = elemementsAsTag[i]
			d.NotAsTag = !elemementsAsTag[i]
			data = append(data, d)
		}
	}

	destinationFile, err := os.Create("../netflow/generated-field-decoders.go")
	if err != nil {
		panic(err)
	}

	if err := template.Execute(destinationFile, data); err != nil {
		panic(err)
	}
}
