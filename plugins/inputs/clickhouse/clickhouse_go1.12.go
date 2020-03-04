// +build go1.12

package clickhouse

// Stop ClickHouse input service
func (ch *ClickHouse) Stop() {
	ch.client.CloseIdleConnections()
}
