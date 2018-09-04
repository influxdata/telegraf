package pgx_test

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/pkg/errors"
)

var pointRegexp *regexp.Regexp = regexp.MustCompile(`^\((.*),(.*)\)$`)

// Point represents a point that may be null.
type Point struct {
	X, Y   float64 // Coordinates of point
	Status pgtype.Status
}

func (dst *Point) Set(src interface{}) error {
	return errors.Errorf("cannot convert %v to Point", src)
}

func (dst *Point) Get() interface{} {
	switch dst.Status {
	case pgtype.Present:
		return dst
	case pgtype.Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *Point) AssignTo(dst interface{}) error {
	return errors.Errorf("cannot assign %v to %T", src, dst)
}

func (dst *Point) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = Point{Status: pgtype.Null}
		return nil
	}

	s := string(src)
	match := pointRegexp.FindStringSubmatch(s)
	if match == nil {
		return errors.Errorf("Received invalid point: %v", s)
	}

	x, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return errors.Errorf("Received invalid point: %v", s)
	}
	y, err := strconv.ParseFloat(match[2], 64)
	if err != nil {
		return errors.Errorf("Received invalid point: %v", s)
	}

	*dst = Point{X: x, Y: y, Status: pgtype.Present}

	return nil
}

func (src *Point) String() string {
	if src.Status == pgtype.Null {
		return "null point"
	}

	return fmt.Sprintf("%.1f, %.1f", src.X, src.Y)
}

func Example_CustomType() {
	conn, err := pgx.Connect(*defaultConnConfig)
	if err != nil {
		fmt.Printf("Unable to establish connection: %v", err)
		return
	}

	// Override registered handler for point
	conn.ConnInfo.RegisterDataType(pgtype.DataType{
		Value: &Point{},
		Name:  "point",
		OID:   600,
	})

	p := &Point{}
	err = conn.QueryRow("select null::point").Scan(p)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(p)

	err = conn.QueryRow("select point(1.5,2.5)").Scan(p)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(p)
	// Output:
	// null point
	// 1.5, 2.5
}
