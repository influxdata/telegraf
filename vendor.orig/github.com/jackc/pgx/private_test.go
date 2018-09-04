package pgx

// This file contains methods that expose internal pgx state to tests.

func (c *Conn) TxStatus() byte {
	return c.txStatus
}
