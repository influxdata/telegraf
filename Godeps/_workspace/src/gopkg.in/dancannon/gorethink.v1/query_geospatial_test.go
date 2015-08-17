package gorethink

import (
	test "gopkg.in/check.v1"

	"github.com/dancannon/gorethink/types"
)

func (s *RethinkSuite) TestGeospatialDecodeGeometryPseudoType(c *test.C) {
	var response types.Geometry

	// setup coordinates
	coords := [][][]float64{
		{
			{-122.423246, 37.779388},
			{-122.423246, 37.329898},
			{-121.88642, 37.329898},
			{-121.88642, 37.329898},
			{-122.423246, 37.779388},
		},
	}

	gt := "Polygon"
	res, err := Expr(map[string]interface{}{
		"$reql_type$": "GEOMETRY",
		"type":        "Polygon",
		"coordinates": coords,
	}).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)

	// test shape
	if response.Type != gt {
		c.Errorf("expected [%v], instead [%v]", gt, response.Type)
	}

	// assert points are within threshold
	c.Assert(response, geometryEquals, "Polygon", coords)
}

func (s *RethinkSuite) TestGeospatialEncodeGeometryPseudoType(c *test.C) {
	encoded, err := encode(types.Geometry{
		Type: "Polygon",
		Lines: types.Lines{
			types.Line{
				types.Point{Lon: -122.423246, Lat: 37.779388},
				types.Point{Lon: -122.423246, Lat: 37.329898},
				types.Point{Lon: -121.88642, Lat: 37.329898},
				types.Point{Lon: -121.88642, Lat: 37.779388},
				types.Point{Lon: -122.423246, Lat: 37.779388},
			},
		},
	})
	c.Assert(err, test.IsNil)
	c.Assert(encoded, jsonEquals, map[string]interface{}{
		"$reql_type$": "GEOMETRY",
		"type":        "Polygon",
		"coordinates": []interface{}{
			[]interface{}{
				[]interface{}{-122.423246, 37.779388},
				[]interface{}{-122.423246, 37.329898},
				[]interface{}{-121.88642, 37.329898},
				[]interface{}{-121.88642, 37.779388},
				[]interface{}{-122.423246, 37.779388},
			},
		},
	})
}

func (s *RethinkSuite) TestGeospatialCircle(c *test.C) {
	var response types.Geometry
	res, err := Circle([]float64{-122.423246, 37.779388}, 10).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, geometryEquals, "Polygon", [][][]float64{
		{
			{-122.423246, 37.77929790366427},
			{-122.42326814543915, 37.77929963483801},
			{-122.4232894398445, 37.779304761831504},
			{-122.42330906488651, 37.77931308761787},
			{-122.42332626638755, 37.77932429224285},
			{-122.42334038330416, 37.77933794512014},
			{-122.42335087313059, 37.77935352157849},
			{-122.42335733274696, 37.77937042302436},
			{-122.4233595139113, 37.77938799994533},
			{-122.42335733279968, 37.7794055768704},
			{-122.42335087322802, 37.779422478327966},
			{-122.42334038343147, 37.77943805480385},
			{-122.42332626652532, 37.779451707701796},
			{-122.42330906501378, 37.77946291234741},
			{-122.42328943994191, 37.77947123815131},
			{-122.42326814549187, 37.77947636515649},
			{-122.423246, 37.779478096334365},
			{-122.42322385450814, 37.77947636515649},
			{-122.4232025600581, 37.77947123815131},
			{-122.42318293498623, 37.77946291234741},
			{-122.42316573347469, 37.779451707701796},
			{-122.42315161656855, 37.77943805480385},
			{-122.423141126772, 37.779422478327966},
			{-122.42313466720033, 37.7794055768704},
			{-122.42313248608872, 37.77938799994533},
			{-122.42313466725305, 37.77937042302436},
			{-122.42314112686942, 37.77935352157849},
			{-122.42315161669585, 37.77933794512014},
			{-122.42316573361246, 37.77932429224285},
			{-122.4231829351135, 37.77931308761787},
			{-122.42320256015552, 37.779304761831504},
			{-122.42322385456086, 37.77929963483801},
			{-122.423246, 37.77929790366427},
		},
	})
}

