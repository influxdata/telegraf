package jboss

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

type BodyContent struct {
	Operation      string                   `json:"operation"`
	Name           string                   `json:"name"`
	IncludeRuntime string                   `json:"include-runtime"`
	AttributesOnly string                   `json:"attributes-only"`
	ChildType      string                   `json:"child-type"`
	RecursiveDepth int                      `json:"recursive-depth"`
	Recursive      string                   `json:"recursive"`
	Address        []map[string]interface{} `json:"address"`
	JsonPretty     int                      `json:"json.pretty"`
}

func readJson(jsonFilePath string) string {
	data, err := ioutil.ReadFile(jsonFilePath)

	if err != nil {
		panic(fmt.Sprintf("could not read from data file %s", jsonFilePath))
	}
	return string(data)
}

func testJBossServer(t *testing.T, eap7 bool) *httptest.Server {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//		fmt.Println("--------------------INIT REQUEST------------------------------------------")
		w.WriteHeader(http.StatusOK)
		decoder := json.NewDecoder(r.Body)
		var b BodyContent
		err := decoder.Decode(&b)
		if err != nil {
			fmt.Printf("ERROR DECODE: %s\n", err)
		}

		//		fmt.Printf("REQUEST:%+v\n", r.Body)
		fmt.Printf("BODYCONTENT:%+v\n", b)
		if b.Operation == "read-resource" {
			if b.AttributesOnly == "true" {
				if eap7 {
					fmt.Fprintln(w, readJson("testdata/jboss_exec_mode.json"))
				} else {
					fmt.Fprintln(w, readJson("testdata/jboss_exec_mode_eap6.json"))
				}
				return
			}
			if v, ok := b.Address[0]["core-service"]; ok {
				if v == "platform-mbean" {
					fmt.Fprintln(w, readJson("testdata/jboss_jvm_out.json"))
					return
				}
			}
			if v, ok := b.Address[0]["subsystem"]; ok {
				if v == "web" && b.Address[1]["connector"] == "http" {
					fmt.Fprintln(w, readJson("testdata/jboss_webcon_http.json"))
					return
				}
				if v == "datasources" {
					fmt.Fprintln(w, readJson("testdata/jboss_database_out.json"))
					return
				}
			}
			if v, ok := b.Address[0]["deployment"]; ok {
				switch v {
				case "sample.war":
					fmt.Fprintln(w, readJson("testdata/jboss_web_app_war.json"))
					return
				case "HelloWorld.ear":
					fmt.Fprintln(w, readJson("testdata/jboss_web_app_ear.json"))
					return
				}
			}
		}
		if b.Operation == "read-children-names" {
			switch b.ChildType {
			case "host":
				fmt.Fprintln(w, readJson("testdata/jboss_host_list.json"))
				return
			case "server":
			case "deployment":
				fmt.Fprintln(w, readJson("testdata/jboss_deployment_list.json"))
				return

			}
		}
		if b.Operation == "read-children-resources" {
			switch b.ChildType {
			case "jms-queue":
				fmt.Fprintln(w, readJson("testdata/jboss_jms_out.json"))
				return
			case "jms-topic":
				fmt.Fprintln(w, readJson("testdata/jboss_jms_out.json"))
				return
			}
		}
	}))

	return ts
}

