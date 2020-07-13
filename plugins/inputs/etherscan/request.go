package etherscan

import (
	"net/http"
	"net/url"
	"sort"
	"time"
)

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
}

type Request interface {
	URL() *url.URL
	LastRequest() time.Time
	MarkTriggered()
	Send(client HTTPClient) (map[string]interface{}, error)
	Tags() map[string]string
}

type requestList []Request

type requestQueue map[Network]requestList

func (rq requestQueue) Len() int {
	return len(rq)
}

type _ sort.Interface

func (rl requestList) Len() int {
	return len(rl)
}

func (rl requestList) Swap(i, j int) {
	rl[i], rl[j] = rl[j], rl[i]
}

func (rl requestList) Less(i, j int) bool {
	return rl[i].LastRequest().After(rl[j].LastRequest())
}
