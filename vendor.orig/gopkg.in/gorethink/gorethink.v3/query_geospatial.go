package gorethink

import (
	p "gopkg.in/gorethink/gorethink.v3/ql2"
)

// CircleOpts contains the optional arguments for the Circle term.
type CircleOpts struct {
	NumVertices interface{} `gorethink:"num_vertices,omitempty"`
	GeoSystem   interface{} `gorethink:"geo_system,omitempty"`
	Unit        interface{} `gorethink:"unit,omitempty"`
	Fill        interface{} `gorethink:"fill,omitempty"`
}

func (o CircleOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Circle constructs a circular line or polygon. A circle in RethinkDB is
// a polygon or line approximating a circle of a given radius around a given
// center, consisting of a specified number of vertices (default 32).
func Circle(point, radius interface{}, optArgs ...CircleOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}

	return constructRootTerm("Circle", p.Term_CIRCLE, []interface{}{point, radius}, opts)
}

// DistanceOpts contains the optional arguments for the Distance term.
type DistanceOpts struct {
	GeoSystem interface{} `gorethink:"geo_system,omitempty"`
	Unit      interface{} `gorethink:"unit,omitempty"`
}

func (o DistanceOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Distance calculates the Haversine distance between two points. At least one
// of the geometry objects specified must be a point.
func (t Term) Distance(point interface{}, optArgs ...DistanceOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}

	return constructMethodTerm(t, "Distance", p.Term_DISTANCE, []interface{}{point}, opts)
}

// Distance calculates the Haversine distance between two points. At least one
// of the geometry objects specified must be a point.
func Distance(point1, point2 interface{}, optArgs ...DistanceOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}

	return constructRootTerm("Distance", p.Term_DISTANCE, []interface{}{point1, point2}, opts)
}

// Fill converts a Line object into a Polygon object. If the last point does not
// specify the same coordinates as the first point, polygon will close the
// polygon by connecting them
func (t Term) Fill() Term {
	return constructMethodTerm(t, "Fill", p.Term_FILL, []interface{}{}, map[string]interface{}{})
}

// GeoJSON converts a GeoJSON object to a ReQL geometry object.
func GeoJSON(args ...interface{}) Term {
	return constructRootTerm("GeoJSON", p.Term_GEOJSON, args, map[string]interface{}{})
}

// ToGeoJSON converts a ReQL geometry object to a GeoJSON object.
func (t Term) ToGeoJSON(args ...interface{}) Term {
	return constructMethodTerm(t, "ToGeoJSON", p.Term_TO_GEOJSON, args, map[string]interface{}{})
}

// GetIntersectingOpts contains the optional arguments for the GetIntersecting term.
type GetIntersectingOpts struct {
	Index interface{} `gorethink:"index,omitempty"`
}

func (o GetIntersectingOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// GetIntersecting gets all documents where the given geometry object intersects
// the geometry object of the requested geospatial index.
func (t Term) GetIntersecting(args interface{}, optArgs ...GetIntersectingOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}

	return constructMethodTerm(t, "GetIntersecting", p.Term_GET_INTERSECTING, []interface{}{args}, opts)
}

// GetNearestOpts contains the optional arguments for the GetNearest term.
type GetNearestOpts struct {
	Index      interface{} `gorethink:"index,omitempty"`
	MaxResults interface{} `gorethink:"max_results,omitempty"`
	MaxDist    interface{} `gorethink:"max_dist,omitempty"`
	Unit       interface{} `gorethink:"unit,omitempty"`
	GeoSystem  interface{} `gorethink:"geo_system,omitempty"`
}

func (o GetNearestOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// GetNearest gets all documents where the specified geospatial index is within a
// certain distance of the specified point (default 100 kilometers).
func (t Term) GetNearest(point interface{}, optArgs ...GetNearestOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}

	return constructMethodTerm(t, "GetNearest", p.Term_GET_NEAREST, []interface{}{point}, opts)
}

// Includes tests whether a geometry object is completely contained within another.
// When applied to a sequence of geometry objects, includes acts as a filter,
// returning a sequence of objects from the sequence that include the argument.
func (t Term) Includes(args ...interface{}) Term {
	return constructMethodTerm(t, "Includes", p.Term_INCLUDES, args, map[string]interface{}{})
}

// Intersects tests whether two geometry objects intersect with one another.
// When applied to a sequence of geometry objects, intersects acts as a filter,
// returning a sequence of objects from the sequence that intersect with the
// argument.
func (t Term) Intersects(args ...interface{}) Term {
	return constructMethodTerm(t, "Intersects", p.Term_INTERSECTS, args, map[string]interface{}{})
}

// Line constructs a geometry object of type Line. The line can be specified in
// one of two ways:
//  - Two or more two-item arrays, specifying longitude and latitude numbers of
//   the line's vertices;
//  - Two or more Point objects specifying the line's vertices.
func Line(args ...interface{}) Term {
	return constructRootTerm("Line", p.Term_LINE, args, map[string]interface{}{})
}

// Point constructs a geometry object of type Point. The point is specified by
// two floating point numbers, the longitude (−180 to 180) and latitude
// (−90 to 90) of the point on a perfect sphere.
func Point(lon, lat interface{}) Term {
	return constructRootTerm("Point", p.Term_POINT, []interface{}{lon, lat}, map[string]interface{}{})
}

// Polygon constructs a geometry object of type Polygon. The Polygon can be
// specified in one of two ways:
//  - Three or more two-item arrays, specifying longitude and latitude numbers of the polygon's vertices;
//  - Three or more Point objects specifying the polygon's vertices.
func Polygon(args ...interface{}) Term {
	return constructRootTerm("Polygon", p.Term_POLYGON, args, map[string]interface{}{})
}

// PolygonSub "punches a hole" out of the parent polygon using the polygon passed
// to the function.
//   polygon1.PolygonSub(polygon2) -> polygon
// In the example above polygon2 must be completely contained within polygon1
// and must have no holes itself (it must not be the output of polygon_sub itself).
func (t Term) PolygonSub(args ...interface{}) Term {
	return constructMethodTerm(t, "PolygonSub", p.Term_POLYGON_SUB, args, map[string]interface{}{})
}
