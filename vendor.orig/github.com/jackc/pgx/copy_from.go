package pgx

import (
	"bytes"
	"fmt"

	"github.com/jackc/pgx/pgio"
	"github.com/jackc/pgx/pgproto3"
	"github.com/pkg/errors"
)

// CopyFromRows returns a CopyFromSource interface over the provided rows slice
// making it usable by *Conn.CopyFrom.
func CopyFromRows(rows [][]interface{}) CopyFromSource {
	return &copyFromRows{rows: rows, idx: -1}
}

type copyFromRows struct {
	rows [][]interface{}
	idx  int
}

func (ctr *copyFromRows) Next() bool {
	ctr.idx++
	return ctr.idx < len(ctr.rows)
}

func (ctr *copyFromRows) Values() ([]interface{}, error) {
	return ctr.rows[ctr.idx], nil
}

func (ctr *copyFromRows) Err() error {
	return nil
}

// CopyFromSource is the interface used by *Conn.CopyFrom as the source for copy data.
type CopyFromSource interface {
	// Next returns true if there is another row and makes the next row data
	// available to Values(). When there are no more rows available or an error
	// has occurred it returns false.
	Next() bool

	// Values returns the values for the current row.
	Values() ([]interface{}, error)

	// Err returns any error that has been encountered by the CopyFromSource. If
	// this is not nil *Conn.CopyFrom will abort the copy.
	Err() error
}

type copyFrom struct {
	conn          *Conn
	tableName     Identifier
	columnNames   []string
	rowSrc        CopyFromSource
	readerErrChan chan error
}

func (ct *copyFrom) readUntilReadyForQuery() {
	for {
		msg, err := ct.conn.rxMsg()
		if err != nil {
			ct.readerErrChan <- err
			close(ct.readerErrChan)
			return
		}

		switch msg := msg.(type) {
		case *pgproto3.ReadyForQuery:
			ct.conn.rxReadyForQuery(msg)
			close(ct.readerErrChan)
			return
		case *pgproto3.CommandComplete:
		case *pgproto3.ErrorResponse:
			ct.readerErrChan <- ct.conn.rxErrorResponse(msg)
		default:
			err = ct.conn.processContextFreeMsg(msg)
			if err != nil {
				ct.readerErrChan <- ct.conn.processContextFreeMsg(msg)
			}
		}
	}
}

func (ct *copyFrom) waitForReaderDone() error {
	var err error
	for err = range ct.readerErrChan {
	}
	return err
}

func (ct *copyFrom) run() (int, error) {
	quotedTableName := ct.tableName.Sanitize()
	cbuf := &bytes.Buffer{}
	for i, cn := range ct.columnNames {
		if i != 0 {
			cbuf.WriteString(", ")
		}
		cbuf.WriteString(quoteIdentifier(cn))
	}
	quotedColumnNames := cbuf.String()

	ps, err := ct.conn.Prepare("", fmt.Sprintf("select %s from %s", quotedColumnNames, quotedTableName))
	if err != nil {
		return 0, err
	}

	err = ct.conn.sendSimpleQuery(fmt.Sprintf("copy %s ( %s ) from stdin binary;", quotedTableName, quotedColumnNames))
	if err != nil {
		return 0, err
	}

	err = ct.conn.readUntilCopyInResponse()
	if err != nil {
		return 0, err
	}

	go ct.readUntilReadyForQuery()
	defer ct.waitForReaderDone()

	buf := ct.conn.wbuf
	buf = append(buf, copyData)
	sp := len(buf)
	buf = pgio.AppendInt32(buf, -1)

	buf = append(buf, "PGCOPY\n\377\r\n\000"...)
	buf = pgio.AppendInt32(buf, 0)
	buf = pgio.AppendInt32(buf, 0)

	var sentCount int

	for ct.rowSrc.Next() {
		select {
		case err = <-ct.readerErrChan:
			return 0, err
		default:
		}

		if len(buf) > 65536 {
			pgio.SetInt32(buf[sp:], int32(len(buf[sp:])))
			_, err = ct.conn.conn.Write(buf)
			if err != nil {
				ct.conn.die(err)
				return 0, err
			}

			// Directly manipulate wbuf to reset to reuse the same buffer
			buf = buf[0:5]
		}

		sentCount++

		values, err := ct.rowSrc.Values()
		if err != nil {
			ct.cancelCopyIn()
			return 0, err
		}
		if len(values) != len(ct.columnNames) {
			ct.cancelCopyIn()
			return 0, errors.Errorf("expected %d values, got %d values", len(ct.columnNames), len(values))
		}

		buf = pgio.AppendInt16(buf, int16(len(ct.columnNames)))
		for i, val := range values {
			buf, err = encodePreparedStatementArgument(ct.conn.ConnInfo, buf, ps.FieldDescriptions[i].DataType, val)
			if err != nil {
				ct.cancelCopyIn()
				return 0, err
			}

		}
	}

	if ct.rowSrc.Err() != nil {
		ct.cancelCopyIn()
		return 0, ct.rowSrc.Err()
	}

	buf = pgio.AppendInt16(buf, -1) // terminate the copy stream
	pgio.SetInt32(buf[sp:], int32(len(buf[sp:])))

	buf = append(buf, copyDone)
	buf = pgio.AppendInt32(buf, 4)

	_, err = ct.conn.conn.Write(buf)
	if err != nil {
		ct.conn.die(err)
		return 0, err
	}

	err = ct.waitForReaderDone()
	if err != nil {
		return 0, err
	}
	return sentCount, nil
}

func (c *Conn) readUntilCopyInResponse() error {
	for {
		msg, err := c.rxMsg()
		if err != nil {
			return err
		}

		switch msg := msg.(type) {
		case *pgproto3.CopyInResponse:
			return nil
		default:
			err = c.processContextFreeMsg(msg)
			if err != nil {
				return err
			}
		}
	}
}

func (ct *copyFrom) cancelCopyIn() error {
	buf := ct.conn.wbuf
	buf = append(buf, copyFail)
	sp := len(buf)
	buf = pgio.AppendInt32(buf, -1)
	buf = append(buf, "client error: abort"...)
	buf = append(buf, 0)
	pgio.SetInt32(buf[sp:], int32(len(buf[sp:])))

	_, err := ct.conn.conn.Write(buf)
	if err != nil {
		ct.conn.die(err)
		return err
	}

	return nil
}

// CopyFrom uses the PostgreSQL copy protocol to perform bulk data insertion.
// It returns the number of rows copied and an error.
//
// CopyFrom requires all values use the binary format. Almost all types
// implemented by pgx use the binary format by default. Types implementing
// Encoder can only be used if they encode to the binary format.
func (c *Conn) CopyFrom(tableName Identifier, columnNames []string, rowSrc CopyFromSource) (int, error) {
	ct := &copyFrom{
		conn:          c,
		tableName:     tableName,
		columnNames:   columnNames,
		rowSrc:        rowSrc,
		readerErrChan: make(chan error),
	}

	return ct.run()
}
