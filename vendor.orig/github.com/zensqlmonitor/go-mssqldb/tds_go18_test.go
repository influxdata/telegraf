package mssql

import (
	"testing"
	"net/url"
	"os"
	"fmt"
	"database/sql"
)

func TestBadConnect(t *testing.T) {
	t.Skip("still fails https://ci.appveyor.com/project/denisenkom/go-mssqldb/build/job/4jm8fmo1rywje9f9")
	var badDSNs []string

	if parsed, err := url.Parse(os.Getenv("SQLSERVER_DSN")); err == nil {
		parsed.User = url.UserPassword("baduser", "badpwd")
		badDSNs = append(badDSNs, parsed.String())
	}
	if len(os.Getenv("HOST")) > 0 && len(os.Getenv("INSTANCE")) > 0 {
		badDSNs = append(badDSNs,
			fmt.Sprintf(
				"Server=%s\\%s;User ID=baduser;Password=badpwd",
				os.Getenv("HOST"), os.Getenv("INSTANCE"),
			),
		)
	}
	SetLogger(testLogger{t})
	for _, badDsn := range badDSNs {
		conn, err := sql.Open("mssql", badDsn)
		if err != nil {
			t.Error("Open connection failed:", err.Error())
		}
		defer conn.Close()
		err = conn.Ping()
		if err == nil {
			t.Error("Ping should fail for connection: ", badDsn)
		}
	}
}