func (s *RethinkSuite) TestGeospatialCirclePoint(c *test.C) {
	var response types.Geometry
	res, err := Circle(Point(-122.423246, 37.779388), 10).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, geometryEquals, "Polygon", [][][]float64{
		{
			{-122.423246, 37.77929790366427},
			{-122.42326814543915, 37.77929963483801},
			{-122.4232894398445, 37.779304761831504},
			{-122.42330906488651, 37.77931308761787},
			{-122.42332626638755, 37.77932429224285},
			{-122.42334038330416, 37.77933794512014},
			{-122.42335087313059, 37.77935352157849},
			{-122.42335733274696, 37.77937042302436},
			{-122.4233595139113, 37.77938799994533},
			{-122.42335733279968, 37.7794055768704},
			{-122.42335087322802, 37.779422478327966},
			{-122.42334038343147, 37.77943805480385},
			{-122.42332626652532, 37.779451707701796},
			{-122.42330906501378, 37.77946291234741},
			{-122.42328943994191, 37.77947123815131},
			{-122.42326814549187, 37.77947636515649},
			{-122.423246, 37.779478096334365},
			{-122.42322385450814, 37.77947636515649},
			{-122.4232025600581, 37.77947123815131},
			{-122.42318293498623, 37.77946291234741},
			{-122.42316573347469, 37.779451707701796},
			{-122.42315161656855, 37.77943805480385},
			{-122.423141126772, 37.779422478327966},
			{-122.42313466720033, 37.7794055768704},
			{-122.42313248608872, 37.77938799994533},
			{-122.42313466725305, 37.77937042302436},
			{-122.42314112686942, 37.77935352157849},
			{-122.42315161669585, 37.77933794512014},
			{-122.42316573361246, 37.77932429224285},
			{-122.4231829351135, 37.77931308761787},
			{-122.42320256015552, 37.779304761831504},
			{-122.42322385456086, 37.77929963483801},
			{-122.423246, 37.77929790366427},
		},
	})
}

func (s *RethinkSuite) TestGeospatialCirclePointFill(c *test.C) {
	var response types.Geometry
	res, err := Circle(Point(-122.423246, 37.779388), 10, CircleOpts{Fill: true}).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, geometryEquals, "Polygon", [][][]float64{
		{
			{-122.423246, 37.77929790366427},
			{-122.42326814543915, 37.77929963483801},
			{-122.4232894398445, 37.779304761831504},
			{-122.42330906488651, 37.77931308761787},
			{-122.42332626638755, 37.77932429224285},
			{-122.42334038330416, 37.77933794512014},
			{-122.42335087313059, 37.77935352157849},
			{-122.42335733274696, 37.77937042302436},
			{-122.4233595139113, 37.77938799994533},
			{-122.42335733279968, 37.7794055768704},
			{-122.42335087322802, 37.779422478327966},
			{-122.42334038343147, 37.77943805480385},
			{-122.42332626652532, 37.779451707701796},
			{-122.42330906501378, 37.77946291234741},
			{-122.42328943994191, 37.77947123815131},
			{-122.42326814549187, 37.77947636515649},
			{-122.423246, 37.779478096334365},
			{-122.42322385450814, 37.77947636515649},
			{-122.4232025600581, 37.77947123815131},
			{-122.42318293498623, 37.77946291234741},
			{-122.42316573347469, 37.779451707701796},
			{-122.42315161656855, 37.77943805480385},
			{-122.423141126772, 37.779422478327966},
			{-122.42313466720033, 37.7794055768704},
			{-122.42313248608872, 37.77938799994533},
			{-122.42313466725305, 37.77937042302436},
			{-122.42314112686942, 37.77935352157849},
			{-122.42315161669585, 37.77933794512014},
			{-122.42316573361246, 37.77932429224285},
			{-122.4231829351135, 37.77931308761787},
			{-122.42320256015552, 37.779304761831504},
			{-122.42322385456086, 37.77929963483801},
			{-122.423246, 37.77929790366427},
		},
	})

}

