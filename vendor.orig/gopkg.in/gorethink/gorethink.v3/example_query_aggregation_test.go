package gorethink

import (
	"fmt"
)

// Group games by player.
func ExampleTerm_Group() {
	cur, err := DB("examples").Table("games").Group("player").Run(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	var res []interface{}
	err = cur.All(&res)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Print(res)
}

// Group games by the index type.
func ExampleTerm_GroupByIndex() {
	cur, err := DB("examples").Table("games").GroupByIndex("type").Run(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	var res []interface{}
	err = cur.All(&res)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Print(res)
}

// Suppose that the table games2 has the following data:
//
//   [
// 	     { id: 1, matches: {'a': [1, 2, 3], 'b': [4, 5, 6]} },
// 	     { id: 2, matches: {'b': [100], 'c': [7, 8, 9]} },
// 	     { id: 3, matches: {'a': [10, 20], 'c': [70, 80]} }
//   ]
// Using MultiGroup we can group data by match A, B or C.
func ExampleTerm_MultiGroup() {
	cur, err := DB("examples").Table("games2").MultiGroup(Row.Field("matches").Keys()).Run(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	var res []interface{}
	err = cur.All(&res)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Print(res)
}

// Ungrouping grouped data.
func ExampleTerm_Ungroup() {
	cur, err := DB("examples").Table("games").
		Group("player").
		Max("points").Field("points").
		Ungroup().
		Run(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	var res []interface{}
	err = cur.All(&res)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Print(res)
}

// Return the number of documents in the table posts.
func ExampleTerm_Reduce() {
	cur, err := DB("examples").Table("posts").
		Map(func(doc Term) interface{} {
			return 1
		}).
		Reduce(func(left, right Term) interface{} {
			return left.Add(right)
		}).
		Run(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	var res int
	err = cur.One(&res)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Print(res)
}

// Concatenate words from a list.
func ExampleTerm_Fold() {
	cur, err := Expr([]string{"a", "b", "c"}).Fold("", func(acc, word Term) Term {
		return acc.Add(Branch(acc.Eq(""), "", ", ")).Add(word)
	}).Run(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	var res string
	err = cur.One(&res)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Print(res)
	// Output:
	// a, b, c
}
