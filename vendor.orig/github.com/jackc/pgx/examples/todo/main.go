package main

import (
	"fmt"
	"github.com/jackc/pgx"
	"os"
	"strconv"
)

var conn *pgx.Conn

func main() {
	config, err := pgx.ParseEnvLibpq()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to parse environment:", err)
		os.Exit(1)
	}

	conn, err = pgx.Connect(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connection to database: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) == 1 {
		printHelp()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "list":
		err = listTasks()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to list tasks: %v\n", err)
			os.Exit(1)
		}

	case "add":
		err = addTask(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to add task: %v\n", err)
			os.Exit(1)
		}

	case "update":
		n, err := strconv.ParseInt(os.Args[2], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable convert task_num into int32: %v\n", err)
			os.Exit(1)
		}
		err = updateTask(int32(n), os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to update task: %v\n", err)
			os.Exit(1)
		}

	case "remove":
		n, err := strconv.ParseInt(os.Args[2], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable convert task_num into int32: %v\n", err)
			os.Exit(1)
		}
		err = removeTask(int32(n))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to remove task: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintln(os.Stderr, "Invalid command")
		printHelp()
		os.Exit(1)
	}
}

func listTasks() error {
	rows, _ := conn.Query("select * from tasks")

	for rows.Next() {
		var id int32
		var description string
		err := rows.Scan(&id, &description)
		if err != nil {
			return err
		}
		fmt.Printf("%d. %s\n", id, description)
	}

	return rows.Err()
}

func addTask(description string) error {
	_, err := conn.Exec("insert into tasks(description) values($1)", description)
	return err
}

func updateTask(itemNum int32, description string) error {
	_, err := conn.Exec("update tasks set description=$1 where id=$2", description, itemNum)
	return err
}

func removeTask(itemNum int32) error {
	_, err := conn.Exec("delete from tasks where id=$1", itemNum)
	return err
}

func printHelp() {
	fmt.Print(`Todo pgx demo

Usage:

  todo list
  todo add task
  todo update task_num item
  todo remove task_num

Example:

  todo add 'Learn Go'
  todo list
`)
}