func (s *RethinkSuite) TestGeospatialPointDistanceMethod(c *test.C) {
	var response float64
	f := 734125.249602186
	res, err := Point(-122.423246, 37.779388).Distance(Point(-117.220406, 32.719464)).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	if !kindaclose(response, f) {
		c.Errorf("the deviation between the compared floats is too great [%v:%v]", response, f)
	}
}

func (s *RethinkSuite) TestGeospatialPointDistanceRoot(c *test.C) {
	var response float64
	f := 734125.249602186
	res, err := Distance(Point(-122.423246, 37.779388), Point(-117.220406, 32.719464)).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	if !kindaclose(response, f) {
		c.Errorf("the deviation between the compared floats is too great [%v:%v]", response, f)
	}
}

func (s *RethinkSuite) TestGeospatialPointDistanceRootKm(c *test.C) {
	var response float64
	f := 734.125249602186
	res, err := Distance(Point(-122.423246, 37.779388), Point(-117.220406, 32.719464), DistanceOpts{Unit: "km"}).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	if !kindaclose(response, f) {
		c.Errorf("the deviation between the compared floats is too great [%v:%v]", response, f)
	}
}

func (s *RethinkSuite) TestGeospatialFill(c *test.C) {
	var response types.Geometry
	res, err := Line(
		[]float64{-122.423246, 37.779388},
		[]float64{-122.423246, 37.329898},
		[]float64{-121.886420, 37.329898},
		[]float64{-121.886420, 37.779388},
	).Fill().Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, geometryEquals, "Polygon", [][][]float64{
		{
			{-122.423246, 37.779388},
			{-122.423246, 37.329898},
			{-121.88642, 37.329898},
			{-121.88642, 37.779388},
			{-122.423246, 37.779388},
		},
	})
}

func (s *RethinkSuite) TestGeospatialGeoJSON(c *test.C) {
	var response types.Geometry
	res, err := GeoJSON(map[string]interface{}{
		"type":        "Point",
		"coordinates": []interface{}{-122.423246, 37.779388},
	}).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, geometryEquals, "Point", []float64{-122.423246, 37.779388})
}

func (s *RethinkSuite) TestGeospatialToGeoJSON(c *test.C) {
	var response map[string]interface{}
	res, err := Point(-122.423246, 37.779388).ToGeoJSON().Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, map[string]interface{}{
		"type":        "Point",
		"coordinates": []interface{}{-122.423246, 37.779388},
	})
}

func (s *RethinkSuite) TestGeospatialGetIntersecting(c *test.C) {
	// Setup table
	DB("test").TableDrop("geospatial").Run(session)
	DB("test").TableCreate("geospatial").Run(session)
	DB("test").Table("geospatial").IndexCreate("area", IndexCreateOpts{
		Geo: true,
	}).Run(session)
	DB("test").Table("geospatial").Insert([]interface{}{
		map[string]interface{}{"area": Circle(Point(-117.220406, 32.719464), 100000)},
		map[string]interface{}{"area": Circle(Point(-100.220406, 20.719464), 100000)},
		map[string]interface{}{"area": Circle(Point(-117.200406, 32.723464), 100000)},
	}).Run(session)

	var response []interface{}
	res, err := DB("test").Table("geospatial").GetIntersecting(
		Circle(Point(-117.220406, 32.719464), 100000),
		GetIntersectingOpts{
			Index: "area",
		},
	).Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, test.HasLen, 2)
}

func (s *RethinkSuite) TestGeospatialGetNearest(c *test.C) {
	// Setup table
	DB("test").TableDrop("geospatial").Run(session)
	DB("test").TableCreate("geospatial").Run(session)
	DB("test").Table("geospatial").IndexCreate("area", IndexCreateOpts{
		Geo: true,
	}).Run(session)
	DB("test").Table("geospatial").Insert([]interface{}{
		map[string]interface{}{"area": Circle(Point(-117.220406, 32.719464), 100000)},
		map[string]interface{}{"area": Circle(Point(-100.220406, 20.719464), 100000)},
		map[string]interface{}{"area": Circle(Point(-115.210306, 32.733364), 100000)},
	}).Run(session)

	var response []interface{}
	res, err := DB("test").Table("geospatial").GetNearest(
		Point(-117.220406, 32.719464),
		GetNearestOpts{
			Index:   "area",
			MaxDist: 1,
		},
	).Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.HasLen, 1)
}

