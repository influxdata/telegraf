package ec2

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

type mockEc2Client struct {
	InstanceId string
}

func (m *mockEc2Client) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	reservations := []*ec2.Reservation{}
	if *input.InstanceIds[0] == m.InstanceId {
		reservations = append(reservations, &ec2.Reservation{
			Instances: []*ec2.Instance{
				&ec2.Instance{
					InstanceType: aws.String("t2.micro"),
					ImageId:      aws.String("ami-12345"),
					Tags: []*ec2.Tag{
						&ec2.Tag{
							Key:   aws.String("Environment"),
							Value: aws.String("acc-test"),
						},
						&ec2.Tag{
							Key:   aws.String("not-included"),
							Value: aws.String("true"),
						},
					},
				},
			},
		})
	}
	return &ec2.DescribeInstancesOutput{
		Reservations: reservations,
	}, nil
}

func cache(client EC2Client, ttl time.Duration) EC2Client {
	return &CachingEC2Client{
		client: client,
		ttl:    ttl,
	}
}

func TestProcess_basic(t *testing.T) {
	p := &EC2{
		MetricNames: []string{
			"cloudwatch_aws_ec2",
		},
		Id:           "instance_id",
		InstanceType: true,
		AmiId:        true,
		Tags:         []string{"Environment"},
		client: cache(&mockEc2Client{
			InstanceId: "i-123abc",
		}, time.Duration(15*time.Minute)),
	}
	metric, _ := metric.New("cloudwatch_aws_ec2",
		map[string]string{
			"instance_id": "i-123abc",
		},
		map[string]interface{}{
			"count": 1,
		},
		time.Now())
	processedMetrics := p.Apply(metric)
	assert.Equal(t, 1, len(processedMetrics))

	pm := processedMetrics[0]
	assert.Equal(t, "t2.micro", pm.Tags()["instance_type"])
	assert.Equal(t, "ami-12345", pm.Tags()["ami_id"])
	assert.Equal(t, "acc-test", pm.Tags()["Environment"])
}

func TestProcess_missingTag(t *testing.T) {
	p := &EC2{
		MetricNames: []string{
			"cloudwatch_aws_ec2",
		},
		Id:           "instance_id",
		InstanceType: false,
		AmiId:        false,
		Tags:         []string{"Name"},
		client: cache(&mockEc2Client{
			InstanceId: "i-123abc",
		}, time.Duration(15*time.Minute)),
	}
	metric, _ := metric.New("cloudwatch_aws_ec2",
		map[string]string{
			"instance_id": "i-123abc",
		},
		map[string]interface{}{
			"count": 1,
		},
		time.Now())
	processedMetrics := p.Apply(metric)
	assert.Equal(t, 1, len(processedMetrics))

	pm := processedMetrics[0]
	assert.Equal(t, "", pm.Tags()["instance_type"])
	assert.Equal(t, "", pm.Tags()["ami_id"])
	assert.Equal(t, "", pm.Tags()["Name"])
}

func TestProcess_missingInstance(t *testing.T) {
	p := &EC2{
		MetricNames: []string{
			"cloudwatch_aws_ec2",
		},
		Id:           "instance_id",
		InstanceType: true,
		AmiId:        true,
		Tags:         []string{"Environment"},
		client: cache(&mockEc2Client{
			InstanceId: "i-xyz987",
		}, time.Duration(15*time.Minute)),
	}
	metric, _ := metric.New("cloudwatch_aws_ec2",
		map[string]string{
			"instance_id": "i-123abc",
		},
		map[string]interface{}{
			"count": 1,
		},
		time.Now())
	processedMetrics := p.Apply(metric)
	assert.Equal(t, 1, len(processedMetrics))

	pm := processedMetrics[0]
	assert.Equal(t, "", pm.Tags()["instance_type"])
	assert.Equal(t, "", pm.Tags()["ami_id"])
	assert.Equal(t, "", pm.Tags()["Environment"])
}
