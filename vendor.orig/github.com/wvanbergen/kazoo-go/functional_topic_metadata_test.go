package kazoo

import (
	"testing"
)

func TestTopics(t *testing.T) {
	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}

	topics, err := kz.Topics()
	if err != nil {
		t.Error(err)
	}

	existingTopic := topics.Find("test.4")
	if existingTopic == nil {
		t.Error("Expected topic test.4 to be returned")
	} else if existingTopic.Name != "test.4" {
		t.Error("Expected topic test.4 to have its name set")
	}

	nonexistingTopic := topics.Find("__nonexistent__")
	if nonexistingTopic != nil {
		t.Error("Expected __nonexistent__ topic to not be defined")
	}

	assertSuccessfulClose(t, kz)
}

func TestTopicPartitions(t *testing.T) {
	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}

	partitions, err := kz.Topic("test.4").Partitions()
	if err != nil {
		t.Fatal(err)
	}

	if len(partitions) != 4 {
		t.Errorf("Expected test.4 to have 4 partitions")
	}

	brokers, err := kz.Brokers()
	if err != nil {
		t.Fatal(err)
	}

	for index, partition := range partitions {
		if partition.ID != int32(index) {
			t.Error("partition.ID is not set properly")
		}

		leader, err := partition.Leader()
		if err != nil {
			t.Fatal(err)
		}

		if _, ok := brokers[leader]; !ok {
			t.Errorf("Expected the leader of test.4/%d to be an existing broker.", partition.ID)
		}

		isr, err := partition.ISR()
		if err != nil {
			t.Fatal(err)
		}

		for _, brokerID := range isr {
			if _, ok := brokers[brokerID]; !ok {
				t.Errorf("Expected all ISRs of test.4/%d to be existing brokers.", partition.ID)
			}
		}
	}

	assertSuccessfulClose(t, kz)
}

func TestTopicConfig(t *testing.T) {
	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}

	topicConfig, err := kz.Topic("test.4").Config()
	if err != nil {
		t.Error(err)
	}
	if topicConfig["retention.ms"] != "604800000" {
		t.Error("Expected retention.ms config for test.4 to be set to 604800000")
	}

	topicConfig, err = kz.Topic("test.1").Config()
	if err != nil {
		t.Error(err)
	}
	if len(topicConfig) > 0 {
		t.Error("Expected no topic level configuration to be set for test.1")
	}

	assertSuccessfulClose(t, kz)
}

func assertSuccessfulClose(t *testing.T, kz *Kazoo) {
	if err := kz.Close(); err != nil {
		t.Error(err)
	}
}
