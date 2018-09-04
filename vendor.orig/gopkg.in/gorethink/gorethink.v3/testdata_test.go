package gorethink

func setupTestData() {
	if *testdata {
		// Delete any preexisting databases
		DBDrop("test").Exec(session)
		DBDrop("examples").Exec(session)
		DBDrop("superheroes").Exec(session)

		DBCreate("test").Exec(session)
		DBCreate("examples").Exec(session)

		DB("test").TableCreate("test").Exec(session)
		DB("test").TableCreate("test2").Exec(session)
		DB("test").TableCreate("changes").Exec(session)

		DB("examples").TableCreate("posts").Exec(session)
		DB("examples").TableCreate("heroes").Exec(session)
		DB("examples").TableCreate("users").Exec(session)
		DB("examples").TableCreate("games").Exec(session)
		DB("examples").TableCreate("games2").Exec(session)
		DB("examples").TableCreate("marvel").Exec(session)

		DB("examples").Table("posts").IndexCreate("date").Exec(session)
		DB("examples").Table("posts").IndexWait().Exec(session)
		DB("examples").Table("posts").IndexCreateFunc(
			"dateAndTitle",
			[]interface{}{Row.Field("date"), Row.Field("title")},
		).Exec(session)

		DB("examples").Table("heroes").IndexCreate("code_name").Exec(session)
		DB("examples").Table("heroes").IndexWait().Exec(session)

		DB("examples").Table("games").IndexCreate("type").Exec(session)
		DB("examples").Table("games").IndexWait().Exec(session)

		// Create heroes table
		DB("examples").Table("heroes").Insert([]interface{}{
			map[string]interface{}{
				"id":        1,
				"code_name": "batman",
				"name":      "Batman",
			},
			map[string]interface{}{
				"id":        2,
				"code_name": "man_of_steel",
				"name":      "Superman",
			},
			map[string]interface{}{
				"id":        3,
				"code_name": "ant_man",
				"name":      "Ant Man",
			},
			map[string]interface{}{
				"id":        4,
				"code_name": "flash",
				"name":      "The Flash",
			},
		}).Exec(session)

		// Create users table
		DB("examples").Table("users").Insert([]interface{}{
			map[string]interface{}{
				"id":    "william",
				"email": "william@rethinkdb.com",
				"age":   30,
			},
			map[string]interface{}{
				"id":    "lara",
				"email": "lara@rethinkdb.com",
				"age":   30,
			},
			map[string]interface{}{
				"id":    "john",
				"email": "john@rethinkdb.com",
				"age":   19,
			},
			map[string]interface{}{
				"id":    "jane",
				"email": "jane@rethinkdb.com",
				"age":   45,
			},
			map[string]interface{}{
				"id":    "bob",
				"email": "bob@rethinkdb.com",
				"age":   24,
			},
			map[string]interface{}{
				"id":    "brad",
				"email": "brad@gmail.com",
				"age":   15,
			},
		}).Exec(session)

		// Create games table
		DB("examples").Table("games").Insert([]interface{}{
			map[string]interface{}{"id": 2, "player": "Bob", "points": 15, "type": "ranked"},
			map[string]interface{}{"id": 5, "player": "Alice", "points": 7, "type": "free"},
			map[string]interface{}{"id": 11, "player": "Bob", "points": 10, "type": "free"},
			map[string]interface{}{"id": 12, "player": "Alice", "points": 2, "type": "free"},
		}).Exec(session)

		// Create games2 table
		DB("examples").Table("games2").Insert([]interface{}{
			map[string]interface{}{"id": 1, "matches": map[string]interface{}{"a": []int{1, 2, 3}, "b": []int{4, 5, 6}}},
			map[string]interface{}{"id": 2, "matches": map[string]interface{}{"b": []int{100}, "c": []int{7, 8, 9}}},
			map[string]interface{}{"id": 3, "matches": map[string]interface{}{"a": []int{10, 20}, "c": []int{70, 80}}},
		}).Exec(session)

		// Create marvel table
		DB("examples").Table("marvel").Insert([]interface{}{
			map[string]interface{}{"name": "Iron Man", "victories": 214},
			map[string]interface{}{"name": "Jubilee", "victories": 9},
		}).Exec(session)
	}
}
