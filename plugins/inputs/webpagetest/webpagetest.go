package webpagetest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	host               string = "https://www.webpagetest.org"
	runTestEndpoint    string = "/runtest.php"
	testStatusEndpoint string = "/testStatus.php"
	cancelTestEndpoint string = "/cancelTest.php"
	jsonResultEndpoint string = "/jsonResult.php"
	format             string = "json"
)

type WebPageTest struct {
	ApiKey        string
	Urls          []string
	PollFrequency int
	MaxPollTime   int

	// Connectivity params
	DownloadBandwidth int
	UploadBandwidth   int
	RoundTripLatency  int
	PacketLossRate    int

	client *http.Client
}

var sampleConfig = `
[[inputs.webpagetest]]

  ## WebPageTest API Key
  ## Get from https://www.webpagetest.org/getkey.php
  api_key = "key"

  ## URLs to test
  urls = ["https://in.hotels.com/"]

  ## Lookup interval. You *probably* want this to run less frequently than
  interval = "1h"

  ## Since test results are not generated instantaneously,
  # pollFrequency = 5      # Polling frequency in seconds
  # maxPollTime = 120      # Maximum poll/wait time in seconds

  ## Network connectivity information
  ## Refer https://sites.google.com/a/webpagetest.org/docs/advanced-features/webpagetest-restful-apis#TOC-Specifying-connectivity
  # downloadBandwidth = 5000    # kbps
  # uploadBandwidth = 1000      # kbps
  # roundTripLatency = 28       # ms
  # packetLossRate = 0
`

func (w *WebPageTest) SampleConfig() string {
	return sampleConfig
}

func (w *WebPageTest) Description() string {
	return "Gathers metrics about URLs from the WebPageTest API"
}

func (w *WebPageTest) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	var outerr error

	for _, lookup := range w.Urls {
		wg.Add(1)
		go func(lookup string) {
			defer wg.Done()
			outerr = w.gatherUrl(lookup, acc)
		}(lookup)
	}

	wg.Wait()

	return outerr
}

func (w *WebPageTest) gatherUrl(lookup string, acc telegraf.Accumulator) error {
	if w.client == nil {
		client := &http.Client{}
		w.client = client
	}

	testId, err := w.runTest(lookup)
	if err != nil {
		return err
	}

	ready := false
	elapsed := 0
	for ready == false {
		time.Sleep(time.Duration(w.PollFrequency) * time.Second)
		elapsed += w.PollFrequency

		ready, err = w.testStatus(testId)
		if elapsed >= w.MaxPollTime {
			err = fmt.Errorf("maximum poll time of %d seconds exceeded, terminating the testId %s", w.MaxPollTime, testId)
			// Fire and forget call
			_, _ = w.cancelTest(testId)
			ready = true
		}
	}

	if err != nil {
		return err
	}

	err = w.getResult(testId, acc)
	if err != nil {
		return err
	}

	return nil
}

func (w *WebPageTest) getResult(testId string, acc telegraf.Accumulator) error {
	u, err := url.Parse(host)
	if err != nil {
		return err
	}

	u.Path = jsonResultEndpoint
	q := u.Query()
	q.Set("test", testId)
	u.RawQuery = q.Encode()

	resp, err := w.client.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result TestResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	w.gatherView(result.Data.Runs.First.FirstView, acc, "firstView")
	w.gatherView(result.Data.Runs.First.RepeatView, acc, "repeatView")

	return nil
}

func (w *WebPageTest) gatherView(view TestResultView, acc telegraf.Accumulator, viewType string) {
	tags := make(map[string]string)
	tags["url"] = view.URL
	tags["type"] = viewType
	fields := map[string]interface{}{
		"ttfb":              view.TTFB,
		"start_render":      view.Render,
		"speed_index":       view.SpeedIndex,
		"document_complete": view.LoadTime,
		"fully_loaded":      view.FullyLoaded,
		"bytes_in":          view.BytesIn,
		"bytes_in_doc":      view.BytesInDoc,
		"requests_doc":      view.RequestsDoc,
		"requests_full":     view.RequestsFull,
		"requests_css":      view.Breakdown.Css.Requests,
		"bytes_css":         view.Breakdown.Css.Bytes,
		"requests_image":    view.Breakdown.Image.Requests,
		"bytes_image":       view.Breakdown.Image.Bytes,
		"requests_js":       view.Breakdown.Js.Requests,
		"bytes_js":          view.Breakdown.Js.Bytes,
		"requests_html":     view.Breakdown.Html.Requests,
		"bytes_html":        view.Breakdown.Html.Bytes,
		"requests_font":     view.Breakdown.Font.Requests,
		"bytes_font":        view.Breakdown.Font.Bytes,
		"requests_other":    view.Breakdown.Other.Requests,
		"bytes_other":       view.Breakdown.Other.Bytes}
	acc.AddFields("webpagetest", fields, tags, time.Now())
}

