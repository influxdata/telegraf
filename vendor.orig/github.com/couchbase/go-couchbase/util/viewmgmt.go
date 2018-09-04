// Package couchbaseutil offers some convenience functions for apps
// that use couchbase.
package couchbaseutil

import (
	"encoding/json"
	"log"
	"time"

	"github.com/couchbase/go-couchbase"
)

// A ViewMarker is stored in your DB to mark a particular view
// version.
type ViewMarker struct {
	Version   int       `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}

// UpdateView installs or updates a view.
//
// This creates a document that tracks the version of design document
// in couchbase and updates it if it's behind the version specified.
//
// A ViewMarker is stored with a type of "viewmarker" under the key
// specified by `markerKey` to keep up with the view info.
func UpdateView(d *couchbase.Bucket,
	ddocName, markerKey, ddocBody string, version int) error {

	marker := ViewMarker{}
	err := d.Get(markerKey, &marker)
	if err != nil {
		log.Printf("Error checking view version: %v", err)
	}
	if marker.Version < version {
		log.Printf("Installing new version of views (old version=%v)",
			marker.Version)
		doc := json.RawMessage([]byte(ddocBody))
		err = d.PutDDoc(ddocName, &doc)
		if err != nil {
			return err
		}
		marker.Version = version
		marker.Timestamp = time.Now().UTC()
		marker.Type = "viewmarker"

		return d.Set(markerKey, 0, &marker)
	}
	return nil
}
