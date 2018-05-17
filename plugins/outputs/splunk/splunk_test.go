package splunk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/telegraf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// default config used by Tests
// AuthString == echo -n "uid:pwd" | base64 
func defaultSplunk() *Splunk {
	return &Splunk{
		Prefix:         "testSplunk.",
		Source: 		"",
		SplunkUrl: 		"http://localhost:8088/services/collector",
		AuthString:		"XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX",		
		SimpleFields:    false,
		MetricSeparator: ".",
		ConvertPaths:    true,
		ConvertBool:     true,
		UseRegex:        false,
	}
}

func TestSplunk(t *testing.T){
	// -----------------------------------------------------------------------------------------------------------------
	//  Create a Splunk metric object 
	// -----------------------------------------------------------------------------------------------------------------
	s := defaultSplunk()	
	testMetric, _ := SplunkMetric.New(
		m.UnixNano(),
		"metric",
		"source",
		"testHost",
		map[string]interface{}{"fields":{	"metric_name":"test.metric.name",
				"_value":123456,
				"region":"us-west-1",
				"datacenter":"us-west-1a",
				"rack":"63",
				"os":"Ubuntu16.10"
			}},
	)
	
	// -----------------------------------------------------------------------------------------------------------------
	//  Create a []byte array to send via an HTTP POST
	// -----------------------------------------------------------------------------------------------------------------
	var payload []byte 
	var err error
	payload, err = json.Marshal(splunkMetric)  
	if err != nil {
		return fmt.Errorf("unable to marshal data, %s\n", err.Error())
	}
	fmt.Printf("Sending Payload: %s\n",payload)

	// -----------------------------------------------------------------------------------------------------------------
	//  Send the data to Splunk 
	// -----------------------------------------------------------------------------------------------------------------
	req, err := http.NewRequest("POST", s.SplunkUrl, bytes.NewBuffer(payload) )
	if err != nil {
		return fmt.Errorf("unable to create http.Request \n    URL:%s\n\n", s.SplunkUrl)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization","Splunk " + s.AuthString)
	
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("error POSTing metrics to Splunk, %s\n", req)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 209 {
		return fmt.Errorf("received bad status code posting to %s, %d\n\n%s\n", s.SplunkUrl, resp.StatusCode,payload)
	}
	return nil
}





