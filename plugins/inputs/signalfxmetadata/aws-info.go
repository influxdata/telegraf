package signalfxmetadata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// NewAWSInfo - returns a new AWSInfo context
func NewAWSInfo() *AWSInfo {
	return &AWSInfo{false, false, ""}
}

// AWSInfo - stores information about the aws instance
type AWSInfo struct {
	aws         bool
	awsSet      bool
	awsUniqueID string
}

// GetAWSInfo - adds aws metadata to the supplied map
func (s *AWSInfo) GetAWSInfo() (info map[string]string) {
	info = map[string]string{}
	if identity, err := requestAWSInfo(); err == nil {
		processAWSInfo(info, identity)
		s.aws = true
		// build aws unique id
		if !s.awsSet {
			s.awsUniqueID, s.awsSet = buildAWSUniqueID(info)
		}
		// set aws unique id property
		if s.awsSet {
			info["AWSUniqueId"] = s.awsUniqueID
		}
		log.Println("I! is an aws box")
	} else {
		log.Println("I! not an aws box")
	}
	return
}

func buildAWSUniqueID(info map[string]string) (awsUniqueID string, awsSet bool) {
	if id, ok := info["aws_instance_id"]; ok {
		if region, ok := info["aws_region"]; ok {
			if account, ok := info["aws_account_id"]; ok {
				awsUniqueID = fmt.Sprintf("%s_%s_%s", id, region, account)
				awsSet = true
			}
		}
	}
	return
}

func processAWSInfo(info map[string]string, identity map[string]interface{}) {
	var want = map[string]string{
		"avaialbility_zone": "availabilityZone",
		"instance_type":     "instanceType",
		"instance_id":       "instanceId",
		"image_id":          "imageId",
		"account_id":        "accountId",
		"region":            "region",
		"architecture":      "architecture",
	}
	// extract desired metadata
	for k, v := range want {
		// if a value exists add it to the host info
		if val, ok := identity[v]; ok {
			info[fmt.Sprintf("aws_%s", k)] = val.(string)
		}
	}
}

func requestAWSInfo() (identity map[string]interface{}, err error) {
	var url = "http://169.254.169.254/latest/dynamic/instance-identity/document"
	var httpClient = &http.Client{Timeout: 200 * time.Millisecond}
	var raw []byte
	var res *http.Response

	// make the request
	res, err = httpClient.Get(url)
	if err == nil {
		// read the response
		raw, err = ioutil.ReadAll(res.Body)
	}
	if err == nil {
		// parse the json response
		err = json.Unmarshal(raw, &identity)
	}
	return
}
