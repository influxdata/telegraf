package gorethink

import (
	"fmt"
)

// Return the first five squares.
func ExampleTerm_Map() {
	cur, err := Expr([]int{1, 2, 3, 4, 5}).Map(func(val Term) Term {
		return val.Mul(val)
	}).Run(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	var res []int
	err = cur.All(&res)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Print(res)

	// Output:
	// [1 4 9 16 25]
}

// Sum the elements of three sequences.
func ExampleMap_multipleSequences() {
	var sequence1 = []int{100, 200, 300, 400}
	var sequence2 = []int{10, 20, 30, 40}
	var sequence3 = []int{1, 2, 3, 4}

	cur, err := Map(sequence1, sequence2, sequence3, func(val1, val2, val3 Term) Term {
		return val1.Add(val2).Add(val3)
	}).Run(session)
	if err != nil {
		fmt.Print(err)
		return
	}

	var res []int
	err = cur.All(&res)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Print(res)

	// Output:
	// [111 222 333 444]
}

// Order all the posts using the index date.
func ExampleTerm_OrderBy_index() {
	cur, err := DB("examples").Table("posts").OrderBy(OrderByOpts{
		Index: "date",
	}).Run(session)
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

// Order all the posts using the index date in descending order.
func ExampleTerm_OrderBy_indexDesc() {
	cur, err := DB("examples").Table("posts").OrderBy(OrderByOpts{
		Index: Desc("date"),
	}).Run(session)
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

// You can efficiently order using multiple fields by using a compound index.
// For example order by date and title.
func ExampleTerm_OrderBy_compound() {
	cur, err := DB("examples").Table("posts").OrderBy(OrderByOpts{
		Index: Desc("dateAndTitle"),
	}).Run(session)
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

// If you have a sequence with fewer documents than the arrayLimit, you can order
// it by multiple fields without an index.
func ExampleTerm_OrderBy_multiple() {
	cur, err := DB("examples").Table("posts").OrderBy(
		"title",
		OrderByOpts{Index: Desc("date")},
	).Run(session)
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

// Notice that an index ordering always has highest precedence. The following
// query orders posts by date, and if multiple posts were published on the same
// date, they will be ordered by title.
func ExampleTerm_OrderBy_multipleWithIndex() {
	cur, err := DB("examples").Table("posts").OrderBy(
		"title",
		OrderByOpts{Index: Desc("date")},
	).Run(session)
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
