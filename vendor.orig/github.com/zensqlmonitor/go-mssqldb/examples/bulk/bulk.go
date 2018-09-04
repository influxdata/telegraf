package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/denisenkom/go-mssqldb"
)

var (
	debug         = flag.Bool("debug", true, "enable debugging")
	password      = flag.String("password", "osmtest", "the database password")
	port     *int = flag.Int("port", 1433, "the database port")
	server        = flag.String("server", "localhost", "the database server")
	user          = flag.String("user", "osmtest", "the database user")
	database      = flag.String("database", "bulktest", "the database name")
)

/*
	CREATE TABLE test_table(
		[id] [int] IDENTITY(1,1) NOT NULL,
		[test_nvarchar] [nvarchar](50) NULL,
		[test_varchar] [varchar](50) NULL,
		[test_float] [float] NULL,
		[test_datetime2_3] [datetime2](3) NULL,
		[test_bitn] [bit] NULL,
		[test_bigint] [bigint] NOT NULL,
		[test_geom] [geometry] NULL,
	 CONSTRAINT [PK_table_test_id] PRIMARY KEY CLUSTERED
	(
		[id] ASC
	) ON [PRIMARY]);
*/

func main() {
	flag.Parse()

	if *debug {
		fmt.Printf(" password:%s\n", *password)
		fmt.Printf(" port:%d\n", *port)
		fmt.Printf(" server:%s\n", *server)
		fmt.Printf(" user:%s\n", *user)
		fmt.Printf(" database:%s\n", *database)
	}

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s", *server, *user, *password, *port, *database)
	if *debug {
		fmt.Printf("connString:%s\n", connString)
	}
	conn, err := sql.Open("mssql", connString)
	if err != nil {
		log.Fatal("Open connection failed:", err.Error())
	}
	defer conn.Close()

	txn, err := conn.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := txn.Prepare(mssql.CopyIn("test_table", mssql.MssqlBulkOptions{}, "test_varchar", "test_nvarchar", "test_float", "test_bigint"))
	if err != nil {
		log.Fatal(err.Error())
	}

	for i := 0; i < 10; i++ {
		_, err = stmt.Exec(generateString(0, 30), generateStringUnicode(0, 30), i, i)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	result, err := stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}

	err = stmt.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = txn.Commit()
	if err != nil {
		log.Fatal(err)
	}
	rowCount, _ := result.RowsAffected()
	log.Printf("%d row copied\n", rowCount)
	log.Printf("bye\n")

}

func generateString(x int, n int) string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}
func generateStringUnicode(x int, n int) string {
	letters := "ab©💾é?ghïjklmnopqЯ☀tuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}
