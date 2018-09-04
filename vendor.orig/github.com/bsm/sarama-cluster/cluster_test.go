package cluster

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Shopify/sarama"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	testGroup     = "sarama-cluster-group"
	testKafkaData = "/tmp/sarama-cluster-test"
)

var (
	testKafkaRoot  = "kafka_2.12-1.0.0"
	testKafkaAddrs = []string{"127.0.0.1:29092"}
	testTopics     = []string{"topic-a", "topic-b"}

	testClient              sarama.Client
	testKafkaCmd, testZkCmd *exec.Cmd
)

func init() {
	if dir := os.Getenv("KAFKA_DIR"); dir != "" {
		testKafkaRoot = dir
	}
}

var _ = Describe("offsetInfo", func() {

	It("should calculate next offset", func() {
		Expect(offsetInfo{-2, ""}.NextOffset(sarama.OffsetOldest)).To(Equal(sarama.OffsetOldest))
		Expect(offsetInfo{-2, ""}.NextOffset(sarama.OffsetNewest)).To(Equal(sarama.OffsetNewest))
		Expect(offsetInfo{-1, ""}.NextOffset(sarama.OffsetOldest)).To(Equal(sarama.OffsetOldest))
		Expect(offsetInfo{-1, ""}.NextOffset(sarama.OffsetNewest)).To(Equal(sarama.OffsetNewest))
		Expect(offsetInfo{0, ""}.NextOffset(sarama.OffsetOldest)).To(Equal(int64(0)))
		Expect(offsetInfo{100, ""}.NextOffset(sarama.OffsetOldest)).To(Equal(int64(100)))
	})

})

var _ = Describe("int32Slice", func() {

	It("should diff", func() {
		Expect(((int32Slice)(nil)).Diff(int32Slice{1, 3, 5})).To(BeNil())
		Expect(int32Slice{1, 3, 5}.Diff((int32Slice)(nil))).To(Equal([]int32{1, 3, 5}))
		Expect(int32Slice{1, 3, 5}.Diff(int32Slice{1, 3, 5})).To(BeNil())
		Expect(int32Slice{1, 3, 5}.Diff(int32Slice{1, 2, 3, 4, 5})).To(BeNil())
		Expect(int32Slice{1, 3, 5}.Diff(int32Slice{2, 3, 4})).To(Equal([]int32{1, 5}))
		Expect(int32Slice{1, 3, 5}.Diff(int32Slice{1, 4})).To(Equal([]int32{3, 5}))
		Expect(int32Slice{1, 3, 5}.Diff(int32Slice{2, 5})).To(Equal([]int32{1, 3}))
	})

})

// --------------------------------------------------------------------

var _ = BeforeSuite(func() {
	testZkCmd = testCmd(
		testDataDir(testKafkaRoot, "bin", "kafka-run-class.sh"),
		"org.apache.zookeeper.server.quorum.QuorumPeerMain",
		testDataDir("zookeeper.properties"),
	)

	testKafkaCmd = testCmd(
		testDataDir(testKafkaRoot, "bin", "kafka-run-class.sh"),
		"-name", "kafkaServer", "kafka.Kafka",
		testDataDir("server.properties"),
	)

	// Remove old test data before starting
	Expect(os.RemoveAll(testKafkaData)).NotTo(HaveOccurred())

	Expect(os.MkdirAll(testKafkaData, 0777)).To(Succeed())
	Expect(testZkCmd.Start()).To(Succeed())
	Expect(testKafkaCmd.Start()).To(Succeed())

	// Wait for client
	Eventually(func() error {
		var err error

		// sync-producer requires Return.Successes set to true
		testConf := sarama.NewConfig()
		testConf.Producer.Return.Successes = true
		testClient, err = sarama.NewClient(testKafkaAddrs, testConf)
		return err
	}, "30s", "1s").Should(Succeed())

	// Ensure we can retrieve partition info
	Eventually(func() error {
		_, err := testClient.Partitions(testTopics[0])
		return err
	}, "30s", "1s").Should(Succeed())

	// Seed a few messages
	Expect(testSeed(1000, testTopics)).To(Succeed())
})

var _ = AfterSuite(func() {
	if testClient != nil {
		_ = testClient.Close()
	}

	_ = testKafkaCmd.Process.Kill()
	_ = testZkCmd.Process.Kill()
	_ = testKafkaCmd.Wait()
	_ = testZkCmd.Wait()
	_ = os.RemoveAll(testKafkaData)
})

// --------------------------------------------------------------------

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "sarama/cluster")
}

func testDataDir(tokens ...string) string {
	tokens = append([]string{"testdata"}, tokens...)
	return filepath.Join(tokens...)
}

func testSeed(n int, testTopics []string) error {
	producer, err := sarama.NewSyncProducerFromClient(testClient)
	if err != nil {
		return err
	}
	defer producer.Close()

	for i := 0; i < n; i++ {
		kv := sarama.StringEncoder(fmt.Sprintf("PLAINDATA-%08d", i))
		for _, t := range testTopics {
			msg := &sarama.ProducerMessage{Topic: t, Key: kv, Value: kv}
			if _, _, err := producer.SendMessage(msg); err != nil {
				return err
			}
		}
	}
	return nil
}

func testCmd(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	if testing.Verbose() || os.Getenv("CI") != "" {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}
	cmd.Env = []string{"KAFKA_HEAP_OPTS=-Xmx1G -Xms1G"}
	return cmd
}

type testConsumerMessage struct {
	sarama.ConsumerMessage
	ConsumerID string
}

// --------------------------------------------------------------------

var _ sarama.Consumer = &mockConsumer{}
var _ sarama.PartitionConsumer = &mockPartitionConsumer{}

type mockClient struct {
	sarama.Client

	topics map[string][]int32
}
type mockConsumer struct{ sarama.Consumer }
type mockPartitionConsumer struct {
	sarama.PartitionConsumer

	Topic     string
	Partition int32
	Offset    int64
}

func (m *mockClient) Partitions(t string) ([]int32, error) {
	pts, ok := m.topics[t]
	if !ok {
		return nil, sarama.ErrInvalidTopic
	}
	return pts, nil
}

func (*mockConsumer) ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	if offset > -1 && offset < 1000 {
		return nil, sarama.ErrOffsetOutOfRange
	}
	return &mockPartitionConsumer{
		Topic:     topic,
		Partition: partition,
		Offset:    offset,
	}, nil
}

func (*mockPartitionConsumer) Close() error { return nil }
