package stackdriver

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"

	// Imports the Stackdriver Monitoring client package.
	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/proto"
	"google.golang.org/api/option"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/grpc"
)

// clientOpt is the option tests should use to connect to the test server.
// It is initialized by TestMain.
var clientOpt option.ClientOption

var (
	mockGroup  mockGroupServer
	mockMetric mockMetricServer
)

type mockGroupServer struct {
	// Embed for forward compatibility.
	// Tests will keep working if more methods are added
	// in the future.
	monitoringpb.GroupServiceServer

	reqs []proto.Message

	// If set, all calls return this error.
	err error

	// responses to return if err == nil
	resps []proto.Message
}

type mockMetricServer struct {
	// Embed for forward compatibility.
	// Tests will keep working if more methods are added
	// in the future.
	monitoringpb.MetricServiceServer

	reqs []proto.Message

	// If set, all calls return this error.
	err error

	// responses to return if err == nil
	resps []proto.Message
}

func TestMain(m *testing.M) {
	serv := grpc.NewServer()
	monitoringpb.RegisterGroupServiceServer(serv, &mockGroup)
	monitoringpb.RegisterMetricServiceServer(serv, &mockMetric)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatal(err)
	}
	go serv.Serve(lis)

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	clientOpt = option.WithGRPCConn(conn)
}

func TestWrite(t *testing.T) {
	c, err := monitoring.NewMetricClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	s := &Stackdriver{
		Project:   fmt.Sprintf("projects/%s", "[PROJECT]"),
		Namespace: "test",
		client:    c,
	}

	err = s.Connect()
	require.NoError(t, err)
	err = s.Write(testutil.MockMetrics())
	require.NoError(t, err)
}

func TestGetStackdriverLabels(t *testing.T) {
	tags := map[string]string{
		"project":   "bar",
		"discuss":   "revolutionary",
		"marble":    "discount",
		"applied":   "falsify",
		"test":      "foo",
		"porter":    "discount",
		"play":      "tiger",
		"fireplace": "display",
		"host":      "this",
		"name":      "bat",
		"device":    "local",
		"reserve":   "publication",
		"xpfqacltlmpguimhtjlou2qlmf9uqqwk3teajwlwqkoxtsppbnjksaxvzc1aa973pho9m96gfnl5op8ku7sv93rexyx42qe3zty12ityv": "keyquota",
		"valuequota": "icym5wcpejnhljcvy2vwk15svmhrtueoppwlvix61vlbaeedufn1g6u4jgwjoekwew9s2dboxtgrkiyuircnl8h1lbzntt9gzcf60qunhxurhiz0g2bynzy1v6eyn4ravndeiiugobsrsj2bfaguahg4gxn7nx4irwfknunhkk6jdlldevawj8levebjajcrcbeugewd14fa8o34ycfwx2ymalyeqxhfqrsksxnii2deqq6cghrzi6qzwmittkzdtye3imoygqmjjshiskvnzz1e4ipd9c6wfor5jsygn1kvcg6jm4clnsl1fnxotbei9xp4swrkjpgursmfmkyvxcgq9hoy435nwnolo3ipnvdlhk6pmlzpdjn6gqi3v9gv7jn5ro2p1t5ufxzfsvqq1fyrgoi7gvmttil1banh3cftkph1dcoaqfhl7y0wkvhwwvrmslmmxp1wedyn8bacd7akmjgfwdvcmrymbzvmrzfvq1gs1xnmmg8rsfxci2h6r1ralo3splf4f3bdg4c7cy0yy9qbxzxhcmdpwekwc7tdjs8uj6wmofm2aor4hum8nwyfwwlxy3yvsnbjy32oucsrmhcnu6l2i8laujkrhvsr9fcix5jflygznlydbqw5uhw1rg1g5wiihqumwmqgggemzoaivm3ut41vjaff4uqtqyuhuwblmuiphfkd7si49vgeeswzg7tpuw0oxmkesgibkcjtev2h9ouxzjs3eb71jffhdacyiuyhuxwvm5bnrjewbm4x2kmhgbirz3eoj7ijgplggdkx5vixufg65ont8zi1jabsuxx0vsqgprunwkugqkxg2r7iy6fmgs4lob4dlseinowkst6gp6x1ejreauyzjz7atzm3hbmr5rbynuqp4lxrnhhcbuoun69mavvaaki0bdz5ybmbbbz5qdv0odtpjo2aezat5uosjuhzbvic05jlyclikynjgfhencdkz3qcqzbzhnsynj1zdke0sk4zfpvfyryzsxv9pu0qm",
	}
	labels := getStackdriverLabels(tags)
	require.Equal(t, len(labels), QuotaLabelsPerMetricDescriptor)
}
