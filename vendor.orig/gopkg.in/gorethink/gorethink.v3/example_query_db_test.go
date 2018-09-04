package gorethink

import (
	"fmt"
)

// Create a database named ’superheroes’.
func ExampleDBCreate() {
	resp, err := DBCreate("superheroes").RunWrite(session)
	if err != nil {
		fmt.Print(err)
	}

	fmt.Printf("%d DB created", resp.DBsCreated)
	// Output:
	// 1 DB created
}

// Drop a database named ‘superheroes’.
func ExampleDBDrop() {
	// Setup database + tables
	DBCreate("superheroes").Exec(session)
	DB("superheroes").TableCreate("superheroes").Exec(session)
	DB("superheroes").TableCreate("battles").Exec(session)

	resp, err := DBDrop("superheroes").RunWrite(session)
	if err != nil {
		fmt.Print(err)
	}

	fmt.Printf("%d DB dropped, %d tables dropped", resp.DBsDropped, resp.TablesDropped)
	// Output:
	// 1 DB dropped, 2 tables dropped
}
