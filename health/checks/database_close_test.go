package checks

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"
)

var errRowsClose = errors.New("test row-close failure")

type closeErrorDriver struct{}

func (closeErrorDriver) Open(string) (driver.Conn, error) { return closeErrorConn{}, nil }

type closeErrorConn struct{}

func (closeErrorConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("not supported") }
func (closeErrorConn) Close() error                        { return nil }
func (closeErrorConn) Begin() (driver.Tx, error)           { return nil, errors.New("not supported") }
func (closeErrorConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return closeErrorRows{}, nil
}

type closeErrorRows struct{}

func (closeErrorRows) Columns() []string         { return []string{"one"} }
func (closeErrorRows) Next([]driver.Value) error { return io.EOF }
func (closeErrorRows) Close() error              { return errRowsClose }

func init() { sql.Register("fastapp-close-error", closeErrorDriver{}) }

func TestDatabaseQueryReportsCloseError(t *testing.T) {
	db, err := sql.Open("fastapp-close-error", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	check := NewDatabaseCheckWithOptions("db", db, DatabaseOptions{Query: "SELECT 1"})
	if err := check.executeQuery(context.Background()); !errors.Is(err, errRowsClose) {
		t.Fatalf("close error lost: %v", err)
	}
}
