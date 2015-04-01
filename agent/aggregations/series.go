package aggregations

import (
	"code.google.com/p/go-sqlite/go1/sqlite3"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Series struct {
	context *Context
	Name    string
}

var cachedSeries = map[string]*Series{}

var seriesNameRegex = regexp.MustCompile("^[A-Za-z_][A-Za-z0-9_]*$")

func validateSeriesName(name string) error {
	if seriesNameRegex.MatchString(name) {
		return nil
	}

	return errors.New(fmt.Sprintf("Invalid series name `%s`. Series names must start with a letter or underscore and can only contain letters, underscores, and digits.", name))
}

func createSeries(context *Context, name string) error {
	result := &Series{
		context: context,
		Name:    name,
	}

	if err := result.createTable(); err != nil {
		return err
	}

	cachedSeries[name] = result

	return nil
}

func GetSeries(context *Context, name string) (*Series, error) {
	if err := validateSeriesName(name); err != nil {
		return nil, err
	}

	result := &Series{
		context: context,
		Name:    name,
	}

	if _, ok := cachedSeries[name]; !ok {
		if err := createSeries(context, name); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (s *Series) prepQuery(query string) string {
	return strings.Replace(query, "??", s.Name, -1)
}

func (s *Series) query(query string, values ...interface{}) (*sqlite3.Stmt, error) {
	return s.context.conn.Query(s.prepQuery(query), values...)
}

func (s *Series) exec(query string, values ...interface{}) error {
	return s.context.conn.Exec(s.prepQuery(query), values...)
}

func (s *Series) fetchRow(query string, values ...interface{}) (sqlite3.RowMap, error) {
	result := sqlite3.RowMap{}

	rs, err := s.query(query, values...)

	if err != nil {
		return result, err
	}

	defer rs.Close()

	rs.Scan(result)

	return result, nil
}

func (s *Series) createTable() error {
	if err := s.exec("CREATE TABLE IF NOT EXISTS ?? (ts DATETIME NOT NULL, value FLOAT)"); err != nil {
		return err
	}

	if err := s.exec("CREATE INDEX IF NOT EXISTS ??_index ON ?? (ts)"); err != nil {
		return err
	}

	return nil
}

func (s *Series) Push(timestamp *time.Time, value float64) error {
	if timestamp == nil {
		timestamp = &time.Time{}
		*timestamp = time.Now()
	}

	return s.exec("INSERT INTO ?? (ts, value) VALUES (?, ?)", *timestamp, value)
}

func (s *Series) Pop(shouldDelete bool) (map[string]interface{}, error) {
	row, err := s.fetchRow("SELECT rowid, ts, value FROM ?? ORDER BY ts DESC LIMIT 1")

	if err != nil {
		return row, err
	}

	rowId := row["rowid"]

	delete(row, "rowid")

	if shouldDelete {
		s.exec("DELETE FROM ?? WHERE rowid = ?", rowId)
	}

	return row, nil
}
