package k8s

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/ericchiang/k8s/runtime"
	"github.com/ericchiang/k8s/watch/versioned"
	"github.com/golang/protobuf/proto"
)

// Decode events from a watch stream.
//
// See: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/protobuf.md#streaming-wire-format

// Watcher receives a stream of events tracking a particular resource within
// a namespace or across all namespaces.
//
// Watcher does not automatically reconnect. If a watch fails, a new watch must
// be initialized.
type Watcher struct {
	watcher interface {
		Next(Resource) (string, error)
		Close() error
	}
}

// Next decodes the next event from the watch stream. Errors are fatal, and
// indicate that the watcher should no longer be used, and must be recreated.
func (w *Watcher) Next(r Resource) (string, error) {
	return w.watcher.Next(r)
}

// Close closes the active connection with the API server being used for
// the watch.
func (w *Watcher) Close() error {
	return w.watcher.Close()
}

type watcherJSON struct {
	d *json.Decoder
	c io.Closer
}

func (w *watcherJSON) Close() error {
	return w.c.Close()
}

func (w *watcherJSON) Next(r Resource) (string, error) {
	var event struct {
		Type   string          `json:"type"`
		Object json.RawMessage `json:"object"`
	}
	if err := w.d.Decode(&event); err != nil {
		return "", fmt.Errorf("decode event: %v", err)
	}
	if event.Type == "" {
		return "", errors.New("wwatch event had no type field")
	}
	if err := json.Unmarshal([]byte(event.Object), r); err != nil {
		return "", fmt.Errorf("decode resource: %v", err)
	}
	return event.Type, nil
}

type watcherPB struct {
	r io.ReadCloser
}

func (w *watcherPB) Next(r Resource) (string, error) {
	msg, ok := r.(proto.Message)
	if !ok {
		return "", errors.New("object was not a protobuf message")
	}
	event, unknown, err := w.next()
	if err != nil {
		return "", err
	}
	if event.Type == nil || *event.Type == "" {
		return "", errors.New("watch event had no type field")
	}
	if err := proto.Unmarshal(unknown.Raw, msg); err != nil {
		return "", err
	}
	return *event.Type, nil
}

func (w *watcherPB) Close() error {
	return w.r.Close()
}

func (w *watcherPB) next() (*versioned.Event, *runtime.Unknown, error) {
	length := make([]byte, 4)
	if _, err := io.ReadFull(w.r, length); err != nil {
		return nil, nil, err
	}

	body := make([]byte, int(binary.BigEndian.Uint32(length)))
	if _, err := io.ReadFull(w.r, body); err != nil {
		return nil, nil, fmt.Errorf("read frame body: %v", err)
	}

	var event versioned.Event
	if err := proto.Unmarshal(body, &event); err != nil {
		return nil, nil, err
	}

	if event.Object == nil {
		return nil, nil, fmt.Errorf("event had no underlying object")
	}

	unknown, err := parseUnknown(event.Object.Raw)
	if err != nil {
		return nil, nil, err
	}

	return &event, unknown, nil
}

var unknownPrefix = []byte{0x6b, 0x38, 0x73, 0x00}

func parseUnknown(b []byte) (*runtime.Unknown, error) {
	if !bytes.HasPrefix(b, unknownPrefix) {
		return nil, errors.New("bytes did not start with expected prefix")
	}

	var u runtime.Unknown
	if err := proto.Unmarshal(b[len(unknownPrefix):], &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// Watch creates a watch on a resource. It takes an example Resource to
// determine what endpoint to watch.
//
// Watch does not automatically reconnect. If a watch fails, a new watch must
// be initialized.
//
// 		// Watch configmaps in the "kube-system" namespace
//		var configMap corev1.ConfigMap
//		watcher, err := client.Watch(ctx, "kube-system", &configMap)
//		if err != nil {
//			// handle error
//		}
//		defer watcher.Close() // Always close the returned watcher.
//
//		for {
//			cm := new(corev1.ConfigMap)
//			eventType, err := watcher.Next(cm)
//			if err != nil {
//				// watcher encountered and error, exit or create a new watcher
//			}
//			fmt.Println(eventType, *cm.Metadata.Name)
//		}
//
func (c *Client) Watch(ctx context.Context, namespace string, r Resource, options ...Option) (*Watcher, error) {
	url, err := resourceWatchURL(c.Endpoint, namespace, r, options...)
	if err != nil {
		return nil, err
	}

	ct := contentTypeFor(r)

	req, err := c.newRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", ct)

	resp, err := c.client().Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 != 2 {
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		return nil, newAPIError(resp.Header.Get("Content-Type"), resp.StatusCode, body)
	}

	if ct == contentTypePB {
		return &Watcher{&watcherPB{r: resp.Body}}, nil
	}

	return &Watcher{&watcherJSON{
		d: json.NewDecoder(resp.Body),
		c: resp.Body,
	}}, nil
}
