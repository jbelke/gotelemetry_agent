package aggregations

import (
	"code.google.com/p/go-sqlite/go1/sqlite3"
	"io"
)

type Series struct {
	manager *AggregationManager
	Name    string
	TTL     int
}

func NewSeries(manager *AggregationManager, name string, ttl int) (*Series, error) {
	if ttl < 0 {
		ttl = manager.ttl
	}

	result := &Series{
		manager: manager,
		Name:    name,
		TTL:     ttl,
	}

	if err := result.createTable(); err != nil {
		return nil, err
	}

	if err := result.setMetadata(); err != nil {
		return nil, err
	}

	return result, nil
}

func GetSeries(manager *AggregationManager, name string, ttl int) (*Series, error) {
	result := &Series{
		manager: manager,
		Name:    name,
		TTL:     0,
	}

	if err := manager.loadSeriesMetadata(result); err != nil {
		if err == io.EOF {
			return NewSeries(manager, name, ttl)
		}

		return nil, err
	}

	return result, nil
}

func (s *Series) createTable() error {
	return s.manager.withConnection(func(c *sqlite3.Conn) error {
		if err := c.Exec("CREATE TABLE IF NOT EXISTS `" + s.Name + "` (when DATETIME NOT NULL, value FLOAT)"); err != nil {
			return err
		}

		if err := c.Exec("CREATE INDEX IF NOT EXISTS `" + s.Name + "_index` ON " + s.Name + " (when)"); err != nil {
			return err
		}

		return nil
	})
}

func (s *Series) setMetadata() error {
	return s.manager.saveSeriesMetadata(s)
}
