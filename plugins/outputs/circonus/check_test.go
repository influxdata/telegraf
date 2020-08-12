// Package circonus contains the output plugin used to write metric data to a
// Circonus broker.
package circonus

import (
	"testing"

	apiclient "github.com/circonus-labs/go-apiclient"
	apiconfig "github.com/circonus-labs/go-apiclient/config"
)

func stringPtr(s string) *string {
	return &s
}

type mockAPIClient struct{}

func (m *mockAPIClient) Get(reqPath string) ([]byte, error) {
	return []byte("test"), nil
}

func (m *mockAPIClient) FetchBrokers() (*[]apiclient.Broker, error) {
	return &[]apiclient.Broker{{
		Name: "test",
		CID:  "/broker/1",
		Details: []apiclient.BrokerDetail{{
			IP:      stringPtr("1.1.1.1"),
			CN:      "test.com",
			Modules: []string{"httptrap"},
		}},
	}}, nil
}

func (m *mockAPIClient) FetchBroker(cid apiclient.CIDType) (*apiclient.Broker, error) {
	return &apiclient.Broker{
		Name: "test",
		CID:  "/broker/1",
		Details: []apiclient.BrokerDetail{{
			IP:      stringPtr("1.1.1.1"),
			CN:      "test.com",
			Modules: []string{"httptrap"},
		}},
	}, nil
}

func (m *mockAPIClient) SearchCheckBundles(searchCriteria *apiclient.SearchQueryType,
	filterCriteria *apiclient.SearchFilterType) (*[]apiclient.CheckBundle, error) {
	return &[]apiclient.CheckBundle{{
		Config: apiclient.CheckBundleConfig{
			apiconfig.Key("submission_url"): "http://test.com:1234",
		},
		CID: "/check_bundle/1",
	}}, nil
}

func (m *mockAPIClient) CreateCheckBundle(
	cfg *apiclient.CheckBundle) (*apiclient.CheckBundle, error) {
	return &apiclient.CheckBundle{
		Brokers: []string{"/broker/1"},
		Config: apiclient.CheckBundleConfig{
			apiconfig.Key("asynch_metrics"): "true",
			apiconfig.Key("secret"):         "test",
			apiconfig.Key("submission_url"): "http://test.com:1234",
		},
		DisplayName:   "telegraf-httptrap",
		MetricFilters: [][]string{{"allow", ".", "default"}},
		Period:        60,
		Status:        "active",
		Tags:          []string{"source:telegraf"},
		Target:        "localhost",
		Timeout:       10,
		Type:          "httptrap",
	}, nil
}

func TestGetRandomBroker(t *testing.T) {
	c := &Circonus{}
	c.api = &mockAPIClient{}

	exp := "/broker/1"
	expCN := "test.com"

	v, vcn, err := c.getRandomBroker()
	if err != nil {
		t.Error(err)
	}

	if v != exp {
		t.Errorf("expected CID: %v, got: %v", exp, v)
	}

	if vcn != expCN {
		t.Errorf("expected CN: %v, got: %v", expCN, vcn)
	}
}

func TestGetBrokerInfo(t *testing.T) {
	c := &Circonus{}
	c.api = &mockAPIClient{}

	expCN := "test.com"

	vcn, err := c.getBrokerInfo("test")
	if err != nil {
		t.Error(err)
	}

	if vcn != expCN {
		t.Errorf("expected CN: %v, got: %v", expCN, vcn)
	}
}

func TestGetCheckSubmissionURL(t *testing.T) {
	c := &Circonus{}
	c.api = &mockAPIClient{}

	exp := "http://test.com:1234"

	v, err := c.getCheckSubmissionURL("test", "test.com")
	if err != nil {
		t.Error(err)
	}

	if v != exp {
		t.Errorf("expected URL: %v, got: %v", exp, v)
	}
}

func TestCreateCheck(t *testing.T) {
	c := &Circonus{}
	c.api = &mockAPIClient{}

	exp := "http://test.com:1234"

	v, err := c.createCheck("test", "test.com")
	if err != nil {
		t.Error(err)
	}

	if v != exp {
		t.Errorf("expected URL: %v, got: %v", exp, v)
	}
}

func TestGetSubmissionURL(t *testing.T) {
	c := &Circonus{}
	c.api = &mockAPIClient{}

	exp := "http://test.com:1234"

	v, err := c.getSubmissionURL()
	if err != nil {
		t.Error(err)
	}

	if v != exp {
		t.Errorf("expected URL: %v, got: %v", exp, v)
	}
}
