package mssql

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"testing"
)

func TestUniqueIdentifier(t *testing.T) {
	dbUUID := UniqueIdentifier{0x67, 0x45, 0x23, 0x01,
		0xAB, 0x89,
		0xEF, 0xCD,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF,
	}

	uuid := UniqueIdentifier{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}

	t.Run("Scan", func(t *testing.T) {
		t.Run("[]byte", func(t *testing.T) {
			var sut UniqueIdentifier
			if err := sut.Scan(dbUUID[:]); err != nil {
				t.Fatal(err)
			}
			if sut != uuid {
				t.Errorf("bytes not swapped correctly: got %q; want %q", sut, uuid)
			}
		})

		t.Run("string", func(t *testing.T) {
			var sut UniqueIdentifier
			if err := sut.Scan(uuid.String()); err != nil {
				t.Fatal(err)
			}
			if sut != uuid {
				t.Errorf("string not scanned correctly: got %q; want %q", sut, uuid)
			}
		})
	})

	t.Run("Value", func(t *testing.T) {
		sut := uuid
		v, err := sut.Value()
		if err != nil {
			t.Fatal(err)
		}

		b, ok := v.([]byte)
		if !ok {
			t.Fatalf("(%T) is not []byte", v)
		}

		if !bytes.Equal(b, dbUUID[:]) {
			t.Errorf("got %q; want %q", b, dbUUID)
		}
	})
}

func TestUniqueIdentifierString(t *testing.T) {
	sut := UniqueIdentifier{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}
	expected := "01234567-89AB-CDEF-0123-456789ABCDEF"
	if actual := sut.String(); actual != expected {
		t.Errorf("sut.String() = %s; want %s", sut, expected)
	}
}

var _ fmt.Stringer = UniqueIdentifier{}
var _ sql.Scanner = &UniqueIdentifier{}
var _ driver.Valuer = UniqueIdentifier{}