func (w *WebPageTest) testStatus(testId string) (bool, error) {
	u, err := url.Parse(host)
	if err != nil {
		return true, err
	}

	u.Path = testStatusEndpoint
	q := u.Query()
	q.Set("test", testId)
	q.Set("f", format)
	u.RawQuery = q.Encode()

	resp, err := w.client.Get(u.String())
	if err != nil {
		return true, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return true, err
	}

	var result TestStatusResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return true, err
	}

	if result.Data.StatusCode == 200 {
		return true, nil
	} else {
		return false, nil
	}
}

func (w *WebPageTest) runTest(lookup string) (string, error) {
	u, err := url.Parse(host)
	if err != nil {
		return "", err
	}

	u.Path = runTestEndpoint
	q := u.Query()
	q.Set("url", lookup)
	q.Set("k", w.ApiKey)

	// Connectivity config
	q.Set("bwDown", strconv.Itoa(w.DownloadBandwidth))
	q.Set("bwUp", strconv.Itoa(w.UploadBandwidth))
	q.Set("latency", strconv.Itoa(w.RoundTripLatency))
	q.Set("plr", strconv.Itoa(w.PacketLossRate))

	q.Set("f", format)
	u.RawQuery = q.Encode()

	resp, err := w.client.Get(u.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s returned HTTP status %s", u.String(), resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result RunTestResult

	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	return result.Data.TestId, nil
}

func (w *WebPageTest) cancelTest(testId string) (bool, error) {
	u, err := url.Parse(host)
	if err != nil {
		return true, err
	}

	u.Path = cancelTestEndpoint
	q := u.Query()
	q.Set("test", testId)
	q.Set("k", w.ApiKey)
	q.Set("f", format)
	u.RawQuery = q.Encode()

	resp, err := w.client.Get(u.String())

	if http.StatusOK != resp.StatusCode {
		return true, err
	}

	return false, nil
}

type TestResult struct {
	Data struct {
		Runs struct {
			First struct {
				FirstView  TestResultView `json:"firstView"`
				RepeatView TestResultView `json:"repeatView"`
			} `json:"1"`
		} `json:"runs"`
	} `json:"data"`
}

type TestResultView struct {
	URL          string `json:"URL"`
	TTFB         int    `json:"TTFB"`
	Render       int    `json:"render"`
	SpeedIndex   int    `json:"SpeedIndex"`
	LoadTime     int    `json:"loadTime"`
	FullyLoaded  int    `json:"fullyLoaded"`
	BytesIn      int    `json:"bytesIn"`
	BytesInDoc   int    `json:"bytesInDoc"`
	RequestsDoc  int    `json:"requestsDoc"`
	RequestsFull int    `json:"requestsFull"`
	Breakdown    struct {
		Css struct {
			Bytes    int `json:"bytes"`
			Requests int `json:"requests"`
		} `json:"css"`
		Flash struct {
			Bytes    int `json:"bytes"`
			Requests int `json:"requests"`
		} `json:"flash"`
		Font struct {
			Bytes    int `json:"bytes"`
			Requests int `json:"requests"`
		} `json:"font"`
		Html struct {
			Bytes    int `json:"bytes"`
			Requests int `json:"requests"`
		} `json:"html"`
		Image struct {
			Bytes    int `json:"bytes"`
			Requests int `json:"requests"`
		} `json:"image"`
		Js struct {
			Bytes    int `json:"bytes"`
			Requests int `json:"requests"`
		} `json:"js"`
		Other struct {
			Bytes    int `json:"bytes"`
			Requests int `json:"requests"`
		} `json:"other"`
	} `json:"breakDown"`
}

type RunTestResult struct {
	Data struct {
		TestId string `json:"testId"`
	}
}

type TestStatusResult struct {
	Data struct {
		StatusCode int `json:"statusCode"`
	}
}

func init() {
	inputs.Add("webpagetest", func() telegraf.Input {
		return &WebPageTest{
			PollFrequency: 5,
			MaxPollTime: 120,
			DownloadBandwidth: 5000,
			UploadBandwidth: 1000,
			RoundTripLatency: 28,
			PacketLossRate: 0,
		}
	})
}