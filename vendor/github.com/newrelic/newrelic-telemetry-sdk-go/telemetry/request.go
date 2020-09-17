// Copyright 2019 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/newrelic/newrelic-telemetry-sdk-go/internal"
)

const (
	maxCompressedSizeBytes = 1 << 20
)

// request contains an http.Request and the UncompressedBody which is provided
// for logging.
type request struct {
	Request          *http.Request
	UncompressedBody json.RawMessage

	compressedBodyLength int
}

type requestsBuilder interface {
	makeBody() json.RawMessage
	split() []requestsBuilder
}

var (
	errUnableToSplit = fmt.Errorf("unable to split large payload further")
)

func requestNeedsSplit(r request) bool {
	return r.compressedBodyLength >= maxCompressedSizeBytes
}

func newRequests(batch requestsBuilder, apiKey string, url string, userAgent string) ([]request, error) {
	return newRequestsInternal(batch, apiKey, url, userAgent, requestNeedsSplit)
}

func newRequestsInternal(batch requestsBuilder, apiKey string, url string, userAgent string, needsSplit func(request) bool) ([]request, error) {
	uncompressed := batch.makeBody()
	compressed, err := internal.Compress(uncompressed)
	if nil != err {
		return nil, fmt.Errorf("error compressing data: %v", err)
	}
	compressedLen := compressed.Len()

	req, err := http.NewRequest("POST", url, compressed)
	if nil != err {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Api-Key", apiKey)
	req.Header.Add("Content-Encoding", "gzip")
	req.Header.Add("User-Agent", userAgent)
	r := request{
		Request:              req,
		UncompressedBody:     uncompressed,
		compressedBodyLength: compressedLen,
	}

	if !needsSplit(r) {
		return []request{r}, nil
	}

	var reqs []request
	batches := batch.split()
	if nil == batches {
		return nil, errUnableToSplit
	}

	for _, b := range batches {
		rs, err := newRequestsInternal(b, apiKey, url, userAgent, needsSplit)
		if nil != err {
			return nil, err
		}
		reqs = append(reqs, rs...)
	}
	return reqs, nil
}
