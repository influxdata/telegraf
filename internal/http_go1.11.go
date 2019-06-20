// +build !go1.12

package internal

import "net/http"

func CloseIdleConnections(c *http.Client) {
	type closeIdler interface {
		CloseIdleConnections()
	}

	if tr, ok := c.Transport.(closeIdler); ok {
		tr.CloseIdleConnections()
	}
}
