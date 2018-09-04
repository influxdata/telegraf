package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx"
)

var pool *pgx.ConnPool

func main() {
	config, err := pgx.ParseEnvLibpq()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to parse environment:", err)
		os.Exit(1)
	}

	pool, err = pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: config})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to connect to database:", err)
		os.Exit(1)
	}

	go listen()

	fmt.Println(`Type a message and press enter.

This message should appear in any other chat instances connected to the same
database.

Type "exit" to quit.`)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := scanner.Text()
		if msg == "exit" {
			os.Exit(0)
		}

		_, err = pool.Exec("select pg_notify('chat', $1)", msg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error sending notification:", err)
			os.Exit(1)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error scanning from stdin:", err)
		os.Exit(1)
	}
}

func listen() {
	conn, err := pool.Acquire()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error acquiring connection:", err)
		os.Exit(1)
	}
	defer pool.Release(conn)

	conn.Listen("chat")

	for {
		notification, err := conn.WaitForNotification(context.Background())
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error waiting for notification:", err)
			os.Exit(1)
		}

		fmt.Println("PID:", notification.PID, "Channel:", notification.Channel, "Payload:", notification.Payload)
	}
}
