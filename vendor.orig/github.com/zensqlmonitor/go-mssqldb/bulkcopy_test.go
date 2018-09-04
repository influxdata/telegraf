package mssql

import (
	"database/sql"
	"encoding/hex"
	"log"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestBulkcopy(t *testing.T) {
	// TDS level Bulk Insert is not supported on Azure SQL Server.
	if dsn := makeConnStr(t); strings.HasSuffix(strings.Split(dsn.Host, ":")[0], ".database.windows.net") {
		t.Skip("TDS level bulk copy is not supported on Azure SQL Server")
	}
	type testValue struct {
		colname string
		val     interface{}
	}
	tableName := "#table_test"
	geom, _ := hex.DecodeString("E6100000010C00000000000034400000000000004440")
	testValues := []testValue{

		{"test_nvarchar", "ab©ĎéⒻghïjklmnopqЯ☀tuvwxyz"},
		{"test_varchar", "abcdefg"},
		{"test_char", "abcdefg   "},
		{"test_nchar", "abcdefg   "},
		{"test_text", "abcdefg"},
		{"test_ntext", "abcdefg"},
		{"test_float", 1234.56},
		{"test_floatn", 1234.56},
		{"test_real", 1234.56},
		{"test_realn", 1234.56},
		{"test_bit", true},
		{"test_bitn", nil},
		{"test_smalldatetime", time.Date(2010, 11, 12, 13, 14, 0, 0, time.UTC)},
		{"test_smalldatetimen", time.Date(2010, 11, 12, 13, 14, 0, 0, time.UTC)},
		{"test_datetime", time.Date(2010, 11, 12, 13, 14, 15, 120000000, time.UTC)},
		{"test_datetimen", time.Date(2010, 11, 12, 13, 14, 15, 120000000, time.UTC)},
		{"test_datetime2_1", time.Date(2010, 11, 12, 13, 14, 15, 0, time.UTC)},
		{"test_datetime2_3", time.Date(2010, 11, 12, 13, 14, 15, 123000000, time.UTC)},
		{"test_datetime2_7", time.Date(2010, 11, 12, 13, 14, 15, 123000000, time.UTC)},
		{"test_date", time.Date(2010, 11, 12, 00, 00, 00, 0, time.UTC)},
		{"test_tinyint", 255},
		{"test_smallint", 32767},
		{"test_smallintn", nil},
		{"test_int", 2147483647},
		{"test_bigint", 9223372036854775807},
		{"test_bigintn", nil},
		{"test_geom", geom},
		//{"test_smallmoney", nil},
		//{"test_money", nil},
		//{"test_decimal_18_0", nil},
		//{"test_decimal_9_2", nil},
		//{"test_decimal_18_0", nil},
	}

	columns := make([]string, len(testValues))
	for i, val := range testValues {
		columns[i] = val.colname
	}

	values := make([]interface{}, len(testValues))
	for i, val := range testValues {
		values[i] = val.val
	}

	conn := open(t)
	defer conn.Close()

	err := setupTable(conn, tableName)
	if (err != nil) {
		t.Error("Setup table failed: ", err.Error())
		return
	}

	log.Println("Preparing copyin statement")

	stmt, err := conn.Prepare(CopyIn(tableName, MssqlBulkOptions{}, columns...))

	for i := 0; i < 10; i++ {
		log.Printf("Executing copy in statement %d time with %d values", i+1, len(values))
		_, err = stmt.Exec(values...)
		if err != nil {
			t.Error("AddRow failed: ", err.Error())
			return
		}
	}

	result, err := stmt.Exec()
	if err != nil {
		t.Fatal("bulkcopy failed: ", err.Error())
	}

	insertedRowCount, _ := result.RowsAffected()
	if insertedRowCount == 0 {
		t.Fatal("0 row inserted!")
	}

	//check that all rows are present
	var rowCount int
	err = conn.QueryRow("select count(*) c from " + tableName).Scan(&rowCount)

	if rowCount != 10 {
		t.Errorf("unexpected row count %d", rowCount)
	}

	//data verification
	rows, err := conn.Query("select " + strings.Join(columns, ",") + " from " + tableName)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {

		ptrs := make([]interface{}, len(columns))
		container := make([]interface{}, len(columns))
		for i, _ := range ptrs {
			ptrs[i] = &container[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			t.Fatal(err)
		}
		for i, c := range testValues {
			if !compareValue(container[i], c.val) {
				t.Errorf("columns %s : %s != %s\n", c.colname, container[i], c.val)
			}
		}
	}
	if err := rows.Err(); err != nil {
		t.Error(err)
	}
}

func compareValue(a interface{}, expected interface{}) bool {
	switch expected := expected.(type) {
	case int:
		return int64(expected) == a
	case int32:
		return int64(expected) == a
	case int64:
		return int64(expected) == a
	case float64:
		return math.Abs(expected-a.(float64)) < 0.0001
	default:
		return reflect.DeepEqual(expected, a)
	}
}

func setupTable(conn *sql.DB, tableName string) (err error) {
	tablesql := `CREATE TABLE ` + tableName + ` (
	[id] [int] IDENTITY(1,1) NOT NULL,
	[test_nvarchar] [nvarchar](50) NULL,
	[test_varchar] [varchar](50) NULL,
	[test_char] [char](10) NULL,
	[test_nchar] [nchar](10) NULL,
	[test_text] [text] NULL,
	[test_ntext] [ntext] NULL,
	[test_float] [float] NOT NULL,
	[test_floatn] [float] NULL,
	[test_real] [real] NULL,
	[test_realn] [real] NULL,
	[test_bit] [bit] NOT NULL,
	[test_bitn] [bit] NULL,
	[test_smalldatetime] [smalldatetime] NOT NULL,
	[test_smalldatetimen] [smalldatetime] NULL,
	[test_datetime] [datetime] NOT NULL,
	[test_datetimen] [datetime] NULL,
	[test_datetime2_1] [datetime2](1) NULL,
	[test_datetime2_3] [datetime2](3) NULL,
	[test_datetime2_7] [datetime2](7) NULL,
	[test_date] [date] NULL,
	[test_smallmoney] [smallmoney] NULL,
	[test_money] [money] NULL,
	[test_tinyint] [tinyint] NULL,
	[test_smallint] [smallint] NOT NULL,
	[test_smallintn] [smallint] NULL,
	[test_int] [int] NULL,
	[test_bigint] [bigint] NOT NULL,
	[test_bigintn] [bigint] NULL,
	[test_geom] [geometry] NULL,
	[test_geog] [geography] NULL,
	[text_xml] [xml] NULL,
	[test_uniqueidentifier] [uniqueidentifier] NULL,
	[test_decimal_18_0] [decimal](18, 0) NULL,
	[test_decimal_9_2] [decimal](9, 2) NULL,
	[test_decimal_20_0] [decimal](20, 0) NULL,
 CONSTRAINT [PK_` + tableName + `_id] PRIMARY KEY CLUSTERED 
(
	[id] ASC
)WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]
) ON [PRIMARY] TEXTIMAGE_ON [PRIMARY];`
	_, err = conn.Exec(tablesql)
	if err != nil {
		log.Fatal("tablesql failed:", err)
	}
	return
}
