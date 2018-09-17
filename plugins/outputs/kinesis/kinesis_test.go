package kinesis

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestPartitionKey(t *testing.T) {

	assert := assert.New(t)
	testPoint := testutil.TestMetric(1)

	k := KinesisOutput{
		Partition: &Partition{
			Method: "static",
			Key:    "-",
		},
	}
	assert.Equal("-", k.getPartitionKey(testPoint), "PartitionKey should be '-'")

	k = KinesisOutput{
		Partition: &Partition{
			Method: "tag",
			Key:    "tag1",
		},
	}
	assert.Equal(testPoint.Tags()["tag1"], k.getPartitionKey(testPoint), "PartitionKey should be value of 'tag1'")

	k = KinesisOutput{
		Partition: &Partition{
			Method: "tag",
			Key:    "doesnotexist",
		},
	}
	assert.Equal("", k.getPartitionKey(testPoint), "PartitionKey should be value of ''")

	k = KinesisOutput{
		Partition: &Partition{
			Method: "not supported",
		},
	}
	assert.Equal("", k.getPartitionKey(testPoint), "PartitionKey should be value of ''")

	k = KinesisOutput{
		Partition: &Partition{
			Method: "measurement",
		},
	}
	assert.Equal(testPoint.Name(), k.getPartitionKey(testPoint), "PartitionKey should be value of measurement name")

	k = KinesisOutput{
		Partition: &Partition{
			Method: "random",
		},
	}
	partitionKey := k.getPartitionKey(testPoint)
	u, err := uuid.FromString(partitionKey)
	assert.Nil(err, "Issue parsing UUID")
	assert.Equal(byte(4), u.Version(), "PartitionKey should be UUIDv4")

	k = KinesisOutput{
		PartitionKey: "-",
	}
	assert.Equal("-", k.getPartitionKey(testPoint), "PartitionKey should be '-'")

	k = KinesisOutput{
		RandomPartitionKey: true,
	}
	partitionKey = k.getPartitionKey(testPoint)
	u, err = uuid.FromString(partitionKey)
	assert.Nil(err, "Issue parsing UUID")
	assert.Equal(byte(4), u.Version(), "PartitionKey should be UUIDv4")
}
