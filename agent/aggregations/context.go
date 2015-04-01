package aggregations

import (
	"code.google.com/p/go-sqlite/go1/sqlite3"
	"fmt"
	"github.com/telemetryapp/gotelemetry"
)

type Context struct {
	conn          *sqlite3.Conn
	hasError      bool
	inTransaction bool
}

func GetContext() (*Context, error) {
	conn, err := sqlite3.Open(manager.path)

	if err != nil {
		return nil, err
	}

	result := &Context{
		conn: conn,
	}

	return result, nil
}

// Logf sends a formatted string to the agent's global log. It works like log.Logf
func (c *Context) Logf(format string, v ...interface{}) {
	if manager.errorChannel != nil {
		manager.errorChannel <- gotelemetry.NewLogError("Data Manager -> %#s", fmt.Sprintf(format, v...))
	}
}

// Debugf sends a formatted string to the agent's debug log, if it exists. It works like log.Logf
func (c *Context) Debugf(format string, v ...interface{}) {
	if manager.errorChannel != nil {
		manager.errorChannel <- gotelemetry.NewDebugError("Data Manager -> %#s", fmt.Sprintf(format, v...))
	}
}

func (c *Context) SetError() {
	c.hasError = true
}

// Transactions

func (c *Context) Begin() error {
	c.inTransaction = true

	return c.conn.Begin()
}

func (c *Context) Close() {
	if c.inTransaction {
		if c.hasError {
			c.conn.Rollback()
		} else {
			c.conn.Commit()
		}
	}

	c.conn.Close()
}
