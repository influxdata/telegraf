package types

import (
	"fmt"
)

type Geometry struct {
	Type  string
	Point Point
	Line  Line
	Lines Lines
}

func (g Geometry) MarshalRQL() (interface{}, error) {
	switch g.Type {
	case "Point":
		return g.Point.MarshalRQL()
	case "LineString":
		return g.Line.MarshalRQL()
	case "Polygon":
		return g.Lines.MarshalRQL()
	default:
		return nil, fmt.Errorf("pseudo-type GEOMETRY object field 'type' %s is not valid", g.Type)
	}
}

func (g *Geometry) UnmarshalRQL(data interface{}) error {
	if data, ok := data.(Geometry); ok {
		g.Type = data.Type
		g.Point = data.Point
		g.Line = data.Line
		g.Lines = data.Lines

		return nil
	}

	m, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("pseudo-type GEOMETRY object is not valid")
	}

	typ, ok := m["type"]
	if !ok {
		return fmt.Errorf("pseudo-type GEOMETRY object is not valid, expects 'type' field")
	}
	coords, ok := m["coordinates"]
	if !ok {
		return fmt.Errorf("pseudo-type GEOMETRY object is not valid, expects 'coordinates' field")
	}

	var err error
	switch typ {
	case "Point":
		g.Type = "Point"
		g.Point, err = UnmarshalPoint(coords)
	case "LineString":
		g.Type = "LineString"
		g.Line, err = UnmarshalLineString(coords)
	case "Polygon":
		g.Type = "Polygon"
		g.Lines, err = UnmarshalPolygon(coords)
	default:
		return fmt.Errorf("pseudo-type GEOMETRY object has invalid type")
	}

	if err != nil {
		return err
	}

	return nil
}

type Point struct {
	Lon float64
	Lat float64
}
type Line []Point
type Lines []Line

func (p Point) Coords() interface{} {
	return []interface{}{p.Lon, p.Lat}
}

func (p Point) MarshalRQL() (interface{}, error) {
	return map[string]interface{}{
		"$reql_type$": "GEOMETRY",
		"coordinates": p.Coords(),
		"type":        "Point",
	}, nil
}

func (p *Point) UnmarshalRQL(data interface{}) error {
	g := &Geometry{}
	err := g.UnmarshalRQL(data)
	if err != nil {
		return err
	}
	if g.Type != "Point" {
		return fmt.Errorf("pseudo-type GEOMETRY object has type %s, expected type %s", g.Type, "Point")
	}

	p.Lat = g.Point.Lat
	p.Lon = g.Point.Lon

	return nil
}

func (l Line) Coords() interface{} {
	coords := make([]interface{}, len(l))
	for i, point := range l {
		coords[i] = point.Coords()
	}
	return coords
}

func (l Line) MarshalRQL() (interface{}, error) {
	return map[string]interface{}{
		"$reql_type$": "GEOMETRY",
		"coordinates": l.Coords(),
		"type":        "LineString",
	}, nil
}

func (l *Line) UnmarshalRQL(data interface{}) error {
	g := &Geometry{}
	err := g.UnmarshalRQL(data)
	if err != nil {
		return err
	}
	if g.Type != "LineString" {
		return fmt.Errorf("pseudo-type GEOMETRY object has type %s, expected type %s", g.Type, "LineString")
	}

	*l = g.Line

	return nil
}

func (l Lines) Coords() interface{} {
	coords := make([]interface{}, len(l))
	for i, line := range l {
		coords[i] = line.Coords()
	}
	return coords
}

func (l Lines) MarshalRQL() (interface{}, error) {
	return map[string]interface{}{
		"$reql_type$": "GEOMETRY",
		"coordinates": l.Coords(),
		"type":        "Polygon",
	}, nil
}

func (l *Lines) UnmarshalRQL(data interface{}) error {
	g := &Geometry{}
	err := g.UnmarshalRQL(data)
	if err != nil {
		return err
	}
	if g.Type != "Polygon" {
		return fmt.Errorf("pseudo-type GEOMETRY object has type %s, expected type %s", g.Type, "Polygon")
	}

	*l = g.Lines

	return nil
}

func UnmarshalPoint(v interface{}) (Point, error) {
	coords, ok := v.([]interface{})
	if !ok {
		return Point{}, fmt.Errorf("pseudo-type GEOMETRY object field 'coordinates' is not valid")
	}
	if len(coords) != 2 {
		return Point{}, fmt.Errorf("pseudo-type GEOMETRY object field 'coordinates' is not valid")
	}
	lon, ok := coords[0].(float64)
	if !ok {
		return Point{}, fmt.Errorf("pseudo-type GEOMETRY object field 'coordinates' is not valid")
	}
	lat, ok := coords[1].(float64)
	if !ok {
		return Point{}, fmt.Errorf("pseudo-type GEOMETRY object field 'coordinates' is not valid")
	}

	return Point{
		Lon: lon,
		Lat: lat,
	}, nil
}

func UnmarshalLineString(v interface{}) (Line, error) {
	points, ok := v.([]interface{})
	if !ok {
		return Line{}, fmt.Errorf("pseudo-type GEOMETRY object field 'coordinates' is not valid")
	}

	var err error
	line := make(Line, len(points))
	for i, coords := range points {
		line[i], err = UnmarshalPoint(coords)
		if err != nil {
			return Line{}, err
		}
	}
	return line, nil
}

func UnmarshalPolygon(v interface{}) (Lines, error) {
	lines, ok := v.([]interface{})
	if !ok {
		return Lines{}, fmt.Errorf("pseudo-type GEOMETRY object field 'coordinates' is not valid")
	}

	var err error
	polygon := make(Lines, len(lines))
	for i, line := range lines {
		polygon[i], err = UnmarshalLineString(line)
		if err != nil {
			return Lines{}, err
		}
	}
	return polygon, nil
}
