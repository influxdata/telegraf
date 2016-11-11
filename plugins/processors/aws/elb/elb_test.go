package elb

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

type mockElbClient struct {
	Name string
}

func (m *mockElbClient) DescribeTags(input *elb.DescribeTagsInput) (*elb.DescribeTagsOutput, error) {
	descriptions := []*elb.TagDescription{}
	if *input.LoadBalancerNames[0] == m.Name {
		descriptions = append(descriptions, &elb.TagDescription{
			LoadBalancerName: &m.Name,
			Tags: []*elb.Tag{
				&elb.Tag{
					Key:   aws.String("Environment"),
					Value: aws.String("acc-test"),
				},
			},
		})
	}
	return &elb.DescribeTagsOutput{
		TagDescriptions: descriptions,
	}, nil
}

func cache(client ELBClient, ttl time.Duration) ELBClient {
	return &CachingELBClient{
		client: client,
		ttl:    ttl,
	}
}

func TestProcess_basic(t *testing.T) {
	p := &ELB{
		MetricNames: []string{
			"cloudwatch_aws_elb",
		},
		Id:   "load_balancer_name",
		Tags: []string{"Environment"},
		client: cache(&mockElbClient{
			Name: "acc-test-lb",
		}, time.Duration(15*time.Minute)),
	}
	metric, _ := metric.New("cloudwatch_aws_elb",
		map[string]string{
			"load_balancer_name": "acc-test-lb",
		},
		map[string]interface{}{
			"count": 1,
		},
		time.Now())
	processedMetrics := p.Apply(metric)
	assert.Equal(t, 1, len(processedMetrics))

	pm := processedMetrics[0]
	assert.Equal(t, "acc-test", pm.Tags()["Environment"])
}

func TestProcess_missingTag(t *testing.T) {
	p := &ELB{
		MetricNames: []string{
			"cloudwatch_aws_elb",
		},
		Id:   "load_balancer_name",
		Tags: []string{"Name"},
		client: cache(&mockElbClient{
			Name: "acc-test-lb",
		}, time.Duration(15*time.Minute)),
	}
	metric, _ := metric.New("cloudwatch_aws_elb",
		map[string]string{
			"load_balancer_name": "acc-test-lb",
		},
		map[string]interface{}{
			"count": 1,
		},
		time.Now())
	processedMetrics := p.Apply(metric)
	assert.Equal(t, 1, len(processedMetrics))

	pm := processedMetrics[0]
	assert.Equal(t, "", pm.Tags()["Name"])
}

func TestProcess_missingInstance(t *testing.T) {
	p := &ELB{
		MetricNames: []string{
			"cloudwatch_aws_elb",
		},
		Id:   "load_balancer_name",
		Tags: []string{"Environment"},
		client: cache(&mockElbClient{
			Name: "acc-test-lb",
		}, time.Duration(15*time.Minute)),
	}
	metric, _ := metric.New("cloudwatch_aws_elb",
		map[string]string{
			"load_balancer_name": "unknown-lb",
		},
		map[string]interface{}{
			"count": 1,
		},
		time.Now())
	processedMetrics := p.Apply(metric)
	assert.Equal(t, 1, len(processedMetrics))

	pm := processedMetrics[0]
	assert.Equal(t, "", pm.Tags()["Environment"])
}