func TestHTTPJboss(t *testing.T) {

	ts := testJBossServer(t, false)
	defer ts.Close()
	j := &JBoss{
		Servers:       []string{ts.URL},
		Username:      "",
		Password:      "",
		Authorization: "digest",
		Metrics: []string{
			"jvm",
			"web",
			"deployment",
			"database",
			"jms",
		},
		Log: testutil.Logger{},
	}

	//var acc testutil.Accumulator
	acc := &testutil.Accumulator{}
	err := acc.GatherError(j.Gather)
	require.NoError(t, err)
	//TEST JVM
	fields_jvm := map[string]interface{}{
		"thread-count":              float64(417),
		"peak-thread-count":         float64(428),
		"daemon-thread-count":       float64(297),
		"ConcurrentMarkSweep_count": float64(1),
		"ConcurrentMarkSweep_time":  float64(703),
		"ParNew_count":              float64(18),
		"ParNew_time":               float64(3259),
		"heap_committed":            float64(8.589869056e+09),
		"heap_init":                 float64(8.589934592e+09),
		"heap_max":                  float64(8.589869056e+09),
		"heap_used":                 float64(5.68541508e+09),
		"nonheap_committed":         float64(5.72850176e+08),
		"nonheap_init":              float64(5.39426816e+08),
		"nonheap_max":               float64(5.8720256e+08),
		"nonheap_used":              float64(3.90926664e+08),
	}
	acc.AssertContainsFields(t, "jboss_jvm", fields_jvm)

	//TEST WEB CONNETOR
	fields_web := map[string]interface{}{
		"bytesReceived":  float64(0),
		"bytesSent":      float64(0),
		"errorCount":     float64(0),
		"maxTime":        float64(0),
		"processingTime": float64(0),
		"requestCount":   float64(0),
	}
	acc.AssertContainsFields(t, "jboss_web", fields_web)

	//TEST WEBAPP WAR
	fields_web_app_sample_war := map[string]interface{}{
		"active-sessions":     float64(0),
		"expired-sessions":    float64(0),
		"max-active-sessions": float64(0),
		"sessions-created":    float64(0),
	}
	tags_web_app_sample_war := map[string]string{
		"jboss_host":   "standalone",
		"jboss_server": "standalone",
		"name":         "sample.war",
		"context-root": "/sample",
		"runtime_name": "sample.war",
	}

	acc.AssertContainsTaggedFields(t, "jboss_web_app", fields_web_app_sample_war, tags_web_app_sample_war)

	//TEST WEBAPP EAR
	fields_web_app_sample_ear := map[string]interface{}{
		"active-sessions":     float64(0),
		"expired-sessions":    float64(0),
		"max-active-sessions": float64(0),
		"sessions-created":    float64(0),
	}
	tags_web_app_sample_ear := map[string]string{
		"jboss_host":   "standalone",
		"jboss_server": "standalone",
		"name":         "web.war",
		"context-root": "/HelloWorld",
		"runtime_name": "HelloWorld.ear",
	}

	acc.AssertContainsTaggedFields(t, "jboss_web_app", fields_web_app_sample_ear, tags_web_app_sample_ear)

	//TEST DATASOURCES
	fields_datasource_exampleDS := map[string]interface{}{
		"in-use-count":    int64(0),
		"active-count":    int64(0),
		"available-count": int64(0),
	}
	fields_datasource_exampleOracle := map[string]interface{}{
		"in-use-count":    int64(0),
		"active-count":    int64(3),
		"available-count": int64(30),
	}

	tags_datasource_exampleDS := map[string]string{
		"jboss_host":   "standalone",
		"jboss_server": "standalone",
		"name":         "ExampleDS",
	}
	tags_datasource_exampleOracle := map[string]string{
		"jboss_host":   "standalone",
		"jboss_server": "standalone",
		"name":         "ExampleOracle",
	}

	acc.AssertContainsTaggedFields(t, "jboss_database", fields_datasource_exampleDS, tags_datasource_exampleDS)
	acc.AssertContainsTaggedFields(t, "jboss_database", fields_datasource_exampleOracle, tags_datasource_exampleOracle)

	// TEST JMS
	fields_jms_DLQ := map[string]interface{}{
		"message-count":   float64(0),
		"messages-added":  float64(0),
		"consumer-count":  float64(0),
		"scheduled-count": float64(0),
	}
	fields_jms_ExpiryQueue := map[string]interface{}{
		"message-count":   float64(0),
		"messages-added":  float64(0),
		"consumer-count":  float64(0),
		"scheduled-count": float64(0),
	}

	tags_jms_DLQ := map[string]string{
		"jboss_host":   "standalone",
		"jboss_server": "standalone",
		"name":         "DLQ",
	}
	tags_jms_ExpiryQueue := map[string]string{
		"jboss_host":   "standalone",
		"jboss_server": "standalone",
		"name":         "ExpiryQueue",
	}

	acc.AssertContainsTaggedFields(t, "jboss_jms", fields_jms_DLQ, tags_jms_DLQ)
	acc.AssertContainsTaggedFields(t, "jboss_jms", fields_jms_ExpiryQueue, tags_jms_ExpiryQueue)
}

func TestHTTPJbossEAP6Domain(t *testing.T) {

	ts := testJBossServer(t, false)
	defer ts.Close()
	j := &JBoss{
		Servers:  []string{ts.URL},
		Username: "",
		Password: "",
		Metrics: []string{
			"jvm",
			"web",
			"deployment",
			"database",
			"jms",
		},
		Log: testutil.Logger{},
	}
	//var acc testutil.Accumulator
	acc := &testutil.Accumulator{}
	err := acc.GatherError(j.Gather)
	require.NoError(t, err)
}
