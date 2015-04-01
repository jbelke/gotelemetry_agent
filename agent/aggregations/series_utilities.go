package aggregations

import (
	"code.google.com/p/go-sqlite/go1/sqlite3"
	"errors"
	"fmt"
	"strings"
)

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
	if err := s.exec("CREATE TABLE IF NOT EXISTS ?? (ts INT NOT NULL, value FLOAT)"); err != nil {
		return err
	}

	if err := s.exec("CREATE INDEX IF NOT EXISTS ??_index ON ?? (ts)"); err != nil {
		return err
	}

	return nil
}
