package gorethink

import (
	"fmt"
)

// Find a document by ID.
func ExampleTerm_Get() {
	// Fetch the row from the database
	res, err := DB("examples").Table("heroes").Get(2).Run(session)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer res.Close()

	if res.IsNil() {
		fmt.Print("Row not found")
		return
	}

	var hero map[string]interface{}
	err = res.One(&hero)
	if err != nil {
		fmt.Printf("Error scanning database result: %s", err)
		return
	}
	fmt.Print(hero["name"])

	// Output: Superman
}

// Find a document by ID.
func ExampleTerm_GetAll() {
	// Fetch the row from the database
	res, err := DB("examples").Table("heroes").GetAll(2).Run(session)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer res.Close()

	if res.IsNil() {
		fmt.Print("Row not found")
		return
	}

	var hero map[string]interface{}
	err = res.One(&hero)
	if err != nil {
		fmt.Printf("Error scanning database result: %s", err)
		return
	}
	fmt.Print(hero["name"])

	// Output: Superman
}

// Find a document by ID.
func ExampleTerm_GetAll_multiple() {
	// Fetch the row from the database
	res, err := DB("examples").Table("heroes").GetAll(1, 2).Run(session)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer res.Close()

	var heroes []map[string]interface{}
	err = res.All(&heroes)
	if err != nil {
		fmt.Printf("Error scanning database result: %s", err)
		return
	}
	fmt.Print(heroes[0]["name"])

	// Output: Superman
}

// Find all document with an indexed value.
func ExampleTerm_GetAll_optArgs() {
	// Fetch the row from the database
	res, err := DB("examples").Table("heroes").GetAll("man_of_steel").OptArgs(GetAllOpts{
		Index: "code_name",
	}).Run(session)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer res.Close()

	if res.IsNil() {
		fmt.Print("Row not found")
		return
	}

	var hero map[string]interface{}
	err = res.One(&hero)
	if err != nil {
		fmt.Printf("Error scanning database result: %s", err)
		return
	}
	fmt.Print(hero["name"])

	// Output: Superman
}

// Find all document with an indexed value.
func ExampleTerm_GetAllByIndex() {
	// Fetch the row from the database
	res, err := DB("examples").Table("heroes").GetAllByIndex("code_name", "man_of_steel").Run(session)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer res.Close()

	if res.IsNil() {
		fmt.Print("Row not found")
		return
	}

	var hero map[string]interface{}
	err = res.One(&hero)
	if err != nil {
		fmt.Printf("Error scanning database result: %s", err)
		return
	}
	fmt.Print(hero["name"])

	// Output: Superman
}

// Find a document and merge another document with it.
func ExampleTerm_Get_merge() {
	// Fetch the row from the database
	res, err := DB("examples").Table("heroes").Get(4).Merge(map[string]interface{}{
		"powers": []string{"speed"},
	}).Run(session)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer res.Close()

	if res.IsNil() {
		fmt.Print("Row not found")
		return
	}

	var hero map[string]interface{}
	err = res.One(&hero)
	if err != nil {
		fmt.Printf("Error scanning database result: %s", err)
		return
	}
	fmt.Printf("%s: %v", hero["name"], hero["powers"])

	// Output: The Flash: [speed]
}

// Get all users who are 30 years old.
func ExampleTerm_Filter() {
	// Fetch the row from the database
	res, err := DB("examples").Table("users").Filter(map[string]interface{}{
		"age": 30,
	}).Run(session)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer res.Close()

	// Scan query result into the person variable
	var users []interface{}
	err = res.All(&users)
	if err != nil {
		fmt.Printf("Error scanning database result: %s", err)
		return
	}
	fmt.Printf("%d users", len(users))

	// Output: 2 users
}

// Get all users who are more than 25 years old.
func ExampleTerm_Filter_row() {
	// Fetch the row from the database
	res, err := DB("examples").Table("users").Filter(Row.Field("age").Gt(25)).Run(session)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer res.Close()

	// Scan query result into the person variable
	var users []interface{}
	err = res.All(&users)
	if err != nil {
		fmt.Printf("Error scanning database result: %s", err)
		return
	}
	fmt.Printf("%d users", len(users))

	// Output: 3 users
}

// Retrieve all users who have a gmail account (whose field email ends with @gmail.com).
func ExampleTerm_Filter_function() {
	// Fetch the row from the database
	res, err := DB("examples").Table("users").Filter(func(user Term) Term {
		return user.Field("email").Match("@gmail.com$")
	}).Run(session)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer res.Close()

	// Scan query result into the person variable
	var users []interface{}
	err = res.All(&users)
	if err != nil {
		fmt.Printf("Error scanning database result: %s", err)
		return
	}
	fmt.Printf("%d users", len(users))

	// Output: 1 users
}