func (s *RethinkSuite) TestGeospatialIncludesTrue(c *test.C) {
	var response bool
	res, err := Polygon(
		Point(-122.4, 37.7),
		Point(-122.4, 37.3),
		Point(-121.8, 37.3),
		Point(-121.8, 37.7),
	).Includes(Point(-122.3, 37.4)).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, true)
}

func (s *RethinkSuite) TestGeospatialIncludesFalse(c *test.C) {
	var response bool
	res, err := Polygon(
		Point(-122.4, 37.7),
		Point(-122.4, 37.3),
		Point(-121.8, 37.3),
		Point(-121.8, 37.7),
	).Includes(Point(100.3, 37.4)).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, false)
}

func (s *RethinkSuite) TestGeospatialIntersectsTrue(c *test.C) {
	var response bool
	res, err := Polygon(
		Point(-122.4, 37.7),
		Point(-122.4, 37.3),
		Point(-121.8, 37.3),
		Point(-121.8, 37.7),
	).Intersects(Polygon(
		Point(-122.3, 37.4),
		Point(-122.4, 37.3),
		Point(-121.8, 37.3),
		Point(-121.8, 37.4),
	)).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, true)
}

func (s *RethinkSuite) TestGeospatialIntersectsFalse(c *test.C) {
	var response bool
	res, err := Polygon(
		Point(-122.4, 37.7),
		Point(-122.4, 37.3),
		Point(-121.8, 37.3),
		Point(-121.8, 37.7),
	).Intersects(Polygon(
		Point(-102.4, 37.7),
		Point(-102.4, 37.3),
		Point(-101.8, 37.3),
		Point(-101.8, 37.7),
	)).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, false)
}

func (s *RethinkSuite) TestGeospatialLineLatLon(c *test.C) {
	var response types.Geometry
	res, err := Line([]float64{-122.423246, 37.779388}, []float64{-121.886420, 37.329898}).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, geometryEquals, "LineString", [][]float64{
		{-122.423246, 37.779388},
		{-121.886420, 37.329898},
	})
}

func (s *RethinkSuite) TestGeospatialLinePoint(c *test.C) {
	var response types.Geometry
	res, err := Line(Point(-122.423246, 37.779388), Point(-121.886420, 37.329898)).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, geometryEquals, "LineString", [][]float64{
		{-122.423246, 37.779388},
		{-121.886420, 37.329898},
	})
}

func (s *RethinkSuite) TestGeospatialPoint(c *test.C) {
	var response types.Geometry
	res, err := Point(-122.423246, 37.779388).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, types.Geometry{
		Type:  "Point",
		Point: types.Point{Lon: -122.423246, Lat: 37.779388},
	})
	c.Assert(response, geometryEquals, "Point", []float64{-122.423246, 37.779388})
}

func (s *RethinkSuite) TestGeospatialPolygon(c *test.C) {
	var response types.Geometry
	res, err := Polygon(Point(-122.423246, 37.779388), Point(-122.423246, 37.329898), Point(-121.886420, 37.329898)).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, geometryEquals, "Polygon", [][][]float64{
		{
			{-122.423246, 37.779388},
			{-122.423246, 37.329898},
			{-121.88642, 37.329898},
			{-122.423246, 37.779388},
		},
	})
}

func (s *RethinkSuite) TestGeospatialPolygonSub(c *test.C) {
	var response types.Geometry
	res, err := Polygon(
		Point(-122.4, 37.7),
		Point(-122.4, 37.3),
		Point(-121.8, 37.3),
		Point(-121.8, 37.7),
	).PolygonSub(Polygon(
		Point(-122.3, 37.4),
		Point(-122.3, 37.6),
		Point(-122.0, 37.6),
		Point(-122.0, 37.4),
	)).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, geometryEquals, "Polygon", [][][]float64{
		{
			{-122.4, 37.7},
			{-122.4, 37.3},
			{-121.8, 37.3},
			{-121.8, 37.7},
			{-122.4, 37.7},
		},
		{
			{-122.3, 37.4},
			{-122.3, 37.6},
			{-122, 37.6},
			{-122, 37.4},
			{-122.3, 37.4},
		},
	})
}
