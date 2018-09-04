package kazoo

import (
	"reflect"
	"testing"
	"time"
)

func TestCreateDeleteTopic(t *testing.T) {
	tests := []struct {
		name           string
		partitionCount int
		config         map[string]string
		err            error
	}{
		{"test.admin.1", 1, nil, nil},
		{"test.admin.1", 1, nil, ErrTopicExists},
		{"test.admin.2", 1, map[string]string{}, nil},
		{"test.admin.3", 4, map[string]string{"retention.ms": "604800000"}, nil},
		{"test.admin.3", 3, nil, ErrTopicExists},
		{"test.admin.4", 12, map[string]string{"retention.bytes": "1000000000", "retention.ms": "9999999"}, nil},
	}

	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}

	for testIdx, test := range tests {
		err = kz.CreateTopic(test.name, test.partitionCount, 1, test.config)
		if err != test.err {
			t.Errorf("Unexpected error (%v) creating %s for test %d", err, test.name, testIdx)
			continue
		}
		if err == nil {
			topic := kz.Topic(test.name)
			conf, err := topic.Config()
			if err != nil {
				t.Errorf("Unable to get topic config (%v) for %s for test %d", err, test.name, testIdx)
			}
			// allow for nil == empty map
			if !reflect.DeepEqual(conf, test.config) && !(test.config == nil && len(conf) == 0) {
				t.Errorf("Invalid config for %s in test %d. Expected (%v) got (%v)", test.name, testIdx, conf, test.config)
			}
		}

	}

	// delete all test topics
	topicMap := make(map[string]bool)
	for _, test := range tests {
		// delete if we haven't seen the topic before
		if _, ok := topicMap[test.name]; !ok {
			err := kz.DeleteTopic(test.name)
			if err != nil {
				t.Errorf("Unable to delete topic %s (%v)", test.name, err)
			}
		}
		topicMap[test.name] = true
	}

	totalToDelete := len(topicMap)

	// wait for deletion (up to 60s)
	for i := 0; i < 15; i++ {
		for name := range topicMap {
			topic := &Topic{kz: kz, Name: name}
			if exists, _ := topic.Exists(); !exists {
				delete(topicMap, name)
			}
		}
		// all topics deleted
		if len(topicMap) == 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if len(topicMap) != 0 {
		t.Errorf("Unable to delete all topics %d out of %d remaining after 15 seconds", len(topicMap), totalToDelete)
	}
}

func TestDeleteTopicSync(t *testing.T) {

	kz, err := NewKazoo(zookeeperPeers, nil)

	topicName := "test.admin.1"

	if err != nil {
		t.Fatal(err)
	}

	err = kz.CreateTopic(topicName, 1, 1, nil)

	if err != nil {
		t.Errorf("Unexpected error (%v) creating topic %s", err, topicName)
	}

	topic := kz.Topic("test.admin.1")
	_, err = topic.Config()

	if err != nil {
		t.Errorf("Unable to get topic config (%v) for %s", err, topicName)
	}

	// delete the topic synchronously
	err = kz.DeleteTopicSync(topicName, 0)

	if err != nil {
		t.Errorf("Unexpected error (%v) while deleting topic synchronously", err)
	}

	exists, err := topic.Exists()

	if err != nil {
		t.Errorf("Unexpected error (%v) while checking if topic exists", err)
	}

	if exists {
		t.Error("Deleted topic still exists.")
	}
}
