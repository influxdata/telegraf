package etherscan

import (
	"net/http"
	"net/url"
	"sort"
	"time"
)

type Request interface {
	URL() *url.URL
	LastRequest() time.Time
	MarkTriggered()
	Send(client *http.Client) (map[string]interface{}, error)
	Tags() map[string]string
}

type requestList []Request

type requestQueue map[Network]requestList

type _ sort.Interface

func (rq requestList) Len() int {
	return len(rq)
}

func (rq requestList) Swap(i, j int) {
	rq[i], rq[j] = rq[j], rq[i]
}

func (rq requestList) Less(i, j int) bool {
	return rq[i].LastRequest().After(rq[j].LastRequest())
}
