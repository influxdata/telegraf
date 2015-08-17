package gorethink

import (
	"fmt"
)

// Insert a document into the table posts using a struct.
func ExampleTerm_Insert_struct() {
	type Post struct {
		ID      int    `gorethink:"id"`
		Title   string `gorethink:"title"`
		Content string `gorethink:"content"`
	}

	resp, err := DB("examples").Table("posts").Insert(Post{
		ID:      1,
		Title:   "Lorem ipsum",
		Content: "Dolor sit amet",
	}).RunWrite(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("%d row inserted", resp.Inserted)

	// Output:
	// 1 row inserted
}

// Insert a document without a defined primary key into the table posts where
// the primary key is id.
func ExampleTerm_Insert_generatedKey() {
	type Post struct {
		Title   string `gorethink:"title"`
		Content string `gorethink:"content"`
	}

	resp, err := DB("examples").Table("posts").Insert(map[string]interface{}{
		"title":   "Lorem ipsum",
		"content": "Dolor sit amet",
	}).RunWrite(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("%d row inserted, %d key generated", resp.Inserted, len(resp.GeneratedKeys))

	// Output:
	// 1 row inserted, 1 key generated
}

// Insert a document into the table posts using a map.
func ExampleTerm_Insert_map() {
	resp, err := DB("examples").Table("posts").Insert(map[string]interface{}{
		"id":      2,
		"title":   "Lorem ipsum",
		"content": "Dolor sit amet",
	}).RunWrite(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("%d row inserted", resp.Inserted)

	// Output:
	// 1 row inserted
}

// Insert multiple documents into the table posts.
func ExampleTerm_Insert_multiple() {
	resp, err := DB("examples").Table("posts").Insert([]interface{}{
		map[string]interface{}{
			"title":   "Lorem ipsum",
			"content": "Dolor sit amet",
		},
		map[string]interface{}{
			"title":   "Lorem ipsum",
			"content": "Dolor sit amet",
		},
	}).RunWrite(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("%d rows inserted", resp.Inserted)

	// Output:
	// 2 rows inserted
}

// Insert a document into the table posts, replacing the document if it already
// exists.
func ExampleTerm_Insert_upsert() {
	resp, err := DB("examples").Table("posts").Insert(map[string]interface{}{
		"id":    1,
		"title": "Lorem ipsum 2",
	}, InsertOpts{
		Conflict: "replace",
	}).RunWrite(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("%d row replaced", resp.Replaced)

	// Output:
	// 1 row replaced
}

// Update the status of the post with id of 1 to published.
func ExampleTerm_Update() {
	resp, err := DB("examples").Table("posts").Get(2).Update(map[string]interface{}{
		"status": "published",
	}).RunWrite(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("%d row replaced", resp.Replaced)

	// Output:
	// 1 row replaced
}

// Update bob's cell phone number.
func ExampleTerm_Update_nested() {
	resp, err := DB("examples").Table("users").Get("bob").Update(map[string]interface{}{
		"contact": map[string]interface{}{
			"phone": "408-555-4242",
		},
	}).RunWrite(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("%d row replaced", resp.Replaced)

	// Output:
	// 1 row replaced
}

// Update the status of all posts to published.
func ExampleTerm_Update_all() {
	resp, err := DB("examples").Table("posts").Update(map[string]interface{}{
		"status": "published",
	}).RunWrite(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("%d row replaced", resp.Replaced)

	// Output:
	// 4 row replaced
}

// Increment the field view of the post with id of 1. If the field views does not
// exist, it will be set to 0.
func ExampleTerm_Update_increment() {
	resp, err := DB("examples").Table("posts").Get(1).Update(map[string]interface{}{
		"views": Row.Field("views").Add(1).Default(0),
	}).RunWrite(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("%d row replaced", resp.Replaced)

	// Output:
	// 1 row replaced
}

// Update the status of the post with id of 1 using soft durability.
func ExampleTerm_Update_softDurability() {
	resp, err := DB("examples").Table("posts").Get(2).Update(map[string]interface{}{
		"status": "draft",
	}, UpdateOpts{
		Durability: "soft",
	}).RunWrite(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("%d row replaced", resp.Replaced)

	// Output:
	// 1 row replaced
}

// Delete a single document from the table posts.
func ExampleTerm_Delete() {
	resp, err := DB("examples").Table("posts").Get(2).Delete().RunWrite(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("%d row deleted", resp.Deleted)

	// Output:
	// 1 row deleted
}

// Delete all comments where the field status is published
func ExampleTerm_Delete_many() {
	resp, err := DB("examples").Table("posts").Filter(map[string]interface{}{
		"status": "published",
	}).Delete().RunWrite(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("%d rows deleted", resp.Deleted)

	// Output:
	// 4 rows deleted
}
