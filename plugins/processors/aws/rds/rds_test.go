package rds

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

type mockRdsClient struct {
	DbIdentifier string
}

func (m *mockRdsClient) DescribeDBInstances(input *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error) {
	instances := []*rds.DBInstance{}
	if *input.DBInstanceIdentifier == m.DbIdentifier {
		instances = append(instances, &rds.DBInstance{
			DBInstanceIdentifier: input.DBInstanceIdentifier,
			Engine:               aws.String("mysql"),
			EngineVersion:        aws.String("5.6"),
			DBInstanceArn:        aws.String("arn:aws:rds:us-east-1:1111111:db:rds-test-instance"),
		})
	}
	return &rds.DescribeDBInstancesOutput{
		DBInstances: instances,
	}, nil
}

func (m *mockRdsClient) ListTagsForResource(input *rds.ListTagsForResourceInput) (*rds.ListTagsForResourceOutput, error) {
	tags := []*rds.Tag{}
	if *input.ResourceName == "arn:aws:rds:us-east-1:1111111:db:rds-test-instance" {
		tags = append(tags, &rds.Tag{
			Key:   aws.String("Environment"),
			Value: aws.String("acc-test"),
		})
		tags = append(tags, &rds.Tag{
			Key:   aws.String("not-included"),
			Value: aws.String("true"),
		})
	}
	return &rds.ListTagsForResourceOutput{
		TagList: tags,
	}, nil
}

func cache(client RDSClient, ttl time.Duration) RDSClient {
	return &CachingRDSClient{
		client: client,
		ttl:    ttl,
	}
}

func TestProcess_basic(t *testing.T) {
	p := &RDS{
		MetricNames: []string{
			"cloudwatch_aws_rds",
		},
		Id:            "db_instance_identifier",
		Engine:        true,
		EngineVersion: true,
		Tags:          []string{"Environment"},
		client: cache(&mockRdsClient{
			DbIdentifier: "rds-test-instance",
		}, time.Duration(15*time.Minute)),
	}
	metric, _ := metric.New("cloudwatch_aws_rds",
		map[string]string{
			"db_instance_identifier": "rds-test-instance",
		},
		map[string]interface{}{
			"count": 1,
		},
		time.Now())
	processedMetrics := p.Apply(metric)
	assert.Equal(t, 1, len(processedMetrics))

	pm := processedMetrics[0]
	assert.Equal(t, "mysql", pm.Tags()["engine"])
	assert.Equal(t, "5.6", pm.Tags()["engine_version"])
	assert.Equal(t, "acc-test", pm.Tags()["Environment"])
}

func TestProcess_missingTag(t *testing.T) {
	p := &RDS{
		MetricNames: []string{
			"cloudwatch_aws_rds",
		},
		Id:            "db_instance_identifier",
		Engine:        false,
		EngineVersion: false,
		Tags:          []string{"Name"},
		client: cache(&mockRdsClient{
			DbIdentifier: "rds-test-instance",
		}, time.Duration(15*time.Minute)),
	}
	metric, _ := metric.New("cloudwatch_aws_rds",
		map[string]string{
			"db_instance_identifier": "rds-test-instance",
		},
		map[string]interface{}{
			"count": 1,
		},
		time.Now())
	processedMetrics := p.Apply(metric)
	assert.Equal(t, 1, len(processedMetrics))

	pm := processedMetrics[0]
	assert.Equal(t, "", pm.Tags()["engine"])
	assert.Equal(t, "", pm.Tags()["engine_version"])
	assert.Equal(t, "", pm.Tags()["Name"])
}

func TestProcess_missingInstance(t *testing.T) {
	p := &RDS{
		MetricNames: []string{
			"cloudwatch_aws_rds",
		},
		Id:            "db_instance_identifier",
		Engine:        true,
		EngineVersion: true,
		Tags:          []string{"Environment"},
		client: cache(&mockRdsClient{
			DbIdentifier: "rds-test-instance2",
		}, time.Duration(15*time.Minute)),
	}
	metric, _ := metric.New("cloudwatch_aws_rds",
		map[string]string{
			"db_instance_identifier": "rds-test-instance",
		},
		map[string]interface{}{
			"count": 1,
		},
		time.Now())
	processedMetrics := p.Apply(metric)
	assert.Equal(t, 1, len(processedMetrics))

	pm := processedMetrics[0]
	assert.Equal(t, "", pm.Tags()["engine"])
	assert.Equal(t, "", pm.Tags()["engine_version"])
	assert.Equal(t, "", pm.Tags()["Environment"])
}
