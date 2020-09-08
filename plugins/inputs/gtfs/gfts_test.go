package gtfs_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	gtfsbindings "github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/inputs/gtfs"
	"github.com/influxdata/telegraf/testutil"
)

func TestGTFS(t *testing.T) {
	var (
		testID                  = "123"
		testStopID              = "stop123"
		testTripID              = "trip123"
		testVehicleID           = "vehicle123"
		testRouteID             = "route123"
		testDirectionID         = uint32(123)
		testTimestamp           = uint64(100)
		testLat                 = float32(42.42)
		testLng                 = float32(20.20)
		testBearing             = float32(270)
		testOdometer            = float64(1000)
		testSpeed               = float32(60)
		testGtfsRealtimeVersion = "v0.0.0"
		testHeader              = &gtfsbindings.FeedHeader{GtfsRealtimeVersion: &testGtfsRealtimeVersion}
	)

	for _, test := range []struct {
		name               string
		measurement        string
		expectedTagCount   int
		expectedFieldCount int
		responses          map[string]*gtfsbindings.FeedMessage
	}{
		{
			name:               "vehicle positions",
			measurement:        "position",
			expectedTagCount:   2,
			expectedFieldCount: 12,
			responses: map[string]*gtfsbindings.FeedMessage{
				"/VehiclePositions.pb": {
					Header: testHeader, // required
					Entity: []*gtfsbindings.FeedEntity{
						{
							Id: &testID,
							Vehicle: &gtfsbindings.VehiclePosition{
								Trip: &gtfsbindings.TripDescriptor{
									TripId:      &testTripID,
									RouteId:     &testRouteID,
									DirectionId: &testDirectionID,
								},
								Vehicle: &gtfsbindings.VehicleDescriptor{
									Id: &testVehicleID,
								},
								Position: &gtfsbindings.Position{
									Latitude:  &testLat,
									Longitude: &testLng,
									Bearing:   &testBearing,
									Odometer:  &testOdometer,
									Speed:     &testSpeed,
								},
								StopId:    &testStopID,
								Timestamp: &testTimestamp,
							},
						},
					},
				},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := newTestServer(t, test.responses)
			defer server.Close()

			subject := &gtfs.GTFS{}
			if test.measurement == "position" {
				subject.VehiclePositionsURL = server.URL + "/VehiclePositions.pb"
			}
			if test.measurement == "update" {
				subject.TripUpdatesURL = server.URL + "/TripUpdates.pb"
			}
			if test.measurement == "alert" {
				subject.ServiceAlertsURL = server.URL + "/ServiceAlerts.pb"
			}

			require.NoError(t, subject.Init())

			var acc testutil.Accumulator
			require.NoError(t, acc.GatherError(subject.Gather))

			require.Len(t, acc.Metrics, 1)
			var metric = acc.Metrics[0]
			require.Equal(t, metric.Measurement, test.measurement)
			require.Len(t, metric.Tags, 2)
			require.Len(t, metric.Fields, 12)
		})
	}
}

func newTestServer(t *testing.T, responses map[string]*gtfsbindings.FeedMessage) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response, found := responses[r.URL.Path]
		if !found {
			t.Fatal("not found")
		}

		b, err := proto.Marshal(response)
		if err != nil {
			t.Fatal("marshaling response")
		}

		if _, err := w.Write(b); err != nil {
			t.Fatal("writing response")
		}
	}))
}
