// +build go1.12

package internal

import "net/http"

func CloseIdleConnections(c *http.Client) {
	c.CloseIdleConnections()
}
