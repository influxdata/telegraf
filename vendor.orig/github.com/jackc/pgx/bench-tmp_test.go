package pgx_test

import (
	"testing"
)

func BenchmarkPgtypeInt4ParseBinary(b *testing.B) {
	conn := mustConnect(b, *defaultConnConfig)
	defer closeConn(b, conn)

	_, err := conn.Prepare("selectBinary", "select n::int4 from generate_series(1, 100) n")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var n int32

		rows, err := conn.Query("selectBinary")
		if err != nil {
			b.Fatal(err)
		}

		for rows.Next() {
			err := rows.Scan(&n)
			if err != nil {
				b.Fatal(err)
			}
		}

		if rows.Err() != nil {
			b.Fatal(rows.Err())
		}
	}
}

func BenchmarkPgtypeInt4EncodeBinary(b *testing.B) {
	conn := mustConnect(b, *defaultConnConfig)
	defer closeConn(b, conn)

	_, err := conn.Prepare("encodeBinary", "select $1::int4, $2::int4, $3::int4, $4::int4, $5::int4, $6::int4, $7::int4")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := conn.Query("encodeBinary", int32(i), int32(i), int32(i), int32(i), int32(i), int32(i), int32(i))
		if err != nil {
			b.Fatal(err)
		}
		rows.Close()
	}
}
